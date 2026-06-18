package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"financial-record/internal/usecase"
)

// TransactionHandler handles HTTP requests for transactions
type TransactionHandler struct {
	transactionUseCase *usecase.TransactionUseCase
}

// NewTransactionHandler creates a new transaction handler
func NewTransactionHandler(transactionUseCase *usecase.TransactionUseCase) *TransactionHandler {
	return &TransactionHandler{
		transactionUseCase: transactionUseCase,
	}
}

// FonteeWebhookRequest represents the webhook request from Fontee
type FonteeWebhookRequest struct {
	Phone    string `json:"phone"`
	Message  string `json:"message"`
	SenderID string `json:"sender_id,omitempty"`
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
		h.sendErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
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

	// Process transaction
	transaction, err := h.transactionUseCase.ProcessTransaction(req.Phone, code, amount, description)
	if err != nil {
		h.sendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Send success response
	response := TransactionResponse{
		Success:   true,
		Message:   "Success",
		Phone:     transaction.Phone,
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

// parseMessage parses the WhatsApp message to extract action and amount
func (h *TransactionHandler) parseMessage(message string) (string, float64, string, error) {
	if len(message) == 0 {
		return "", 0, "", fmt.Errorf("empty message")
	} else if len(message) < 3 {
		return "", 0, "", fmt.Errorf("invalid format")
	}

	parts := strings.SplitN(message, " ", 3)

	code := string(parts[0])
	fmt.Println("code:", code)

	amountStr := string(parts[1])
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return "", 0, "", fmt.Errorf("invalid amount format: %s", amountStr)
	}

	description := parts[2]

	if amount <= 0 {
		return "", 0, "", fmt.Errorf("amount must be greater than 0")
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
