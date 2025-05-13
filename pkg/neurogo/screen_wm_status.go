// NeuroScript Version: 0.3.0
// File version: 0.1.1
// Updated to use WMStatusViewDataProvider interface to break import cycle.
// filename: pkg/neurogo/tui/screen_wm_status.go
// nlines: 55 // Approximate
// risk_rating: LOW
package neurogo

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core" // For AIWorkerManager type
	// "github.com/aprice2704/neuroscript/pkg/neurogo" // REMOVED to break import cycle
	"github.com/charmbracelet/lipgloss"
)

// Define some styles for the WM status display (styles remain the same)
var (
	wmTitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63")).Underline(true)
	wmHeaderStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	wmValueStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
	wmInactiveStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	wmErrorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
)

// FormatWMStatusView generates the string content for the Worker Manager status screen.
// It now accepts the WMStatusViewDataProvider interface.
func FormatWMStatusView(dataProvider WMStatusViewDataProvider) string {
	if dataProvider == nil {
		return wmErrorStyle.Render("Error: Data provider not available for WM Status.")
	}
	logger := dataProvider.GetLogger() // Get logger from the dataProvider

	aiWm := dataProvider.GetAIWorkerManager()
	if aiWm == nil {
		if logger != nil {
			logger.Warn("FormatWMStatusView: AI Worker Manager is not available or not initialized.")
		}
		return wmValueStyle.Render("AI Worker Manager not available.")
	}
	if logger != nil {
		logger.Debug("FormatWMStatusView: Fetching AI Worker Manager status.")
	}

	var sb strings.Builder

	// sb.WriteString(wmTitleStyle.Render("AI Worker Manager Status"))
	// sb.WriteString("\n\n")

	definitions := aiWm.ListWorkerDefinitions(nil)

	sb.WriteString(wmHeaderStyle.Render(
		fmt.Sprintf("--- %d Defined Workers ---\n", len(definitions))))

	if len(definitions) == 0 {
		sb.WriteString(wmValueStyle.Render("No worker definitions loaded.\n"))
	} else {
		for i, def := range definitions {
			name := def.Name
			if name == "" {
				name = "[Unnamed Definition]"
			}
			id := def.DefinitionID
			if len(id) > 8 {
				id = id[:8] + "..."
			}

			statusStyle := wmValueStyle
			if def.Status != core.DefinitionStatusActive { // Use core.DefinitionStatusActive
				statusStyle = wmInactiveStyle
			}

			sb.WriteString(fmt.Sprintf("%s: %s (ID: %s)\n",
				wmHeaderStyle.Render(fmt.Sprintf("%2d. %s", i+1, name)),
				statusStyle.Render(string(def.Status)),
				wmValueStyle.Render(id),
			))

			activeInstances := 0
			if def.AggregatePerformanceSummary != nil {
				activeInstances = def.AggregatePerformanceSummary.ActiveInstancesCount
			}

			sb.WriteString(fmt.Sprintf("   %s: %s, %s: %s, %s: %d\n",
				wmValueStyle.Render("Provider"),
				wmValueStyle.Render(string(def.Provider)),
				wmValueStyle.Render("Model"),
				wmValueStyle.Render(def.ModelName),
				wmValueStyle.Render("#Active"),
				activeInstances,
			))
			// if i < len(definitions)-1 {
			// 	sb.WriteString("\n")
			// }
		}
	}
	sb.WriteString("\n")
	sb.WriteString(wmHeaderStyle.Render("--- Overall Status ---"))

	return sb.String()
}
