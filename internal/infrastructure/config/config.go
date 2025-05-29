package config

import (
	"os"

	"github.com/joho/godotenv"
)

type BitpinConfig struct {
	APIKey    string
	APISecret string
	BaseURL   string
}

type WallexConfig struct {
	APIKey  string
	BaseURL string
}

type Config struct {
	Exchange string // "bitpin", or "wallex"
	HTTPPort string
	LogLevel string

	Bitpin BitpinConfig
	Wallex WallexConfig
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	return &Config{
		Exchange: getEnv("EXCHANGE", "bitpin"),
		HTTPPort: getEnv("HTTP_PORT", "8080"),
		LogLevel: getEnv("LOG_LEVEL", "info"),

		Bitpin: BitpinConfig{
			APIKey:    mustGetEnv("BITPIN_API_KEY"),
			APISecret: mustGetEnv("BITPIN_API_SECRET"),
			BaseURL:   getEnv("BITPIN_BASE_URL", "https://api.bitpin.ir"),
		},

		Wallex: WallexConfig{
			APIKey:  mustGetEnv("WALLEX_API_KEY"),
			BaseURL: getEnv("WALLEX_BASE_URL", "https://api.wallex.ir"),
		},
	}, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func mustGetEnv(key string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	panic("environment variable " + key + " required")
}
