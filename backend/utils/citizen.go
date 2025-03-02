package utils

import (
	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"app/db"
)


func CreateCitizen(c *gin.Context) {
	type CreateCitizenRequest struct {
		Email    string `json:"email" binding:"required,email"`
		Name     string `json:"name" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
	}
	var req CreateCitizenRequest
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

	// Create citizen record
	newCitizen := db.Citizen{
		Email:    req.Email,
		Password: string(hashedPassword),
		Name:     req.Name,
	}

	// Save to database
	if result := db.DB.Create(&newCitizen); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create citizen",
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "User created",
	})
}

func GetAllCitizen(c *gin.Context) {
	var citizens []db.Citizen
	db.DB.Preload("StolenBikes").Find(&citizens)
	c.JSON(http.StatusOK, gin.H{
		"data": citizens,
	})
}