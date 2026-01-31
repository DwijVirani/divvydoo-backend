package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRole string

const (
	RoleMember UserRole = "member"
	RoleAdmin  UserRole = "admin"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    string             `bson:"user_id" json:"user_id"`
	Name      string             `bson:"name" json:"name"`
	Email     string             `bson:"email" json:"email"`
	Phone     string             `bson:"phone,omitempty" json:"phone,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
	Password  string             `bson:"password,omitempty" json:"-"`
}

type UserPreferences struct {
	DefaultCurrency string `bson:"default_currency,omitempty" json:"default_currency,omitempty"`
}
