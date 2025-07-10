// NeuroScript Version: 0.5.4
// File version: 2
// Purpose: Corrects test failures for meta tools by updating to fully qualified names, using the standard testing library, and ensuring proper test setup.
// filename: pkg/tool/meta/tools_meta_test.go
// nlines: 168
// risk_rating: LOW
package meta

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// newMetaTestInterpreter sets up a clean interpreter for each test run.
// It registers the 'meta' tools and some dummy tools to test filtering.
func newMetaTestInterpreter(t *testing.T) *interpreter.Interpreter {
	t.Helper()

	interp := interpreter.NewInterpreter(interpreter.WithLogger(logging.NewTestLogger(t)))

	// Register the actual 'meta' tools from the current package
	if err := tool.RegisterCoreTools(interp.ToolRegistry()); err != nil {
		t.Fatalf("Failed to register meta tools: %v", err)
	}

	// Register some dummy tools from other groups to test filtering logic
	dummyFSSpec := tool.ToolSpec{Name: "Read", Group: "FS", Description: "Dummy FS tool.", ReturnType: tool.ArgTypeAny}
	dummyFSFunc := func(rt tool.Runtime, args []interface{}) (interface{}, error) { return "dummy fs read", nil }
	if err := interp.ToolRegistry().RegisterTool(tool.ToolImplementation{Spec: dummyFSSpec, Func: dummyFSFunc}); err != nil {
		t.Fatalf("Failed to register dummy FS tool: %v", err)
	}

	dummyListSpec := tool.ToolSpec{Name: "Head", Group: "List", Description: "Dummy List tool.", ReturnType: tool.ArgTypeAny}
	dummyListFunc := func(rt tool.Runtime, args []interface{}) (interface{}, error) { return "dummy list head", nil }
	if err := interp.ToolRegistry().RegisterTool(tool.ToolImplementation{Spec: dummyListSpec, Func: dummyListFunc}); err != nil {
		t.Fatalf("Failed to register dummy List tool: %v", err)
	}

	return interp
}

func TestToolMetaListTools(t *testing.T) {
	interpreter := newMetaTestInterpreter(t)
	fullName := types.MakeFullName(group, "ListTools")

	listToolsImpl, found := interpreter.ToolRegistry().GetTool(fullName)
	if !found {
		t.Fatalf("Tool %q not found in registry", fullName)
	}

	result, err := listToolsImpl.Func(interpreter, []interface{}{})
	if err != nil {
		t.Fatalf("ListTools execution failed unexpectedly: %v", err)
	}

	resultStr, ok := result.(string)
	if !ok {
		t.Fatalf("ListTools did not return a string, got %T", result)
	}

	// Check for the presence of fully qualified tool names
	expectedTools := []string{
		"tool.Meta.ListTools",
		"tool.FS.Read",
		"tool.List.Head",
	}
	for _, expected := range expectedTools {
		if !strings.Contains(resultStr, expected) {
			t.Errorf("ListTools output is missing expected tool: %s\nOutput was:\n%s", expected, resultStr)
		}
	}
}

func TestToolMetaGetToolSpecificationsJSON(t *testing.T) {
	interpreter := newMetaTestInterpreter(t)
	fullName := types.MakeFullName(group, "GetToolSpecificationsJSON")
	getJsonImpl, found := interpreter.ToolRegistry().GetTool(fullName)
	if !found {
		t.Fatalf("Tool %q not found in registry", fullName)
	}

	result, err := getJsonImpl.Func(interpreter, []interface{}{})
	if err != nil {
		t.Fatalf("GetToolSpecificationsJSON execution failed unexpectedly: %v", err)
	}
	jsonStr, ok := result.(string)
	if !ok {
		t.Fatalf("GetToolSpecificationsJSON did not return a string, got %T", result)
	}
	// Basic sanity check: is it valid JSON?
	var specs []tool.ToolSpec
	if err := json.Unmarshal([]byte(jsonStr), &specs); err != nil {
		t.Fatalf("Failed to unmarshal JSON output from GetToolSpecificationsJSON: %v\nOutput:\n%s", err, jsonStr)
	}
	if len(specs) < 3 { // Meta tools + dummy tools
		t.Errorf("Expected at least 3 tool specs in JSON, got %d", len(specs))
	}
}

func TestToolMetaToolsHelp(t *testing.T) {
	interpreter := newMetaTestInterpreter(t)
	fullName := types.MakeFullName(group, "ToolsHelp")

	toolsHelpImpl, found := interpreter.ToolRegistry().GetTool(fullName)
	if !found {
		t.Fatalf("Tool %q not found in registry", fullName)
	}

	testCases := []struct {
		name                 string
		filterArg            []interface{}
		expectedToContain    []string
		expectedToNotContain []string
	}{
		{
			name:      "No_filter_(all_tools)",
			filterArg: []interface{}{},
			expectedToContain: []string{
				"## `tool.Meta.ListTools`",
				"## `tool.FS.Read`",
				"## `tool.List.Head`",
			},
		},
		{
			name:      "Filter_for_Meta_tools_(case_insensitive)",
			filterArg: []interface{}{"meta"}, // Use lowercase to test case-insensitivity
			expectedToContain: []string{
				"## `tool.Meta.ListTools`",
				"## `tool.Meta.ToolsHelp`",
			},
			expectedToNotContain: []string{
				"## `tool.FS.Read`",
				"## `tool.List.Head`",
			},
		},
		{
			name:      "Filter_with_no_results",
			filterArg: []interface{}{"NonExistentToolFilter"},
			expectedToContain: []string{
				"No tools found matching filter: `NonExistentToolFilter`",
			},
			expectedToNotContain: []string{"## `"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := toolsHelpImpl.Func(interpreter, tc.filterArg)
			if err != nil {
				t.Fatalf("ToolsHelp execution failed: %v", err)
			}
			resultStr, ok := result.(string)
			if !ok {
				t.Fatalf("ToolsHelp did not return a string, got %T", result)
			}
			for _, sub := range tc.expectedToContain {
				if !strings.Contains(resultStr, sub) {
					t.Errorf("ToolsHelp output for '%s' does not contain expected substring: '%s'\n---Output---\n%s\n------------", tc.name, sub, resultStr)
				}
			}
			for _, sub := range tc.expectedToNotContain {
				if strings.Contains(resultStr, sub) {
					t.Errorf("ToolsHelp output for '%s' unexpectedly contains substring: '%s'\n---Output---\n%s\n------------", tc.name, sub, resultStr)
				}
			}
		})
	}
}
