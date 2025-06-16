// NeuroScript Version: 0.3.8
// File version: 0.2.1
// Purpose: Corrected test to only expect tools registered by NewDefaultTestInterpreter.
// Filename: pkg/core/tools_meta_test.go
// nlines: 180
// risk_rating: LOW

package core

import (
	"fmt"
	"strings"
	"testing"
)

func TestToolMetaListTools(t *testing.T) {
	interpreter, err := NewDefaultTestInterpreter(t) // Correctly handle potential error
	if err != nil {
		t.Fatalf("NewDefaultTestInterpreter failed: %v", err)
	}

	result, err := interpreter.ExecuteTool("Meta.ListTools", map[string]Value{})
	if err != nil {
		t.Fatalf("Meta.ListTools execution failed: %v", err)
	}

	resultStr, ok := result.(StringValue)
	if !ok {
		t.Fatalf("Meta.ListTools did not return a StringValue, got %T", result)
	}
	resultOutput := resultStr.Value

	// FIX: The NewDefaultTestInterpreter only registers Core tools.
	// It does not register Go.* or AIWorker.* tools, so they should not be expected here.
	expectedSignatures := []string{
		"Meta.ListTools() -> string",
		"Meta.ToolsHelp(filter:string?) -> string",
		"FS.Read(filepath:string) -> string",
	}

	for _, sig := range expectedSignatures {
		if !strings.Contains(resultOutput, sig) {
			t.Errorf("Meta.ListTools output does not contain expected signature: %s", sig)
		}
	}
}

func TestToolMetaToolsHelp(t *testing.T) {
	interpreter, err := NewDefaultTestInterpreter(t)
	if err != nil {
		t.Fatalf("NewDefaultTestInterpreter failed: %v", err)
	}

	tests := []struct {
		name                 string
		filterArg            map[string]Value
		expectedToContain    []string
		expectedToNotContain []string
		checkNoToolsMsg      bool
		noToolsFilter        string
	}{
		{
			name:      "No filter (all tools)",
			filterArg: map[string]Value{},
			expectedToContain: []string{
				"## `tool.Meta.ListTools`",
				"## `tool.FS.Read`",
			},
			expectedToNotContain: []string{"No tools found matching filter"},
		},
		{
			name:      "Filter for Meta tools",
			filterArg: map[string]Value{"filter": StringValue{Value: "Meta."}},
			expectedToContain: []string{
				"## `tool.Meta.ListTools`",
				"## `tool.Meta.ToolsHelp`",
				"Showing tools matching filter: `Meta.`",
			},
			expectedToNotContain: []string{"## `tool.FS.Read`"},
		},
		{
			name:            "Filter with no results",
			filterArg:       map[string]Value{"filter": StringValue{Value: "NoSucchToolExistsFilter123"}},
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

			resultStr, ok := result.(StringValue)
			if !ok {
				t.Fatalf("Meta.ToolsHelp did not return a StringValue, got %T. Args: %#v", result, tt.filterArg)
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
					if resultOutput != "No tools are currently registered." {
						t.Errorf("Meta.ToolsHelp output for '%s' expected to contain '%s', got '\n%s'", tt.name, expectedMsg, resultOutput)
					}
				}
			}
		})
	}
}
