package httpclient

import (
	"net/http"
	"time"
)

type Option func(*Client)

// WithBaseURL set base URL, sehingga endpoint di request cukup path relatif
// contoh: "/auth/login" -> akan digabung jadi "https://api.example.com/auth/login"
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithTimeout mengatur timeout per-request (default 15 detik)
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithHeader menambahkan satu header default yang dikirim di setiap request
// contoh: WithHeader("Authorization", "Bearer xxx")
func WithHeader(key, value string) Option {
	return func(c *Client) {
		c.headers[key] = value
	}
}

// WithHeaders menambahkan banyak header sekaligus
func WithHeaders(headers map[string]string) Option {
	return func(c *Client) {
		for k, v := range headers {
			c.headers[k] = v
		}
	}
}

// WithRetry mengaktifkan retry otomatis untuk status 429 dan 5xx
// maxRetries: jumlah percobaan ulang (di luar percobaan pertama)
// waitMin/waitMax: rentang exponential backoff
func WithRetry(maxRetries int, waitMin, waitMax time.Duration) Option {
	return func(c *Client) {
		c.maxRetries = maxRetries
		c.retryWaitMin = waitMin
		c.retryWaitMax = waitMax
	}
}

// WithHTTPClient mengganti http.Client bawaan (misal untuk custom transport/proxy)
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}
