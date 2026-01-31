package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Balance struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    string             `bson:"user_id" json:"user_id"`
	GroupID   *string            `bson:"group_id,omitempty" json:"group_id,omitempty"` // null for personal balances
	Balance   float64            `bson:"balance" json:"balance"`
	Currency  string             `bson:"currency" json:"currency"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
	Version   int                `bson:"version" json:"version"` // For optimistic concurrency
}

type BalanceHistory struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      string             `bson:"user_id" json:"user_id"`
	GroupID     *string            `bson:"group_id,omitempty" json:"group_id,omitempty"`
	Amount      float64            `bson:"amount" json:"amount"`
	Currency    string             `bson:"currency" json:"currency"`
	Type        BalanceChangeType  `bson:"type" json:"type"`
	ReferenceID string             `bson:"reference_id" json:"reference_id"` // expense_id or settlement_id
	Description string             `bson:"description" json:"description"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}

type BalanceChangeType string

const (
	BalanceChangeExpense    BalanceChangeType = "expense"
	BalanceChangeSettlement BalanceChangeType = "settlement"
	BalanceChangeAdjustment BalanceChangeType = "adjustment"
	BalanceChangeCorrection BalanceChangeType = "correction"
)

type UserBalanceSummary struct {
	UserID        string         `json:"user_id"`
	TotalBalance  float64        `json:"total_balance"`
	GroupBalances []GroupBalance `json:"group_balances"`
	PeerBalances  []PeerBalance  `json:"peer_balances"`
	Currency      string         `json:"currency"`
	LastUpdated   time.Time      `json:"last_updated"`
}

type GroupBalance struct {
	GroupID   string  `json:"group_id"`
	GroupName string  `json:"group_name"`
	Balance   float64 `json:"balance"`
}

type PeerBalance struct {
	PeerID   string  `json:"peer_id"`
	PeerName string  `json:"peer_name"`
	Balance  float64 `json:"balance"` // Positive: peer owes you, Negative: you owe peer
}
