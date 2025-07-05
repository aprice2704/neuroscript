// NeuroScript Version: 0.4.0
// File version: 1.1.0
// Purpose: Corrected typos for the Runtime interface.
// filename: pkg/tool/tools_registry.go
// nlines: 165
// risk_rating: HIGH

package tool

import (
	"fmt"
	"log"
	"sync"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/utils"
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

// ToolRegistryImpl manages the available tools for an Interpreter instance.
type ToolRegistryImpl struct {
	tools map[string]ToolImplementation
	// FIX: Use the correct 'Runtime' interface name.
	interpreter Runtime
	mu          sync.RWMutex
}

// NewToolRegistry creates a new registry instance.
// FIX: Use the correct 'Runtime' interface name.
func NewToolRegistry(interpreter Runtime) *ToolRegistryImpl {
	r := &ToolRegistryImpl{
		tools:       make(map[string]ToolImplementation),
		interpreter: interpreter,
	}
	globalRegMutex.Lock()
	defer globalRegMutex.Unlock()
	for _, impl := range globalToolImplementations {
		if err := r.RegisterTool(impl); err != nil {
			log.Printf("[ERROR] NewToolRegistry: Failed to register tool '%s': %v\n", impl.Spec.Name, err)
		}
	}
	return r
}

// RegisterTool adds a tool to the registry.
func (r *ToolRegistryImpl) RegisterTool(impl ToolImplementation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if impl.Spec.Name == "" {
		return fmt.Errorf("tool registration failed: name is empty")
	}
	if impl.Func == nil {
		return fmt.Errorf("tool registration failed for '%s': function is nil", impl.Spec.Name)
	}
	r.tools[impl.Spec.Name] = impl
	return nil
}

// GetTool retrieves a tool by name.
func (r *ToolRegistryImpl) GetTool(name string) (ToolImplementation, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, found := r.tools[name]
	return tool, found
}

// ListTools returns the specs of all registered tools.
func (r *ToolRegistryImpl) ListTools() []ToolSpec {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]ToolSpec, 0, len(r.tools))
	for _, impl := range r.tools {
		list = append(list, impl.Spec)
	}
	return list
}

// --- BRIDGE IMPLEMENTATION ---

// CallFromInterpreter is the single bridge between the Value-based interpreter and primitive-based tools.
func (r *ToolRegistryImpl) CallFromInterpreter(interp Runtime, toolName string, args []lang.Value) (lang.Value, error) {
	impl, ok := r.GetTool(toolName)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeToolNotFound, fmt.Sprintf("tool '%s' not found", toolName), lang.ErrToolNotFound)
	}

	// 1. Unwrap all arguments from Value to primitives
	rawArgs := make([]interface{}, len(args))
	for i, arg := range args {
		// This function call needs to be verified against the lang package
		rawArgs[i] = lang.Unwrap(arg)
	}

	// 2. Validate and coerce the primitive arguments
	if len(rawArgs) < len(impl.Spec.Args) {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("tool '%s': expected at least %d args, got %d", toolName, len(impl.Spec.Args), len(rawArgs)), lang.ErrArgumentMismatch)
	}

	coercedArgs := make([]interface{}, len(impl.Spec.Args))
	for i, spec := range impl.Spec.Args {
		var err error
		coercedArgs[i], err = coerceArg(rawArgs[i], spec.Type)
		if err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("tool '%s' arg '%s': %v", toolName, spec.Name, err), lang.ErrArgumentMismatch)
		}
	}
	if impl.Spec.Variadic {
		coercedArgs = append(coercedArgs, rawArgs[len(impl.Spec.Args):]...)
	}

	// 3. Call the tool's implementation function with primitives
	out, err := impl.Func(interp, coercedArgs)
	if err != nil {
		return nil, err // Assume tool returns a compliant RuntimeError
	}

	// 4. Wrap the primitive result back into a Value
	return lang.Wrap(out)
}

// ExecuteTool is the bridge for external callers that have named arguments.
func (r *ToolRegistryImpl) ExecuteTool(toolName string, args map[string]lang.Value) (lang.Value, error) {
	impl, ok := r.GetTool(toolName)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeToolNotFound, fmt.Sprintf("tool '%s' not found", toolName), lang.ErrToolNotFound)
	}

	// Tool functions require a `tool.Runtime`. The registry was initialized with one.
	if r.interpreter == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeConfiguration, "ToolRegistry not configured with a runtime context", lang.ErrConfiguration)
	}

	// Convert the map of named arguments into an ordered slice of positional arguments.
	orderedLangArgs := make([]lang.Value, len(impl.Spec.Args))
	for i, spec := range impl.Spec.Args {
		val, ok := args[spec.Name]
		if !ok {
			if spec.Required {
				return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("missing required argument '%s' for tool '%s'", spec.Name, toolName), lang.ErrArgumentMismatch)
			}
			orderedLangArgs[i] = lang.NilValue{} // Use nil for optional args that are not provided.
		} else {
			orderedLangArgs[i] = val
		}
	}

	// Now we have an ordered slice of `lang.Value`, which is what CallFromInterpreter expects.
	return r.CallFromInterpreter(r.interpreter, toolName, orderedLangArgs)
}

// coerceArg attempts to convert a primitive value `x` to the specified ArgType.
func coerceArg(x interface{}, t ArgType) (interface{}, error) {
	if x == nil {
		return nil, nil // Let the tool handle nil for optional args.
	}

	switch t {
	case ArgTypeString:
		s, ok := x.(string)
		if !ok {
			return nil, fmt.Errorf("expected string, got %T", x)
		}
		return s, nil
	case ArgTypeInt:
		i, ok := lang.ToInt64(x)
		if !ok {
			return nil, fmt.Errorf("expected integer, got %T", x)
		}
		return i, nil
	case ArgTypeFloat:
		f, ok := lang.ToFloat64(x)
		if !ok {
			return nil, fmt.Errorf("expected float, got %T", x)
		}
		return f, nil
	case ArgTypeBool:
		b, ok := utils.ConvertToBool(x)
		if !ok {
			return nil, fmt.Errorf("expected boolean, got %T", x)
		}
		return b, nil
	case ArgTypeSliceAny:
		s, ok, _ := utils.ConvertToSliceOfAny(x)
		if !ok {
			return nil, fmt.Errorf("expected list, got %T", x)
		}
		return s, nil
	case ArgTypeMap:
		m, ok := x.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("expected map, got %T", x)
		}
		return m, nil
	case ArgTypeAny:
		return x, nil
	default:
		return nil, fmt.Errorf("unknown argument type for coercion: %s", t)
	}
}
