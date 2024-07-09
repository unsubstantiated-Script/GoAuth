package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"os"
	"time"
)

type Client struct {
	ID          string `gorm:"primaryKey"`
	Name        string `gorm:"uniqueIndex"`
	Website     string
	Logo        string
	RedirectURI string         `json:"redirect_uri"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		panic("DATABASE_URL environment variable not set")
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database")
	}

	err = db.AutoMigrate(&Client{})
	if err != nil {
		panic("Migration failed")
	}

	// Insert dummy client OnConflict allows an update when ID conflicts or matches again.
	db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "website", "logo", "redirect_uri"}),
	}).Create(&Client{
		ID:          "1",
		Name:        "fiber",
		Website:     "https://fiber.com",
		RedirectURI: "http://localhost:8080/auth/callback",
		Logo:        "https://placehold.co/600x400",
	})

	api := fiber.New(fiber.Config{
		AppName: "Authorization Service",
	})

	api.Use(logger.New())
	api.Use(recover.New())

	api.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("hello!")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	err = api.Listen(":" + port)
	if err != nil {
		panic("API has failed")
	}
}
