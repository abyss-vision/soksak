package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is a thin HTTP client for the Soksak server API.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

// NewClient returns a new Client targeting baseURL, authenticating with token.
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// testClientOverride allows tests to inject a mock client without touching Viper.
var testClientOverride *Client

// NewClientFromConfig builds a Client from the current Viper configuration.
// Tests can set testClientOverride to bypass Viper entirely.
func NewClientFromConfig() *Client {
	if testClientOverride != nil {
		return testClientOverride
	}
	cfg := GetConfig()
	return NewClient(cfg.ServerURL, cfg.ServerToken)
}

// Get performs a GET request and JSON-decodes the response into dst.
func (c *Client) Get(path string, dst any) error {
	return c.do(http.MethodGet, path, nil, dst)
}

// Post performs a POST request with body JSON-encoded, decoding the response into dst.
func (c *Client) Post(path string, body, dst any) error {
	return c.do(http.MethodPost, path, body, dst)
}

// Patch performs a PATCH request with body JSON-encoded, decoding the response into dst.
func (c *Client) Patch(path string, body, dst any) error {
	return c.do(http.MethodPatch, path, body, dst)
}

// Delete performs a DELETE request, decoding any response into dst.
func (c *Client) Delete(path string, dst any) error {
	return c.do(http.MethodDelete, path, nil, dst)
}

func (c *Client) do(method, path string, body, dst any) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	url := c.baseURL + "/" + strings.TrimLeft(path, "/")
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("http %s %s: %w", method, url, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		// Try to extract a message from the JSON body.
		var errBody struct {
			Message string `json:"message"`
			Error   string `json:"error"`
		}
		if json.Unmarshal(respBody, &errBody) == nil {
			msg := errBody.Message
			if msg == "" {
				msg = errBody.Error
			}
			if msg != "" {
				return fmt.Errorf("server error %d: %s", resp.StatusCode, msg)
			}
		}
		return fmt.Errorf("server error %d: %s", resp.StatusCode, string(respBody))
	}

	if dst != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, dst); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}
