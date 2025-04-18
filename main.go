package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"gitlab.com/sketch-to-ui-final-proj/auth" // Replace with the correct import path
)

func main() {
	db, err := sql.Open("postgres", "postgres://user:pass@localhost:5432/yourdb?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.LoadHTMLGlob("templates/**/*")

	auth.Init(router, db)

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to the home page!")
	})

	log.Println("Listening on :3000")
	router.Run(":3000")
}
