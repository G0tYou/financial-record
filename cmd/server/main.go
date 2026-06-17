package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"financial-record/internal/config"
	"financial-record/internal/repository"
	httptransport "financial-record/internal/transport/http"
	"financial-record/internal/usecase"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Validate required configuration
	if cfg.GoogleCredentials == "" {
		log.Fatal("GOOGLE_CREDENTIALS environment variable is required")
	}
	if cfg.SpreadsheetID == "" {
		log.Fatal("SPREADSHEET_ID environment variable is required")
	}

	// Initialize repository
	ctx := context.Background()
	transactionRepo, err := repository.NewGoogleSheetsRepository(ctx, cfg.GoogleCredentials, cfg.SpreadsheetID)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}

	// Initialize use case
	transactionUseCase := usecase.NewTransactionUseCase(transactionRepo)

	// Initialize HTTP handler
	transactionHandler := httptransport.NewTransactionHandler(transactionUseCase)

	// Setup router
	router := httptransport.NewRouter(transactionHandler)
	mux := router.SetupRoutes()

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Starting server on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
