// NeuroScript Version: 0.8.0
// File version: 4
// Purpose: Defines internal AeiouOrchestrator. Uses 'any' for return value to break 'lang' import cycle.
// Latest change: Added ActiveLoopInfo and loop management methods to the interface.
// filename: pkg/interfaces/aeiou.go
// nlines: 48

package interfaces

import "time"

// FIX: Removed import "github.com/aprice2704/neuroscript/pkg/lang" to break import cycle.

// AeiouServiceKey is the standard key used to retrieve the
// AEIOU orchestrator from the host's ServiceRegistry.
const AeiouServiceKey = "aeiou.service"

// --- NEW: Added for loop management ---
// ActiveLoopInfo provides a snapshot of an in-flight 'ask' loop.
type ActiveLoopInfo struct {
	LoopID       string
	AgentName    string
	StartTime    time.Time
	CurrentTurn  int
	LastActivity time.Time
}

// --- END NEW ---

// AeiouOrchestrator defines the *internal* interface for the FDM service
// that orchestrates the 'ask' loop.
type AeiouOrchestrator interface {
	// RunAskLoop takes control of the 'ask' execution.
	// ... (comments unchanged) ...
	//
	// FIX: Return value is 'any' to avoid importing 'pkg/lang'.
	// 'executeAsk' is responsible for type-asserting this to 'lang.Value'.
	RunAskLoop(
		callingInterp any,
		agentModelName string,
		initialPrompt string,
	) (any, error)

	// --- NEW: Added for loop management ---
	// ListActiveLoops returns a snapshot of all currently in-flight 'ask' loops.
	ListActiveLoops() []ActiveLoopInfo

	// CancelLoop forcefully terminates an in-flight 'ask' loop by its ID.
	CancelLoop(loopID string) error
	// --- END NEW ---
}
