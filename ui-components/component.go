package uicomponents

import (
	"database/sql"
	"sketch-to-ui-final-proj/ai"
	"sketch-to-ui-final-proj/sketch"
	"time"

	"github.com/gin-gonic/gin"
)

type UIComponent struct {
	ID         int       `db:"id" json:"ID"`
	Title      string    `db:"title" json:"Title"`
	Type       string    `db:"type" json:"Type"`
	Code       string    `db:"code" json:"Code"`
	IsPublic   bool      `db:"is_public"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
	ArchivedAt time.Time `db:"archived_at"`
	UserID     int       `db:"user_id"`
}

// PublicComponentWithUser holds a public component and its owner's name
type PublicComponentWithUser struct {
	ID         int
	Title      string
	Type       string
	Code       string
	IsPublic   bool
	UserID     int
	FirstName  string
	LastName   string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ArchivedAt time.Time
}


func SetupComponents(router *gin.Engine ,db *sql.DB, sketchStore *sketch.SketchStore, openRouterProvider *ai.OpenRouterProvider){


	componentStore := NewUIComponentsStore(db)
	componentHandler := NewUIComponentHandler(componentStore, sketchStore,  openRouterProvider)

	componentHandler.RegisterRoutes(router)


}
