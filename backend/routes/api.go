package api

import (

	"github.com/gin-gonic/gin"
	"app/utils"
)


func SetupAPIRoutes(r *gin.Engine) {
	// Create an api group/router
	api := r.Group("/api")
	{
		police := api.Group("/police")
		police.GET("/", utils.GetAllPolice)
		police.POST("/", utils.CreatePolice)
		police.PATCH("/:id", utils.UpdatePolice)
		police.DELETE("/:id", utils.DeletePolice)
	}
	citizen := api.Group("/citizen")
	{
		citizen.GET("/", utils.GetAllCitizen)
		citizen.POST("/", utils.CreateCitizen)
	}
	bike := api.Group("/bike")
	bike.Use((utils.AuthMiddleware()))
	{
		bike.GET("/", utils.GetAllBikes)
		bike.POST("/", utils.CreateBike)
		bike.GET("/found/:id", utils.FoundBike)
	}
	login := r.Group("/login")
	{
		login.POST("citizen", utils.LoginCitizen)
	}
}







