package main

import (
    "fmt"
    "log"
    "time"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

// Models
type User struct {
    gorm.Model           // Embeds ID, CreatedAt, UpdatedAt, and DeletedAt
    Email     string     `gorm:"uniqueIndex"`
    Name      string
    Posts     []Post     `gorm:"foreignKey:AuthorID"`
}

type Post struct {
    gorm.Model           // Embeds ID, CreatedAt, UpdatedAt, and DeletedAt
    Title     string
    Content   string
    AuthorID  uint
    Author    User
}

func Database() {
    // Database connection parameters
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
        "postgres",     // host
        "myuser",       // user
        "mypassword",   // password
        "mydb",         // dbname
        5432,          // port
    )

    // Custom GORM logger configuration
    gormLogger := logger.New(
        log.New(log.Writer(), "\r\n", log.LstdFlags),
        logger.Config{
            SlowThreshold:             time.Second,
            LogLevel:                  logger.Info,
            IgnoreRecordNotFoundError: true,
            Colorful:                  true,
        },
    )

    // Open connection to database
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: gormLogger,
    })
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

	if err := db.Migrator().DropTable(&Post{}); err != nil {
		log.Fatalf("Failed to drop Post table: %v", err)
	}
	if err := db.Migrator().DropTable(&User{}); err != nil {
		log.Fatalf("Failed to drop User table: %v", err)
	}

    // Auto migrate the schema
    err = db.AutoMigrate(&User{}, &Post{})
    if err != nil {
        log.Fatalf("Failed to migrate database: %v", err)
    }

    // Create a test user
    user := User{
		Email: "john.doe@example.com",
		Name:  "John Doe",
		Posts: []Post{
			{
				Title:   "My First Post",
				Content: "This is the content of my first post",
			},
			{
				Title:   "My Second Post",
				Content: "This is the content of my second post",
			},
		},
	}

    result := db.Create(&user)
    if result.Error != nil {
        log.Fatalf("Failed to create user: %v", result.Error)
    }

    // Query the user with their posts
    var retrievedUser User
	result = db.Preload("Posts").First(&retrievedUser)
    if result.Error != nil {
        log.Fatalf("Failed to retrieve user: %v", result.Error)
    }

    fmt.Printf("User: %v\n", retrievedUser)
}