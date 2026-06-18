package http

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"
)

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

// LoggingMiddleware logs HTTP requests and responses
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Read request body for logging
		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Log request details
		log.Printf("[%s] %s %s - Query: %s - Body: %s",
			start.Format("2006-01-02 15:04:05"),
			r.Method,
			r.URL.Path,
			r.URL.RawQuery,
			string(bodyBytes))

		// Create a response writer wrapper to capture status code
		rw := &responseWriter{w, http.StatusOK}

		// Call the next handler
		next.ServeHTTP(rw, r)

		// Log response
		log.Printf("[%s] Response: %d %s - Duration: %v",
			start.Format("2006-01-02 15:04:05"),
			rw.statusCode,
			http.StatusText(rw.statusCode),
			time.Since(start))
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
