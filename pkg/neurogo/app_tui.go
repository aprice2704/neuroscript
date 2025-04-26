// filename: pkg/neurogo/app_tui.go
package neurogo

import (
	"context"
	"fmt"
	"io" // Import io for io.Discard
	"log"

	// Import the TUI package to call Start
	"github.com/aprice2704/neuroscript/pkg/neurogo/tui"
)

// runTuiMode handles the execution logic for the TUI mode.
func (a *App) runTuiMode(ctx context.Context) error {
	a.Logger.Info("Starting TUI mode...")

	// Ensure loggers are not nil before using them
	warnLog := a.WarnLog
	if warnLog == nil {
		warnLog = log.New(io.Discard, "WARN-FALLBACK: ", log.LstdFlags|log.Lshortfile)
	}
	errorLog := a.ErrorLog
	if errorLog == nil {
		errorLog = log.New(io.Discard, "ERROR-FALLBACK: ", log.LstdFlags|log.Lshortfile)
	}

	// Check if LLM client is needed and available
	// Use the correct exported interface method GetLLMClient()
	llmClient := a.GetLLMClient()

	if !a.Config.EnableLLM {
		warnLog.Println("TUI mode running with LLM disabled via -enable-llm=false flag.")
	} else if llmClient == nil || llmClient.Client() == nil {
		warnLog.Println("TUI mode starting without a functional LLM Client (initialization failed or key missing?).")
	}

	// Call the Start function from the tui package.
	// 'a' (type *App) satisfies the tui.AppAccess interface because
	// we added the necessary methods to App.
	err := tui.Start(a) // Pass the app instance 'a'
	if err != nil {
		errorLog.Printf("TUI execution failed: %v", err)
		// Return the error, potentially wrapping it if more context is needed here
		// Wrapping adds context about where the error occurred.
		return fmt.Errorf("TUI execution failed: %w", err)
	}

	a.Logger.Info("TUI mode finished.")
	return nil
}
