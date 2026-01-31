package services

import (
	"context"
	"errors"
	"time"

	"divvydoo/backend/internal/models"
	"divvydoo/backend/internal/repositories"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
)

type UserService struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Phone    string `json:"phone,omitempty"`
}

type UpdateUserRequest struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*models.User, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		UserID:    uuid.New().String(),
		Name:      req.Name,
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return s.userRepo.Create(ctx, user)
}

func (s *UserService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	user, err := s.userRepo.GetByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) LookupUser(ctx context.Context, query string) (*models.User, error) {
	// Try email first
	user, err := s.userRepo.GetByEmail(ctx, query)
	if err == nil {
		return user, nil
	}
	if !errors.Is(err, repositories.ErrUserNotFound) {
		return nil, err
	}

	// Try phone number
	user, err = s.userRepo.GetByPhone(ctx, query)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, userID string, req UpdateUserRequest) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}

	return s.userRepo.Update(ctx, user)
}

func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	return s.userRepo.Delete(ctx, userID)
}

func (s *UserService) ValidateCredentials(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}
