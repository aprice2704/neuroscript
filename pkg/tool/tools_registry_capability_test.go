// NeuroScript Version: 0.4.0
// File version: 6
// Purpose: Corrected test failures by properly setting grants within the mock execution policy.
// filename: pkg/tool/tools_registry_capability_test.go
// nlines: 162
// risk_rating: MEDIUM

package tool_test

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// testRuntime is a minimal mock implementation of tool.Runtime for this test file.
type testRuntime struct {
	registry   tool.ToolRegistry
	execPolicy *policy.ExecPolicy
}

// Statically assert that *testRuntime satisfies the tool.Runtime interface.
var _ tool.Runtime = (*testRuntime)(nil)

func (t *testRuntime) ToolRegistry() tool.ToolRegistry   { return t.registry }
func (t *testRuntime) GetExecPolicy() *policy.ExecPolicy { return t.execPolicy }
func (t *testRuntime) GetGrantSet() *capability.GrantSet {
	if t.execPolicy != nil {
		return &t.execPolicy.Grants
	}
	return &capability.GrantSet{}
}

// --- Unused tool.Runtime methods, stubbed out to satisfy the interface ---
func (t *testRuntime) Println(...any)                                        {}
func (t *testRuntime) PromptUser(prompt string) (string, error)              { return "", nil }
func (t *testRuntime) GetVar(name string) (any, bool)                        { return nil, false }
func (t *testRuntime) SetVar(name string, val any)                           {}
func (t *testRuntime) CallTool(name types.FullName, args []any) (any, error) { return nil, nil }
func (t *testRuntime) GetLogger() interfaces.Logger                          { return nil }
func (t *testRuntime) SandboxDir() string                                    { return "" }
func (t *testRuntime) LLM() interfaces.LLMClient                             { return nil }
func (t *testRuntime) RegisterHandle(obj interface{}, typePrefix string) (string, error) {
	return "", nil
}
func (t *testRuntime) GetHandleValue(handle string, expectedTypePrefix string) (interface{}, error) {
	return nil, nil
}
func (t *testRuntime) AgentModels() interfaces.AgentModelReader     { return nil }
func (t *testRuntime) AgentModelsAdmin() interfaces.AgentModelAdmin { return nil }

// secureTool is a sample tool implementation that requires a specific capability.
var secureTool = tool.ToolImplementation{
	Spec: tool.ToolSpec{
		Name:  "writeFile",
		Group: "fs",
	},
	Func: func(rt tool.Runtime, args []interface{}) (interface{}, error) {
		return "wrote file successfully", nil
	},
	RequiredCaps: []capability.Capability{
		{Resource: "fs", Verbs: []string{"write"}, Scopes: []string{"/data/user.txt"}},
	},
}

// TestCallFromInterpreter_CapabilityCheck_Success verifies that a tool call is
// allowed when the interpreter's policy grants the required capabilities.
func TestCallFromInterpreter_CapabilityCheck_Success(t *testing.T) {
	mockRuntime := &testRuntime{}
	registry := tool.NewToolRegistry(mockRuntime)
	_, err := registry.RegisterTool(secureTool)
	if err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}
	mockRuntime.registry = registry

	grantedCaps := []capability.Capability{
		{Resource: "fs", Verbs: []string{"write"}, Scopes: []string{"/data/*"}},
	}
	mockRuntime.execPolicy = &policy.ExecPolicy{
		Context: policy.ContextNormal,
		Allow:   []string{"tool.fs.writefile"},
		Grants:  capability.GrantSet{Grants: grantedCaps},
	}

	_, execErr := registry.ExecuteTool("tool.fs.writeFile", map[string]lang.Value{})

	if execErr != nil {
		t.Errorf("Expected tool call to succeed, but it failed with error: %v", execErr)
	}
}

// TestCallFromInterpreter_CapabilityCheck_Failure verifies that a tool call is
// blocked when the required capabilities are not granted.
func TestCallFromInterpreter_CapabilityCheck_Failure(t *testing.T) {
	mockRuntime := &testRuntime{}
	registry := tool.NewToolRegistry(mockRuntime)
	_, err := registry.RegisterTool(secureTool)
	if err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}
	mockRuntime.registry = registry

	grantedCaps := []capability.Capability{
		{Resource: "fs", Verbs: []string{"read"}, Scopes: []string{"*"}},
	}
	mockRuntime.execPolicy = &policy.ExecPolicy{
		Context: policy.ContextNormal,
		Allow:   []string{"tool.fs.writefile"},
		Grants:  capability.GrantSet{Grants: grantedCaps},
	}

	_, execErr := registry.ExecuteTool("tool.fs.writeFile", map[string]lang.Value{})

	if execErr == nil {
		t.Fatal("Expected tool call to fail due to insufficient capabilities, but it succeeded.")
	}
	var rtErr *lang.RuntimeError
	if !errors.As(execErr, &rtErr) || rtErr.Code != lang.ErrorCodePolicy {
		t.Errorf("Expected a policy error, but got: %v", execErr)
	}
}

// TestCallFromInterpreter_CapabilityCheck_NoRequirements confirms that a tool
// with no capability requirements can be called successfully.
func TestCallFromInterpreter_CapabilityCheck_NoRequirements(t *testing.T) {
	insecureTool := tool.ToolImplementation{
		Spec: tool.ToolSpec{Name: "add", Group: "math"},
		Func: func(rt tool.Runtime, args []interface{}) (interface{}, error) { return 42.0, nil },
	}
	mockRuntime := &testRuntime{}
	registry := tool.NewToolRegistry(mockRuntime)
	_, err := registry.RegisterTool(insecureTool)
	if err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}
	mockRuntime.registry = registry
	mockRuntime.execPolicy = &policy.ExecPolicy{Context: policy.ContextNormal, Allow: []string{"tool.math.add"}}

	result, execErr := registry.ExecuteTool("tool.math.add", map[string]lang.Value{})

	if execErr != nil {
		t.Errorf("Expected tool call to succeed, but it failed: %v", execErr)
	}
	if val, _ := lang.Unwrap(result).(float64); val != 42.0 {
		t.Errorf("Expected result 42.0, got %v", result)
	}
}

// TestCallFromInterpreter_CapabilityCheck_MultipleRequirements verifies that a tool
// requiring multiple capabilities is only allowed when all are granted.
func TestCallFromInterpreter_CapabilityCheck_MultipleRequirements(t *testing.T) {
	multiCapTool := tool.ToolImplementation{
		Spec: tool.ToolSpec{Name: "deployApp", Group: "cloud"},
		Func: func(rt tool.Runtime, args []interface{}) (interface{}, error) { return "deployed", nil },
		RequiredCaps: []capability.Capability{
			{Resource: "net", Verbs: []string{"write"}, Scopes: []string{"*.cloudprovider.com"}},
			{Resource: "fs", Verbs: []string{"read"}, Scopes: []string{"/src/app/**"}},
		},
	}

	testCases := []struct {
		name        string
		grantedCaps []capability.Capability
		shouldFail  bool
	}{
		{"Success - All capabilities granted", []capability.Capability{{Resource: "net", Verbs: []string{"write"}, Scopes: []string{"*"}}, {Resource: "fs", Verbs: []string{"read"}, Scopes: []string{"/src/app/**"}}}, false},
		{"Failure - Missing one capability", []capability.Capability{{Resource: "net", Verbs: []string{"write"}, Scopes: []string{"*"}}}, true},
		{"Failure - No capabilities granted", []capability.Capability{}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRuntime := &testRuntime{}
			registry := tool.NewToolRegistry(mockRuntime)
			_, err := registry.RegisterTool(multiCapTool)
			if err != nil {
				t.Fatalf("Failed to register tool: %v", err)
			}
			mockRuntime.registry = registry
			mockRuntime.execPolicy = &policy.ExecPolicy{
				Context: policy.ContextNormal,
				Allow:   []string{"tool.cloud.deployapp"},
				Grants:  capability.GrantSet{Grants: tc.grantedCaps},
			}

			_, execErr := registry.ExecuteTool("tool.cloud.deployApp", map[string]lang.Value{})

			if tc.shouldFail && execErr == nil {
				t.Error("Expected tool call to fail but it succeeded.")
			}
			if !tc.shouldFail && execErr != nil {
				t.Errorf("Expected tool call to succeed but it failed: %v", execErr)
			}
		})
	}
}

// TestCallFromInterpreter_CapabilityCheck_VerbAndScopeMismatch verifies grant rejection.
func TestCallFromInterpreter_CapabilityCheck_VerbAndScopeMismatch(t *testing.T) {
	mockRuntime := &testRuntime{}
	registry := tool.NewToolRegistry(mockRuntime)
	_, err := registry.RegisterTool(secureTool)
	if err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}
	mockRuntime.registry = registry
	mockRuntime.execPolicy = &policy.ExecPolicy{Context: policy.ContextNormal, Allow: []string{"tool.fs.writefile"}}

	mockRuntime.execPolicy.Grants = capability.GrantSet{
		Grants: []capability.Capability{{Resource: "fs", Verbs: []string{"read"}, Scopes: []string{"/data/*"}}},
	}

	_, execErr := registry.ExecuteTool("tool.fs.writeFile", map[string]lang.Value{})
	if execErr == nil {
		t.Fatal("Expected failure for wrong verb, but call succeeded.")
	}

	mockRuntime.execPolicy.Grants = capability.GrantSet{
		Grants: []capability.Capability{{Resource: "fs", Verbs: []string{"write"}, Scopes: []string{"/tmp/*"}}},
	}

	_, execErr = registry.ExecuteTool("tool.fs.writeFile", map[string]lang.Value{})
	if execErr == nil {
		t.Fatal("Expected failure for wrong scope, but call succeeded.")
	}
}

// TestCallFromInterpreter_CapabilityCheck_CaseInsensitive verifies case-insensitivity.
func TestCallFromInterpreter_CapabilityCheck_CaseInsensitive(t *testing.T) {
	mockRuntime := &testRuntime{}
	registry := tool.NewToolRegistry(mockRuntime)
	_, err := registry.RegisterTool(secureTool)
	if err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}
	mockRuntime.registry = registry
	mockRuntime.execPolicy = &policy.ExecPolicy{
		Context: policy.ContextNormal,
		Allow:   []string{"tool.fs.writefile"},
		Grants: capability.GrantSet{
			Grants: []capability.Capability{{Resource: "FS", Verbs: []string{"WRITE"}, Scopes: []string{"/data/*"}}},
		},
	}

	_, execErr := registry.ExecuteTool("tool.fs.writeFile", map[string]lang.Value{})
	if execErr != nil {
		t.Errorf("Expected call to succeed with case-insensitive grant, but it failed: %v", execErr)
	}
}
