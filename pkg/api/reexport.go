// NeuroScript Version: 0.7.0
// File version: 26
// Purpose: Re-exported policy context constants (e.g., ContextConfig).
// filename: pkg/api/reexport.go
// nlines: 161
// risk_rating: LOW

package api

import (
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/policy/capability"
	"github.com/aprice2704/neuroscript/pkg/provider"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// Re-exported types for the public API, as per the v0.7 contract.
type (
	// ... (other types unchanged) ...
	Kind         = types.Kind
	Position     = types.Position
	Node         = interfaces.Node
	Tree         = interfaces.Tree
	Logger       = interfaces.Logger
	LogLevel     = interfaces.LogLevel
	RuntimeError = lang.RuntimeError
	SignedAST    struct {
		Blob []byte
		Sum  [32]byte
		Sig  []byte
	}
	Value              any
	Option             = interpreter.InterpreterOption
	ExecPolicy         = policy.ExecPolicy
	ExecContext        = policy.ExecContext // Re-export the type
	Capability         = capability.Capability
	AIProvider         = provider.AIProvider
	ToolImplementation = tool.ToolImplementation
	ArgSpec            = tool.ArgSpec
	Runtime            = tool.Runtime
	ToolFunc           = tool.ToolFunc
	ToolSpec           = tool.ToolSpec
	FullName           = types.FullName
	ToolName           = types.ToolName
	ToolGroup          = types.ToolGroup
	ArgType            = tool.ArgType

	// LoopControl holds the parsed result of an AEIOU LOOP signal.
	LoopControl struct {
		Control string // "continue", "done", or "abort"
		Notes   string
		Reason  string
	}
)

// ... (resource/verb consts unchanged) ...
const (
	ResFS     = capability.ResFS
	ResNet    = capability.ResNet
	ResEnv    = capability.ResEnv
	ResModel  = capability.ResModel
	ResTool   = capability.ResTool
	ResSecret = capability.ResSecret
	ResBudget = capability.ResBudget
	ResBus    = capability.ResBus

	VerbRead  = capability.VerbRead
	VerbWrite = capability.VerbWrite
	VerbAdmin = capability.VerbAdmin
	VerbUse   = capability.VerbUse
	VerbExec  = capability.VerbExec
)

// Re-exported constants for policy contexts
const (
	ContextConfig ExecContext = policy.ContextConfig
	ContextNormal ExecContext = policy.ContextNormal
	ContextTest   ExecContext = policy.ContextTest
)

const (
	ArgTypeAny         = tool.ArgTypeAny
	ArgTypeString      = tool.ArgTypeString
	ArgTypeInt         = tool.ArgTypeInt
	ArgTypeFloat       = tool.ArgTypeFloat
	ArgTypeBool        = tool.ArgTypeBool
	ArgTypeMap         = tool.ArgTypeMap
	ArgTypeSlice       = tool.ArgTypeSlice
	ArgTypeSliceString = tool.ArgTypeSliceString
	ArgTypeSliceInt    = tool.ArgTypeSliceInt
	ArgTypeSliceFloat  = tool.ArgTypeSliceFloat
	ArgTypeSliceBool   = tool.ArgTypeSliceBool
	ArgTypeSliceMap    = tool.ArgTypeSliceMap
	ArgTypeSliceAny    = tool.ArgTypeSliceAny
	ArgTypeNil         = tool.ArgTypeNil
)

// ... (vars unchanged) ...
var (
	NewCapability   = capability.New
	ParseCapability = capability.Parse
	MustParse       = capability.MustParse
	NewWithVerbs    = capability.NewWithVerbs
)

// WithTool creates an interpreter option to register a custom tool.
func WithTool(t ToolImplementation) Option {
	return func(i *interpreter.Interpreter) {
		if _, err := i.ToolRegistry().RegisterTool(t); err != nil {
			if logger := i.GetLogger(); logger != nil {
				logger.Error("failed to register tool via WithTool option", "tool", t.Spec.Name, "error", err)
			}
		}
	}
}

// WithEmitFunc creates an interpreter option to set a custom emit handler.
func WithEmitFunc(f func(Value)) Option {
	return func(i *interpreter.Interpreter) {
		// We wrap the api.Value in the function signature to avoid exposing lang.Value.
		i.SetEmitFunc(func(v lang.Value) {
			f(v)
		})
	}
}

// RegisterCriticalErrorHandler allows the host application to override the default
// panic behavior for critical errors.
func RegisterCriticalErrorHandler(h func(*lang.RuntimeError)) {
	lang.RegisterCriticalHandler(h)
}

// MakeToolFullName creates a correctly formatted, fully-qualified tool name.
func MakeToolFullName(group, name string) types.FullName {
	return types.MakeFullName(group, name)
}

// WithExecPolicy applies a runtime execution policy to the interpreter.
func WithExecPolicy(policy *ExecPolicy) Option {
	return interpreter.WithExecPolicy(policy)
}

// WithInterpreter creates an option to reuse the internal state of an existing
// interpreter. This is useful for the host-managed ask-loop pattern.
func WithInterpreter(existing *Interpreter) Option {
	return func(i *interpreter.Interpreter) {
		if existing != nil && existing.internal != nil {
			*i = *existing.internal
		}
	}
}
