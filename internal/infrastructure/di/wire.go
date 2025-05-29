package di

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"trade/internal/adapters/bitpin"
	"trade/internal/adapters/logger"
	"trade/internal/adapters/wallex"
	"trade/internal/application"
	"trade/internal/domain"
	"trade/internal/infrastructure/config"
	"trade/pkg/transport"
)

func BuildApp(cfg *config.Config) (*fiber.App, error) {
	logPort, err := logger.NewLogrusAdapter(cfg.LogLevel)
	if err != nil {
		return nil, err
	}

	var exch domain.ExchangePort
	switch cfg.Exchange {
	case "bitpin":
		exch = bitpin.NewAdapter(
			cfg.Bitpin.APIKey,
			cfg.Bitpin.APISecret,
			cfg.Bitpin.BaseURL,
			logPort,
		)

	case "wallex":
		exch = wallex.NewAdapter(
			cfg.Wallex.APIKey,
			cfg.Wallex.BaseURL,
			logPort,
		)
	default:
		return nil, fmt.Errorf("unsupported exchange: %s", cfg.Exchange)
	}

	svc := application.NewTradingService(exch, logPort)

	app := transport.NewRouter(svc, logPort)
	return app, nil
}
