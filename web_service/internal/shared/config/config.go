package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	DBDriver       string
	DBUrl          string
	JWTSecret      string
	AllowedOrigins []string
	UploadPath     string
	BaseURL        string
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	rawOrigins := getEnv("ALLOWED_ORIGINS", "http://localhost:3000")
	allowedOrigins := strings.Split(rawOrigins, ",")

	cfg := &Config{
		Port:           getEnv("PORT", ":8080"),
		DBDriver:       getEnv("DB_DRIVER", "oracle"),
		DBUrl:          getEnv("DATABASE_DSN", "oracle://plastik:k4r4w4ng@192.168.3.3:1521/XE"),
		JWTSecret:      getEnv("JWT_SECRET", "your-very-secret-key"),
		AllowedOrigins: allowedOrigins,
		UploadPath:     getEnv("UPLOAD_PATH", "./uploads/article/images"),
		BaseURL:        getEnv("BASE_URL", "http://localhost:8080"),
	}

	return cfg, nil
}

// Helper untuk membaca env dengan nilai default
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
