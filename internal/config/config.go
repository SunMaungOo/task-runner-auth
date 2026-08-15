package config

import (
	"os"
	"time"
)

type Config struct {
	Port                   string
	LogLevel               string
	MongoURI               string
	MongoDatabase          string
	JWTSecret              string
	TokenTTL               time.Duration
	ShutdownTimeout        time.Duration
	DatabaseConnectTimeout time.Duration
	DatabasePingTimeout    time.Duration
}

func Load() Config {
	return Config{
		Port:                   getEnvStr("PORT", "8080"),
		LogLevel:               getEnvStr("LOG_LEVEL", "info"),
		MongoURI:               getEnvStr("MONGO_URI", "mongodb://localhost:27017"),
		MongoDatabase:          getEnvStr("MONGO_DATABASE", "authdb"),
		JWTSecret:              getEnvStr("JWT_SECRET", ""),
		TokenTTL:               getEnvDuration("TOKEN_TTL", 24*time.Hour),
		ShutdownTimeout:        getEnvDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
		DatabaseConnectTimeout: getEnvDuration("DATABASE_TIMEOUT", 10*time.Second),
		DatabasePingTimeout:    getEnvDuration("DATABASE_PING_TIMEOUT", 5*time.Second),
	}
}

func getEnvStr(key string, fallback string) string {

	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {

	if value := os.Getenv(key); value != "" {

		if duration, err := time.ParseDuration(value); err != nil {
			return duration
		}

	}

	return fallback

}
