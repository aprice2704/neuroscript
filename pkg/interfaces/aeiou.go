// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Defines internal AeiouOrchestrator. Uses 'any' for return value to break 'lang' import cycle.
// filename: pkg/interfaces/aeiou.go
// nlines: 32

package interfaces

// FIX: Removed import "github.com/aprice2704/neuroscript/pkg/lang" to break import cycle.

// AeiouServiceKey is the standard key used to retrieve the
// AEIOU orchestrator from the host's ServiceRegistry.
const AeiouServiceKey = "AeiouService"

// AeiouOrchestrator defines the *internal* interface for the FDM service
// that orchestrates the 'ask' loop.
// The 'executeAsk' step will look for this interface in the
// HostContext.ServiceRegistry.
type AeiouOrchestrator interface {
	// RunAskLoop takes control of the 'ask' execution.
	// It receives the *public API* interpreter instance, the name
	// of the agent model to use, and the initial prompt.
	//
	//
	// 'callingInterp' is 'any' to avoid an import cycle.
	// 'executeAsk' will pass its 'i.PublicAPI'.
	//
	// FIX: Return value is 'any' to avoid importing 'pkg/lang'.
	// 'executeAsk' is responsible for type-asserting this to 'lang.Value'.
	RunAskLoop(
		callingInterp any,
		agentModelName string,
		initialPrompt string,
	) (any, error)
}
