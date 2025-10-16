// NeuroScript Version: 0.8.0
// File version: 22
// Purpose: Fixes compiler errors in ExecuteTool and simplifies its logic.
// filename: pkg/tool/tools_registry.go
// nlines: 201
// risk_rating: HIGH

package tool

import (
	"crypto/sha256"
	"fmt"
	"os"
	"sync"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
	"github.com/aprice2704/neuroscript/pkg/utils"
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
func calculateChecksum(spec ToolSpec) string {
	data := fmt.Sprintf("%s:%s:%d", spec.FullName, spec.ReturnType, len(spec.Args))
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("sha256:%x", hash)
}

// RegisterTool adds a tool to the registry. It canonicalizes the tool's name
// and automatically calculates its integrity checksum.
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
	impl.SignatureChecksum = calculateChecksum(impl.Spec)

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

// ListTools returns all registered tool implementations.
func (r *ToolRegistryImpl) ListTools() []ToolImplementation {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]ToolImplementation, 0, len(r.tools))
	for _, impl := range r.tools {
		list = append(list, impl)
	}
	return list
}

// NewViewForInterpreter creates a new registry instance that shares the toolset
// of the original registry but is bound to a new interpreter runtime.
func (r *ToolRegistryImpl) NewViewForInterpreter(interpreter Runtime) ToolRegistry {
	return &ToolRegistryImpl{
		tools:       r.tools,
		interpreter: interpreter,
		mu:          r.mu,
	}
}

// --- BRIDGE IMPLEMENTATION ---

// CallFromInterpreter is the single bridge between the Value-based interpreter and primitive-based tools.
func (r *ToolRegistryImpl) CallFromInterpreter(interp Runtime, fullname types.FullName, args []lang.Value) (lang.Value, error) {
	// DEBUG: Add extensive logging to trace the runtime context.
	fmt.Fprintf(os.Stderr, "--- DEBUG: CallFromInterpreter for tool '%s' ---\n", fullname)
	fmt.Fprintf(os.Stderr, "  - Runtime from argument (interp): %T\n", interp)
	fmt.Fprintf(os.Stderr, "  - Runtime from registry (r.interpreter): %T\n", r.interpreter)

	impl, ok := r.GetTool(fullname)
	if !ok {
		canonicalName := CanonicalizeToolName(string(fullname))
		return nil, lang.NewRuntimeError(lang.ErrorCodeToolNotFound, fmt.Sprintf("tool '%s' not found", canonicalName), lang.ErrToolNotFound)
	}

	// Centralized policy enforcement. This correctly checks trust, then grants.
	// Policy checks MUST always run against the live, public-facing runtime (`interp`).
	if err := CanCall(interp, impl); err != nil {
		fmt.Fprintf(os.Stderr, "  - DEBUG: CanCall failed: %v\n", err)
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "  - DEBUG: CanCall succeeded.\n")

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

	// --- START FIX ---
	// By default, pass the live, public-facing interpreter (`interp`).
	runtimeForTool := interp

	// If the tool is marked as internal, unwrap the runtime to pass the
	// internal interpreter, which the tool expects.
	if impl.IsInternal {
		if wrapper, ok := interp.(Wrapper); ok {
			runtimeForTool = wrapper.Unwrap()
			fmt.Fprintf(os.Stderr, "  - DEBUG: Unwrapped runtime for internal tool. Type is now: %T\n", runtimeForTool)
		} else {
			// This should not happen if the architecture is correct.
			fmt.Fprintf(os.Stderr, "  - DEBUG: WARNING: Internal tool '%s' running with wrapped runtime %T (does not implement tool.Wrapper).\n", fullname, interp)
		}
	}
	// --- END FIX ---

	fmt.Fprintf(os.Stderr, "  - DEBUG: Calling tool Func with runtime type: %T\n", runtimeForTool)
	out, err := impl.Func(runtimeForTool, coercedArgs)
	if err != nil {
		// Pass the error through. Do not wrap it again here.
		return nil, err
	}

	fmt.Fprintln(os.Stderr, "-------------------------------------------------")
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

	// This function is for external calls, so it should use the public-facing
	// runtime stored in r.interpreter.
	orderedLangArgs := make([]lang.Value, len(impl.Spec.Args))
	for i, spec := range impl.Spec.Args {
		val, ok := args[spec.Name]
		if !ok {
			if spec.Required {
				return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("missing required argument '%s' for tool '%s'", spec.Name, impl.FullName), lang.ErrArgumentMismatch)
			}
			orderedLangArgs[i] = &lang.NilValue{}
		} else {
			orderedLangArgs[i] = val
		}
	}

	// We call CallFromInterpreter, passing the registry's root *public*
	// interpreter as the runtime. CallFromInterpreter will then correctly
	// unwrap it if needed.
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
