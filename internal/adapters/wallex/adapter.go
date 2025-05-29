package wallex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"trade/internal/domain"
	"trade/internal/ports"
)

type WallexAdapter struct {
	client *Client
	log    ports.LoggerPort
}

func NewAdapter(apiKey, baseURL string, log ports.LoggerPort) *WallexAdapter {
	return &WallexAdapter{
		client: NewClient(apiKey, baseURL, log),
		log:    log,
	}
}

func (w *WallexAdapter) CreateOrder(ctx context.Context, req domain.OrderRequest) (domain.OrderResponse, error) {
	w.log.Info(ctx, "CreateOrder start", ports.Fields{
		"symbol":   req.Symbol,
		"side":     req.Side,
		"type":     req.Type,
		"quantity": req.Quantity,
	})
	start := time.Now()

	payload := map[string]string{
		"symbol":   req.Symbol,
		"type":     string(req.Type),
		"side":     string(req.Side),
		"quantity": fmt.Sprintf("%f", req.Quantity),
	}
	if req.Price != nil {
		payload["price"] = fmt.Sprintf("%f", *req.Price)
	}
	if req.ClientID != nil {
		payload["client_id"] = *req.ClientID
	}
	bodyBytes, _ := json.Marshal(payload)
	w.log.Debug(ctx, "CreateOrder payload", ports.Fields{"body": string(bodyBytes)})

	url := fmt.Sprintf("%s/v1/account/orders", w.client.baseURL)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	resp, err := w.client.Do(ctx, httpReq)
	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		w.log.Error(ctx, "CreateOrder HTTP error", ports.Fields{"error": err.Error(), "latency_ms": elapsed})
		return domain.OrderResponse{}, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	w.log.Debug(ctx, "CreateOrder response", ports.Fields{
		"status":     resp.StatusCode,
		"body":       string(data),
		"latency_ms": elapsed,
	})

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		err := fmt.Errorf("status %d: %s", resp.StatusCode, data)
		w.log.Error(ctx, "CreateOrder failed", ports.Fields{"error": err.Error(), "latency_ms": elapsed})
		return domain.OrderResponse{}, err
	}

	var wrap struct {
		Success bool `json:"success"`
		Result  struct {
			Symbol        string `json:"symbol"`
			Type          string `json:"type"`
			Side          string `json:"side"`
			Price         string `json:"price"`
			OrigQty       string `json:"origQty"`
			ExecutedQty   string `json:"executedQty"`
			TransactTime  int64  `json:"transactTime"`
			ClientOrderId string `json:"clientOrderId"`
			Status        string `json:"status"`
			Active        bool   `json:"active"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &wrap); err != nil {
		w.log.Error(ctx, "CreateOrder decode error", ports.Fields{"error": err.Error(), "latency_ms": elapsed})
		return domain.OrderResponse{}, err
	}

	origQty, _ := strconv.ParseFloat(wrap.Result.OrigQty, 64)
	exQty, _ := strconv.ParseFloat(wrap.Result.ExecutedQty, 64)
	priceVal, _ := strconv.ParseFloat(wrap.Result.Price, 64)
	ts := time.Unix(wrap.Result.TransactTime, 0)

	res := domain.OrderResponse{
		ID:        wrap.Result.ClientOrderId,
		Symbol:    wrap.Result.Symbol,
		Side:      domain.OrderSide(wrap.Result.Side),
		Type:      domain.OrderType(wrap.Result.Type),
		Quantity:  origQty,
		Price:     priceVal,
		Status:    wrap.Result.Status,
		Timestamp: ts,
	}

	w.log.Info(ctx, "CreateOrder succeeded", ports.Fields{
		"orderID":     res.ID,
		"origQty":     origQty,
		"executedQty": exQty,
		"price":       priceVal,
		"active":      wrap.Result.Active,
		"latency_ms":  elapsed,
	})
	return res, nil
}

func (w *WallexAdapter) CancelOrder(ctx context.Context, symbol, orderID string) error {
	w.log.Info(ctx, "CancelOrder start", ports.Fields{"symbol": symbol, "orderID": orderID})
	start := time.Now()

	url := fmt.Sprintf("%s/v1/account/orders?clientOrderId=%s", w.client.baseURL, orderID)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	resp, err := w.client.Do(ctx, httpReq)
	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		w.log.Error(ctx, "CancelOrder HTTP error", ports.Fields{"error": err.Error(), "latency_ms": elapsed})
		return err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	w.log.Info(ctx, "CancelOrder response", ports.Fields{"status": resp.StatusCode, "body": string(data), "latency_ms": elapsed})

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("status %d: %s", resp.StatusCode, data)
		w.log.Error(ctx, "CancelOrder failed", ports.Fields{"error": err.Error(), "latency_ms": elapsed})
		return err
	}

	w.log.Info(ctx, "CancelOrder succeeded", ports.Fields{"orderID": orderID, "latency_ms": elapsed})
	return nil
}

func (w *WallexAdapter) GetBalance(ctx context.Context) ([]domain.Balance, error) {
	w.log.Info(ctx, "GetBalance start", nil)
	start := time.Now()

	url := fmt.Sprintf("%s/v1/account/balances", w.client.baseURL)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := w.client.Do(ctx, httpReq)
	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		w.log.Error(ctx, "GetBalance HTTP error", ports.Fields{"error": err.Error(), "latency_ms": elapsed})
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	w.log.Info(ctx, "GetBalance response", ports.Fields{"status": resp.StatusCode, "body": string(data), "latency_ms": elapsed})

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("status %d: %s", resp.StatusCode, data)
		w.log.Error(ctx, "GetBalance failed", ports.Fields{"error": err.Error(), "latency_ms": elapsed})
		return nil, err
	}

	var wrap struct {
		Success bool `json:"success"`
		Result  struct {
			Balances map[string]struct {
				Asset  string `json:"asset"`
				Fiat   bool   `json:"fiat"`
				Value  string `json:"value"`
				Locked string `json:"locked"`
			} `json:"balances"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &wrap); err != nil {
		w.log.Error(ctx, "GetBalance decode error", ports.Fields{"error": err.Error(), "latency_ms": elapsed})
		return nil, err
	}

	out := make([]domain.Balance, 0, len(wrap.Result.Balances))
	for _, b := range wrap.Result.Balances {
		total, _ := strconv.ParseFloat(b.Value, 64)
		locked, _ := strconv.ParseFloat(b.Locked, 64)
		out = append(out, domain.Balance{
			Asset:  b.Asset,
			Free:   total - locked,
			Locked: locked,
		})
	}

	w.log.Info(ctx, "GetBalance succeeded", ports.Fields{
		"count":      len(out),
		"latency_ms": elapsed,
	})
	return out, nil
}

func (w *WallexAdapter) GetOrderBook(ctx context.Context, symbol string) (domain.OrderBook, error) {
	w.log.Info(ctx, "GetOrderBook start", ports.Fields{"symbol": symbol})
	start := time.Now()

	url := fmt.Sprintf("%s/v1/depth?symbol=%s", w.client.baseURL, symbol)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := w.client.Do(ctx, httpReq)
	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		w.log.Error(ctx, "GetOrderBook HTTP error", ports.Fields{"error": err.Error(), "latency_ms": elapsed})
		return domain.OrderBook{}, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	w.log.Info(ctx, "GetOrderBook response", ports.Fields{"status": resp.StatusCode, "latency_ms": elapsed})

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("status %d: %s", resp.StatusCode, data)
		w.log.Error(ctx, "GetOrderBook failed", ports.Fields{"error": err.Error(), "latency_ms": elapsed})
		return domain.OrderBook{}, err
	}

	type rawLevel struct {
		Price    json.RawMessage `json:"price"`
		Quantity json.RawMessage `json:"quantity"`
	}
	var wrap struct {
		Success bool `json:"success"`
		Result  struct {
			Ask []rawLevel `json:"ask"`
			Bid []rawLevel `json:"bid"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &wrap); err != nil {
		w.log.Error(ctx, "GetOrderBook decode error", ports.Fields{"error": err.Error(), "latency_ms": elapsed})
		return domain.OrderBook{}, err
	}

	parseFloat := func(raw json.RawMessage) (float64, error) {
		b := bytes.Trim(raw, `"`)
		return strconv.ParseFloat(string(b), 64)
	}

	bids := make([]domain.DepthLevel, len(wrap.Result.Bid))
	for i, lvl := range wrap.Result.Bid {
		p, err := parseFloat(lvl.Price)
		if err != nil {
			w.log.Error(ctx, "parse bid price error", ports.Fields{"error": err.Error(), "latency_ms": elapsed})
			return domain.OrderBook{}, err
		}
		q, err := parseFloat(lvl.Quantity)
		if err != nil {
			w.log.Error(ctx, "parse bid quantity error", ports.Fields{"error": err.Error(), "latency_ms": elapsed})
			return domain.OrderBook{}, err
		}
		bids[i] = domain.DepthLevel{Price: p, Quantity: q}
	}
	asks := make([]domain.DepthLevel, len(wrap.Result.Ask))
	for i, lvl := range wrap.Result.Ask {
		p, err := parseFloat(lvl.Price)
		if err != nil {
			w.log.Error(ctx, "parse ask price error", ports.Fields{"error": err.Error(), "latency_ms": elapsed})
			return domain.OrderBook{}, err
		}
		q, err := parseFloat(lvl.Quantity)
		if err != nil {
			w.log.Error(ctx, "parse ask quantity error", ports.Fields{"error": err.Error(), "latency_ms": elapsed})
			return domain.OrderBook{}, err
		}
		asks[i] = domain.DepthLevel{Price: p, Quantity: q}
	}

	w.log.Info(ctx, "GetOrderBook succeeded", ports.Fields{
		"symbol":     symbol,
		"bids_count": len(bids),
		"asks_count": len(asks),
		"latency_ms": elapsed,
	})
	return domain.OrderBook{Symbol: symbol, Bids: bids, Asks: asks}, nil
}
