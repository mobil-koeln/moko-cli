package testutil

import (
	"net/http"
	"net/http/httptest"
)

// MockServer wraps httptest.Server with convenience methods
type MockServer struct {
	*httptest.Server
	Requests []*http.Request
}

// NewMockServer creates a new mock HTTP server
func NewMockServer(handler http.HandlerFunc) *MockServer {
	ms := &MockServer{
		Requests: make([]*http.Request, 0),
	}

	ms.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ms.Requests = append(ms.Requests, r)
		handler(w, r)
	}))

	return ms
}

// LastRequest returns the most recent request
func (ms *MockServer) LastRequest() *http.Request {
	if len(ms.Requests) == 0 {
		return nil
	}
	return ms.Requests[len(ms.Requests)-1]
}

// RequestCount returns the number of requests received
func (ms *MockServer) RequestCount() int {
	return len(ms.Requests)
}

// Reset clears the request history
func (ms *MockServer) Reset() {
	ms.Requests = make([]*http.Request, 0)
}
