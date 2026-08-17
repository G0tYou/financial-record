package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"financial-record/internal/config"
	"financial-record/internal/repository"
	"financial-record/internal/scheduler"
	httptransport "financial-record/internal/transport/http"
	"financial-record/internal/usecase"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Validate required configuration
	if cfg.GoogleCredentials == "" {
		log.Println("GOOGLE_CREDENTIALS environment variable is required")
	}
	if cfg.SpreadsheetID == "" {
		log.Println("SPREADSHEET_ID environment variable is required")
	}

	// Initialize repository
	ctx := context.Background()
	transactionRepo, err := repository.NewGoogleSheetsRepository(ctx, cfg.GoogleCredentials, cfg.SpreadsheetID)
	if err != nil {
		log.Printf("Failed to initialize repository: %v", err)
	}

	categoryRepo, err := repository.NewGoogleSheetsCategoryRepository(ctx, cfg.GoogleCredentials, cfg.SpreadsheetID)
	if err != nil {
		log.Printf("Failed to initialize category repository: %v", err)
		categoryRepo = nil // Set to nil to prevent panic
	}

	// Add default categories if needed
	if categoryRepo != nil {
		if err := categoryRepo.AddDefaultCategories(); err != nil {
			log.Printf("Failed to add default categories: %v", err)
		}
	}

	// Initialize use cases
	transactionUseCase := usecase.NewTransactionUseCase(transactionRepo)
	categoryUseCase := usecase.NewCategoryUseCase(categoryRepo)
	reportUseCase := usecase.NewReportUseCase(transactionRepo)

	// Initialize HTTP handler
	transactionHandler := httptransport.NewTransactionHandler(transactionUseCase, categoryUseCase, reportUseCase)

	// Setup router
	router := httptransport.NewRouter(transactionHandler)
	mux := router.SetupRoutes()

	// Wrap with logging middleware
	loggedMux := httptransport.LoggingMiddleware(mux)

	// Initialize scheduler if default phone is configured
	var reportScheduler *scheduler.MonthlyReportScheduler
	if cfg.DefaultPhone != "" {
		reportScheduler = scheduler.NewMonthlyReportScheduler(reportUseCase, cfg.DefaultPhone)
		if err := reportScheduler.Start(); err != nil {
			log.Printf("Failed to start report scheduler: %v", err)
		}
		defer reportScheduler.Stop()
	}

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Starting server on %s", addr)

	if err := http.ListenAndServe(addr, loggedMux); err != nil {
		log.Printf("Server failed to start: %v", err)
	}
}
