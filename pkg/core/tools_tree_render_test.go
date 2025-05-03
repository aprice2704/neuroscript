// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 20:29:00 PM PDT // Update error expectations after interpreter fix
// filename: pkg/core/tools_tree_render_test.go

package core

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp" // Using cmp for better diffs
)

// --- Tests ---

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
			// Use cmp.Diff for better reporting
			if diff := cmp.Diff(originalIntf, formattedIntf); diff != "" {
				t.Errorf("Formatted JSON mismatch (-want +got):\n%s", diff)
			} else {
				t.Logf("DeepEqual check passed.")
			}
		})
	}
	t.Run("InvalidHandle_WrongPrefix", func(t *testing.T) {
		_, err := toolTreeFormatJSON(interp, MakeArgs("badprefix::123"))
		if err == nil {
			t.Error("Expected error, got nil")
			// UPDATED Error Expectation
		} else if !errors.Is(err, ErrHandleWrongType) {
			t.Errorf("Expected ErrHandleWrongType, got: %v (Type: %T)", err, err)
		} else {
			t.Logf("Got expected ErrHandleWrongType: %v", err)
		}
	})
	t.Run("InvalidHandle_NotFound", func(t *testing.T) {
		validLookingHandle := GenericTreeHandleType + "::not-a-real-uuid"
		_, err := toolTreeFormatJSON(interp, MakeArgs(validLookingHandle))
		if err == nil {
			t.Error("Expected error, got nil")
			// UPDATED Error Expectation
		} else if !errors.Is(err, ErrNotFound) {
			t.Errorf("Expected ErrNotFound, got: %v (Type: %T)", err, err)
		} else {
			t.Logf("Got expected ErrNotFound: %v", err)
		}
	})
	t.Run("InvalidHandle_Malformed", func(t *testing.T) {
		_, err := toolTreeFormatJSON(interp, MakeArgs("badhandle"))
		if err == nil {
			t.Error("Expected error, got nil")
			// UPDATED Error Expectation
		} else if !errors.Is(err, ErrInvalidArgument) {
			t.Errorf("Expected ErrInvalidArgument, got: %v (Type: %T)", err, err)
		} else {
			t.Logf("Got expected ErrInvalidArgument: %v", err)
		}
	})
}

// TestTreeRenderText tests the new text rendering tool.
func TestTreeRenderText(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	nodeIdRegex := regexp.MustCompile(`\[node-\d+\] `)

	tests := []struct {
		name      string
		jsonInput string
		wantText  string
	}{
		{
			name:      "Simple Object",
			jsonInput: `{"key": "value", "num": 123, "active": true}`,
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
			wantText: `- (array) (len: 3)
  - (number): 1
  - (string): "two"
  - (null): null
`,
		},
		{
			name:      "Nested Structure",
			jsonInput: `{"d": null, "a": [true, {"b": "c"}]}`,
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
		{name: "Empty Object", jsonInput: `{}`, wantText: "- (object) (attrs: 0)\n"},
		{name: "Empty Array", jsonInput: `[]`, wantText: "- (array) (len: 0)\n"},
		{name: "Just String", jsonInput: `"hello"`, wantText: "- (string): \"hello\"\n"},
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
			want := strings.TrimSpace(tt.wantText) + "\n"
			gotRaw := strings.TrimSpace(renderedStr) + "\n"
			got := nodeIdRegex.ReplaceAllString(gotRaw, "")
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("Rendered text mismatch (-want +got):\n%s", diff)
			}
		})
	}

	t.Run("InvalidHandle_WrongPrefix", func(t *testing.T) {
		_, err := toolTreeRenderText(interp, MakeArgs("badprefix::123"))
		if err == nil {
			t.Error("Expected error, got nil")
			// UPDATED Error Expectation
		} else if !errors.Is(err, ErrHandleWrongType) {
			t.Errorf("Expected ErrHandleWrongType, got: %v (Type: %T)", err, err)
		} else {
			t.Logf("Got expected ErrHandleWrongType: %v", err)
		}
	})
	t.Run("InvalidHandle_NotFound", func(t *testing.T) {
		validLookingHandle := GenericTreeHandleType + "::not-a-real-uuid"
		_, err := toolTreeRenderText(interp, MakeArgs(validLookingHandle))
		if err == nil {
			t.Error("Expected error, got nil")
			// UPDATED Error Expectation
		} else if !errors.Is(err, ErrNotFound) {
			t.Errorf("Expected ErrNotFound, got: %v (Type: %T)", err, err)
		} else {
			t.Logf("Got expected ErrNotFound: %v", err)
		}
	})
	t.Run("InvalidHandle_Malformed", func(t *testing.T) {
		_, err := toolTreeRenderText(interp, MakeArgs("badhandle"))
		if err == nil {
			t.Error("Expected error, got nil")
			// UPDATED Error Expectation
		} else if !errors.Is(err, ErrInvalidArgument) {
			t.Errorf("Expected ErrInvalidArgument, got: %v (Type: %T)", err, err)
		} else {
			t.Logf("Got expected ErrInvalidArgument: %v", err)
		}
	})
}
