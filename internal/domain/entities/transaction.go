package entities

import "time"

// Action represents the type of transaction
type Action string

const (
	ActionAdd      Action = "add"
	ActionSubtract Action = "subtract"
)

// Transaction represents a financial transaction
type Transaction struct {
	ID        string
	Phone     string
	Action    Action
	Amount    float64
	Balance   float64
	Timestamp time.Time
	Notes     string
}

// NewTransaction creates a new transaction
func NewTransaction(phone string, action Action, amount float64, balance float64, notes string) *Transaction {
	return &Transaction{
		ID:        generateID(),
		Phone:     phone,
		Action:    action,
		Amount:    amount,
		Balance:   balance,
		Timestamp: time.Now().UTC(),
		Notes:     notes,
	}
}

// generateID generates a unique transaction ID
func generateID() string {
	return time.Now().Format("20060102150405")
}
