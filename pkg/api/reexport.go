// NeuroScript Version: 0.6.0
// File version: 20
// Purpose: Re-export the provider.AIProvider interface for use with the new RegisterProvider API method.
// filename: pkg/api/reexport.go
// nlines: 90
// risk_rating: MEDIUM

package api

import (
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy/capability"
	"github.com/aprice2704/neuroscript/pkg/provider"
	"github.com/aprice2704/neuroscript/pkg/runtime"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// Re-exported types for the public API, as per the v0.7 contract.
type (
	// Foundational types from pkg/types, ensuring a stable AST contract.
	Kind     = types.Kind
	Position = types.Position
	Node     = interfaces.Node
	Tree     = interfaces.Tree

	// Logging and error types.
	Logger       = interfaces.Logger
	LogLevel     = interfaces.LogLevel
	RuntimeError = lang.RuntimeError

	// SignedAST is the transport wrapper for a canonicalized and signed tree.
	SignedAST struct {
		Blob []byte
		Sum  [32]byte
		Sig  []byte
	}

	Value  any
	Option = interpreter.InterpreterOption

	// --- POLICY & CAPABILITY TYPES ---
	// Re-exported for building trusted interpreter configurations.
	ExecPolicy = runtime.ExecPolicy
	Capability = capability.Capability

	// --- PROVIDER TYPES ---
	AIProvider = provider.AIProvider

	// Tool-related types needed to define custom tools.
	ToolImplementation = tool.ToolImplementation
	ArgSpec            = tool.ArgSpec
	Runtime            = tool.Runtime
	ToolFunc           = tool.ToolFunc
	ToolSpec           = tool.ToolSpec
	FullName           = types.FullName
	ToolName           = types.ToolName
	ToolGroup          = types.ToolGroup
)

// WithTool creates an interpreter option to register a custom tool.
func WithTool(t ToolImplementation) Option {
	// The RegisterTool function is the authority on canonicalizing names.
	// We simply pass the implementation through to it.
	return func(i *interpreter.Interpreter) {
		if _, err := i.ToolRegistry().RegisterTool(t); err != nil {
			if logger := i.GetLogger(); logger != nil {
				logger.Error("failed to register tool via WithTool option", "tool", t.Spec.Name, "error", err)
			}
		}
	}
}

// InstantiateAllStandardTools returns a slice of ToolImplementation structs for all
// standard tools that come with NeuroScript.
func InstantiateAllStandardTools() []ToolImplementation {
	// This remains a placeholder as the standard tools are auto-bundled.
	return []ToolImplementation{}
}

// RegisterCriticalErrorHandler allows the host application to override the default
// panic behavior for critical errors.
func RegisterCriticalErrorHandler(h func(*lang.RuntimeError)) {
	lang.RegisterCriticalHandler(h)
}

// MakeToolFullName creates a correctly formatted, fully-qualified tool name.
// It acts as a public facade for the internal types.MakeFullName function.
func MakeToolFullName(group, name string) types.FullName {
	return types.MakeFullName(group, name)
}

// WithExecPolicy applies a runtime execution policy to the interpreter.
// This provides an escape hatch for advanced users who need to construct
// a policy manually.
func WithExecPolicy(policy *ExecPolicy) Option {
	return interpreter.WithExecPolicy(policy)
}
