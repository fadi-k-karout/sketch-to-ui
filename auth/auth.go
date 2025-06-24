package auth

import (
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// DB connection (you should initialize this in main.go)
var DB *sql.DB
var sessionStore *SessionStore // Session store instance

func Init(router *gin.Engine, db *sql.DB, secretKey string) {
	DB = db
	isProduction := os.Getenv("IS_PRODUCTION") == "true"
	sessionStore = NewSessionStore("session-name", 24*time.Hour, secretKey, isProduction)
	router.Use(SessionMiddleware(sessionStore))
	setupRoutes(router)
}

func setupRoutes(router *gin.Engine) {
	router.GET("/signup", signupFormHandler)
	router.POST("/signup", signupHandler)
	router.GET("/login", loginFormHandler)
	router.POST("/login", loginHandler)
	router.GET("/logout", logoutHandler)
	router.GET("/profile", AuthRequiredMiddleware(), profileHandler)
}

func signupFormHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "register", gin.H{})
}

func loginFormHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "login", gin.H{})
}

type SignupForm struct {
	FirstName string `form:"first_name" binding:"required" validate:"required"`
	LastName  string `form:"last_name" binding:"required" validate:"required"`
	Email     string `form:"email" binding:"required" validate:"required,email"`
	Password  string `form:"password" binding:"required" validate:"required,min=8"`
}

type LoginForm struct {
	Email    string `form:"email" binding:"required" validate:"required,email"`
	Password string `form:"password" binding:"required" validate:"required,min=8"`
}

func signupHandler(c *gin.Context) {
	var form SignupForm
	if err := c.ShouldBind(&form); err != nil {
		slog.Error("Signup error", "err", err)
		c.String(http.StatusBadRequest, "Bad request")
		return
	}

	user := User{
		FirstName: form.FirstName,
		LastName:  form.LastName,
		Email:     form.Email,
		Password:  form.Password,
	}

	avatarURI, err := user.generateAvatar()
	if err != nil {
		c.String(http.StatusInternalServerError, "Could not generate avatar")
		return
	}

	user.AvatarURI = avatarURI

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		c.String(http.StatusInternalServerError, "Could not hash password")
		return
	}
	user.Password = string(hashedPassword)

	// In the INSERT statement:
	err = DB.QueryRow(
		"INSERT INTO users (first_name, last_name, email, password, avatar_uri) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		user.FirstName, user.LastName, user.Email, user.Password, user.AvatarURI,
	).Scan(&user.ID)
	if err != nil {
		log.Println("Signup error:", err)
		c.String(http.StatusInternalServerError, "Signup failed")
		return
	}
	// **Set Session on successful login**
	sessionStore.SetSession(c, &user)
	if c.GetHeader("HX-Request") == "true" {
		c.Header("HX-Trigger", "navbarChanged")

		c.HTML(http.StatusOK, "welcome", gin.H{
			"firstName":  user.FirstName,
			"userID":     user.ID,
			"avatarPath": user.AvatarURI,
		})
		return
	}

}

func loginHandler(c *gin.Context) {
	var form LoginForm
	if err := c.ShouldBind(&form); err != nil {
		slog.Error("Login error", "err", err)
		c.String(http.StatusBadRequest, "Bad request")
		return
	}

	user := User{
		Email:    form.Email,
		Password: form.Password,
	}

	slog.Info("User attempting login", "user", user)

	var storedUser User
	err := DB.QueryRow("SELECT id, first_name, email, password, avatar_uri FROM users WHERE email = $1", user.Email).Scan(&storedUser.ID, &storedUser.FirstName, &storedUser.Email, &storedUser.Password, &storedUser.AvatarURI)
	if err != nil {
		if err == sql.ErrNoRows {
			c.String(http.StatusUnauthorized, "Invalid credentials")
		} else {
			c.String(http.StatusInternalServerError, "Login failed")
		}
		return
	}
	slog.Info("Stored user found", "storedUser", storedUser)

	err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password))
	if err != nil {
		slog.Error("Login error", "err", err)
		c.String(http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// **Set Session on successful login**
	sessionStore.SetSession(c, &storedUser)
	if c.GetHeader("HX-Request") == "true" {
		c.Header("HX-Trigger", "navbarChanged")

		c.HTML(http.StatusOK, "welcome", gin.H{
			"firstName":  storedUser.FirstName,
			"userID":     storedUser.ID,
			"avatarPath": storedUser.AvatarURI,
		})
		return
	}
}

func logoutHandler(c *gin.Context) {
	sessionStore.ClearSession(c)
	if c.GetHeader("HX-Request") == "true" {
		c.Header("HX-Redirect", "/")
		c.Status(http.StatusSeeOther)
		return
	}
	c.Redirect(http.StatusSeeOther, "/") // Redirect to home page or login page
}

func AuthRequiredMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		isLoggedIn, _ := c.Get("isLoggedIn")
		if loggedIn, ok := isLoggedIn.(bool); !ok || !loggedIn {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

// profileHandler is an example protected handler.
func profileHandler(c *gin.Context) {
	userID, _ := GetUserIDFromContext(c) // User ID is guaranteed to exist due to middleware
	c.String(http.StatusOK, "Profile page for user ID: %d", userID)
}

// GetUserIDFromContext retrieves the user ID from the Gin context.
func GetUserIDFromContext(c *gin.Context) (int, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, false
	}
	if id, ok := userID.(ID); ok {
		return int(id), true
	}
	return 0, false // Type assertion failed
}
