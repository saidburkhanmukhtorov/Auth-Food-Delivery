package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cast"
)

// Config struct holds the configuration settings.
type Config struct {
	Environment string // Development, Production, etc.
	HTTPPort    string

	// PostgreSQL Configuration
	PostgresUser     string
	PostgresPassword string
	PostgresHost     string
	PostgresPort     string
	PostgresDatabase string

	// Redis Configuration
	RedisAddress  string
	RedisPassword string
	RedisDB       int

	// JWT Configuration
	JWTSecretKey string
	JWTExpiry    int // In minutes

	// Email Configuration (if using email OTP)
	EmailSender      string
	EmailPassword    string
	EmailHost        string
	EmailPort        int
	EmailFromAddress string
}

// Load loads the configuration from environment variables.
func Load() Config {
	// Load environment variables from .env file (if it exists)
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, loading from environment variables")
	}

	config := Config{}

	// General Configuration
	config.Environment = cast.ToString(getOrReturnDefault("ENVIRONMENT", "development"))
	config.HTTPPort = cast.ToString(getOrReturnDefault("HTTP_PORT", ":8080"))

	// PostgreSQL Configuration
	config.PostgresUser = cast.ToString(getOrReturnDefault("POSTGRES_USER", "sayyidmuhammad"))
	config.PostgresPassword = cast.ToString(getOrReturnDefault("POSTGRES_PASSWORD", "root"))
	config.PostgresHost = cast.ToString(getOrReturnDefault("POSTGRES_HOST", "localhost"))
	config.PostgresPort = cast.ToString(getOrReturnDefault("POSTGRES_PORT", "5432"))
	config.PostgresDatabase = cast.ToString(getOrReturnDefault("POSTGRES_DATABASE", "postgres"))

	// Redis Configuration
	config.RedisAddress = cast.ToString(getOrReturnDefault("REDIS_ADDRESS", "localhost:6379"))
	config.RedisPassword = cast.ToString(getOrReturnDefault("REDIS_PASSWORD", ""))
	config.RedisDB = cast.ToInt(getOrReturnDefault("REDIS_DB", 0))

	// JWT Configuration
	config.JWTSecretKey = cast.ToString(getOrReturnDefault("JWT_SECRET_KEY", "your_secret_key"))
	config.JWTExpiry = cast.ToInt(getOrReturnDefault("JWT_EXPIRY", 60))

	// Email Configuration (if using email OTP)
	config.EmailSender = cast.ToString(getOrReturnDefault("EMAIL_SENDER", "qodirovazizbek1129@gmail.com"))
	config.EmailPassword = cast.ToString(getOrReturnDefault("EMAIL_PASSWORD", "jkzt mtab wvaq ewlm llldsf"))
	config.EmailHost = cast.ToString(getOrReturnDefault("EMAIL_HOST", "smtp.gmail.com"))
	config.EmailPort = cast.ToInt(getOrReturnDefault("EMAIL_PORT", 587))
	config.EmailFromAddress = cast.ToString(getOrReturnDefault("EMAIL_FROM_ADDRESS", "your_email@example.com"))

	return config
}

// getOrReturnDefault retrieves the value of an environment variable or returns a default value if it's not set.
func getOrReturnDefault(key string, defaultValue interface{}) interface{} {
	val, exists := os.LookupEnv(key)
	if exists {
		return val
	}
	return defaultValue
}
