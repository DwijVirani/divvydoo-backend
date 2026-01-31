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
	ErrExpenseNotFound = errors.New("expense not found")
)

type ExpenseRepository interface {
	StartSession() (mongo.Session, error)
	CreateExpense(ctx context.Context, expense models.Expense) (*models.Expense, error)
	GetByID(ctx context.Context, expenseID string) (*models.Expense, error)
	GetByGroupID(ctx context.Context, groupID string, limit, offset int64) ([]*models.Expense, error)
	GetByUserID(ctx context.Context, userID string, limit, offset int64) ([]*models.Expense, error)
	Update(ctx context.Context, expense *models.Expense) (*models.Expense, error)
	SoftDelete(ctx context.Context, expenseID string) error
	HardDelete(ctx context.Context, expenseID string) error
	CountByGroupID(ctx context.Context, groupID string) (int64, error)
	CountByUserID(ctx context.Context, userID string) (int64, error)
}

type expenseRepository struct {
	collection *mongo.Collection
	client     *mongo.Client
}

func NewExpenseRepository(db *mongo.Database) ExpenseRepository {
	return &expenseRepository{
		collection: db.Collection("expenses"),
		client:     db.Client(),
	}
}

func (r *expenseRepository) StartSession() (mongo.Session, error) {
	return r.client.StartSession()
}

func (r *expenseRepository) CreateExpense(ctx context.Context, expense models.Expense) (*models.Expense, error) {
	expense.CreatedAt = time.Now()
	expense.UpdatedAt = expense.CreatedAt
	expense.IsDeleted = false

	result, err := r.collection.InsertOne(ctx, expense)
	if err != nil {
		return nil, err
	}

	expense.ID = result.InsertedID.(primitive.ObjectID)
	return &expense, nil
}

func (r *expenseRepository) GetByID(ctx context.Context, expenseID string) (*models.Expense, error) {
	var expense models.Expense
	filter := bson.M{
		"expense_id": expenseID,
		"is_deleted": false,
	}

	err := r.collection.FindOne(ctx, filter).Decode(&expense)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrExpenseNotFound
		}
		return nil, err
	}

	return &expense, nil
}

func (r *expenseRepository) GetByGroupID(ctx context.Context, groupID string, limit, offset int64) ([]*models.Expense, error) {
	filter := bson.M{
		"group_id":   groupID,
		"is_deleted": false,
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

	var expenses []*models.Expense
	if err := cursor.All(ctx, &expenses); err != nil {
		return nil, err
	}

	return expenses, nil
}

func (r *expenseRepository) GetByUserID(ctx context.Context, userID string, limit, offset int64) ([]*models.Expense, error) {
	// User is either the creator, a payer, or in the split
	filter := bson.M{
		"is_deleted": false,
		"$or": []bson.M{
			{"creator_id": userID},
			{"paid_by.user_id": userID},
			{"split.details.user_id": userID},
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

	var expenses []*models.Expense
	if err := cursor.All(ctx, &expenses); err != nil {
		return nil, err
	}

	return expenses, nil
}

func (r *expenseRepository) Update(ctx context.Context, expense *models.Expense) (*models.Expense, error) {
	expense.UpdatedAt = time.Now()

	filter := bson.M{
		"expense_id": expense.ExpenseID,
		"is_deleted": false,
	}

	update := bson.M{
		"$set": bson.M{
			"title":      expense.Title,
			"amount":     expense.Amount,
			"currency":   expense.Currency,
			"paid_by":    expense.PaidBy,
			"split":      expense.Split,
			"updated_at": expense.UpdatedAt,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedExpense models.Expense

	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedExpense)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrExpenseNotFound
		}
		return nil, err
	}

	return &updatedExpense, nil
}

func (r *expenseRepository) SoftDelete(ctx context.Context, expenseID string) error {
	filter := bson.M{"expense_id": expenseID}
	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrExpenseNotFound
	}

	return nil
}

func (r *expenseRepository) HardDelete(ctx context.Context, expenseID string) error {
	filter := bson.M{"expense_id": expenseID}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrExpenseNotFound
	}

	return nil
}

func (r *expenseRepository) CountByGroupID(ctx context.Context, groupID string) (int64, error) {
	filter := bson.M{
		"group_id":   groupID,
		"is_deleted": false,
	}

	return r.collection.CountDocuments(ctx, filter)
}

func (r *expenseRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	filter := bson.M{
		"is_deleted": false,
		"$or": []bson.M{
			{"creator_id": userID},
			{"paid_by.user_id": userID},
			{"split.details.user_id": userID},
		},
	}

	return r.collection.CountDocuments(ctx, filter)
}
