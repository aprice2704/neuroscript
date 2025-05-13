// NeuroScript Version: 0.3.0
// File version: 0.0.4
// Defines the comprehensive AppAccess interface required by the TUI package.
// Added missing methods used by TUI components.
// filename: pkg/neurogo/interfaces.go
// nlines: 38 // Approximate
// risk_rating: LOW
package neurogo

import (

	// For io.Writer, if GetInterpreter().SetStdout needs to be interfaced
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// // AppAccess defines the methods the TUI components need
// // from the embedding application.
// type AppAccess interface {
// 	// Config/State Accessors
// 	GetModelName() string
// 	GetLogger() logging.Logger
// 	GetAIWorkerManager() *core.AIWorkerManager
// 	GetInterpreter() *core.Interpreter // REQUIRED by tui.go and update_helpers.go/runSyncCmd
// 	Context() context.Context

// 	// Script Execution
// 	ExecuteScriptFile(ctx context.Context, scriptPath string) error

// 	// Sync related methods (used by runSyncCmd in update_helpers.go)
// 	GetSyncDir() string           // REQUIRED by update_helpers.go/runSyncCmd
// 	GetSyncFilter() string        // REQUIRED by update_helpers.go/runSyncCmd
// 	GetSyncIgnoreGitignore() bool // REQUIRED by update_helpers.go/runSyncCmd

// 	// LLM Client access - might be needed if other TUI screens make direct LLM calls
// 	// GetLLMClient() core.LLMClient // Currently not directly called by TUI on app, but good to have if needed
// }

// WMStatusViewDataProvider defines the methods needed by the WM Status screen.
// tui.AppAccess above must satisfy this.
type WMStatusViewDataProvider interface {
	GetAIWorkerManager() *core.AIWorkerManager
	GetLogger() logging.Logger
}
