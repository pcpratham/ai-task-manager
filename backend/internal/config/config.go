package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DBDSN          string
	JWTSecret      string
	GeminiKey      string
	AllowedOrigins []string
}

func Load() (*Config, error) {
	// Load .env file if it exists (for local development)
	godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		dbDSN = "taskuser:taskpassword@tcp(localhost:3306)/taskmanager?parseTime=true"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-change-in-production"
	}

	geminiKey := os.Getenv("GEMINI_API_KEY")

	allowedOriginsStr := os.Getenv("ALLOWED_ORIGINS")
	if allowedOriginsStr == "" {
		allowedOriginsStr = "http://localhost:3000"
	}
	allowedOrigins := strings.Split(allowedOriginsStr, ",")
	for i := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
	}

	return &Config{
		Port:           port,
		DBDSN:          dbDSN,
		JWTSecret:      jwtSecret,
		GeminiKey:      geminiKey,
		AllowedOrigins: allowedOrigins,
	}, nil
}
