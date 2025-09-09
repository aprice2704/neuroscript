// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Tests for special string type validation (email, url, etc.).
// filename: pkg/json-lite/special_types_test.go
// nlines: 133
// risk_rating: LOW

package json_lite

import (
	"errors"
	"testing"
)

func TestShapeValidate_Email(t *testing.T) {
	shapeDef := map[string]any{"user_email": "email"}
	s, err := ParseShape(shapeDef)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	testCases := []struct {
		name        string
		email       any
		expectedErr error
	}{
		{"valid email", "test@example.com", nil},
		{"valid email with subdomain", "test@sub.example.co.uk", nil},
		{"valid email with plus", "test+alias@example.com", nil},
		{"invalid format - no at", "test.example.com", ErrValidationFailed},
		{"invalid format - no domain", "test@", ErrValidationFailed},
		{"invalid format - no tld", "test@example", ErrValidationFailed},
		{"invalid format - whitespace", "test@ example.com", ErrValidationFailed},
		{"wrong type - int", 12345, ErrValidationTypeMismatch},
		{"wrong type - bool", true, ErrValidationTypeMismatch},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := map[string]any{"user_email": tc.email}
			err := s.Validate(data, nil)

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error type %v, but got %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("validation should have passed, but got: %v", err)
			}
		})
	}
}

func TestShapeValidate_URL(t *testing.T) {
	shapeDef := map[string]any{"website": "url"}
	s, err := ParseShape(shapeDef)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	testCases := []struct {
		name        string
		url         any
		expectedErr error
	}{
		{"valid http", "http://example.com", nil},
		{"valid https", "https://example.com", nil},
		{"valid with path", "https://example.com/path/to/resource", nil},
		{"valid with subdomain", "https://sub.domain.com/path", nil},
		{"invalid - no scheme", "example.com", ErrValidationFailed},
		{"invalid - wrong scheme", "ftp://example.com", ErrValidationFailed},
		{"invalid - just text", "not a url", ErrValidationFailed},
		{"wrong type - int", 123, ErrValidationTypeMismatch},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := map[string]any{"website": tc.url}
			err := s.Validate(data, nil)
			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("validation should have passed, got %v", err)
			}
		})
	}
}

func TestShapeValidate_ISODateTime(t *testing.T) {
	shapeDef := map[string]any{"timestamp": "isoDatetime"}
	s, err := ParseShape(shapeDef)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	testCases := []struct {
		name        string
		datetime    any
		expectedErr error
	}{
		{"valid Z timezone", "2025-01-01T12:00:00Z", nil},
		{"valid with offset", "2025-01-01T12:00:00+01:00", nil},
		{"valid with fractional seconds", "2025-01-01T12:00:00.12345Z", nil},
		{"invalid - just date", "2025-01-01", ErrValidationFailed},
		{"invalid - wrong separator", "2025-01-01 12:00:00Z", ErrValidationFailed},
		{"invalid - no timezone", "2025-01-01T12:00:00", ErrValidationFailed},
		{"wrong type - float", 123.45, ErrValidationTypeMismatch},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := map[string]any{"timestamp": tc.datetime}
			err := s.Validate(data, nil)
			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("expected error %v, got %v", tc.expectedErr, err)
				}
			} else if err != nil {
				t.Fatalf("validation should have passed, got %v", err)
			}
		})
	}
}

func TestShapeValidate_SpecialTypeNonStringError(t *testing.T) {
	shapeDef := map[string]any{"contact": "email"}
	s, err := ParseShape(shapeDef)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	data := map[string]any{"contact": 123}
	err = s.Validate(data, nil)
	if !errors.Is(err, ErrValidationTypeMismatch) {
		t.Fatalf("expected type mismatch for non-string special type, got: %v", err)
	}
}
