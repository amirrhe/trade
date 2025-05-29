package bitpin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"trade/internal/ports"
)

type AuthResponse struct {
	Refresh string `json:"refresh"`
	Access  string `json:"access"`
}

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
	apiSecret  string

	mu     sync.Mutex
	token  string
	expiry time.Time

	log ports.LoggerPort
}

func NewClient(apiKey, apiSecret, baseURL string, log ports.LoggerPort) *Client {
	if baseURL == "" {
		baseURL = "https://api.bitpin.ir"
	}
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		log:        log,
	}
}

func (c *Client) authenticate(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Until(c.expiry) > time.Minute {
		c.log.Info(ctx, "auth: token still valid", nil)
		return nil
	}

	url := fmt.Sprintf("%s/api/v1/usr/authenticate/", c.baseURL)
	payload := map[string]string{"api_key": c.apiKey, "secret_key": c.apiSecret}
	bodyBytes, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Error(ctx, "auth: http error", ports.Fields{"error": err.Error()})
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("auth failed: status %d", resp.StatusCode)
		c.log.Error(ctx, "auth: non-OK status", ports.Fields{"status": resp.StatusCode})
		return err
	}

	var ar AuthResponse
	if err := json.Unmarshal(respBody, &ar); err != nil {
		c.log.Error(ctx, "auth: decode error", ports.Fields{"error": err.Error()})
		return err
	}

	c.token = ar.Access
	c.expiry = time.Now().Add(15 * time.Minute)
	return nil
}

func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if err := c.authenticate(ctx); err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Error(ctx, "http: request error", ports.Fields{"error": err.Error()})
		return nil, err
	}

	bodyBytes, _ := io.ReadAll(resp.Body)

	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return resp, nil
}
