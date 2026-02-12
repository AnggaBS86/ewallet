package config

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type Config struct {
	Port      string
	JWTSecret string
	DBDSN     string
}

func Load() Config {
	cfg := Config{
		Port:      getEnv("PORT", "8080"),
		JWTSecret: getEnv("JWT_SECRET", ""),
		DBDSN:     resolveDBDSN(),
	}

	if strings.TrimSpace(cfg.JWTSecret) == "" {
		log.Fatal("JWT_SECRET is required")
	}
	if strings.TrimSpace(cfg.DBDSN) == "" {
		log.Fatal("DB_DSN is required")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

func resolveDBDSN() string {
	if dsn := strings.TrimSpace(getEnv("DB_DSN", "")); dsn != "" {
		return dsn
	}

	host := strings.TrimSpace(getEnv("DB_HOST", ""))
	port := strings.TrimSpace(getEnv("DB_PORT", ""))
	user := strings.TrimSpace(getEnv("DB_USER", ""))
	password := getEnv("DB_PASSWORD", "")
	name := strings.TrimSpace(getEnv("DB_NAME", ""))
	sslmode := strings.TrimSpace(getEnv("DB_SSLMODE", "disable"))

	if host == "" || port == "" || user == "" || name == "" {
		return ""
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, name, sslmode)
}
