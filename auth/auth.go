package auth

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/authboss/v3"
	abclientstate "github.com/volatiletech/authboss-clientstate"
)

var Auth *authboss.Authboss

func Init(router *gin.Engine, db *sql.DB) {
	Auth = authboss.New()

	Auth.Config.Paths.RootURL = "http://localhost:3000"
	Auth.Config.Paths.Mount = "/auth"
	Auth.Config.Paths.AuthLoginOK = "/"
	Auth.Config.Paths.RegisterOK = "/"
	Auth.Config.Paths.LogoutOK = "/"

	Auth.Config.Storage.Server = &UserStorer{DB: db}
	Auth.Config.Storage.SessionState = abclientstate.NewSessionStorer("ab_session", []byte("session-secret"), nil)
	Auth.Config.Storage.CookieState = abclientstate.NewCookieStorer([]byte("cookie-secret"), nil)

	Auth.Config.Modules.RegisterPreserveFields = []string{"email"}
	Auth.Config.Modules.LogoutMethod = "GET"

	if err := Auth.Init(); err != nil {
		log.Fatal(err)
	}

	// Mount Authboss into Gin using http.Handler
	router.Use(gin.WrapH(Auth.LoadClientStateMiddleware(router.Handler())))

	// Use http.ServeMux to serve Authboss routes
	authMux := http.NewServeMux()
	authbossRoutes := Auth.Config.Core.Router
	authMux.Handle("/auth/", http.StripPrefix("/auth", authbossRoutes))

	router.Any("/auth/*w", gin.WrapH(authMux))
}
