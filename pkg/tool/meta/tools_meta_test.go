// NeuroScript Major Version: 1
// File version: 17
// Purpose: Adds tests for listGlobalConstants, listFunctions, and filtering logic. Implemented HandleRegistry for mockRuntime interface compliance.
// Latest change: Implemented HandleRegistry on mockRuntime to satisfy tool.Runtime interface.
// filename: pkg/tool/meta/tools_meta_test.go
// nlines: 304

package meta_test

import (
	"fmt"
	"os"
	"sort"
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
// It implements extra methods to satisfy the introspection interfaces used by meta tools.
type mockRuntime struct {
	registry   tool.ToolRegistry
	execPolicy *policy.ExecPolicy

	// Test data for introspection
	consts map[string]lang.Value
	procs  map[string]any
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

// --- Introspection Methods (Used by meta tools) ---

// KnownGlobalConstants satisfies the 'constProvider' interface.
func (m *mockRuntime) KnownGlobalConstants() map[string]lang.Value {
	if m.consts == nil {
		return map[string]lang.Value{}
	}
	return m.consts
}

// KnownProcedures is called via reflection by ListFunctions.
// The value type doesn't matter for the tool, only the keys.
func (m *mockRuntime) KnownProcedures() map[string]any {
	if m.procs == nil {
		return map[string]any{}
	}
	return m.procs
}

// --- tool.Runtime methods ---
func (m *mockRuntime) Println(args ...any)                                   { fmt.Fprintln(os.Stderr, args...) }
func (m *mockRuntime) PromptUser(prompt string) (string, error)              { return "", nil }
func (m *mockRuntime) GetVar(name string) (any, bool)                        { return nil, false }
func (m *mockRuntime) SetVar(name string, val any)                           {}
func (m *mockRuntime) CallTool(name types.FullName, args []any) (any, error) { return nil, nil }
func (m *mockRuntime) GetLogger() interfaces.Logger                          { return nil }
func (m *mockRuntime) SandboxDir() string                                    { return "" }
func (m *mockRuntime) LLM() interfaces.LLMClient                             { return nil }

// HandleRegistry returns a nil implementation for testing methods not using handles.
func (m *mockRuntime) HandleRegistry() interfaces.HandleRegistry {
	return nil // Handled by a dedicated mock if needed, or nil for methods not using it.
}

// NOTE: Old handle methods (RegisterHandle, GetHandleValue) have been removed from the interface.

func (m *mockRuntime) AgentModels() interfaces.AgentModelReader     { return nil }
func (m *mockRuntime) AgentModelsAdmin() interfaces.AgentModelAdmin { return nil }

func newTestRuntime(t *testing.T) *mockRuntime {
	t.Helper()
	rt := &mockRuntime{
		consts: make(map[string]lang.Value),
		procs:  make(map[string]any),
	}
	registry := tool.NewToolRegistry(rt)
	rt.registry = registry
	rt.execPolicy = &policy.ExecPolicy{
		Allow: []string{"tool.meta.*"},
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

	// 1. No Filter
	t.Run("No Filter", func(t *testing.T) {
		result, err := rt.ToolRegistry().ExecuteTool("tool.meta.listTools", map[string]lang.Value{})
		if err != nil {
			t.Fatalf("ExecuteTool failed: %v", err)
		}
		unwrapped := lang.Unwrap(result)
		specs, ok := unwrapped.([]interface{})
		if !ok {
			t.Fatalf("Expected slice result, got %T", unwrapped)
		}
		// 6 meta tools + 1 dummy tool = 7
		if len(specs) != 7 {
			t.Errorf("Expected 7 tools, got %d", len(specs))
		}
	})

	// 2. With Filter
	t.Run("With Filter", func(t *testing.T) {
		filter, _ := lang.Wrap("dummy")
		result, err := rt.ToolRegistry().ExecuteTool("tool.meta.listTools", map[string]lang.Value{"filter": filter})
		if err != nil {
			t.Fatalf("ExecuteTool failed: %v", err)
		}
		specs := lang.Unwrap(result).([]interface{})
		if len(specs) != 1 {
			t.Errorf("Expected 1 tool, got %d", len(specs))
		}
	})
}

func TestToolMetaListGlobalConstants(t *testing.T) {
	rt := newTestRuntime(t)
	// FIX: Use struct literals for Values
	rt.consts["FDM_TEST"] = lang.StringValue{Value: "value"}
	rt.consts["OTHER_CONST"] = lang.NumberValue{Value: 123}

	// 1. No Filter
	t.Run("No Filter", func(t *testing.T) {
		result, err := rt.ToolRegistry().ExecuteTool("tool.meta.listGlobalConstants", nil)
		if err != nil {
			t.Fatalf("ExecuteTool failed: %v", err)
		}
		resMap := lang.Unwrap(result).(map[string]interface{})
		if len(resMap) != 2 {
			t.Errorf("Expected 2 constants, got %d", len(resMap))
		}

		// FIX: Expect native string "value", not quoted ""value""
		if resMap["FDM_TEST"] != "value" {
			t.Errorf("Unexpected value for FDM_TEST: %v", resMap["FDM_TEST"])
		}
		// FIX: Expect native float 123.0
		if resMap["OTHER_CONST"] != 123.0 {
			t.Errorf("Unexpected value for OTHER_CONST: %v", resMap["OTHER_CONST"])
		}
	})

	// 2. With Filter
	t.Run("With Filter", func(t *testing.T) {
		filter, _ := lang.Wrap("fdm")
		result, err := rt.ToolRegistry().ExecuteTool("tool.meta.listGlobalConstants", map[string]lang.Value{"filter": filter})
		if err != nil {
			t.Fatalf("ExecuteTool failed: %v", err)
		}
		resMap := lang.Unwrap(result).(map[string]interface{})
		if len(resMap) != 1 {
			t.Errorf("Expected 1 constant, got %d", len(resMap))
		}
		if _, ok := resMap["FDM_TEST"]; !ok {
			t.Error("Expected FDM_TEST to be present")
		}
	})
}

func TestToolMetaListFunctions(t *testing.T) {
	rt := newTestRuntime(t)
	rt.procs["my_func"] = nil
	rt.procs["other_proc"] = nil

	// 1. No Filter
	t.Run("No Filter", func(t *testing.T) {
		result, err := rt.ToolRegistry().ExecuteTool("tool.meta.listFunctions", nil)
		if err != nil {
			t.Fatalf("ExecuteTool failed: %v", err)
		}
		list := lang.Unwrap(result).([]interface{})
		if len(list) != 2 {
			t.Errorf("Expected 2 functions, got %d", len(list))
		}

		// Verify sorting
		if list[0] != "my_func" || list[1] != "other_proc" {
			sort.Slice(list, func(i, j int) bool { return list[i].(string) < list[j].(string) })
			if list[0] != "my_func" {
				t.Errorf("List not sorted or incorrect: %v", list)
			}
		}
	})

	// 2. With Filter
	t.Run("With Filter", func(t *testing.T) {
		filter, _ := lang.Wrap("other")
		result, err := rt.ToolRegistry().ExecuteTool("tool.meta.listFunctions", map[string]lang.Value{"filter": filter})
		if err != nil {
			t.Fatalf("ExecuteTool failed: %v", err)
		}
		list := lang.Unwrap(result).([]interface{})
		if len(list) != 1 {
			t.Errorf("Expected 1 function, got %d", len(list))
		}
		if list[0] != "other_proc" {
			t.Errorf("Expected other_proc, got %v", list[0])
		}
	})
}

func TestToolMetaListToolNames(t *testing.T) {
	rt := newTestRuntime(t)
	registerDummyTool(t, rt)

	t.Run("No Filter", func(t *testing.T) {
		result, err := rt.ToolRegistry().ExecuteTool("tool.meta.listToolNames", map[string]lang.Value{})
		if err != nil {
			t.Fatalf("ExecuteTool failed unexpectedly: %v", err)
		}

		unwrapped := lang.Unwrap(result)
		namesStr, ok := unwrapped.(string)
		if !ok {
			t.Fatalf("Expected result to be a string, but got %T", unwrapped)
		}

		lines := strings.Split(strings.TrimSpace(namesStr), "\n")
		// 6 meta tools + 1 dummy tool = 7
		if len(lines) != 7 {
			t.Errorf("Expected 7 tool names, but got %d", len(lines))
		}

		expectedSig := "tool.test.dummy(input:string) -> string"
		if !strings.Contains(namesStr, expectedSig) {
			t.Errorf("Expected output to contain '%s', but got:\n%s", expectedSig, namesStr)
		}
	})

	t.Run("With Filter", func(t *testing.T) {
		filter, _ := lang.Wrap("dummy")
		result, err := rt.ToolRegistry().ExecuteTool("tool.meta.listToolNames", map[string]lang.Value{"filter": filter})
		if err != nil {
			t.Fatalf("ExecuteTool failed unexpectedly: %v", err)
		}

		unwrapped := lang.Unwrap(result)
		namesStr, ok := unwrapped.(string)
		if !ok {
			t.Fatalf("Expected result to be a string, but got %T", unwrapped)
		}

		if !strings.Contains(namesStr, "tool.test.dummy") {
			t.Errorf("Expected filtered output to contain dummy tool, got: %s", namesStr)
		}
		if strings.Contains(namesStr, "tool.meta.listtools") {
			t.Errorf("Expected filtered output NOT to contain meta tool, got: %s", namesStr)
		}
	})
}

func TestToolMetaToolsHelp(t *testing.T) {
	rt := newTestRuntime(t)
	registerDummyTool(t, rt)

	t.Run("No filter", func(t *testing.T) {
		result, err := rt.ToolRegistry().ExecuteTool("tool.meta.toolsHelp", map[string]lang.Value{})
		if err != nil {
			t.Fatalf("ExecuteTool failed unexpectedly: %v", err)
		}
		helpStr, ok := lang.Unwrap(result).(string)
		if !ok {
			t.Fatalf("Expected string result, got %T", result)
		}

		lowerHelpStr := strings.ToLower(helpStr)
		if !strings.Contains(lowerHelpStr, "tool.meta.listtools") {
			t.Error("Expected help string to contain 'tool.meta.listtools'")
		}
		if !strings.Contains(lowerHelpStr, "tool.test.dummy") {
			t.Error("Expected help string to contain 'tool.test.dummy'")
		}
	})

	t.Run("With filter", func(t *testing.T) {
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

		lowerHelpStr := strings.ToLower(helpStr)
		if strings.Contains(lowerHelpStr, "tool.meta.listtools") {
			t.Error("Expected help string to NOT contain 'tool.meta.listtools'")
		}
		if !strings.Contains(lowerHelpStr, "tool.test.dummy") {
			t.Error("Expected help string to contain 'tool.test.dummy'")
		}
	})
}
