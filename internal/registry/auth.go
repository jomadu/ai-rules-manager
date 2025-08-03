package registry

import (
	"net/http"
)

// AuthProvider defines the interface for authentication methods
type AuthProvider interface {
	SetAuth(req *http.Request)
}

// BearerAuth implements Bearer token authentication
type BearerAuth struct {
	Token string
}

// SetAuth adds Bearer token to request headers
func (b *BearerAuth) SetAuth(req *http.Request) {
	if b.Token != "" {
		req.Header.Set("Authorization", "Bearer "+b.Token)
	}
}

// BasicAuth implements Basic authentication
type BasicAuth struct {
	Username string
	Password string
}

// SetAuth adds Basic auth to request headers
func (b *BasicAuth) SetAuth(req *http.Request) {
	if b.Username != "" || b.Password != "" {
		req.SetBasicAuth(b.Username, b.Password)
	}
}

// HeaderAuth implements custom header authentication
type HeaderAuth struct {
	Header string
	Value  string
}

// SetAuth adds custom header to request
func (h *HeaderAuth) SetAuth(req *http.Request) {
	if h.Header != "" && h.Value != "" {
		req.Header.Set(h.Header, h.Value)
	}
}
