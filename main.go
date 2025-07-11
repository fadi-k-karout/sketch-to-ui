package main

import (
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"sketch-to-ui-final-proj/ai"
	"sketch-to-ui-final-proj/auth"
	"sketch-to-ui-final-proj/sketch"
	uicomponents "sketch-to-ui-final-proj/ui-components"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		slog.Warn("Error loading .env file")
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
	router.Static("/static", "./static")

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))

	secretKey := os.Getenv("SESSION_SECRET_KEY") // Get secret key from env
	if secretKey == "" {
		log.Fatal("SESSION_SECRET_KEY is not set in .env") // Important: Secret key is needed!
	}

	auth.Init(router, db, secretKey) // Initialize auth with the database
	sketchStore := sketch.SetupSketch(router)

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	baseURL := os.Getenv("OPENROUTER_BASE_URL")
	client := &http.Client{
		Timeout: 30 * time.Second, // Set a timeout of 30 seconds

	}

	aiProvider := ai.NewOpenRouterProvider(apiKey, baseURL, client)
	uicomponents.SetupComponents(router, db, sketchStore, aiProvider)

	router.GET("/", func(c *gin.Context) {
		isLoggedIn, _ := c.Get("isLoggedIn")
		avatarURI, _ := c.Get("avatarURI")
		user, _ := c.Get("user")

		var firstName, lastName string
		if authUser, ok := user.(*auth.User); ok && authUser != nil {
			firstName = authUser.FirstName
			lastName = authUser.LastName
		}
		slog.Info("Rendering home page", "isLoggedIn", isLoggedIn, "avatarURI", avatarURI, "firstName", firstName, "lastName", lastName)

		c.HTML(http.StatusOK, "base", gin.H{
			"title":      "Home",
			"avatarPath": avatarURI,
			"isLoggedIn": isLoggedIn,
			"firstName":  firstName,
			"lastName":   lastName,
		})
	})

	router.GET("/navbar", func(c *gin.Context) {
		isLoggedIn, _ := c.Get("isLoggedIn")
		avatarURI, _ := c.Get("avatarURI")
		user, _ := c.Get("user")

		var firstName, lastName string
		if authUser, ok := user.(*auth.User); ok && authUser != nil {
			firstName = authUser.FirstName
			lastName = authUser.LastName
		}

		c.Header("HX-Push-Url", "/")
		c.HTML(http.StatusOK, "navbar", gin.H{
			"isLoggedIn": isLoggedIn,
			"avatarPath": avatarURI,
			"firstName":  firstName,
			"lastName":   lastName,
		})
	})

	log.Println("Listening on :3000")
	router.Run(":3000")
}
