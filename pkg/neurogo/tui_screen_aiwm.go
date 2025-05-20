// NeuroScript Version: 0.4.0
// File version: 0.1.6 // Use FormatEventKeyForLogging
// Description: TUI screen for AIWM status. InputHandler now attempts to start chat.
// filename: pkg/neurogo/tui_screen_aiwm.go
package neurogo

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// AIWMStatusScreen displays the status of AI Worker Definitions in a table.
type AIWMStatusScreen struct {
	app *App

	name         string
	title        string
	displayInfo  []*core.AIWorkerDefinitionDisplayInfo
	lastFetchErr error

	table *tview.Table
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

func (s *AIWMStatusScreen) fetchDisplayInfo() {
	aiwm := s.app.GetAIWorkerManager()
	if aiwm == nil {
		s.lastFetchErr = fmt.Errorf("AIWorkerManager is not available in the application")
		s.displayInfo = nil
		s.table = nil
		return
	}

	infos, err := aiwm.ListWorkerDefinitionsForDisplay()
	if err != nil {
		s.lastFetchErr = fmt.Errorf("error fetching AI worker definitions: %w", err)
		s.displayInfo = nil
	} else {
		s.lastFetchErr = nil
		s.displayInfo = infos
	}
	s.table = nil
}

// Primitive returns the tview.Table widget for this screen.
func (s *AIWMStatusScreen) Primitive() tview.Primitive {
	s.fetchDisplayInfo()

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
					chatColor = tcell.ColorGreen // Keep green for "Yes"
				}
				apiKeyStatusText := string(info.APIKeyStatus)
				apiKeyColor := tcell.ColorWhite // Default
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
					statusText = string(core.DefinitionStatusActive)
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
		if s.table.GetRowCount() > 1 && len(s.displayInfo) > 0 {
			s.table.Select(1, 0)
		}
	}
}

// OnBlur is called when this screen's primitive is about to lose focus.
func (s *AIWMStatusScreen) OnBlur() {}

// InputHandler allows the screen to process key events.
func (s *AIWMStatusScreen) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) *tcell.EventKey {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) *tcell.EventKey {
		if s.table == nil {
			if s.app.tui != nil {
				s.app.tui.LogToDebugScreen("[AIWM_INPUT] Table is nil, returning event.")
			}
			return event
		}

		logger := s.app.GetLogger()
		debugLog := func(format string, args ...interface{}) {
			if s.app.tui != nil {
				s.app.tui.LogToDebugScreen(format, args...)
			}
			// Also log to main logger if desired, e.g., logger.Debug(fmt.Sprintf(format, args...))
		}

		// Use the new helper function for logging the key
		debugLog("[AIWM_INPUT] Received key: %s, mod: %v", FormatEventKeyForLogging(event), event.Modifiers())

		initiateChatAttempt := false
		if event.Key() == tcell.KeyEnter {
			initiateChatAttempt = true
			debugLog("[AIWM_INPUT] Enter key pressed.")
		} else if event.Key() == tcell.KeyRune && (event.Rune() == 'c' || event.Rune() == 'C') {
			initiateChatAttempt = true
			debugLog("[AIWM_INPUT] 'c' or 'C' key pressed.")
		}

		if initiateChatAttempt {
			row, col := s.table.GetSelection()
			debugLog("[AIWM_INPUT] Attempting chat initiation. Selected row: %d, col: %d", row, col)

			if s.displayInfo == nil {
				debugLog("[AIWM_INPUT] displayInfo is nil. Cannot initiate chat.")
				return nil
			}
			debugLog("[AIWM_INPUT] displayInfo length: %d", len(s.displayInfo))

			if row > 0 && row-1 < len(s.displayInfo) {
				selectedInfo := s.displayInfo[row-1]
				debugLog("[AIWM_INPUT] Valid row selected. Definition: %s, Name: %s", selectedInfo.Definition.DefinitionID, selectedInfo.Definition.Name)

				if selectedInfo.IsChatCapable && selectedInfo.APIKeyStatus == core.APIKeyStatusFound {
					debugLog("[AIWM_INPUT] Worker is chat capable and API key found. Calling CreateNewChatSession.")
					chatSession, err := s.app.CreateNewChatSession(selectedInfo.Definition.DefinitionID)
					if err != nil {
						errMsg := fmt.Sprintf("Failed to start chat with worker %s", selectedInfo.Definition.DefinitionID)
						debugLog("[AIWM_INPUT] %s: %v", errMsg, err)
						logger.Error(errMsg, "error", err)
						if s.app.tui != nil && s.app.tui.statusBar != nil {
							s.app.tui.tviewApp.QueueUpdateDraw(func() {
								s.app.tui.statusBar.SetText(fmt.Sprintf("[red]%s: %v[-]", EscapeTviewTags(errMsg), EscapeTviewTags(err.Error())))
							})
						}
					} else {
						successMsg := fmt.Sprintf("Chat session %s created for %s", chatSession.SessionID, chatSession.DefinitionID)
						debugLog("[AIWM_INPUT] %s. Instance: %s. Calling switchToChatViewAndUpdate.", successMsg, chatSession.WorkerInstance.InstanceID)
						logger.Info(successMsg, "instanceID", chatSession.WorkerInstance.InstanceID)
						if s.app.tui != nil {
							s.app.tui.switchToChatViewAndUpdate(chatSession.SessionID)
						}
					}
				} else {
					warnMsg := fmt.Sprintf("Worker %s not chat capable or API key missing. Capable: %v, KeyStatus: %s",
						selectedInfo.Definition.Name, selectedInfo.IsChatCapable, selectedInfo.APIKeyStatus)
					debugLog("[AIWM_INPUT] %s", warnMsg)
					logger.Warn(warnMsg)
					if s.app.tui != nil && s.app.tui.statusBar != nil {
						s.app.tui.tviewApp.QueueUpdateDraw(func() { s.app.tui.statusBar.SetText(fmt.Sprintf("[orange]%s[-]", EscapeTviewTags(warnMsg))) })
					}
				}
			} else {
				debugLog("[AIWM_INPUT] No valid data row selected (row index %d). Cannot initiate chat.", row)
				if s.app.tui != nil && s.app.tui.statusBar != nil {
					s.app.tui.tviewApp.QueueUpdateDraw(func() { s.app.tui.statusBar.SetText("[yellow]No definition selected in table to start chat.[-]") })
				}
			}
			return nil
		}

		tableHandler := s.table.InputHandler()
		if tableHandler != nil {
			// Use FormatEventKeyForLogging for debug logging
			debugLog("[AIWM_INPUT] Passing key %s to table's default handler.", FormatEventKeyForLogging(event))
			tableHandler(event, setFocus)
		} else {
			debugLog("[AIWM_INPUT] Table has no default handler. Key: %s", FormatEventKeyForLogging(event))
		}
		return event
	}
}

// IsFocusable indicates if the screen's main primitive should be part of the focus cycle.
func (s *AIWMStatusScreen) IsFocusable() bool {
	return true
}

var _ PrimitiveScreener = (*AIWMStatusScreen)(nil)
