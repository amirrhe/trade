package bitpin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"trade/internal/domain"
	"trade/internal/ports"
)

type BitpinAdapter struct {
	client *Client
	log    ports.LoggerPort
}

func NewAdapter(apiKey, apiSecret, baseURL string, log ports.LoggerPort) *BitpinAdapter {
	c := NewClient(apiKey, apiSecret, baseURL, log)
	return &BitpinAdapter{client: c, log: log}
}

func (b *BitpinAdapter) CreateOrder(ctx context.Context, req domain.OrderRequest) (domain.OrderResponse, error) {
	b.log.Info(ctx, "CreateOrder start", ports.Fields{"symbol": req.Symbol, "side": req.Side})
	url := fmt.Sprintf("%s/api/v1/odr/orders/", b.client.baseURL)

	payload := map[string]interface{}{
		"symbol":      req.Symbol,
		"type":        strings.ToLower(string(req.Type)),
		"side":        strings.ToLower(string(req.Side)),
		"base_amount": fmt.Sprintf("%f", req.Quantity),
	}
	if req.Price != nil {
		payload["price"] = fmt.Sprintf("%f", *req.Price)
	}
	if req.ClientID != nil {
		payload["identifier"] = *req.ClientID
	}

	bodyBytes, _ := json.Marshal(payload)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(ctx, httpReq)
	if err != nil {
		return domain.OrderResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		data, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf("status %d: %s", resp.StatusCode, data)
		b.log.Error(ctx, "CreateOrder failed", ports.Fields{"error": err.Error()})
		return domain.OrderResponse{}, err
	}

	var r struct {
		ID                int64  `json:"id"`
		Symbol            string `json:"symbol"`
		Side, Type        string
		Price             string `json:"price"`
		DealedBaseAmount  string `json:"dealed_base_amount"`
		DealedQuoteAmount string `json:"dealed_quote_amount"`
		State             string `json:"state"`
		CreatedAt         string `json:"created_at"`
	}
	json.NewDecoder(resp.Body).Decode(&r)

	qty, _ := strconv.ParseFloat(r.DealedBaseAmount, 64)
	priceF, _ := strconv.ParseFloat(r.Price, 64)
	ts, _ := time.Parse(time.RFC3339, r.CreatedAt)

	result := domain.OrderResponse{
		ID:        strconv.FormatInt(r.ID, 10),
		Symbol:    r.Symbol,
		Side:      domain.OrderSide(strings.ToUpper(r.Side)),
		Type:      domain.OrderType(strings.ToUpper(r.Type)),
		Quantity:  qty,
		Price:     priceF,
		Status:    r.State,
		Timestamp: ts,
	}
	b.log.Info(ctx, "CreateOrder succeeded", ports.Fields{"orderID": result.ID})
	return result, nil
}

func (b *BitpinAdapter) CancelOrder(ctx context.Context, symbol, orderID string) error {
	b.log.Info(ctx, "CancelOrder start", ports.Fields{"symbol": symbol, "orderID": orderID})
	url := fmt.Sprintf("%s/api/v1/odr/orders/%s/", b.client.baseURL, orderID)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)

	resp, err := b.client.Do(ctx, httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		data, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf("status %d: %s", resp.StatusCode, data)
		b.log.Error(ctx, "CancelOrder failed", ports.Fields{"error": err.Error()})
		return err
	}
	b.log.Info(ctx, "CancelOrder succeeded", ports.Fields{"orderID": orderID})
	return nil
}

func (b *BitpinAdapter) GetBalance(ctx context.Context) ([]domain.Balance, error) {
	url := fmt.Sprintf("%s/api/v1/wlt/wallets/", b.client.baseURL)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	resp, err := b.client.Do(ctx, httpReq)
	if err != nil {
		b.log.Error(ctx, "GetBalance request error", ports.Fields{"error": err.Error()})
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf("status %d: %s", resp.StatusCode, data)
		b.log.Error(ctx, "GetBalance failed", ports.Fields{"error": err.Error()})
		return nil, err
	}

	var wallets []struct {
		Asset   string `json:"asset"`
		Balance string `json:"balance"`
		Frozen  string `json:"frozen"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wallets); err != nil {
		b.log.Error(ctx, "GetBalance decode error", ports.Fields{"error": err.Error()})
		return nil, err
	}

	balances := make([]domain.Balance, 0, len(wallets))
	for _, w := range wallets {
		total, _ := strconv.ParseFloat(w.Balance, 64)
		frozen, _ := strconv.ParseFloat(w.Frozen, 64)
		free := total - frozen
		balances = append(balances, domain.Balance{Asset: w.Asset, Free: free, Locked: frozen})
	}

	b.log.Info(ctx, "GetBalance succeeded", ports.Fields{
		"count": len(balances),
	})
	return balances, nil
}

func (b *BitpinAdapter) GetOrderBook(ctx context.Context, symbol string) (domain.OrderBook, error) {
	b.log.Info(ctx, "GetOrderBook start", ports.Fields{"symbol": symbol})
	url := fmt.Sprintf("%s/api/v1/mth/orderbook/%s/", b.client.baseURL, symbol)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	resp, err := b.client.Do(ctx, httpReq)
	if err != nil {
		return domain.OrderBook{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		err := fmt.Errorf("status %d: %s", resp.StatusCode, data)
		b.log.Error(ctx, "GetOrderBook failed", ports.Fields{"error": err.Error()})
		return domain.OrderBook{}, err
	}

	var r struct {
		Asks [][]string `json:"asks"`
		Bids [][]string `json:"bids"`
	}
	json.NewDecoder(resp.Body).Decode(&r)

	asks := make([]domain.DepthLevel, len(r.Asks))
	for i, lvl := range r.Asks {
		price, _ := strconv.ParseFloat(lvl[0], 64)
		qty, _ := strconv.ParseFloat(lvl[1], 64)
		asks[i] = domain.DepthLevel{Price: price, Quantity: qty}
	}
	bids := make([]domain.DepthLevel, len(r.Bids))
	for i, lvl := range r.Bids {
		price, _ := strconv.ParseFloat(lvl[0], 64)
		qty, _ := strconv.ParseFloat(lvl[1], 64)
		bids[i] = domain.DepthLevel{Price: price, Quantity: qty}
	}

	book := domain.OrderBook{Symbol: symbol, Bids: bids, Asks: asks}
	b.log.Info(ctx, "GetOrderBook succeeded", ports.Fields{"bids": len(bids), "asks": len(asks)})
	return book, nil
}
