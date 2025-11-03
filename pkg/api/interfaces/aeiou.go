// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Defines the AeiouOrchestrator interface for the external FDM service hook.
// filename: pkg/api/interfaces/aeiou.go
// nlines: 27

package interfaces

import (
	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// AeiouOrchestrator defines the interface for the FDM service
// that orchestrates the 'ask' loop.
//
// This interface is expected to be implemented by an FDM service
// and injected into the interpreter's HostContext via
// api.WithServiceRegistry. The 'executeAsk' step will then
// call this service to perform the entire multi-turn
// 'ask' conversation.
type AeiouOrchestrator interface {
	// RunAskLoop takes control of the 'ask' execution.
	// It receives the agent's interpreter instance, the name
	// of the agent model to use, and the initial prompt.
	// It is responsible for the entire multi-turn loop, including
	// parsing, validating, and executing actions via
	// api.ExecuteSandboxedAST.
	// It returns the final lang.Value (e.g., from a 'return')
	// or an error if the loop fails.
	RunAskLoop(
		callingInterp *api.Interpreter,
		agentModelName string,
		initialPrompt string,
	) (lang.Value, error)
}
