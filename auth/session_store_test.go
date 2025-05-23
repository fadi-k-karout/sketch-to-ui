package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupSessionStore() *SessionStore {
	return NewSessionStore("testsession", time.Hour, "test-secret-key", false)
}

func TestSetAndGetSession(t *testing.T) {
	store := setupSessionStore()
	router := gin.New()

	// Handler to set session
	router.GET("/set", func(c *gin.Context) {
		user := &User{ID: 42, AvatarURI: "/avatar.png"}
		store.SetSession(c, user)
		c.String(200, "session set")
	})

	// Handler to get session
	router.GET("/get", func(c *gin.Context) {
		user, ok := store.GetSession(c)
		if !ok {
			c.String(401, "no session")
			return
		}
		c.String(200, "%d|%s", user.ID, user.AvatarURI)
	})

	// Set session and capture cookie
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/set", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	cookies := w.Result().Cookies()
	assert.NotEmpty(t, cookies)

	// Use cookie to get session
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/get", nil)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}
	router.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)
	assert.Contains(t, w2.Body.String(), "42|/avatar.png")
}

func TestClearSession(t *testing.T) {
	store := setupSessionStore()
	router := gin.New()

	// Set session
	router.GET("/set", func(c *gin.Context) {
		user := &User{ID: 99}
		store.SetSession(c, user)
		c.String(200, "session set")
	})

	// Clear session
	router.GET("/clear", func(c *gin.Context) {
		store.ClearSession(c)
		c.String(200, "session cleared")
	})

	// Get session
	router.GET("/get", func(c *gin.Context) {
		_, ok := store.GetSession(c)
		if !ok {
			c.String(401, "no session")
			return
		}
		c.String(200, "session exists")
	})

	// Set session and capture cookie
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/set", nil)
	router.ServeHTTP(w, req)
	cookies := w.Result().Cookies()

	// Clear session
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/clear", nil)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}
	router.ServeHTTP(w2, req2)

	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/get", nil)
	// Do NOT add any cookies
	router.ServeHTTP(w3, req3)
	assert.Equal(t, 401, w3.Code)
	assert.Contains(t, w3.Body.String(), "no session")
}

func TestSessionMiddleware(t *testing.T) {
	store := setupSessionStore()
	router := gin.New()
	router.Use(SessionMiddleware(store))

	// Set session
	router.GET("/set", func(c *gin.Context) {
		user := &User{ID: 7, AvatarURI: "/avatar7.png"}
		store.SetSession(c, user)
		c.String(200, "session set")
	})

	// Protected route
	router.GET("/protected", func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.String(401, "not logged in")
			return
		}
		avatarURI, _ := c.Get("avatarURI")
		c.String(200, "%v|%v", userID, avatarURI)
	})

	// Set session and capture cookie
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/set", nil)
	router.ServeHTTP(w, req)
	cookies := w.Result().Cookies()

	// Access protected route
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/protected", nil)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}
	router.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)
	assert.Contains(t, w2.Body.String(), "7|/avatar7.png")
}
