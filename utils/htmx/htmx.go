package htmx

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
)

type ToastLevel string

const (
	InfoLevel  ToastLevel = "info"
	ErrorLevel ToastLevel = "error"
)

func TriggerToast(c *gin.Context, level ToastLevel, message string) {
	payload := map[string]interface{}{
		"showMessage": map[string]string{
			"level":   string(level),
			"message": message,
		},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		// Optionally log error
		return
	}

	c.Header("HX-Trigger", string(jsonPayload))
}