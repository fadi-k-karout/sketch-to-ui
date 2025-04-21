package auth

import (
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
        MaxAge: int(sessionDuration.Seconds()),
		
	}

	return &SessionStore{
		store:           cookieStore,
		cookieName:      cookieName,
		sessionDuration: sessionDuration,
	}
}

// SetSession creates a new session and sets the session cookie using gorilla/sessions.
func (s *SessionStore) SetSession(c *gin.Context, userID int) {
	session, err := s.store.Get(c.Request, s.cookieName) // Get session from gorilla/sessions
	if err != nil {
		// Handle error, maybe log it. For now, continue.
	}

	session.Values["userID"] = userID                           // Set user ID in session values


	err = session.Save(c.Request, c.Writer) // Save session using gorilla/sessions
	if err != nil {
		// Handle save error, maybe log it. For now, continue.
	}
}

// GetSession retrieves the user ID from the session cookie using gorilla/sessions.
func (s *SessionStore) GetSession(c *gin.Context) (int, bool) {
	session, err := s.store.Get(c.Request, s.cookieName) // Get session from gorilla/sessions
	if err != nil {
		return 0, false // No session or error
	}

	// Check if user ID exists and is of the correct type
	userIDValue := session.Values["userID"]
	if userIDValue == nil {
		return 0, false // User ID not in session
	}
	userID, ok := userIDValue.(int) // Type assertion to int
	if !ok {
		return 0, false // Type assertion failed
	}

	return userID, true
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
		// Handle save error, maybe log it. For now, continue.
	}
	// gorilla/sessions handles server-side clearing of cookie-based sessions automatically
}

// SessionMiddleware is a Gin middleware to handle session management.
func (s *SessionStore) SessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, isLoggedIn := s.GetSession(c)
		if isLoggedIn {
			// Make user ID available in context for subsequent handlers
			c.Set("userID", userID)
		}
		c.Next()
	}
}

