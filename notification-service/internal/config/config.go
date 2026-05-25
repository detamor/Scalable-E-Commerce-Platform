package config

import "os"

// Config holds application configuration values.
type Config struct {
	RabbitMQURL string
}

// LoadConfig reads configuration from environment variables.
func LoadConfig() *Config {
	return &Config{
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
