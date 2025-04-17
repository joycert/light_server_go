package api

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs the details of each request and response
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Start timer
		start := time.Now()

		// Read the request body
		var requestBody []byte
		if r.Body != nil {
			requestBody, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Create a custom response writer to capture the response
		responseWriter := &responseWriter{
			ResponseWriter: w,
			body:          new(bytes.Buffer),
			statusCode:    http.StatusOK,
		}

		// Call the next handler
		next(responseWriter, r)

		// Calculate duration
		duration := time.Since(start)

		// Log the request and response details
		log.Printf("[%s] %s %s - Duration: %v\nRequest Body: %s\nResponse Status: %d\nResponse Body: %s",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			duration,
			string(requestBody),
			responseWriter.statusCode,
			responseWriter.body.String(),
		)
	}
}

// responseWriter is a custom response writer that captures the response
type responseWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the response body
func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
} 