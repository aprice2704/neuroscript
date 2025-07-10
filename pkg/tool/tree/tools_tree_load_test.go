// NeuroScript Version: 0.5.4
// File version: 7
// Purpose: Corrects load tests by using the updated helpers and compacting JSON for robust comparison.
// filename: pkg/tool/tree/tools_tree_load_test.go
// nlines: 75
// risk_rating: LOW
package tree

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestTreeLoadJSONAndToJSON(t *testing.T) {
	testCases := []treeTestCase{
		{
			Name:      "Load_and_ToJSON",
			JSONInput: `{"a": 1, "b": "hello"}`,
			ToolName:  "ToJSON",
			Expected:  `{"a":1,"b":"hello"}`,
		},
		{
			Name:        "LoadJSON_Invalid_JSON",
			JSONInput:   `{"a": 1, "b": }`,
			ToolName:    "LoadJSON",
			ExpectedErr: lang.ErrTreeJSONUnmarshal,
		},
		{
			Name:        "LoadJSON_Empty_Input",
			ToolName:    "LoadJSON",
			Args:        []interface{}{""},
			ExpectedErr: lang.ErrTreeJSONUnmarshal,
		},
		{
			Name:        "LoadJSON_Wrong_Arg_Type",
			ToolName:    "LoadJSON",
			Args:        []interface{}{12345},
			ExpectedErr: lang.ErrInvalidArgument,
		},
	}

	for _, tc := range testCases {
		testTreeToolHelper(t, tc.Name, func(t *testing.T, interp *interpreter.Interpreter) {
			var treeHandle string
			var err error

			if tc.JSONInput != "" {
				treeHandle, err = setupTreeWithJSON(t, interp, tc.JSONInput)
				if err != nil {
					if tc.ExpectedErr != nil && tc.ToolName == "LoadJSON" {
						assertResult(t, nil, err, nil, tc.ExpectedErr)
						return
					}
					t.Fatalf("Tree setup failed unexpectedly: %v", err)
				}
			}

			var result interface{}
			// The logic needs to differentiate between calling LoadJSON directly vs. another tool
			if tc.ToolName == "LoadJSON" {
				result, err = runTool(t, interp, tc.ToolName, tc.Args...)
			} else {
				// For other tools like ToJSON, the handle is the first argument
				args := append([]interface{}{treeHandle}, tc.Args...)
				result, err = runTool(t, interp, tc.ToolName, args...)
			}

			// Compacting the JSON for comparison
			if jsonStr, ok := result.(string); ok {
				var compactBuf bytes.Buffer
				if err := json.Compact(&compactBuf, []byte(jsonStr)); err == nil {
					result = compactBuf.String()
				}
			}

			assertResult(t, result, err, tc.Expected, tc.ExpectedErr)
		})
	}
}
