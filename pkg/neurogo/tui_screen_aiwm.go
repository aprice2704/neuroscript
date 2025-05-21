// NeuroScript Version: 0.4.0
// File version: 0.1.2M (Modified from user's 0.1.1)
// Description: TUI screen for AIWM status.
// - Primitive() creates and populates table ONCE.
// - OnFocus() is minimal.
// - InputHandler uses CreateNewChatSession.
// - Includes fmt.Println for hang diagnosis.
// filename: pkg/neurogo/tui_screen_aiwm.go
package neurogo

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type AIWMStatusScreen struct {
	app          *App
	name         string
	title        string
	displayInfo  []*core.AIWorkerDefinitionDisplayInfo
	lastFetchErr error
	table        *tview.Table
}

func NewAIWMStatusScreen(app *App) *AIWMStatusScreen {
	if app == nil {
		panic("AIWMStatusScreen: app cannot be nil")
	}
	s := &AIWMStatusScreen{
		app:   app,
		name:  "AIWM",
		title: "AI Worker Definitions",
	}
	// Log creation
	if app.tui != nil { // Assuming app.tui is accessible for logging
		app.tui.LogToDebugScreen("[AIWM_NEW] NewAIWMStatusScreen created (v0.1.2M) for %s.", s.name)
	}
	return s
}

func (s *AIWMStatusScreen) Name() string { return s.name }

func (s *AIWMStatusScreen) Title() string {
	count := 0
	if s.displayInfo != nil { // displayInfo is populated when Primitive() first builds the table
		count = len(s.displayInfo)
	}
	return fmt.Sprintf("%s (%d)", s.title, count)
}

// Primitive creates and populates the table structure ONCE.
// Subsequent calls return the existing table.
func (s *AIWMStatusScreen) Primitive() tview.Primitive {
	fmt.Println("[STDOUT_AIWM_PRIMITIVE_0.1.2M] Primitive() called.")
	if s.app != nil && s.app.tui != nil {
		s.app.tui.LogToDebugScreen("[AIWM_PRIMITIVE_0.1.2M] Primitive() called for %s.", s.name)
	}

	if s.table == nil {
		fmt.Println("[STDOUT_AIWM_PRIMITIVE_0.1.2M] Table is nil. Creating, fetching data, and populating.")
		if s.app != nil && s.app.tui != nil {
			s.app.tui.LogToDebugScreen("[AIWM_PRIMITIVE_0.1.2M] Table is nil, creating new table and populating ONCE.")
		}
		s.table = tview.NewTable().
			SetFixed(1, 0).
			SetSelectable(true, false)
		s.table.SetBorder(false)

		headers := []string{"Idx", "Name", "Status", "Chat?", "API Key"}
		headerColor := tcell.ColorYellow
		for c, header := range headers {
			s.table.SetCell(0, c,
				tview.NewTableCell(header).
					SetTextColor(headerColor).
					SetAlign(tview.AlignCenter).
					SetSelectable(false))
		}

		// Fetch data and populate (combines your original fetchDisplayInfo and population)
		aiwm := s.app.GetAIWorkerManager() // Corrected method name
		if aiwm == nil {
			s.lastFetchErr = fmt.Errorf("AIWorkerManager is not available in the application")
			s.displayInfo = nil
		} else {
			infos, err := aiwm.ListWorkerDefinitionsForDisplay()
			if err != nil {
				s.lastFetchErr = fmt.Errorf("error fetching AI worker definitions: %w", err)
				s.displayInfo = nil
			} else {
				s.lastFetchErr = nil
				s.displayInfo = infos
			}
		}
		if s.app != nil && s.app.tui != nil {
			s.app.tui.LogToDebugScreen("[AIWM_PRIMITIVE_0.1.2M] Data fetched. Error: %v. Info count: %d.", s.lastFetchErr, len(s.displayInfo))
		}

		// Populate table with data or error message
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
				rowNum := r + 1 // Data rows start at index 1
				if info == nil || info.Definition == nil {
					s.table.SetCell(rowNum, 0, tview.NewTableCell("Error: Invalid data").SetTextColor(tcell.ColorRed).SetExpansion(len(headers)))
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
		if s.app != nil && s.app.tui != nil {
			s.app.tui.LogToDebugScreen("[AIWM_PRIMITIVE_0.1.2M] Table created and populated ONCE.")
		}
	} else {
		if s.app != nil && s.app.tui != nil {
			s.app.tui.LogToDebugScreen("[AIWM_PRIMITIVE_0.1.2M] Table already exists, returning same instance.")
		}
	}
	fmt.Println("[STDOUT_AIWM_PRIMITIVE_0.1.2M] Primitive() returning table.")
	return s.table
}

// OnFocus is now minimal: ensures table exists, sets focus, and selects.
// Data is loaded by Primitive() when the table is first created.
func (s *AIWMStatusScreen) OnFocus(setFocus func(p tview.Primitive)) {
	fmt.Println("[STDOUT_AIWM_ONFOCUS_0.1.2M] Entered OnFocus.")
	if s.app.tui != nil {
		s.app.tui.LogToDebugScreen("[AIWM_ONFOCUS_0.1.2M] Entered OnFocus for %s.", s.name)
	}

	if s.table == nil {
		// This should ideally not happen if addScreen called Primitive correctly.
		fmt.Println("[STDOUT_AIWM_ONFOCUS_0.1.2M] Table is nil! Calling Primitive() to ensure it exists.")
		if s.app.tui != nil {
			s.app.tui.LogToDebugScreen("[AIWM_ONFOCUS_0.1.2M] Table is nil. This is unexpected. Calling Primitive() to create.")
		}
		s.Primitive() // Will create and populate if nil.
		if s.table == nil {
			fmt.Println("[STDOUT_AIWM_ONFOCUS_0.1.2M] Table STILL nil after defensive Primitive call. Aborting OnFocus.")
			if s.app.tui != nil {
				s.app.tui.LogToDebugScreen("[AIWM_ONFOCUS_0.1.2M] Table still nil after Primitive(). Cannot set focus.")
			}
			setFocus(tview.NewBox().SetBorder(true).SetTitle("Error: AIWM Table Nil in OnFocus")) // Focus fallback
			return
		}
	}

	fmt.Println("[STDOUT_AIWM_ONFOCUS_0.1.2M] Calling setFocus(s.table).")
	setFocus(s.table)
	fmt.Println("[STDOUT_AIWM_ONFOCUS_0.1.2M] setFocus(s.table) returned.")

	if s.table.GetRowCount() > 1 { // If headers + data/message row(s)
		s.table.Select(1, 0)
		fmt.Println("[STDOUT_AIWM_ONFOCUS_0.1.2M] Selected row 1,0.")
	} else if s.table.GetRowCount() == 1 { // Only header row
		s.table.Select(0, 0)
		fmt.Println("[STDOUT_AIWM_ONFOCUS_0.1.2M] Selected row 0,0 (header).")
	}
	if s.app.tui != nil {
		s.app.tui.LogToDebugScreen("[AIWM_ONFOCUS_0.1.2M] Completed OnFocus for %s.", s.name)
	}
	fmt.Println("[STDOUT_AIWM_ONFOCUS_0.1.2M] Exiting OnFocus.")
}

func (s *AIWMStatusScreen) OnBlur() {
	fmt.Println("[STDOUT_AIWM_ONBLUR_0.1.2M] Entered OnBlur.")
	if s.app.tui != nil {
		s.app.tui.LogToDebugScreen("[AIWM_ONBLUR_0.1.2M] OnBlur called for %s.", s.name)
	}
}

func (s *AIWMStatusScreen) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) *tcell.EventKey {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) *tcell.EventKey {
		// Assuming s.app and s.app.tui are non-nil due to halting policy.
		s.app.tui.LogToDebugScreen("[AIWM_INPUT_HANDLER] Key: %s, Rune: %c, Mod: %v", event.Name(), event.Rune(), event.Modifiers())

		// Assuming s.table is non-nil due to halting policy.
		// The original 'if s.table == nil' block and its contents are removed.

		// Handle Enter key for chat creation specifically.
		if event.Key() == tcell.KeyEnter {
			s.app.tui.LogToDebugScreen("[AIWM_INPUT_HANDLER] Enter key pressed.")
			row, _ := s.table.GetSelection()

			// Assuming s.displayInfo is non-nil due to halting policy.
			// The original 'if s.displayInfo == nil' block and its contents are removed.

			// The check for row bounds is a logical data validation, not a nil check of a disallowed state.
			if row > 0 && row-1 < len(s.displayInfo) { //
				selectedInfo := s.displayInfo[row-1] //

				// Assuming selectedInfo and selectedInfo.Definition are non-nil due to halting policy.
				// The original 'if selectedInfo == nil || selectedInfo.Definition == nil' block is removed.

				logger := s.app.GetLogger() // Assuming GetLogger() guarantees a non-nil logger or panics.

				if selectedInfo.IsChatCapable && selectedInfo.APIKeyStatus == core.APIKeyStatusFound { //
					logger.Info("Attempting to start chat with worker...", //
						"definitionID", selectedInfo.Definition.DefinitionID,
						"name", selectedInfo.Definition.Name)

					chatSession, err := s.app.CreateNewChatSession(selectedInfo.Definition.DefinitionID) //
					if err != nil {                                                                      // Error check (not a nil policy violation) - KEEP
						logger.Error("Failed to create chat session with worker", //
							"definitionID", selectedInfo.Definition.DefinitionID, "error", err)
						s.app.tui.LogToDebugScreen("[AIWM_ERROR_HANDLER] Failed to create chat session: %v", err) //
					} else {
						logger.Info("Successfully created chat session", //
							"definitionID", chatSession.DefinitionID, "sessionID", chatSession.SessionID)
						s.app.tui.LogToDebugScreen("[AIWM_INFO_HANDLER] Chat session created for %s. Main TUI should handle view/focus switch.", chatSession.DefinitionID) //
					}
				} else {
					logger.Warn("Selected worker is not chat capable or API key not found/configured.", //
						"name", selectedInfo.Definition.Name,
						"chatCapable", selectedInfo.IsChatCapable,
						"apiKeyStatus", selectedInfo.APIKeyStatus)
					s.app.tui.LogToDebugScreen("[AIWM_WARN_HANDLER] Worker %s not chat capable or API key issue.", selectedInfo.Definition.Name) //
				}
			}
			return nil // Enter key is consumed by AIWM screen's action
		}

		// For all other keys, let the table handle it first.
		tableHandler := s.table.InputHandler()
		// This 'if' checks if the table *provides* an input handler.
		// This is a capability check, not a "this should never be nil" state check, so it remains.
		if tableHandler != nil {
			s.app.tui.LogToDebugScreen("[AIWM_INPUT_HANDLER] Passing key %s to table's default handler.", event.Name())
			tableHandler(event, setFocus)
		}

		s.app.tui.LogToDebugScreen("[AIWM_INPUT_HANDLER] Returning event %s for further processing by global key handler.", event.Name())
		return event // Pass all other events to global key handler
	}
}

func (s *AIWMStatusScreen) IsFocusable() bool { return true }

var _ PrimitiveScreener = (*AIWMStatusScreen)(nil)
