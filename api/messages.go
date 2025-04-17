package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ryanjoyce/lights/lights_server_go/utils"
)

// Current RGB value
var currentRGB = utils.RGB{R: 0, G: 0, B: 0}

// Message represents an incoming message
// @Description Message object containing the text command to control the lights
type Message struct {
	Message string `json:"message"` // The text command to control the lights
}

// Response represents the server response
// @Description Response object containing the status, message and current RGB values
type Response struct {
	Status  string     `json:"status"`  // Status of the operation (success/error)
	Message string     `json:"message"` // Detailed message about the operation
	RGB     utils.RGB  `json:"rgb"`     // Current RGB values of the lights
}

// @Summary Process a message to control lights
// @Description Takes a text message and converts it to RGB values to control the lights
// @Tags messages
// @Accept json
// @Produce json
// @Param message body Message true "Message containing the light control command"
// @Success 200 {object} Response "Successfully processed the message"
// @Failure 405 {object} Response "Method not allowed"
// @Failure 500 {object} Response "Internal server error"
// @Router /messages [post]
func ReceiveMessage(w http.ResponseWriter, r *http.Request) {
	// Start a new span for this request
	ctx, span := utils.StartSpan(r.Context(), "receive_message")
	defer span.End()

	// Set content type for all responses
	w.Header().Set("Content-Type", "application/json")
	
	// Only allow POST method
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(Response{
			Status:  "error",
			Message: "Method not allowed",
			RGB:     currentRGB,
		})
		utils.LogEvent(ctx, "method_not_allowed", map[string]interface{}{
			"method": r.Method,
		})
		return
	}
	
	var data Message
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Response{
			Status:  "error",
			Message: "Invalid request body: " + err.Error(),
			RGB:     currentRGB,
		})
		utils.LogEvent(ctx, "invalid_request", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	utils.LogEvent(ctx, "message_received", map[string]interface{}{
		"message": data.Message,
	})

	// parse message to rgb using openai
	rgb, err := utils.ParseToRGB(data.Message, currentRGB)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Status:  "error",
			Message: "Failed to parse message: " + err.Error(),
			RGB:     currentRGB,
		})
		utils.LogEvent(ctx, "parse_error", map[string]interface{}{
			"error": err.Error(),
		})
		return	
	}

	// update to new rgb values
	currentRGB = rgb

	// Publish RGB values to MQTT
	log.Printf("Publishing RGB values: %v", rgb)
	if err := utils.PublishRGB(rgb); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Response{
			Status:  "error",
			Message: "Failed to send RGB values: " + err.Error(),
			RGB:     rgb,
		})
		utils.LogEvent(ctx, "publish_error", map[string]interface{}{
			"error": err.Error(),
			"rgb":   rgb,
		})
		return
	}

	utils.LogEvent(ctx, "rgb_published", map[string]interface{}{
		"rgb": rgb,
	})

	// Success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Status:  "success",
		Message: "Message sent successfully",
		RGB:     rgb,
	})
} 