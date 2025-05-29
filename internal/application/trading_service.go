package application

import (
	"context"
	"fmt"

	"trade/internal/domain"
	"trade/internal/ports"
)

type TradingService struct {
	exchange domain.ExchangePort
	log      ports.LoggerPort
}

func NewTradingService(exch domain.ExchangePort, log ports.LoggerPort) *TradingService {
	return &TradingService{exchange: exch, log: log}
}

func (s *TradingService) CreateOrder(ctx context.Context, req domain.OrderRequest) (domain.OrderResponse, error) {
	resp, err := s.exchange.CreateOrder(ctx, req)
	if err != nil {
		s.log.Error(ctx, "CreateOrder failed", ports.Fields{"error": err})
		return domain.OrderResponse{}, fmt.Errorf("CreateOrder failed: %w", err)
	}
	return resp, nil
}

func (s *TradingService) CancelOrder(ctx context.Context, symbol, orderID string) error {
	err := s.exchange.CancelOrder(ctx, symbol, orderID)
	if err != nil {
		s.log.Error(ctx, "CancelOrder failed", ports.Fields{"error": err})
		return fmt.Errorf("CancelOrder failed: %w", err)
	}
	return nil
}

func (s *TradingService) GetBalance(ctx context.Context) ([]domain.Balance, error) {
	balances, err := s.exchange.GetBalance(ctx)
	if err != nil {
		s.log.Error(ctx, "GetBalance failed", ports.Fields{"error": err})
		return nil, fmt.Errorf("GetBalance failed: %w", err)
	}
	return balances, nil
}

func (s *TradingService) GetOrderBook(ctx context.Context, symbol string) (domain.OrderBook, error) {
	book, err := s.exchange.GetOrderBook(ctx, symbol)
	if err != nil {
		s.log.Error(ctx, "GetOrderBook failed", ports.Fields{"error": err})
		return domain.OrderBook{}, fmt.Errorf("GetOrderBook failed: %w", err)
	}
	return book, nil
}
