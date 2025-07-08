// NeuroScript Version: 0.3.8
// File version: 0.5.1
// Purpose: Combined suite and main test files, fixing helper visibility and registration calls. Corrected tool execution in tests.
// nlines: 120
// risk_rating: LOW
// filename: pkg/tool/meta/tools_meta_test.go
package meta

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// newMetaTestInterpreter sets up an interpreter instance specifically for meta tool testing.
// It registers the meta tools so they can be executed.
func newMetaTestInterpreter(t *testing.T) (*interpreter.Interpreter, error) {
	t.Helper()

	// Use WithLogger for verbose test output if needed
	interp := interpreter.NewInterpreter(interpreter.WithLogger(logging.NewTestLogger(t)))

	// Manually register the meta tools for this test suite.
	for _, toolImpl := range metaToolsToRegister {
		if err := interp.ToolRegistry().RegisterTool(toolImpl); err != nil {
			return nil, fmt.Errorf("failed to register tool '%s': %w", toolImpl.Spec.Name, err)
		}
	}

	// Register a few other dummy tools to test filtering.
	dummySpec := tool.ToolSpec{Name: "FS.Read", Description: "Dummy FS tool.", ReturnType: tool.ArgTypeAny}
	dummyFunc := func(rt tool.Runtime, args []interface{}) (interface{}, error) { return "dummy fs read", nil }
	if err := interp.ToolRegistry().RegisterTool(tool.ToolImplementation{Spec: dummySpec, Func: dummyFunc}); err != nil {
		return nil, fmt.Errorf("failed to register dummy tool: %w", err)
	}

	return interp, nil
}

func TestToolMetaListTools(t *testing.T) {
	interpreter, err := newMetaTestInterpreter(t)
	if err != nil {
		t.Fatalf("newMetaTestInterpreter failed: %v", err)
	}

	// Get the tool implementation from the registry
	listToolsImpl, found := interpreter.ToolRegistry().GetToolShort(group, "ListTools")
	if !found {
		t.Fatal("ListTools tool not found")
	}

	// Execute the tool's function directly
	result, err := listToolsImpl.Func(interpreter, []interface{}{})
	if err != nil {
		t.Fatalf("ListTools execution failed: %v", err)
	}

	resultStr, ok := result.(string)
	if !ok {
		t.Fatalf("ListTools did not return a string, got %T", result)
	}

	// We need to check for the dummy tool as well now.
	expectedSignatures := []string{
		"ListTools() -> string",
		"ToolsHelp(filter:string?) -> string",
		"FS.Read() -> any",
	}

	for _, sig := range expectedSignatures {
		if !strings.Contains(resultStr, sig) {
			t.Errorf("ListTools output does not contain expected signature: %s\nOutput was:\n%s", sig, resultStr)
		}
	}
}

func TestToolMetaToolsHelp(t *testing.T) {
	interpreter, err := newMetaTestInterpreter(t)
	if err != nil {
		t.Fatalf("newMetaTestInterpreter failed: %v", err)
	}

	// Get the tool implementation from the registry
	toolsHelpImpl, found := interpreter.ToolRegistry().GetToolShort(group, "ToolsHelp")
	if !found {
		t.Fatal("ToolsHelp tool not found")
	}

	tests := []struct {
		name                 string
		filterArg            map[string]lang.Value
		args                 []interface{}
		expectedToContain    []string
		expectedToNotContain []string
		checkNoToolsMsg      bool
		noToolsFilter        string
	}{
		{
			name: "No filter (all tools)",
			args: []interface{}{},
			expectedToContain: []string{
				"## `tool.Meta.ListTools`",
				"## `tool.FS.Read`",
			},
			expectedToNotContain: []string{"No tools found matching filter"},
		},
		{
			name: "Filter for Meta tools",
			args: []interface{}{""},
			expectedToContain: []string{
				"## `tool.Meta.ListTools`",
				"## `tool.Meta.ToolsHelp`",
			},
			expectedToNotContain: []string{"## `tool.FS.Read`"},
		},
		{
			name:            "Filter with no results",
			args:            []interface{}{"NoSucchToolExistsFilter123"},
			checkNoToolsMsg: true,
			noToolsFilter:   "NoSucchToolExistsFilter123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toolsHelpImpl.Func(interpreter, tt.args)
			if err != nil {
				t.Fatalf("ToolsHelp execution failed: %v. Args: %#v", err, tt.filterArg)
			}

			resultStr, ok := result.(string)
			if !ok {
				t.Fatalf("ToolsHelp did not return a lang.StringValue, got %T. Args: %#v", result, tt.filterArg)
			}

			for _, sub := range tt.expectedToContain {
				if !strings.Contains(resultStr, sub) {
					t.Errorf("ToolsHelp output for '%s' does not contain expected substring: '%s'\nOutput was:\n%s", tt.name, sub, resultStr)
				}
			}
			for _, sub := range tt.expectedToNotContain {
				if strings.Contains(resultStr, sub) {
					t.Errorf("ToolsHelp output for '%s' unexpectedly contains substring: '%s'\nOutput was:\n%s", tt.name, sub, resultStr)
				}
			}
			if tt.checkNoToolsMsg {
				expectedMsg := fmt.Sprintf("No tools found matching filter: `%s`", tt.noToolsFilter)
				if !strings.Contains(resultStr, expectedMsg) {
					t.Errorf("ToolsHelp output for '%s' expected to contain '%s', got '\n%s'", tt.name, expectedMsg, resultStr)
				}
			}
		})
	}
}
