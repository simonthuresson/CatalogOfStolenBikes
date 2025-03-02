package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"app/db"
	"golang.org/x/crypto/bcrypt"
)

type JWTClaim struct {
	Email  string `json:"email"`
	UserID uint   `json:"user_id"`
	jwt.RegisteredClaims
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("jwt_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaim{}, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Check if the token is valid
		claims, ok := token.Claims.(*JWTClaim)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Find user by ID
		var citizen db.Citizen
		if result :=db.DB.First(&citizen, claims.ID); result.Error != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}
		c.Set("citizen", citizen)
		c.Next()
	}

	}
	
var jwtKey = []byte("some_secret_key")

func generateJWT(user db.Citizen) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token expires in 24 hours

	claims := &JWTClaim{
		Email:  user.Email,
		UserID: user.ID,
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

func LoginCitizen(c *gin.Context) {
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
	var citizen db.Citizen
	if result := db.DB.Where("email = ?", req.Email).First(&citizen); result.Error != nil {
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
		"jwt_token", // cookie name
		token,       // cookie value
		60*60*24,    // max age in seconds (24 hours)
		"/",         // path
		"",          // domain
		false,       // secure (set to true in production with HTTPS)
		true,        // http only (prevents JavaScript access)
	)

	c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
}