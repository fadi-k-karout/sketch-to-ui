package auth

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions" // Import gorilla/sessions
)

// SessionStore is now using gorilla/sessions.CookieStore
type SessionStore struct {
	store           *sessions.CookieStore // Use gorilla/sessions CookieStore
	cookieName      string
	sessionDuration time.Duration
}

func NewSessionStore(cookieName string, sessionDuration time.Duration, secretKey string, isProduction bool) *SessionStore {
	cookieStore := sessions.NewCookieStore([]byte(secretKey)) // Initialize gorilla/sessions CookieStore
	// Configure cookie options (optional, but good practice)
	cookieStore.Options = &sessions.Options{
		HttpOnly: true,
		Secure:   isProduction, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   int(sessionDuration.Seconds()),
	}

	return &SessionStore{
		store:           cookieStore,
		cookieName:      cookieName,
		sessionDuration: sessionDuration,
	}
}

// SetSession creates a new session and sets the session cookie using gorilla/sessions.
func (s *SessionStore) SetSession(c *gin.Context, user *User) {
	session, err := s.store.Get(c.Request, s.cookieName) // Get session from gorilla/sessions
	if err != nil {
		slog.Error("Error getting session", "err", err)
		return // Handle error, maybe create a new session
	}

	session.Values["userID"] = int(user.ID) // Set user ID in session values
	session.Values["avatarURI"] = user.AvatarURI

	err = session.Save(c.Request, c.Writer) // Save session using gorilla/sessions
	if err != nil {
		// Handle save error, maybe log it. For now, continue.
		slog.Error("Error saving session", "err", err)
	}
}

// GetSession retrieves the user ID from the session cookie using gorilla/sessions.
func (s *SessionStore) GetSession(c *gin.Context) (*User, bool) {
	var user User
	session, err := s.store.Get(c.Request, s.cookieName)
	if err != nil {
		return nil, false
	}

	userIDValue := session.Values["userID"]
	if userIDValue == nil {
		return nil, false
	}
	userID, ok := userIDValue.(int)
	if !ok {
		return nil, false
	}
	user.ID = ID(userID)

	// Restore AvatarURI
	if avatar, ok := session.Values["avatarURI"].(string); ok {
		user.AvatarURI = avatar
	}

	return &user, true
}

// ClearSession removes the session cookie and server-side session data using gorilla/sessions.
func (s *SessionStore) ClearSession(c *gin.Context) {
	session, err := s.store.Get(c.Request, s.cookieName) // Get session from gorilla/sessions
	if err != nil {
		return // Session likely doesn't exist, nothing to clear
	}

	session.Options.MaxAge = -1             // Set MaxAge to -1 to delete the cookie
	err = session.Save(c.Request, c.Writer) // Save session to delete cookie
	if err != nil {
		slog.Error("Error clearing session", "err", err)
		return
	}
	// gorilla/sessions handles server-side clearing of cookie-based sessions automatically
}

// SessionMiddleware is a Gin middleware to handle session management.
func SessionMiddleware(sessionStore *SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, isLoggedIn := sessionStore.GetSession(c)
		if isLoggedIn {
			c.Set("userID", user.ID)
			c.Set("avatarURI", user.AvatarURI)
		}
		c.Set("isLoggedIn", isLoggedIn)
		c.Next()
	}
}
