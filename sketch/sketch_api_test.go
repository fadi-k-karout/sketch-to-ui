package sketch

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestUploadSketchHandler_ValidImage tests the image upload handler with a valid image.
func TestUploadSketchHandler_ValidImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a dummy valid PNG header (first 8 bytes)
	validImage := []byte("\x89\x50\x4E\x47\x0D\x0A\x1A\x0A")
	// Append some extra content to simulate a file.
	validImage = append(validImage, []byte("dummydata")...)

	// Ensure the dummy image is at least 265 bytes, append zero bytes if necessary.
	if len(validImage) < 265 {
		padding := make([]byte, 265-len(validImage))
		validImage = append(validImage, padding...)
	}

	// Create a multipart form file.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("sketch", "test.png")

	assert.NoError(t, err)
	_, err = part.Write(validImage)
	assert.NoError(t, err)
	writer.Close()

	// Prepare the test HTTP request.
	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Create a gin context with the request.
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Simulate user ID in context using a test-specific key.
	c.Set("userID", 123)

	// Initialize the SketchStore and handler.
	sketchStore := NewSketchStore()
	handler := uploadSketchHandler(sketchStore)

	// Call the handler.
	handler(c)

	// Assert success (200 OK).
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestUploadSketchHandler_NoFile tests the image upload handler with no file provided.
func TestUploadSketchHandler_NoFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a multipart form without the "sketch" field.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("dummy", "value")
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// Simulate user ID in context using a test-specific key.
	c.Set("userID", 123)

	sketchStore := NewSketchStore()
	handler := uploadSketchHandler(sketchStore)

	handler(c)

	// Expect a 400 Bad Request because no file is uploaded.
	assert.Equal(t, http.StatusBadRequest, w.Code)
}


