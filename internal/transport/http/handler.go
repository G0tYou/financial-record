package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"financial-record/internal/usecase"
	"financial-record/internal/utils"
)

// TransactionHandler handles HTTP requests for transactions
type TransactionHandler struct {
	transactionUseCase *usecase.TransactionUseCase
	categoryUseCase    *usecase.CategoryUseCase
	reportUseCase      *usecase.ReportUseCase
}

// NewTransactionHandler creates a new transaction handler
func NewTransactionHandler(transactionUseCase *usecase.TransactionUseCase, categoryUseCase *usecase.CategoryUseCase, reportUseCase *usecase.ReportUseCase) *TransactionHandler {
	return &TransactionHandler{
		transactionUseCase: transactionUseCase,
		categoryUseCase:    categoryUseCase,
		reportUseCase:      reportUseCase,
	}
}

// FonteeWebhookRequest represents the webhook request from Fontee
type FonteeWebhookRequest struct {
	Phone    string `json:"sender"`
	Message  string `json:"message"`
	SenderID string `json:"senderlid,omitempty"`
}

// TransactionResponse represents the response for a transaction
type TransactionResponse struct {
	Success   bool    `json:"success"`
	Message   string  `json:"message"`
	Phone     string  `json:"phone,omitempty"`
	Action    string  `json:"action,omitempty"`
	Amount    float64 `json:"amount,omitempty"`
	Balance   float64 `json:"balance,omitempty"`
	Timestamp string  `json:"timestamp,omitempty"`
}

// HandleWebhook handles the webhook from Fontee
func (h *TransactionHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendJSONResponse(w, http.StatusOK, "OK")
		return
	}

	var req FonteeWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Parse message to extract action and amount
	code, amount, description, err := h.parseMessage(req.Message)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Normalize phone number
	normalizedPhone := normalizePhoneNumber(req.Phone)

	// Categorize the transaction based on description
	category := "Other"
	if h.categoryUseCase != nil {
		categorizedCategory, categorizeErr := h.categoryUseCase.CategorizeTransaction(description)
		if categorizeErr != nil {
			log.Printf("Failed to categorize transaction: %v", categorizeErr)
		} else {
			category = categorizedCategory
		}
	}

	// Process transaction with category
	transaction, err := h.transactionUseCase.ProcessTransaction(normalizedPhone, code, amount, category, description)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Send success response
	response := TransactionResponse{
		Success:   true,
		Message:   "Success",
		Phone:     normalizedPhone,
		Action:    string(transaction.Action),
		Amount:    transaction.Amount,
		Balance:   transaction.Balance,
		Timestamp: transaction.Timestamp.Format("2006-01-02 15:04:05"),
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

// HandleGetBalance handles GET requests to retrieve balance
func (h *TransactionHandler) HandleGetBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	phone := r.URL.Query().Get("phone")
	if phone == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	// Normalize phone number
	phone = normalizePhoneNumber(phone)

	balance, err := h.transactionUseCase.GetBalance(phone)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"success": true,
		"phone":   phone,
		"balance": balance,
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

// HandleGetHistory handles GET requests to retrieve transaction history
func (h *TransactionHandler) HandleGetHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	phone := r.URL.Query().Get("phone")
	if phone == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	// Normalize phone number
	phone = normalizePhoneNumber(phone)

	transactions, err := h.transactionUseCase.GetTransactionHistory(phone)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"success":      true,
		"phone":        phone,
		"transactions": transactions,
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

// HandleCategories handles GET requests to retrieve all categories
func (h *TransactionHandler) HandleCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	categories, err := h.categoryUseCase.GetAllCategories()
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"success":    true,
		"categories": categories,
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

// HandleCreateCategory handles POST requests to create a new category
func (h *TransactionHandler) HandleCreateCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Category name is required")
		return
	}

	if err := h.categoryUseCase.CreateCategory(req.Name, req.Description, req.Color); err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Category created successfully",
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

// HandleKeywords handles GET requests to retrieve all keywords
func (h *TransactionHandler) HandleKeywords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	keywords, err := h.categoryUseCase.GetAllKeywords()
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"success":  true,
		"keywords": keywords,
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

// HandleCreateKeyword handles POST requests to create a new keyword
func (h *TransactionHandler) HandleCreateKeyword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Keyword    string `json:"keyword"`
		CategoryID string `json:"category_id"`
		Priority   int    `json:"priority"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Keyword == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Keyword is required")
		return
	}

	if req.CategoryID == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Category ID is required")
		return
	}

	if err := h.categoryUseCase.CreateKeyword(req.Keyword, req.CategoryID, req.Priority); err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Keyword created successfully",
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

// HandleMonthlyReport handles GET requests to generate monthly summary report
func (h *TransactionHandler) HandleMonthlyReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	phone := r.URL.Query().Get("phone")
	if phone == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	// Normalize phone number: convert "0..." to "+62..."
	phone = normalizePhoneNumber(phone)

	// Get month and year from query params, default to current month
	month := r.URL.Query().Get("month")
	yearStr := r.URL.Query().Get("year")

	if month == "" {
		now := time.Now()
		month = now.Month().String()
		log.Printf("No month specified, using current month: %s", month)
	}

	year := 0
	if yearStr != "" {
		year, _ = strconv.Atoi(yearStr)
	}
	if year == 0 {
		year = time.Now().Year()
	}

	log.Printf("Generating monthly report for phone: %s, month: %s, year: %d", phone, month, year)

	// Generate report
	report, err := h.reportUseCase.GenerateMonthlyReport(phone, month, year)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"success": true,
		"report":  report,
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

// HandleSendReport handles POST requests to send monthly report via WhatsApp
func (h *TransactionHandler) HandleSendReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Phone string `json:"phone"`
		Month string `json:"month,omitempty"`
		Year  int    `json:"year,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Phone == "" {
		h.sendErrorResponse(w, http.StatusBadRequest, "Phone number is required")
		return
	}

	// Normalize phone number
	normalizedPhone := normalizePhoneNumber(req.Phone)

	// Get month and year, default to current month
	month := req.Month
	year := req.Year

	if month == "" {
		now := time.Now()
		month = now.Month().String()
	}
	if year == 0 {
		year = time.Now().Year()
	}

	// Generate report
	report, err := h.reportUseCase.GenerateMonthlyReport(normalizedPhone, month, year)
	if err != nil {
		h.sendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Format as WhatsApp message
	message := h.reportUseCase.FormatReportAsWhatsAppMessage(report)

	// In a real implementation, this would send via Fontee API
	// For now, return the formatted message
	response := map[string]interface{}{
		"success":          true,
		"message":          "Report generated successfully",
		"phone":            normalizedPhone,
		"whatsapp_message": message,
	}

	h.sendJSONResponse(w, http.StatusOK, response)
}

// parseMessage parses the WhatsApp message to extract action and amount
func (h *TransactionHandler) parseMessage(message string) (string, float64, string, error) {
	if len(message) == 0 {
		return "", 0, "", fmt.Errorf("empty message")
	} else if len(message) < 2 {
		return "", 0, "", fmt.Errorf("invalid format")
	}

	// Extract action (first character)
	action := string(message[0])

	// Extract amount (rest of the string until space or end)
	remaining := message[1:]
	var amountStr string
	var description string

	// Split by space to separate amount from description
	parts := strings.Fields(remaining)
	if len(parts) > 0 {
		amountStr = parts[0]
		if len(parts) > 1 {
			description = strings.Join(parts[1:], " ")
		}
	} else {
		amountStr = remaining
	}

	// Parse Indonesian number format (e.g., "3.000.000")
	amount, err := utils.ParseIndonesianNumber(amountStr)
	if err != nil {
		return "", 0, "", fmt.Errorf("invalid amount format: %s", amountStr)
	}

	if amount <= 0 {
		return "", 0, "", fmt.Errorf("amount must be greater than 0")
	}

	// Convert + to add, - to subtract
	var code string
	if action == "+" {
		code = "add"
	} else if action == "-" {
		code = "subtract"
	} else {
		return "", 0, "", fmt.Errorf("invalid action: %s (use + or -)", action)
	}

	return code, amount, description, nil
}

// sendJSONResponse sends a JSON response
func (h *TransactionHandler) sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// sendErrorResponse sends an error response
func (h *TransactionHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := map[string]interface{}{
		"success": false,
		"error":   message,
	}
	h.sendJSONResponse(w, statusCode, response)
}

// normalizePhoneNumber normalizes phone number format
// Converts "0..." to "+62..." format
func normalizePhoneNumber(phone string) string {
	// Remove any existing + prefix for consistency
	phone = strings.TrimPrefix(phone, "+")

	// If starts with "0", replace with "62"
	if strings.HasPrefix(phone, "0") {
		phone = "62" + phone[1:]
	}

	// Add + prefix
	return "+" + phone
}
