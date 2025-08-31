// NeuroScript Version: 0.7.0
// File version: 5
// Purpose: Adds a recursion depth limit to prevent stack overflow DoS attacks.
// filename: aeiou/canonical_json.go
// nlines: 91
// risk_rating: HIGH

package aeiou

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
)

const maxRecursionDepth = 100

// Canonicalize takes a struct or raw JSON and canonicalizes it according to a strict
// subset of JCS (RFC-8785) rules sufficient for AEIOU v3.
func Canonicalize(data interface{}) ([]byte, error) {
	var raw json.RawMessage
	var err error

	switch v := data.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	case json.RawMessage:
		raw = v
	default:
		raw, err = json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal data: %w", err)
		}
	}

	return transform(raw, 0)
}

// transform is the recursive heart of the canonicalizer.
func transform(raw json.RawMessage, depth int) ([]byte, error) {
	if depth > maxRecursionDepth {
		return nil, ErrMaxRecursionDepth
	}

	// Is it a JSON object?
	if bytes.HasPrefix(raw, []byte{'{'}) {
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(raw, &obj); err != nil {
			return nil, err
		}

		keys := make([]string, 0, len(obj))
		for k := range obj {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var b bytes.Buffer
		b.WriteByte('{')
		for i, k := range keys {
			b.WriteString(strconv.Quote(k))
			b.WriteByte(':')

			val, err := transform(obj[k], depth+1)
			if err != nil {
				return nil, err
			}
			b.Write(val)

			if i < len(keys)-1 {
				b.WriteByte(',')
			}
		}
		b.WriteByte('}')
		return b.Bytes(), nil
	}

	// Is it a JSON array?
	if bytes.HasPrefix(raw, []byte{'['}) {
		var arr []json.RawMessage
		if err := json.Unmarshal(raw, &arr); err != nil {
			return nil, err
		}

		var b bytes.Buffer
		b.WriteByte('[')
		for i, v := range arr {
			val, err := transform(v, depth+1)
			if err != nil {
				return nil, err
			}
			b.Write(val)
			if i < len(arr)-1 {
				b.WriteByte(',')
			}
		}
		b.WriteByte(']')
		return b.Bytes(), nil
	}

	// It's a scalar (string, number, bool, null), return as is.
	return raw, nil
}
