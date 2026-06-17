package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	ServerPort        string
	GoogleCredentials string
	SpreadsheetID     string
	SheetName         string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	return &Config{
		ServerPort:        getEnv("SERVER_PORT", "8080"),
		GoogleCredentials: getEnv("GOOGLE_CREDENTIALS", ""),
		SpreadsheetID:     getEnv("SPREADSHEET_ID", ""),
	}
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
