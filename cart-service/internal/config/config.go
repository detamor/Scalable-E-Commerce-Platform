package config

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

// Config holds application configuration values.
type Config struct {
	RedisAddr string
	JWTSecret string
	Port      string
}

// LoadConfig reads configuration from environment variables.
func LoadConfig() *Config {
	return &Config{
		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret: getEnv("JWT_SECRET", "secret"),
		Port:      getEnv("PORT", "8083"),
	}
}

// ConnectRedis establishes a connection to Redis.
func ConnectRedis(cfg *Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Redis connected successfully")
	return client
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
