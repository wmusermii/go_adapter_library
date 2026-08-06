package httpclient

import "encoding/base64"

// requestConfig menampung override khusus untuk satu request (bukan default Client)
type requestConfig struct {
	headers map[string]string
}

// RequestOption adalah opsi yang berlaku hanya untuk satu kali panggilan request,
// tidak mengubah konfigurasi default Client.
type RequestOption func(*requestConfig)

// WithRequestHeader menambahkan/override satu header untuk request ini saja
// contoh: WithRequestHeader("X-Request-Id", "abc-123")
func WithRequestHeader(key, value string) RequestOption {
	return func(rc *requestConfig) {
		rc.headers[key] = value
	}
}

// WithRequestHeaders menambahkan/override banyak header sekaligus untuk request ini saja
func WithRequestHeaders(headers map[string]string) RequestOption {
	return func(rc *requestConfig) {
		for k, v := range headers {
			rc.headers[k] = v
		}
	}
}

// WithBasicAuth menambahkan header Authorization: Basic <base64(username:password)>
// untuk request ini saja
func WithBasicAuth(username, password string) RequestOption {
	return func(rc *requestConfig) {
		raw := username + ":" + password
		encoded := base64.StdEncoding.EncodeToString([]byte(raw))
		rc.headers["Authorization"] = "Basic " + encoded
	}
}

// WithBearerToken menambahkan header Authorization: Bearer <token> untuk request ini saja
// berguna kalau token berbeda per user/per request (misal token dari JWT login user)
func WithBearerToken(token string) RequestOption {
	return func(rc *requestConfig) {
		rc.headers["Authorization"] = "Bearer " + token
	}
}

func newRequestConfig(opts ...RequestOption) *requestConfig {
	rc := &requestConfig{
		headers: make(map[string]string),
	}
	for _, opt := range opts {
		opt(rc)
	}
	return rc
}
