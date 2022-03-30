package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
)

type BaseClient interface {
	NewRequest(method string, path string, body []byte) (*http.Request, error)
	Do(req *http.Request) (AWSConfig, *http.Response, error)
}

type ClientConfig struct {
	BaseURL    string
	httpClient *http.Client
}

func NewBaseClient(url string, httpClient *http.Client) BaseClient {
	return &ClientConfig{
		BaseURL:    url,
		httpClient: httpClient,
	}
}

func (c *ClientConfig) NewRequest(method string, path string, body []byte) (*http.Request, error) {
	url := "http://" + c.BaseURL + path
	var buf io.ReadWriter
	if body != nil {
		buf = bytes.NewBuffer(body)
	}
	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	return req, nil
}
func (c *ClientConfig) Do(req *http.Request) (AWSConfig, *http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return AWSConfig{}, nil, err
	}
	if resp.StatusCode != 200 {
		return AWSConfig{}, nil, errors.New("Response with status code: " + strconv.Itoa(resp.StatusCode))
	}
	defer resp.Body.Close()
	var config AWSConfig
	err = json.NewDecoder(resp.Body).Decode(&config)
	return config, resp, err
}
