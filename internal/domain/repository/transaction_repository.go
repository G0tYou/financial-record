package repository

import (
	"financial-record/internal/domain/entities"
)

// TransactionRepository defines the interface for transaction data operations
type TransactionRepository interface {
	// SaveTransaction saves a transaction to the spreadsheet
	SaveTransaction(transaction *entities.Transaction) error
	
	// GetBalance retrieves the current balance for a phone number
	GetBalance(phone string) (float64, error)
	
	// GetTransactions retrieves all transactions for a phone number
	GetTransactions(phone string) ([]*entities.Transaction, error)
}
