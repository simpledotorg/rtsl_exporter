package alphasms

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const BaseUrl string = "https://api.sms.net.bd"

type Client struct {
	APIKey string
}

type BalanceElement struct {
	Balance  string `json:"balance"`
	Validity string `json:"validity"`
}

type ApiResponse struct {
	Error int             `json:"error"`
	Msg   string          `json:"msg"`
	Data  json.RawMessage `json:"data"`
}

// method to get user balance
func (c *Client) GetUserBalance() (*ApiResponse, *BalanceElement, error) {
	apiResp := &ApiResponse{}
	err := c.doRequest("/user/balance", apiResp)
	if err != nil {
		return nil, nil, err
	}

	var balanceData BalanceElement
	err = json.Unmarshal(apiResp.Data, &balanceData)
	if err != nil {
		return apiResp, nil, err
	}

	return apiResp, &balanceData, nil
}

// common method to do request
func (c *Client) doRequest(path string, result interface{}) error {
	url := fmt.Sprintf("%s%s?api_key=%s", BaseUrl, path, c.APIKey)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, result)

	return err
}
