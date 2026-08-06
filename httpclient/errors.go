package httpclient

import "fmt"

// APIError merepresentasikan error dari response HTTP dengan status code >= 400
type APIError struct {
	StatusCode int
	Body       []byte
	Method     string
	URL        string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("httpclient: %s %s gagal dengan status %d: %s", e.Method, e.URL, e.StatusCode, string(e.Body))
}

// IsClientError true jika status code 4xx
func (e *APIError) IsClientError() bool {
	return e.StatusCode >= 400 && e.StatusCode < 500
}

// IsServerError true jika status code 5xx
func (e *APIError) IsServerError() bool {
	return e.StatusCode >= 500
}
