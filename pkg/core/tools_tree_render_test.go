// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-01 22:19:01 PDT // Remove node IDs from wantText, fix indentation
// filename: pkg/core/tools_tree_render_test.go

package core

import (
	"encoding/json"
	"errors"
	"reflect"
	"regexp" // Import regexp
	"strings"
	"testing"
)

// --- Test Helper ---

// --- Tests ---

// TestTreeFormatJSON
// --- (TestTreeFormatJSON remains unchanged) ---
func TestTreeFormatJSON(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	tests := []struct {
		name      string
		jsonInput string
	}{{"Simple Object", `{"key": "value", "num": 123, "bool": true, "nil": null}`}, {"Simple Array", `[1, "two", true, null] `}, {"Nested Structure", `{"a": [1, {"b": null}], "c": true}`}, {"Empty Object", `{}`}, {"Empty Array", `[]`}, {"Order Check Array", `[{"id":1}, {"id":0}]`}, {"Order Check Object", `{"z":1, "a":2}`}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handle := loadJSONHelper(t, interp, tt.jsonInput)
			formattedData, err := toolTreeFormatJSON(interp, MakeArgs(handle))
			if err != nil {
				t.Fatalf("toolTreeFormatJSON failed: %v", err)
			}
			formattedStr, ok := formattedData.(string)
			if !ok {
				t.Fatalf("toolTreeFormatJSON did not return string, got %T", formattedData)
			}
			var originalIntf, formattedIntf interface{}
			trimmedInput := strings.TrimSpace(tt.jsonInput)
			errOrig := json.Unmarshal([]byte(trimmedInput), &originalIntf)
			if errOrig != nil {
				t.Fatalf("Unmarshal original failed: %v", errOrig)
			}
			errFmt := json.Unmarshal([]byte(formattedStr), &formattedIntf)
			if errFmt != nil {
				t.Fatalf("Unmarshal formatted failed: %v\nString:\n%s", errFmt, formattedStr)
			}
			if !reflect.DeepEqual(originalIntf, formattedIntf) {
				t.Errorf("Formatted JSON mismatch.\nOriginal:\n%s\n\nFormatted:\n%s", trimmedInput, formattedStr)
			} else {
				t.Logf("DeepEqual check passed.")
			}
		})
	}
	t.Run("InvalidHandle_WrongPrefix", func(t *testing.T) {
		_, err := toolTreeFormatJSON(interp, MakeArgs("badprefix::123"))
		if err == nil {
			t.Error("Expected error, got nil")
		} else if !errors.Is(err, ErrCacheObjectWrongType) {
			t.Errorf("Expected ErrCacheObjectWrongType, got: %v", err)
		} else {
			t.Logf("Got expected ErrCacheObjectWrongType: %v", err)
		}
	})
	t.Run("InvalidHandle_NotFound", func(t *testing.T) {
		validLookingHandle := GenericTreeHandleType + "::not-a-real-uuid"
		_, err := toolTreeFormatJSON(interp, MakeArgs(validLookingHandle))
		if err == nil {
			t.Error("Expected error, got nil")
		} else if !errors.Is(err, ErrCacheObjectNotFound) {
			t.Errorf("Expected ErrCacheObjectNotFound, got: %v", err)
		} else {
			t.Logf("Got expected ErrCacheObjectNotFound: %v", err)
		}
	})
}

// TestTreeRenderText tests the new text rendering tool.
func TestTreeRenderText(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)

	// Regex to remove the [node-X] part for comparison
	nodeIdRegex := regexp.MustCompile(`\[node-\d+\] `)

	tests := []struct {
		name      string
		jsonInput string
		wantText  string // Expected indented text output (WITHOUT node IDs)
	}{
		{
			name:      "Simple Object",
			jsonInput: `{"key": "value", "num": 123, "active": true}`,
			// *** UPDATED WANTTEXT: No node IDs, corrected indentation ***
			wantText: `- (object) (attrs: 3)
  * Key: "active"
    - (boolean): true
  * Key: "key"
    - (string): "value"
  * Key: "num"
    - (number): 123
`,
		},
		{
			name:      "Simple Array",
			jsonInput: `[1, "two", null]`,
			// *** UPDATED WANTTEXT: No node IDs, corrected indentation ***
			wantText: `- (array) (len: 3)
  - (number): 1
  - (string): "two"
  - (null): null
`,
		},
		{
			name:      "Nested Structure",
			jsonInput: `{"d": null, "a": [true, {"b": "c"}]}`, // Keys sorted a, d
			// *** UPDATED WANTTEXT: No node IDs, corrected indentation ***
			wantText: `- (object) (attrs: 2)
  * Key: "a"
    - (array) (len: 2)
      - (boolean): true
      - (object) (attrs: 1)
        * Key: "b"
          - (string): "c"
  * Key: "d"
    - (null): null
`,
		},
		{
			name:      "Empty Object",
			jsonInput: `{}`,
			wantText: `- (object) (attrs: 0)
`,
		},
		{
			name:      "Empty Array",
			jsonInput: `[]`,
			wantText: `- (array) (len: 0)
`,
		},
		{
			name:      "Just String",
			jsonInput: `"hello"`,
			wantText: `- (string): "hello"
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handle := loadJSONHelper(t, interp, tt.jsonInput)
			renderedData, err := toolTreeRenderText(interp, MakeArgs(handle))
			if err != nil {
				t.Fatalf("toolTreeRenderText failed: %v", err)
			}
			renderedStr, ok := renderedData.(string)
			if !ok {
				t.Fatalf("toolTreeRenderText did not return string, got %T", renderedData)
			}

			// Normalize expected output
			want := strings.TrimSpace(tt.wantText) + "\n"
			// Normalize actual output AND remove node IDs
			gotRaw := strings.TrimSpace(renderedStr) + "\n"
			got := nodeIdRegex.ReplaceAllString(gotRaw, "") // Remove [node-X]

			if got != want {
				t.Errorf("Rendered text mismatch (Node IDs ignored).\n--- GOT (Raw) ---\n%s\n--- GOT (Cleaned) ---\n%s\n--- WANT (No IDs) ---\n%s", gotRaw, got, want)
			}
		})
	}

	// Test invalid handle remains unchanged (checking error type)
	t.Run("InvalidHandle_WrongPrefix", func(t *testing.T) {
		_, err := toolTreeRenderText(interp, MakeArgs("badprefix::123"))
		if err == nil {
			t.Error("Expected error, got nil")
		} else if !errors.Is(err, ErrCacheObjectWrongType) {
			t.Errorf("Expected ErrCacheObjectWrongType, got: %v", err)
		} else {
			t.Logf("Got expected ErrCacheObjectWrongType: %v", err)
		}
	})
	t.Run("InvalidHandle_NotFound", func(t *testing.T) {
		validLookingHandle := GenericTreeHandleType + "::not-a-real-uuid"
		_, err := toolTreeRenderText(interp, MakeArgs(validLookingHandle))
		if err == nil {
			t.Error("Expected error, got nil")
		} else if !errors.Is(err, ErrCacheObjectNotFound) {
			t.Errorf("Expected ErrCacheObjectNotFound, got: %v", err)
		} else {
			t.Logf("Got expected ErrCacheObjectNotFound: %v", err)
		}
	})
}
