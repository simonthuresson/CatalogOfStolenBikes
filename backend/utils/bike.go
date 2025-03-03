package utils

import (
	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
	"app/db"
)

func FoundBike(c *gin.Context) {
	id := c.Param("id")

	// First check if the bike exists
	var bike db.Bike
	if err := db.DB.First(&bike, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Bike not found",
		})
		return
	}

	// Get the current police ID before updating
	previousPoliceID := bike.PoliceID

	// Update the bike: mark as found and clear police reference
	result := db.DB.Model(&db.Bike{}).Where("id = ?", id).Updates(map[string]interface{}{
		"found":     true,
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
		var unassignedBike db.Bike
		if err := db.DB.Where("found = ? AND police_id IS NULL", false).First(&unassignedBike).Error; err != nil {
			// No unassigned bikes found, that's fine
			fmt.Println("No unassigned bikes available for reassignment")
		} else {
			// Find an available police officer (one without an assigned bike)
			var availablePolice db.Police
			if err := db.DB.Where("id = ?", *previousPoliceID).First(&availablePolice).Error; err != nil {
				fmt.Println("Could not find the previously assigned police officer")
			} else {
				// Assign this police officer to the unassigned bike
				if err := db.DB.Model(&unassignedBike).Update("police_id", *previousPoliceID).Error; err != nil {
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

func GetAllBikes(c *gin.Context) {
	var bikes []db.Bike

	db.DB.Preload("Citizen").
		Preload("Police").
		Find(&bikes)

	c.JSON(http.StatusOK, gin.H{
		"bikes": bikes,
	})
}

func CreateBike(c *gin.Context) {
    citizenValue, exists := c.Get("citizen")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{
            "error": "Authentication required",
        })
        return
    }
    
    citizen, ok := citizenValue.(db.Citizen)
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Invalid citizen data",
        })
        return
    }

    type ReportStolenBikeRequest struct {
        Description string `json:"description" binding:"required"`
    }

    var req ReportStolenBikeRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": err.Error(),
        })
        return
    }

    newBike := db.Bike{
        Description: req.Description,
        CitizenID:   citizen.ID, 
        Citizen:     citizen,
        Found:       false,
        PoliceID:    nil,
        Police:      nil,
    }

    // Try to find an available police officer
    var police db.Police
    result := db.DB.Where("id NOT IN (SELECT police_id FROM bikes WHERE police_id IS NOT NULL)").First(&police)

    // If a police officer is found, assign them to the bike
    if result.Error == nil {
        newBike.PoliceID = &police.ID
        newBike.Police = &police
    }

    // Create the bike
    if result := db.DB.Create(&newBike); result.Error != nil {
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
        "message":               responseMsg,
        "bike_id":               newBike.ID,
        "assigned_to_police_id": newBike.PoliceID,
    })
}