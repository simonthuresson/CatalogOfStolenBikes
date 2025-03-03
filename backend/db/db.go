package db

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Police struct {
	gorm.Model
	Email        string `gorm:"uniqueIndex"`
	Password     string `gorm:"not null" json:"-"`
	Name         string
	AssignedBike *Bike `gorm:"foreignKey:PoliceID"`
}

type Citizen struct {
	gorm.Model
	Email       string `gorm:"uniqueIndex"`
	Password    string `gorm:"not null" json:"-"`
	Name        string
	StolenBikes []Bike `gorm:"foreignKey:CitizenID"` 
}

type Bike struct {
	gorm.Model
	Description string
	PoliceID    *uint `gorm:"unique"`
	Police      *Police
	CitizenID   uint    `gorm:"not null"`
	Citizen     Citizen 
	Found       bool
}

var (
	DB *gorm.DB
)

func Database() {
	// Database connection parameters
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		"postgres",   // host
		"myuser",     // user
		"mypassword", // password
		"mydb",       // dbname
		5432,         // port
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

func (c Citizen) GetEmail() string {
    return c.Email
}

func (c Citizen) GetID() uint {
    return c.ID
}

// For Police
func (p Police) GetEmail() string {
    return p.Email
}

func (p Police) GetID() uint {
    return p.ID
}
