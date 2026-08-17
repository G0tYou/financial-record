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
func (uc *TransactionUseCase) ProcessTransaction(phone string, action string, amount float64, category string, description string) (*entities.Transaction, error) {
	// Validate input
	if phone == "" {
		return nil, fmt.Errorf("phone number is required")
	}

	// Allow 0 amount for carry-over transactions
	if amount <= 0 && action != "carry-over" {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	// Get current balance
	currentBalance, err := uc.transactionRepo.GetBalance(phone)
	if err != nil {
		return nil, fmt.Errorf("unable to get current balance: %w", err)
	}

	// If no balance in current month, try to get from previous month
	if currentBalance == 0 {
		previousMonthBalance, err := uc.transactionRepo.GetBalanceFromPreviousMonth(phone)
		if err == nil && previousMonthBalance > 0 {
			currentBalance = previousMonthBalance
		}
	}

	// Calculate new balance based on action
	var newBalance float64
	if action == "add" {
		newBalance = currentBalance + amount
	} else if action == "subtract" {
		newBalance = currentBalance - amount
	} else if action == "carry-over" {
		newBalance = amount // For carry-over, amount represents the carried balance
	} else {
		return nil, fmt.Errorf("invalid action: %s", action)
	}

	// Create transaction with calculated balance
	transaction := entities.NewTransaction(phone, action, amount, newBalance, category, description)

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

	currentBalance, err := uc.transactionRepo.GetBalance(phone)
	if err != nil {
		return 0, err
	}

	// If no balance in current month, try to get from previous month and create carry-over transaction
	if currentBalance == 0 {
		previousMonthBalance, err := uc.transactionRepo.GetBalanceFromPreviousMonth(phone)
		if err == nil && previousMonthBalance > 0 {
			// Create a carry-over transaction to initialize the new month
			carryOverTransaction := entities.NewTransaction(phone, "carry-over", 0, previousMonthBalance, "", "Balance carried over from previous month")
			if err := uc.transactionRepo.SaveTransaction(carryOverTransaction); err != nil {
				// If save fails, still return the previous month balance
				return previousMonthBalance, nil
			}
			return previousMonthBalance, nil
		}
	}

	return currentBalance, nil
}

// GetTransactionHistory retrieves transaction history for a phone number
func (uc *TransactionUseCase) GetTransactionHistory(phone string) ([]*entities.Transaction, error) {
	if phone == "" {
		return nil, fmt.Errorf("phone number is required")
	}

	return uc.transactionRepo.GetTransactions(phone)
}
