package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort         string
	MongoURI           string
	MongoDBName        string
	JWTSecret          string
	JWTExpiration      time.Duration
	RedisAddr          string
	RedisPassword      string
	RedisDB            int
	EnableTLS          bool
	TLSCertFile        string
	TLSKeyFile         string
	WorkerPoolSize     int
	MaxRequestSize     int64
	RateLimitPerSecond int
}

func LoadConfig() *Config {
	// Load .env file if it exists (ignore error if not found)
	_ = godotenv.Load()

	cfg := &Config{
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		MongoURI:           getEnv("MONGO_URI", "mongodb://127.0.0.1:27017/?replicaSet=rs0"),
		MongoDBName:        getEnv("MONGO_DB_NAME", "divvydoo"),
		JWTSecret:          getEnv("JWT_SECRET", "default-secret-key"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		EnableTLS:          getEnvAsBool("ENABLE_TLS", false),
		TLSCertFile:        getEnv("TLS_CERT_FILE", ""),
		TLSKeyFile:         getEnv("TLS_KEY_FILE", ""),
		WorkerPoolSize:     getEnvAsInt("WORKER_POOL_SIZE", 10),
		MaxRequestSize:     getEnvAsInt64("MAX_REQUEST_SIZE", 1048576), // 1MB
		RateLimitPerSecond: getEnvAsInt("RATE_LIMIT_PER_SECOND", 100),
	}

	jwtExp := getEnvAsInt("JWT_EXPIRATION_HOURS", 24)
	cfg.JWTExpiration = time.Duration(jwtExp) * time.Hour

	redisDB := getEnvAsInt("REDIS_DB", 0)
	cfg.RedisDB = redisDB

	return cfg
}

// Helper functions for environment variables
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if int64Value, err := strconv.ParseInt(value, 10, 64); err == nil {
			return int64Value
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
