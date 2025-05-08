// NeuroScript Version: 0.3.0
// File version: 0.1.1 // Updated version
// Removed calls to undefined tui.RestoreTerminal
// filename: pkg/neurogo/app_tui.go
package neurogo

import (
	"context"
	"fmt"

	//"runtime" // No longer needed without the defer/panic handler here

	"github.com/aprice2704/neuroscript/pkg/adapters" // For checking NoOp type
	// For checking LLMClient
	"github.com/aprice2704/neuroscript/pkg/neurogo/tui"
)

// runTuiMode handles the execution logic for the TUI mode.
func (a *App) runTuiMode(ctx context.Context) error {
	logger := a.GetLogger() // Use safe getter
	logger.Info("Starting TUI mode...")

	llmClient := a.GetLLMClient() // Use the correct exported interface method GetLLMClient()

	// Check if the client is functional or the NoOp client
	if _, isNoOp := llmClient.(*adapters.NoOpLLMClient); isNoOp {
		logger.Warn("TUI mode running with NoOpLLMClient (API key likely missing or invalid). LLM features unavailable.")
	} else if llmClient == nil {
		logger.Error("TUI mode starting but LLM Client is nil.")
	} else if llmClient.Client() == nil {
		logger.Warn("TUI mode starting but underlying LLM client is nil (check provider integration?).")
	} else {
		logger.Info("TUI mode starting with a functional LLM Client.")
	}

	// Call the Start function from the tui package.
	// 'a' (type *App) satisfies the tui.AppAccess interface.
	// Assume tui.Start() and the bubbletea program handle terminal restoration.

	// --- Removed defer block calling tui.RestoreTerminal ---
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		// tui.RestoreTerminal() // Removed - undefined
	// 		panic(r)
	// 	}
	// }()

	// Start the TUI
	err := tui.Start(a) // Pass the app instance 'a'

	// --- Removed explicit call to tui.RestoreTerminal ---
	// tui.RestoreTerminal() // Removed - undefined

	if err != nil {
		logger.Error("TUI execution failed.", "error", err)
		// Return the error, potentially wrapping it if more context is needed here
		return fmt.Errorf("TUI execution failed: %w", err)
	}

	logger.Info("TUI mode finished.")
	return nil
}
