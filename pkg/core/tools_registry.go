// neuroscript/pkg/core/tools_registry.go
package core

import "fmt"

// ToolRegistry holds the collection of registered tools.
type ToolRegistry struct {
	tools map[string]ToolImplementation
}

// NewToolRegistry creates a new, empty ToolRegistry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{tools: make(map[string]ToolImplementation)}
}

// RegisterTool adds a tool implementation to the registry.
// Returns an error if a tool with the same name already exists or if the implementation is nil.
func (tr *ToolRegistry) RegisterTool(impl ToolImplementation) error {
	if _, exists := tr.tools[impl.Spec.Name]; exists {
		return fmt.Errorf("tool '%s' already registered", impl.Spec.Name)
	}
	if impl.Func == nil {
		return fmt.Errorf("tool '%s' registration is missing implementation function", impl.Spec.Name)
	}
	tr.tools[impl.Spec.Name] = impl
	return nil
}

// GetTool retrieves a tool implementation by name.
// Returns the implementation and a boolean indicating if the tool was found.
func (tr *ToolRegistry) GetTool(name string) (ToolImplementation, bool) {
	impl, found := tr.tools[name]
	return impl, found
}
