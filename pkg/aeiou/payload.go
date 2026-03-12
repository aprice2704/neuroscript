// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 1
// :: description: Centralized LLM payload parsing and sanitization. Multi-tier Waterfall parser with scrubbing JSON parser.
// :: latestChange: Initial backport from FDM aeiou service to NS native aeiou pkg.
// :: filename: pkg/aeiou/payload.go
// :: serialization: go

package aeiou

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
)

// ParsePayloadWaterfall implements a multi-tier sanitization logic for LLM outputs.
// It attempts to recover valid JSON from mangled text or Go slice stringification artifacts.
// It returns a map suitable for NeuroScript consumption:
// {"success": bool, "data": list, "required_recovery": bool, "warning": string, "err": error}
func ParsePayloadWaterfall(raw string) map[string]any {
	rawTrimmed := strings.TrimSpace(raw)

	// Tier 1: Pure Strict Parse (The Happy Path)
	// Must be valid JSON from start to finish with zero noise.
	data, err := parseJSONStrict(rawTrimmed)
	if err == nil && len(data) > 0 {
		return map[string]any{
			"success":           true,
			"data":              data,
			"required_recovery": false,
			"warning":           "",
			"err":               nil,
		}
	}

	// Tier 2: Cleaned Strict Parse
	// Strips markdown fences, labels, and artifacts, then attempts a strict parse.
	cleaned := CleanLLMPayload(rawTrimmed)

	// Check for Go slice artifact coercion specifically: "[{payload} marker]"
	if strings.HasPrefix(cleaned, "[") && strings.HasSuffix(cleaned, "]") {
		inner := strings.TrimSpace(cleaned[1 : len(cleaned)-1])
		inner = strings.ReplaceAll(inner, "<<<LOOP:DONE>>>", "")
		inner = strings.TrimSpace(inner)

		if innerData, innerErr := parseJSONStrict(inner); innerErr == nil && len(innerData) > 0 {
			return map[string]any{
				"success":           true,
				"data":              innerData,
				"required_recovery": true,
				"warning":           "Payload required bracket unwrapping.",
				"err":               nil,
			}
		}
	}

	// If cleaning actually changed the string and results in valid JSON, we mark recovery as true.
	if cleaned != rawTrimmed {
		if cleanedData, cleanedErr := parseJSONStrict(cleaned); cleanedErr == nil && len(cleanedData) > 0 {
			return map[string]any{
				"success":           true,
				"data":              cleanedData,
				"required_recovery": true,
				"warning":           "",
				"err":               nil,
			}
		}
	}

	// Tier 3: Structural Scrubbing (The Salvage Operation)
	// Skips noise and conversational filler to find valid structural blocks.
	scrubData, scrubErr := parseJSONScrubbed(rawTrimmed)
	if len(scrubData) > 0 {
		return map[string]any{
			"success":           true,
			"data":              scrubData,
			"required_recovery": true,
			"warning":           "Payload required structural extraction.",
			"err":               scrubErr,
		}
	}

	// Tier 4: Bailout
	return map[string]any{
		"success":           false,
		"data":              data,
		"required_recovery": false,
		"warning":           "Failed to extract valid JSON payload from the response.",
		"err":               err,
	}
}

// parseJSONStrict parses a string expecting one or more JSON objects with zero noise.
func parseJSONStrict(raw string) ([]any, error) {
	if raw == "" {
		return nil, errors.New("empty payload")
	}

	dec := json.NewDecoder(strings.NewReader(raw))
	var results []any
	for {
		var v any
		if err := dec.Decode(&v); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return results, err
		}
		results = append(results, v)
	}

	// Final validation: check if the decoder consumed everything but whitespace.
	var sink any
	if err := dec.Decode(&sink); err != nil && !errors.Is(err, io.EOF) {
		// If there is trailing data that isn't valid JSON (like " - Thanks!"),
		// Decode will return a syntax error, which means the string isn't "Strictly JSON".
		return results, err
	}

	return results, nil
}

// parseJSONScrubbed searches for potential JSON starts ('{' or '[') and attempts to decode from there.
func parseJSONScrubbed(raw string) ([]any, error) {
	if raw == "" {
		return nil, errors.New("empty payload")
	}

	var results []any
	var lastErr error

	// Iterate through the string looking for potential start characters.
	for i := 0; i < len(raw); i++ {
		char := raw[i]
		if char == '{' || char == '[' {
			// Construct a reader for the remaining part of the string.
			chunk := raw[i:]
			dec := json.NewDecoder(strings.NewReader(chunk))
			var v any
			err := dec.Decode(&v)

			if err == nil {
				results = append(results, v)
				// Success! Advance i based on how much the decoder consumed.
				consumed := int(dec.InputOffset())
				i += (consumed - 1)
				lastErr = nil // Found valid data, so ignore errors from previous noise.
			} else {
				// Record error (in case it's a real trailing error) and keep searching.
				lastErr = err
			}
		}
	}

	if len(results) == 0 {
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, errors.New("no JSON structures found in text")
	}

	return results, lastErr
}

// CleanLLMPayload strips common LLM hallucinated wrappers and control markers.
func CleanLLMPayload(raw string) string {
	cleaned := strings.TrimSpace(raw)

	// 1. Remove the loop marker wherever it appears.
	cleaned = strings.ReplaceAll(cleaned, "<<<LOOP:DONE>>>", "")
	cleaned = strings.TrimSpace(cleaned)

	// 2. Normalize literal '\n' hallucinations.
	cleaned = strings.ReplaceAll(cleaned, "\\n", "\n")

	// 3. Strip NeuroScript line-continuations (\)
	lines := strings.Split(cleaned, "\n")
	for i, l := range lines {
		trimmedLine := strings.TrimRight(l, " \t\r")
		if strings.HasSuffix(trimmedLine, "\\") {
			lines[i] = strings.TrimSuffix(trimmedLine, "\\")
		}
	}
	cleaned = strings.Join(lines, "\n")

	// 4. Strip Conversational Labels.
	prefixes := []string{
		"result:", "answer:", "output:", "final result:", "final answer:", "response:", "result is:", "answer is:",
		"here is the output:", "here is the result:", "here is the json output:", "here is the json result:",
	}

	for {
		found := false
		lowerRes := strings.ToLower(cleaned)
		for _, p := range prefixes {
			if strings.HasPrefix(lowerRes, p) {
				cleaned = cleaned[len(p):]
				cleaned = strings.TrimSpace(cleaned)
				found = true
				break
			}
		}
		if !found {
			break
		}
	}

	// 5. Strip Markdown Code Fences.
	if strings.HasPrefix(cleaned, "```") {
		firstNL := strings.Index(cleaned, "\n")
		if firstNL != -1 && firstNL < 40 {
			cleaned = cleaned[firstNL+1:]
		} else {
			cleaned = strings.TrimPrefix(cleaned, "```")
			tags := []string{"json", "neuroscript", "ns", "text", "markdown", "md"}
			lowCleaned := strings.ToLower(strings.TrimSpace(cleaned))
			for _, tag := range tags {
				if strings.HasPrefix(lowCleaned, tag) {
					cleaned = strings.TrimSpace(cleaned)[len(tag):]
					break
				}
			}
		}

		cleaned = strings.TrimSpace(cleaned)
		if strings.HasSuffix(cleaned, "```") {
			cleaned = strings.TrimSuffix(cleaned, "```")
		}
	}

	return strings.TrimSpace(cleaned)
}

// ParseJSONL is the high-level compatibility function for tools.
func ParseJSONL(raw string) ([]any, error) {
	res := ParsePayloadWaterfall(raw)
	data, _ := res["data"].([]any)

	if res["success"].(bool) {
		// If success is true, check for partial errors (e.g. malformed JSONL stream)
		if err, ok := res["err"].(error); ok && err != nil {
			return data, err
		}
		return data, nil
	}

	// Failure path
	if err, ok := res["err"].(error); ok && err != nil {
		return data, err
	}
	return data, errors.New(res["warning"].(string))
}
