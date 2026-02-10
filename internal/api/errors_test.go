package api

import (
	"errors"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *APIError
		wantStr string
	}{
		{
			name: "with message",
			err: &APIError{
				StatusCode: 404,
				Endpoint:   "/test",
				Message:    "Resource not found",
			},
			wantStr: "API error 404 (/test): Resource not found",
		},
		{
			name: "without message",
			err: &APIError{
				StatusCode: 500,
				Status:     "Internal Server Error",
				Endpoint:   "/api/test",
			},
			wantStr: "API error 500: Internal Server Error (endpoint: /api/test)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantStr {
				t.Errorf("Error() = %q, want %q", got, tt.wantStr)
			}
		})
	}
}

func TestAPIError_Is(t *testing.T) {
	tests := []struct {
		name      string
		err       *APIError
		target    error
		wantMatch bool
	}{
		{
			name:      "404 matches ErrNotFound",
			err:       &APIError{StatusCode: 404},
			target:    ErrNotFound,
			wantMatch: true,
		},
		{
			name:      "500 matches ErrServerError",
			err:       &APIError{StatusCode: 500},
			target:    ErrServerError,
			wantMatch: true,
		},
		{
			name:      "502 matches ErrServerError",
			err:       &APIError{StatusCode: 502},
			target:    ErrServerError,
			wantMatch: true,
		},
		{
			name:      "400 matches ErrInvalidRequest",
			err:       &APIError{StatusCode: 400},
			target:    ErrInvalidRequest,
			wantMatch: true,
		},
		{
			name:      "404 does not match ErrServerError",
			err:       &APIError{StatusCode: 404},
			target:    ErrServerError,
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errors.Is(tt.err, tt.target); got != tt.wantMatch {
				t.Errorf("Is() = %v, want %v", got, tt.wantMatch)
			}
		})
	}
}

func TestNewAPIError(t *testing.T) {
	err := NewAPIError(404, "Not Found", "/test/endpoint")

	if err.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", err.StatusCode)
	}
	if err.Status != "Not Found" {
		t.Errorf("Status = %q, want %q", err.Status, "Not Found")
	}
	if err.Endpoint != "/test/endpoint" {
		t.Errorf("Endpoint = %q, want %q", err.Endpoint, "/test/endpoint")
	}
}

func TestValidationError(t *testing.T) {
	err := NewValidationError("station", "field is required")

	if err.Field != "station" {
		t.Errorf("Field = %q, want %q", err.Field, "station")
	}
	if err.Message != "field is required" {
		t.Errorf("Message = %q, want %q", err.Message, "field is required")
	}

	expectedStr := "validation error: station - field is required"
	if err.Error() != expectedStr {
		t.Errorf("Error() = %q, want %q", err.Error(), expectedStr)
	}
}

func TestErrMissingField(t *testing.T) {
	err := ErrMissingField("eva")

	ve := &ValidationError{}
	ok := errors.As(err, &ve)
	if !ok {
		t.Fatal("Expected *ValidationError")
	}
	if ve.Field != "eva" {
		t.Errorf("Field = %q, want %q", ve.Field, "eva")
	}
}

func TestErrInvalidFormat(t *testing.T) {
	err := ErrInvalidFormat("date", "YYYY-MM-DD")

	ve := &ValidationError{}
	ok := errors.As(err, &ve)
	if !ok {
		t.Fatal("Expected *ValidationError")
	}
	if ve.Field != "date" {
		t.Errorf("Field = %q, want %q", ve.Field, "date")
	}
}
