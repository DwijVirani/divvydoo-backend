package services

import (
	"context"
	"errors"

	"divvydoo/backend/internal/models"
	"divvydoo/backend/internal/repositories"
)

var (
	ErrBalanceNotFound = errors.New("balance not found")
)

type BalanceService struct {
	balanceRepo repositories.BalanceRepository
	expenseRepo repositories.ExpenseRepository
	userRepo    repositories.UserRepository
}

func NewBalanceService(
	balanceRepo repositories.BalanceRepository,
	expenseRepo repositories.ExpenseRepository,
	userRepo repositories.UserRepository,
) *BalanceService {
	return &BalanceService{
		balanceRepo: balanceRepo,
		expenseRepo: expenseRepo,
		userRepo:    userRepo,
	}
}

func (s *BalanceService) GetUserBalances(ctx context.Context, userID string) (*models.UserBalanceSummary, error) {
	// Verify user exists
	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	return s.balanceRepo.GetUserBalanceSummary(ctx, userID)
}

func (s *BalanceService) GetGroupBalances(ctx context.Context, groupID string) ([]*models.Balance, error) {
	return s.balanceRepo.GetByGroupID(ctx, groupID)
}

func (s *BalanceService) GetBalanceHistory(ctx context.Context, userID string, groupID *string, limit, offset int64) ([]*models.BalanceHistory, error) {
	return s.balanceRepo.GetBalanceHistory(ctx, userID, groupID, limit, offset)
}

func (s *BalanceService) GetUserBalanceInGroup(ctx context.Context, userID string, groupID string) (*models.Balance, error) {
	balance, err := s.balanceRepo.GetByUserAndGroup(ctx, userID, &groupID)
	if err != nil {
		if errors.Is(err, repositories.ErrBalanceNotFound) {
			// Return zero balance if not found
			return &models.Balance{
				UserID:   userID,
				GroupID:  &groupID,
				Balance:  0,
				Currency: "USD",
			}, nil
		}
		return nil, err
	}
	return balance, nil
}

func (s *BalanceService) RecordBalanceChange(ctx context.Context, history *models.BalanceHistory) error {
	return s.balanceRepo.CreateBalanceHistory(ctx, history)
}
