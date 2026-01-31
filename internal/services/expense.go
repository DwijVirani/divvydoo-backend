package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"divvydoo/backend/internal/models"
	"divvydoo/backend/internal/repositories"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

type ExpenseService struct {
	expenseRepo repositories.ExpenseRepository
	balanceRepo repositories.BalanceRepository
	groupRepo   repositories.GroupRepository
	userRepo    repositories.UserRepository
}

func NewExpenseService(
	expenseRepo repositories.ExpenseRepository,
	balanceRepo repositories.BalanceRepository,
	groupRepo repositories.GroupRepository,
	userRepo repositories.UserRepository,
) *ExpenseService {
	return &ExpenseService{
		expenseRepo: expenseRepo,
		balanceRepo: balanceRepo,
		groupRepo:   groupRepo,
		userRepo:    userRepo,
	}
}

func (s *ExpenseService) CreateExpense(ctx context.Context, expense models.Expense) (*models.Expense, error) {
	// Validate the expense
	if err := validateExpense(expense); err != nil {
		return nil, err
	}

	// Check if all users exist
	if err := s.validateUsersExist(ctx, expense); err != nil {
		return nil, err
	}

	// Check group membership if it's a group expense
	if expense.GroupID != nil {
		if err := s.validateGroupMembership(ctx, *expense.GroupID, expense); err != nil {
			return nil, err
		}
	}

	// Calculate shares based on split type
	shares, err := s.calculateShares(expense)
	if err != nil {
		return nil, err
	}

	// Set calculated shares back to expense
	expense.Split.Details = shares

	// Generate expense ID
	expense.ExpenseID = uuid.New().String()
	expense.CreatedAt = time.Now()
	expense.UpdatedAt = expense.CreatedAt

	// Start MongoDB transaction
	session, err := s.expenseRepo.StartSession()
	if err != nil {
		return nil, fmt.Errorf("failed to start session: %v", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Save the expense
		createdExpense, err := s.expenseRepo.CreateExpense(sessCtx, expense)
		if err != nil {
			return nil, err
		}

		// Update balances
		if err := s.updateBalances(sessCtx, *createdExpense); err != nil {
			return nil, err
		}

		return createdExpense, nil
	})

	if err != nil {
		return nil, fmt.Errorf("transaction failed: %v", err)
	}

	return &expense, nil
}

func validateExpense(expense models.Expense) error {
	if expense.Amount <= 0 {
		return errors.New("amount must be positive")
	}

	if len(expense.PaidBy) == 0 {
		return errors.New("at least one payer must be specified")
	}

	totalPaid := 0.0
	for _, pb := range expense.PaidBy {
		if pb.Amount <= 0 {
			return fmt.Errorf("invalid amount for user %s", pb.UserID)
		}
		totalPaid += pb.Amount
	}

	if math.Abs(totalPaid-expense.Amount) > 0.01 { // Allow for small floating point differences
		return fmt.Errorf("total paid amount %.2f does not match expense amount %.2f", totalPaid, expense.Amount)
	}

	switch expense.Split.Type {
	case models.SplitEqual, models.SplitExact, models.SplitPercentage, models.SplitShares:
		// Valid types
	default:
		return fmt.Errorf("invalid split type: %s", expense.Split.Type)
	}

	return nil
}

func (s *ExpenseService) calculateShares(expense models.Expense) ([]models.SplitShare, error) {
	switch expense.Split.Type {
	case models.SplitEqual:
		return s.calculateEqualShares(expense)
	case models.SplitExact:
		return s.calculateExactShares(expense)
	case models.SplitPercentage:
		return s.calculatePercentageShares(expense)
	case models.SplitShares:
		return s.calculateShareBased(expense)
	default:
		return nil, fmt.Errorf("unsupported split type: %s", expense.Split.Type)
	}
}

func (s *ExpenseService) calculateEqualShares(expense models.Expense) ([]models.SplitShare, error) {
	// Get all participants (unique user IDs from paid_by and split details)
	participants := make(map[string]bool)
	for _, pb := range expense.PaidBy {
		participants[pb.UserID] = true
	}

	// For equal split, we expect all users in the group to participate
	// If it's a group expense, we need to get all group members
	// For simplicity, we'll assume split.details contains all participating users
	for _, share := range expense.Split.Details {
		participants[share.UserID] = true
	}

	numParticipants := len(participants)
	if numParticipants == 0 {
		return nil, errors.New("no participants found for equal split")
	}

	equalShare := expense.Amount / float64(numParticipants)

	var shares []models.SplitShare
	for userID := range participants {
		shares = append(shares, models.SplitShare{
			UserID: userID,
			Value:  equalShare,
		})
	}

	// Handle rounding errors by adjusting the first user's share
	if len(shares) > 0 {
		total := equalShare * float64(numParticipants)
		diff := expense.Amount - total
		shares[0].Value += diff
	}

	return shares, nil
}

func (s *ExpenseService) calculateExactShares(expense models.Expense) ([]models.SplitShare, error) {
	if len(expense.Split.Details) == 0 {
		return nil, errors.New("exact split requires split details with specific amounts")
	}

	// Validate that all specified amounts are positive
	totalSpecified := 0.0
	for _, share := range expense.Split.Details {
		if share.Value <= 0 {
			return nil, fmt.Errorf("invalid amount %.2f for user %s in exact split", share.Value, share.UserID)
		}
		totalSpecified += share.Value
	}

	// Check if total specified amounts match the expense amount
	if math.Abs(totalSpecified-expense.Amount) > 0.01 { // Allow for small floating point differences
		return nil, fmt.Errorf("total specified amounts %.2f do not match expense amount %.2f", totalSpecified, expense.Amount)
	}

	// Return the exact shares as specified
	var shares []models.SplitShare
	for _, share := range expense.Split.Details {
		shares = append(shares, models.SplitShare{
			UserID: share.UserID,
			Value:  share.Value,
		})
	}

	return shares, nil
}

func (s *ExpenseService) calculatePercentageShares(expense models.Expense) ([]models.SplitShare, error) {
	if len(expense.Split.Details) == 0 {
		return nil, errors.New("percentage split requires split details with percentages")
	}

	// Validate that all percentages are positive and sum to 100
	totalPercentage := 0.0
	for _, share := range expense.Split.Details {
		if share.Value <= 0 || share.Value > 100 {
			return nil, fmt.Errorf("invalid percentage %.2f for user %s", share.Value, share.UserID)
		}
		totalPercentage += share.Value
	}

	if math.Abs(totalPercentage-100.0) > 0.01 {
		return nil, fmt.Errorf("total percentage %.2f does not equal 100", totalPercentage)
	}

	// Calculate actual amounts based on percentages
	var shares []models.SplitShare
	totalCalculated := 0.0

	for i, share := range expense.Split.Details {
		amount := (share.Value / 100.0) * expense.Amount

		// Handle rounding for the last share to ensure total matches exactly
		if i == len(expense.Split.Details)-1 {
			amount = expense.Amount - totalCalculated
		}

		shares = append(shares, models.SplitShare{
			UserID: share.UserID,
			Value:  amount,
		})
		totalCalculated += amount
	}

	return shares, nil
}

func (s *ExpenseService) calculateShareBased(expense models.Expense) ([]models.SplitShare, error) {
	if len(expense.Split.Details) == 0 {
		return nil, errors.New("share-based split requires split details with share counts")
	}

	// Calculate total shares
	totalShares := 0.0
	for _, share := range expense.Split.Details {
		if share.Value <= 0 {
			return nil, fmt.Errorf("invalid share count %.2f for user %s", share.Value, share.UserID)
		}
		totalShares += share.Value
	}

	if totalShares == 0 {
		return nil, errors.New("total shares cannot be zero")
	}

	// Calculate actual amounts based on share ratios
	var shares []models.SplitShare
	totalCalculated := 0.0

	for i, share := range expense.Split.Details {
		amount := (share.Value / totalShares) * expense.Amount

		// Handle rounding for the last share to ensure total matches exactly
		if i == len(expense.Split.Details)-1 {
			amount = expense.Amount - totalCalculated
		}

		shares = append(shares, models.SplitShare{
			UserID: share.UserID,
			Value:  amount,
		})
		totalCalculated += amount
	}

	return shares, nil
}

func (s *ExpenseService) validateUsersExist(ctx context.Context, expense models.Expense) error {
	// Collect all unique user IDs from the expense
	userIDSet := make(map[string]bool)
	userIDSet[expense.CreatorID] = true
	for _, pb := range expense.PaidBy {
		userIDSet[pb.UserID] = true
	}
	for _, share := range expense.Split.Details {
		userIDSet[share.UserID] = true
	}

	// Convert to slice
	userIDs := make([]string, 0, len(userIDSet))
	for id := range userIDSet {
		userIDs = append(userIDs, id)
	}

	// Batch check all users in one query
	missingUsers, err := s.userRepo.ExistMultiple(ctx, userIDs)
	if err != nil {
		return fmt.Errorf("failed to check user existence: %v", err)
	}
	if len(missingUsers) > 0 {
		return fmt.Errorf("user %s does not exist", missingUsers[0])
	}

	return nil
}

func (s *ExpenseService) validateGroupMembership(ctx context.Context, groupID string, expense models.Expense) error {
	// Collect all unique user IDs from the expense
	userIDSet := make(map[string]bool)
	userIDSet[expense.CreatorID] = true
	for _, pb := range expense.PaidBy {
		userIDSet[pb.UserID] = true
	}
	for _, share := range expense.Split.Details {
		userIDSet[share.UserID] = true
	}

	// Convert to slice
	userIDs := make([]string, 0, len(userIDSet))
	for id := range userIDSet {
		userIDs = append(userIDs, id)
	}

	// Batch check all memberships in one query
	nonMembers, err := s.groupRepo.GetNonMembers(ctx, groupID, userIDs)
	if err != nil {
		return fmt.Errorf("failed to check group membership: %v", err)
	}
	if len(nonMembers) > 0 {
		return fmt.Errorf("user %s is not a member of group %s", nonMembers[0], groupID)
	}

	return nil
}

func (s *ExpenseService) GetExpense(ctx context.Context, expenseID string, userID string) (*models.Expense, error) {
	expense, err := s.expenseRepo.GetByID(ctx, expenseID)
	if err != nil {
		return nil, err
	}

	// Check if user has access to this expense
	hasAccess := expense.CreatorID == userID
	if !hasAccess {
		for _, pb := range expense.PaidBy {
			if pb.UserID == userID {
				hasAccess = true
				break
			}
		}
	}
	if !hasAccess {
		for _, share := range expense.Split.Details {
			if share.UserID == userID {
				hasAccess = true
				break
			}
		}
	}

	if !hasAccess {
		return nil, fmt.Errorf("user does not have access to this expense")
	}

	return expense, nil
}

func (s *ExpenseService) GetGroupExpenses(ctx context.Context, groupID string, limit, offset int64) ([]*models.Expense, error) {
	return s.expenseRepo.GetByGroupID(ctx, groupID, limit, offset)
}

func (s *ExpenseService) GetUserExpenses(ctx context.Context, userID string, limit, offset int64) ([]*models.Expense, error) {
	return s.expenseRepo.GetByUserID(ctx, userID, limit, offset)
}

func (s *ExpenseService) updateBalances(ctx context.Context, expense models.Expense) error {
	// For each user in the split, update their balance
	for _, share := range expense.Split.Details {
		// For each payer, reduce what they owe by what they paid
		for _, pb := range expense.PaidBy {
			if pb.UserID == share.UserID {
				// This user paid some amount and owes some amount
				netChange := pb.Amount - share.Value
				if err := s.balanceRepo.UpdateBalance(ctx, pb.UserID, expense.GroupID, netChange); err != nil {
					return err
				}
			} else {
				// Other users owe the payer
				if share.Value > 0 {
					if err := s.balanceRepo.UpdateBalance(ctx, share.UserID, expense.GroupID, -share.Value); err != nil {
						return err
					}
					if err := s.balanceRepo.UpdateBalance(ctx, pb.UserID, expense.GroupID, share.Value); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
