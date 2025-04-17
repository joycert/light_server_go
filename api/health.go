package api

import (
	"encoding/json"
	"net/http"
)

// HealthResponse represents the health check response
// @Description Response object containing the health status of the server
type HealthResponse struct {
	Status string `json:"status"` // The health status of the server
}

// @Summary Health check endpoint
// @Description Returns the health status of the server
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HealthResponse{
		Status: "healthy",
	})
} 