// NeuroScript Version: 0.4.0
// File version: 0.3.18
// Description: Main TUI entry point. Ctrl+Q to quit. Compiler nits fixed.
// filename: pkg/neurogo/tview_tui.go
package neurogo

import (
	"context"
	"fmt" // Keep for tvP.originalStdout and Printf debugging
	"log"
	"os" // For redirecting Println if chosen (though shell redirection is better)
	"path/filepath"
	"strings" // Keep for strings.TrimSpace
	"time"    // Keep for context.WithTimeout

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// debugFile is a global variable to hold the debug log file.
var debugFile *os.File

// InitTUIDebugLog sets up a file for redirecting Println statements for debugging the TUI.
func InitTUIDebugLog(filePath string) error {
	var err error
	debugFile, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open debug log file %s: %w", filePath, err)
	}
	return nil
}

// CloseTUIDebugLog closes the debug log file if it was opened.
func CloseTUIDebugLog() {
	if debugFile != nil {
		fmt.Fprintln(debugFile, "Closing TUI debug log.")
		debugFile.Close()
		debugFile = nil
	}
}

// TuiPrintf is a helper to print to the debugFile if it's open, otherwise to stdout.
func TuiPrintf(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if debugFile != nil {
		// Ensure the message ends with a newline for file logging
		if !strings.HasSuffix(msg, "\n") {
			msg += "\n"
		}
		fmt.Fprint(debugFile, msg)
	} else {
		fmt.Println(msg) // Fallback to stdout if debugFile not initialized
	}
}

// StartTviewTUI initializes and runs the tview-based Text User Interface.
func StartTviewTUI(mainApp *App, initialScriptPath string) error {
	// Example: Initialize TUI debug log (uncomment to use)
	// if err := InitTUIDebugLog("tui_debug.log"); err != nil {
	// 	log.Printf("Warning: Could not initialize TUI debug log: %v", err)
	// }
	// defer CloseTUIDebugLog()

	logInfo := func(msg string, keyvals ...interface{}) {
		if mainApp != nil && mainApp.Log != nil {
			mainApp.Log.Info(msg, keyvals...)
		} else {
			log.Printf("INFO: %s %v\n", msg, keyvals)
		}
	}
	logError := func(msg string, keyvals ...interface{}) {
		if mainApp != nil && mainApp.Log != nil {
			mainApp.Log.Error(msg, keyvals...)
		} else {
			log.Printf("ERROR: %s %v\n", msg, keyvals)
		}
	}

	logInfo("StartTviewTUI initializing...")
	if mainApp == nil {
		log.Println("CRITICAL ERROR in StartTviewTUI: mainApp parameter is nil.")
		return fmt.Errorf("mainApp instance cannot be nil")
	}

	tvApp := tview.NewApplication()
	tvP := &tviewAppPointers{
		tviewApp:      tvApp,
		app:           mainApp,
		chatScreenMap: make(map[string]*ChatConversationScreen),
	}
	if mainApp.tui == nil {
		mainApp.tui = tvP
	}

	tvP.debugScreen = NewDynamicOutputScreen("DebugLog", "Debug Log", tvP.tviewApp)
	scriptOutputScreen := NewDynamicOutputScreen("ScriptOut", "Script Output", tvP.tviewApp)

	tvP.localOutputView = tview.NewPages().SetChangedFunc(func() { tvP.onPanePageChange(tvP.localOutputView) })
	tvP.aiOutputView = tview.NewPages().SetChangedFunc(func() { tvP.onPanePageChange(tvP.aiOutputView) })

	tvP.localInputArea = tview.NewTextArea().SetWrap(false).SetWordWrap(false)
	tvP.localInputArea.SetBorder(false).SetTitle("C: Local Input")

	tvP.aiInputArea = tview.NewTextArea().SetWrap(true).SetWordWrap(true)
	tvP.aiInputArea.SetBorder(false).SetTitle("D: AI Input")

	tvP.aiInputArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter && event.Modifiers() == tcell.ModNone {
			text := strings.TrimSpace(tvP.aiInputArea.GetText())
			if text != "" {
				activeChatSession := tvP.app.GetActiveChatSession()
				if activeChatSession == nil {
					TuiPrintf("[AI_INPUT] Enter pressed. No active chat session. Text: %s", text)
					logError("Cannot send chat: No active chat session.", "text", text)
					if tvP.tviewApp != nil && tvP.statusBar != nil {
						tvP.tviewApp.QueueUpdateDraw(func() {
							tvP.statusBar.SetText("[red]No active chat. Select worker from AIWM (Tab to Pane A, Enter or 'c') then type here.[-]")
						})
					}
					return nil
				}
				TuiPrintf("[AI_INPUT] Enter pressed. Active session: %s. Sending: %s", activeChatSession.SessionID, text)
				logInfo("Sending chat message from AI Input", "text", text, "sessionID", activeChatSession.SessionID)

				go func(msgToSend string, sessID string) {
					ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
					defer cancel()
					TuiPrintf("[AI_INPUT_GOROUTINE] Sending to session %s: %s", sessID, msgToSend)
					_, err := tvP.app.SendChatMessageToActiveSession(ctx, msgToSend) //

					if tvP.tviewApp != nil {
						tvP.tviewApp.QueueUpdateDraw(func() {
							TuiPrintf("[AI_INPUT_UI_UPDATE] Updating screen for session %s after send.", sessID)
							if screen, ok := tvP.chatScreenMap[sessID]; ok {
								if err != nil {
									errMsg := fmt.Sprintf("Error sending/processing chat: %v", err)
									TuiPrintf("[AI_INPUT_UI_UPDATE] %s (Session: %s)", errMsg, sessID)
									logError("Error sending/processing chat message", "sessionID", sessID, "error", err)
									if screen.textView != nil {
										fmt.Fprintln(screen.textView, "[red]"+EscapeTviewTags(errMsg)+"[-]")
										screen.textView.ScrollToEnd()
									}
								}
								screen.UpdateConversation()
							} else {
								TuiPrintf("[AI_INPUT_UI_UPDATE] Chat screen NOT FOUND in map for session ID: %s", sessID)
								logError("Chat screen not found in map for session ID", "sessionID", sessID)
							}
							tvP.updateStatusText()
						})
					}
				}(text, activeChatSession.SessionID)
				tvP.aiInputArea.SetText("", true)
			}
			return nil
		}
		return event
	})

	tvP.initialActivityText = "NeuroScript TUI Ready"
	if initialScriptPath != "" {
		tvP.initialActivityText = fmt.Sprintf("Script: %s | Ready", filepath.Base(initialScriptPath))
	}
	tvP.statusBar = tview.NewTextView().SetDynamicColors(true)

	TuiPrintf("[TUI_INIT] UI Primitives (like debugScreen, statusbar) initialized.")

	helpStaticScreen := NewStaticPrimitiveScreen("Help", "Help", helpText)
	aiwmScreen := NewAIWMStatusScreen(tvP.app) //

	tvP.addScreen(scriptOutputScreen, true)
	tvP.addScreen(aiwmScreen, true)
	tvP.addScreen(helpStaticScreen, true)
	tvP.helpScreenIndex = 2

	tvP.addScreen(tvP.debugScreen, false)
	tvP.addScreen(NewStaticPrimitiveScreen("HelpRight", "Help (Right)", helpText), false)

	tvP.focusablePrimitives = []tview.Primitive{
		tvP.localInputArea, tvP.aiInputArea, tvP.aiOutputView, tvP.localOutputView,
	}
	tvP.numFocusablePrimitives = len(tvP.focusablePrimitives)
	tvP.paneCIndex = 0 // localInputArea
	tvP.paneDIndex = 1 // aiInputArea
	tvP.paneBIndex = 2 // aiOutputView
	tvP.paneAIndex = 3 // localOutputView
	tvP.currentFocusIndex = tvP.paneCIndex

	if len(tvP.leftScreens) > 0 {
		tvP.setScreen(0, true)
	}
	if len(tvP.rightScreens) > 0 {
		tvP.setScreen(0, false)
	}

	if interp := mainApp.GetInterpreter(); interp != nil { //
		if tvP.originalStdout == nil {
			tvP.originalStdout = interp.Stdout()
		}
		interp.SetStdout(scriptOutputScreen)
		TuiPrintf("[TUI_INIT] Interpreter stdout redirected to ScriptOut screen.")
	}

	if initialScriptPath != "" {
		TuiPrintf("[TUI_INIT] Executing initial TUI script: %s", initialScriptPath)
		originalActivityText := tvP.initialActivityText
		tvP.initialActivityText = fmt.Sprintf("Running: %s...", filepath.Base(initialScriptPath))
		tvP.updateStatusText()
		err := mainApp.ExecuteScriptFile(context.Background(), initialScriptPath)

		if err != nil {
			errMsg := fmt.Sprintf("Initial script error: %s", err.Error())
			TuiPrintf("[TUI_INIT] %s (%s)", errMsg, initialScriptPath)
			logError("Initial script execution error", "script", initialScriptPath, "error", err)
			if scriptOutputScreen != nil {
				scriptOutputScreen.Write([]byte("[red]" + EscapeTviewTags(errMsg) + "[-]\n"))
			}
			tvP.initialActivityText = fmt.Sprintf("[red]Script Error: %s[-]", filepath.Base(initialScriptPath))
		} else {
			TuiPrintf("[TUI_INIT] Initial script completed: %s", initialScriptPath)
			logInfo("Initial script completed.", "script", initialScriptPath)
			if scriptOutputScreen != nil {
				scriptOutputScreen.Write([]byte("[green]Initial script completed.[-]\n"))
			}
			tvP.initialActivityText = fmt.Sprintf("Finished: %s. %s", filepath.Base(initialScriptPath), strings.TrimSpace(originalActivityText))
		}
		tvP.updateStatusText()
	}

	tvP.grid = tview.NewGrid().
		SetRows(0, 5, 1).SetColumns(0, 0).SetBorders(false).SetGap(0, 0).
		AddItem(tvP.localOutputView, 0, 0, 1, 1, 0, 0, false).
		AddItem(tvP.aiOutputView, 0, 1, 1, 1, 0, 0, false).
		AddItem(tvP.localInputArea, 1, 0, 1, 1, 0, 100, true).
		AddItem(tvP.aiInputArea, 1, 1, 1, 1, 0, 100, false).
		AddItem(tvP.statusBar, 2, 0, 1, 2, 0, 0, false)
	TuiPrintf("[TUI_INIT] Grid layout configured.")

	keyHandle := func(event *tcell.EventKey) *tcell.EventKey {
		var activeScreener PrimitiveScreener
		// var targetPaneIsLeft bool // Not strictly needed here anymore for general event dispatch
		var paneToCheck *tview.Pages
		var focusedComponent tview.Primitive

		if tvP.currentFocusIndex < 0 || tvP.currentFocusIndex >= tvP.numFocusablePrimitives {
			TuiPrintf("[KEY_HANDLE_ERROR] currentFocusIndex %d out of bounds [0-%d]. Resetting.", tvP.currentFocusIndex, tvP.numFocusablePrimitives-1)
			if tvP.numFocusablePrimitives > 0 {
				tvP.currentFocusIndex = 0 // Default to first focusable item
			} else {
				return event // No focusable items, let tview handle if it can
			}
		}
		focusedComponent = tvP.focusablePrimitives[tvP.currentFocusIndex]

		// Determine which pane (Pages primitive) is notionally active based on current focus
		// This helps in finding the activeScreener
		targetPaneIsLeftForScreenerLookup := false
		if focusedComponent == tvP.localInputArea || focusedComponent == tvP.localOutputView {
			paneToCheck = tvP.localOutputView
			targetPaneIsLeftForScreenerLookup = true
		} else if focusedComponent == tvP.aiInputArea || focusedComponent == tvP.aiOutputView {
			paneToCheck = tvP.aiOutputView
			targetPaneIsLeftForScreenerLookup = false
		}

		if paneToCheck != nil {
			_, primOnPage := paneToCheck.GetFrontPage()
			if primOnPage != nil {
				activeScreener, _ = tvP.getScreenerFromPrimitive(primOnPage, targetPaneIsLeftForScreenerLookup)
			}
		}

		shouldLogFullKeyEvent := true
		switch event.Key() {
		case tcell.KeyTab, tcell.KeyBacktab:
			shouldLogFullKeyEvent = false
		}
		// if event.Key() == tcell.KeyRune { // Optionally suppress for all typed runes
		// 	shouldLogFullKeyEvent = false
		// }

		if shouldLogFullKeyEvent {
			activeScreenerName := "None"
			if activeScreener != nil {
				activeScreenerName = activeScreener.Name()
			}
			actualTviewFocusName := "nil"
			if actualFocusPrim := tvP.tviewApp.GetFocus(); actualFocusPrim != nil {
				actualTviewFocusName = fmt.Sprintf("%T", actualFocusPrim)
			}
			focusedCompName := "nil"
			if focusedComponent != nil {
				focusedCompName = fmt.Sprintf("%T", focusedComponent)
			}
			TuiPrintf("[KEY_HANDLE] Key: %s, Rune: %c, Mod: %v (FocusIndex: %d [%s], Screener: %s, TviewFocus: %s)",
				event.Name(), event.Rune(), event.Modifiers(),
				tvP.currentFocusIndex, focusedCompName, activeScreenerName, actualTviewFocusName)
		}

		if event.Key() == tcell.KeyCtrlQ {
			TuiPrintf("[KEY_HANDLE] Ctrl+Q pressed. Stopping app.")
			if tvP.tviewApp != nil {
				tvP.tviewApp.Stop()
			}
			return nil
		}

		if activeScreener != nil {
			if handler := activeScreener.InputHandler(); handler != nil {
				TuiPrintf("[KEY_HANDLE] Passing event to screener '%s' InputHandler", activeScreener.Name())
				returnedEvent := handler(event, func(p tview.Primitive) {
					TuiPrintf("[KEY_HANDLE] Screener '%s' requests focus on %T", activeScreener.Name(), p)
					if tvP.tviewApp != nil {
						tvP.tviewApp.SetFocus(p)
					}
				})
				if returnedEvent == nil {
					TuiPrintf("[KEY_HANDLE] Event consumed by screener '%s'", activeScreener.Name())
					return nil
				}
			}
		}

		switch event.Key() {
		case tcell.KeyTab:
			// TuiPrintf("[KEY_HANDLE] Tab pressed.") // dFocus logs more details
			tvP.dFocus(1)
			return nil
		case tcell.KeyBacktab:
			// TuiPrintf("[KEY_HANDLE] Shift+Tab pressed.") // dFocus logs more details
			tvP.dFocus(-1)
			return nil
		case tcell.KeyRune:
			if event.Rune() == '?' {
				isInputField := false
				currentFocus := tvP.tviewApp.GetFocus()
				if _, ok := currentFocus.(*tview.TextArea); ok {
					isInputField = true
				}
				if _, ok := currentFocus.(*tview.InputField); ok { // tview.InputField also exists
					isInputField = true
				}

				if !isInputField || currentFocus == tvP.localOutputView || currentFocus == tvP.aiOutputView {
					TuiPrintf("[KEY_HANDLE] '?' pressed (global). Switching to help on left.")
					if tvP.leftScreens != nil && tvP.helpScreenIndex >= 0 && tvP.helpScreenIndex < len(tvP.leftScreens) {
						tvP.setScreen(tvP.helpScreenIndex, true)
						// Consider focusing the help screen if it's interactive, or Pane A (localOutputView)
						if tvP.localOutputView != nil { // Ensure pane A (localOutputView) gets tview focus
							tvP.currentFocusIndex = tvP.paneAIndex // Update internal focus tracking
							tvP.tviewApp.SetFocus(tvP.localOutputView)
						}
					} else {
						TuiPrintf("[KEY_HANDLE] Help screen index invalid or leftScreens nil for '?' key.")
					}
					return nil
				}
			}
		case tcell.KeyCtrlB:
			TuiPrintf("[KEY_HANDLE] Ctrl+B pressed. Prev screen left.")
			tvP.nextScreen(-1, true)
			return nil
		case tcell.KeyCtrlN:
			TuiPrintf("[KEY_HANDLE] Ctrl+N pressed. Next screen left.")
			tvP.nextScreen(1, true)
			return nil
		case tcell.KeyCtrlP:
			TuiPrintf("[KEY_HANDLE] Ctrl+P pressed. Prev screen right.")
			tvP.nextScreen(-1, false)
			return nil
		case tcell.KeyCtrlF:
			TuiPrintf("[KEY_HANDLE] Ctrl+F pressed. Next screen right.")
			tvP.nextScreen(1, false)
			return nil
		}

		// The problematic block trying to use tview.MouseHandler has been removed.
		// If the event reaches here, it means it was not consumed by a screener's InputHandler
		// and was not a global hotkey. We return the event to allow tview
		// to dispatch it to the InputHandler of the currently focused primitive.
		// For example, a tview.TextArea will handle character input this way.

		if shouldLogFullKeyEvent {
			TuiPrintf("[KEY_HANDLE] Event not handled by custom logic, passing to tview for default dispatch (%s).", FormatEventKeyForLogging(event))
		}
		return event
	}

	if tvP.tviewApp != nil {
		tvP.tviewApp.SetInputCapture(keyHandle)
	}
	TuiPrintf("[TUI_INIT] Global InputCapture function set.")

	tvP.dFocus(0)
	TuiPrintf("[TUI_INIT] Initial dFocus(0) called.")

	TuiPrintf("[TUI_INIT] Starting tview event loop (app.Run())...")
	var runErr error
	if tvP.tviewApp != nil && tvP.grid != nil {
		runErr = tvP.tviewApp.SetRoot(tvP.grid, true).EnableMouse(true).Run()
	} else {
		runErr = fmt.Errorf("tviewApp or grid was nil before Run()")
		logError("Cannot start TUI", "error", runErr.Error())
	}

	if runErr != nil {
		logError("tview.Application.Run() exited with error", "error", runErr)
		fmt.Fprintf(os.Stderr, "tview.Application.Run() exited with error: %v\n", runErr)
		if interp := mainApp.GetInterpreter(); interp != nil && tvP.originalStdout != nil { //
			interp.SetStdout(tvP.originalStdout)
		}
		return fmt.Errorf("tview application run error: %w", runErr)
	}

	logInfo("tview.Application.Run() exited normally.")
	if interp := mainApp.GetInterpreter(); interp != nil && tvP.originalStdout != nil { //
		interp.SetStdout(tvP.originalStdout)
	}
	return nil
}
