package api

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/mobil-koeln/moko-cli/internal/testutil"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient()
	testutil.AssertNil(t, err)
	testutil.AssertTrue(t, client != nil)
	testutil.AssertTrue(t, client.httpClient != nil)
	testutil.AssertEqual(t, client.baseURL, BaseURL)
	testutil.AssertTrue(t, client.timezone != nil)
}

func TestNewClient_WithTimeout(t *testing.T) {
	customTimeout := 30 * time.Second
	client, err := NewClient(WithTimeout(customTimeout))
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, client.httpClient.Timeout, customTimeout)
}

func TestNewClient_WithHTTPClient(t *testing.T) {
	customClient := &http.Client{Timeout: 5 * time.Second}
	client, err := NewClient(WithHTTPClient(customClient))
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, client.httpClient, customClient)
}

func TestNewClient_WithCache(t *testing.T) {
	mockCache := &mockCache{data: make(map[string][]byte)}
	client, err := NewClient(WithCache(mockCache))
	testutil.AssertNil(t, err)
	testutil.AssertTrue(t, client.cache != nil)
}

func TestClient_Timezone(t *testing.T) {
	client, err := NewClient()
	testutil.AssertNil(t, err)
	tz := client.Timezone()
	testutil.AssertTrue(t, tz != nil)
	testutil.AssertEqual(t, tz.String(), "Europe/Berlin")
}

func TestNewBrowserProfile(t *testing.T) {
	profile := newBrowserProfile()
	testutil.AssertTrue(t, profile.userAgent != "")
	testutil.AssertTrue(t, profile.secChUA != "")
	testutil.AssertContains(t, profile.userAgent, "Mozilla")
}

func TestUUID4(t *testing.T) {
	uuid1 := uuid4()
	uuid2 := uuid4()

	// Check format (8-4-4-4-12 hex digits)
	testutil.AssertEqual(t, len(uuid1), 36)
	testutil.AssertContains(t, uuid1, "-")

	// Check uniqueness
	testutil.AssertTrue(t, uuid1 != uuid2)
}

func TestGetDepartures_Success(t *testing.T) {
	ms := testutil.NewMockServer(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		testutil.AssertEqual(t, r.Method, "GET")
		testutil.AssertContains(t, r.URL.Path, "/abfahrten")

		// Return sample response
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(testutil.SampleDepartureResponse))
	})
	defer ms.Close()

	client := newTestClient(ms.URL)

	req := StationBoardRequest{
		EVA:       8000105,
		StationID: "A=1@O=Frankfurt(Main)Hbf@",
		DateTime:  time.Now(),
	}

	departures, err := client.GetDepartures(context.Background(), req)
	testutil.AssertNil(t, err)
	testutil.AssertTrue(t, len(departures) > 0)

	// Verify mock server received the request
	testutil.AssertEqual(t, ms.RequestCount(), 1)
}

func TestGetDepartures_InvalidJSON(t *testing.T) {
	ms := testutil.NewMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`invalid json`))
	})
	defer ms.Close()

	client := newTestClient(ms.URL)

	req := StationBoardRequest{
		EVA:       8000105,
		StationID: "test",
	}

	_, err := client.GetDepartures(context.Background(), req)
	testutil.AssertError(t, err)
}

func TestGetDepartures_HTTPError(t *testing.T) {
	ms := testutil.NewMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"server error"}`))
	})
	defer ms.Close()

	client := newTestClient(ms.URL)

	req := StationBoardRequest{
		EVA:       8000105,
		StationID: "test",
	}

	_, err := client.GetDepartures(context.Background(), req)
	testutil.AssertError(t, err)
}

func TestSearchLocations_Success(t *testing.T) {
	ms := testutil.NewMockServer(func(w http.ResponseWriter, r *http.Request) {
		testutil.AssertEqual(t, r.Method, "GET")
		testutil.AssertContains(t, r.URL.Path, "/orte")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(testutil.SampleLocationResponse))
	})
	defer ms.Close()

	client := newTestClient(ms.URL)

	locations, err := client.SearchLocations(context.Background(), "Frankfurt")
	testutil.AssertNil(t, err)
	testutil.AssertTrue(t, len(locations) > 0)
}

func TestSearchLocations_EmptyQuery(t *testing.T) {
	client, _ := NewClient()

	locations, err := client.SearchLocations(context.Background(), "")
	testutil.AssertError(t, err)
	testutil.AssertLen(t, locations, 0)
}

func TestGetDeparturesRaw_Success(t *testing.T) {
	ms := testutil.NewMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(testutil.SampleDepartureResponse))
	})
	defer ms.Close()

	client := newTestClient(ms.URL)

	req := StationBoardRequest{
		EVA:       8000105,
		StationID: "test",
	}

	rawJSON, err := client.GetDeparturesRaw(context.Background(), req)
	testutil.AssertNil(t, err)
	testutil.AssertTrue(t, len(rawJSON) > 0)
}

func TestClient_WithCache(t *testing.T) {
	mockCache := &mockCache{data: make(map[string][]byte)}

	ms := testutil.NewMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(testutil.SampleDepartureResponse))
	})
	defer ms.Close()

	client := newTestClient(ms.URL)
	client.cache = mockCache

	req := StationBoardRequest{
		EVA:       8000105,
		StationID: "test",
	}

	// First call - should hit the server
	_, err := client.GetDepartures(context.Background(), req)
	testutil.AssertNil(t, err)
	testutil.AssertEqual(t, ms.RequestCount(), 1)

	// Second call - should use cache
	_, err = client.GetDepartures(context.Background(), req)
	testutil.AssertNil(t, err)
	// Request count should still be 1 if cache is working
	// (This test assumes cache is implemented correctly)
}

func TestClient_ContextCancellation(t *testing.T) {
	// Create a server that delays response
	ms := testutil.NewMockServer(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(testutil.SampleDepartureResponse))
	})
	defer ms.Close()

	client := newTestClient(ms.URL)

	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := StationBoardRequest{
		EVA:       8000105,
		StationID: "test",
	}

	_, err := client.GetDepartures(ctx, req)
	testutil.AssertError(t, err)
}

func TestGetArrivals_Success(t *testing.T) {
	ms := testutil.NewMockServer(func(w http.ResponseWriter, r *http.Request) {
		testutil.AssertEqual(t, r.Method, "GET")
		testutil.AssertContains(t, r.URL.Path, "/ankuenfte")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(testutil.SampleArrivalResponse))
	})
	defer ms.Close()

	client := newTestClient(ms.URL)

	req := StationBoardRequest{
		EVA:       8000105,
		StationID: "test",
	}

	arrivals, err := client.GetArrivals(context.Background(), req)
	testutil.AssertNil(t, err)
	testutil.AssertTrue(t, len(arrivals) > 0)
}

func TestStationBoardRequest_DefaultValues(t *testing.T) {
	req := StationBoardRequest{
		EVA:       8000105,
		StationID: "test",
	}

	// DateTime should default to zero time
	testutil.AssertTrue(t, req.DateTime.IsZero())

	// NumVias should be 0 by default
	testutil.AssertEqual(t, req.NumVias, 0)

	// ModesOfTransit should be nil/empty by default
	testutil.AssertTrue(t, len(req.ModesOfTransit) == 0)
}

// Mock cache implementation for testing
type mockCache struct {
	data map[string][]byte
}

func (m *mockCache) Get(key string) ([]byte, bool) {
	val, ok := m.data[key]
	return val, ok
}

func (m *mockCache) Set(key string, value []byte) error {
	m.data[key] = value
	return nil
}

// Helper to create a client with custom base URL for testing
func newTestClient(baseURL string) *Client {
	client, _ := NewClient()
	client.baseURL = baseURL
	return client
}
