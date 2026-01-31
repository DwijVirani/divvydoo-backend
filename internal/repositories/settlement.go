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
	ErrSettlementNotFound = errors.New("settlement not found")
)

type SettlementRepository interface {
	Create(ctx context.Context, settlement *models.Settlement) (*models.Settlement, error)
	GetByID(ctx context.Context, settlementID string) (*models.Settlement, error)
	GetByUserID(ctx context.Context, userID string, limit, offset int64) ([]*models.Settlement, error)
	GetByGroupID(ctx context.Context, groupID string, limit, offset int64) ([]*models.Settlement, error)
	GetBetweenUsers(ctx context.Context, userID1, userID2 string, limit, offset int64) ([]*models.Settlement, error)
	UpdateStatus(ctx context.Context, settlementID string, status models.SettlementStatus) error
	MarkCompleted(ctx context.Context, settlementID string, transactionID *string) error
	MarkFailed(ctx context.Context, settlementID string, reason string) error
	MarkCancelled(ctx context.Context, settlementID string) error
	GetPendingSettlements(ctx context.Context, userID string) ([]*models.Settlement, error)
	CountByUserID(ctx context.Context, userID string) (int64, error)
	StartSession() (mongo.Session, error)
}

type settlementRepository struct {
	collection *mongo.Collection
	client     *mongo.Client
}

func NewSettlementRepository(db *mongo.Database) SettlementRepository {
	return &settlementRepository{
		collection: db.Collection("settlements"),
		client:     db.Client(),
	}
}

func (r *settlementRepository) StartSession() (mongo.Session, error) {
	return r.client.StartSession()
}

func (r *settlementRepository) Create(ctx context.Context, settlement *models.Settlement) (*models.Settlement, error) {
	settlement.CreatedAt = time.Now()
	settlement.UpdatedAt = settlement.CreatedAt
	settlement.Status = models.SettlementPending

	result, err := r.collection.InsertOne(ctx, settlement)
	if err != nil {
		return nil, err
	}

	settlement.ID = result.InsertedID.(primitive.ObjectID)
	return settlement, nil
}

func (r *settlementRepository) GetByID(ctx context.Context, settlementID string) (*models.Settlement, error) {
	var settlement models.Settlement
	filter := bson.M{"settlement_id": settlementID}

	err := r.collection.FindOne(ctx, filter).Decode(&settlement)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrSettlementNotFound
		}
		return nil, err
	}

	return &settlement, nil
}

func (r *settlementRepository) GetByUserID(ctx context.Context, userID string, limit, offset int64) ([]*models.Settlement, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"from_user_id": userID},
			{"to_user_id": userID},
		},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(offset)

	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var settlements []*models.Settlement
	if err := cursor.All(ctx, &settlements); err != nil {
		return nil, err
	}

	return settlements, nil
}

func (r *settlementRepository) GetByGroupID(ctx context.Context, groupID string, limit, offset int64) ([]*models.Settlement, error) {
	filter := bson.M{"group_id": groupID}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(offset)

	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var settlements []*models.Settlement
	if err := cursor.All(ctx, &settlements); err != nil {
		return nil, err
	}

	return settlements, nil
}

func (r *settlementRepository) GetBetweenUsers(ctx context.Context, userID1, userID2 string, limit, offset int64) ([]*models.Settlement, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"from_user_id": userID1, "to_user_id": userID2},
			{"from_user_id": userID2, "to_user_id": userID1},
		},
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(offset)

	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var settlements []*models.Settlement
	if err := cursor.All(ctx, &settlements); err != nil {
		return nil, err
	}

	return settlements, nil
}

func (r *settlementRepository) UpdateStatus(ctx context.Context, settlementID string, status models.SettlementStatus) error {
	filter := bson.M{"settlement_id": settlementID}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrSettlementNotFound
	}

	return nil
}

func (r *settlementRepository) MarkCompleted(ctx context.Context, settlementID string, transactionID *string) error {
	now := time.Now()
	filter := bson.M{"settlement_id": settlementID}
	update := bson.M{
		"$set": bson.M{
			"status":         models.SettlementCompleted,
			"completed_at":   now,
			"updated_at":     now,
			"transaction_id": transactionID,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrSettlementNotFound
	}

	return nil
}

func (r *settlementRepository) MarkFailed(ctx context.Context, settlementID string, reason string) error {
	now := time.Now()
	filter := bson.M{"settlement_id": settlementID}
	update := bson.M{
		"$set": bson.M{
			"status":         models.SettlementFailed,
			"failed_at":      now,
			"failure_reason": reason,
			"updated_at":     now,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrSettlementNotFound
	}

	return nil
}

func (r *settlementRepository) MarkCancelled(ctx context.Context, settlementID string) error {
	filter := bson.M{"settlement_id": settlementID}
	update := bson.M{
		"$set": bson.M{
			"status":     models.SettlementCancelled,
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrSettlementNotFound
	}

	return nil
}

func (r *settlementRepository) GetPendingSettlements(ctx context.Context, userID string) ([]*models.Settlement, error) {
	filter := bson.M{
		"status": models.SettlementPending,
		"$or": []bson.M{
			{"from_user_id": userID},
			{"to_user_id": userID},
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var settlements []*models.Settlement
	if err := cursor.All(ctx, &settlements); err != nil {
		return nil, err
	}

	return settlements, nil
}

func (r *settlementRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"from_user_id": userID},
			{"to_user_id": userID},
		},
	}

	return r.collection.CountDocuments(ctx, filter)
}
