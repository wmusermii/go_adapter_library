package httpclient

import "net/http"

// Response adalah hasil mentah dari request, selain hasil unmarshal ke struct target
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// IsSuccess true jika status code 2xx
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}
