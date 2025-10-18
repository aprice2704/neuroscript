// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Prevents overwriting existing tool registrations; returns an error instead.
// filename: pkg/tool/tools_registration.go
// nlines: 100+
// risk_rating: MEDIUM

package tool

import (
	"crypto/sha256"
	"fmt"
	"os" // Import os for Fprintf
	"sync"

	"github.com/aprice2704/neuroscript/pkg/lang" // Import lang for error
	"github.com/aprice2704/neuroscript/pkg/types"
)

// ToolRegistryImpl manages the available tools for an Interpreter instance.
type ToolRegistryImpl struct {
	tools       map[types.FullName]ToolImplementation
	interpreter Runtime // This should be the public *api.Interpreter
	mu          *sync.RWMutex
}

// NewToolRegistry creates a new, empty registry instance.
func NewToolRegistry(interpreter Runtime) *ToolRegistryImpl {
	r := &ToolRegistryImpl{
		tools:       make(map[types.FullName]ToolImplementation),
		interpreter: interpreter,
		mu:          &sync.RWMutex{},
	}
	return r
}

// calculateChecksum generates a stable hash of a tool's essential signature.
// It uses FullName, ReturnType, and number of args.
func calculateChecksum(spec ToolSpec) string {
	data := fmt.Sprintf("%s:%s:%d", spec.FullName, spec.ReturnType, len(spec.Args))
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("sha256:%x", hash)
}

// RegisterTool adds a tool to the registry. It canonicalizes the tool's name
// and automatically calculates its integrity checksum. It returns an error if
// a tool with the same canonical name already exists.
func (r *ToolRegistryImpl) RegisterTool(impl ToolImplementation) (ToolImplementation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if impl.Spec.Name == "" {
		return impl, fmt.Errorf("tool registration failed: name is empty")
	}
	if impl.Func == nil {
		// Log this critical issue, as it's a developer error during setup.
		err := fmt.Errorf("tool registration failed for '%s.%s': function is nil", impl.Spec.Group, impl.Spec.Name)
		// Use logger if available, otherwise stderr
		if r.interpreter != nil && r.interpreter.GetLogger() != nil {
			r.interpreter.GetLogger().Error(err.Error())
		} else {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		}
		// Return error even after logging
		return impl, err
	}

	baseName := string(impl.Spec.Group) + "." + string(impl.Spec.Name)
	canonicalName := CanonicalizeToolName(baseName)
	fullName := types.FullName(canonicalName)

	impl.FullName = fullName
	impl.Spec.FullName = fullName // Ensure spec also has canonical name
	impl.SignatureChecksum = calculateChecksum(impl.Spec)

	// <<< ADDED OVERWRITE CHECK >>>
	if existingImpl, exists := r.tools[fullName]; exists {
		errMsg := fmt.Sprintf("tool '%s' already registered", fullName)
		// Provide more context in the error if possible (e.g., existing func pointer)
		fmt.Fprintf(os.Stderr, "[DEBUG][RegisterTool] ERROR: Attempted to overwrite tool '%s'. Existing Func: %p, New Func: %p\n", fullName, existingImpl.Func, impl.Func) // DEBUG
		return impl, fmt.Errorf("%w: %s", lang.ErrDuplicateKey, errMsg)                                                                                                    // Use lang.ErrDuplicateKey
	}
	// <<< END OVERWRITE CHECK >>>

	r.tools[fullName] = impl
	//fmt.Fprintf(os.Stderr, "[DEBUG][RegisterTool] Registered tool '%s' with Func: %p\n", fullName, impl.Func) // DEBUG

	return impl, nil
}

// GetTool finds a tool by its fully qualified (and potentially non-canonical) name.
func (r *ToolRegistryImpl) GetTool(name types.FullName) (ToolImplementation, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	canonicalName := CanonicalizeToolName(string(name))
	tool, found := r.tools[types.FullName(canonicalName)]
	return tool, found
}

// GetToolShort finds a tool using its group and short name.
func (r *ToolRegistryImpl) GetToolShort(group types.ToolGroup, name types.ToolName) (ToolImplementation, bool) {
	baseName := types.FullName(string(group) + "." + string(name))
	return r.GetTool(baseName) // Relies on GetTool's canonicalization
}

// ListTools returns a slice of all registered tool implementations.
func (r *ToolRegistryImpl) ListTools() []ToolImplementation {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]ToolImplementation, 0, len(r.tools))
	for _, impl := range r.tools {
		list = append(list, impl)
	}
	return list
}

// NTools returns the number of tools currently registered.
func (r *ToolRegistryImpl) NTools() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}

// NewViewForInterpreter creates a new registry instance that shares the toolset
// of the original registry but is bound to a new interpreter runtime context.
// This is useful for creating isolated execution environments (like forks).
func (r *ToolRegistryImpl) NewViewForInterpreter(interpreter Runtime) ToolRegistry {
	// Returns a new struct sharing the mutex and tools map, but with its own
	// interpreter reference.
	return &ToolRegistryImpl{
		tools:       r.tools,     // Shared map (read-only after init)
		interpreter: interpreter, // Specific to this view
		mu:          r.mu,        // Shared mutex
	}
}
