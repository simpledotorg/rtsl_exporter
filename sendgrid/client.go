package sendgrid

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Client struct {
	APIKeys map[string]string
}

// Metrics struct maps the SendGrid API response.
type Metrics struct {
	Total     float64 `json:"total"`
	Remaining float64 `json:"remain"`
	Used      float64 `json:"used"`
	NextReset string  `json:"next_reset"`
}

// NewClient creates a new Client instance.
func NewClient(apiKeys map[string]string) *Client {
	return &Client{APIKeys: apiKeys}
}

// FetchMetrics retrieves metrics from the SendGrid API for a specific account.
func (c *Client) FetchMetrics(accountName string) (*Metrics, int, time.Duration, error) {
	apiKey, exists := c.APIKeys[accountName]
	if !exists {
		return nil, 0, 0, fmt.Errorf("API key for account %s not found", accountName)
	}

	start := time.Now()
	url := "https://api.sendgrid.com/v3/user/credits"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	duration := time.Since(start)
	if err != nil {
		return nil, 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Received non-200 response code: %d", resp.StatusCode)
		return nil, resp.StatusCode, duration, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var metrics Metrics
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return nil, resp.StatusCode, duration, err
	}

	return &metrics, resp.StatusCode, duration, nil
}
