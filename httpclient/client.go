package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

// Client adalah HTTP client yang bisa dipakai lintas project.
// Cukup panggil Get/Post/Put/Patch/Delete dengan endpoint, dan payload jika ada.
type Client struct {
	baseURL      string
	httpClient   *http.Client
	headers      map[string]string
	maxRetries   int
	retryWaitMin time.Duration
	retryWaitMax time.Duration
}

// New membuat instance Client baru dengan konfigurasi default,
// lalu menerapkan semua Option yang diberikan.
func New(opts ...Option) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		headers: map[string]string{
			"Content-Type": "application/json",
			"Accept":       "application/json",
		},
		maxRetries:   0,
		retryWaitMin: 500 * time.Millisecond,
		retryWaitMax: 5 * time.Second,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Get melakukan request GET. result diisi otomatis dari JSON response (boleh nil).
// reqOpts opsional: WithBasicAuth, WithBearerToken, WithRequestHeader, dst.
func (c *Client) Get(ctx context.Context, endpoint string, result interface{}, reqOpts ...RequestOption) (*Response, error) {
	return c.Request(ctx, http.MethodGet, endpoint, nil, result, reqOpts...)
}

// Post melakukan request POST dengan payload JSON (boleh nil jika tidak ada body).
func (c *Client) Post(ctx context.Context, endpoint string, payload interface{}, result interface{}, reqOpts ...RequestOption) (*Response, error) {
	return c.Request(ctx, http.MethodPost, endpoint, payload, result, reqOpts...)
}

// Put melakukan request PUT dengan payload JSON.
func (c *Client) Put(ctx context.Context, endpoint string, payload interface{}, result interface{}, reqOpts ...RequestOption) (*Response, error) {
	return c.Request(ctx, http.MethodPut, endpoint, payload, result, reqOpts...)
}

// Patch melakukan request PATCH dengan payload JSON.
func (c *Client) Patch(ctx context.Context, endpoint string, payload interface{}, result interface{}, reqOpts ...RequestOption) (*Response, error) {
	return c.Request(ctx, http.MethodPatch, endpoint, payload, result, reqOpts...)
}

// Delete melakukan request DELETE. payload opsional (boleh nil).
func (c *Client) Delete(ctx context.Context, endpoint string, payload interface{}, result interface{}, reqOpts ...RequestOption) (*Response, error) {
	return c.Request(ctx, http.MethodDelete, endpoint, payload, result, reqOpts...)
}

// Request adalah method inti: method, endpoint, payload (opsional), result (opsional),
// dan reqOpts untuk override header per-request (basic auth, bearer token, header custom).
// Header di reqOpts akan menimpa header default Client jika key-nya sama.
func (c *Client) Request(ctx context.Context, method, endpoint string, payload interface{}, result interface{}, reqOpts ...RequestOption) (*Response, error) {
	fullURL := c.buildURL(endpoint)
	rc := newRequestConfig(reqOpts...)

	var rawBody []byte
	if payload != nil {
		jsonBody, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("httpclient: gagal marshal payload: %w", err)
		}
		rawBody = jsonBody
	}

	attempts := c.maxRetries + 1
	var resp *Response
	var err error

	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			wait := c.calculateBackoff(attempt)
			select {
			case <-time.After(wait):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		var bodyReader io.Reader
		if rawBody != nil {
			bodyReader = bytes.NewReader(rawBody)
		}

		resp, err = c.doRequest(ctx, method, fullURL, bodyReader, rc.headers)
		if err != nil {
			if attempt == attempts-1 {
				return nil, err
			}
			continue
		}

		if !c.shouldRetry(resp.StatusCode) {
			break
		}
	}

	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return resp, &APIError{
			StatusCode: resp.StatusCode,
			Body:       resp.Body,
			Method:     method,
			URL:        fullURL,
		}
	}

	if result != nil && len(resp.Body) > 0 {
		if err := json.Unmarshal(resp.Body, result); err != nil {
			return resp, fmt.Errorf("httpclient: gagal unmarshal response: %w", err)
		}
	}

	return resp, nil
}

// doRequest menggabungkan header default Client dengan header override per-request.
// Header override (extraHeaders) menang jika ada key yang sama.
func (c *Client) doRequest(ctx context.Context, method, url string, body io.Reader, extraHeaders map[string]string) (*Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("httpclient: gagal membuat request: %w", err)
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("httpclient: gagal eksekusi request: %w", err)
	}
	defer httpResp.Body.Close()

	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("httpclient: gagal membaca response body: %w", err)
	}

	return &Response{
		StatusCode: httpResp.StatusCode,
		Headers:    httpResp.Header,
		Body:       bodyBytes,
	}, nil
}

func (c *Client) shouldRetry(statusCode int) bool {
	if c.maxRetries == 0 {
		return false
	}
	return statusCode == http.StatusTooManyRequests || statusCode >= 500
}

func (c *Client) calculateBackoff(attempt int) time.Duration {
	wait := time.Duration(math.Pow(2, float64(attempt-1))) * c.retryWaitMin
	if wait > c.retryWaitMax {
		wait = c.retryWaitMax
	}
	return wait
}

func (c *Client) buildURL(endpoint string) string {
	if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return endpoint
	}
	if c.baseURL == "" {
		return endpoint
	}
	return strings.TrimRight(c.baseURL, "/") + "/" + strings.TrimLeft(endpoint, "/")
}
