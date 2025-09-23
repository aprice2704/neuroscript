// NeuroScript Version: 0.7.3
// File version: 39
// Purpose: Re-exports the AccountStore and AgentModelStore types and their constructors for host-managed persistence.
// filename: pkg/api/reexport.go
package api

import (
	"github.com/aprice2704/neuroscript/pkg/account"
	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/agentmodel"
	"github.com/aprice2704/neuroscript/pkg/ast"
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
	Option               = interpreter.InterpreterOption
	ExecPolicy           = policy.ExecPolicy
	ExecContext          = policy.ExecContext // Re-export the type
	Capability           = capability.Capability
	AIProvider           = provider.AIProvider
	ToolImplementation   = tool.ToolImplementation
	ToolRegistry         = tool.ToolRegistry
	ArgSpec              = tool.ArgSpec
	Runtime              = tool.Runtime
	ToolFunc             = tool.ToolFunc
	ToolSpec             = tool.ToolSpec
	FullName             = types.FullName
	ToolName             = types.ToolName
	ToolGroup            = types.ToolGroup
	ArgType              = tool.ArgType
	CapsuleRegistry      = capsule.Registry
	AdminCapsuleRegistry = CapsuleRegistry // FIX: Alias the public type to make it exportable.
	Capsule              = capsule.Capsule // So hosts can construct capsules
	AgentModel           = types.AgentModel
	AgentModelReader     = interfaces.AgentModelReader
	AgentModelAdmin      = interfaces.AgentModelAdmin
	LLMClient            = interfaces.LLMClient
	GrantSet             = capability.GrantSet
	RootNode             = ast.Node
	Program              = ast.Program
	AccountStore         = account.Store
	AgentModelStore      = agentmodel.AgentModelStore

	// --- LLM Telemetry Emitter ---
	Emitter            = interfaces.Emitter
	LLMCallStartInfo   = interfaces.LLMCallStartInfo
	LLMCallSuccessInfo = interfaces.LLMCallSuccessInfo
	LLMCallFailureInfo = interfaces.LLMCallFailureInfo

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

const (
	ResFS      = capability.ResFS
	ResNet     = capability.ResNet
	ResEnv     = capability.ResEnv
	ResModel   = capability.ResModel
	ResTool    = capability.ResTool
	ResSecret  = capability.ResSecret
	ResBudget  = capability.ResBudget
	ResBus     = capability.ResBus
	ResCapsule = capability.ResCapsule
	ResAccount = capability.ResAccount
	ResIPC     = capability.ResIPC

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

var (
	NewCapability           = capability.New
	ParseCapability         = capability.Parse
	MustParse               = capability.MustParse
	NewWithVerbs            = capability.NewWithVerbs
	NewLoopController       = aeiou.NewLoopController
	NewReplayCache          = aeiou.NewReplayCache
	NewProgressTracker      = aeiou.NewProgressTracker
	NewMagicVerifier        = aeiou.NewMagicVerifier
	ComputeHostDigest       = aeiou.ComputeHostDigest
	NewRotatingKeyProvider  = aeiou.NewRotatingKeyProvider
	NewCapsuleRegistry      = capsule.NewRegistry
	NewAdminCapsuleRegistry = capsule.NewRegistry // The new constructor
	NewPolicyBuilder        = policy.NewBuilder
	DecodeWithRegistry      = canon.DecodeWithRegistry
	NewAccountStore         = account.NewStore
	NewAgentModelStore      = agentmodel.NewAgentModelStore
)
