package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const defaultBaseURL = "https://app-prod-ws.meteoswiss-app.ch"
const openDataBaseURL = "https://data.geo.admin.ch"

type Client struct {
	http    *http.Client
	baseURL string
	lang    string
}

func NewClient(lang string) *Client {
	return &Client{
		http:    &http.Client{},
		baseURL: defaultBaseURL,
		lang:    lang,
	}
}

func (c *Client) DoJSON(method, path string, body any, result any) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", c.lang)
	req.Header.Set("User-Agent", "SwissCLI/1.0")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("could not reach API. Check your internet connection")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (c *Client) DoRaw(method, url string) ([]byte, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "SwissCLI/1.0")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach API. Check your internet connection")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
