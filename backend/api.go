package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
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
		bike.GET("/found/:id", foundBike)
	}
	login := r.Group("/login")
	{
		login.POST("police")
		login.POST("citizen")
	}
}

func foundBike(c *gin.Context) {
    id := c.Param("id")
    
    // First check if the bike exists
    var bike Bike
    if err := DB.First(&bike, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "error": "Bike not found",
        })
        return
    }
    
    // Get the current police ID before updating
    previousPoliceID := bike.PoliceID
    
    // Update the bike: mark as found and clear police reference
    result := DB.Model(&Bike{}).Where("id = ?", id).Updates(map[string]interface{}{
        "found": true,
        "police_id": nil,
    })
    
    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to update bike",
        })
        return
    }
    
    // Only try to reassign if a police officer was freed up
    if previousPoliceID != nil {
        // Find another stolen bike without an assigned police officer
        var unassignedBike Bike
        if err := DB.Where("found = ? AND police_id IS NULL", false).First(&unassignedBike).Error; err != nil {
            // No unassigned bikes found, that's fine
            fmt.Println("No unassigned bikes available for reassignment")
        } else {
            // Find an available police officer (one without an assigned bike)
            var availablePolice Police
            if err := DB.Where("id = ?", *previousPoliceID).First(&availablePolice).Error; err != nil {
                fmt.Println("Could not find the previously assigned police officer")
            } else {
                // Assign this police officer to the unassigned bike
                if err := DB.Model(&unassignedBike).Update("police_id", *previousPoliceID).Error; err != nil {
                    fmt.Println("Error assigning police to new bike:", err)
                } else {
                    fmt.Printf("Police officer %s (ID: %d) reassigned to bike ID: %d\n", 
                               availablePolice.Name, *previousPoliceID, unassignedBike.ID)
                }
            }
        }
    }
    
    c.JSON(http.StatusOK, gin.H{
        "message": "Bike marked as found and police officer unassigned",
    })
}

func getAllBikes(c *gin.Context) {
    var bikes []Bike
    
    // Preload relationships
    DB.Preload("Citizen").
       Preload("Police").
       Find(&bikes)
    
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
    
    // Verify citizen exists
    var citizen Citizen
    if result := DB.First(&citizen, req.CitizenID); result.Error != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "error": "Citizen not found",
        })
        return
    }
    
    // Initialize new bike with nil Police fields
    newBike := Bike{
        Description: req.Description,
        CitizenID:   req.CitizenID,
        Citizen:     citizen, 
        Found:       false,
        PoliceID:    nil,
        Police:      nil,
    }
    
    // Try to find an available police officer
    var police Police
    result := DB.Where("id NOT IN (SELECT police_id FROM bikes WHERE police_id IS NOT NULL)").First(&police)
    
    // If a police officer is found, assign them to the bike
    if result.Error == nil {
        newBike.PoliceID = &police.ID
        newBike.Police = &police
    }
    
    // Create the bike
    if result := DB.Create(&newBike); result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to report stolen bike",
        })
        return
    }
    
    // Prepare response message
    responseMsg := "Stolen bike reported"
    if newBike.PoliceID != nil {
        responseMsg += " and assigned to police"
    } else {
        responseMsg += " but no available police to assign"
    }
    
    c.JSON(http.StatusCreated, gin.H{
        "message": responseMsg,
        "bike_id": newBike.ID,
        "assigned_to_police_id": newBike.PoliceID,
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create citizen",
		})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "User created",
	})
}

func getAllCitizen(c *gin.Context) {
	var citizens []Citizen
	DB.Preload("StolenBikes").Find(&citizens)
	c.JSON(http.StatusOK, gin.H{
		"data": citizens,
	})
}

// Handler functions
func getAllPolice(c *gin.Context) {
	var polices []Police
	DB.Preload("AssignedBike").Find(&polices)
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
			"id":    police.ID,
			"name":  police.Name,
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create police",
		})
		return
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
