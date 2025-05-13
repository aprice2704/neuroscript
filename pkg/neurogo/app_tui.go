// NeuroScript Version: 0.3.0
// File version: 0.1.2 // Pass StartupScript path to tui.Start.
// Removed calls to undefined tui.RestoreTerminal
// filename: pkg/neurogo/app_tui.go
package neurogo

import (
	"context"
	"fmt"

	// "runtime" // No longer needed without the defer/panic handler here

	"github.com/aprice2704/neuroscript/pkg/adapters" // For checking NoOp type
	// For checking LLMClient
	//"github.com/aprice2704/neuroscript/pkg/neurogo"
)

// runTuiMode handles the execution logic for the TUI mode.
// It now passes the configured startup script path to the TUI's Start function.
func (a *App) runTuiMode(ctx context.Context) error {
	logger := a.GetLogger()             // Use safe getter
	logger.Info("Starting TUI mode...") // This seems like a good INFO level log

	llmClient := a.GetLLMClient() // Use the correct exported interface method GetLLMClient()

	// Check if the client is functional or the NoOp client
	if _, isNoOp := llmClient.(*adapters.NoOpLLMClient); isNoOp {
		logger.Warn("TUI mode running with NoOpLLMClient (API key likely missing or invalid). LLM features unavailable.")
	} else if llmClient == nil {
		logger.Error("TUI mode starting but LLM Client is nil.")
	} else if llmClient.Client() == nil { // Assuming LLMClient has a method to get the underlying client
		logger.Warn("TUI mode starting but underlying LLM client is nil (check provider integration?).")
	} else {
		logger.Info("TUI mode starting with a functional LLM Client.") // Good INFO level log
	}

	// Call the Start function from the tui package.
	// 'a' (type *App) satisfies the tui.AppAccess interface.
	// Pass the StartupScript path from the app's configuration.
	// If no script is specified in the config, StartupScript will be an empty string,
	// and tui.Start/model.Init should handle that gracefully (i.e., not run any initial script).
	var initialScriptPath string
	if a.Config != nil {
		initialScriptPath = a.Config.StartupScript
	}

	err := Start(a, initialScriptPath) // MODIFIED: Pass the initial script path

	if err != nil {
		logger.Error("TUI execution failed.", "error", err)
		return fmt.Errorf("TUI execution failed: %w", err)
	}

	logger.Info("TUI mode finished.") // Good INFO level log
	return nil
}
