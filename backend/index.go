package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	Database()
	r := gin.Default()

	setupAPIRoutes(r)

	// NoRoute will handle any unmatched routes
	r.NoRoute(func(c *gin.Context) {
		// Try to serve static file
		fileServer := http.FileServer(http.Dir("./public"))
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	// Set up port based on environment
	port := ":80"
	if os.Getenv("GO_ENV") == "development" {
		port = ":8080"
	}

	fmt.Println("Server starting on" + port)

	err := r.Run(port)
	if err != nil {
		fmt.Println(err)
	}
}
