// NeuroScript Version: 0.4.0
// File version: 0.1.1
// Description: TUI screen for AIWM status. InputHandler now attempts to start chat.
// filename: pkg/neurogo/tui_screen_aiwm.go
// nlines: 190 // Approximate
// risk_rating: MEDIUM
package neurogo

import (
	"fmt"
	// For logging a timestamp or context timeout if needed
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
func (s *AIWMStatusScreen) Primitive() tview.Primitive {
	s.fetchDisplayInfo() // Always fetch data when primitive is requested

	if s.table == nil {
		s.table = tview.NewTable().
			SetFixed(1, 0).
			SetSelectable(true, false)

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
					SetExpansion(len(headers)).
					SetAlign(tview.AlignCenter))
		} else if len(s.displayInfo) == 0 {
			s.table.SetCell(1, 0,
				tview.NewTableCell("No AI Worker Definitions found or loaded.").
					SetExpansion(len(headers)).
					SetAlign(tview.AlignCenter))
		} else {
			for r, info := range s.displayInfo {
				if info == nil || info.Definition == nil {
					s.table.SetCell(r+1, 0, tview.NewTableCell("Error: Invalid data").SetTextColor(tcell.ColorRed).SetExpansion(len(headers)))
					continue
				}

				chatCapableText := "No"
				chatColor := tcell.ColorDarkGray
				if info.IsChatCapable {
					chatCapableText = "Yes"
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
				if len(defName) > 30 {
					defName = defName[:27] + "..."
				}
				statusText := string(info.Definition.Status)
				if statusText == "" {
					statusText = string(core.DefinitionStatusActive) // Default if empty for some reason
				}

				s.table.SetCell(r+1, 0, tview.NewTableCell(fmt.Sprintf("%d", r+1)).SetAlign(tview.AlignRight))
				s.table.SetCell(r+1, 1, tview.NewTableCell(defName))
				s.table.SetCell(r+1, 2, tview.NewTableCell(statusText))
				s.table.SetCell(r+1, 3, tview.NewTableCell(chatCapableText).SetTextColor(chatColor).SetAlign(tview.AlignCenter))
				s.table.SetCell(r+1, 4, tview.NewTableCell(apiKeyStatusText).SetTextColor(apiKeyColor))
			}
		}
		s.table.SetBorder(false)
	}
	return s.table
}

// OnFocus is called when this screen's primitive is about to receive focus.
func (s *AIWMStatusScreen) OnFocus(setFocus func(p tview.Primitive)) {
	if s.table != nil {
		setFocus(s.table)
		if s.table.GetRowCount() > 1 {
			s.table.Select(1, 0)
		}
	}
}

// OnBlur is called when this screen's primitive is about to lose focus.
func (s *AIWMStatusScreen) OnBlur() {
	// Optional: Deselect rows or other cleanup
}

// InputHandler allows the screen to process key events.
func (s *AIWMStatusScreen) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) *tcell.EventKey {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) *tcell.EventKey {
		if s.table == nil {
			return event
		}

		if event.Key() == tcell.KeyEnter {
			row, _ := s.table.GetSelection()
			if row > 0 && row-1 < len(s.displayInfo) {
				selectedInfo := s.displayInfo[row-1]
				logger := s.app.GetLogger()

				if selectedInfo.IsChatCapable && selectedInfo.APIKeyStatus == core.APIKeyStatusFound {
					logger.Info("Attempting to start chat with worker...",
						"definitionID", selectedInfo.Definition.DefinitionID,
						"name", selectedInfo.Definition.Name)

					// Potentially use a context with timeout for starting chat
					// ctx, cancel := context.WithTimeout(s.app.Context(), 10*time.Second)
					// defer cancel()
					// For now, use app's main context or background.
					instance, err := s.app.StartChatWithWorker(selectedInfo.Definition.DefinitionID)
					if err != nil {
						logger.Error("Failed to start chat with worker",
							"definitionID", selectedInfo.Definition.DefinitionID, "error", err)
						// Optionally, display this error in a status bar or popup in TUI
					} else {
						logger.Info("Successfully started/resumed chat with worker",
							"definitionID", instance.DefinitionID, "instanceID", instance.InstanceID)
						// The main TUI (tview_tui.go) will need to react to this change,
						// e.g., by checking app.GetActiveChatInstanceDetails() and then
						// switching focus to AI input and updating the chat view.
						// This screen (AIWMStatusScreen) has done its job of initiating the chat.
					}
				} else {
					logger.Warn("Selected worker is not chat capable or API key not found/configured.",
						"name", selectedInfo.Definition.Name,
						"chatCapable", selectedInfo.IsChatCapable,
						"apiKeyStatus", selectedInfo.APIKeyStatus)
				}
				return nil // Event handled
			}
		}
		return event // Pass to table's default handler
	}
}

// IsFocusable indicates if the screen's main primitive should be part of the focus cycle.
func (s *AIWMStatusScreen) IsFocusable() bool {
	return true
}

var _ PrimitiveScreener = (*AIWMStatusScreen)(nil)
