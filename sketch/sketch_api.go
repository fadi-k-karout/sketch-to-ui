package sketch

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"sketch-to-ui-final-proj/auth"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// isImageFile checks if the given file is an image by attempting to decode its header.
func isImageFile(file *multipart.FileHeader) (bool, error) {
	if file == nil {
		return false, nil
	}

	// Open the file
	f, err := file.Open()
	if err != nil {
		slog.Error("Error opening file", slog.Any("error", err)) // {{ edit_2 }} Use slog.Error
		return false, fmt.Errorf("error opening file: %w", err)  // Wrap error
	}
	defer f.Close()

	// Read the file header to determine the image type
	buf := make([]byte, 265) // Read enough bytes to determine the image type
	_, err = io.ReadFull(f, buf)
	if err != nil && err != io.EOF {
		slog.Error("Error reading file header", slog.Any("error", err)) // {{ edit_3 }} Use slog.Error
		return false, fmt.Errorf("error reading file header: %w", err)  // Wrap error
	}

	// Check for known image headers
	switch {
	case bytes.HasPrefix(buf, []byte("\x89\x50\x4E\x47\x0D\x0A\x1A\x0A")): // PNG
		return true, nil
	case bytes.HasPrefix(buf, []byte("\xFF\xD8")): // JPEG
		return true, nil
	case bytes.HasPrefix(buf, []byte("\x42\x4D")): // BMP
		return true, nil

	}

	return false, nil
}

// uploadSketchHandler handles the upload of sketch files.
func uploadSketchHandler(sketchStore *SketchStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the multipart form
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse multipart form"})
			return
		}

		// Extract the uploaded file
		sketchFiles := form.File["sketch"]
		if len(sketchFiles) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No sketch file uploaded"})
			return
		}

		// Process each uploaded file
		for _, fileHeader := range sketchFiles {

			isImage, err := isImageFile(fileHeader) // Get boolean and error
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking file type"}) // Handle error
				slog.Error("File type check error", slog.Any("error", err))                       
				return
			}
			if !isImage {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Uploaded file is not an image"})
				return
			}

			// Open the file
			f, err := fileHeader.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
				slog.Error("Error opening file", slog.Any("error", err)) 
				return
			}
			defer f.Close()

			// Read the entire file contents
			buf := bytes.NewBuffer(nil)
			_, err = io.Copy(buf, f)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
				slog.Error("Error reading file", slog.Any("error", err)) 
				return
			}

			// Convert the file contents to base64
			base64Image := base64.StdEncoding.EncodeToString(buf.Bytes())

			// Store or process the base64 image as needed
			// For example, you can save it to a database or send it to another service
			// Example:
			// SaveToDatabase(base64Image)

			sketchID := uuid.New().String()
			userID, ok := auth.GetUserIDFromContext(c)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
				return
			}

			sketchStore.SetSketch(sketchID, &Sketch{
				ID:       sketchID,
				ImageURL: base64Image,
				OwnerID:  strconv.Itoa(userID),
			}, 24*time.Hour)

			c.JSON(http.StatusOK, gin.H{
				"message":   "File uploaded successfully",
				"sketch_id": sketchID,
			})
		}
	}
}

// sketchpadHandler serves the sketchpad HTML page.
func sketchpadHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "sketch", nil)
}

func RegisterRoutes(r *gin.Engine, sketchStore *SketchStore) {
	slog.Info("Registering Sketch Routes")

	r.GET("/sketchpad", auth.AuthRequiredMiddleware(), sketchpadHandler)
	r.POST("/upload", auth.AuthRequiredMiddleware(), uploadSketchHandler(sketchStore))
}
