package main

import (
	"fmt"
	"net/http"
    "time"

    "github.com/golang-jwt/jwt/v5"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type JWTClaim struct {
	Email string `json:"email"`
	UserID   uint   `json:"user_id"`
	jwt.RegisteredClaims
}

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
		login.POST("citizen", loginCitizen)
	}
}

func loginCitizen(c *gin.Context) {
    type LoginRequest struct {
        Email    string `json:"email" binding:"required,email"`
        Password string `json:"password" binding:"required"`
    }
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": err.Error(),
        })
        return
    }

    // Find the citizen
    var citizen Citizen
    if result := DB.Where("email = ?", req.Email).First(&citizen); result.Error != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "error": "Citizen not found",
        })
        return
    }

    // Compare the password
    if err := bcrypt.CompareHashAndPassword([]byte(citizen.Password), []byte(req.Password)); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "error": "Invalid password",
        })
        return
    }

    // Generate JWT
    token, err := generateJWT(citizen)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	// Set JWT token as a cookie
	c.SetCookie(
		"jwt_token",  // cookie name
		token,        // cookie value
		60*60*24,     // max age in seconds (24 hours)
		"/",          // path
		"",           // domain
		false,        // secure (set to true in production with HTTPS)
		true,         // http only (prevents JavaScript access)
	)

	c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
}

var jwtKey = []byte("some_secret_key")

func generateJWT(user Citizen) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token expires in 24 hours

	claims := &JWTClaim{
		Email: user.Email,
		UserID:   user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
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
    if err := DB.Model(&Bike{}).Where("police_id = ?", id).Update("police_id", nil).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update assigned bikes",
		})
		return
	}
	DB.Delete(&Police{}, id)
	c.JSON(http.StatusOK, gin.H{
		"message": "Police " + id + " deleted",
	})
}
