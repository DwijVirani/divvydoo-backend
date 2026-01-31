package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Settlement struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	SettlementID  string             `bson:"settlement_id" json:"settlement_id"`
	FromUserID    string             `bson:"from_user_id" json:"from_user_id"`
	ToUserID      string             `bson:"to_user_id" json:"to_user_id"`
	GroupID       *string            `bson:"group_id,omitempty" json:"group_id,omitempty"`
	Amount        float64            `bson:"amount" json:"amount"`
	Currency      string             `bson:"currency" json:"currency"`
	Status        SettlementStatus   `bson:"status" json:"status"`
	Method        SettlementMethod   `bson:"method" json:"method"`
	Description   string             `bson:"description" json:"description"`
	TransactionID *string            `bson:"transaction_id,omitempty" json:"transaction_id,omitempty"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
	CompletedAt   *time.Time         `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
	FailedAt      *time.Time         `bson:"failed_at,omitempty" json:"failed_at,omitempty"`
	FailureReason *string            `bson:"failure_reason,omitempty" json:"failure_reason,omitempty"`
}

type SettlementStatus string

const (
	SettlementPending   SettlementStatus = "pending"
	SettlementCompleted SettlementStatus = "completed"
	SettlementFailed    SettlementStatus = "failed"
	SettlementCancelled SettlementStatus = "cancelled"
)

type SettlementMethod string

const (
	SettlementMethodCash   SettlementMethod = "cash"
	SettlementMethodBank   SettlementMethod = "bank_transfer"
	SettlementMethodUPI    SettlementMethod = "upi"
	SettlementMethodPayPal SettlementMethod = "paypal"
	SettlementMethodVenmo  SettlementMethod = "venmo"
	SettlementMethodOther  SettlementMethod = "other"
)

type SettlementRequest struct {
	FromUserID  string           `json:"from_user_id" binding:"required"`
	ToUserID    string           `json:"to_user_id" binding:"required"`
	GroupID     *string          `json:"group_id,omitempty"`
	Amount      float64          `json:"amount" binding:"required,gt=0"`
	Currency    string           `json:"currency" binding:"required"`
	Method      SettlementMethod `json:"method" binding:"required"`
	Description string           `json:"description,omitempty"`
}

type SettlementResponse struct {
	SettlementID string           `json:"settlement_id"`
	Status       SettlementStatus `json:"status"`
	Message      string           `json:"message,omitempty"`
}
