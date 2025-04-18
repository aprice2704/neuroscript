// neuroscript/pkg/core/tools_registry.go
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

var (
	// GlobalToolRegistry is a default registry instance.
	// Consider initializing it more robustly if concurrent access during init is possible.
	GlobalToolRegistry = NewToolRegistry()
)

// RegisterTool adds a tool implementation to the registry using its base name from the Spec as the key.
// Returns an error if a tool with the same name already exists or if the implementation is nil.
func (tr *ToolRegistry) RegisterTool(impl ToolImplementation) error {
	if impl.Spec.Name == "" {
		return fmt.Errorf("tool registration failed: ToolSpec.Name cannot be empty")
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
	// Lookup uses the base name directly.
	impl, found := tr.tools[name]
	return impl, found
}
