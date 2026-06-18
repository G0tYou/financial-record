package usecase

import (
	"fmt"

	"financial-record/internal/domain/entities"
	"financial-record/internal/domain/repository"
)

// TransactionUseCase handles business logic for transactions
type TransactionUseCase struct {
	transactionRepo repository.TransactionRepository
}

// NewTransactionUseCase creates a new transaction use case
func NewTransactionUseCase(transactionRepo repository.TransactionRepository) *TransactionUseCase {
	return &TransactionUseCase{
		transactionRepo: transactionRepo,
	}
}

// ProcessTransaction processes a transaction request
func (uc *TransactionUseCase) ProcessTransaction(phone string, action string, amount float64, description string) (*entities.Transaction, error) {
	// Validate input
	if phone == "" {
		return nil, fmt.Errorf("phone number is required")
	}

	if amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	// Create transaction
	transaction := entities.NewTransaction(phone, action, amount, 0, description)

	// Save transaction
	if err := uc.transactionRepo.SaveTransaction(transaction); err != nil {
		return nil, fmt.Errorf("unable to save transaction: %w", err)
	}

	return transaction, nil
}

// GetBalance retrieves the current balance for a phone number
func (uc *TransactionUseCase) GetBalance(phone string) (float64, error) {
	if phone == "" {
		return 0, fmt.Errorf("phone number is required")
	}

	return uc.transactionRepo.GetBalance(phone)
}

// GetTransactionHistory retrieves transaction history for a phone number
func (uc *TransactionUseCase) GetTransactionHistory(phone string) ([]*entities.Transaction, error) {
	if phone == "" {
		return nil, fmt.Errorf("phone number is required")
	}

	return uc.transactionRepo.GetTransactions(phone)
}
