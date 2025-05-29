package config

import (
	"os"

	"github.com/joho/godotenv"
)

type BinanceConfig struct {
	APIKey    string
	APISecret string
	BaseURL   string
}

type KuCoinConfig struct {
	APIKey     string
	APISecret  string
	Passphrase string
	BaseURL    string
}

type Config struct {
	Exchange string
	HTTPPort string
	Binance  BinanceConfig
	KuCoin   KuCoinConfig
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Exchange: getEnv("EXCHANGE", "binance"),
		HTTPPort: getEnv("HTTP_PORT", "8080"),
		Binance: BinanceConfig{
			APIKey:    getEnv("BINANCE_API_KEY", ""),
			APISecret: getEnv("BINANCE_API_SECRET", ""),
			BaseURL:   getEnv("BINANCE_BASE_URL", "https://api.binance.com"),
		},
		KuCoin: KuCoinConfig{
			APIKey:     getEnv("KUCOIN_API_KEY", ""),
			APISecret:  getEnv("KUCOIN_API_SECRET", ""),
			Passphrase: getEnv("KUCOIN_PASSPHRASE", ""),
			BaseURL:    getEnv("KUCOIN_BASE_URL", "https://api.kucoin.com"),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
