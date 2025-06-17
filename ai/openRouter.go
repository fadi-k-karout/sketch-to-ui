package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// OpenRouterProvider implements both LLMProvider and LLMStreamingProvider
// using the OpenRouter API. OpenRouter acts as a gateway to multiple LLM providers
// with a unified API and flexible model selection.
type OpenRouterProvider struct {
	// APIKey is the authentication key for the OpenRouter API
	APIKey string

	// BaseURL is the root URL for the OpenRouter API endpoints
	BaseURL string


	// Client is a reusable HTTP client for making API requests
	Client *http.Client
}

// NewOpenRouterProvider creates a new instance of OpenRouterProvider with default settings.
// It requires an API key for authentication with the OpenRouter service.
//
// Parameters:
//   - apiKey: The authentication key for the OpenRouter API
//
// Returns:
//   - *OpenRouterProvider: A configured provider instance with sensible defaults
func NewOpenRouterProvider(apiKey string, baseURL string, client *http.Client) *OpenRouterProvider {
	return &OpenRouterProvider{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Client:  client,
	}
}

// RequestChatCompletion makes a synchronous call to OpenRouter and returns a complete response.
// This method blocks until the entire response is generated.
//
// Parameters:
//   - ctx: Context for request cancellation and timeout
//   - messages: Array of message objects with role and content fields
//   - modelID: Model identifier to use for generation
//
// Returns:
//   - string: The generated response text
//   - error: Any error encountered during the request
func (p *OpenRouterProvider) RequestChatCompletion(ctx context.Context, messages []map[string]any, modelID string) (string, error) {
	requestID := fmt.Sprintf("%d", time.Now().UnixNano())
	slog.Info("Starting RequestChatCompletion", "requestID", requestID)

	url := fmt.Sprintf("%s/v1/chat/completions", p.BaseURL)

	payload := map[string]any{
		"model":    modelID,
		"messages": messages,
		"stream":   false,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.APIKey)

	resp, err := p.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("non-200 response: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", errors.New("no choices returned")
	}

	slog.Info("Completed RequestChatCompletion", "requestID", requestID)
	return result.Choices[0].Message.Content, nil
}
