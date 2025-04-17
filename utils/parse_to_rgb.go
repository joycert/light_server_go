package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// RGB represents an RGB color value
// @Description RGB color value with red, green, and blue components
type RGB struct {
	R int `json:"r"` // Red component (0-255)
	G int `json:"g"` // Green component (0-255)
	B int `json:"b"` // Blue component (0-255)
}

// OpenAI request structure
type OpenAIRequest struct {
	Model       string        `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAI response structure
type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// ParseToRGB parses a text message to RGB values using OpenAI
func ParseToRGB(message string, currentRGB RGB) (RGB, error) {
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		return RGB{}, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	// Create the prompt similar to the Python version
	prompt := fmt.Sprintf(`Given the following current light color and message, return the new light color in an RGB value format.

The lights are very bright, so colors with high RGB values (e.g., 255, 240, 240) appear as bright white. To maintain distinguishable colors, keep individual RGB values **below 100** unless an explicitly bright color is requested.

- Soft or warm colors should use **low intensity** (e.g., soft white should be around (80, 40, 30)).
- Saturated colors should still be distinguishable at lower brightness levels.
- Avoid using values above **100-120** unless the message specifies high brightness.

Current light color: {"r":%d,"g":%d,"b":%d}
Message: %s
Return only a JSON object with r, g, b keys.`, currentRGB.R, currentRGB.G, currentRGB.B, message)

	openaiReq := OpenAIRequest{
		Model: "gpt-4o-mini",
		Messages: []OpenAIMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.2,
		MaxTokens:   100,
	}

	jsonData, err := json.Marshal(openaiReq)
	if err != nil {
		return RGB{}, fmt.Errorf("failed to marshal OpenAI request: %w", err)
	}

	// Make the OpenAI API request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return RGB{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+openaiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return RGB{}, fmt.Errorf("failed to make OpenAI request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return RGB{}, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return RGB{}, fmt.Errorf("OpenAI API returned non-200 status: %d, body: %s", resp.StatusCode, string(body))
	}

	var openaiResp OpenAIResponse
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return RGB{}, fmt.Errorf("failed to unmarshal OpenAI response: %w", err)
	}

	if len(openaiResp.Choices) == 0 {
		return RGB{}, fmt.Errorf("no choices in OpenAI response")
	}

	// Clean up the response content - remove markdown formatting
	content := openaiResp.Choices[0].Message.Content
	// Clean up any markdown formatting
	for _, marker := range []string{"```json", "```", "`"} {
		content = replaceAll(content, marker, "")
	}
	content = trimSpace(content)

	// Parse the RGB values
	var rgb RGB
	if err := json.Unmarshal([]byte(content), &rgb); err != nil {
		return RGB{}, fmt.Errorf("failed to parse RGB from response: %w", err)
	}

	return rgb, nil
}

// clean up the string
func replaceAll(s, old, new string) string {
	for {
		if s2 := replace(s, old, new); s2 == s {
			return s
		} else {
			s = s2
		}
	}
}

func replace(s, old, new string) string {
	var result string
	for i := 0; i < len(s); {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			result += new
			i += len(old)
		} else {
			result += string(s[i])
			i++
		}
	}
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\n' || s[start] == '\t' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\n' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
} 