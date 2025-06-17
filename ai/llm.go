package ai

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
)

type UIGenerationResponse struct {
	Components      []UIComponentDTO `json:"components"`
	FailureResponse string           `json:"failure response,omitempty"`
}

// UIComponent represents a UI component with its title, type, and code.
type UIComponentDTO struct {
	Title string `json:"title"`
	Type  string `json:"type"`
	Code  string `json:"code"`
}

//go:embed prompts/system_prompt.txt
var systemPrompt string

// GenerateUICode generates UI code from a user prompt and image using OpenRouter's LLM.
// It sends the prompt and base64-encoded image to the specified model and returns
// a structured UIGenerationResponse containing the generated UI code.
func GenerateUICode(ctx context.Context, userPrompt string, imageBase64URI string, op *OpenRouterProvider) (UIGenerationResponse, error) {
	messages := []map[string]any{
		{"role": "system", "content": systemPrompt},
		{"role": "user", "content": []map[string]any{{"type": "text", "text": userPrompt}, {
			"type": "image_url",
			"image_url": map[string]string{
				"url": imageBase64URI,
			},
		}},
		}}
	modelID := "opengvlab/internvl3-14b:free"
	response, err := op.RequestChatCompletion(ctx, messages, modelID)
	if err != nil {
		return UIGenerationResponse{}, fmt.Errorf("failed to generate response from OpenRouter: %w", err)
	}

	cleanResponse := cleanLLMResponse(response)

	var uiGenResp UIGenerationResponse
	err = json.Unmarshal([]byte(cleanResponse), &uiGenResp)
	if err != nil {
		slog.Debug("Failed to parse UI Generation Response", "error", err, "response", cleanResponse)
		return UIGenerationResponse{}, fmt.Errorf("failed to parse response JSON: %w", err)
	}
	

	return uiGenResp, nil
}

// cleanLLMResponse sanitizes and formats raw LLM response text by removing
// markdown code blocks, fixing escaped characters, and validating JSON structure.
func cleanLLMResponse(response string) string {
	// Log the initial response
	slog.Debug("Initial response", "response", response)

	// Remove leading/trailing whitespace
	response = strings.TrimSpace(response)

	// Remove markdown code block markers if present
	if strings.HasPrefix(response, "```") {
		lines := strings.Split(response, "\n")
		if len(lines) > 1 {
			// Skip first line (```json or similar) and last line (```)
			lines = lines[1 : len(lines)-1]
			response = strings.Join(lines, "\n")
		}
	}

	// First pass: Fix basic formatting
	response = strings.ReplaceAll(response, "\n", " ") // Replace newlines with spaces
	response = strings.ReplaceAll(response, "\r", "")  // Remove carriage returns

	// Second pass: Handle code field escaping
	// This specifically targets the issue with double escaped newlines \\\\n
	response = strings.ReplaceAll(response, "\\\\n", "\\n")
	response = strings.ReplaceAll(response, "\\\\\"", "\\\"")

	// Log the cleaned response
	slog.Debug("Cleaned response", "response", response)

	// Validate JSON structure
	var js json.RawMessage
	if err := json.Unmarshal([]byte(response), &js); err != nil {
		slog.Error("Invalid JSON after cleaning", "error", err)
	}

	return response
}
