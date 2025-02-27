package main

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "net/http"
	"golang.org/x/crypto/bcrypt"
)

func setupAPIRoutes(r *gin.Engine) {
    // Create an api group/router
    api := r.Group("/api")
    {
        police := api.Group("/police")
        police.GET("/", getAllPolice)
        police.POST("/", createPolice)

    }
    login := r.Group("/login")
    {
        login.POST("police")
        login.POST("citizen")
    }
}

// Handler functions
func getAllPolice(c *gin.Context) {
    var polices []Police
    DB.Find(&polices)
    c.JSON(http.StatusOK, gin.H{
        "data": polices,
    })
}

func getUser(c *gin.Context) {
    id := c.Param("id")
    c.JSON(http.StatusOK, gin.H{
        "id": id,
        "name": "Sample User",
    })
}

func createPolice(c *gin.Context) {
    type CreatePoliceRequest struct {
        Email    string `json:"email" binding:"required,email"`
        Name     string `json:"name" binding:"required"`
        Password string `json:"password" binding:"required,min=6"`
    }
    var req CreatePoliceRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": err.Error(),
        })
        return
    }
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("failed to hash password: %w", err)
	}
	
	// Create police record
	newPolice := Police{
		Email:    req.Email,
		Password: string(hashedPassword),
		Name:     req.Name,
	}
	
	// Save to database
	if result := DB.Create(&newPolice); result.Error != nil {
		fmt.Println("failed to create police record: %w", result.Error)
	}
    c.JSON(http.StatusCreated, gin.H{
        "message": "User created",
    })
}

func updateUser(c *gin.Context) {
    id := c.Param("id")
    c.JSON(http.StatusOK, gin.H{
        "message": "User " + id + " updated",
    })
}

func deleteUser(c *gin.Context) {
    id := c.Param("id")
    c.JSON(http.StatusOK, gin.H{
        "message": "User " + id + " deleted",
    })
}

func getAllPosts(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "posts": []string{"post1", "post2"},
    })
}