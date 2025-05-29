package domain

import "time"

type OrderSide string

const (
	SideBuy  OrderSide = "BUY"
	SideSell OrderSide = "SELL"
)

type OrderType string

const (
	TypeMarket OrderType = "MARKET"
	TypeLimit  OrderType = "LIMIT"
)

type OrderRequest struct {
	Symbol    string
	Side      OrderSide
	Type      OrderType
	Quantity  float64
	Price     *float64
	ClientID  *string
	Timestamp time.Time
}

type OrderResponse struct {
	ID        string
	Symbol    string
	Side      OrderSide
	Type      OrderType
	Quantity  float64
	Price     float64
	Status    string
	Timestamp time.Time
}

type Balance struct {
	Asset  string
	Free   float64
	Locked float64
}

type DepthLevel struct {
	Price    float64
	Quantity float64
}

type OrderBook struct {
	Symbol string
	Bids   []DepthLevel
	Asks   []DepthLevel
}
