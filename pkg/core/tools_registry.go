// NeuroScript Version: 0.3.1
// File version: 0.0.2 // Add central list and func for init-based registration.
// filename: pkg/core/tools_registry.go

package core

import (
	"fmt"
	"sync"
)

// globalToolImplementations holds tools registered via init() functions.
// This avoids import cycles between core and tool sub-packages.
var (
	globalToolImplementations []ToolImplementation
	globalRegMutex            sync.Mutex // Protect access during registration phase
)

// AddToolImplementations allows tool packages (like gosemantic, goast, toolsets)
// to register their ToolImplementation specs during their init() phase.
// This should typically only be called from init() functions.
func AddToolImplementations(impls ...ToolImplementation) {
	globalRegMutex.Lock()
	defer globalRegMutex.Unlock()
	globalToolImplementations = append(globalToolImplementations, impls...)
}

// --- ToolRegistry definition remains the same ---

// ToolRegistry manages the available tools for an Interpreter instance.
type ToolRegistry struct {
	tools       map[string]ToolImplementation
	interpreter *Interpreter // Reference back to the interpreter
	mu          sync.RWMutex
}

// NewToolRegistry creates a new registry associated with an interpreter.
func NewToolRegistry(interpreter *Interpreter) *ToolRegistry {
	return &ToolRegistry{
		tools:       make(map[string]ToolImplementation),
		interpreter: interpreter,
	}
}

// RegisterTool adds or updates a tool in the registry.
// It checks for naming conflicts.
func (r *ToolRegistry) RegisterTool(impl ToolImplementation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if impl.Spec.Name == "" {
		return fmt.Errorf("attempted to register tool with empty name")
	}
	if impl.Func == nil {
		return fmt.Errorf("attempted to register tool '%s' with nil function", impl.Spec.Name)
	}

	// Basic validation for spec could go here if needed

	// Allow overwriting for now, maybe add configuration later?
	// if _, exists := r.tools[impl.Spec.Name]; exists {
	//  return fmt.Errorf("tool '%s' already registered", impl.Spec.Name)
	// }

	r.tools[impl.Spec.Name] = impl
	// Add logging if interpreter and logger are available and configured
	if r.interpreter != nil && r.interpreter.logger != nil {
		r.interpreter.logger.Debug("Tool registered", "name", impl.Spec.Name)
	}
	return nil
}

// GetTool retrieves a tool implementation by name.
func (r *ToolRegistry) GetTool(name string) (ToolImplementation, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, found := r.tools[name]
	return tool, found
}

// ListTools returns a list of specifications for all registered tools.
func (r *ToolRegistry) ListTools() []ToolSpec {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]ToolSpec, 0, len(r.tools))
	for _, impl := range r.tools {
		list = append(list, impl.Spec)
	}
	// Sort list alphabetically by name?
	// sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
	return list
}
