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
	router.Use(sessionStore.SessionMiddleware())
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

type AuthForm struct {
	FirstName string `form:"first_name" binding:"required" validate:"required"`
	Email    string `form:"email" binding:"required" validate:"required,email"`
	Password string `form:"password" binding:"required" validate:"required,min=8"`
}

func signupHandler(c *gin.Context) {
	var form AuthForm
	if err := c.ShouldBind(&form); err != nil {
		slog.Error("Signup error:", err)
		c.String(http.StatusBadRequest, "Bad request")
		return
	}

	

	user := User{
		FirstName: form.FirstName,
		Email:    form.Email,
		
	}

	avatarURI, err := user.generateAvatar()
	if err != nil {
		c.String(http.StatusInternalServerError, "Could not generate avatar")
		return
	}

	user.AvatarURI = avatarURI

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.String(http.StatusInternalServerError, "Could not hash password")
		return
	}
	user.Password = string(hashedPassword)

	_, err = DB.Exec("INSERT INTO users (email, password, avatar_uri) VALUES ($1, $2, $3)", user.Email, user.Password, user.AvatarURI)
	if err != nil {
		log.Println("Signup error:", err)
		c.String(http.StatusInternalServerError, "Signup failed")
		return
	}

	c.String(http.StatusOK, "Signup successful") // Redirect to login or home page later
}

func loginHandler(c *gin.Context) {
	var form AuthForm
	if err := c.ShouldBind(&form); err != nil {
		
		c.String(http.StatusBadRequest, "Bad request")
		return
	}

	user := User{
		Email:    form.Email,
		Password: form.Password,
	}

	var storedUser User
	err := DB.QueryRow("SELECT id, email, password, avatar_uri FROM users WHERE email = $1", user.Email).Scan(&storedUser.ID, &storedUser.Email, &storedUser.Password, &storedUser.AvatarURI)
	if err != nil {
		if err == sql.ErrNoRows {
			c.String(http.StatusUnauthorized, "Invalid credentials")
		} else {
			c.String(http.StatusInternalServerError, "Login failed")
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password))
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid credentials")
		return	
	}

	// **Set Session on successful login**
	sessionStore.SetSession(c, int(storedUser.ID))
	c.HTML(http.StatusOK, "base", gin.H{
		"userID":     storedUser.ID,
		"avatarPath": storedUser.AvatarURI,
	}) 
}

func logoutHandler(c *gin.Context) {
	sessionStore.ClearSession(c)
	c.Redirect(http.StatusSeeOther, "/") // Redirect to home page or login page
}

func AuthRequiredMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, isLoggedIn := sessionStore.GetSession(c)
		if !isLoggedIn {
			c.Redirect(http.StatusSeeOther, "/login") // Redirect to login if not logged in
			c.Abort()
			return
		}
		c.Set("userID", userID) // Make user ID available in context
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
	if id, ok := userID.(int); ok {
		return id, true
	}
	return 0, false // Type assertion failed
}
