// NeuroScript Version: 0.4.0
// File version: 0.2.2
// Description: TUI screen for AIWM status. Synchronous populateTable.
//
//	Removed explicit Draw() call and simplified loading message in populateTable.
//	Further simplified table manipulations in OnFocus.
//
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
	// This log helps see if Title() is called frequently or at odd times
	// if s.app != nil && s.app.tui != nil {
	// 	s.app.tui.LogToDebugScreen("[AIWM_TITLE] Title() called, count: %d", count)
	// }
	return fmt.Sprintf("%s (%d)", s.title, count)
}

// Primitive returns the tview.Table widget for this screen.
func (s *AIWMStatusScreen) Primitive() tview.Primitive {
	fmt.Printf("start prim\n")
	if s.app != nil && s.app.tui != nil {
		s.app.tui.LogToDebugScreen("[AIWM_PRIMITIVE] Primitive() called for %s.", s.name)
	}
	if s.table == nil {
		if s.app != nil && s.app.tui != nil {
			s.app.tui.LogToDebugScreen("[AIWM_PRIMITIVE] Table is nil, creating new table.")
		}
		s.table = tview.NewTable().
			SetFixed(1, 0).
			SetSelectable(true, false)
		s.table.SetBorder(false)

		// Headers are set only once when table is created.
		// populateTable will clear data rows but should leave headers if SetFixed works as expected.
		headers := []string{"Idx", "Name", "Status", "Chat?", "API Key"}
		headerColor := tcell.ColorYellow
		for c, header := range headers {
			s.table.SetCell(0, c,
				tview.NewTableCell(header).
					SetTextColor(headerColor).
					SetAlign(tview.AlignCenter).
					SetSelectable(false))
		}
		if s.app != nil && s.app.tui != nil {
			s.app.tui.LogToDebugScreen("[AIWM_PRIMITIVE] Table created and headers set.")
		}
	}
	fmt.Printf("end prim\n")
	return s.table
}

// populateTable (synchronous)
func (s *AIWMStatusScreen) populateTable() {
	fmt.Printf("populate\n")
	if s.app.tui != nil {
		s.app.tui.LogToDebugScreen("[AIWM_POPULATE] Entered populateTable for %s.", s.name)
	}
	if s.table == nil {
		if s.app.tui != nil {
			s.app.tui.LogToDebugScreen("[AIWM_POPULATE] Table is nil. Aborting populateTable.")
		}
		return
	}

	// Clear non-fixed rows. Row 0 is fixed for headers.
	// Start clearing from row 1.
	for r := s.table.GetRowCount() - 1; r >= 1; r-- {
		s.table.RemoveRow(r)
	}
	if s.app.tui != nil {
		s.app.tui.LogToDebugScreen("[AIWM_POPULATE] Data rows cleared (if any). Header row should be intact.")
	}

	// Fetch data (synchronous, expected to be very fast)
	aiwm := s.app.GetAIWorkerManager()
	if aiwm == nil {
		s.lastFetchErr = fmt.Errorf("AIWorkerManager is not available")
		s.displayInfo = nil
		if s.app.tui != nil {
			s.app.tui.LogToDebugScreen("[AIWM_POPULATE] AIWorkerManager is nil.")
		}
	} else {
		if s.app.tui != nil {
			s.app.tui.LogToDebugScreen("[AIWM_POPULATE] Fetching worker definitions...")
		}
		infos, err := aiwm.ListWorkerDefinitionsForDisplay()
		if err != nil {
			s.lastFetchErr = fmt.Errorf("error fetching AI worker definitions: %w", err)
			s.displayInfo = nil
			if s.app.tui != nil {
				s.app.tui.LogToDebugScreen("[AIWM_POPULATE] Error fetching definitions: %v", err)
			}
		} else {
			s.lastFetchErr = nil
			s.displayInfo = infos
			if s.app.tui != nil {
				s.app.tui.LogToDebugScreen("[AIWM_POPULATE] Fetched %d definitions.", len(infos))
			}
		}
	}
	fmt.Printf("got defs\n")

	// Populate table with data or error message
	if s.lastFetchErr != nil {
		s.table.SetCell(1, 0, // Row 1 for the message
			tview.NewTableCell(fmt.Sprintf("Error: %v", s.lastFetchErr)).
				SetTextColor(tcell.ColorRed).
				SetExpansion(s.table.GetColumnCount()). // Use actual column count
				SetAlign(tview.AlignCenter))
	} else if len(s.displayInfo) == 0 {
		s.table.SetCell(1, 0,
			tview.NewTableCell("No AI Worker Definitions found or loaded.").
				SetExpansion(s.table.GetColumnCount()).
				SetAlign(tview.AlignCenter))
	} else {
		for r, info := range s.displayInfo {
			rowNum := r + 1 // Data rows start from 1
			if info == nil || info.Definition == nil {
				s.table.SetCell(rowNum, 0, tview.NewTableCell("Error: Invalid data").SetTextColor(tcell.ColorRed).SetExpansion(s.table.GetColumnCount()))
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
				statusText = string(core.DefinitionStatusActive)
			}
			s.table.SetCell(rowNum, 0, tview.NewTableCell(fmt.Sprintf("%d", rowNum)).SetAlign(tview.AlignRight))
			s.table.SetCell(rowNum, 1, tview.NewTableCell(EscapeTviewTags(defName)))
			s.table.SetCell(rowNum, 2, tview.NewTableCell(EscapeTviewTags(statusText)))
			s.table.SetCell(rowNum, 3, tview.NewTableCell(chatCapableText).SetTextColor(chatColor).SetAlign(tview.AlignCenter))
			s.table.SetCell(rowNum, 4, tview.NewTableCell(EscapeTviewTags(apiKeyStatusText)).SetTextColor(apiKeyColor))
		}
	}
	if s.app.tui != nil {
		s.app.tui.LogToDebugScreen("[AIWM_POPULATE] Finished populating table cells for %s.", s.name)
	}
}

// OnFocus is called when this screen's primitive is about to receive focus.
func (s *AIWMStatusScreen) OnFocus(setFocus func(p tview.Primitive)) {
	if s.app.tui != nil {
		s.app.tui.LogToDebugScreen("[AIWM_ONFOCUS] Entered OnFocus for %s.", s.name)
	}
	if s.table == nil {
		if s.app.tui != nil {
			s.app.tui.LogToDebugScreen("[AIWM_ONFOCUS] Table is nil, calling Primitive().")
		}
		s.Primitive() // This creates the table and sets headers if s.table was nil.
		if s.app.tui != nil {
			s.app.tui.LogToDebugScreen("[AIWM_ONFOCUS] Primitive() returned.")
		}
	}

	if s.app.tui != nil {
		s.app.tui.LogToDebugScreen("[AIWM_ONFOCUS] Calling populateTable (synchronous).")
	}
	s.populateTable() // Populates data rows, headers should already be there or re-added.
	if s.app.tui != nil {
		s.app.tui.LogToDebugScreen("[AIWM_ONFOCUS] populateTable returned.")
	}

	if s.table != nil {
		if s.app.tui != nil {
			s.app.tui.LogToDebugScreen("[AIWM_ONFOCUS] Calling setFocus on the table.")
		}
		setFocus(s.table)
		if s.app.tui != nil {
			s.app.tui.LogToDebugScreen("[AIWM_ONFOCUS] setFocus on table returned.")
		}

		// Row selection logic
		if s.table.GetRowCount() > 1 && len(s.displayInfo) > 0 {
			s.table.Select(1, 0) // Default to selecting the first data row.
			if s.app.tui != nil {
				s.app.tui.LogToDebugScreen("[AIWM_ONFOCUS] Selected row 1,0 in table (first data row).")
			}
		} else if s.table.GetRowCount() > 0 { // Only header exists or no data
			s.table.Select(0, 0) // Select header
			if s.app.tui != nil {
				s.app.tui.LogToDebugScreen("[AIWM_ONFOCUS] Selected row 0,0 (header or no data).")
			}
		}
		if s.app.tui != nil {
			s.app.tui.LogToDebugScreen("[AIWM_ONFOCUS] Row selection logic complete.")
		}
	} else {
		if s.app.tui != nil {
			s.app.tui.LogToDebugScreen("[AIWM_ONFOCUS] Table is nil after populate, cannot set focus/selection.")
		}
	}
	if s.app.tui != nil {
		s.app.tui.LogToDebugScreen("[AIWM_ONFOCUS] OnFocus completed for %s.", s.name)
	}
}

// OnBlur is called when this screen's primitive is about to lose focus.
func (s *AIWMStatusScreen) OnBlur() {
	if s.app.tui != nil {
		s.app.tui.LogToDebugScreen("[AIWM_SCREEN] OnBlur called for %s.", s.name)
	}
}

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
		}
		// debugLog("[AIWM_INPUT] Key: %s", FormatEventKeyForLogging(event)) // Keep this minimal for now

		initiateChatAttempt := false
		if event.Key() == tcell.KeyEnter {
			initiateChatAttempt = true
			debugLog("[AIWM_INPUT] Enter key pressed.")
		} else if event.Key() == tcell.KeyRune && (event.Rune() == 'c' || event.Rune() == 'C') {
			initiateChatAttempt = true
			debugLog("[AIWM_INPUT] 'c' or 'C' key pressed.")
		}

		if initiateChatAttempt {
			row, _ := s.table.GetSelection()
			debugLog("[AIWM_INPUT] Attempting chat initiation. Selected row: %d", row)

			if s.displayInfo == nil {
				debugLog("[AIWM_INPUT] displayInfo is nil. Cannot initiate chat.")
				return nil // Event consumed
			}

			if row > 0 && row-1 < len(s.displayInfo) {
				selectedInfo := s.displayInfo[row-1]
				if selectedInfo == nil || selectedInfo.Definition == nil {
					debugLog("[AIWM_INPUT] Selected info or definition is nil at row %d.", row)
					return nil // Event consumed
				}
				debugLog("[AIWM_INPUT] Valid row. DefID: %s, Name: %s", selectedInfo.Definition.DefinitionID, selectedInfo.Definition.Name)

				if selectedInfo.IsChatCapable && selectedInfo.APIKeyStatus == core.APIKeyStatusFound {
					debugLog("[AIWM_INPUT] Worker chat capable & API key OK. Calling CreateNewChatSession.")
					if s.app.tui == nil {
						debugLog("[AIWM_INPUT] s.app.tui is nil, cannot switch view.")
						if logger != nil {
							logger.Error("TUI is nil, cannot switch to chat view.")
						}
						return nil // Event consumed
					}

					chatSession, err := s.app.CreateNewChatSession(selectedInfo.Definition.DefinitionID)
					if err != nil {
						errMsg := fmt.Sprintf("Failed to start chat with %s", selectedInfo.Definition.Name)
						debugLog("[AIWM_INPUT] %s: %v", errMsg, err)
						if logger != nil {
							logger.Error(errMsg, "defID", selectedInfo.Definition.DefinitionID, "error", err)
						}
						s.app.tui.LogToDebugScreen("[AIWM_ERROR] %s: %v", EscapeTviewTags(errMsg), EscapeTviewTags(err.Error()))
					} else {
						debugLog("[AIWM_INPUT] Chat session %s created for %s. Instance: %s. Switching view.", chatSession.SessionID, chatSession.DefinitionID, chatSession.WorkerInstance.InstanceID)
						if logger != nil {
							logger.Info("Chat session created", "sessionID", chatSession.SessionID, "instanceID", chatSession.WorkerInstance.InstanceID)
						}
						s.app.tui.switchToChatViewAndUpdate(chatSession.SessionID)
						s.app.tui.LogToDebugScreen("[AIWM_INFO] Chat started with %s", EscapeTviewTags(selectedInfo.Definition.Name))
					}
				} else {
					warnMsg := fmt.Sprintf("Worker %s not chat capable or API key issue. Capable: %v, KeyStatus: %s",
						EscapeTviewTags(selectedInfo.Definition.Name), selectedInfo.IsChatCapable, selectedInfo.APIKeyStatus)
					debugLog("[AIWM_INPUT] %s", warnMsg)
					if logger != nil {
						logger.Warn(warnMsg)
					}
					s.app.tui.LogToDebugScreen("[AIWM_WARN] %s", warnMsg)
				}
			} else {
				debugLog("[AIWM_INPUT] No valid data row selected (row index %d). Cannot initiate chat.", row)
				s.app.tui.LogToDebugScreen("[AIWM_INFO] No definition selected in table to start chat.")
			}
			return nil // Event consumed by chat attempt logic
		}

		// If not a chat attempt, pass to table's default input handler
		if s.table != nil {
			tableHandler := s.table.InputHandler()
			if tableHandler != nil {
				tableHandler(event, setFocus)
				return nil // Event is handled by the table's default handler
			}
		}
		return event // Event not handled
	}
}

// IsFocusable indicates if the screen's main primitive should be part of the focus cycle.
func (s *AIWMStatusScreen) IsFocusable() bool {
	return true
}

var _ PrimitiveScreener = (*AIWMStatusScreen)(nil)
