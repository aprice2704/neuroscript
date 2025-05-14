// NeuroScript Version: 0.3.0
// File version: 0.1.2
// Updated FormatWMStatusView to display base-36 numbers and chat capability.
// filename: pkg/neurogo/screen_wm_status.go
// nlines: 75 // Approximate
// risk_rating: MEDIUM
package neurogo

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core" // For AIWorkerManager and types
	"github.com/charmbracelet/lipgloss"
)

// Styles remain the same
var (
	wmTitleStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63")).Underline(true) // Keep for potential future use
	wmHeaderStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	wmValueStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("248"))
	wmChatCapableStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("40")) // Bright green for chat capable
	wmInactiveStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	wmErrorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
)

// FormatWMStatusView generates the string content for the Worker Manager status screen.
// It now uses the WMStatusViewDataProvider interface (which App implements).
func FormatWMStatusView(dataProvider WMStatusViewDataProvider) ([]*core.AIWorkerDefinition, string) {
	if dataProvider == nil {
		return nil, wmErrorStyle.Render("Error: Data provider not available for WM Status.")
	}
	logger := dataProvider.GetLogger()

	aiWm := dataProvider.GetAIWorkerManager()
	if aiWm == nil {
		if logger != nil {
			logger.Warn("FormatWMStatusView: AI Worker Manager is not available or not initialized.")
		}
		return nil, wmValueStyle.Render("AI Worker Manager not available.")
	}
	if logger != nil {
		logger.Debug("FormatWMStatusView: Fetching AI Worker Manager status.")
	}

	var sb strings.Builder
	definitions := aiWm.ListWorkerDefinitions(nil) // Get all definitions

	// sb.WriteString(wmTitleStyle.Render("AI Worker Manager Status")) // Title can be handled by the screen view itself
	// sb.WriteString("\n\n")

	sb.WriteString(wmHeaderStyle.Render(
		fmt.Sprintf("--- %d Defined Workers (Select with //chat <id>) ---", len(definitions))))
	sb.WriteString("\n") // Add a newline for better spacing

	if len(definitions) == 0 {
		sb.WriteString(wmValueStyle.Render("No worker definitions loaded.\n"))
	} else {
		// Max width for ID column for alignment
		maxIDWidth := 0
		for i := range definitions {
			idStr := indexToBase36(i) // 0-indexed for selection logic, display is user's choice
			if len(idStr) > maxIDWidth {
				maxIDWidth = len(idStr)
			}
		}
		if maxIDWidth == 0 {
			maxIDWidth = 1
		} // Ensure at least 1 for alignment

		for i, def := range definitions {
			name := def.Name
			if name == "" {
				name = "[Unnamed Definition]"
			}
			defIDShort := def.DefinitionID
			if len(defIDShort) > 12 { // Slightly longer display for actual ID
				defIDShort = defIDShort[:12] + "..."
			}

			statusStyle := wmValueStyle
			if def.Status != core.DefinitionStatusActive {
				statusStyle = wmInactiveStyle
			}

			// Base36 number for selection (0-indexed internally, displayed as 0-z, 10, ...)
			displayID := indexToBase36(i) // Use the helper

			// Check chat capability
			chatCapableIndicator := "   " // 3 spaces if not chat capable
			isChatCapable := false
			for _, im := range def.InteractionModels {
				if im == core.InteractionModelConversational || im == core.InteractionModelBoth {
					isChatCapable = true
					break
				}
			}
			if isChatCapable {
				chatCapableIndicator = wmChatCapableStyle.Render("[C]")
			}

			// Format the main definition line
			// Using Sprintf for padding: %*s means "pad string to width *"
			// The negative sign in %-*s means left-justify
			sb.WriteString(fmt.Sprintf("%s %-*s %s %s (ID: %s)\n",
				chatCapableIndicator,
				maxIDWidth, wmHeaderStyle.Render(displayID), // Display the base36 ID
				wmHeaderStyle.Render(name),
				statusStyle.Render(string(def.Status)),
				wmValueStyle.Render(defIDShort),
			))

			activeInstances := 0
			if def.AggregatePerformanceSummary != nil {
				activeInstances = def.AggregatePerformanceSummary.ActiveInstancesCount
			}

			sb.WriteString(fmt.Sprintf("     %s: %s, %s: %s, %s: %d\n", // Indent details
				wmValueStyle.Render("Provider"),
				wmValueStyle.Render(string(def.Provider)),
				wmValueStyle.Render("Model"),
				wmValueStyle.Render(def.ModelName),
				wmValueStyle.Render("#Active"),
				activeInstances,
			))
			// Add a small separator if not the last item, for readability
			if i < len(definitions)-1 {
				sb.WriteString(wmValueStyle.Render("     ---------------------\n"))
			}
		}
	}
	// sb.WriteString("\n") // Removed, as the title for overall status can be part of status bar or another screen
	// sb.WriteString(wmHeaderStyle.Render("--- Overall Status ---")) // This can be a separate small view or part of status bar

	// Return both the definitions (for the main model to cache) and the formatted string
	return definitions, sb.String()
}
