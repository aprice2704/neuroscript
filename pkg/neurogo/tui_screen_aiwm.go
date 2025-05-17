// NeuroScript Version: 0.4.0
// File version: 0.1.0
// Description: TUI screen for displaying AI Worker Management status using a tview.Table.
// filename: pkg/neurogo/tui_screen_aiwm.go
// nlines: 180 // Approximate
// risk_rating: MEDIUM
package neurogo

import (
	"fmt"
	// For table cell content if needed
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// AIWMStatusScreen displays the status of AI Worker Definitions in a table.
type AIWMStatusScreen struct {
	app *App // Reference to the main application

	name         string
	title        string
	displayInfo  []*core.AIWorkerDefinitionDisplayInfo
	lastFetchErr error

	table *tview.Table // The tview.Primitive for this screen

	// tviewApp is needed if we want to QueueUpdateDraw from here,
	// but for now, Primitive() will recreate the table on data fetch.
	// If live updates were needed without full redraw, this would be necessary.
	// tviewApp *tview.Application
}

// NewAIWMStatusScreen creates a new screen for AI Worker Management status.
func NewAIWMStatusScreen(app *App) *AIWMStatusScreen {
	if app == nil {
		panic("AIWMStatusScreen: app cannot be nil")
	}
	s := &AIWMStatusScreen{
		app:   app,
		name:  "AIWM",
		title: "AI Worker Definitions",
	}
	// Initial fetch and table creation can happen here or lazily in Primitive()
	return s
}

// Name returns the short identifier for the screen.
func (s *AIWMStatusScreen) Name() string {
	return s.name
}

// Title returns the title to be displayed for the screen.
func (s *AIWMStatusScreen) Title() string {
	count := 0
	if s.displayInfo != nil {
		count = len(s.displayInfo)
	}
	return fmt.Sprintf("%s (%d)", s.title, count)
}

// fetchDisplayInfo calls the AIWorkerManager to get the latest definition display information.
// If data is fetched successfully, it invalidates the cached table to force a redraw.
func (s *AIWMStatusScreen) fetchDisplayInfo() {
	if s.app.AIWorkerManager() == nil {
		s.lastFetchErr = fmt.Errorf("AIWorkerManager is not available in the application")
		s.displayInfo = nil
		s.table = nil // Invalidate table
		return
	}

	infos, err := s.app.AIWorkerManager().ListWorkerDefinitionsForDisplay()
	if err != nil {
		s.lastFetchErr = fmt.Errorf("error fetching AI worker definitions: %w", err)
		s.displayInfo = nil
	} else {
		s.lastFetchErr = nil
		s.displayInfo = infos
	}
	s.table = nil // Invalidate table so Primitive() recreates it with new data
}

// Primitive returns the tview.Table widget for this screen.
// It creates or recreates the table if necessary.
func (s *AIWMStatusScreen) Primitive() tview.Primitive {
	// Always fetch data when primitive is requested to ensure it's up-to-date.
	// For a more optimized approach, fetch could be triggered by other events.
	s.fetchDisplayInfo()

	if s.table == nil { // Recreate table if invalidated or first time
		s.table = tview.NewTable().
			SetFixed(1, 0).            // Fix header row
			SetSelectable(true, false) // Enable row selection, disable column selection

		// Set Headers
		headers := []string{"Idx", "Name", "Status", "Chat?", "API Key"}
		headerColor := tcell.ColorYellow
		for c, header := range headers {
			s.table.SetCell(0, c,
				tview.NewTableCell(header).
					SetTextColor(headerColor).
					SetAlign(tview.AlignCenter).
					SetSelectable(false))
		}

		if s.lastFetchErr != nil {
			s.table.SetCell(1, 0,
				tview.NewTableCell(fmt.Sprintf("Error: %v", s.lastFetchErr)).
					SetTextColor(tcell.ColorRed).
					SetExpansion(len(headers)). // Span all columns
					SetAlign(tview.AlignCenter))
		} else if len(s.displayInfo) == 0 {
			s.table.SetCell(1, 0,
				tview.NewTableCell("No AI Worker Definitions found or loaded.").
					SetExpansion(len(headers)).
					SetAlign(tview.AlignCenter))
		} else {
			for r, info := range s.displayInfo {
				if info == nil || info.Definition == nil {
					s.table.SetCell(r+1, 0, tview.NewTableCell("Error: Invalid data").SetTextColor(tcell.ColorRed))
					continue
				}

				chatCapable := "No"
				chatColor := tcell.ColorDarkGray
				if info.IsChatCapable {
					chatCapable = "Yes"
					chatColor = tcell.ColorGreen
				}

				apiKeyStatusText := string(info.APIKeyStatus)
				apiKeyColor := tcell.ColorWhite
				switch info.APIKeyStatus {
				case core.APIKeyStatusFound:
					apiKeyColor = tcell.ColorGreen
				case core.APIKeyStatusNotFound, core.APIKeyStatusNotConfigured:
					apiKeyColor = tcell.ColorOrange
				case core.APIKeyStatusError:
					apiKeyColor = tcell.ColorRed
				}

				defName := info.Definition.Name
				if len(defName) > 30 { // Simple truncation
					defName = defName[:27] + "..."
				}
				statusText := string(info.Definition.Status)
				if statusText == "" {
					statusText = "unknown"
				}

				s.table.SetCell(r+1, 0, tview.NewTableCell(fmt.Sprintf("%d", r+1)).SetAlign(tview.AlignRight))
				s.table.SetCell(r+1, 1, tview.NewTableCell(defName))
				s.table.SetCell(r+1, 2, tview.NewTableCell(statusText))
				s.table.SetCell(r+1, 3, tview.NewTableCell(chatCapable).SetTextColor(chatColor).SetAlign(tview.AlignCenter))
				s.table.SetCell(r+1, 4, tview.NewTableCell(apiKeyStatusText).SetTextColor(apiKeyColor))
			}
		}
		// s.table.SetDoneFunc(func(key tcell.Key) { ... }) // If table itself should handle 'done'
		s.table.SetBorder(false)
	}
	return s.table
}

// OnFocus is called when this screen's primitive is about to receive focus.
func (s *AIWMStatusScreen) OnFocus(setFocus func(p tview.Primitive)) {
	if s.table != nil {
		setFocus(s.table) // Ensure the table itself gets focus
		// Select the first data row if there are items, otherwise header (or stay unselected)
		if s.table.GetRowCount() > 1 {
			s.table.Select(1, 0)
		}
	}
}

// OnBlur is called when this screen's primitive is about to lose focus.
func (s *AIWMStatusScreen) OnBlur() {
	// Can deselect rows or other cleanup if needed
	// if s.table != nil {
	// s.table.Select(-1, -1) // Deselect (implementation might vary)
	// }
}

// InputHandler allows the screen to process key events when its primitive is focused.
func (s *AIWMStatusScreen) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) *tcell.EventKey {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) *tcell.EventKey {
		if s.table == nil {
			return event // Table not initialized
		}

		// Handle 'Enter' key on a selected row to initiate chat (example)
		if event.Key() == tcell.KeyEnter {
			row, _ := s.table.GetSelection()
			if row > 0 && row-1 < len(s.displayInfo) { // row 0 is header
				selectedInfo := s.displayInfo[row-1]
				// TODO: Implement starting chat with selectedInfo.Definition
				// This would involve calling a method on s.app
				if s.app != nil && s.app.GetLogger() != nil {
					s.app.GetLogger().Info("Enter pressed on AIWM screen",
						"selectedDefName", selectedInfo.Definition.Name,
						"chatCapable", selectedInfo.IsChatCapable,
						"apiKeyStatus", selectedInfo.APIKeyStatus)
				}
				// Example: s.app.StartChatWithWorker(selectedInfo.Definition.DefinitionID)
				// For now, just log and consume the event.
				return nil // Event handled
			}
		}

		// Let the table handle its own navigation keys (arrows, pgup/pgdn, home/end)
		// This is done by returning the event if not handled above.
		// If tview.Table's default InputHandler is needed, it will be invoked after this.
		// However, tview.Table handles navigation internally if no custom handler consumes keys.
		return event
	}
}

// IsFocusable indicates if the screen's main primitive should be part of the focus cycle.
func (s *AIWMStatusScreen) IsFocusable() bool {
	return true // The table should be focusable for selection and navigation
}

// Ensure AIWMStatusScreen satisfies the PrimitiveScreener interface.
var _ PrimitiveScreener = (*AIWMStatusScreen)(nil)
