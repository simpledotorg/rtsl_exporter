package dhis2

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const httpTimeOutSec = 1

type Client struct {
	Username string
	Password string
	BaseURL  string
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
	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), httpTimeOutSec*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s%s", c.BaseURL, path)

	// Create a new request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	// Encode credentials
	credentials := base64.StdEncoding.EncodeToString([]byte(c.Username + ":" + c.Password))

	// Set Authorization header
	req.Header.Set("Authorization", "Basic "+credentials)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read the body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Unmarshal JSON response into result
	err = json.Unmarshal(body, result)
	return err
}
