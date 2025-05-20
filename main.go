package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"sketch-to-ui-final-proj/auth"
	"sketch-to-ui-final-proj/sketch"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		log.Fatal("GIN_MODE is not set in .env")
	}

	gin.SetMode(ginMode) // Set Gin mode

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Database connection error:", err)
	}
	defer db.Close()

	router := gin.Default()
	router.LoadHTMLGlob("templates/**/*.html") // Load your templates

	secretKey := os.Getenv("SESSION_SECRET_KEY") // Get secret key from env
	if secretKey == "" {
		log.Fatal("SESSION_SECRET_KEY is not set in .env") // Important: Secret key is needed!
	}

	auth.Init(router, db, secretKey) // Initialize auth with the database
	sketch.SetupSketch(router)

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "base", gin.H{
			"title": "Home",
		})
	})

	log.Println("Listening on :3000")
	router.Run(":3000")
}
