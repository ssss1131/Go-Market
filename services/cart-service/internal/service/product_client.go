package service

import (
	"fmt"
	"net/http"
)

type ProductHTTPClient struct {
	BaseURL string
}

func NewProductHTTPClient(baseURL string) *ProductHTTPClient {
	return &ProductHTTPClient{BaseURL: baseURL}
}

func (c *ProductHTTPClient) CheckProductExists(productID uint) bool {
	url := fmt.Sprintf("%s/products/%d", c.BaseURL, productID)

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return false
	}

	return true
}
