package testutil

import (
	"context"
	"io"
	"net/http"
	"testing"
)

func TestMockServer(t *testing.T) {
	// Create mock server
	ms := NewMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	defer ms.Close()

	// Make request
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, ms.URL, nil)
	AssertNil(t, err)
	resp, err := http.DefaultClient.Do(req) //nolint:gosec // URL is from httptest.Server (localhost)
	AssertNil(t, err)
	defer func() { _ = resp.Body.Close() }()

	AssertEqual(t, resp.StatusCode, http.StatusOK)

	body, err := io.ReadAll(resp.Body)
	AssertNil(t, err)
	AssertEqual(t, string(body), `{"status":"ok"}`)

	// Check request tracking
	AssertEqual(t, ms.RequestCount(), 1)
	lastReq := ms.LastRequest()
	AssertTrue(t, lastReq != nil)
	AssertEqual(t, lastReq.Method, "GET")
}

func TestMockServerMultipleRequests(t *testing.T) {
	ms := NewMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer ms.Close()

	// Make multiple requests
	for i := 0; i < 3; i++ {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, ms.URL, nil)
		AssertNil(t, err)
		resp, err := http.DefaultClient.Do(req) //nolint:gosec // URL is from httptest.Server (localhost)
		AssertNil(t, err)
		_ = resp.Body.Close()
	}

	AssertEqual(t, ms.RequestCount(), 3)
}

func TestMockServerReset(t *testing.T) {
	ms := NewMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer ms.Close()

	// Make request
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, ms.URL, nil)
	AssertNil(t, err)
	resp, err := http.DefaultClient.Do(req) //nolint:gosec // URL is from httptest.Server (localhost)
	AssertNil(t, err)
	_ = resp.Body.Close()

	AssertEqual(t, ms.RequestCount(), 1)

	// Reset
	ms.Reset()
	AssertEqual(t, ms.RequestCount(), 0)
	AssertTrue(t, ms.LastRequest() == nil)
}
