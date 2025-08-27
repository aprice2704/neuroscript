// NeuroScript Version: 0.7.0
// File version: 3
// Purpose: Replaces obsolete V1 tests with a single, focused test for V2 robust parsing via the tool interface.
// filename: pkg/tool/aeiou_proto/parser_robust_test.go
// nlines: 48
// risk_rating: LOW

package aeiou_proto

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func TestToolRobustParseV2(t *testing.T) {
	interp := interpreter.NewInterpreter()

	runTool := func(name string, args ...interface{}) (interface{}, error) {
		fullname := types.MakeFullName(group, name)
		toolImpl, found := interp.ToolRegistry().GetTool(fullname)
		if !found {
			t.Fatalf("tool %s not found", fullname)
		}
		return toolImpl.Func(interp, args)
	}

	startMarker, _ := aeiou.Wrap(aeiou.SectionStart, nil)
	endMarker, _ := aeiou.Wrap(aeiou.SectionEnd, nil)
	actionsMarker, _ := aeiou.Wrap(aeiou.SectionActions, nil)

	// V2 payload with a preamble
	noisyPayload := "Here is the envelope:\n" +
		startMarker + "\n" +
		actionsMarker + "\n" +
		"action content" + "\n" +
		endMarker

	handleVal, err := runTool("parse", noisyPayload)
	if err != nil {
		t.Fatalf("tool.aeiou.parse failed on noisy V2 payload: %v", err)
	}
	handle, _ := handleVal.(string)

	contentVal, err := runTool("get_section", handle, "ACTIONS")
	if err != nil {
		t.Fatalf("get_section failed: %v", err)
	}

	if strings.TrimSpace(contentVal.(string)) != "action content" {
		t.Errorf("parsed content mismatch: got %q, want %q", contentVal, "action content")
	}
}
