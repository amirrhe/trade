package wallex

import (
	"context"
	"net/http"
	"time"

	"trade/internal/ports"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	log        ports.LoggerPort
}

func NewClient(apiKey, baseURL string, log ports.LoggerPort) *Client {
	if baseURL == "" {
		baseURL = "https://api.wallex.ir"
	}
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
		apiKey:     apiKey,
		log:        log,
	}
}

func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	req.Header.Set("X-API-Key", c.apiKey) //
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Error(ctx, "http: request error", ports.Fields{
			"error": err.Error(),
		})
		return nil, err
	}
	return resp, nil
}
