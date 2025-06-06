// NeuroScript Version: 0.4.0
// File version: 0.3.18 (with Ctrl+C copy, Ctrl+N fix, and reduced TuiPrintf logging)
// Description: Main TUI entry point. Ctrl+Q to quit. Ctrl+C to copy.
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

	"github.com/atotto/clipboard" // For Ctrl+C copy functionality
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
// (This will be redirected to the in-TUI debug pane by LogToDebugScreen if called from there)
func TuiPrintf(format string, a ...interface{}) {
	// msg := fmt.Sprintf(format, a...)
	// if debugFile != nil {
	// 	// Ensure the message ends with a newline for file logging
	// 	if !strings.HasSuffix(msg, "\n") {
	// 		msg += "\n"
	// 	}
	// 	fmt.Fprint(debugFile, msg)
	// } else {
	// 	fmt.Println(msg) // Fallback to stdout if debugFile not initialized
	// }
}

// StartTviewTUI initializes and runs the tview-based Text User Interface.
func StartTviewTUI(mainApp *App, initialScriptPath string) error {

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
					TuiPrintf("[AI_INPUT] Enter pressed. No active chat session. Text: %s", text) // Goes to debug file or stdout
					logError("Cannot send chat: No active chat session.", "text", text)
					// Status bar not primary, so removing this visual error, rely on logs.
					// if tvP.tviewApp != nil && tvP.statusBar != nil {
					// 	tvP.tviewApp.QueueUpdateDraw(func() {
					// 		tvP.statusBar.SetText("[red]No active chat. Select worker from AIWM (Tab to Pane A, Enter or 'c') then type here.[-]")
					// 	})
					// }
					return nil
				}
				TuiPrintf("[AI_INPUT] Enter pressed. Active session: %s. Sending: %s", activeChatSession.SessionID, text) // Debug file/stdout
				logInfo("Sending chat message from AI Input", "text", text, "sessionID", activeChatSession.SessionID)

				go func(msgToSend string, sessID string) {
					ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
					defer cancel()
					TuiPrintf("[AI_INPUT_GOROUTINE] Sending to session %s: %s", sessID, msgToSend) // Debug file/stdout
					_, err := tvP.app.SendChatMessageToActiveSession(ctx, msgToSend)

					if tvP.tviewApp != nil {
						tvP.tviewApp.QueueUpdateDraw(func() {
							TuiPrintf("[AI_INPUT_UI_UPDATE] Updating screen for session %s after send.", sessID) // Debug file/stdout
							if screen, ok := tvP.chatScreenMap[sessID]; ok {
								if err != nil {
									errMsg := fmt.Sprintf("Error sending/processing chat: %v", err)
									TuiPrintf("[AI_INPUT_UI_UPDATE] %s (Session: %s)", errMsg, sessID) // Debug file/stdout
									logError("Error sending/processing chat message", "sessionID", sessID, "error", err)
									if screen.textView != nil {
										fmt.Fprintln(screen.textView, "[red]"+EscapeTviewTags(errMsg)+"[-]")
										screen.textView.ScrollToEnd()
									}
								}
								screen.UpdateConversation()
							} else {
								TuiPrintf("[AI_INPUT_UI_UPDATE] Chat screen NOT FOUND in map for session ID: %s", sessID) // Debug file/stdout
								logError("Chat screen not found in map for session ID", "sessionID", sessID)
							}
							tvP.updateStatusText() // Update status bar with screen names
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
	tvP.statusBar = tview.NewTextView().SetDynamicColors(true) // Status bar still used for screen names

	TuiPrintf("[TUI_INIT] UI Primitives (like debugScreen, statusbar) initialized.") // Debug file/stdout

	helpStaticScreen := NewStaticPrimitiveScreen("Help", "Help", helpText)
	aiwmScreen := NewAIWMStatusScreen(tvP.app)
	aiwmStringScreen := NewAIWMStringScreen(tvP.app)

	tvP.addScreen(scriptOutputScreen, true)
	tvP.addScreen(aiwmScreen, true)
	tvP.addScreen(helpStaticScreen, true)
	tvP.helpScreenIndex = 2

	tvP.addScreen(aiwmStringScreen, false) // Add the new AIWMStringScreen to the right pane
	tvP.addScreen(tvP.debugScreen, false)  // DebugLog screen
	tvP.addScreen(NewStaticPrimitiveScreen("HelpRight", "Help (Right)", helpText), false)

	tvP.focusablePrimitives = []tview.Primitive{
		tvP.localInputArea, tvP.aiInputArea, tvP.aiOutputView, tvP.localOutputView,
	}
	tvP.numFocusablePrimitives = len(tvP.focusablePrimitives)
	tvP.paneCIndex = 0                     // localInputArea
	tvP.paneDIndex = 1                     // aiInputArea
	tvP.paneBIndex = 2                     // aiOutputView
	tvP.paneAIndex = 3                     // localOutputView
	tvP.currentFocusIndex = tvP.paneCIndex // Default focus to Local Input

	if len(tvP.leftScreens) > 0 {
		tvP.setScreen(0, true)
	}
	if len(tvP.rightScreens) > 0 {
		tvP.setScreen(0, false)
	}

	if interp := mainApp.GetInterpreter(); interp != nil {
		if tvP.originalStdout == nil {
			tvP.originalStdout = interp.Stdout()
		}
		interp.SetStdout(scriptOutputScreen)
		TuiPrintf("[TUI_INIT] Interpreter stdout redirected to ScriptOut screen.") // Debug file/stdout
	}

	if initialScriptPath != "" {
		TuiPrintf("[TUI_INIT] Executing initial TUI script: %s", initialScriptPath) // Debug file/stdout
		originalActivityText := tvP.initialActivityText
		tvP.initialActivityText = fmt.Sprintf("Running: %s...", filepath.Base(initialScriptPath))
		tvP.updateStatusText()
		err := mainApp.ExecuteScriptFile(context.Background(), initialScriptPath)

		if err != nil {
			errMsg := fmt.Sprintf("Initial script error: %s", err.Error())
			TuiPrintf("[TUI_INIT] %s (%s)", errMsg, initialScriptPath) // Debug file/stdout
			logError("Initial script execution error", "script", initialScriptPath, "error", err)
			if scriptOutputScreen != nil {
				scriptOutputScreen.Write([]byte("[red]" + EscapeTviewTags(errMsg) + "[-]\n"))
			}
			tvP.initialActivityText = fmt.Sprintf("[red]Script Error: %s[-]", filepath.Base(initialScriptPath))
		} else {
			TuiPrintf("[TUI_INIT] Initial script completed: %s", initialScriptPath) // Debug file/stdout
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
		AddItem(tvP.localInputArea, 1, 0, 1, 1, 0, 100, true). // Initially focus localInputArea
		AddItem(tvP.aiInputArea, 1, 1, 1, 1, 0, 100, false).
		AddItem(tvP.statusBar, 2, 0, 1, 2, 0, 0, false)
	TuiPrintf("[TUI_INIT] Grid layout configured.") // Debug file/stdout

	keyHandle := func(event *tcell.EventKey) *tcell.EventKey {
		var activeScreener PrimitiveScreener
		var paneToCheck *tview.Pages
		var focusedComponent tview.Primitive

		if tvP.currentFocusIndex < 0 || tvP.currentFocusIndex >= tvP.numFocusablePrimitives {
			if tvP.numFocusablePrimitives > 0 {
				tvP.currentFocusIndex = 0
			} else {
				return event
			}
		}
		focusedComponent = tvP.focusablePrimitives[tvP.currentFocusIndex]

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

		// --- Ctrl+C for Copy ---
		if event.Key() == tcell.KeyCtrlC {
			// LogToDebugScreen goes to the TUI debug pane
			tvP.LogToDebugScreen("[KEY_HANDLE] Ctrl+C pressed for copy.")
			var textToCopy string
			var copyErrorEncountered bool = false

			if tvP.currentFocusIndex < 0 || tvP.currentFocusIndex >= tvP.numFocusablePrimitives {
				tvP.LogToDebugScreen("[KEY_HANDLE_COPY] No valid focus target for copy.")
				copyErrorEncountered = true
			} else {
				focusedComponentForCopy := tvP.focusablePrimitives[tvP.currentFocusIndex]
				switch comp := focusedComponentForCopy.(type) {
				case *tview.TextArea:
					textToCopy = comp.GetText()
					tvP.LogToDebugScreen("[KEY_HANDLE_COPY] Preparing to copy from TextArea. Length: %d", len(textToCopy))
				case *tview.Pages:
					_, pagePrimitive := comp.GetFrontPage()
					if textView, ok := pagePrimitive.(*tview.TextView); ok {
						textToCopy = textView.GetText(false)
						tvP.LogToDebugScreen("[KEY_HANDLE_COPY] Preparing to copy from TextView in Pane A/B. Length: %d", len(textToCopy))
					} else {
						paneName := "A/B" // Simplified
						errMsg := fmt.Sprintf("Focused content in Pane %s is not a TextView (type: %T). Cannot copy.", paneName, pagePrimitive)
						tvP.LogToDebugScreen("[KEY_HANDLE_COPY] %s", errMsg)
						copyErrorEncountered = true
					}
				default:
					errMsg := fmt.Sprintf("Copy not supported for focused primitive type: %T", focusedComponentForCopy)
					tvP.LogToDebugScreen("[KEY_HANDLE_COPY] %s", errMsg)
					copyErrorEncountered = true
				}
			}

			if !copyErrorEncountered {
				if textToCopy != "" {
					err := clipboard.WriteAll(textToCopy)
					if err != nil {
						tvP.LogToDebugScreen("[KEY_HANDLE_COPY] Error writing to clipboard: %v", err)
					} else {
						tvP.LogToDebugScreen("[KEY_HANDLE_COPY] Text copied to clipboard successfully (%d chars).", len(textToCopy))
					}
				} else {
					tvP.LogToDebugScreen("[KEY_HANDLE_COPY] Focused pane is empty. Nothing to copy.")
				}
			}
			return nil // Consume Ctrl+C event
		}
		// --- END: Ctrl+C for Copy ---

		if event.Key() == tcell.KeyCtrlQ {
			tvP.LogToDebugScreen("[KEY_HANDLE] Ctrl+Q pressed. Stopping app.") // LogToDebugScreen now goes to TUI debug pane
			if tvP.tviewApp != nil {
				tvP.tviewApp.Stop()
			}
			return nil
		}

		if activeScreener != nil {
			if handler := activeScreener.InputHandler(); handler != nil {
				returnedEvent := handler(event, func(p tview.Primitive) {
					if tvP.tviewApp != nil {
						tvP.tviewApp.SetFocus(p)
					}
				})
				if returnedEvent == nil {
					return nil
				}
			}
		}

		switch event.Key() {
		case tcell.KeyTab:
			tvP.LogToDebugScreen("[KEY_HANDLE] Tab pressed.") // Goes to TUI debug pane
			tvP.dFocus(1)
			return nil
		case tcell.KeyBacktab:
			tvP.LogToDebugScreen("[KEY_HANDLE] Shift+Tab pressed.") // Goes to TUI debug pane
			tvP.dFocus(-1)
			return nil
		case tcell.KeyRune:
			if event.Rune() == '?' {
				isInputField := false
				currentFocus := tvP.tviewApp.GetFocus()
				if _, ok := currentFocus.(*tview.TextArea); ok {
					isInputField = true
				}
				if _, ok := currentFocus.(*tview.InputField); ok {
					isInputField = true
				}
				if !isInputField || currentFocus == tvP.localOutputView || currentFocus == tvP.aiOutputView {
					tvP.LogToDebugScreen("[KEY_HANDLE] '?' pressed (global). Switching to help on left.") // Goes to TUI debug pane
					if tvP.leftScreens != nil && tvP.helpScreenIndex >= 0 && tvP.helpScreenIndex < len(tvP.leftScreens) {
						tvP.setScreen(tvP.helpScreenIndex, true)
						if tvP.localOutputView != nil {
							tvP.currentFocusIndex = tvP.paneAIndex
							tvP.tviewApp.SetFocus(tvP.localOutputView)
						}
					} else {
						tvP.LogToDebugScreen("[KEY_HANDLE] Help screen index invalid or leftScreens nil for '?' key.") // Goes to TUI debug pane
					}
					return nil
				}
			}
		case tcell.KeyCtrlB: // Next screen left (as per your current 0.3.18 code, design doc might differ for direction)
			tvP.LogToDebugScreen("[KEY_HANDLE] Ctrl+B pressed. Next screen left.") // Goes to TUI debug pane
			tvP.nextScreen(1, true)                                                // Note: Your file had (1, true), design doc implied cycling. Keep as per your latest code.
			return nil
		case tcell.KeyCtrlN: // Next screen right (corrected behavior)
			tvP.LogToDebugScreen("[KEY_HANDLE] Ctrl+N pressed. Next screen right.") // Goes to TUI debug pane
			tvP.nextScreen(1, false)
			return nil
		case tcell.KeyCtrlP: // Added from my previous suggestion for completeness, if you want Prev Right
			tvP.LogToDebugScreen("[KEY_HANDLE] Ctrl+P pressed. Prev screen right.") // Goes to TUI debug pane
			tvP.nextScreen(-1, false)
			return nil
		case tcell.KeyCtrlF: // Added from my previous suggestion for completeness, if you want Next Right (or if Ctrl+N was meant for something else)
			tvP.LogToDebugScreen("[KEY_HANDLE] Ctrl+F pressed. Next screen right.") // Goes to TUI debug pane
			tvP.nextScreen(1, false)                                                // This is same as Ctrl+N now, adjust if needed
			return nil
		}

		// Minimal logging for unhandled keys by global logic if desired
		// tvP.LogToDebugScreen("[KEY_HANDLE] Event not handled by custom global logic, passing to tview: %s", event.Name())
		return event
	}

	if tvP.tviewApp != nil {
		tvP.tviewApp.SetInputCapture(keyHandle)
	}
	TuiPrintf("[TUI_INIT] Global InputCapture function set.") // Debug file/stdout

	tvP.dFocus(0)                                     // Set initial focus
	TuiPrintf("[TUI_INIT] Initial dFocus(0) called.") // Debug file/stdout

	TuiPrintf("[TUI_INIT] Starting tview event loop (app.Run())...") // Debug file/stdout
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
		if interp := mainApp.GetInterpreter(); interp != nil && tvP.originalStdout != nil {
			interp.SetStdout(tvP.originalStdout)
		}
		return fmt.Errorf("tview application run error: %w", runErr)
	}

	logInfo("tview.Application.Run() exited normally.")
	if interp := mainApp.GetInterpreter(); interp != nil && tvP.originalStdout != nil {
		interp.SetStdout(tvP.originalStdout)
	}
	return nil
}

// Ensure EscapeTviewTags is available if used by any TUI components.
// If it's in tui_utils.go, it would be imported or part of the same package.
// For this complete file, if it's not imported, you might need it:
/*
func EscapeTviewTags(s string) string {
	s = strings.ReplaceAll(s, "[", "[[")
	return s
}
*/

// Ensure helpText is defined if NewStaticPrimitiveScreen("Help", "Help", helpText) is used.
// It might be in tui_screens.go. For a self-contained example, it could be here:
/*
var helpText = `[green]Navigation:[white]
... (your help text content) ...
`
*/
