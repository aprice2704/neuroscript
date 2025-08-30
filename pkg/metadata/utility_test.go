// NeuroScript Version: 0.3.0
// File version: 2
// Purpose: Contains tests for the metadata utility and extractor.
// filename: pkg/metadata/utility_test.go
// nlines: 118
// risk_rating: LOW
package metadata_test

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/metadata"
)

func TestExtractor(t *testing.T) {
	store := metadata.Store{
		"schema":       "spec",
		"file-Version": "123",
		"description":  "A test file.",
		"camelCaseKey": "value",
		"not_an_int":   "abc",
		"optional.key": "present",
	}
	extractor := metadata.NewExtractor(store)

	t.Run("GetFound", func(t *testing.T) {
		// Test different variations of the same key
		keysToTest := []string{"file-Version", "file_version", "file.version", "FileVersion"}
		for _, key := range keysToTest {
			val, ok := extractor.Get(key)
			if !ok {
				t.Errorf("Get(%q) expected to find key", key)
			}
			if val != "123" {
				t.Errorf("Get(%q) got %q, want %q", key, val, "123")
			}
		}

		// Test camelCase
		val, ok := extractor.Get("Camel.Case_Key")
		if !ok || val != "value" {
			t.Errorf("Get(Camel.Case_Key) failed, got: %q, %v", val, ok)
		}
	})

	t.Run("GetNotFound", func(t *testing.T) {
		_, ok := extractor.Get("nonexistent")
		if ok {
			t.Error("Get(\"nonexistent\") expected not to find key")
		}
	})

	t.Run("GetWithDefaults", func(t *testing.T) {
		// Test GetOr
		if val := extractor.GetOr("schema", "default"); val != "spec" {
			t.Errorf("GetOr on existing key failed, got %q", val)
		}
		if val := extractor.GetOr("nonexistent", "default"); val != "default" {
			t.Errorf("GetOr on missing key failed, got %q", val)
		}

		// Test GetIntOr
		i, err := extractor.GetIntOr("file.version", 999)
		if err != nil || i != 123 {
			t.Errorf("GetIntOr on existing key failed, got %d, %v", i, err)
		}
		i, err = extractor.GetIntOr("nonexistent", 999)
		if err != nil || i != 999 {
			t.Errorf("GetIntOr on missing key failed, got %d, %v", i, err)
		}
		_, err = extractor.GetIntOr("not_an_int", 999)
		if err == nil {
			t.Error("GetIntOr on non-int value expected an error")
		}
	})

	t.Run("MustGet", func(t *testing.T) {
		if val := extractor.MustGet("schema"); val != "spec" {
			t.Errorf("MustGet(\"schema\") got %q, want %q", val, "spec")
		}
		if val := extractor.MustGet("nonexistent"); val != "" {
			t.Errorf("MustGet(\"nonexistent\") got %q, want \"\"", val)
		}
	})

	t.Run("GetInt", func(t *testing.T) {
		i, ok, err := extractor.GetInt("file.version")
		if !ok || err != nil || i != 123 {
			t.Errorf("GetInt(\"file.version\") got %d, %v, %v; want 123, true, nil", i, ok, err)
		}

		_, ok, _ = extractor.GetInt("nonexistent")
		if ok {
			t.Error("GetInt(\"nonexistent\") expected ok=false")
		}

		_, _, err = extractor.GetInt("not_an_int")
		if err == nil {
			t.Error("GetInt(\"not_an_int\") expected a parse error")
		}
	})

	t.Run("CheckRequired", func(t *testing.T) {
		// Test with keys that will be normalized
		err := extractor.CheckRequired("schema", "file-version")
		if err != nil {
			t.Errorf("CheckRequired() returned unexpected error: %v", err)
		}

		err = extractor.CheckRequired("schema", "missing-key")
		if err == nil {
			t.Error("CheckRequired() expected an error for missing key")
		}
	})
}
