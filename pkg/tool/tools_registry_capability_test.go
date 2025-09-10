// NeuroScript Version: 0.4.0
// File version: 4
// Purpose: Resolved import cycle by defining a local test mock for the tool.Runtime interface.
// filename: pkg/tool/tools_registry_capability_test.go
// nlines: 151
// risk_rating: MEDIUM

package tool_test

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// testRuntime is a minimal mock implementation of tool.Runtime for this test file.
// Defining it locally avoids the circular dependency with the pkg/tool/internal package.
type testRuntime struct {
	registry tool.ToolRegistry
	grantSet *capability.GrantSet
}

func (t *testRuntime) GetGrantSet() *capability.GrantSet { return t.grantSet }
func (t *testRuntime) ToolRegistry() tool.ToolRegistry   { return t.registry }

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
	// 1. Setup
	mockRuntime := &testRuntime{}
	registry := tool.NewToolRegistry(mockRuntime)
	_, err := registry.RegisterTool(secureTool)
	if err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}
	mockRuntime.registry = registry

	// 2. Configure the mock runtime with a policy that GRANTS the required capability.
	grantedCaps := []capability.Capability{
		{Resource: "fs", Verbs: []string{"write"}, Scopes: []string{"/data/*"}}, // Grant with wildcard
	}
	mockRuntime.grantSet = &capability.GrantSet{
		Grants: grantedCaps,
	}

	// 3. Execute
	_, execErr := registry.ExecuteTool("tool.fs.writeFile", map[string]lang.Value{})

	// 4. Assert
	if execErr != nil {
		t.Errorf("Expected tool call to succeed, but it failed with error: %v", execErr)
	}
}

// TestCallFromInterpreter_CapabilityCheck_Failure verifies that a tool call is
// blocked with a PermissionDenied error when the required capabilities are not granted.
func TestCallFromInterpreter_CapabilityCheck_Failure(t *testing.T) {
	// 1. Setup
	mockRuntime := &testRuntime{}
	registry := tool.NewToolRegistry(mockRuntime)
	_, err := registry.RegisterTool(secureTool)
	if err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}
	mockRuntime.registry = registry

	// 2. Configure a policy that DOES NOT grant the required capability.
	grantedCaps := []capability.Capability{
		{Resource: "fs", Verbs: []string{"read"}, Scopes: []string{"*"}},
	}
	mockRuntime.grantSet = &capability.GrantSet{
		Grants: grantedCaps,
	}

	// 3. Execute
	_, execErr := registry.ExecuteTool("tool.fs.writeFile", map[string]lang.Value{})

	// 4. Assert
	if execErr == nil {
		t.Fatal("Expected tool call to fail due to insufficient capabilities, but it succeeded.")
	}
	var rtErr *lang.RuntimeError
	if !errors.As(execErr, &rtErr) || rtErr.Code != lang.ErrorCodePermissionDenied {
		t.Errorf("Expected a permission denied error, but got: %v", execErr)
	}
}

// TestCallFromInterpreter_CapabilityCheck_NoRequirements confirms that a tool
// with no capability requirements can be called successfully, even with an empty policy.
func TestCallFromInterpreter_CapabilityCheck_NoRequirements(t *testing.T) {
	// 1. Setup
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

	// 2. Policy is empty (default).
	mockRuntime.grantSet = &capability.GrantSet{}

	// 3. Execute
	result, execErr := registry.ExecuteTool("tool.math.add", map[string]lang.Value{})

	// 4. Assert
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
		{
			name: "Success - All capabilities granted",
			grantedCaps: []capability.Capability{
				{Resource: "net", Verbs: []string{"write"}, Scopes: []string{"*"}},
				{Resource: "fs", Verbs: []string{"read"}, Scopes: []string{"/src/app/**"}},
			},
			shouldFail: false,
		},
		{
			name: "Failure - Missing one capability",
			grantedCaps: []capability.Capability{
				{Resource: "net", Verbs: []string{"write"}, Scopes: []string{"*"}},
			},
			shouldFail: true,
		},
		{
			name:        "Failure - No capabilities granted",
			grantedCaps: []capability.Capability{},
			shouldFail:  true,
		},
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
			mockRuntime.grantSet = &capability.GrantSet{Grants: tc.grantedCaps}

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

// TestCallFromInterpreter_CapabilityCheck_VerbAndScopeMismatch verifies that a
// grant with the correct resource but wrong verb or scope is rejected.
func TestCallFromInterpreter_CapabilityCheck_VerbAndScopeMismatch(t *testing.T) {
	// 1. Setup
	mockRuntime := &testRuntime{}
	registry := tool.NewToolRegistry(mockRuntime)
	_, err := registry.RegisterTool(secureTool) // Needs fs:write:/data/user.txt
	if err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}
	mockRuntime.registry = registry

	// 2. Configure a policy that grants READ access, not WRITE.
	mockRuntime.grantSet = &capability.GrantSet{
		Grants: []capability.Capability{{Resource: "fs", Verbs: []string{"read"}, Scopes: []string{"/data/*"}}},
	}

	// 3. Execute and assert failure for wrong verb.
	_, execErr := registry.ExecuteTool("tool.fs.writeFile", map[string]lang.Value{})
	if execErr == nil {
		t.Fatal("Expected failure for wrong verb, but call succeeded.")
	}

	// 4. Configure a policy that grants access to a different directory.
	mockRuntime.grantSet = &capability.GrantSet{
		Grants: []capability.Capability{{Resource: "fs", Verbs: []string{"write"}, Scopes: []string{"/tmp/*"}}},
	}

	// 5. Execute and assert failure for wrong scope.
	_, execErr = registry.ExecuteTool("tool.fs.writeFile", map[string]lang.Value{})
	if execErr == nil {
		t.Fatal("Expected failure for wrong scope, but call succeeded.")
	}
}

// TestCallFromInterpreter_CapabilityCheck_CaseInsensitive verifies that resource
// and verb matching is case-insensitive.
func TestCallFromInterpreter_CapabilityCheck_CaseInsensitive(t *testing.T) {
	// 1. Setup
	mockRuntime := &testRuntime{}
	registry := tool.NewToolRegistry(mockRuntime)
	_, err := registry.RegisterTool(secureTool) // Needs fs:write
	if err != nil {
		t.Fatalf("Failed to register tool: %v", err)
	}
	mockRuntime.registry = registry

	// 2. Grant the capability using different casing.
	mockRuntime.grantSet = &capability.GrantSet{
		Grants: []capability.Capability{{Resource: "FS", Verbs: []string{"WRITE"}, Scopes: []string{"/data/*"}}},
	}

	// 3. Execute and assert success.
	_, execErr := registry.ExecuteTool("tool.fs.writeFile", map[string]lang.Value{})
	if execErr != nil {
		t.Errorf("Expected call to succeed with case-insensitive grant, but it failed: %v", execErr)
	}
}
