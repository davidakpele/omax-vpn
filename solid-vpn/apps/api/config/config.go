package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port        string
	Environment string

	DatabaseURL string

	JWTSecret          string
	JWTAccessExpiryMin int
	JWTRefreshExpiryH  int

	VPNEngineURL   string
	VPNEngineToken string

	VPNDNS string

	LogLevel string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),

		DatabaseURL: mustGetEnv("DATABASE_URL"),

		JWTSecret:          mustGetEnv("JWT_SECRET"),
		JWTAccessExpiryMin: getEnvInt("JWT_ACCESS_EXPIRY_MINUTES", 15),
		JWTRefreshExpiryH:  getEnvInt("JWT_REFRESH_EXPIRY_HOURS", 168),

		VPNEngineURL:   getEnv("VPN_ENGINE_URL", "http://vpn-engine:9090"),
		VPNEngineToken: mustGetEnv("VPN_ENGINE_TOKEN"),

		VPNDNS: getEnv("VPN_DNS", "1.1.1.1"),

		LogLevel: getEnv("LOG_LEVEL", "info"),
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
