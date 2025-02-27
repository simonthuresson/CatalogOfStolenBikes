package main

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

func setupAPIRoutes(r *gin.Engine) {
    // Create an api group/router
    api := r.Group("/api")
    {
        // GET /api/health
        api.GET("/health", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{
                "status": "healthy",
            })
        })

        // Example of grouped endpoints for a resource
        users := api.Group("/users")
        {
            users.GET("/", getAllUsers)      // GET /api/users
            users.GET("/:id", getUser)       // GET /api/users/123
            users.POST("/", createUser)      // POST /api/users
            users.PUT("/:id", updateUser)    // PUT /api/users/123
            users.DELETE("/:id", deleteUser) // DELETE /api/users/123
        }

        // You can add more resource groups here
        posts := api.Group("/posts")
        {
            posts.GET("/", getAllPosts)
            // ... other post routes
        }
    }
}

// Handler functions
func getAllUsers(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "users": []string{"user1", "user2"},
    })
}

func getUser(c *gin.Context) {
    id := c.Param("id")
    c.JSON(http.StatusOK, gin.H{
        "id": id,
        "name": "Sample User",
    })
}

func createUser(c *gin.Context) {
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