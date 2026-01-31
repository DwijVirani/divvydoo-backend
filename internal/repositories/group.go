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
	ErrGroupNotFound        = errors.New("group not found")
	ErrGroupAlreadyExists   = errors.New("group with this ID already exists")
	ErrMemberNotInGroup     = errors.New("member not found in group")
	ErrMemberAlreadyInGroup = errors.New("member already in group")
)

// MemberWithUser contains member info joined with user details
type MemberWithUser struct {
	UserID   string          `bson:"user_id" json:"user_id"`
	Role     models.UserRole `bson:"role" json:"role"`
	JoinedAt time.Time       `bson:"joined_at" json:"joined_at"`
	IsActive bool            `bson:"is_active" json:"is_active"`
	Name     string          `bson:"name" json:"name"`
	Email    string          `bson:"email" json:"email"`
}

type GroupRepository interface {
	Create(ctx context.Context, group *models.Group) (*models.Group, error)
	GetByID(ctx context.Context, groupID string) (*models.Group, error)
	GetByUserID(ctx context.Context, userID string) ([]*models.Group, error)
	Update(ctx context.Context, group *models.Group) (*models.Group, error)
	Delete(ctx context.Context, groupID string) error
	AddMember(ctx context.Context, groupID string, member models.GroupMember) error
	RemoveMember(ctx context.Context, groupID string, userID string) error
	UpdateMemberRole(ctx context.Context, groupID string, userID string, role models.UserRole) error
	IsMember(ctx context.Context, groupID string, userID string) (bool, error)
	GetNonMembers(ctx context.Context, groupID string, userIDs []string) ([]string, error) // Returns user IDs that are not members
	GetMembers(ctx context.Context, groupID string) ([]models.GroupMember, error)
	GetMembersWithDetails(ctx context.Context, groupID string) ([]MemberWithUser, error)
	SetActive(ctx context.Context, groupID string, isActive bool) error
}

type groupRepository struct {
	collection *mongo.Collection
}

func NewGroupRepository(db *mongo.Database) GroupRepository {
	return &groupRepository{
		collection: db.Collection("groups"),
	}
}

func (r *groupRepository) Create(ctx context.Context, group *models.Group) (*models.Group, error) {
	group.CreatedAt = time.Now()
	group.UpdatedAt = group.CreatedAt
	group.IsActive = true

	result, err := r.collection.InsertOne(ctx, group)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrGroupAlreadyExists
		}
		return nil, err
	}

	group.ID = result.InsertedID.(primitive.ObjectID)
	return group, nil
}

func (r *groupRepository) GetByID(ctx context.Context, groupID string) (*models.Group, error) {
	var group models.Group
	filter := bson.M{"group_id": groupID}

	err := r.collection.FindOne(ctx, filter).Decode(&group)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}

	return &group, nil
}

func (r *groupRepository) GetByUserID(ctx context.Context, userID string) ([]*models.Group, error) {
	filter := bson.M{
		"members": bson.M{
			"$elemMatch": bson.M{
				"user_id":   userID,
				"is_active": true,
			},
		},
		"is_active": true,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var groups []*models.Group
	if err := cursor.All(ctx, &groups); err != nil {
		return nil, err
	}

	return groups, nil
}

func (r *groupRepository) Update(ctx context.Context, group *models.Group) (*models.Group, error) {
	group.UpdatedAt = time.Now()

	filter := bson.M{"group_id": group.GroupID}
	update := bson.M{
		"$set": bson.M{
			"name":       group.Name,
			"currency":   group.Currency,
			"updated_at": group.UpdatedAt,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedGroup models.Group

	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedGroup)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}

	return &updatedGroup, nil
}

func (r *groupRepository) Delete(ctx context.Context, groupID string) error {
	filter := bson.M{"group_id": groupID}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrGroupNotFound
	}

	return nil
}

func (r *groupRepository) AddMember(ctx context.Context, groupID string, member models.GroupMember) error {
	// Check if member already exists
	isMember, err := r.IsMember(ctx, groupID, member.UserID)
	if err != nil {
		return err
	}
	if isMember {
		return ErrMemberAlreadyInGroup
	}

	member.JoinedAt = time.Now()
	member.IsActive = true

	filter := bson.M{"group_id": groupID}
	update := bson.M{
		"$push": bson.M{"members": member},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrGroupNotFound
	}

	return nil
}

func (r *groupRepository) RemoveMember(ctx context.Context, groupID string, userID string) error {
	filter := bson.M{
		"group_id":        groupID,
		"members.user_id": userID,
	}
	update := bson.M{
		"$set": bson.M{
			"members.$.is_active": false,
			"updated_at":          time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrMemberNotInGroup
	}

	return nil
}

func (r *groupRepository) UpdateMemberRole(ctx context.Context, groupID string, userID string, role models.UserRole) error {
	filter := bson.M{
		"group_id":        groupID,
		"members.user_id": userID,
	}
	update := bson.M{
		"$set": bson.M{
			"members.$.role": role,
			"updated_at":     time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrMemberNotInGroup
	}

	return nil
}

func (r *groupRepository) IsMember(ctx context.Context, groupID string, userID string) (bool, error) {
	filter := bson.M{
		"group_id": groupID,
		"members": bson.M{
			"$elemMatch": bson.M{
				"user_id":   userID,
				"is_active": true,
			},
		},
	}

	count, err := r.collection.CountDocuments(ctx, filter, options.Count().SetLimit(1))
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetNonMembers returns user IDs from the provided list that are NOT active members of the group
func (r *groupRepository) GetNonMembers(ctx context.Context, groupID string, userIDs []string) ([]string, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	// Use aggregation to find which users are members
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"group_id": groupID}}},
		{{Key: "$unwind", Value: "$members"}},
		{{Key: "$match", Value: bson.M{
			"members.user_id":   bson.M{"$in": userIDs},
			"members.is_active": true,
		}}},
		{{Key: "$project", Value: bson.M{
			"_id":     0,
			"user_id": "$members.user_id",
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Collect member user IDs
	memberIDs := make(map[string]bool)
	for cursor.Next(ctx) {
		var result struct {
			UserID string `bson:"user_id"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		memberIDs[result.UserID] = true
	}

	// Find non-members
	var nonMembers []string
	for _, id := range userIDs {
		if !memberIDs[id] {
			nonMembers = append(nonMembers, id)
		}
	}

	return nonMembers, nil
}

func (r *groupRepository) GetMembers(ctx context.Context, groupID string) ([]models.GroupMember, error) {
	group, err := r.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Return only active members
	var activeMembers []models.GroupMember
	for _, member := range group.Members {
		if member.IsActive {
			activeMembers = append(activeMembers, member)
		}
	}

	return activeMembers, nil
}

func (r *groupRepository) GetMembersWithDetails(ctx context.Context, groupID string) ([]MemberWithUser, error) {
	pipeline := mongo.Pipeline{
		// Match the group by group_id
		{{Key: "$match", Value: bson.M{"group_id": groupID}}},
		// Unwind the members array
		{{Key: "$unwind", Value: "$members"}},
		// Filter only active members
		{{Key: "$match", Value: bson.M{"members.is_active": true}}},
		// Lookup user details from users collection
		{{Key: "$lookup", Value: bson.M{
			"from":         "users",
			"localField":   "members.user_id",
			"foreignField": "user_id",
			"as":           "user_info",
		}}},
		// Unwind the user_info array (will be single element)
		{{Key: "$unwind", Value: bson.M{
			"path":                       "$user_info",
			"preserveNullAndEmptyArrays": true,
		}}},
		// Project the final shape
		{{Key: "$project", Value: bson.M{
			"_id":       0,
			"user_id":   "$members.user_id",
			"role":      "$members.role",
			"joined_at": "$members.joined_at",
			"is_active": "$members.is_active",
			"name":      "$user_info.name",
			"email":     "$user_info.email",
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var members []MemberWithUser
	if err := cursor.All(ctx, &members); err != nil {
		return nil, err
	}

	if members == nil {
		members = []MemberWithUser{}
	}

	return members, nil
}

func (r *groupRepository) SetActive(ctx context.Context, groupID string, isActive bool) error {
	filter := bson.M{"group_id": groupID}
	update := bson.M{
		"$set": bson.M{
			"is_active":  isActive,
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrGroupNotFound
	}

	return nil
}
