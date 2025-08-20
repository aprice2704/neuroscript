// NeuroScript Version: 0.6.0
// File version: 6.0.2
// Purpose: Fixed a test-only import cycle issue by changing MockRuntime.Registry to type any and using a type assertion.
// filename: pkg/tool/internal/tools_helpers.go
// nlines: 247
// risk_rating: MEDIUM

package internal

import (
	"bytes"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"sync"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy/capability"
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

// --- Mock Runtime for Testing ---

type MockRuntime struct {
	mu             sync.RWMutex
	Vars           map[string]interface{}
	Output         *bytes.Buffer
	Handles        map[string]interface{}
	handleCounter  int
	Models         map[types.AgentModelName]types.AgentModel
	SandboxDirStr  string
	PromptResponse string
	PromptErr      error
	Logger         interfaces.Logger
	LlmClient      interfaces.LLMClient
	Registry       any // Changed to `any` to break import cycle in tests.
	GrantSet       *capability.GrantSet
}

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

func (m *MockRuntime) RegisterHandle(obj interface{}, typePrefix string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handleCounter++
	handle := fmt.Sprintf("%s-%d", typePrefix, m.handleCounter)
	m.Handles[handle] = obj
	return handle, nil
}

func (m *MockRuntime) GetHandleValue(handle string, expectedTypePrefix string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if !strings.HasPrefix(handle, expectedTypePrefix+"-") {
		return nil, fmt.Errorf("invalid handle prefix for %s: expected '%s'", handle, expectedTypePrefix)
	}
	val, ok := m.Handles[handle]
	if !ok {
		return nil, fmt.Errorf("handle not found: %s", handle)
	}
	return val, nil
}

func (m *MockRuntime) GetGrantSet() *capability.GrantSet {
	if m.GrantSet != nil {
		return m.GrantSet
	}
	// Return a default GrantSet that allows tools without specific requirements to run.
	return &capability.GrantSet{}
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

func (v *mockModelReader) List() []types.AgentModelName {
	v.rt.mu.RLock()
	defer v.rt.mu.RUnlock()
	out := make([]types.AgentModelName, 0, len(v.rt.Models))
	for name := range v.rt.Models {
		out = append(out, name)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func (v *mockModelReader) Get(name types.AgentModelName) (any, bool) {
	v.rt.mu.RLock()
	defer v.rt.mu.RUnlock()
	model, ok := v.rt.Models[name]
	return model, ok
}

type mockModelAdmin struct {
	rt *MockRuntime
}

func (v *mockModelAdmin) List() []types.AgentModelName {
	return v.rt.AgentModels().List()
}

func (v *mockModelAdmin) Get(name types.AgentModelName) (any, bool) {
	return v.rt.AgentModels().Get(name)
}

func (v *mockModelAdmin) Register(name types.AgentModelName, cfg map[string]any) error {
	v.rt.mu.Lock()
	defer v.rt.mu.Unlock()
	if _, exists := v.rt.Models[name]; exists {
		return lang.ErrDuplicateKey
	}
	model := types.AgentModel{Name: name}
	if p, ok := cfg["provider"].(string); ok {
		model.Provider = p
	}
	if m, ok := cfg["model"].(string); ok {
		model.Model = m
	}
	v.rt.Models[name] = model
	return nil
}

func (v *mockModelAdmin) Update(name types.AgentModelName, updates map[string]any) error {
	v.rt.mu.Lock()
	defer v.rt.mu.Unlock()
	model, exists := v.rt.Models[name]
	if !exists {
		return lang.ErrNotFound
	}
	if p, ok := updates["provider"].(string); ok {
		model.Provider = p
	}
	v.rt.Models[name] = model
	return nil
}

func (v *mockModelAdmin) Delete(name types.AgentModelName) bool {
	v.rt.mu.Lock()
	defer v.rt.mu.Unlock()
	if _, exists := v.rt.Models[name]; exists {
		delete(v.rt.Models, name)
		return true
	}
	return false
}
