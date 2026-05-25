package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Config holds application configuration values.
type Config struct {
	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
	DBName    string
	DBSSLMode string
	JWTSecret string
	Port      string
}

// LoadConfig reads configuration from environment variables.
func LoadConfig() *Config {
	return &Config{
		DBHost:    getEnv("DB_HOST", "localhost"),
		DBPort:    getEnv("DB_PORT", "5432"),
		DBUser:    getEnv("DB_USER", "postgres"),
		DBPass:    getEnv("DB_PASSWORD", "postgres"),
		DBName:    getEnv("DB_NAME", "ecommerce_users"),
		DBSSLMode: getEnv("DB_SSLMODE", "disable"),
		JWTSecret: getEnv("JWT_SECRET", "secret"),
		Port:      getEnv("PORT", "8081"),
	}
}

// ConnectDB establishes a connection to PostgreSQL and runs SQL migrations.
func ConnectDB(cfg *Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run SQL migrations instead of AutoMigrate
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying sql.DB: %v", err)
	}
	runMigrations(sqlDB, getEnv("MIGRATIONS_PATH", "/app/migrations"))

	log.Println("Database connected and migrated successfully")
	return db
}

// runMigrations reads and executes all *.up.sql files from the migrations directory.
func runMigrations(db *sql.DB, migrationsPath string) {
	// Create schema_migrations tracking table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create schema_migrations table: %v", err)
	}

	// Find all .up.sql files
	files, err := filepath.Glob(filepath.Join(migrationsPath, "*.up.sql"))
	if err != nil {
		log.Printf("Warning: could not read migrations directory: %v", err)
		return
	}
	sort.Strings(files)

	for _, file := range files {
		version := filepath.Base(file)

		// Check if already applied
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = $1", version).Scan(&count)
		if err != nil {
			log.Fatalf("Failed to check migration status: %v", err)
		}
		if count > 0 {
			continue // Already applied
		}

		// Read and execute migration
		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", file, err)
		}

		// Execute each statement
		statements := strings.Split(string(content), ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}
			if _, err := db.Exec(stmt); err != nil {
				log.Fatalf("Failed to execute migration %s: %v", version, err)
			}
		}

		// Record migration
		_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version)
		if err != nil {
			log.Fatalf("Failed to record migration %s: %v", version, err)
		}

		log.Printf("Applied migration: %s", version)
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
