package usecase

import (
	"fmt"
	"strings"
	"time"

	"financial-record/internal/domain/entities"
	"financial-record/internal/domain/repository"
	"financial-record/internal/utils"
)

// ReportUseCase handles business logic for reports
type ReportUseCase struct {
	transactionRepo repository.TransactionRepository
}

// NewReportUseCase creates a new report use case
func NewReportUseCase(transactionRepo repository.TransactionRepository) *ReportUseCase {
	return &ReportUseCase{
		transactionRepo: transactionRepo,
	}
}

// GenerateMonthlyReport generates a monthly summary report
func (uc *ReportUseCase) GenerateMonthlyReport(phone string, month string, year int) (*entities.MonthlyReport, error) {
	fmt.Printf("Generating report for phone: %s, month: %s, year: %d\n", phone, month, year)

	// Normalize month name (case insensitive)
	month = strings.ToLower(month)

	// Get all transactions for the specified month
	transactions, err := uc.transactionRepo.GetAllTransactionsForMonth(phone, month, year)
	if err != nil {
		return nil, fmt.Errorf("unable to get transactions: %w", err)
	}

	fmt.Printf("Found %d transactions for report generation\n", len(transactions))

	// Create report
	report := entities.NewMonthlyReport(phone, month, year)

	// Calculate totals and category breakdown
	categoryTotals := make(map[string]*entities.CategoryBreakdown)
	var totalIncome, totalExpense float64

	for _, transaction := range transactions {
		amount := transaction.Amount

		// Separate income and expenses
		if transaction.Action == "add" {
			totalIncome += amount
		} else if transaction.Action == "subtract" {
			totalExpense += amount
		}

		// Group by category
		category := transaction.Category
		if category == "" {
			category = "Other"
		}

		if _, exists := categoryTotals[category]; !exists {
			categoryTotals[category] = &entities.CategoryBreakdown{
				Category: category,
			}
		}

		categoryTotals[category].TotalAmount += amount
		categoryTotals[category].TransactionCount++
	}

	// Calculate percentages and build breakdown
	totalTransactions := float64(len(transactions))
	for _, breakdown := range categoryTotals {
		if totalTransactions > 0 {
			breakdown.Percentage = (breakdown.TotalAmount / (totalIncome + totalExpense)) * 100
		}
		report.CategoryBreakdown = append(report.CategoryBreakdown, breakdown)
	}

	report.TotalIncome = totalIncome
	report.TotalExpense = totalExpense
	report.NetBalance = totalIncome - totalExpense

	fmt.Printf("Report generated: Income=%.2f, Expense=%.2f, Net=%.2f\n", totalIncome, totalExpense, report.NetBalance)

	return report, nil
}

// GenerateCurrentMonthReport generates a report for the current month
func (uc *ReportUseCase) GenerateCurrentMonthReport(phone string) (*entities.MonthlyReport, error) {
	now := time.Now()
	month := now.Month().String()
	year := now.Year()

	return uc.GenerateMonthlyReport(phone, month, year)
}

// FormatReportAsWhatsAppMessage formats the monthly report as a WhatsApp message
func (uc *ReportUseCase) FormatReportAsWhatsAppMessage(report *entities.MonthlyReport) string {
	message := fmt.Sprintf("📊 *LAPORAN BULANAN - %s %d*\n\n", report.Month, report.Year)
	message += fmt.Sprintf("💰 *Total Pemasukan*: Rp %s\n", formatCurrency(report.TotalIncome))
	message += fmt.Sprintf("💸 *Total Pengeluaran*: Rp %s\n", formatCurrency(report.TotalExpense))
	message += fmt.Sprintf("📈 *Saldo Bersih*: Rp %s\n\n", formatCurrency(report.NetBalance))

	message += "*Kategori:*\n"
	for _, breakdown := range report.CategoryBreakdown {
		message += fmt.Sprintf("• %s: Rp %s (%.1f%%) - %d transaksi\n",
			breakdown.Category,
			formatCurrency(breakdown.TotalAmount),
			breakdown.Percentage,
			breakdown.TransactionCount)
	}

	message += fmt.Sprintf("\n📅 *Dibuat pada*: %s", report.GeneratedAt.Format("2006-01-02 15:04:05"))

	return message
}

// formatCurrency formats a number as Indonesian currency
func formatCurrency(amount float64) string {
	return utils.FormatIndonesianNumber(amount)
}
