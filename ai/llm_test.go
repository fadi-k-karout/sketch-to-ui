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

	// Load a sample image (replace with a real image path if needed)
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
