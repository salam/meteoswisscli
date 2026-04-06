package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	bulletinBaseURL    = "https://aws.slf.ch"
	measurementBaseURL = "https://measurement-api.slf.ch"
)

type Client struct {
	http            *http.Client
	bulletinBase    string
	measurementBase string
	lang            string
}

func NewClient(lang string) *Client {
	return &Client{
		http:            &http.Client{},
		bulletinBase:    bulletinBaseURL,
		measurementBase: measurementBaseURL,
		lang:            lang,
	}
}

func NewClientWithBase(baseURL, lang string) *Client {
	return &Client{
		http:            &http.Client{},
		bulletinBase:    baseURL,
		measurementBase: baseURL,
		lang:            lang,
	}
}

func (c *Client) DoJSON(method, url string, result any) error {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "SwissCLI/1.0")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("could not reach API. Check your internet connection")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}
