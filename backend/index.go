package main

import (
	"fmt"
    "app/db"
    "app/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	db.Database()
	r := gin.Default()

	api.SetupAPIRoutes(r)

	port := ":8080"

	fmt.Println("Server starting on" + port)

	err := r.Run(port)
	if err != nil {
		fmt.Println(err)
	}
}
