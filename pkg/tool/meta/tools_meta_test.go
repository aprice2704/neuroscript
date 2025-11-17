// NeuroScript Major Version: 1
// File version: 12
// Purpose: Adds tests for new listToolNames and toolsHelp functions.
// Latest change: Made test assertions in TestToolMetaToolsHelp case-insensitive to match registry behavior.
// filename: pkg/tool/meta/tools_meta_test.go
// nlines: 161

package meta_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/tool/meta"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// mockRuntime is a minimal mock implementation of tool.Runtime for this test file.
type mockRuntime struct {
	registry   tool.ToolRegistry
	execPolicy *policy.ExecPolicy
}

// Statically assert that *mockRuntime satisfies the tool.Runtime interface.
var _ tool.Runtime = (*mockRuntime)(nil)

func (m *mockRuntime) ToolRegistry() tool.ToolRegistry   { return m.registry }
func (m *mockRuntime) GetExecPolicy() *policy.ExecPolicy { return m.execPolicy }
func (m *mockRuntime) GetGrantSet() *capability.GrantSet {
	if m.execPolicy != nil {
		return &m.execPolicy.Grants
	}
	return &capability.GrantSet{}
}

// --- Unused tool.Runtime methods ---
func (m *mockRuntime) Println(args ...any)                                   { fmt.Fprintln(os.Stderr, args...) } // DEBUG
func (m *mockRuntime) PromptUser(prompt string) (string, error)              { return "", nil }
func (m *mockRuntime) GetVar(name string) (any, bool)                        { return nil, false }
func (m *mockRuntime) SetVar(name string, val any)                           {}
func (m *mockRuntime) CallTool(name types.FullName, args []any) (any, error) { return nil, nil }
func (m *mockRuntime) GetLogger() interfaces.Logger                          { return nil } // Use Println for test debug
func (m *mockRuntime) SandboxDir() string                                    { return "" }
func (m *mockRuntime) LLM() interfaces.LLMClient                             { return nil }
func (m *mockRuntime) RegisterHandle(obj interface{}, typePrefix string) (string, error) {
	return "", nil
}
func (m *mockRuntime) GetHandleValue(handle, prefix string) (interface{}, error) { return nil, nil }
func (m *mockRuntime) AgentModels() interfaces.AgentModelReader                  { return nil }
func (m *mockRuntime) AgentModelsAdmin() interfaces.AgentModelAdmin              { return nil }

func newTestRuntime(t *testing.T) *mockRuntime {
	t.Helper()
	rt := &mockRuntime{}
	registry := tool.NewToolRegistry(rt)
	rt.registry = registry
	rt.execPolicy = &policy.ExecPolicy{
		Allow: []string{"tool.meta.*"}, // Allow all meta tools for testing
	}
	if err := meta.RegisterTools(registry); err != nil {
		t.Fatalf("Failed to register meta tools: %v", err)
	}
	return rt
}

func registerDummyTool(t *testing.T, rt *mockRuntime) {
	t.Helper()
	dummyTool := tool.ToolImplementation{
		Spec: tool.ToolSpec{
			Name:        "dummy",
			Group:       "test",
			Description: "A dummy test tool.",
			Args: []tool.ArgSpec{
				{Name: "input", Type: tool.ArgTypeString, Required: true, Description: "Some input."},
			},
			ReturnType: tool.ArgTypeString,
			ReturnHelp: "Returns 'ok'",
		},
		Func: func(rt tool.Runtime, args []any) (any, error) { return "ok", nil },
	}
	if _, err := rt.ToolRegistry().RegisterTool(dummyTool); err != nil {
		t.Fatalf("Failed to register dummy tool: %v", err)
	}
}

func TestToolMetaListTools(t *testing.T) {
	rt := newTestRuntime(t)
	registerDummyTool(t, rt)

	result, err := rt.ToolRegistry().ExecuteTool("tool.meta.listTools", map[string]lang.Value{})
	if err != nil {
		t.Fatalf("ExecuteTool failed unexpectedly: %v", err)
	}

	unwrapped := lang.Unwrap(result)
	specs, ok := unwrapped.([]interface{})
	if !ok {
		t.Fatalf("Expected result to be a slice of specs, but got %T", unwrapped)
	}

	// Should find the 4 meta tools + the 1 dummy tool
	if len(specs) != 5 {
		t.Errorf("Expected to find 5 tools, but got %d", len(specs))
	}
}

func TestToolMetaListToolNames(t *testing.T) {
	rt := newTestRuntime(t)
	registerDummyTool(t, rt) // Adds tool.test.dummy

	result, err := rt.ToolRegistry().ExecuteTool("tool.meta.listToolNames", map[string]lang.Value{})
	if err != nil {
		t.Fatalf("ExecuteTool failed unexpectedly: %v", err)
	}

	unwrapped := lang.Unwrap(result)
	namesStr, ok := unwrapped.(string)
	if !ok {
		t.Fatalf("Expected result to be a string, but got %T", unwrapped)
	}

	// We expect 5 tools: 4 meta tools + 1 dummy tool
	lines := strings.Split(strings.TrimSpace(namesStr), "\n")
	if len(lines) != 5 {
		t.Errorf("Expected 5 tool names, but got %d", len(lines))
	}

	// Check for the dummy tool's signature specifically
	expectedSig := "tool.test.dummy(input:string) -> string"
	if !strings.Contains(namesStr, expectedSig) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedSig, namesStr)
	}
}

func TestToolMetaToolsHelp(t *testing.T) {
	rt := newTestRuntime(t)
	registerDummyTool(t, rt) // Adds tool.test.dummy

	t.Run("No filter", func(t *testing.T) {
		result, err := rt.ToolRegistry().ExecuteTool("tool.meta.toolsHelp", map[string]lang.Value{})
		if err != nil {
			t.Fatalf("ExecuteTool failed unexpectedly: %v", err)
		}
		helpStr, ok := lang.Unwrap(result).(string)
		if !ok {
			t.Fatalf("Expected string result, got %T", result)
		}

		// FIX: Use case-insensitive comparison
		lowerHelpStr := strings.ToLower(helpStr)

		if !strings.Contains(lowerHelpStr, "tool.meta.listtools") {
			t.Error("Expected help string to contain 'tool.meta.listtools' (case-insensitive)")
		}
		if !strings.Contains(lowerHelpStr, "tool.test.dummy") {
			t.Error("Expected help string to contain 'tool.test.dummy' (case-insensitive)")
		}
		if !strings.Contains(lowerHelpStr, "## tool.test.dummy") {
			t.Error("Expected help string to contain markdown header for dummy tool")
		}
	})

	t.Run("With filter", func(t *testing.T) {
		// FIX: Call lang.Wrap separately and check for error
		wrappedFilter, err := lang.Wrap("dummy")
		if err != nil {
			t.Fatalf("Failed to wrap filter string: %v", err)
		}
		args := map[string]lang.Value{"filter": wrappedFilter}

		result, err := rt.ToolRegistry().ExecuteTool("tool.meta.toolsHelp", args)
		if err != nil {
			t.Fatalf("ExecuteTool failed unexpectedly: %v", err)
		}
		helpStr, ok := lang.Unwrap(result).(string)
		if !ok {
			t.Fatalf("Expected string result, got %T", result)
		}

		// FIX: Use case-insensitive comparison
		lowerHelpStr := strings.ToLower(helpStr)

		if strings.Contains(lowerHelpStr, "tool.meta.listtools") {
			t.Error("Expected help string to NOT contain 'tool.meta.listtools'")
		}
		if !strings.Contains(lowerHelpStr, "tool.test.dummy") {
			t.Error("Expected help string to contain 'tool.test.dummy' (case-insensitive)")
		}
		if !strings.Contains(helpStr, "| `input` | `string` | true | Some input. |") {
			t.Error("Expected help string to contain args table for dummy tool")
		}
	})

	t.Run("No match filter", func(t *testing.T) {
		// FIX: Call lang.Wrap separately and check for error
		wrappedFilter, err := lang.Wrap("nonexistent.tool")
		if err != nil {
			t.Fatalf("Failed to wrap filter string: %v", err)
		}
		args := map[string]lang.Value{"filter": wrappedFilter}

		result, err := rt.ToolRegistry().ExecuteTool("tool.meta.toolsHelp", args)
		if err != nil {
			t.Fatalf("ExecuteTool failed unexpectedly: %v", err)
		}
		helpStr, ok := lang.Unwrap(result).(string)
		if !ok {
			t.Fatalf("Expected string result, got %T", result)
		}

		expectedMsg := `No tools found matching filter: "nonexistent.tool"`
		if helpStr != expectedMsg {
			t.Errorf("Expected no-match message, got: %s", helpStr)
		}
	})
}
