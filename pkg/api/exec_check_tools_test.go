// NeuroScript Version: 0.8.0
// File version: 8
// Purpose: Adds a test case for tool calls within event handlers.
// filename: pkg/api/exec_check_tools_test.go
// nlines: 120+
// risk_rating: LOW

package api_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// Helper to create a basic interpreter with optional dummy tools for testing CheckScriptTools.
func setupCheckToolsTest(t *testing.T, toolNames ...string) *api.Interpreter {
	t.Helper()
	// Assumes newTestHostContext is available from harness_test.go
	interp := api.New(api.WithHostContext(newTestHostContext(nil)))

	for _, name := range toolNames {
		parts := strings.Split(name, ".")
		if len(parts) < 3 {
			t.Fatalf("Invalid dummy tool name format for test setup: %s", name)
		}
		dummyTool := api.ToolImplementation{
			Spec: api.ToolSpec{
				Group: api.ToolGroup(parts[1]),
				Name:  api.ToolName(parts[2]),
			},
			Func: func(rt api.Runtime, args []interface{}) (interface{}, error) { return nil, nil },
		}
		if _, err := interp.ToolRegistry().RegisterTool(dummyTool); err != nil {
			t.Fatal(fmt.Sprintf("Failed to register dummy tool %s: %v", name, err))
		}
	}
	return interp
}

func TestCheckScriptTools(t *testing.T) {
	tests := []struct {
		name               string
		script             string
		registeredTools    []string
		wantErr            bool
		wantErrIs          error
		wantErrMsgContains []string
	}{
		{
			name:            "No tool calls",
			script:          "func main() means\n return 1 \nendfunc",
			registeredTools: []string{},
			wantErr:         false,
		},
		{
			name:            "Required tool exists",
			script:          "command\n call tool.test.dummy() \nendcommand",
			registeredTools: []string{"tool.test.dummy"},
			wantErr:         false,
		},
		{
			name:               "Required tool missing",
			script:             "command\n call tool.test.missing() \nendcommand",
			registeredTools:    []string{"tool.test.dummy"},
			wantErr:            true,
			wantErrIs:          lang.ErrToolNotFound,
			wantErrMsgContains: []string{"test.missing"},
		},
		{
			name: "Multiple tools, one missing",
			script: `
                command
                    call tool.test.exists1()
                    call tool.test.missing()
                    call tool.test.exists2()
                endcommand`,
			registeredTools:    []string{"tool.test.exists1", "tool.test.exists2"},
			wantErr:            true,
			wantErrIs:          lang.ErrToolNotFound,
			wantErrMsgContains: []string{"test.missing"},
		},
		{
			name: "Multiple missing tools",
			script: `
                command
                    call tool.test.missing1()
                    call tool.test.exists1()
                    call tool.test.missing2()
                endcommand`,
			registeredTools:    []string{"tool.test.exists1"},
			wantErr:            true,
			wantErrIs:          lang.ErrToolNotFound,
			wantErrMsgContains: []string{"test.missing1", "test.missing2"},
		},
		{
			name: "Tool call inside function",
			script: `
                func myfunc() means
                    call tool.test.needed()
                endfunc`,
			registeredTools:    []string{},
			wantErr:            true,
			wantErrIs:          lang.ErrToolNotFound,
			wantErrMsgContains: []string{"test.needed"},
		},
		{
			name: "Tool call inside event handler",
			script: `
                on event "test.event" do
                    call tool.test.event_tool()
                endon`,
			registeredTools:    []string{},
			wantErr:            true,
			wantErrIs:          lang.ErrToolNotFound,
			wantErrMsgContains: []string{"test.event_tool"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := setupCheckToolsTest(t, tt.registeredTools...)

			tree, err := api.Parse([]byte(tt.script), api.ParseSkipComments)
			if err != nil {
				t.Fatalf("api.Parse failed: %v", err)
			}

			checkErr := api.CheckScriptTools(tree, interp)

			if tt.wantErr {
				if checkErr == nil {
					t.Fatal("Expected an error from CheckScriptTools, but got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(checkErr, tt.wantErrIs) {
					t.Errorf("Expected error to wrap [%v], but got: %v", tt.wantErrIs, checkErr)
				}
				for _, substring := range tt.wantErrMsgContains {
					expectedSubstring := "tool." + substring
					if !strings.Contains(checkErr.Error(), expectedSubstring) {
						t.Errorf("Expected error message to contain %q, but got: %v", expectedSubstring, checkErr)
					}
				}
			} else {
				if checkErr != nil {
					t.Fatalf("Expected CheckScriptTools to succeed, but got error: %v", checkErr)
				}
			}
		})
	}

	// --- Edge Cases ---
	t.Run("Nil Tree", func(t *testing.T) {
		interp := setupCheckToolsTest(t)
		err := api.CheckScriptTools(nil, interp)
		if err == nil {
			t.Fatal("Expected an error for nil tree, but got nil")
		}
	})

	t.Run("Nil Interpreter", func(t *testing.T) {
		tree, _ := api.Parse([]byte("command\n emit 1 \nendcommand"), api.ParseSkipComments)
		err := api.CheckScriptTools(tree, nil)
		if err == nil {
			t.Fatal("Expected an error for nil interpreter, but got nil")
		}
	})
}
