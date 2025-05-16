// NeuroScript Version: 0.4.0
// File version: 0.1.0
// filename: pkg/neurogo/tui_screen_aiwm.go
// nlines: 90 // Approximate
// risk_rating: LOW
// Short description: Defines the AIWMStatusScreen for displaying AI Worker Manager status (definitions, key status) in the TUI.

package neurogo

import (
	"fmt"
	"strings"
	// For core.WorkerDefinition if ListWorkerDefinitions returns this directly
)

// AIWMStatusScreen displays key information about the AI Worker Manager,
// focusing on worker definitions, their models, and API key status.
// It implements the Screener interface.
type AIWMStatusScreen struct {
	name  string
	title string
	app   *App // Reference to the main application to access AIWorkerManager
}

// NewAIWMStatusScreen creates a new AIWMStatusScreen.
func NewAIWMStatusScreen(name, title string, app *App) *AIWMStatusScreen {
	return &AIWMStatusScreen{
		name:  name,
		title: title,
		app:   app,
	}
}

// Name returns the name of the AIWMStatusScreen.
func (s *AIWMStatusScreen) Name() string {
	return s.name
}

// Title returns the title of the AIWMStatusScreen.
func (s *AIWMStatusScreen) Title() string {
	return s.title
}

// Contents generates and returns the formatted string displaying AIWM status.
func (s *AIWMStatusScreen) Contents() string {
	if s.app == nil {
		return "[red]Error: AIWMStatusScreen app reference is nil. Cannot fetch AIWM status.[-]"
	}
	aiWm := s.app.GetAIWorkerManager() // This is *core.AIWorkerManager
	if aiWm == nil {
		return "[red]Error: AI Worker Manager not available in app. Cannot fetch status.[-]"
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("[::b]%s[::-]\n\n", "AI Worker Manager Status"))

	// Basic WM Status
	defs := aiWm.ListWorkerDefinitions(nil) // Returns []core.WorkerDefinition
	sb.WriteString(fmt.Sprintf("Loaded Worker Definitions: %d\n", len(defs)))
	aiWmSandboxPath := aiWm.GetSandboxDir() // Get actual AIWM sandbox
	sb.WriteString(fmt.Sprintf("AIWM Sandbox: %s\n", aiWmSandboxPath))
	sb.WriteString("\n")

	if len(defs) == 0 {
		sb.WriteString("No AI worker definitions found.\n")
	} else {
		sb.WriteString("[::u]Worker Definitions:[::-]\n\n")
		for i, def := range defs {
			sb.WriteString(fmt.Sprintf("[%d] [yellow]%s[-] ([green]%s[-])\n", i+1, def.Name, def.DefinitionID))
			sb.WriteString(fmt.Sprintf("    Model        : %s\n", def.ModelName))

			// // Check if API key is configured for this worker's provider
			// keyFoundStr := "[red]Not Found[-]"
			// // IsWorkerConfigured takes workerID, it internally resolves provider & checks key.
			// if aiWm.IsWorkerConfigured(def.ID) {
			// 	keyFoundStr = "[green]Found[-]"
			// }
			// sb.WriteString(fmt.Sprintf("    API Key      : %s\n", keyFoundStr))

			// if len(def.Tools) > 0 {
			// 	toolsStr := strings.Join(def.Tools, ", ")
			// 	maxToolsDisplayLen := 60 // Keep the line from getting too long
			// 	if len(toolsStr) > maxToolsDisplayLen {
			// 		toolsStr = toolsStr[:maxToolsDisplayLen-3] + "..."
			// 	}
			// 	sb.WriteString(fmt.Sprintf("    Tools        : %s\n", toolsStr))
			// } else {
			// 	sb.WriteString("    Tools        : (none defined)\n")
			// }
			// "Can chat" determination can be complex. Listing tools gives an indication.
			// For now, we won't explicitly state "Can Chat: Yes/No".
			// User can infer from tools or we can add a helper to WorkerDefinition later.
			sb.WriteString("\n")
		}
	}

	return sb.String()
}
