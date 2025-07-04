// NeuroScript Version: 0.3.8
// File version: 0.4.0
// Purpose: Removed duplicate test helpers and corrected struct literal syntax.
// nlines: 180
// risk_rating: LOW

// filename: pkg/tool/meta/tools_meta_test.go
package meta

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestToolMetaListTools(t *testing.T) {
	// FIX: Use the single, correct helper from the suite file.
	interpreter, err := newMetaTestInterpreter(t)
	if err != nil {
		t.Fatalf("newMetaTestInterpreter failed: %v", err)
	}

	result, err := interpreter.ExecuteTool("Meta.ListTools", nil)
	if err != nil {
		t.Fatalf("Meta.ListTools execution failed: %v", err)
	}

	resultStr, ok := result.(lang.StringValue)
	if !ok {
		t.Fatalf("Meta.ListTools did not return a lang.StringValue, got %T", result)
	}
	resultOutput := resultStr.Value

	expectedSignatures := []string{
		"Meta.ListTools() -> string",
		"Meta.ToolsHelp(filter:string?) -> string",
		"FS.Read() -> any",
	}

	for _, sig := range expectedSignatures {
		if !strings.Contains(resultOutput, sig) {
			t.Errorf("Meta.ListTools output does not contain expected signature: %s", sig)
		}
	}
}

func TestToolMetaToolsHelp(t *testing.T) {
	interpreter, err := newMetaTestInterpreter(t)
	if err != nil {
		t.Fatalf("newMetaTestInterpreter failed: %v", err)
	}

	tests := []struct {
		name                 string
		filterArg            map[string]lang.Value
		expectedToContain    []string
		expectedToNotContain []string
		checkNoToolsMsg      bool
		noToolsFilter        string
	}{
		{
			name:      "No filter (all tools)",
			filterArg: map[string]lang.Value{},
			expectedToContain: []string{
				"## `tool.Meta.ListTools`",
				"## `tool.FS.Read`",
			},
			expectedToNotContain: []string{"No tools found matching filter"},
		},
		{
			name:      "Filter for Meta tools",
			filterArg: map[string]lang.Value{"filter": lang.StringValue{Value: "Meta."}},
			expectedToContain: []string{
				"## `tool.Meta.ListTools`",
				"## `tool.Meta.ToolsHelp`",
			},
			expectedToNotContain: []string{"## `tool.FS.Read`"},
		},
		{
			name: "Filter with no results",
			// FIX: Corrected the struct literal syntax.
			filterArg:       map[string]lang.Value{"filter": lang.StringValue{Value: "NoSucchToolExistsFilter123"}},
			checkNoToolsMsg: true,
			noToolsFilter:   "NoSucchToolExistsFilter123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := interpreter.ExecuteTool("Meta.ToolsHelp", tt.filterArg)
			if err != nil {
				t.Fatalf("Meta.ToolsHelp execution failed: %v. Args: %#v", err, tt.filterArg)
			}

			resultStr, ok := result.(lang.StringValue)
			if !ok {
				t.Fatalf("Meta.ToolsHelp did not return a lang.StringValue, got %T. Args: %#v", result, tt.filterArg)
			}
			resultOutput := resultStr.Value

			for _, sub := range tt.expectedToContain {
				if !strings.Contains(resultOutput, sub) {
					t.Errorf("Meta.ToolsHelp output for '%s' does not contain expected substring: '%s'\nOutput was:\n%s", tt.name, sub, resultOutput)
				}
			}
			for _, sub := range tt.expectedToNotContain {
				if strings.Contains(resultOutput, sub) {
					t.Errorf("Meta.ToolsHelp output for '%s' unexpectedly contains substring: '%s'\nOutput was:\n%s", tt.name, sub, resultOutput)
				}
			}
			if tt.checkNoToolsMsg {
				expectedMsg := fmt.Sprintf("No tools found matching filter: `%s`", tt.noToolsFilter)
				if !strings.Contains(resultOutput, expectedMsg) {
					t.Errorf("Meta.ToolsHelp output for '%s' expected to contain '%s', got '\n%s'", tt.name, expectedMsg, resultOutput)
				}
			}
		})
	}
}
