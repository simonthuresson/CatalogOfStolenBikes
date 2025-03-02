package utils

import (
	"fmt"
	"net/http"
	"golang.org/x/crypto/bcrypt"
	"github.com/gin-gonic/gin"
	"app/db"
)

// Handler functions
func GetAllPolice(c *gin.Context) {
	var polices []db.Police
	db.DB.Preload("AssignedBike").Find(&polices)
	c.JSON(http.StatusOK, gin.H{
		"data": polices,
	})
}

func UpdatePolice(c *gin.Context) {
	id := c.Param("id")

	// Define request structure
	type UpdateRequest struct {
		Name     string `json:"name" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
	}

	// Parse request
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to hash password",
		})
		return
	}

	// Find the police officer
	var police db.Police
	if result := db.DB.First(&police, id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Police officer not found",
		})
		return
	}

	// Update the record
	result := db.DB.Model(&police).Updates(db.Police{
		Name:     req.Name,
		Password: string(hashedPassword),
	})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Police officer updated successfully",
		"data": gin.H{
			"id":    police.ID,
			"name":  police.Name,
			"email": police.Email,
		},
	})
}

func CreatePolice(c *gin.Context) {
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
	newPolice := db.Police{
		Email:    req.Email,
		Password: string(hashedPassword),
		Name:     req.Name,
	}

	// Save to database
	if result := db.DB.Create(&newPolice); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create police",
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "User created",
	})
}

func DeletePolice(c *gin.Context) {
	id := c.Param("id")
	if err := db.DB.Model(&db.Bike{}).Where("police_id = ?", id).Update("police_id", nil).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update assigned bikes",
		})
		return
	}
	db.DB.Delete(&db.Police{}, id)
	c.JSON(http.StatusOK, gin.H{
		"message": "Police " + id + " deleted",
	})
}
