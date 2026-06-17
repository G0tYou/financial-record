package http

import "net/http"

// Router sets up the HTTP routes
type Router struct {
	handler *TransactionHandler
}

// NewRouter creates a new router
func NewRouter(handler *TransactionHandler) *Router {
	return &Router{
		handler: handler,
	}
}

// SetupRoutes configures the HTTP routes
func (r *Router) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Webhook endpoint for Fontee
	mux.HandleFunc("/api/webhook", r.handler.HandleWebhook)

	// Balance endpoint
	mux.HandleFunc("/api/balance", r.handler.HandleGetBalance)

	// Transaction history endpoint
	mux.HandleFunc("/api/history", r.handler.HandleGetHistory)

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return mux
}
