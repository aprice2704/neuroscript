// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Provides helpers for request interpolation and response parsing for the http.Provider.
// filename: pkg/provider/httpprovider/helpers.go
// nlines: 150
// risk_rating: HIGH

package httpprovider

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jmespath/go-jmespath"
)

// buildRequestHeaders interpolates token values into a map of headers.
func buildRequestHeaders(headerTemplate map[string]string, ctx map[string]string) map[string]string {
	headers := make(map[string]string, len(headerTemplate))
	for k, v := range headerTemplate {
		headers[k] = interpolateString(v, ctx)
	}
	return headers
}

// buildRequestBody interpolates token values into the body template and marshals it to JSON.
func buildRequestBody(bodyTemplate any, ctx map[string]string) ([]byte, error) {
	interpolatedBody, err := interpolateRecursive(bodyTemplate, ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInterpolation, err)
	}
	return json.Marshal(interpolatedBody)
}

// interpolateString replaces all known tokens in a single string.
// e.g., "Bearer {APIKEY}" -> "Bearer sk-123..."
func interpolateString(s string, ctx map[string]string) string {
	for token, value := range ctx {
		placeholder := fmt.Sprintf("{%s}", token)
		// Use strings.Replace instead of template engines for simplicity
		// and to avoid issues with JSON string literal escaping.
		s = strings.ReplaceAll(s, placeholder, value)
	}
	return s
}

// interpolateRecursive walks a nested structure (map/slice) and interpolates all string values.
func interpolateRecursive(data any, ctx map[string]string) (any, error) {
	switch v := data.(type) {
	case string:
		// Check for a special case: if the string *is* the prompt token,
		// we must respect the prompt's type (e.g., it might be a JSON string).
		// We only JSON-escape the prompt if it's part of a larger string.
		if v == "{PROMPT}" {
			return ctx["PROMPT"], nil
		}
		// Otherwise, just do simple string replacement.
		return interpolateString(v, ctx), nil

	case map[string]any:
		newMap := make(map[string]any, len(v))
		for key, val := range v {
			interpolatedVal, err := interpolateRecursive(val, ctx)
			if err != nil {
				return nil, err
			}
			newMap[key] = interpolatedVal
		}
		return newMap, nil

	case []any:
		newSlice := make([]any, len(v))
		for i, val := range v {
			interpolatedVal, err := interpolateRecursive(val, ctx)
			if err != nil {
				return nil, err
			}
			newSlice[i] = interpolatedVal
		}
		return newSlice, nil

	default:
		// Pass through other types (bool, number, nil) unchanged.
		return data, nil
	}
}

// extractResponseText uses a JMESPath expression to find and return a string
// from a parsed JSON response (represented as any).
func extractResponseText(data any, path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("%w: response path is empty", ErrConfigInvalid)
	}

	result, err := jmespath.Search(path, data)
	if err != nil {
		return "", fmt.Errorf("%w: jmespath search failed for path '%s': %v", ErrResponseFormat, path, err)
	}

	strResult, ok := result.(string)
	if !ok {
		// This can happen if the path is valid but points to a non-string (e.g., an object or list).
		// Try to stringify it as a fallback.
		if result != nil {
			return fmt.Sprintf("%v", result), nil
		}
		return "", fmt.Errorf("%w: path '%s' did not return a string or any value", ErrResponseFormat, path)
	}

	return strResult, nil
}

// parseErrorResponse attempts to find a structured error message in a non-200 response.
func parseErrorResponse(body []byte, path string) string {
	if path == "" {
		return "" // No error path specified
	}

	var data any
	if err := json.Unmarshal(body, &data); err != nil {
		return "" // Body wasn't valid JSON
	}

	result, err := jmespath.Search(path, data)
	if err != nil {
		return "" // Path search failed
	}

	strResult, ok := result.(string)
	if !ok {
		return "" // Path didn't return a string
	}

	return strResult
}
