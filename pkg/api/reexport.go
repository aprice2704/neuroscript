// NeuroScript Version: 0.7.1
// File version: 33
// Purpose: Re-exported canon.DecodeWithRegistry for FDM integration.
// filename: pkg/api/reexport.go
// nlines: 191
// risk_rating: LOW
package api

import (
	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/canon"
	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/provider"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// Re-exported types for the public API, as per the v0.7 contract.
type (
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
	CapsuleRegistry    = capsule.Registry

	// --- AEIOU v3 Host Components ---
	// LoopControl holds the parsed result of an AEIOU LOOP signal. (DEPRECATED)
	LoopControl struct {
		Control string
		Notes   string
		Reason  string
	}
	// HostContext provides the necessary host-side information for a turn.
	HostContext = aeiou.HostContext
	// Decision represents the outcome of a turn.
	Decision = aeiou.Decision
	// LoopController orchestrates the host's decision-making process.
	LoopController = aeiou.LoopController
	// ReplayCache detects and prevents token replay attacks.
	ReplayCache = aeiou.ReplayCache
	// ProgressTracker detects repetitive, non-progressing loops.
	ProgressTracker = aeiou.ProgressTracker
	// KeyProvider is an interface for looking up public keys.
	KeyProvider = aeiou.KeyProvider
	// MagicVerifier parses and validates AEIOU v3 control tokens.
	MagicVerifier = aeiou.MagicVerifier
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
	NewCapability          = capability.New
	ParseCapability        = capability.Parse
	MustParse              = capability.MustParse
	NewWithVerbs           = capability.NewWithVerbs
	NewLoopController      = aeiou.NewLoopController
	NewReplayCache         = aeiou.NewReplayCache
	NewProgressTracker     = aeiou.NewProgressTracker
	NewMagicVerifier       = aeiou.NewMagicVerifier
	ComputeHostDigest      = aeiou.ComputeHostDigest
	NewRotatingKeyProvider = aeiou.NewRotatingKeyProvider
	NewCapsuleRegistry     = capsule.NewRegistry
	NewPolicyBuilder       = policy.NewBuilder
	DecodeWithRegistry     = canon.DecodeWithRegistry
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

// WithCapsuleRegistry adds a custom capsule registry to the interpreter's store.
// This allows host applications to layer in their own documentation.
func WithCapsuleRegistry(registry *CapsuleRegistry) Option {
	return func(i *interpreter.Interpreter) {
		if cs := i.CapsuleStore(); cs != nil {
			cs.Add(registry)
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

// WithEventHandlerErrorCallback creates an interpreter option to set a custom
// callback for handling errors that occur within event handlers.
func WithEventHandlerErrorCallback(f func(eventName, source string, err *RuntimeError)) Option {
	return interpreter.WithEventHandlerErrorCallback(f)
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
