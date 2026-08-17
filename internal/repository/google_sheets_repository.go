package repository

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
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
	readRange := fmt.Sprintf("%s!A1:G1", r.sheetName)
	resp, err := r.sheetsService.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return fmt.Errorf("unable to read sheet: %w", err)
	}

	if len(resp.Values) == 0 {
		// Add headers with Category column
		headers := []interface{}{"Date", "Phone", "Action", "Amount", "Balance", "Category", "Notes"}
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

// getPreviousMonthSheetName returns the sheet name for the previous month
func (r *GoogleSheetsRepository) getPreviousMonthSheetName() string {
	now := time.Now()
	previousMonth := now.AddDate(0, -1, 0)
	return previousMonth.Month().String() + " " + strconv.Itoa(previousMonth.Year())
}

// getBalanceFromSheet retrieves the current balance for a phone number from a specific sheet
func (r *GoogleSheetsRepository) getBalanceFromSheet(phone, sheetName string) (float64, error) {
	readRange := fmt.Sprintf("%s!A:G", sheetName)
	resp, err := r.sheetsService.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return 0, fmt.Errorf("unable to read sheet: %w", err)
	}

	var balance float64
	if len(resp.Values) > 1 { // Skip header row
		// Get the last transaction for this phone number
		for i := len(resp.Values) - 1; i >= 1; i-- {
			row := resp.Values[i]
			if len(row) >= 6 {
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

// GetBalanceFromPreviousMonth retrieves the balance from the previous month's sheet
func (r *GoogleSheetsRepository) GetBalanceFromPreviousMonth(phone string) (float64, error) {
	previousMonthSheet := r.getPreviousMonthSheetName()
	return r.getBalanceFromSheet(phone, previousMonthSheet)
}

// ensureCurrentMonthSheet checks if we're using the correct month's sheet and switches if needed
func (r *GoogleSheetsRepository) ensureCurrentMonthSheet() error {
	currentSheetName := time.Now().Month().String() + " " + strconv.Itoa(time.Now().Year())

	if r.sheetName != currentSheetName {
		log.Printf("Month changed from %s to %s, switching sheets", r.sheetName, currentSheetName)
		r.sheetName = currentSheetName
		if err := r.initializeSheet(); err != nil {
			return fmt.Errorf("unable to initialize new month sheet: %w", err)
		}
		log.Printf("Successfully switched to sheet: %s", r.sheetName)
	}

	return nil
}

// SaveTransaction saves a transaction to the spreadsheet
func (r *GoogleSheetsRepository) SaveTransaction(transaction *entities.Transaction) error {
	if err := r.ensureCurrentMonthSheet(); err != nil {
		return err
	}

	readRange := fmt.Sprintf("%s!A:G", r.sheetName)

	values := []interface{}{
		transaction.Timestamp.Format("2006-01-02 15:04:05"),
		transaction.Phone,
		string(transaction.Action),
		transaction.Amount,
		transaction.Balance,
		transaction.Category,
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
	if err := r.ensureCurrentMonthSheet(); err != nil {
		return 0, err
	}

	readRange := fmt.Sprintf("%s!A:G", r.sheetName)
	resp, err := r.sheetsService.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return 0, fmt.Errorf("unable to read sheet: %w", err)
	}

	var balance float64
	if len(resp.Values) > 1 { // Skip header row
		// Get the last transaction for this phone number
		for i := len(resp.Values) - 1; i >= 1; i-- {
			row := resp.Values[i]
			if len(row) >= 6 {
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
	if err := r.ensureCurrentMonthSheet(); err != nil {
		return nil, err
	}

	readRange := fmt.Sprintf("%s!A:G", r.sheetName)
	resp, err := r.sheetsService.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to read sheet: %w", err)
	}

	var transactions []*entities.Transaction
	if len(resp.Values) > 1 { // Skip header row
		for i := 1; i < len(resp.Values); i++ {
			row := resp.Values[i]
			if len(row) >= 6 {
				if rowPhone, ok := row[1].(string); ok && rowPhone == phone {
					transaction := &entities.Transaction{
						Phone:    rowPhone,
						Action:   row[2].(string),
						Category: "",
						Notes:    "",
					}

					if amountStr, ok := row[3].(string); ok {
						transaction.Amount, _ = strconv.ParseFloat(amountStr, 64)
					}

					if balanceStr, ok := row[4].(string); ok {
						transaction.Balance, _ = strconv.ParseFloat(balanceStr, 64)
					}

					if len(row) >= 6 {
						if category, ok := row[5].(string); ok {
							transaction.Category = category
						}
					}

					if len(row) >= 7 {
						if notes, ok := row[6].(string); ok {
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

// GetAllTransactionsForMonth retrieves all transactions for a specific month from all sheets
func (r *GoogleSheetsRepository) GetAllTransactionsForMonth(phone string, month string, year int) ([]*entities.Transaction, error) {
	// Get all sheets in the spreadsheet
	spreadsheet, err := r.sheetsService.Spreadsheets.Get(r.spreadsheetID).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to get spreadsheet: %w", err)
	}

	var allTransactions []*entities.Transaction
	targetSheetName := strings.ToLower(month + " " + strconv.Itoa(year))

	log.Printf("Looking for transactions in sheet: %s for phone: %s", targetSheetName, phone)

	// Check each sheet for transactions matching the target month
	for _, sheet := range spreadsheet.Sheets {
		sheetName := sheet.Properties.Title
		lowerSheetName := strings.ToLower(sheetName)

		// Skip if it's not a month/year sheet (like Categories or Keywords)
		if lowerSheetName == "categories" || lowerSheetName == "keywords" {
			continue
		}

		// If we're looking for a specific month, only check that sheet (case insensitive)
		if targetSheetName != "" && lowerSheetName != targetSheetName {
			log.Printf("Skipping sheet %s (looking for %s)", sheetName, targetSheetName)
			continue
		}

		log.Printf("Reading transactions from sheet: %s", sheetName)

		// Read transactions from this sheet
		readRange := fmt.Sprintf("%s!A:G", sheetName)
		resp, err := r.sheetsService.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
		if err != nil {
			log.Printf("Unable to read sheet %s: %v", sheetName, err)
			continue
		}

		if len(resp.Values) > 1 { // Skip header row
			for i := 1; i < len(resp.Values); i++ {
				row := resp.Values[i]
				if len(row) >= 6 {
					if rowPhone, ok := row[1].(string); ok && rowPhone == phone {
						transaction := &entities.Transaction{
							Phone:    rowPhone,
							Action:   row[2].(string),
							Category: "",
							Notes:    "",
						}

						if amountStr, ok := row[3].(string); ok {
							transaction.Amount, _ = strconv.ParseFloat(amountStr, 64)
						}

						if balanceStr, ok := row[4].(string); ok {
							transaction.Balance, _ = strconv.ParseFloat(balanceStr, 64)
						}

						if len(row) >= 6 {
							if category, ok := row[5].(string); ok {
								transaction.Category = category
							}
						}

						if len(row) >= 7 {
							if notes, ok := row[6].(string); ok {
								transaction.Notes = notes
							}
						}

						allTransactions = append(allTransactions, transaction)
					}
				}
			}
		}
	}

	log.Printf("Found %d transactions for phone %s in month %s %d", len(allTransactions), phone, month, year)
	return allTransactions, nil
}
