// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 3
// :: description: Re-exports internal types for the AEIOU v2+ "LLM Orchestration SDK". Corrects emitter and loop controller.
// :: latestChange: Removed deprecated aeiou.NewLoopController export.
// :: filename: pkg/api/reexport_sdk.go
// :: serialization: go

package api

import (
	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/llmconn"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// This file provides the "LLM Orchestration SDK" types required by
// an external AEIOU orchestrator (like FDM's AeiouService)
// to manage the 'ask' loop, as specified in ns_hook.md.

// Re-exported types for the AEIOU SDK
type (
	// Language Primitives
	StringValue = lang.StringValue
	MapValue    = lang.MapValue
	NilValue    = lang.NilValue

	// AEIOU Protocol Types
	AeiouEnvelope = aeiou.Envelope

	// LLM Connection Types
	// AIProvider is already re-exported in reexport.go
	LLMConnector = llmconn.Connector
	LLMEmitter   = interfaces.Emitter // This is the full interface
)

// Re-exported functions for the AEIOU SDK
var (
	// Language Primitives
	LangToString = lang.ToString
	LangWrap     = lang.Wrap
	LangUnwrap   = lang.Unwrap

	// AEIOU Protocol Functions
	ParseAeiouEnvelope = aeiou.Parse
	ComputeHostDigest  = aeiou.ComputeHostDigest
	NewProgressTracker = aeiou.NewProgressTracker

	// LLM Connection Constructor
	// This wraps the internal llmconn.New, adapting it for the public API.
	NewConnector = func(agentModel *types.AgentModel, provider AIProvider, emitter LLMEmitter) (LLMConnector, error) {
		// FIX: The internal llmconn.New expects the full interfaces.Emitter,
		// not a simple func. We pass it through directly.
		// This fixes the 'cannot use...' and 'missing method...' errors.
		//
		return llmconn.New(agentModel, provider, emitter)
	}
)
