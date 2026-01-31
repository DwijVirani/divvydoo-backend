package repositories

import (
	"context"
	"errors"
	"time"

	"divvydoo/backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrBalanceNotFound       = errors.New("balance not found")
	ErrOptimisticLockFailure = errors.New("optimistic lock failure: balance was modified")
)

type BalanceRepository interface {
	Create(ctx context.Context, balance *models.Balance) (*models.Balance, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Balance, error)
	GetByGroupID(ctx context.Context, groupID string) ([]*models.Balance, error)
	GetByUserAndGroup(ctx context.Context, userID string, groupID *string) (*models.Balance, error)
	UpdateBalance(ctx context.Context, userID string, groupID *string, amount float64) error
	UpdateBalanceWithVersion(ctx context.Context, balance *models.Balance) error
	GetUserBalanceSummary(ctx context.Context, userID string) (*models.UserBalanceSummary, error)
	CreateBalanceHistory(ctx context.Context, history *models.BalanceHistory) error
	GetBalanceHistory(ctx context.Context, userID string, groupID *string, limit, offset int64) ([]*models.BalanceHistory, error)
}

type balanceRepository struct {
	balanceCollection *mongo.Collection
	historyCollection *mongo.Collection
}

func NewBalanceRepository(db *mongo.Database) BalanceRepository {
	return &balanceRepository{
		balanceCollection: db.Collection("balances"),
		historyCollection: db.Collection("balance_history"),
	}
}

func (r *balanceRepository) Create(ctx context.Context, balance *models.Balance) (*models.Balance, error) {
	balance.UpdatedAt = time.Now()
	balance.Version = 1

	result, err := r.balanceCollection.InsertOne(ctx, balance)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			// Balance already exists, return it
			return r.GetByUserAndGroup(ctx, balance.UserID, balance.GroupID)
		}
		return nil, err
	}

	balance.ID = result.InsertedID.(primitive.ObjectID)
	return balance, nil
}

func (r *balanceRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Balance, error) {
	filter := bson.M{"user_id": userID}

	cursor, err := r.balanceCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var balances []*models.Balance
	if err := cursor.All(ctx, &balances); err != nil {
		return nil, err
	}

	return balances, nil
}

func (r *balanceRepository) GetByGroupID(ctx context.Context, groupID string) ([]*models.Balance, error) {
	filter := bson.M{"group_id": groupID}

	cursor, err := r.balanceCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var balances []*models.Balance
	if err := cursor.All(ctx, &balances); err != nil {
		return nil, err
	}

	return balances, nil
}

func (r *balanceRepository) GetByUserAndGroup(ctx context.Context, userID string, groupID *string) (*models.Balance, error) {
	filter := bson.M{"user_id": userID}
	if groupID != nil {
		filter["group_id"] = *groupID
	} else {
		filter["group_id"] = nil
	}

	var balance models.Balance
	err := r.balanceCollection.FindOne(ctx, filter).Decode(&balance)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrBalanceNotFound
		}
		return nil, err
	}

	return &balance, nil
}

func (r *balanceRepository) UpdateBalance(ctx context.Context, userID string, groupID *string, amount float64) error {
	filter := bson.M{"user_id": userID}
	if groupID != nil {
		filter["group_id"] = *groupID
	} else {
		filter["group_id"] = nil
	}

	update := bson.M{
		"$inc": bson.M{
			"balance": amount,
			"version": 1,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
		"$setOnInsert": bson.M{
			"user_id":  userID,
			"group_id": groupID,
			"currency": "USD", // Default currency
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.balanceCollection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *balanceRepository) UpdateBalanceWithVersion(ctx context.Context, balance *models.Balance) error {
	currentVersion := balance.Version
	balance.Version++
	balance.UpdatedAt = time.Now()

	filter := bson.M{
		"user_id": balance.UserID,
		"version": currentVersion,
	}
	if balance.GroupID != nil {
		filter["group_id"] = *balance.GroupID
	} else {
		filter["group_id"] = nil
	}

	update := bson.M{
		"$set": bson.M{
			"balance":    balance.Balance,
			"version":    balance.Version,
			"updated_at": balance.UpdatedAt,
		},
	}

	result, err := r.balanceCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrOptimisticLockFailure
	}

	return nil
}

func (r *balanceRepository) GetUserBalanceSummary(ctx context.Context, userID string) (*models.UserBalanceSummary, error) {
	balances, err := r.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	summary := &models.UserBalanceSummary{
		UserID:        userID,
		TotalBalance:  0,
		GroupBalances: []models.GroupBalance{},
		PeerBalances:  []models.PeerBalance{},
		Currency:      "USD",
		LastUpdated:   time.Now(),
	}

	for _, balance := range balances {
		summary.TotalBalance += balance.Balance
		if balance.GroupID != nil {
			summary.GroupBalances = append(summary.GroupBalances, models.GroupBalance{
				GroupID: *balance.GroupID,
				Balance: balance.Balance,
			})
		}
		if summary.Currency == "" {
			summary.Currency = balance.Currency
		}
		if balance.UpdatedAt.After(summary.LastUpdated) {
			summary.LastUpdated = balance.UpdatedAt
		}
	}

	return summary, nil
}

func (r *balanceRepository) CreateBalanceHistory(ctx context.Context, history *models.BalanceHistory) error {
	history.CreatedAt = time.Now()

	result, err := r.historyCollection.InsertOne(ctx, history)
	if err != nil {
		return err
	}

	history.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *balanceRepository) GetBalanceHistory(ctx context.Context, userID string, groupID *string, limit, offset int64) ([]*models.BalanceHistory, error) {
	filter := bson.M{"user_id": userID}
	if groupID != nil {
		filter["group_id"] = *groupID
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(offset)

	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := r.historyCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var history []*models.BalanceHistory
	if err := cursor.All(ctx, &history); err != nil {
		return nil, err
	}

	return history, nil
}
