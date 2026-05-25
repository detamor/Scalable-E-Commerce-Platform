package config

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"order-service/internal/model"
)

// Config holds application configuration values.
type Config struct {
	DBHost            string
	DBPort            string
	DBUser            string
	DBPass            string
	DBName            string
	DBSSLMode         string
	JWTSecret         string
	Port              string
	CartServiceURL    string
	ProductServiceURL string
	PaymentServiceURL string
	RabbitMQURL       string
}

// LoadConfig reads configuration from environment variables.
func LoadConfig() *Config {
	return &Config{
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBUser:            getEnv("DB_USER", "postgres"),
		DBPass:            getEnv("DB_PASSWORD", "postgres"),
		DBName:            getEnv("DB_NAME", "ecommerce_orders"),
		DBSSLMode:         getEnv("DB_SSLMODE", "disable"),
		JWTSecret:         getEnv("JWT_SECRET", "secret"),
		Port:              getEnv("PORT", "8084"),
		CartServiceURL:    getEnv("CART_SERVICE_URL", "http://localhost:8083"),
		ProductServiceURL: getEnv("PRODUCT_SERVICE_URL", "http://localhost:8082"),
		PaymentServiceURL: getEnv("PAYMENT_SERVICE_URL", "http://localhost:8085"),
		RabbitMQURL:       getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
	}
}

// ConnectDB establishes a connection to PostgreSQL and auto-migrates tables.
func ConnectDB(cfg *Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&model.Order{}, &model.OrderItem{}); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}

	log.Println("Database connected and migrated successfully")
	return db
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
