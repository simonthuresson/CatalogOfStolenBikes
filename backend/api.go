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
        police.PATCH("/:id", updatePolice)
        police.DELETE("/:id", deletePolice)
    }
    citizen := api.Group("/citizen")
    {
        citizen.GET("/", getAllCitizen)
        citizen.POST("/", createCitizen)
    }
    bike := api.Group("/bike")
    {
        bike.GET("/", getAllBikes)
        bike.POST("/", createBike)
    }
    login := r.Group("/login")
    {
        login.POST("police")
        login.POST("citizen")
    }
}

func getAllBikes(c *gin.Context) {
    var bikes []Bike
    DB.Find(&bikes)
    c.JSON(http.StatusOK, gin.H{
        "bikes": bikes,
    })
}

func createBike(c *gin.Context) {
    type ReportStolenBikeRequest struct {
        Description string `json:"description" binding:"required"`
        CitizenID   uint   `json:"citizen_id" binding:"required"`
    }
    var req ReportStolenBikeRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": err.Error(),
        })
        return
    }
    newBike := Bike{
        Description: req.Description,
        CitizenID:   req.CitizenID,
        Found:       false,
    }
    if result := DB.Create(&newBike); result.Error != nil {
        fmt.Println("failed to create bike record: %w", result.Error)
    }
    c.JSON(http.StatusCreated, gin.H{
        "message": "Stolen bike reported",
    })
}

func createCitizen(c *gin.Context) {
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
    newCitizen := Citizen{
        Email:    req.Email,
        Password: string(hashedPassword),
        Name:     req.Name,
    }
    
    // Save to database
    if result := DB.Create(&newCitizen); result.Error != nil {
        fmt.Println("failed to create citizen record: %w", result.Error)
    }
    c.JSON(http.StatusCreated, gin.H{
        "message": "User created",
    })
}

func getAllCitizen(c *gin.Context) {
    var citizens []Citizen
    DB.Find(&citizens)
    c.JSON(http.StatusOK, gin.H{
        "data": citizens,
    })
}


// Handler functions
func getAllPolice(c *gin.Context) {
    var polices []Police
    DB.Find(&polices)
    c.JSON(http.StatusOK, gin.H{
        "data": polices,
    })
}

func updatePolice(c *gin.Context) {
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
    var police Police
    if result := DB.First(&police, id); result.Error != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "error": "Police officer not found",
        })
        return
    }
    
    // Update the record
    result := DB.Model(&police).Updates(Police{
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
            "id": police.ID,
            "name": police.Name,
            "email": police.Email,
        },
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

func deletePolice(c *gin.Context) {
    id := c.Param("id")
    DB.Delete(&Police{}, id)
    c.JSON(http.StatusOK, gin.H{
        "message": "Police " + id + " deleted",
    })
}