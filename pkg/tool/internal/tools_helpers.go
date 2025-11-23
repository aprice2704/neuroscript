// NeuroScript Version: 0.6.0
// File version: 6.0.5
// Purpose: Aligned mock implementations with interface changes (string instead of types.AgentModelName). Implemented HandleRegistry for tool.Runtime interface compliance.
// filename: pkg/tool/internal/tools_helpers.go
// nlines: 297
// risk_rating: MEDIUM

package internal

import (
	"bytes"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"sync"

	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func MakeArgs(vals ...interface{}) []interface{} {
	if vals == nil {
		return []interface{}{}
	}
	return vals
}

func toolExec(interpreter tool.Runtime, cmdAndArgs ...string) (string, error) {
	if len(cmdAndArgs) == 0 {
		return "", fmt.Errorf("toolExec requires at least a command")
	}
	commandPath := cmdAndArgs[0]
	commandArgs := cmdAndArgs[1:]

	if strings.Contains(commandPath, "..") || strings.ContainsAny(commandPath, "|;&$><`\\") {
		errMsg := fmt.Sprintf("toolExec blocked suspicious command path: %q", commandPath)
		if logger := interpreter.GetLogger(); logger != nil {
			logger.Errorf("[toolExec] %s", errMsg)
		}
		return errMsg, fmt.Errorf("%w: %s", lang.ErrInternalTool, errMsg)
	}

	if logger := interpreter.GetLogger(); logger != nil {
		logArgs := make([]string, len(commandArgs))
		for i, arg := range commandArgs {
			if strings.Contains(arg, " ") {
				logArgs[i] = fmt.Sprintf("%q", arg)
			} else {
				logArgs[i] = arg
			}
		}
		logger.Debugf("[toolExec] Executing: %s %s", commandPath, strings.Join(logArgs, " "))
	}

	cmd := exec.Command(commandPath, commandArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	execErr := cmd.Run()

	stdoutStr := stdout.String()
	stderrStr := stderr.String()
	combinedOutput := stdoutStr + stderrStr

	if execErr != nil {
		errMsg := fmt.Sprintf("command '%s %s' failed with exit error: %v. Output:\n%s",
			commandPath, strings.Join(commandArgs, " "), execErr, combinedOutput)
		if logger := interpreter.GetLogger(); logger != nil {
			logger.Errorf("[toolExec] %s", errMsg)
		}
		return combinedOutput, fmt.Errorf("%w: %s", lang.ErrInternalTool, errMsg)
	}

	if logger := interpreter.GetLogger(); logger != nil {
		logger.Debugf("[toolExec] Command successful. Output:\n%s", combinedOutput)
	}
	return combinedOutput, nil
}

func getStringArg(args map[string]interface{}, key string) (string, error) {
	val, ok := args[key]
	if !ok {
		return "", fmt.Errorf("missing required argument '%s'", key)
	}
	strVal, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("invalid type for argument '%s': expected string, got %T", key, val)
	}
	return strVal, nil
}

func makeArgMap(kvPairs ...interface{}) (map[string]interface{}, error) {
	if len(kvPairs)%2 != 0 {
		return nil, fmt.Errorf("makeArgMap requires an even number of arguments (key-value pairs)")
	}
	args := make(map[string]interface{})
	for i := 0; i < len(kvPairs); i += 2 {
		key, ok := kvPairs[i].(string)
		if !ok {
			return nil, fmt.Errorf("makeArgMap requires string keys, got %T at index %d", kvPairs[i], i)
		}
		args[key] = kvPairs[i+1]
	}
	return args, nil
}

// --- Mock Handle Value/Registry for Testing ---

// mockHandleValue implements interfaces.HandleValue for testing only.
type mockHandleValue struct {
	id   string
	kind string
}

func (m mockHandleValue) Type() interfaces.NeuroScriptType { return lang.TypeHandle }
func (m mockHandleValue) String() string                   { return fmt.Sprintf("<handle %s#%s>", m.kind, m.id) }
func (m mockHandleValue) IsTruthy() bool                   { return true }
func (m mockHandleValue) HandleID() string                 { return m.id }
func (m mockHandleValue) HandleKind() string               { return m.kind }

// mockHandleRegistry implements interfaces.HandleRegistry, wrapping MockRuntime's internal handle map.
type mockHandleRegistry struct {
	rt *MockRuntime
}

func (m *mockHandleRegistry) NewHandle(payload any, kind string) (interfaces.HandleValue, error) {
	m.rt.mu.Lock()
	defer m.rt.mu.Unlock()

	if kind == "" {
		return nil, lang.ErrInvalidArgument // Mocking NewHandle requirement
	}

	m.rt.handleCounter++
	// Note: We use the existing MockRuntime handle ID format for compatibility with old test logic
	handleID := fmt.Sprintf("%s-%d", kind, m.rt.handleCounter)
	m.rt.Handles[handleID] = payload
	return mockHandleValue{id: handleID, kind: kind}, nil
}

func (m *mockHandleRegistry) GetHandle(id string) (any, error) {
	m.rt.mu.RLock()
	defer m.rt.mu.RUnlock()
	val, ok := m.rt.Handles[id]
	if !ok {
		return nil, lang.ErrHandleNotFound
	}
	return val, nil
}

func (m *mockHandleRegistry) DeleteHandle(id string) error {
	m.rt.mu.Lock()
	defer m.rt.mu.Unlock()
	if _, exists := m.rt.Handles[id]; !exists {
		return lang.ErrHandleNotFound
	}
	delete(m.rt.Handles, id)
	return nil
}

// --- Mock Runtime for Testing ---

type MockRuntime struct {
	mu             sync.RWMutex
	Vars           map[string]interface{}
	Output         *bytes.Buffer
	Handles        map[string]interface{} // Internal handle map
	handleCounter  int
	Models         map[types.AgentModelName]types.AgentModel
	SandboxDirStr  string
	PromptResponse string
	PromptErr      error
	Logger         interfaces.Logger
	LlmClient      interfaces.LLMClient
	Registry       any // Changed to `any` to break import cycle in tests.
	GrantSet       *capability.GrantSet
	ExecPolicy     *policy.ExecPolicy
}

// Statically assert that *MockRuntime satisfies the tool.Runtime interface.
var _ tool.Runtime = (*MockRuntime)(nil)

func NewMockRuntime() *MockRuntime {
	return &MockRuntime{
		Vars:          make(map[string]interface{}),
		Output:        new(bytes.Buffer),
		Handles:       make(map[string]interface{}),
		Models:        make(map[types.AgentModelName]types.AgentModel),
		SandboxDirStr: "/tmp/sandbox",
	}
}

// --- tool.Runtime Interface Implementation ---

func (m *MockRuntime) Println(a ...any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	fmt.Fprintln(m.Output, a...)
}

func (m *MockRuntime) PromptUser(prompt string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.PromptResponse, m.PromptErr
}

func (m *MockRuntime) GetVar(name string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.Vars[name]
	return val, ok
}

func (m *MockRuntime) SetVar(name string, val any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Vars[name] = val
}

func (m *MockRuntime) CallTool(name types.FullName, args []any) (any, error) {
	impl, ok := m.ToolRegistry().GetTool(name)
	if !ok {
		return nil, fmt.Errorf("tool '%s' not found in mock registry", name)
	}
	return impl.Func(m, args)
}

func (m *MockRuntime) GetLogger() interfaces.Logger {
	return m.Logger
}

func (m *MockRuntime) SandboxDir() string {
	return m.SandboxDirStr
}

func (m *MockRuntime) ToolRegistry() tool.ToolRegistry {
	if m.Registry == nil {
		return nil
	}
	// This type assertion resolves the compiler's circular dependency confusion.
	return m.Registry.(tool.ToolRegistry)
}

func (m *MockRuntime) LLM() interfaces.LLMClient {
	return m.LlmClient
}

// HandleRegistry returns the mock HandleRegistry that wraps the internal state.
func (m *MockRuntime) HandleRegistry() interfaces.HandleRegistry {
	return &mockHandleRegistry{rt: m}
}

// NOTE: The old RegisterHandle and GetHandleValue methods were removed to comply with the new tool.Runtime interface.

func (m *MockRuntime) GetGrantSet() *capability.GrantSet {
	if m.GrantSet != nil {
		return m.GrantSet
	}
	// Return a default GrantSet that allows tools without specific requirements to run.
	return &capability.GrantSet{}
}

// GetExecPolicy returns the currently active execution policy.
func (m *MockRuntime) GetExecPolicy() *policy.ExecPolicy {
	return m.ExecPolicy
}

// --- AgentModel Management ---

func (m *MockRuntime) AgentModels() interfaces.AgentModelReader {
	return &mockModelReader{rt: m}
}

func (m *MockRuntime) AgentModelsAdmin() interfaces.AgentModelAdmin {
	return &mockModelAdmin{rt: m}
}

// --- Mock AgentModel Reader/Admin Implementations ---

type mockModelReader struct {
	rt *MockRuntime
}

// FIX: Signature changed to return []string
func (v *mockModelReader) List() []string {
	v.rt.mu.RLock()
	defer v.rt.mu.RUnlock()
	// FIX: Return []string
	out := make([]string, 0, len(v.rt.Models))
	for name := range v.rt.Models {
		// FIX: Convert name to string
		out = append(out, string(name))
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// FIX: Signature changed to accept string
func (v *mockModelReader) Get(name string) (any, bool) {
	v.rt.mu.RLock()
	defer v.rt.mu.RUnlock()
	// FIX: Convert string name back to types.AgentModelName for map lookup
	model, ok := v.rt.Models[types.AgentModelName(name)]
	return model, ok
}

type mockModelAdmin struct {
	rt *MockRuntime
}

// FIX: Signature changed to return []string
func (v *mockModelAdmin) List() []string {
	return v.rt.AgentModels().List()
}

// FIX: Signature changed to accept string
func (v *mockModelAdmin) Get(name string) (any, bool) {
	return v.rt.AgentModels().Get(name)
}

// FIX: Signature changed to accept string
func (v *mockModelAdmin) Register(name string, cfg map[string]any) error {
	v.rt.mu.Lock()
	defer v.rt.mu.Unlock()
	// FIX: Convert string name to types.AgentModelName for map operations
	key := types.AgentModelName(name)
	if _, exists := v.rt.Models[key]; exists {
		return lang.ErrDuplicateKey
	}
	model := types.AgentModel{Name: key}
	if p, ok := cfg["provider"].(string); ok {
		model.Provider = p
	}
	if m, ok := cfg["model"].(string); ok {
		model.Model = m
	}
	v.rt.Models[key] = model
	return nil
}

// FIX: Added missing RegisterFromModel to satisfy the interface
func (v *mockModelAdmin) RegisterFromModel(model any) error {
	modelStruct, ok := model.(types.AgentModel)
	if !ok {
		return fmt.Errorf("mockModelAdmin: invalid type for RegisterFromModel, expected types.AgentModel")
	}
	return v.Register(string(modelStruct.Name), nil) // Simple mock
}

// FIX: Signature changed to accept string
func (v *mockModelAdmin) Update(name string, updates map[string]any) error {
	v.rt.mu.Lock()
	defer v.rt.mu.Unlock()
	// FIX: Convert string name to types.AgentModelName for map operations
	key := types.AgentModelName(name)
	model, exists := v.rt.Models[key]
	if !exists {
		return lang.ErrNotFound
	}
	if p, ok := updates["provider"].(string); ok {
		model.Provider = p
	}
	v.rt.Models[key] = model
	return nil
}

// FIX: Signature changed to accept string
func (v *mockModelAdmin) Delete(name string) bool {
	v.rt.mu.Lock()
	defer v.rt.mu.Unlock()
	// FIX: Convert string name to types.AgentModelName for map operations
	key := types.AgentModelName(name)
	if _, exists := v.rt.Models[key]; exists {
		delete(v.rt.Models, key)
		return true
	}
	return false
}
