package repository

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"

	"financial-record/internal/domain/entities"
	"financial-record/internal/domain/repository"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// GoogleSheetsCategoryRepository implements CategoryRepository using Google Sheets
type GoogleSheetsCategoryRepository struct {
	sheetsService *sheets.Service
	spreadsheetID string
	categories    []*entities.Category
	keywords      []*entities.CategoryKeyword
	mu            sync.RWMutex
}

// NewGoogleSheetsCategoryRepository creates a new Google Sheets category repository
func NewGoogleSheetsCategoryRepository(ctx context.Context, credentialsJSON, spreadsheetID string) (repository.CategoryRepository, error) {
	sheetsService, err := sheets.NewService(ctx, option.WithCredentialsJSON([]byte(credentialsJSON)))
	if err != nil {
		return nil, fmt.Errorf("unable to create sheets service: %w", err)
	}

	repo := &GoogleSheetsCategoryRepository{
		sheetsService: sheetsService,
		spreadsheetID: spreadsheetID,
	}

	// Initialize categories and keywords sheets
	if err := repo.initializeCategoriesSheet(); err != nil {
		return nil, fmt.Errorf("unable to initialize categories sheet: %w", err)
	}

	if err := repo.initializeKeywordsSheet(); err != nil {
		return nil, fmt.Errorf("unable to initialize keywords sheet: %w", err)
	}

	// Load categories and keywords into memory
	if err := repo.loadCategories(); err != nil {
		return nil, fmt.Errorf("unable to load categories: %w", err)
	}

	if err := repo.loadKeywords(); err != nil {
		return nil, fmt.Errorf("unable to load keywords: %w", err)
	}

	return repo, nil
}

// initializeCategoriesSheet creates the Categories sheet if it doesn't exist
func (r *GoogleSheetsCategoryRepository) initializeCategoriesSheet() error {
	sheetName := "Categories"

	// Check if sheet exists
	spreadsheet, err := r.sheetsService.Spreadsheets.Get(r.spreadsheetID).Do()
	if err != nil {
		return fmt.Errorf("unable to get spreadsheet: %w", err)
	}

	sheetExists := false
	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			sheetExists = true
			break
		}
	}

	// Create sheet if it doesn't exist
	if !sheetExists {
		addSheetRequest := &sheets.Request{
			AddSheet: &sheets.AddSheetRequest{
				Properties: &sheets.SheetProperties{
					Title: sheetName,
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
	readRange := fmt.Sprintf("%s!A1:D1", sheetName)
	resp, err := r.sheetsService.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return fmt.Errorf("unable to read sheet: %w", err)
	}

	if len(resp.Values) == 0 {
		headers := []interface{}{"ID", "Name", "Description", "Color"}
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

// initializeKeywordsSheet creates the Keywords sheet if it doesn't exist
func (r *GoogleSheetsCategoryRepository) initializeKeywordsSheet() error {
	sheetName := "Keywords"

	// Check if sheet exists
	spreadsheet, err := r.sheetsService.Spreadsheets.Get(r.spreadsheetID).Do()
	if err != nil {
		return fmt.Errorf("unable to get spreadsheet: %w", err)
	}

	sheetExists := false
	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			sheetExists = true
			break
		}
	}

	// Create sheet if it doesn't exist
	if !sheetExists {
		addSheetRequest := &sheets.Request{
			AddSheet: &sheets.AddSheetRequest{
				Properties: &sheets.SheetProperties{
					Title: sheetName,
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
	readRange := fmt.Sprintf("%s!A1:D1", sheetName)
	resp, err := r.sheetsService.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return fmt.Errorf("unable to read sheet: %w", err)
	}

	if len(resp.Values) == 0 {
		headers := []interface{}{"ID", "Keyword", "CategoryID", "Priority"}
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

// loadCategories loads categories from Google Sheets into memory
func (r *GoogleSheetsCategoryRepository) loadCategories() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	readRange := "Categories!A:D"
	resp, err := r.sheetsService.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return fmt.Errorf("unable to read categories: %w", err)
	}

	r.categories = []*entities.Category{}
	if len(resp.Values) > 1 { // Skip header row
		for i := 1; i < len(resp.Values); i++ {
			row := resp.Values[i]
			if len(row) >= 4 {
				category := &entities.Category{
					ID:          row[0].(string),
					Name:        row[1].(string),
					Description: row[2].(string),
					Color:       row[3].(string),
				}
				r.categories = append(r.categories, category)
			}
		}
	}

	return nil
}

// loadKeywords loads keywords from Google Sheets into memory
func (r *GoogleSheetsCategoryRepository) loadKeywords() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	readRange := "Keywords!A:D"
	resp, err := r.sheetsService.Spreadsheets.Values.Get(r.spreadsheetID, readRange).Do()
	if err != nil {
		return fmt.Errorf("unable to read keywords: %w", err)
	}

	r.keywords = []*entities.CategoryKeyword{}
	if len(resp.Values) > 1 { // Skip header row
		for i := 1; i < len(resp.Values); i++ {
			row := resp.Values[i]
			if len(row) >= 4 {
				priority, _ := strconv.Atoi(row[3].(string))
				keyword := &entities.CategoryKeyword{
					ID:         row[0].(string),
					Keyword:    row[1].(string),
					CategoryID: row[2].(string),
					Priority:   priority,
				}
				r.keywords = append(r.keywords, keyword)
			}
		}
	}

	return nil
}

// GetAllCategories retrieves all categories
func (r *GoogleSheetsCategoryRepository) GetAllCategories() ([]*entities.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to avoid external modification
	categories := make([]*entities.Category, len(r.categories))
	copy(categories, r.categories)

	return categories, nil
}

// CreateCategory creates a new category
func (r *GoogleSheetsCategoryRepository) CreateCategory(category *entities.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	readRange := "Categories!A:D"
	values := []interface{}{
		category.ID,
		category.Name,
		category.Description,
		category.Color,
	}

	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{values},
	}

	_, err := r.sheetsService.Spreadsheets.Values.Append(r.spreadsheetID, readRange, valueRange).ValueInputOption("RAW").Do()
	if err != nil {
		return fmt.Errorf("unable to append category: %w", err)
	}

	r.categories = append(r.categories, category)
	log.Printf("Category created: %+v", category)
	return nil
}

// GetAllKeywords retrieves all category keywords
func (r *GoogleSheetsCategoryRepository) GetAllKeywords() ([]*entities.CategoryKeyword, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to avoid external modification
	keywords := make([]*entities.CategoryKeyword, len(r.keywords))
	copy(keywords, r.keywords)

	return keywords, nil
}

// CreateKeyword creates a new category keyword
func (r *GoogleSheetsCategoryRepository) CreateKeyword(keyword *entities.CategoryKeyword) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	readRange := "Keywords!A:D"
	values := []interface{}{
		keyword.ID,
		keyword.Keyword,
		keyword.CategoryID,
		keyword.Priority,
	}

	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{values},
	}

	_, err := r.sheetsService.Spreadsheets.Values.Append(r.spreadsheetID, readRange, valueRange).ValueInputOption("RAW").Do()
	if err != nil {
		return fmt.Errorf("unable to append keyword: %w", err)
	}

	r.keywords = append(r.keywords, keyword)
	log.Printf("Keyword created: %+v", keyword)
	return nil
}

// CategorizeTransaction categorizes a transaction based on its description
func (r *GoogleSheetsCategoryRepository) CategorizeTransaction(description string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.keywords) == 0 {
		return "Other", nil
	}

	// Sort keywords by priority (higher priority first)
	sortedKeywords := make([]*entities.CategoryKeyword, len(r.keywords))
	copy(sortedKeywords, r.keywords)
	sort.Slice(sortedKeywords, func(i, j int) bool {
		return sortedKeywords[i].Priority > sortedKeywords[j].Priority
	})

	// Convert description to lowercase for case-insensitive matching
	lowerDescription := strings.ToLower(description)

	// Check each keyword for matches
	for _, keyword := range sortedKeywords {
		if strings.Contains(lowerDescription, strings.ToLower(keyword.Keyword)) {
			// Find category name - match by CategoryID (which can be name or ID)
			for _, category := range r.categories {
				// Match by either category name or category ID for flexibility
				if category.Name == keyword.CategoryID || category.ID == keyword.CategoryID {
					log.Printf("Categorized '%s' as '%s' (matched keyword: '%s', CategoryID: '%s')", description, category.Name, keyword.Keyword, keyword.CategoryID)
					return category.Name, nil
				}
			}
		}
	}

	// If no match found, return "Other"
	log.Printf("No category match for '%s', using 'Other'", description)
	return "Other", nil
}

// AddDefaultCategories adds default categories for Indonesian context
func (r *GoogleSheetsCategoryRepository) AddDefaultCategories() error {
	// Check if categories already exist in Google Sheets
	if len(r.categories) > 0 {
		log.Printf("Categories already exist in Google Sheets, skipping default category creation")
		return nil
	}

	log.Printf("Categories sheet is empty, adding default categories")

	defaultCategories := []*entities.Category{
		entities.NewCategory("Food", "Makanan dan minuman", "#FF6B6B"),
		entities.NewCategory("Transportation", "Transportasi dan perjalanan", "#4ECDC4"),
		entities.NewCategory("Shopping", "Belanja dan kebutuhan", "#45B7D1"),
		entities.NewCategory("Entertainment", "Hiburan dan rekreasi", "#96CEB4"),
		entities.NewCategory("Bills", "Tagihan dan utilitas", "#FFEAA7"),
		entities.NewCategory("Health", "Kesehatan dan medis", "#DDA0DD"),
		entities.NewCategory("Education", "Pendidikan dan pelatihan", "#98D8C8"),
		entities.NewCategory("Income", "Pemasukan dan gaji", "#82E0AA"),
		entities.NewCategory("Other", "Lain-lain", "#95A5A6"),
	}

	for _, category := range defaultCategories {
		if err := r.CreateCategory(category); err != nil {
			log.Printf("Failed to create default category %s: %v", category.Name, err)
		}
	}

	// Reload categories after creation to get the correct IDs
	if err := r.loadCategories(); err != nil {
		log.Printf("Failed to reload categories: %v", err)
	}

	// Add default keywords for Indonesian context using category names
	defaultKeywords := []*entities.CategoryKeyword{
		// Food keywords
		entities.NewCategoryKeyword("makan", "Food", 10),
		entities.NewCategoryKeyword("makanan", "Food", 10),
		entities.NewCategoryKeyword("minum", "Food", 10),
		entities.NewCategoryKeyword("kopi", "Food", 9),
		entities.NewCategoryKeyword("teh", "Food", 9),
		entities.NewCategoryKeyword("restoran", "Food", 8),
		entities.NewCategoryKeyword("warung", "Food", 8),
		entities.NewCategoryKeyword("cafe", "Food", 8),

		// Transportation keywords
		entities.NewCategoryKeyword("transport", "Transportation", 10),
		entities.NewCategoryKeyword("travel", "Transportation", 10),
		entities.NewCategoryKeyword("bus", "Transportation", 9),
		entities.NewCategoryKeyword("kereta", "Transportation", 9),
		entities.NewCategoryKeyword("ojek", "Transportation", 9),
		entities.NewCategoryKeyword("grab", "Transportation", 9),
		entities.NewCategoryKeyword("gojek", "Transportation", 9),
		entities.NewCategoryKeyword("bensin", "Transportation", 8),
		entities.NewCategoryKeyword("parkir", "Transportation", 8),

		// Shopping keywords
		entities.NewCategoryKeyword("belanja", "Shopping", 10),
		entities.NewCategoryKeyword("shopping", "Shopping", 10),
		entities.NewCategoryKeyword("toko", "Shopping", 9),
		entities.NewCategoryKeyword("mall", "Shopping", 9),
		entities.NewCategoryKeyword("pasar", "Shopping", 9),
		entities.NewCategoryKeyword("minimarket", "Shopping", 8),
		entities.NewCategoryKeyword("indomaret", "Shopping", 8),
		entities.NewCategoryKeyword("alfamart", "Shopping", 8),

		// Entertainment keywords
		entities.NewCategoryKeyword("hiburan", "Entertainment", 10),
		entities.NewCategoryKeyword("film", "Entertainment", 9),
		entities.NewCategoryKeyword("bioskop", "Entertainment", 9),
		entities.NewCategoryKeyword("game", "Entertainment", 9),
		entities.NewCategoryKeyword("musik", "Entertainment", 8),
		entities.NewCategoryKeyword("konser", "Entertainment", 8),

		// Bills keywords
		entities.NewCategoryKeyword("tagihan", "Bills", 10),
		entities.NewCategoryKeyword("listrik", "Bills", 9),
		entities.NewCategoryKeyword("air", "Bills", 9),
		entities.NewCategoryKeyword("internet", "Bills", 9),
		entities.NewCategoryKeyword("puls", "Bills", 8),
		entities.NewCategoryKeyword("kuota", "Bills", 8),

		// Health keywords
		entities.NewCategoryKeyword("kesehatan", "Health", 10),
		entities.NewCategoryKeyword("dokter", "Health", 9),
		entities.NewCategoryKeyword("obat", "Health", 9),
		entities.NewCategoryKeyword("rumah sakit", "Health", 9),
		entities.NewCategoryKeyword("apotek", "Health", 8),

		// Education keywords
		entities.NewCategoryKeyword("pendidikan", "Education", 10),
		entities.NewCategoryKeyword("sekolah", "Education", 9),
		entities.NewCategoryKeyword("kuliah", "Education", 9),
		entities.NewCategoryKeyword("buku", "Education", 9),
		entities.NewCategoryKeyword("kursus", "Education", 8),

		// Income keywords
		entities.NewCategoryKeyword("gaji", "Income", 10),
		entities.NewCategoryKeyword("jatah", "Income", 9),
		entities.NewCategoryKeyword("bonus", "Income", 9),
		entities.NewCategoryKeyword("pemasukan", "Income", 8),
		entities.NewCategoryKeyword("income", "Income", 8),
	}

	for _, keyword := range defaultKeywords {
		if err := r.CreateKeyword(keyword); err != nil {
			log.Printf("Failed to create default keyword %s: %v", keyword.Keyword, err)
		}
	}

	return nil
}
