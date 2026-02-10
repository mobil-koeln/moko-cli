package api

import (
	"context"
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/mobil-koeln/moko-cli/internal/cache"
	"github.com/mobil-koeln/moko-cli/internal/models"
)

const (
	defaultTimeout  = 10 * time.Second
	defaultCacheTTL = 90 * time.Second
)

// browserProfile holds a consistent browser identity for a client session.
type browserProfile struct {
	userAgent string
	secChUA   string // sec-ch-ua header value matching the UA
	mobile    bool
}

var userAgentTemplates = []struct {
	ua     string
	major  int // Chrome major version for sec-ch-ua
	mobile bool
}{
	{"Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.XXXX.YYY Mobile Safari/537.36", 114, true},
	{"Mozilla/5.0 (Linux; Android 14; SM-S928B/DS) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.XXXX.YYY Mobile Safari/537.36", 120, true},
	{"Mozilla/5.0 (Linux; Android 14; Pixel 9 Pro Build/AD1A.240418.003; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/124.0.XXXX.YYY Mobile Safari/537.36", 124, true},
	{"Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.XXXX.YYY Mobile Safari/537.36", 112, true},
	{"Mozilla/5.0 (Linux; Android 15; moto g - 2025 Build/V1VK35.22-13-2; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/132.0.XXXX.YYY Mobile Safari/537.36", 132, true},
	{"Mozilla/5.0 (X11; CrOS x86_64 14541.0.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.XXXX.YYY Safari/537.36", 134, false},
	{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.XXXX.YYY Safari/537.36", 131, false},
	{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.XXXX.YYY Safari/537.36", 129, false},
}

// cryptoRandIntn returns a random integer [0, n) using crypto/rand
func cryptoRandIntn(n int) int {
	nBig, err := crand.Int(crand.Reader, big.NewInt(int64(n)))
	if err != nil {
		// Fallback to time-based seed if crypto/rand fails (very unlikely)
		return int(time.Now().UnixNano() % int64(n))
	}
	return int(nBig.Int64())
}

// newBrowserProfile generates a randomized but internally-consistent browser identity.
func newBrowserProfile() browserProfile {
	tmpl := userAgentTemplates[cryptoRandIntn(len(userAgentTemplates))]

	major := cryptoRandIntn(1000)
	minor := cryptoRandIntn(100)
	ua := strings.NewReplacer("XXXX", fmt.Sprintf("%d", major), "YYY", fmt.Sprintf("%d", minor)).Replace(tmpl.ua)

	secChUA := fmt.Sprintf(`"Chromium";v="%d", "Not?A_Brand";v="24", "Google Chrome";v="%d"`, tmpl.major, tmpl.major)

	return browserProfile{
		userAgent: ua,
		secChUA:   secChUA,
		mobile:    tmpl.mobile,
	}
}

// uuid4 generates a random UUID v4 string.
func uuid4() string {
	var b [16]byte
	if _, err := crand.Read(b[:]); err != nil {
		// This should never happen with crypto/rand, but provide a time-based fallback
		now := time.Now().UnixNano()
		// #nosec G115 -- intentional overflow for fallback UUID generation
		return fmt.Sprintf("%08x-%04x-4%03x-%04x-%012x",
			uint32(now), uint16(now>>32)&0xffff, uint16(now>>48)&0xfff,
			uint16((now>>16)&0x3fff)|0x8000, uint64(now)&0xffffffffffff)
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 2
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// Cache interface for caching HTTP responses
type Cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, value []byte) error
}

// Client is the API client for bahn.de
type Client struct {
	httpClient *http.Client
	baseURL    string
	timezone   *time.Location
	cache      Cache
	browser    browserProfile
}

// ClientOption configures the Client
type ClientOption func(*Client)

// WithTimeout sets the HTTP client timeout
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// WithCache enables caching with the provided cache implementation
func WithCache(cache Cache) ClientOption {
	return func(c *Client) {
		c.cache = cache
	}
}

// WithDefaultCache enables caching with the default file cache
func WithDefaultCache() ClientOption {
	return func(c *Client) {
		fc, err := cache.NewFileCache(cache.DefaultCacheDir(), defaultCacheTTL)
		if err == nil {
			c.cache = fc
		}
	}
}

// NewClient creates a new API client
func NewClient(opts ...ClientOption) (*Client, error) {
	tz, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		return nil, fmt.Errorf("failed to load timezone: %w", err)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	c := &Client{
		httpClient: &http.Client{
			Timeout: defaultTimeout,
			Jar:     jar,
		},
		baseURL:  BaseURL,
		timezone: tz,
		browser:  newBrowserProfile(),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// Timezone returns the client's timezone
func (c *Client) Timezone() *time.Location {
	return c.timezone
}

// StationBoardRequest contains parameters for a departure/arrival query
type StationBoardRequest struct {
	EVA            int64     // Station EVA number (required)
	StationID      string    // Station ID (required)
	DateTime       time.Time // Query time (defaults to now)
	NumVias        int       // Number of via stations (default: 5)
	ModesOfTransit []string  // Filter by transport mode (default: all)
}

// DepartureRequest is an alias for StationBoardRequest for backward compatibility
type DepartureRequest = StationBoardRequest

// GetDepartures fetches departures for a station
func (c *Client) GetDepartures(ctx context.Context, req StationBoardRequest) ([]models.Departure, error) {
	body, err := c.GetDeparturesRaw(ctx, req)
	if err != nil {
		return nil, err
	}

	// Parse response
	var resp models.DeparturesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse departures response: %w", err)
	}

	// Convert to domain models
	departures := make([]models.Departure, 0, len(resp.Entries))
	for _, entry := range resp.Entries {
		departures = append(departures, *entry.ToDeparture(c.timezone))
	}

	return departures, nil
}

// GetDeparturesRaw fetches departures and returns raw JSON
func (c *Client) GetDeparturesRaw(ctx context.Context, req StationBoardRequest) (json.RawMessage, error) {
	return c.getStationBoardRaw(ctx, req, EndpointDepartures)
}

// GetArrivals fetches arrivals for a station
func (c *Client) GetArrivals(ctx context.Context, req StationBoardRequest) ([]models.Departure, error) {
	body, err := c.GetArrivalsRaw(ctx, req)
	if err != nil {
		return nil, err
	}

	// Parse response (same format as departures)
	var resp models.DeparturesResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse arrivals response: %w", err)
	}

	// Convert to domain models
	arrivals := make([]models.Departure, 0, len(resp.Entries))
	for _, entry := range resp.Entries {
		arrivals = append(arrivals, *entry.ToDeparture(c.timezone))
	}

	return arrivals, nil
}

// GetArrivalsRaw fetches arrivals and returns raw JSON
func (c *Client) GetArrivalsRaw(ctx context.Context, req StationBoardRequest) (json.RawMessage, error) {
	return c.getStationBoardRaw(ctx, req, EndpointArrivals)
}

// getStationBoardRaw is a helper for fetching departures/arrivals
func (c *Client) getStationBoardRaw(ctx context.Context, req StationBoardRequest, endpoint string) (json.RawMessage, error) {
	// Use current time if not specified
	dt := req.DateTime
	if dt.IsZero() {
		dt = time.Now().In(c.timezone)
	}

	// Build query parameters
	params := url.Values{}
	params.Set("datum", dt.Format("2006-01-02"))
	params.Set("zeit", dt.Format("15:04:00"))
	params.Set("ortExtId", fmt.Sprintf("%d", req.EVA))
	params.Set("ortId", req.StationID)
	params.Set("mitVias", "true")

	numVias := req.NumVias
	if numVias == 0 {
		numVias = 5
	}
	params.Set("maxVias", fmt.Sprintf("%d", numVias))

	// Set modes of transit
	mots := req.ModesOfTransit
	if len(mots) == 0 {
		mots = ModesOfTransit
	}
	for _, mot := range mots {
		params.Add("verkehrsmittel[]", mot)
	}

	reqURL := c.baseURL + endpoint + "?" + params.Encode()

	return c.doRequest(ctx, reqURL)
}

// NearbyRequest contains parameters for a nearby search
type NearbyRequest struct {
	Latitude  float64 // Latitude (required)
	Longitude float64 // Longitude (required)
	Radius    int     // Search radius in meters (default: 9999)
	MaxNo     int     // Maximum number of results (default: 100)
}

// SearchNearby searches for stations near a location
func (c *Client) SearchNearby(ctx context.Context, req NearbyRequest) ([]models.Location, error) {
	body, err := c.SearchNearbyRaw(ctx, req)
	if err != nil {
		return nil, err
	}

	var resp []models.LocationResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse nearby response: %w", err)
	}

	locations := make([]models.Location, 0, len(resp))
	for _, entry := range resp {
		locations = append(locations, *entry.ToLocation())
	}

	return locations, nil
}

// SearchNearbyRaw searches for nearby stations and returns raw JSON
func (c *Client) SearchNearbyRaw(ctx context.Context, req NearbyRequest) (json.RawMessage, error) {
	radius := req.Radius
	if radius == 0 {
		radius = 9999
	}
	maxNo := req.MaxNo
	if maxNo == 0 {
		maxNo = 100
	}

	params := url.Values{}
	params.Set("lat", fmt.Sprintf("%f", req.Latitude))
	params.Set("long", fmt.Sprintf("%f", req.Longitude))
	params.Set("radius", fmt.Sprintf("%d", radius))
	params.Set("maxNo", fmt.Sprintf("%d", maxNo))

	reqURL := c.baseURL + EndpointNearby + "?" + params.Encode()

	return c.doRequest(ctx, reqURL)
}

// SearchLocations searches for stations by name
func (c *Client) SearchLocations(ctx context.Context, query string) ([]models.Location, error) {
	body, err := c.SearchLocationsRaw(ctx, query)
	if err != nil {
		return nil, err
	}

	var resp []models.LocationResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse locations response: %w", err)
	}

	locations := make([]models.Location, 0, len(resp))
	for _, entry := range resp {
		locations = append(locations, *entry.ToLocation())
	}

	return locations, nil
}

// SearchLocationsRaw searches for stations and returns raw JSON
func (c *Client) SearchLocationsRaw(ctx context.Context, query string) (json.RawMessage, error) {
	params := url.Values{}
	params.Set("suchbegriff", query)
	params.Set("typ", "ALL")
	params.Set("limit", "10")

	reqURL := c.baseURL + EndpointLocations + "?" + params.Encode()

	return c.doRequest(ctx, reqURL)
}

// GetJourney fetches journey details by journey ID
func (c *Client) GetJourney(ctx context.Context, journeyID string, withPolyline bool) (*models.Journey, error) {
	body, err := c.GetJourneyRaw(ctx, journeyID, withPolyline)
	if err != nil {
		return nil, err
	}

	var resp models.JourneyResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse journey response: %w", err)
	}

	return resp.ToJourney(journeyID, c.timezone), nil
}

// GetJourneyRaw fetches journey details and returns raw JSON
func (c *Client) GetJourneyRaw(ctx context.Context, journeyID string, withPolyline bool) (json.RawMessage, error) {
	params := url.Values{}
	params.Set("journeyId", journeyID)
	if withPolyline {
		params.Set("poly", "true")
	} else {
		params.Set("poly", "false")
	}

	reqURL := c.baseURL + EndpointJourney + "?" + params.Encode()

	return c.doRequest(ctx, reqURL)
}

// FormationRequest contains parameters for a formation query
type FormationRequest struct {
	EVA         int64     // Station EVA number
	TrainType   string    // Train type (e.g., "ICE")
	TrainNumber string    // Train number (e.g., "623")
	Departure   time.Time // Departure time
}

// GetFormation fetches train formation/composition data
func (c *Client) GetFormation(ctx context.Context, req FormationRequest) (*models.Formation, error) {
	body, err := c.GetFormationRaw(ctx, req)
	if err != nil {
		return nil, err
	}

	var resp models.FormationResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse formation response: %w", err)
	}

	return resp.ToFormation(req.TrainType), nil
}

// GetFormationRaw fetches train formation data and returns raw JSON
func (c *Client) GetFormationRaw(ctx context.Context, req FormationRequest) (json.RawMessage, error) {
	// Convert departure time to UTC
	departure := req.Departure
	if departure.IsZero() {
		departure = time.Now().In(c.timezone)
	}
	utcTime := departure.UTC()

	params := url.Values{}
	params.Set("administrationId", "80") // DB Fernverkehr
	params.Set("category", req.TrainType)
	params.Set("date", utcTime.Format("2006-01-02"))
	params.Set("evaNumber", fmt.Sprintf("%d", req.EVA))
	params.Set("number", req.TrainNumber)
	params.Set("time", utcTime.Format("2006-01-02T15:04:05.000Z"))

	reqURL := c.baseURL + EndpointFormation + "?" + params.Encode()

	return c.doRequest(ctx, reqURL)
}

// doRequest performs an HTTP GET request with optional caching
func (c *Client) doRequest(ctx context.Context, reqURL string) ([]byte, error) {
	// Check cache first
	if c.cache != nil {
		if data, ok := c.cache.Get(reqURL); ok {
			return data, nil
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	bp := c.browser

	// Standard browser headers in Chrome-typical order
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "de-DE,de;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Origin", "https://www.bahn.de")
	req.Header.Set("Referer", "https://www.bahn.de/buchung/fahrplan/suche")
	req.Header.Set("User-Agent", bp.userAgent)

	// Sec-Fetch headers (Chrome always sends these on XHR/fetch)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	// Client hints (Chrome sends these)
	req.Header.Set("sec-ch-ua", bp.secChUA)
	if bp.mobile {
		req.Header.Set("sec-ch-ua-mobile", "?1")
		req.Header.Set("sec-ch-ua-platform", `"Android"`)
	} else {
		req.Header.Set("sec-ch-ua-mobile", "?0")
		platforms := []string{`"Windows"`, `"macOS"`, `"ChromeOS"`}
		req.Header.Set("sec-ch-ua-platform", platforms[cryptoRandIntn(len(platforms))])
	}

	// Correlation ID per request
	req.Header.Set("x-correlation-id", uuid4()+"_"+uuid4())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Check for context errors
		if ctx.Err() != nil {
			return nil, fmt.Errorf("%w: %w", ErrTimeout, ctx.Err())
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Handle non-OK status codes with proper error types
	if resp.StatusCode != http.StatusOK {
		// Extract endpoint from URL for error message
		endpoint := extractEndpoint(reqURL)
		return nil, NewAPIError(resp.StatusCode, resp.Status, endpoint)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Store in cache
	if c.cache != nil {
		_ = c.cache.Set(reqURL, body)
	}

	return body, nil
}

// extractEndpoint extracts the endpoint path from a full URL
func extractEndpoint(fullURL string) string {
	u, err := url.Parse(fullURL)
	if err != nil {
		return fullURL
	}
	return u.Path
}
