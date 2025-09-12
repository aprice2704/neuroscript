// NeuroScript Version: 0.7.2
// File version: 2
// Purpose: Defines a neutral, dependency-free interface for emitting LLM telemetry.
// filename: pkg/ns_interfaces/emitter.go

package interfaces

import (
	"context"
	"time"

	"github.com/aprice2704/neuroscript/pkg/provider"
)

// LLMCallStartInfo contains the metadata available at the start of an LLM call.
type LLMCallStartInfo struct {
	Ctx     context.Context
	CallID  string
	Request provider.AIRequest
	Start   time.Time
}

// LLMCallSuccessInfo contains the results of a successful LLM call.
type LLMCallSuccessInfo struct {
	Ctx      context.Context
	CallID   string
	Request  provider.AIRequest
	Response provider.AIResponse
	Latency  time.Duration
}

// LLMCallFailureInfo contains the details of a failed LLM call.
type LLMCallFailureInfo struct {
	Ctx     context.Context
	CallID  string
	Request provider.AIRequest
	Err     error
	Latency time.Duration
}

// Emitter is an interface for a component that can receive telemetry about
// the lifecycle of LLM calls. This decouples llmconn from any specific
// event bus implementation (like FDM's).
type Emitter interface {
	EmitLLMCallStarted(info LLMCallStartInfo)
	EmitLLMCallSucceeded(info LLMCallSuccessInfo)
	EmitLLMCallFailed(info LLMCallFailureInfo)
}
