package main

import (
	"github.com/joho/godotenv"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	dbURL := os.Getenv("DATABASE_URL")

	if dbURL == "" {
		panic("DATABASE_URL environment variable not set")
	}
}
