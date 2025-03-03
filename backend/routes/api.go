package api

import (

	"github.com/gin-gonic/gin"
	"app/utils"
)


func SetupAPIRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		police := api.Group("/police")
		police.POST("/", utils.CreatePolice)
		police.Use(utils.AuthMiddlewarePolice())
		police.GET("/", utils.GetAllPolice)
		police.PATCH("/:id", utils.UpdatePolice)
		police.DELETE("/:id", utils.DeletePolice)
	}
	citizen := api.Group("/citizen")
	{
		citizen.GET("/", utils.GetAllCitizen)
		citizen.POST("/", utils.CreateCitizen)
	}
	bike := api.Group("/bike")
	bike.Use((utils.AuthMiddlewareCitizen()))
	{
		bike.GET("/", utils.GetAllBikes)
		bike.GET("/create", utils.CreateBike)
		bike.GET("/found/:id", utils.FoundBike)
	}
	login := r.Group("/login")
	{
		login.POST("/citizen", utils.LoginCitizen)
		login.POST("/police", utils.LoginPolice)
	}
}







