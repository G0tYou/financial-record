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
func (uc *TransactionUseCase) ProcessTransaction(phone string, action entities.Action, amount float64, notes string) (*entities.Transaction, error) {
	// Validate input
	if phone == "" {
		return nil, fmt.Errorf("phone number is required")
	}

	if amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	if action != entities.ActionAdd && action != entities.ActionSubtract {
		return nil, fmt.Errorf("invalid action: %s", action)
	}

	// Get current balance
	currentBalance, err := uc.transactionRepo.GetBalance(phone)
	if err != nil {
		return nil, fmt.Errorf("unable to get current balance: %w", err)
	}

	// Calculate new balance
	var newBalance float64
	switch action {
	case entities.ActionAdd:
		newBalance = currentBalance + amount
	case entities.ActionSubtract:
		newBalance = currentBalance - amount
		if newBalance < 0 {
			return nil, fmt.Errorf("insufficient balance: current balance is %.2f", currentBalance)
		}
	}

	// Create transaction
	transaction := entities.NewTransaction(phone, action, amount, newBalance, notes)

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
