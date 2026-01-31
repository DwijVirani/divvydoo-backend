package services

import (
	"context"
	"errors"
	"time"

	"divvydoo/backend/internal/models"
	"divvydoo/backend/internal/repositories"

	"github.com/google/uuid"
)

var (
	ErrGroupNotFound       = errors.New("group not found")
	ErrNotGroupMember      = errors.New("user is not a member of this group")
	ErrNotGroupAdmin       = errors.New("user is not an admin of this group")
	ErrMemberAlreadyExists = errors.New("user is already a member of this group")
)

type GroupService struct {
	groupRepo repositories.GroupRepository
	userRepo  repositories.UserRepository
}

func NewGroupService(groupRepo repositories.GroupRepository, userRepo repositories.UserRepository) *GroupService {
	return &GroupService{
		groupRepo: groupRepo,
		userRepo:  userRepo,
	}
}

type CreateGroupRequest struct {
	Name     string `json:"name" binding:"required"`
	Currency string `json:"currency" binding:"required"`
}

type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role,omitempty"`
}


func (s *GroupService) CreateGroup(ctx context.Context, creatorID string, req CreateGroupRequest) (*models.Group, error) {
	// Verify creator exists
	exists, err := s.userRepo.Exists(ctx, creatorID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	group := &models.Group{
		GroupID:  uuid.New().String(),
		Name:     req.Name,
		Currency: req.Currency,
		Members: []models.GroupMember{
			{
				UserID:   creatorID,
				Role:     models.RoleAdmin,
				JoinedAt: time.Now(),
				IsActive: true,
			},
		},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return s.groupRepo.Create(ctx, group)
}

func (s *GroupService) GetGroup(ctx context.Context, groupID string, userID string) (*models.Group, error) {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, repositories.ErrGroupNotFound) {
			return nil, ErrGroupNotFound
		}
		return nil, err
	}

	// Check if user is a member
	isMember, err := s.groupRepo.IsMember(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotGroupMember
	}

	return group, nil
}

func (s *GroupService) GetUserGroups(ctx context.Context, userID string) ([]*models.Group, error) {
	return s.groupRepo.GetByUserID(ctx, userID)
}

func (s *GroupService) UpdateGroup(ctx context.Context, groupID string, userID string, req CreateGroupRequest) (*models.Group, error) {
	// Check if user is an admin
	isAdmin, err := s.isGroupAdmin(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, ErrNotGroupAdmin
	}

	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	group.Name = req.Name
	group.Currency = req.Currency

	return s.groupRepo.Update(ctx, group)
}

func (s *GroupService) AddMember(ctx context.Context, groupID string, adminUserID string, req AddMemberRequest) error {
	// Check if requester is an admin
	isAdmin, err := s.isGroupAdmin(ctx, groupID, adminUserID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return ErrNotGroupAdmin
	}

	// Verify new member exists
	exists, err := s.userRepo.Exists(ctx, req.UserID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrUserNotFound
	}

	role := models.RoleMember
	if req.Role == string(models.RoleAdmin) {
		role = models.RoleAdmin
	}

	member := models.GroupMember{
		UserID:   req.UserID,
		Role:     role,
		JoinedAt: time.Now(),
		IsActive: true,
	}

	err = s.groupRepo.AddMember(ctx, groupID, member)
	if err != nil {
		if errors.Is(err, repositories.ErrMemberAlreadyInGroup) {
			return ErrMemberAlreadyExists
		}
		return err
	}

	return nil
}

func (s *GroupService) RemoveMember(ctx context.Context, groupID string, adminUserID string, memberUserID string) error {
	// Check if requester is an admin
	isAdmin, err := s.isGroupAdmin(ctx, groupID, adminUserID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return ErrNotGroupAdmin
	}

	return s.groupRepo.RemoveMember(ctx, groupID, memberUserID)
}

func (s *GroupService) LeaveGroup(ctx context.Context, groupID string, userID string) error {
	return s.groupRepo.RemoveMember(ctx, groupID, userID)
}

func (s *GroupService) GetMembers(ctx context.Context, groupID string, userID string) ([]repositories.MemberWithUser, error) {
	// Check if user is a member
	isMember, err := s.groupRepo.IsMember(ctx, groupID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotGroupMember
	}

	return s.groupRepo.GetMembersWithDetails(ctx, groupID)
}

func (s *GroupService) isGroupAdmin(ctx context.Context, groupID string, userID string) (bool, error) {
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, repositories.ErrGroupNotFound) {
			return false, ErrGroupNotFound
		}
		return false, err
	}

	for _, member := range group.Members {
		if member.UserID == userID && member.IsActive && member.Role == models.RoleAdmin {
			return true, nil
		}
	}

	return false, nil
}
