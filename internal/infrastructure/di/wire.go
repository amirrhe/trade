// internal/infrastructure/di/wire.go
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
	// 1) Logger
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

	// 3) Core service
	svc := application.NewTradingService(exch, logPort)

	// 4) HTTP transport
	app := transport.NewRouter(svc, logPort)
	return app, nil
}
