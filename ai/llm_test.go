// ai/llm_test.go
package ai

import (
	"context"
	"encoding/base64"
	"log/slog"
	"net/http"
	"os"
	"time"

	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)



// TestGenerateUICode tests the GenerateUICode function by loading environment variables,
// creating an OpenRouter provider, reading a test image, and verifying that UI code
// generation works correctly with image input.
func TestGenerateUICode(t *testing.T) {
	err := godotenv.Load("../.env")
	require.NoError(t, err, "Failed to load .env file")
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	baseURL := os.Getenv("OPENROUTER_BASE_URL")

	if apiKey == "" || baseURL == "" {
		t.Skip("OPENROUTER_API_KEY not set, skipping integration test")
	}
	client := &http.Client{
		Timeout: 30 * time.Second, // Set a timeout of 30 seconds

	}

	openrouter := NewOpenRouterProvider(apiKey, baseURL, client)

	// Load a sample image 
	imagePath := "test_data/test_image1.png" // Ensure this file exists
	imgBytes, err := os.ReadFile(imagePath)
	require.NoError(t, err, "Failed to read test image")

	imageBase64 := base64.StdEncoding.EncodeToString(imgBytes)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second) // 2 minutes timeout for image processing
	defer cancel()

	userPrompt := "Analyze the following sketch image from the image url I sent and generate the corresponding UI component code (using  HTML and tailwindcss) in JSON format."
	uiCode, err := GenerateUICode(ctx, userPrompt, "data:image/png;base64,"+imageBase64, openrouter)

	assert.NoError(t, err, "GenerateUICode should not return an error")
	assert.NotEmpty(t, uiCode, "GenerateUICode should return a non-empty string")

	// Log the response for manual inspection
	slog.Info("Generated UI Code:", "code", uiCode)

}

// TestUpdateCode tests the UpdateCode function by loading environment variables,
// creating an OpenRouter provider, and verifying that code update works correctly.
func TestUpdateCode(t *testing.T) {
	err := godotenv.Load("../.env")
	require.NoError(t, err, "Failed to load .env file")
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	baseURL := os.Getenv("OPENROUTER_BASE_URL")

	if apiKey == "" || baseURL == "" {
		t.Skip("OPENROUTER_API_KEY not set, skipping integration test")
	}
	
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	openrouter := NewOpenRouterProvider(apiKey, baseURL, client)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Sample existing component to update
	oldComponent := `{
		"title": "Simple Button",
		"type": "button",
		"code": "<button class=\"px-4 py-2 bg-gray-500 text-white\">Click Me</button>"
	}`
	
	userPrompt := "Update the following button component to have a blue background and rounded corners. Here is the current component: " + oldComponent + ". Return the updated code in JSON format."
	codeUpdateResp, err := UpdateCode(ctx, userPrompt, openrouter)

	assert.NoError(t, err, "UpdateCode should not return an error")
	assert.NotEmpty(t, codeUpdateResp, "UpdateCode should return a non-empty response")

	// Log the response for manual inspection
	slog.Info("Code Update Response:", "response", codeUpdateResp)
}
