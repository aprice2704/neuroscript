// NeuroScript Version: 0.3.1
// File version: 0.0.6 // Rename ToolRegistry struct to toolRegistryImpl
// nlines: 96
// risk_rating: MEDIUM
// filename: pkg/core/tools_registry.go

package core

import (
	"fmt"
	"log" // Standard Go logging package for init-time/early phase logging
	"sync"
)

// globalToolImplementations holds tools registered via init() functions.
var (
	globalToolImplementations []ToolImplementation
	globalRegMutex            sync.Mutex
)

// AddToolImplementations allows tool packages to register their ToolImplementation specs.
func AddToolImplementations(impls ...ToolImplementation) {
	globalRegMutex.Lock()
	defer globalRegMutex.Unlock()
	globalToolImplementations = append(globalToolImplementations, impls...)
}

// toolRegistryImpl manages the available tools for an Interpreter instance.
// This is the concrete struct implementation. The ToolRegistry interface is in tools_types.go.
type toolRegistryImpl struct {
	tools       map[string]ToolImplementation
	interpreter *Interpreter // Reference back to the interpreter
	mu          sync.RWMutex
}

// NewToolRegistry creates a new registry (toolRegistryImpl instance) associated with an interpreter.
// It processes the globalToolImplementations collected during init phases.
func NewToolRegistry(interpreter *Interpreter) *toolRegistryImpl { // Returns the concrete type
	r := &toolRegistryImpl{
		tools:       make(map[string]ToolImplementation),
		interpreter: interpreter,
	}

	globalRegMutex.Lock()
	toolsToProcess := make([]ToolImplementation, len(globalToolImplementations))
	copy(toolsToProcess, globalToolImplementations)
	globalRegMutex.Unlock()

	for _, impl := range toolsToProcess {
		if err := r.RegisterTool(impl); err != nil {
			logMsg := fmt.Sprintf("NewToolRegistry: Failed to register tool '%s' from global list: %v", impl.Spec.Name, err)
			if r.interpreter != nil && r.interpreter.logger != nil {
				r.interpreter.logger.Error(logMsg)
			} else {
				log.Printf("[ERROR] %s\n", logMsg)
			}
		}
	}
	return r
}

// RegisterTool adds or updates a tool in the registry.
func (r *toolRegistryImpl) RegisterTool(impl ToolImplementation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if impl.Spec.Name == "" {
		err := fmt.Errorf("attempted to register tool with empty name")
		if r.interpreter != nil && r.interpreter.logger != nil {
			r.interpreter.logger.Error("[toolRegistryImpl] Registration failed", "error", err.Error())
		} else {
			log.Printf("[ERROR] toolRegistryImpl: Registration failed: %v\n", err)
		}
		return err
	}
	if impl.Func == nil {
		err := fmt.Errorf("attempted to register tool '%s' with nil function", impl.Spec.Name)
		if r.interpreter != nil && r.interpreter.logger != nil {
			r.interpreter.logger.Error("[toolRegistryImpl] Registration failed", "tool_name", impl.Spec.Name, "error", err.Error())
		} else {
			log.Printf("[ERROR] toolRegistryImpl: Registration failed for tool '%s': %v\n", impl.Spec.Name, err)
		}
		return err
	}

	if _, exists := r.tools[impl.Spec.Name]; exists {
		logMsg := fmt.Sprintf("toolRegistryImpl: Attempted to re-register tool '%s'. First registration wins.", impl.Spec.Name)
		if r.interpreter != nil && r.interpreter.logger != nil {
			r.interpreter.logger.Warn(logMsg)
		} else {
			log.Printf("[WARN] %s\n", logMsg)
		}
		return nil
	}

	r.tools[impl.Spec.Name] = impl
	return nil
}

// GetTool retrieves a tool implementation by name.
func (r *toolRegistryImpl) GetTool(name string) (ToolImplementation, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, found := r.tools[name]
	return tool, found
}

// ListTools returns a list of specifications for all registered tools.
func (r *toolRegistryImpl) ListTools() []ToolSpec {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]ToolSpec, 0, len(r.tools))
	for _, impl := range r.tools {
		list = append(list, impl.Spec)
	}
	return list
}
