package workertest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	BaseURL          string
	APIKey           string
	RuntimeTaskToken string
	CallbackToken    string
	HTTPClient       *http.Client
}

func (c *Client) httpClient() *http.Client {
	if c != nil && c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 30 * time.Second}
}

func (c *Client) EngineJSON(method, path string, body any, extraHeaders map[string]string, out any) error {
	return c.doJSON(method, c.BaseURL+path, body, mergeHeaders(map[string]string{
		"X-API-Key": c.APIKey,
	}, extraHeaders), out)
}

func (c *Client) RuntimeTaskJSON(method, path string, body any, out any) error {
	return c.EngineJSON(method, path, body, map[string]string{"X-Runtime-Task-Token": c.RuntimeTaskToken}, out)
}

func (c *Client) CallbackJSON(method, path string, body any, extraHeaders map[string]string, out any) error {
	headers := mergeHeaders(map[string]string{"X-Callback-Token": c.CallbackToken}, extraHeaders)
	return c.doJSON(method, c.BaseURL+path, body, headers, out)
}

func (c *Client) doJSON(method, rawURL string, body any, headers map[string]string, out any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, rawURL, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		if strings.TrimSpace(value) != "" {
			req.Header.Set(key, value)
		}
	}
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("%s %s failed: %d %s", method, rawURL, resp.StatusCode, strings.TrimSpace(string(data)))
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}

func WaitHealthy(rawURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 3 * time.Second}
	for time.Now().Before(deadline) {
		resp, err := client.Get(rawURL)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode < 400 {
				return nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("health endpoint did not become ready within %s", timeout)
}

func QueryWithValues(path string, values map[string]string) string {
	if len(values) == 0 {
		return path
	}
	q := url.Values{}
	for key, value := range values {
		q.Set(key, value)
	}
	if strings.Contains(path, "?") {
		return path + "&" + q.Encode()
	}
	return path + "?" + q.Encode()
}

func mergeHeaders(base map[string]string, extra map[string]string) map[string]string {
	if len(extra) == 0 {
		return base
	}
	out := map[string]string{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}
