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
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	GetByID(ctx context.Context, userID string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByPhone(ctx context.Context, phone string) (*models.User, error)
	GetByIDs(ctx context.Context, userIDs []string) ([]*models.User, error)
	Update(ctx context.Context, user *models.User) (*models.User, error)
	Delete(ctx context.Context, userID string) error
	Exists(ctx context.Context, userID string) (bool, error)
	ExistMultiple(ctx context.Context, userIDs []string) ([]string, error) // Returns missing user IDs
}

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) UserRepository {
	return &userRepository{
		collection: db.Collection("users"),
	}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) (*models.User, error) {
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return user, nil
}

func (r *userRepository) GetByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	filter := bson.M{"user_id": userID}

	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	filter := bson.M{"email": email}

	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByPhone(ctx context.Context, phone string) (*models.User, error) {
	var user models.User
	filter := bson.M{"phone": phone}

	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByIDs(ctx context.Context, userIDs []string) ([]*models.User, error) {
	filter := bson.M{"user_id": bson.M{"$in": userIDs}}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*models.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) (*models.User, error) {
	user.UpdatedAt = time.Now()

	filter := bson.M{"user_id": user.UserID}
	update := bson.M{
		"$set": bson.M{
			"name":       user.Name,
			"email":      user.Email,
			"phone":      user.Phone,
			"updated_at": user.UpdatedAt,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedUser models.User

	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedUser)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &updatedUser, nil
}

func (r *userRepository) Delete(ctx context.Context, userID string) error {
	filter := bson.M{"user_id": userID}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *userRepository) Exists(ctx context.Context, userID string) (bool, error) {
	filter := bson.M{"user_id": userID}
	count, err := r.collection.CountDocuments(ctx, filter, options.Count().SetLimit(1))
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistMultiple checks if all provided user IDs exist and returns the IDs that don't exist
func (r *userRepository) ExistMultiple(ctx context.Context, userIDs []string) ([]string, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	// Find all existing users with the given IDs
	filter := bson.M{"user_id": bson.M{"$in": userIDs}}
	cursor, err := r.collection.Find(ctx, filter, options.Find().SetProjection(bson.M{"user_id": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Collect existing user IDs
	existingIDs := make(map[string]bool)
	for cursor.Next(ctx) {
		var result struct {
			UserID string `bson:"user_id"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		existingIDs[result.UserID] = true
	}

	// Find missing IDs
	var missingIDs []string
	for _, id := range userIDs {
		if !existingIDs[id] {
			missingIDs = append(missingIDs, id)
		}
	}

	return missingIDs, nil
}
