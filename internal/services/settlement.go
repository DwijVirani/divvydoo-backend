package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"divvydoo/backend/internal/models"
	"divvydoo/backend/internal/repositories"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrSettlementNotFound  = errors.New("settlement not found")
	ErrInvalidSettlement   = errors.New("invalid settlement request")
	ErrSettlementCompleted = errors.New("settlement is already completed")
)

type SettlementService struct {
	settlementRepo repositories.SettlementRepository
	balanceRepo    repositories.BalanceRepository
	userRepo       repositories.UserRepository
}

func NewSettlementService(
	settlementRepo repositories.SettlementRepository,
	balanceRepo repositories.BalanceRepository,
	userRepo repositories.UserRepository,
) *SettlementService {
	return &SettlementService{
		settlementRepo: settlementRepo,
		balanceRepo:    balanceRepo,
		userRepo:       userRepo,
	}
}

func (s *SettlementService) CreateSettlement(ctx context.Context, req models.SettlementRequest) (*models.Settlement, error) {
	if req.FromUserID == req.ToUserID {
		return nil, ErrInvalidSettlement
	}

	// Validate both users exist in one query
	missingUsers, err := s.userRepo.ExistMultiple(ctx, []string{req.FromUserID, req.ToUserID})
	if err != nil {
		return nil, err
	}
	if len(missingUsers) > 0 {
		return nil, fmt.Errorf("user %s does not exist", missingUsers[0])
	}

	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	settlement := &models.Settlement{
		SettlementID: uuid.New().String(),
		FromUserID:   req.FromUserID,
		ToUserID:     req.ToUserID,
		GroupID:      req.GroupID,
		Amount:       req.Amount,
		Currency:     req.Currency,
		Status:       models.SettlementPending,
		Method:       req.Method,
		Description:  req.Description,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return s.settlementRepo.Create(ctx, settlement)
}

func (s *SettlementService) GetSettlement(ctx context.Context, settlementID string, userID string) (*models.Settlement, error) {
	settlement, err := s.settlementRepo.GetByID(ctx, settlementID)
	if err != nil {
		if errors.Is(err, repositories.ErrSettlementNotFound) {
			return nil, ErrSettlementNotFound
		}
		return nil, err
	}

	// Check if user is involved in the settlement
	if settlement.FromUserID != userID && settlement.ToUserID != userID {
		return nil, ErrSettlementNotFound
	}

	return settlement, nil
}

func (s *SettlementService) GetUserSettlements(ctx context.Context, userID string, limit, offset int64) ([]*models.Settlement, error) {
	return s.settlementRepo.GetByUserID(ctx, userID, limit, offset)
}

func (s *SettlementService) GetGroupSettlements(ctx context.Context, groupID string, limit, offset int64) ([]*models.Settlement, error) {
	return s.settlementRepo.GetByGroupID(ctx, groupID, limit, offset)
}

func (s *SettlementService) CompleteSettlement(ctx context.Context, settlementID string, userID string, transactionID *string) error {
	settlement, err := s.settlementRepo.GetByID(ctx, settlementID)
	if err != nil {
		return err
	}

	// Only the person who owes money can mark it as complete
	if settlement.FromUserID != userID {
		return fmt.Errorf("only the payer can complete the settlement")
	}

	if settlement.Status != models.SettlementPending {
		return ErrSettlementCompleted
	}

	// Start a transaction to update both settlement and balances
	session, err := s.settlementRepo.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Mark settlement as completed
		if err := s.settlementRepo.MarkCompleted(sessCtx, settlementID, transactionID); err != nil {
			return nil, err
		}

		// Update balances: from_user pays to_user
		// from_user's balance increases (they owe less)
		if err := s.balanceRepo.UpdateBalance(sessCtx, settlement.FromUserID, settlement.GroupID, settlement.Amount); err != nil {
			return nil, err
		}

		// to_user's balance decreases (they are owed less)
		if err := s.balanceRepo.UpdateBalance(sessCtx, settlement.ToUserID, settlement.GroupID, -settlement.Amount); err != nil {
			return nil, err
		}

		// Record balance history
		now := time.Now()
		fromHistory := &models.BalanceHistory{
			UserID:      settlement.FromUserID,
			GroupID:     settlement.GroupID,
			Amount:      settlement.Amount,
			Currency:    settlement.Currency,
			Type:        models.BalanceChangeSettlement,
			ReferenceID: settlementID,
			Description: fmt.Sprintf("Settlement payment to user"),
			CreatedAt:   now,
		}
		if err := s.balanceRepo.CreateBalanceHistory(sessCtx, fromHistory); err != nil {
			return nil, err
		}

		toHistory := &models.BalanceHistory{
			UserID:      settlement.ToUserID,
			GroupID:     settlement.GroupID,
			Amount:      -settlement.Amount,
			Currency:    settlement.Currency,
			Type:        models.BalanceChangeSettlement,
			ReferenceID: settlementID,
			Description: fmt.Sprintf("Settlement received from user"),
			CreatedAt:   now,
		}
		if err := s.balanceRepo.CreateBalanceHistory(sessCtx, toHistory); err != nil {
			return nil, err
		}

		return nil, nil
	})

	return err
}

func (s *SettlementService) CancelSettlement(ctx context.Context, settlementID string, userID string) error {
	settlement, err := s.settlementRepo.GetByID(ctx, settlementID)
	if err != nil {
		return err
	}

	// Only involved parties can cancel
	if settlement.FromUserID != userID && settlement.ToUserID != userID {
		return ErrSettlementNotFound
	}

	if settlement.Status != models.SettlementPending {
		return fmt.Errorf("can only cancel pending settlements")
	}

	return s.settlementRepo.MarkCancelled(ctx, settlementID)
}

func (s *SettlementService) GetPendingSettlements(ctx context.Context, userID string) ([]*models.Settlement, error) {
	return s.settlementRepo.GetPendingSettlements(ctx, userID)
}
