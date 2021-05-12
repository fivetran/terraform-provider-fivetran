package gofivetran

import (
	"encoding/base64"
	"fmt"
	"time"
)

var BaseURL string = "https://beta-api.fivetran.com/v1"

type Client struct {
	BaseURL       string
	Timeout       time.Duration
	Authorization string
}

func NewClient(apiKey string, apiSecret string) *Client {
	return &Client{
		BaseURL:       BaseURL,
		Timeout:       Timeout,
		Authorization: fmt.Sprintf("Basic %v", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", apiKey, apiSecret)))),
	}
}
