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

type RefreshResponse struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
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
	c := &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    baseURL,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		log:        log,
	}

	ctx := context.Background()
	if err := c.authenticate(ctx); err != nil {
		log.Error(ctx, "initial auth failed", ports.Fields{"error": err.Error()})
	}

	go c.autoRefreshLoop(ctx)

	return c
}

func (c *Client) authenticate(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Until(c.expiry) > time.Minute {
		c.log.Info(ctx, "auth: token still valid", nil)
		return nil
	}

	url := fmt.Sprintf("%s/api/v1/usr/authenticate/", c.baseURL)
	payload := map[string]string{
		"api_key":    c.apiKey,
		"secret_key": c.apiSecret,
	}
	bodyBytes, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Error(ctx, "auth: http error", ports.Fields{"error": err.Error()})
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.log.Error(ctx, "auth: non-OK status", ports.Fields{"status": resp.StatusCode})
		return fmt.Errorf("auth failed: status %d", resp.StatusCode)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var ar AuthResponse
	if err := json.Unmarshal(respBody, &ar); err != nil {
		c.log.Error(ctx, "auth: decode error", ports.Fields{"error": err.Error()})
		return err
	}

	c.token = ar.Access
	c.expiry = time.Now().Add(15 * time.Minute)
	c.log.Info(ctx, "auth: token acquired", ports.Fields{"expiry": c.expiry})

	return nil
}

func (c *Client) refreshToken(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	url := fmt.Sprintf("%s/api/v1/usr/refresh_token/", c.baseURL)
	payload := map[string]string{"refresh": c.token}
	bodyBytes, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Error(ctx, "refresh: http error", ports.Fields{"error": err.Error()})
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.log.Error(ctx, "refresh: non-OK status", ports.Fields{"status": resp.StatusCode})
		return fmt.Errorf("refresh failed: status %d", resp.StatusCode)
	}

	respBody, _ := io.ReadAll(resp.Body)
	var rr RefreshResponse
	if err := json.Unmarshal(respBody, &rr); err != nil {
		c.log.Error(ctx, "refresh: decode error", ports.Fields{"error": err.Error()})
		return err
	}

	c.token = rr.Access
	c.expiry = time.Now().Add(15 * time.Minute)
	c.log.Info(ctx, "refresh: token updated", ports.Fields{"expiry": c.expiry})

	return nil
}

func (c *Client) autoRefreshLoop(ctx context.Context) {
	for {
		c.mu.Lock()
		wait := time.Until(c.expiry) - time.Minute
		c.mu.Unlock()

		if wait < 0 {
			wait = 10 * time.Second
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(wait):
			if err := c.refreshToken(ctx); err != nil {
				_ = c.authenticate(ctx)
			}
		}
	}
}

func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if time.Until(c.expiry) < 30*time.Second {
		_ = c.refreshToken(ctx)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	return c.httpClient.Do(req)
}
