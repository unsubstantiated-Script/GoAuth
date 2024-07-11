package main

import (
	"crypto/rand"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
	"github.com/lucsky/cuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"os"
	"strings"
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

type AuthRequest struct {
	ResponseType string `json:"response_type" query:"response_type"`
	ClientID     string `json:"client_id" query:"client_id"`
	RedirectURI  string `json:"redirect_uri" query:"redirect_uri"`
	Scope        string
	State        string
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
		Website:     "http://localhost:8080",
		RedirectURI: "http://localhost:8080/auth/callback",
		Logo:        "https://placehold.co/600x400",
	})

	views := html.New("./views", ".html")

	api := fiber.New(fiber.Config{
		AppName: "Authorization Service",
		Views:   views,
	})

	api.Use(logger.New())
	api.Use(recover.New())

	api.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("hello!")
	})

	api.Get("/auth", func(c *fiber.Ctx) error {
		authRequest := new(AuthRequest)
		if err := c.QueryParser(authRequest); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
		}

		if authRequest.ResponseType != "code" {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
		}

		if authRequest.ClientID != os.Getenv("CLIENT_ID") {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
		}

		if !strings.Contains(authRequest.RedirectURI, "https") {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
		}

		if authRequest.Scope == "" {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
		}

		// Check for client
		client := new(Client)
		if err := db.Where("name = ?", authRequest.ClientID).First(client).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid client"})
		}

		//Generate temp code
		code, err := cuid.NewCrypto(rand.Reader)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "server error"})
		}

		c.Cookie(&fiber.Cookie{
			Name:     "auth_request_code",
			Value:    code,
			Secure:   true,
			Expires:  time.Now().Add(1 * time.Minute),
			HTTPOnly: true,
		})

		return c.Render("authorize_client", fiber.Map{
			"Logo":    client.Logo,
			"Name":    client.Name,
			"Website": client.Website,
		})

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
