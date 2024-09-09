package dhis2

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

const DefaultConnectionTimeout = 1 * time.Second

type Client struct {
	Username          string
	Password          string
	BaseURL           string
	ConnectionTimeout time.Duration
}

type Info struct {
	ContextPath string `json:"contextPath"`
	Version     string `json:"version"`
	Revision    string `json:"revision"`
	BuildTime   string `json:"buildTime"`
}

// get dhis2 info
func (c *Client) GetInfo() (*Info, error) {
	info := &Info{}
	err := c.doRequest("/api/system/info", info)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// common method to do request
func (c *Client) doRequest(path string, result interface{}) error {
	url := fmt.Sprintf("%s%s", c.BaseURL, path)

	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Encode credentials
	credentials := base64.StdEncoding.EncodeToString([]byte(c.Username + ":" + c.Password))

	// Set Authorization header
	req.Header.Set("Authorization", "Basic "+credentials)

	// Custom Transport with Connect Timeout
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: c.ConnectionTimeout,
		}).DialContext,
	}

	// Make the request
	client := &http.Client{
		Transport: transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check if the status code is not 200
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read the body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Unmarshal JSON response into result
	err = json.Unmarshal(body, result)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return nil
}
