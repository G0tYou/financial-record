package entities

import (
	"financial-record/internal/utils"
	"time"
)

// Action represents the type of transaction
type Action string

const (
	ActionAdd       Action = "add"
	ActionSubtract  Action = "subtract"
	ActionCarryOver Action = "carry-over"
)

// Transaction represents a financial transaction
type Transaction struct {
	ID        string
	Phone     string
	Action    string
	Amount    float64
	Balance   float64
	Category  string
	Timestamp time.Time
	Notes     string
}

// NewTransaction creates a new transaction
func NewTransaction(phone string, action string, amount float64, balance float64, category string, notes string) *Transaction {
	jakartaLoc, _ := time.LoadLocation("Asia/Jakarta")
	return &Transaction{
		ID:        utils.GenerateID(),
		Phone:     phone,
		Action:    action,
		Amount:    amount,
		Balance:   balance,
		Category:  category,
		Timestamp: time.Now().In(jakartaLoc),
		Notes:     notes,
	}
}
