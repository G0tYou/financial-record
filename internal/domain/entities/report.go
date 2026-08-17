package entities

import "time"

// MonthlyReport represents a monthly financial summary report
type MonthlyReport struct {
	Phone         string
	Month         string
	Year          int
	TotalIncome   float64
	TotalExpense  float64
	NetBalance    float64
	CategoryBreakdown []*CategoryBreakdown
	GeneratedAt   time.Time
}

// CategoryBreakdown represents spending/income by category
type CategoryBreakdown struct {
	Category     string
	TotalAmount  float64
	TransactionCount int
	Percentage   float64
}

// NewMonthlyReport creates a new monthly report
func NewMonthlyReport(phone string, month string, year int) *MonthlyReport {
	return &MonthlyReport{
		Phone:       phone,
		Month:       month,
		Year:        year,
		GeneratedAt: time.Now(),
	}
}