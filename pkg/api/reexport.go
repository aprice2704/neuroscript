// NeuroScript Version: 0.8.0
// File version: 61
// Purpose: Centralizes all public API re-exports. Removes the ambiguous WithActor option and adds TurnContextProvider. Corrects ToolGroup re-export.
// filename: pkg/api/reexport.go
// nlines: 144
// risk_rating: LOW
package api

import (
	"github.com/aprice2704/neuroscript/pkg/account"
	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/agentmodel"
	"github.com/aprice2704/neuroscript/pkg/ast"
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

// Re-exported types for the public API
type (
	// Core Interpreter Configuration
	HostContext        = interpreter.HostContext
	HostContextBuilder = interpreter.HostContextBuilder

	// Core Types
	Value        = lang.Value
	Kind         = types.Kind
	Position     = types.Position
	Node         = interfaces.Node
	Tree         = interfaces.Tree
	Logger       = interfaces.Logger
	LogLevel     = interfaces.LogLevel
	RuntimeError = lang.RuntimeError
	Actor        = interfaces.Actor
	SignedAST    struct {
		Blob []byte
		Sum  [32]byte
		Sig  []byte
	}

	// Policy & Capability Types
	ExecPolicy  = policy.ExecPolicy
	ExecContext = policy.ExecContext
	Capability  = capability.Capability
	GrantSet    = capability.GrantSet

	// AI & State Store Types
	AIProvider           = provider.AIProvider
	CapsuleRegistry      = capsule.Registry
	AdminCapsuleRegistry = capsule.Registry
	Capsule              = capsule.Capsule
	AgentModel           = types.AgentModel
	Account              = account.Account
	AgentModelReader     = interfaces.AgentModelReader
	AgentModelAdmin      = interfaces.AgentModelAdmin
	AccountReader        = interfaces.AccountReader
	AccountAdmin         = interfaces.AccountAdmin
	AccountStore         = account.Store
	AgentModelStore      = agentmodel.AgentModelStore

	// Tooling Types
	ToolImplementation = tool.ToolImplementation
	ToolRegistry       = tool.ToolRegistry
	ArgSpec            = tool.ArgSpec
	Runtime            = tool.Runtime
	ToolFunc           = tool.ToolFunc
	ToolSpec           = tool.ToolSpec
	FullName           = types.FullName
	ToolName           = types.ToolName
	ToolGroup          = types.ToolGroup
	ArgType            = tool.ArgType

	// Context Provider for Tools
	TurnContextProvider = interpreter.TurnContextProvider

	// AST Types (for advanced use)
	RootNode = ast.Node
	Program  = ast.Program

	// Telemetry & AEIOU
	Emitter            = interfaces.Emitter
	LLMCallStartInfo   = interfaces.LLMCallStartInfo
	LLMCallSuccessInfo = interfaces.LLMCallSuccessInfo
	LLMCallFailureInfo = interfaces.LLMCallFailureInfo
	Decision           = aeiou.Decision
	LoopController     = aeiou.LoopController
)

// Re-exported constants
const (
	ContextConfig ExecContext = policy.ContextConfig
	ContextNormal ExecContext = policy.ContextNormal
	ContextTest   ExecContext = policy.ContextTest

	// Capability Resources
	ResFS      = capability.ResFS
	ResNet     = capability.ResNet
	ResAccount = capability.ResAccount
	ResModel   = capability.ResModel
	ResCapsule = capability.ResCapsule
	ResEnv     = capability.ResEnv
	ResTool    = capability.ResTool
	ResSecret  = capability.ResSecret
	ResBudget  = capability.ResBudget
	ResBus     = capability.ResBus

	// Capability Verbs
	VerbRead  = capability.VerbRead
	VerbWrite = capability.VerbWrite
	VerbAdmin = capability.VerbAdmin
	VerbUse   = capability.VerbUse
	VerbExec  = capability.VerbExec
	VerbSign  = capability.VerbSign
)

// Re-exported functions and constructors
var (
	// Configuration
	NewHostContextBuilder = interpreter.NewHostContextBuilder
	// WithActor is intentionally removed. Identity must be set via WithHostContext.
	WithGlobals              = interpreter.WithGlobals
	WithExecPolicy           = interpreter.WithExecPolicy
	WithSandboxDir           = interpreter.WithSandboxDir
	WithoutStandardTools     = interpreter.WithoutStandardTools
	WithAccountStore         = interpreter.WithAccountStore
	WithAgentModelStore      = interpreter.WithAgentModelStore
	WithCapsuleRegistry      = interpreter.WithCapsuleRegistry
	WithCapsuleAdminRegistry = interpreter.WithCapsuleAdminRegistry
	WithAITranscriptWriter   = interpreter.WithAITranscriptWriter

	// Capability Constructors
	NewCapability   = capability.New
	ParseCapability = capability.Parse
	MustParse       = capability.MustParse
	NewWithVerbs    = capability.NewWithVerbs

	// Other Constructors
	NewPolicyBuilder        = policy.NewBuilder
	NewAccountStore         = account.NewStore
	NewAgentModelStore      = agentmodel.NewAgentModelStore
	NewAdminCapsuleRegistry = capsule.NewRegistry // Used for admin purposes
	MakeToolFullName        = types.MakeFullName
)
