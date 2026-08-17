package repository

import (
	"financial-record/internal/domain/entities"
)

// ReportRepository defines the interface for report data operations
type ReportRepository interface {
	// GenerateMonthlyReport generates a monthly summary report
	GenerateMonthlyReport(phone string, month string, year int) (*entities.MonthlyReport, error)
	
	// GetAllTransactionsForMonth retrieves all transactions for a specific month
	GetAllTransactionsForMonth(phone string, month string, year int) ([]*entities.Transaction, error)
}