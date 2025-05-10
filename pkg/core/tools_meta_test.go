// NeuroScript Version: 0.3.8
// File version: 0.1.1 // Minor log adjustment to reflect code changes
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
	interpreter, _ := NewDefaultTestInterpreter(t)

	result, err := interpreter.ExecuteTool("Meta.ListTools", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Meta.ListTools execution failed: %v", err)
	}

	resultStr, ok := result.(string)
	if !ok {
		t.Fatalf("Meta.ListTools did not return a string, got %T", result)
	}

	// t.Logf("Meta.ListTools output:\n%s", resultStr) // Optional: keep for debugging

	expectedSignatures := []string{
		"Meta.ListTools() -> string",
		"Meta.ToolsHelp(filter:string?) -> string",
		"FS.Read(filepath:string) -> string",
		"Go.Build(path:string?) -> map", // Assuming 'path' is its optional string arg
		// AIWorker.ExecuteStatelessTask is very long, checking for its presence is enough
		"AIWorker.ExecuteStatelessTask(",
	}

	for _, sig := range expectedSignatures {
		if !strings.Contains(resultStr, sig) {
			parts := strings.Split(sig, "(")
			baseName := parts[0]
			returnTypeAndRest := strings.Split(sig, "-> ")
			returnType := ""
			if len(returnTypeAndRest) > 1 {
				returnType = strings.TrimSpace(returnTypeAndRest[1])
			}

			partialSigFound := strings.Contains(resultStr, baseName) && (returnType == "" || strings.Contains(resultStr, "-> "+returnType))

			if !partialSigFound {
				t.Errorf("Meta.ListTools output does not contain expected signature element: looking for base '%s' and return '%s'.\nFull expected: %s\nOutput was:\n%s", baseName, returnType, sig, resultStr)
			}
		}
	}

	idxFSRead := strings.Index(resultStr, "FS.Read(")
	idxGoBuild := strings.Index(resultStr, "Go.Build(")
	idxMetaList := strings.Index(resultStr, "Meta.ListTools(")

	if idxFSRead == -1 || idxGoBuild == -1 || idxMetaList == -1 {
		t.Errorf("One or more key tools for sorting check not found in output: FS.Read (found at %d), Go.Build (found at %d), Meta.ListTools (found at %d)", idxFSRead, idxGoBuild, idxMetaList)
	} else {
		if !(idxFSRead < idxGoBuild && idxGoBuild < idxMetaList) {
			t.Errorf("Meta.ListTools output does not appear to be sorted correctly. Indices: FS.Read=%d, Go.Build=%d, Meta.ListTools=%d", idxFSRead, idxGoBuild, idxMetaList)
		}
	}

	if strings.TrimSpace(resultStr) == "" {
		t.Error("Meta.ListTools output is empty")
	}
}

func TestToolMetaToolsHelp(t *testing.T) {
	interpreter, _ := NewDefaultTestInterpreter(t)

	tests := []struct {
		name                 string
		filterArg            map[string]interface{}
		expectedToContain    []string
		expectedToNotContain []string
		checkNoToolsMsg      bool
		noToolsFilter        string // Original casing for message check
	}{
		{
			name:      "No filter (all tools)",
			filterArg: map[string]interface{}{},
			expectedToContain: []string{
				"## `tool.Meta.ListTools`",
				"**Description:**",
				"**Parameters:**",
				"**Returns:** (`string`)",
				"## `tool.FS.Read`",
				"## `tool.Go.Build`",
			},
			expectedToNotContain: []string{"No tools found matching filter", "Showing tools matching filter:"}, // Should not show filter-related messages
		},
		{
			name:      "Filter for Meta tools",
			filterArg: map[string]interface{}{"filter": "Meta."}, // Original casing
			expectedToContain: []string{
				"## `tool.Meta.ListTools`",
				"## `tool.Meta.ToolsHelp`",
				"Showing tools matching filter: `Meta.`", // Expect original casing in message
			},
			expectedToNotContain: []string{"## `tool.FS.Read`"},
		},
		{
			name:      "Filter for FS tools",
			filterArg: map[string]interface{}{"filter": "FS."}, // Original casing
			expectedToContain: []string{
				"## `tool.FS.Read`",
				"## `tool.FS.Write`",
				"Showing tools matching filter: `FS.`", // Expect original casing
			},
			expectedToNotContain: []string{"## `tool.Meta.ListTools`"},
		},
		{
			name:            "Filter with no results",
			filterArg:       map[string]interface{}{"filter": "NoSucchToolExistsFilter123"}, // Original casing
			checkNoToolsMsg: true,
			noToolsFilter:   "NoSucchToolExistsFilter123", // Original casing for message check
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Log the arguments being passed to ExecuteTool for clarity during testing
			// t.Logf("Executing Meta.ToolsHelp with args: %#v", tt.filterArg)

			result, err := interpreter.ExecuteTool("Meta.ToolsHelp", tt.filterArg)
			if err != nil {
				t.Fatalf("Meta.ToolsHelp execution failed: %v. Args: %#v", err, tt.filterArg)
			}

			resultStr, ok := result.(string)
			if !ok {
				t.Fatalf("Meta.ToolsHelp did not return a string, got %T. Args: %#v", result, tt.filterArg)
			}

			// t.Logf("Meta.ToolsHelp output for '%s':\n%s", tt.name, resultStr) // Optional

			for _, sub := range tt.expectedToContain {
				if !strings.Contains(resultStr, sub) {
					t.Errorf("Meta.ToolsHelp output for '%s' does not contain expected substring: '%s'\nOutput was:\n%s", tt.name, sub, resultStr)
				}
			}
			for _, sub := range tt.expectedToNotContain {
				if strings.Contains(resultStr, sub) {
					t.Errorf("Meta.ToolsHelp output for '%s' unexpectedly contains substring: '%s'\nOutput was:\n%s", tt.name, sub, resultStr)
				}
			}
			if tt.checkNoToolsMsg {
				// The tool now uses the original casing of the filter in its "not found" message.
				expectedMsg := fmt.Sprintf("No tools found matching filter: `%s`", tt.noToolsFilter)
				if !strings.Contains(resultStr, expectedMsg) {
					// Also handle the edge case where NO tools are registered AT ALL (unlikely here as core tools are registered)
					// or if the message about *no filter provided but no tools* is returned.
					if resultStr != "No tools are currently registered." {
						t.Errorf("Meta.ToolsHelp output for '%s' expected to contain '%s', got '\n%s'", tt.name, expectedMsg, resultStr)
					}
				}
			}
		})
	}
}
