// filename: pkg/neurogo/app_tui.go
package neurogo

import (
	"context"
	"fmt"

	// Import io for io.Discard
	// Import the TUI package to call Start
	"github.com/aprice2704/neuroscript/pkg/neurogo/tui"
)

// runTuiMode handles the execution logic for the TUI mode.
func (a *App) runTuiMode(ctx context.Context) error {
	a.Logger.Info("Starting TUI mode...")

	// Check if LLM client is needed and available
	// Use the correct exported interface method GetLLMClient()
	llmClient := a.GetLLMClient()

	if !a.Config.EnableLLM {
		a.Logger.Warn("TUI mode running with LLM disabled via -enable-llm=false flag.")
	} else if llmClient == nil || llmClient.Client() == nil {
		a.Logger.Warn("TUI mode starting without a functional LLM Client (initialization failed or key missing?).")
	}

	// Call the Start function from the tui package.
	// 'a' (type *App) satisfies the tui.AppAccess interface because
	// we added the necessary methods to App.
	err := tui.Start(a) // Pass the app instance 'a'
	if err != nil {
		a.Logger.Error("TUI execution failed: %v", err)
		// Return the error, potentially wrapping it if more context is needed here
		// Wrapping adds context about where the error occurred.
		return fmt.Errorf("TUI execution failed: %w", err)
	}

	a.Logger.Info("TUI mode finished.")
	return nil
}
