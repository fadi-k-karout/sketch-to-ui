package sketch

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

type Sketch struct {
	ID string `json:"id"`
	ImageURL string `json:"image_url"`
	OwnerID  string `json:"owner_id"`
}

func SetupSketch(r *gin.Engine) {
	slog.Info("Setting up sketch")

	sketchStore := NewSketchStore()

	RegisterRoutes(r, sketchStore)

}
