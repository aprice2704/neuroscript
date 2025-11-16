// NeuroScript Version: 0.8.0
// File version: 77
// Purpose: Re-exports all types for the facade, correcting store interfaces AND concrete store names.
// Latest change: Removed duplicate NewCapsuleStore (defined in capsule.go).
// filename: pkg/api/reexport.go
// nlines: 176
package api

import (
	"github.com/aprice2704/neuroscript/pkg/account"
	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/agentmodel"
	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/capsule"
	"github.com/aprice2704/neuroscript/pkg/interfaces" // <--- MUST BE IMPORTED
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
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

	// SymbolProvider
	SymbolProvider = interfaces.SymbolProvider

	// AI & State Store Types
	AIProvider             = provider.AIProvider
	ProviderRegistry       = provider.Registry
	ProviderRegistryReader = interfaces.ProviderRegistryReader
	ProviderRegistryAdmin  = interfaces.ProviderRegistryAdmin
	CapsuleRegistry        = capsule.Registry
	// AdminCapsuleRegistry   = capsule.Registry // REMOVED: Stale
	Capsule = capsule.Capsule
	// CapsuleProvider        = interfaces.CapsuleProvider // REMOVED: Stale
	CapsuleStore = capsule.Store
	AgentModel   = types.AgentModel
	Account      = account.Account

	// --- CONCRETE STORES (for old behavior) ---
	// Renamed to avoid collision with admin interfaces
	AccountStoreConcrete    = account.Store
	AgentModelStoreConcrete = agentmodel.AgentModelStore
	// ------------------------------------------

	// --- CORRECTED FACADE INTERFACES ---
	// Admin interfaces for facade injection
	// We re-export them with the simpler names for public API use.
	AccountStore    = interfaces.AccountAdmin
	AgentModelStore = interfaces.AgentModelAdmin
	// -----------------------------------

	// --- AEIOU HOOK INTERFACE ---
	AeiouOrchestrator = interfaces.AeiouOrchestrator //
	// ----------------------------

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

	// SymbolProvider
	SymbolProviderKey = interfaces.SymbolProviderKey

	// --- AEIOU HOOK KEY ---
	AeiouServiceKey = interfaces.AeiouServiceKey //
	// ----------------------

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
	WithGlobals           = interpreter.WithGlobals
	WithExecPolicy        = interpreter.WithExecPolicy
	WithSandboxDir        = interpreter.WithSandboxDir
	WithoutStandardTools  = interpreter.WithoutStandardTools
	// --- FIXED TYPO ---
	WithAITranscriptWriter = interpreter.WithAITranscriptWriter //
	// ------------------
	WithCapsuleStore = interpreter.WithCapsuleStore

	// Loggers
	NewNoOpLogger = logging.NewNoOpLogger
	NewTestLogger = logging.NewTestLogger

	// Capability Constructors
	NewCapability   = capability.New
	ParseCapability = capability.Parse
	MustParse       = capability.MustParse
	NewWithVerbs    = capability.NewWithVerbs

	// Other Constructors
	NewPolicyBuilder          = policy.NewBuilder
	NewProviderRegistry       = provider.NewRegistry
	NewProviderRegistryReader = provider.NewReader
	NewProviderAdmin          = provider.NewAdmin
	// --- THE FIX: Export correct store building blocks ---
	NewCapsuleRegistry     = capsule.NewRegistry
	BuiltInCapsuleRegistry = capsule.BuiltInRegistry
	// NewCapsuleStore          = capsule.NewStore // REMOVED: Defined in capsule.go
	// --- END FIX ---
	MakeToolFullName = types.MakeFullName

	// --- STORE CONSTRUCTORS & OPTIONS ---
	// Concrete store constructors
	NewAccountStore    = account.NewStore
	NewAgentModelStore = agentmodel.NewAgentModelStore

	// Concrete store options (renamed)
	WithAccountStoreConcrete    = interpreter.WithAccountStore
	WithAgentModelStoreConcrete = interpreter.WithAgentModelStore
	WithProviderRegistry        = interpreter.WithProviderRegistry

	// New facade interface options
	WithAccountStore    = interpreter.WithAccountAdmin
	WithAgentModelStore = interpreter.WithAgentModelAdmin
	// ----------------------------------
)

// AIRequest is a re-export of types.AIRequest for the public API.
type AIRequest = types.AIRequest

// AIResponse is a re-export of types.AIRequest for the public API.
type AIResponse = types.AIResponse

type ProgressTracker = aeiou.ProgressTracker

// ActiveLoopInfo is the re-exported struct for 'ask' loop observability.
type ActiveLoopInfo = interfaces.ActiveLoopInfo // This should be line 209
