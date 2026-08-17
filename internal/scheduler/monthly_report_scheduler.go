package scheduler

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"financial-record/internal/usecase"
)

// MonthlyReportScheduler handles scheduled monthly report generation
type MonthlyReportScheduler struct {
	cron          *cron.Cron
	reportUseCase *usecase.ReportUseCase
	phone         string // Default phone number for single-user mode
}

// NewMonthlyReportScheduler creates a new monthly report scheduler
func NewMonthlyReportScheduler(reportUseCase *usecase.ReportUseCase, phone string) *MonthlyReportScheduler {
	return &MonthlyReportScheduler{
		cron:          cron.New(),
		reportUseCase: reportUseCase,
		phone:         phone,
	}
}

// Start starts the scheduler
func (s *MonthlyReportScheduler) Start() error {
	// Schedule monthly report for the last day of each month at 23:59
	// Cron format: second minute hour day month weekday
	// 59 23 28-31 * * - runs at 23:59 on days 28-31 of every month
	// We'll check if it's actually the last day in the job
	
	_, err := s.cron.AddFunc("59 23 28-31 * *", s.generateEndOfMonthReport)
	if err != nil {
		return err
	}

	s.cron.Start()
	log.Println("Monthly report scheduler started")
	return nil
}

// Stop stops the scheduler
func (s *MonthlyReportScheduler) Stop() {
	s.cron.Stop()
	log.Println("Monthly report scheduler stopped")
}

// generateEndOfMonthReport generates the end-of-month report
func (s *MonthlyReportScheduler) generateEndOfMonthReport() {
	now := time.Now()
	
	// Check if today is actually the last day of the month
	if !s.isLastDayOfMonth(now) {
		log.Printf("Today is not the last day of the month, skipping report generation")
		return
	}

	log.Printf("Generating end-of-month report for %s %d", now.Month().String(), now.Year())
	
	// Generate report for the current month
	report, err := s.reportUseCase.GenerateCurrentMonthReport(s.phone)
	if err != nil {
		log.Printf("Failed to generate monthly report: %v", err)
		return
	}

	// Format the report as WhatsApp message
	message := s.reportUseCase.FormatReportAsWhatsAppMessage(report)
	
	log.Printf("Monthly report generated: %s", message)
	
	// TODO: Send the report via Fontee API
	// This would involve calling the Fontee API to send the message to the user
}

// isLastDayOfMonth checks if the given date is the last day of its month
func (s *MonthlyReportScheduler) isLastDayOfMonth(date time.Time) bool {
	firstDayOfNextMonth := date.AddDate(0, 1, 0)
	firstDayOfNextMonth = time.Date(firstDayOfNextMonth.Year(), firstDayOfNextMonth.Month(), 1, 0, 0, 0, 0, date.Location())
	
	return date.After(firstDayOfNextMonth.Add(-time.Second))
}