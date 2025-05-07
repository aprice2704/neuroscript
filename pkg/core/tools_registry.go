// NeuroScript Version: 0.3.1
// File version: 0.0.4 // Removed DEBUG log for successful first-time tool registration.
// filename: pkg/core/tools_registry.go

package core

import (
	"fmt"
	"log" // Standard Go logging package for init-time/early phase logging
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

// ToolRegistry manages the available tools for an Interpreter instance.
type ToolRegistry struct {
	tools       map[string]ToolImplementation
	interpreter *Interpreter // Reference back to the interpreter for its logger, if available
	mu          sync.RWMutex
}

// NewToolRegistry creates a new registry associated with an interpreter.
// It processes the globalToolImplementations collected during init phases.
func NewToolRegistry(interpreter *Interpreter) *ToolRegistry {
	r := &ToolRegistry{
		tools:       make(map[string]ToolImplementation),
		interpreter: interpreter,
	}

	globalRegMutex.Lock()
	toolsToProcess := make([]ToolImplementation, len(globalToolImplementations))
	copy(toolsToProcess, globalToolImplementations)
	globalRegMutex.Unlock()

	for _, impl := range toolsToProcess {
		if err := r.RegisterTool(impl); err != nil {
			// Log critical errors (e.g., nil func, empty name)
			log.Printf("[ERROR] NewToolRegistry: Failed to register tool '%s' from global list: %v\n", impl.Spec.Name, err)
		}
	}
	return r
}

// RegisterTool adds or updates a tool in the registry.
// If a tool with the same name already exists, the first registration wins,
// an error is logged, and no error is returned from this function for that case.
// Returns an error for other issues like empty name or nil function.
func (r *ToolRegistry) RegisterTool(impl ToolImplementation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if impl.Spec.Name == "" {
		err := fmt.Errorf("attempted to register tool with empty name")
		// Use interpreter's logger if available, otherwise standard log
		if r.interpreter != nil && r.interpreter.logger != nil {
			r.interpreter.logger.Error("[ToolRegistry] Registration failed", "error", err.Error())
		} else {
			log.Printf("[ERROR] ToolRegistry: Registration failed: %v\n", err)
		}
		return err // Return error for this critical issue
	}
	if impl.Func == nil {
		err := fmt.Errorf("attempted to register tool '%s' with nil function", impl.Spec.Name)
		if r.interpreter != nil && r.interpreter.logger != nil {
			r.interpreter.logger.Error("[ToolRegistry] Registration failed", "tool_name", impl.Spec.Name, "error", err.Error())
		} else {
			log.Printf("[ERROR] ToolRegistry: Registration failed for tool '%s': %v\n", impl.Spec.Name, err)
		}
		return err // Return error for this critical issue
	}

	// Check for duplicate registration
	if _, exists := r.tools[impl.Spec.Name]; exists {
		log.Printf("[ERROR] ToolRegistry: Attempted to re-register tool '%s'. First registration wins.\n", impl.Spec.Name)
		return nil // Do not overwrite, do not return an error from this function for this case
	}

	// Register the new tool
	r.tools[impl.Spec.Name] = impl
	// DEBUG log for successful registration has been removed as per request.
	// The summary logs from zz_core_tools_registrar.go and error logs for duplicates remain.
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
	return list
}
