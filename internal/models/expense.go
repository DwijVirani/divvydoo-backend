package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SplitType string

const (
	SplitEqual      SplitType = "equal"
	SplitExact      SplitType = "exact"
	SplitPercentage SplitType = "percentage"
	SplitShares     SplitType = "shares"
)

type Expense struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ExpenseID string             `bson:"expense_id" json:"expense_id"`
	GroupID   *string            `bson:"group_id,omitempty" json:"group_id,omitempty"`
	CreatorID string             `bson:"creator_id" json:"creator_id"`
	Title     string             `bson:"title" json:"title"`
	Amount    float64            `bson:"amount" json:"amount"`
	Currency  string             `bson:"currency" json:"currency"`
	PaidBy    []PaidBy           `bson:"paid_by" json:"paid_by"`
	Split     SplitDetail        `bson:"split" json:"split"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
	IsDeleted bool               `bson:"is_deleted" json:"is_deleted"`
}

type PaidBy struct {
	UserID string  `bson:"user_id" json:"user_id"`
	Amount float64 `bson:"amount" json:"amount"`
}

type SplitDetail struct {
	Type    SplitType    `bson:"type" json:"type"`
	Details []SplitShare `bson:"details" json:"details"`
}

type SplitShare struct {
	UserID string  `bson:"user_id" json:"user_id"`
	Value  float64 `bson:"value" json:"value"`
}
