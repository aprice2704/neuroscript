// NeuroScript Version: 0.4.0
// File version: 8
// Purpose: Removed the conflicting auto-registration loop from the constructor to centralize tool setup in the api.New function.
// filename: pkg/tool/tools_registry.go
// nlines: 147
// risk_rating: HIGH

package tool

import (
	"fmt"
	"log"
	"sync"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
	"github.com/aprice2704/neuroscript/pkg/utils"
)

// The global tool registration mechanism below is part of a conflicting pattern
// and is being deprecated in favor of the 'register.go' pattern. It is commented
// out to prevent its use.
/*
var (
	globalToolImplementations []ToolImplementation
	globalRegMutex            sync.Mutex
)

func AddToolImplementations(impls ...ToolImplementation) {
	globalRegMutex.Lock()
	defer globalRegMutex.Unlock()
	globalToolImplementations = append(globalToolImplementations, impls...)
}
*/

// ToolRegistryImpl manages the available tools for an Interpreter instance.
type ToolRegistryImpl struct {
	tools       map[types.FullName]ToolImplementation
	interpreter Runtime
	mu          sync.RWMutex
}

// NewToolRegistry creates a new, empty registry instance.
// Tool registration is now handled by the high-level api.New() function.
func NewToolRegistry(interpreter Runtime) *ToolRegistryImpl {
	r := &ToolRegistryImpl{
		tools:       make(map[types.FullName]ToolImplementation),
		interpreter: interpreter,
	}
	return r
}

// RegisterTool adds a tool to the registry. It canonicalizes the tool's name
// to ensure consistent storage and lookup.
func (r *ToolRegistryImpl) RegisterTool(impl ToolImplementation) (ToolImplementation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if impl.Spec.Name == "" {
		return impl, fmt.Errorf("tool registration failed: name is empty")
	}
	if impl.Func == nil {
		return impl, fmt.Errorf("tool registration failed for '%s': function is nil", impl.Spec.Name)
	}

	baseName := string(impl.Spec.Group) + "." + string(impl.Spec.Name)
	canonicalName := CanonicalizeToolName(baseName)

	impl.FullName = types.FullName(canonicalName)
	impl.Spec.FullName = types.FullName(canonicalName)

	log.Printf("[DEBUG] Registering tool. Group: '%s', Name: '%s', Final Key: '%s'", impl.Spec.Group, impl.Spec.Name, canonicalName)

	r.tools[types.FullName(canonicalName)] = impl

	return impl, nil
}

// GetTool finds a tool by name.
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
	return r.GetTool(baseName)
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
func (r *ToolRegistryImpl) CallFromInterpreter(interp Runtime, fullname types.FullName, args []lang.Value) (lang.Value, error) {
	impl, ok := r.GetTool(fullname)
	if !ok {
		canonicalName := CanonicalizeToolName(string(fullname))
		return nil, lang.NewRuntimeError(lang.ErrorCodeToolNotFound, fmt.Sprintf("tool '%s' not found", canonicalName), lang.ErrToolNotFound)
	}

	rawArgs := make([]interface{}, len(args))
	for i, arg := range args {
		rawArgs[i] = lang.Unwrap(arg)
	}

	if len(rawArgs) < len(impl.Spec.Args) {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("tool '%s': expected at least %d args, got %d", impl.FullName, len(impl.Spec.Args), len(rawArgs)), lang.ErrArgumentMismatch)
	}

	coercedArgs := make([]interface{}, len(impl.Spec.Args))
	for i, spec := range impl.Spec.Args {
		var err error
		coercedArgs[i], err = coerceArg(rawArgs[i], spec.Type)
		if err != nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("tool '%s' arg '%s': %v", impl.FullName, spec.Name, err), lang.ErrArgumentMismatch)
		}
	}
	if impl.Spec.Variadic {
		coercedArgs = append(coercedArgs, rawArgs[len(impl.Spec.Args):]...)
	}

	out, err := impl.Func(interp, coercedArgs)
	if err != nil {
		return nil, err
	}

	return lang.Wrap(out)
}

// Simple internal fn to return the number of tools registered
func (r *ToolRegistryImpl) NTools() (ntools int) {
	return len(r.tools)
}

// ExecuteTool is the bridge for external callers that have named arguments.
func (r *ToolRegistryImpl) ExecuteTool(fullname types.FullName, args map[string]lang.Value) (lang.Value, error) {
	impl, ok := r.GetTool(fullname)
	if !ok {
		canonicalName := CanonicalizeToolName(string(fullname))
		return nil, lang.NewRuntimeError(lang.ErrorCodeToolNotFound, fmt.Sprintf("tool '%s' not found", canonicalName), lang.ErrToolNotFound)
	}

	if r.interpreter == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeConfiguration, "ToolRegistry not configured with a runtime context", lang.ErrConfiguration)
	}

	orderedLangArgs := make([]lang.Value, len(impl.Spec.Args))
	for i, spec := range impl.Spec.Args {
		val, ok := args[spec.Name]
		if !ok {
			if spec.Required {
				return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("missing required argument '%s' for tool '%s'", spec.Name, impl.FullName), lang.ErrArgumentMismatch)
			}
			orderedLangArgs[i] = lang.NilValue{}
		} else {
			orderedLangArgs[i] = val
		}
	}

	return r.CallFromInterpreter(r.interpreter, fullname, orderedLangArgs)
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
