package main

import (
    "fmt"
    "log"
    "time"
    "sync"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
	"golang.org/x/crypto/bcrypt"
)

// Models
type Police struct {
	gorm.Model           // Embeds ID, CreatedAt, UpdatedAt, and DeletedAt
	Email     string     `gorm:"uniqueIndex"`
	Password  string     `gorm:"not null" json:"-"` // Password field, excluded from JSON responses
	Name      string
	AssignedCases []Bike `gorm:"foreignKey:PoliceID"` // One-to-many relationship
}

type Citizen struct {
	gorm.Model           // Embeds ID, CreatedAt, UpdatedAt, and DeletedAt
	Email     string     `gorm:"uniqueIndex"` // Added email for login
	Password  string     `gorm:"not null" json:"-"` // Password field, excluded from JSON responses
    Name      string
	AuthorID  uint
	StolenBikes []Bike `gorm:"foreignKey:CitizenID"` // One-to-many relationship
}

type Bike struct {
	gorm.Model           // Embeds ID, CreatedAt, UpdatedAt, and DeletedAt
	Description string
	PoliceID    uint     // Foreign key for Police
	Police      Police   `gorm:"foreignKey:PoliceID"` // Belongs-to relationship
	CitizenID   uint     // Foreign key for Citizen
	Citizen     Citizen  `gorm:"foreignKey:CitizenID"` // Belongs-to relationship
	Found       bool     // Fixed lowercase to uppercase
}


var (
	DB   *gorm.DB
	once sync.Once
)

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
    var err error
    // Open connection to database
    DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: gormLogger,
    })
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

	if err := DB.Migrator().DropTable(&Police{}); err != nil {
		log.Fatalf("Failed to drop Post table: %v", err)
	}
	if err := DB.Migrator().DropTable(&Citizen{}); err != nil {
		log.Fatalf("Failed to drop User table: %v", err)
	}
    if err := DB.Migrator().DropTable(&Bike{}); err != nil {
		log.Fatalf("Failed to drop User table: %v", err)
	}

    // Auto migrate the schema
    err = DB.AutoMigrate(&Citizen{}, &Bike{}, &Police{})
    if err != nil {
        log.Fatalf("Failed to migrate database: %v", err)
    }
}

func AddPoliceOfficer(email, name, password string) (*Police, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	
	// Create police record
	newPolice := Police{
		Email:    email,
		Password: string(hashedPassword),
		Name:     name,
	}
	
	// Save to database
	if result := DB.Create(&newPolice); result.Error != nil {
		return nil, fmt.Errorf("failed to create police record: %w", result.Error)
	}
	
	return &newPolice, nil
}