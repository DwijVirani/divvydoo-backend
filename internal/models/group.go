package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Group struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GroupID   string             `bson:"group_id" json:"group_id"`
	Name      string             `bson:"name" json:"name"`
	Members   []GroupMember      `bson:"members" json:"members"`
	Currency  string             `bson:"currency" json:"currency"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
	IsActive  bool               `bson:"is_active" json:"is_active"`
}

type GroupMember struct {
	UserID   string    `bson:"user_id" json:"user_id"`
	Role     UserRole  `bson:"role" json:"role"`
	JoinedAt time.Time `bson:"joined_at" json:"joined_at"`
	IsActive bool      `bson:"is_active" json:"is_active"`
}

type GroupInvitation struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	InvitationID string             `bson:"invitation_id" json:"invitation_id"`
	GroupID      string             `bson:"group_id" json:"group_id"`
	InviterID    string             `bson:"inviter_id" json:"inviter_id"`
	InviteeEmail string             `bson:"invitee_email" json:"invitee_email"`
	Token        string             `bson:"token" json:"-"`
	Status       InvitationStatus   `bson:"status" json:"status"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	ExpiresAt    time.Time          `bson:"expires_at" json:"expires_at"`
}

type InvitationStatus string

const (
	InvitationPending  InvitationStatus = "pending"
	InvitationAccepted InvitationStatus = "accepted"
	InvitationRejected InvitationStatus = "rejected"
	InvitationExpired  InvitationStatus = "expired"
)
