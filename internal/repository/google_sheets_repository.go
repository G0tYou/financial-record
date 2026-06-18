package repository

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"financial-record/internal/domain/entities"
	"financial-record/internal/domain/repository"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// GoogleSheetsRepository implements TransactionRepository using Google Sheets
type GoogleSheetsRepository struct {
	sheetsService *sheets.Service
	spreadsheetID string
	sheetName     string
}

// NewGoogleSheetsRepository creates a new Google Sheets repository
func NewGoogleSheetsRepository(ctx context.Context, credentialsJSON, spreadsheetID string) (repository.TransactionRepository, error) {
	sheetsService, err := sheets.NewService(ctx, option.WithCredentialsJSON([]byte(credentialsJSON)))
	if err != nil {
		return nil, fmt.Errorf("unable to create sheets service: %w", err)
	}

	repo := &GoogleSheetsRepository{
		sheetsService: sheetsService,
		spreadsheetID: spreadsheetID,
		sheetName:     time.Now().Month().String() + " " + strconv.Itoa(time.Now().Year()),
	}

	// Initialize sheet if needed
	if err := repo.initializeSheet(); err != nil {
		return nil, fmt.Errorf("unable to initialize sheet: %w", err)
	}

	return repo, nil
}

// initializeSheet creates the sheet with headers if it doesn't exist
func (r *GoogleSheetsRepository) initializeSheet() error {
	// Check if sheet exists
	spreadsheet, err := r.sheetsService.Spreadsheets.Get(r.spreadsheetID).Do()
	if err != nil {
		return fmt.Errorf("unable to get spreadsheet: %w", err)
	}

	// Check if sheet with the given name exists
	sheetExists := false
	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == r.sheetName {
			sheetExists = true
			break
		}
	}

	// Create sheet if it doesn't exist
	if !sheetExists {
		addSheetRequest := &sheets.Request{
			AddSheet: &sheets.AddSheetRequest{
				Properties: &sheets.SheetProperties{
					Title: r.sheetName,
				},
			},
		}

		batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{addSheetRequest},
		}

		_, err = r.sheetsService.Spreadsheets.BatchUpdate(r.spreadsheetID, batchUpdateRequest).Do()
		if err != nil {
			return fmt.Errorf("unable to create sheet: %w", err)
		}
	}

	// Add headers if the sheet is empty
	readRange := fmt.Sprintf("%s!A1:F1", r.sheetName)
	resp, err := r.sheetsService.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return fmt.Errorf("unable to read sheet: %w", err)
	}

	if len(resp.Values) == 0 {
		// Add headers
		headers := []interface{}{"Date", "Phone", "Action", "Amount", "Balance", "Notes"}
		valueRange := &sheets.ValueRange{
			Values: [][]interface{}{headers},
		}

		_, err = r.sheetsService.Spreadsheets.Values.Append(r.spreadsheetID, readRange, valueRange).ValueInputOption("RAW").Do()
		if err != nil {
			return fmt.Errorf("unable to add headers: %w", err)
		}
	}

	return nil
}

// SaveTransaction saves a transaction to the spreadsheet
func (r *GoogleSheetsRepository) SaveTransaction(transaction *entities.Transaction) error {
	readRange := fmt.Sprintf("%s!A:F", r.sheetName)

	values := []interface{}{
		transaction.Timestamp.Format("2006-01-02 15:04:05"),
		transaction.Phone,
		string(transaction.Action),
		transaction.Amount,
		transaction.Balance,
		transaction.Notes,
	}

	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{values},
	}

	_, err := r.sheetsService.Spreadsheets.Values.Append(r.spreadsheetID, readRange, valueRange).ValueInputOption("RAW").Do()
	if err != nil {
		return fmt.Errorf("unable to append transaction: %w", err)
	}

	log.Printf("Transaction saved: %+v", transaction)
	return nil
}

// GetBalance retrieves the current balance for a phone number
func (r *GoogleSheetsRepository) GetBalance(phone string) (float64, error) {
	readRange := fmt.Sprintf("%s!A:F", r.sheetName)
	resp, err := r.sheetsService.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return 0, fmt.Errorf("unable to read sheet: %w", err)
	}

	var balance float64
	if len(resp.Values) > 1 { // Skip header row
		// Get the last transaction for this phone number
		for i := len(resp.Values) - 1; i >= 1; i-- {
			row := resp.Values[i]
			if len(row) >= 5 {
				if rowPhone, ok := row[1].(string); ok && rowPhone == phone {
					if balanceStr, ok := row[4].(string); ok {
						balance, err = strconv.ParseFloat(balanceStr, 64)
						if err != nil {
							return 0, fmt.Errorf("unable to parse balance: %w", err)
						}
						return balance, nil
					}
				}
			}
		}
	}

	return 0, nil // No transactions found, return 0
}

// GetTransactions retrieves all transactions for a phone number
func (r *GoogleSheetsRepository) GetTransactions(phone string) ([]*entities.Transaction, error) {
	readRange := fmt.Sprintf("%s!A:F", r.sheetName)
	resp, err := r.sheetsService.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to read sheet: %w", err)
	}

	var transactions []*entities.Transaction
	if len(resp.Values) > 1 { // Skip header row
		for i := 1; i < len(resp.Values); i++ {
			row := resp.Values[i]
			if len(row) >= 5 {
				if rowPhone, ok := row[1].(string); ok && rowPhone == phone {
					transaction := &entities.Transaction{
						Phone:  rowPhone,
						Action: row[2].(string),
						Notes:  "",
					}

					if amountStr, ok := row[3].(string); ok {
						transaction.Amount, _ = strconv.ParseFloat(amountStr, 64)
					}

					if balanceStr, ok := row[4].(string); ok {
						transaction.Balance, _ = strconv.ParseFloat(balanceStr, 64)
					}

					if len(row) >= 6 {
						if notes, ok := row[5].(string); ok {
							transaction.Notes = notes
						}
					}

					transactions = append(transactions, transaction)
				}
			}
		}
	}

	return transactions, nil
}
