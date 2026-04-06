package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/salam/swissmeteocli/pkg/cache"
)

const defaultBaseURL = "https://app-prod-ws.meteoswiss-app.ch"
const openDataBaseURL = "https://data.geo.admin.ch"

type Client struct {
	http    *http.Client
	baseURL string
	lang    string
	cache   *cache.Cache
}

func NewClient(lang string) *Client {
	return &Client{
		http:    &http.Client{},
		baseURL: defaultBaseURL,
		lang:    lang,
	}
}

func NewClientWithCache(lang string, c *cache.Cache) *Client {
	return &Client{
		http:    &http.Client{},
		baseURL: defaultBaseURL,
		lang:    lang,
		cache:   c,
	}
}

func (c *Client) DoJSON(method, path string, body any, result any) error {
	url := c.baseURL + path
	cacheKey := method + " " + url

	// Check cache for GET requests without body
	if c.cache != nil && method == "GET" && body == nil {
		if data, ok := c.cache.Get(cacheKey); ok && result != nil {
			return json.Unmarshal(data, result)
		}
	}

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, reqBody)
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	// Cache successful GET responses
	if c.cache != nil && method == "GET" && body == nil {
		c.cache.Set(cacheKey, respBody)
	}

	if result != nil {
		return json.Unmarshal(respBody, result)
	}
	return nil
}

func (c *Client) DoRaw(method, url string) ([]byte, error) {
	cacheKey := method + " " + url

	// Cache DoRaw for non-HDF5 requests (skip large binary files)
	canCache := c.cache != nil && method == "GET" && !strings.HasSuffix(url, ".h5")
	if canCache {
		if data, ok := c.cache.Get(cacheKey); ok {
			return data, nil
		}
	}

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

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if canCache {
		c.cache.Set(cacheKey, data)
	}

	return data, nil
}
