// NeuroScript Version: 0.7.0
// File version: 7
// Purpose: Fixes a variable redeclaration compiler error in the 'magic tool' test case.
// filename: pkg/tool/aeiou_proto/aeiou_tool_test.go
// nlines: 132
// risk_rating: LOW

package aeiou_proto

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func TestAeiouV2Tools(t *testing.T) {
	interp := interpreter.NewInterpreter()

	// Helper to run a tool and get the result
	runTool := func(name string, args ...interface{}) (interface{}, error) {
		fullname := types.MakeFullName(group, name)
		toolImpl, found := interp.ToolRegistry().GetTool(fullname)
		if !found {
			t.Fatalf("tool %s not found", fullname)
		}
		return toolImpl.Func(interp, args)
	}

	// --- Test .magic ---
	t.Run("magic tool", func(t *testing.T) {
		// Test with payload
		payload := map[string]interface{}{"control": "continue"}
		magicVal, err := runTool("magic", "LOOP", payload)
		if err != nil {
			t.Fatalf("aeiou.magic failed: %v", err)
		}
		magicStr, ok := magicVal.(string)
		if !ok {
			t.Fatalf("aeiou.magic did not return a string, got %T", magicVal)
		}
		expected, _ := aeiou.Wrap(aeiou.SectionLoop, payload)
		if magicStr != expected {
			t.Errorf("magic string mismatch:\ngot:  %s\nwant: %s", magicStr, expected)
		}

		// Test without payload
		magicVal, err = runTool("magic", "START") // THE FIX IS HERE
		if err != nil {
			t.Fatalf("aeiou.magic failed: %v", err)
		}
		magicStr, ok = magicVal.(string)
		if !ok {
			t.Fatalf("aeiou.magic did not return a string, got %T", magicVal)
		}
		expected, _ = aeiou.Wrap(aeiou.SectionStart, nil)
		if magicStr != expected {
			t.Errorf("magic string mismatch:\ngot:  %s\nwant: %s", magicStr, expected)
		}
	})

	// --- Test Full V2 Lifecycle ---
	t.Run("V2 lifecycle", func(t *testing.T) {
		// 1. Create a new envelope
		handleVal, err := runTool("new")
		if err != nil {
			t.Fatalf("aeiou.new failed: %v", err)
		}
		handle, _ := handleVal.(string)

		// 2. Set some content
		_, err = runTool("set_section", handle, "ACTIONS", "command {}")
		if err != nil {
			t.Fatalf("aeiou.set_section failed: %v", err)
		}

		// 3. Compose it to a V2 string
		payloadVal, err := runTool("compose", handle)
		if err != nil {
			t.Fatalf("aeiou.compose failed: %v", err)
		}
		payload, _ := payloadVal.(string)

		// Verify V2 markers
		if !strings.Contains(payload, "V2:START") || !strings.Contains(payload, "V2:END") {
			t.Fatalf("composed payload is missing V2 START/END markers:\n%s", payload)
		}

		// 4. Parse it back
		parsedHandleVal, err := runTool("parse", payload)
		if err != nil {
			t.Fatalf("aeiou.parse failed on V2 payload: %v", err)
		}
		parsedHandle, _ := parsedHandleVal.(string)

		// 5. Get the content back and verify
		contentVal, err := runTool("get_section", parsedHandle, "ACTIONS")
		if err != nil {
			t.Fatalf("get_section on parsed handle failed: %v", err)
		}
		if !reflect.DeepEqual(contentVal, "command {}") {
			t.Errorf("round trip content mismatch: got %q, want %q", contentVal, "command {}")
		}
	})

	// --- Test Error Cases ---
	t.Run("error cases", func(t *testing.T) {
		// Test parse with invalid payload
		_, err := runTool("parse", "this is not an envelope")
		if !errors.Is(err, aeiou.ErrEnvelopeNoStart) {
			t.Errorf("expected ErrEnvelopeNoStart on invalid parse, got %v", err)
		}

		// Create a real handle to derive a well-formatted but non-existent one.
		realHandleVal, _ := runTool("new")
		realHandle := realHandleVal.(string)
		parts := strings.Split(realHandle, "::")
		if len(parts) != 2 {
			t.Fatalf("unexpected handle format: %s", realHandle)
		}
		prefix := parts[0]
		nonExistentHandle := fmt.Sprintf("%s::%s", prefix, "00000000-0000-0000-0000-000000000000")

		_, err = runTool("get_section", nonExistentHandle, "ACTIONS")
		if !errors.Is(err, lang.ErrHandleNotFound) {
			t.Errorf("expected ErrHandleNotFound for invalid handle, got %v", err)
		}
	})
}
