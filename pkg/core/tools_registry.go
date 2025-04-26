// filename: pkg/core/tools_registry.go
package core

import "fmt"

// ToolRegistry holds the collection of registered tools.
// NOTE: Tools are registered and retrieved using their BASE NAME (e.g., "ReadFile"),
// not the fully qualified name used in NeuroScript CALL syntax (e.g., "TOOL.ReadFile").
// The interpreter handles the "TOOL." prefix during CALL execution.
type ToolRegistry struct {
	tools map[string]ToolImplementation
}

// NewToolRegistry creates a new, empty ToolRegistry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{tools: make(map[string]ToolImplementation)}
}

var DefaultRegistry = NewToolRegistry()

// GlobalToolRegistry is a default registry instance.
// Deprecated: Avoid using global registry; pass registries explicitly.
var GlobalToolRegistry = NewToolRegistry()

// RegisterTool adds a tool implementation to the registry using its base name from the Spec as the key.
// Returns an error if a tool with the same name already exists or if the implementation is nil.
func (tr *ToolRegistry) RegisterTool(impl ToolImplementation) error {
	if impl.Spec.Name == "" {
		return fmt.Errorf("tool registration failed: ToolSpec.Name cannot be empty")
	}
	// Ensure the map is initialized (defensive check)
	if tr.tools == nil {
		tr.tools = make(map[string]ToolImplementation)
	}
	if _, exists := tr.tools[impl.Spec.Name]; exists {
		return fmt.Errorf("tool '%s' already registered", impl.Spec.Name)
	}
	if impl.Func == nil {
		return fmt.Errorf("tool '%s' registration is missing implementation function", impl.Spec.Name)
	}
	tr.tools[impl.Spec.Name] = impl
	return nil
}

// GetTool retrieves a tool implementation by its base name (e.g., "ReadFile").
// Returns the implementation and a boolean indicating if the tool was found.
func (tr *ToolRegistry) GetTool(name string) (ToolImplementation, bool) {
	// Ensure the map is initialized before access (defensive check)
	if tr.tools == nil {
		return ToolImplementation{}, false
	}
	// Lookup uses the base name directly.
	impl, found := tr.tools[name]
	return impl, found
}

// +++ ADDED: GetAllTools +++
// GetAllTools returns a copy of the internal map of all registered tools.
// The returned map's keys are the base tool names (e.g., "ReadFile").
func (tr *ToolRegistry) GetAllTools() map[string]ToolImplementation {
	// Ensure the map is initialized before access (defensive check)
	if tr.tools == nil {
		return make(map[string]ToolImplementation) // Return empty map if nil
	}
	// Return a copy to prevent modification of the internal map
	toolsCopy := make(map[string]ToolImplementation, len(tr.tools))
	for name, impl := range tr.tools {
		toolsCopy[name] = impl
	}
	return toolsCopy
}

// --- END ADDED ---
