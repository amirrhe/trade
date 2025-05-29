package domain

import "context"

type ExchangePort interface {
	CreateOrder(ctx context.Context, req OrderRequest) (OrderResponse, error)

	CancelOrder(ctx context.Context, symbol, orderID string) error

	GetBalance(ctx context.Context) ([]Balance, error)

	GetOrderBook(ctx context.Context, symbol string) (OrderBook, error)
}
