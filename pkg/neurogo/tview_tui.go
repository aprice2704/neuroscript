// NeuroScript Version: 0.4.0
// File version: 0.3.14 // Slimmed down, methods moved to tui_methods.go
// Description: Main TUI entry point.
// filename: pkg/neurogo/tview_tui.go
package neurogo

import (
	"context"
	"fmt" // Keep for tvP.originalStdout
	"log"
	"path/filepath"

	// "strconv" // No longer directly used in this file
	"strings" // Keep for strings.TrimSpace
	"time"    // Keep for context.WithTimeout

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// StartTviewTUI initializes and runs the tview-based Text User Interface.
// tviewAppPointers struct and its methods are now in tui_types.go and tui_methods.go.
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
	// tviewAppPointers is now defined in tui_types.go
	tvP := &tviewAppPointers{
		tviewApp:      tvApp,
		app:           mainApp,
		chatScreenMap: make(map[string]*ChatConversationScreen),
		// Other fields will be initialized as components are created
	}
	if mainApp.tui == nil { // Assign our TUI controller to the main app
		mainApp.tui = tvP
	}

	// Create core UI components
	// Methods on tvP will now be called (e.g., tvP.LogToDebugScreen)
	// These methods are defined in tui_methods.go

	// Pass tvP.tviewApp to NewDynamicOutputScreen as it's now required
	tvP.debugScreen = NewDynamicOutputScreen("DebugLog", "Debug Log", tvP.tviewApp)
	scriptOutputScreen := NewDynamicOutputScreen("ScriptOut", "Script Output", tvP.tviewApp)

	tvP.localOutputView = tview.NewPages().SetChangedFunc(func() { tvP.onPanePageChange(tvP.localOutputView) })
	tvP.aiOutputView = tview.NewPages().SetChangedFunc(func() { tvP.onPanePageChange(tvP.aiOutputView) })

	tvP.localInputArea = tview.NewTextArea().SetWrap(false).SetWordWrap(false)
	tvP.localInputArea.SetBorder(false).SetTitle("C: Local Input") // It's useful to see titles for panes

	tvP.aiInputArea = tview.NewTextArea().SetWrap(true).SetWordWrap(true)
	tvP.aiInputArea.SetBorder(false).SetTitle("D: AI Input")

	tvP.aiInputArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter && event.Modifiers() == tcell.ModNone {
			text := strings.TrimSpace(tvP.aiInputArea.GetText())
			if text != "" {
				activeChatSession := tvP.app.GetActiveChatSession()
				if activeChatSession == nil {
					tvP.LogToDebugScreen("[AI_INPUT] Enter pressed. No active chat session. Text: %s", text)
					logError("Cannot send chat: No active chat session.", "text", text)
					if tvP.tviewApp != nil && tvP.statusBar != nil { // Check tviewApp as well for QueueUpdateDraw
						tvP.tviewApp.QueueUpdateDraw(func() {
							tvP.statusBar.SetText("[red]No active chat. Select worker from AIWM (Tab to Pane A, Enter or 'c') then type here.[-]")
						})
					}
					return nil
				}
				tvP.LogToDebugScreen("[AI_INPUT] Enter pressed. Active session: %s. Sending: %s", activeChatSession.SessionID, text)
				logInfo("Sending chat message from AI Input", "text", text, "sessionID", activeChatSession.SessionID)

				go func(msgToSend string, sessID string) {
					ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
					defer cancel()
					tvP.LogToDebugScreen("[AI_INPUT_GOROUTINE] Sending to session %s: %s", sessID, msgToSend)
					_, err := tvP.app.SendChatMessageToActiveSession(ctx, msgToSend)

					if tvP.tviewApp != nil { // Check before QueueUpdateDraw
						tvP.tviewApp.QueueUpdateDraw(func() {
							tvP.LogToDebugScreen("[AI_INPUT_UI_UPDATE] Updating screen for session %s after send.", sessID)
							if screen, ok := tvP.chatScreenMap[sessID]; ok {
								if err != nil {
									errMsg := fmt.Sprintf("Error sending/processing chat: %v", err)
									tvP.LogToDebugScreen("[AI_INPUT_UI_UPDATE] %s (Session: %s)", errMsg, sessID)
									logError("Error sending/processing chat message", "sessionID", sessID, "error", err)
									if screen.textView != nil { // Ensure textView exists
										fmt.Fprintln(screen.textView, "[red]"+EscapeTviewTags(errMsg)+"[-]")
										screen.textView.ScrollToEnd()
									}
								}
								screen.UpdateConversation() // This updates its internal state
							} else {
								tvP.LogToDebugScreen("[AI_INPUT_UI_UPDATE] Chat screen NOT FOUND in map for session ID: %s", sessID)
								logError("Chat screen not found in map for session ID", "sessionID", sessID)
							}
							tvP.updateStatusText() // Method on tvP from tui_methods.go
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

	// This log uses the LogToDebugScreen method on tvP
	tvP.LogToDebugScreen("[TUI_INIT] UI Primitives (like debugScreen, statusbar) initialized.")

	helpStaticScreen := NewStaticPrimitiveScreen("Help", "Help", helpText)
	aiwmScreen := NewAIWMStatusScreen(tvP.app)

	tvP.addScreen(scriptOutputScreen, true)
	tvP.addScreen(aiwmScreen, true)
	tvP.addScreen(helpStaticScreen, true)
	tvP.helpScreenIndex = 2

	tvP.addScreen(tvP.debugScreen, false)
	tvP.addScreen(helpStaticScreen, false)

	tvP.focusablePrimitives = []tview.Primitive{
		tvP.localInputArea, tvP.aiInputArea, tvP.aiOutputView, tvP.localOutputView,
	}
	tvP.numFocusablePrimitives = len(tvP.focusablePrimitives)
	// These indices are informational, dFocus uses the slice directly
	tvP.paneCIndex, tvP.paneDIndex, tvP.paneBIndex, tvP.paneAIndex = 0, 1, 2, 3
	tvP.currentFocusIndex = tvP.paneCIndex // Start focus on Local Input

	if len(tvP.leftScreens) > 0 {
		tvP.setScreen(0, true)
	} // Show ScriptOut on left
	if len(tvP.rightScreens) > 0 {
		tvP.setScreen(0, false)
	} // Show DebugLog on right

	if interp := mainApp.GetInterpreter(); interp != nil {
		if tvP.originalStdout == nil {
			tvP.originalStdout = interp.Stdout()
		}
		interp.SetStdout(scriptOutputScreen)
		tvP.LogToDebugScreen("[TUI_INIT] Interpreter stdout redirected to ScriptOut screen.")
	}

	if initialScriptPath != "" {
		tvP.LogToDebugScreen("[TUI_INIT] Executing initial TUI script: %s", initialScriptPath)
		originalActivityText := tvP.initialActivityText
		tvP.initialActivityText = fmt.Sprintf("Running: %s...", filepath.Base(initialScriptPath))
		tvP.updateStatusText()
		err := mainApp.ExecuteScriptFile(context.Background(), initialScriptPath)

		if err != nil {
			errMsg := fmt.Sprintf("Initial script error: %s", err.Error())
			tvP.LogToDebugScreen("[TUI_INIT] %s (%s)", errMsg, initialScriptPath)
			logError("Initial script execution error", "script", initialScriptPath, "error", err)
			if scriptOutputScreen != nil {
				scriptOutputScreen.Write([]byte("[red]" + EscapeTviewTags(errMsg) + "[-]\n"))
			}
			tvP.initialActivityText = fmt.Sprintf("[red]Script Error: %s[-]", filepath.Base(initialScriptPath))
		} else {
			tvP.LogToDebugScreen("[TUI_INIT] Initial script completed: %s", initialScriptPath)
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
		AddItem(tvP.localInputArea, 1, 0, 1, 1, 0, 100, true). // localInputArea gets initial focus in dFocus(0)
		AddItem(tvP.aiInputArea, 1, 1, 1, 1, 0, 100, false).
		AddItem(tvP.statusBar, 2, 0, 1, 2, 0, 0, false)
	tvP.LogToDebugScreen("[TUI_INIT] Grid layout configured.")

	keyHandle := func(event *tcell.EventKey) *tcell.EventKey {
		var activeScreener PrimitiveScreener
		currentFocusPrimitive := tvP.tviewApp.GetFocus()

		// Determine active screener
		if tvP.localOutputView != nil && (currentFocusPrimitive == tvP.localOutputView || tvP.localOutputView.HasFocus()) {
			_, prim := tvP.localOutputView.GetFrontPage()
			if prim != nil {
				activeScreener, _ = tvP.getScreenerFromPrimitive(prim, true)
			}
		} else if tvP.aiOutputView != nil && (currentFocusPrimitive == tvP.aiOutputView || tvP.aiOutputView.HasFocus()) {
			_, prim := tvP.aiOutputView.GetFrontPage()
			if prim != nil {
				activeScreener, _ = tvP.getScreenerFromPrimitive(prim, false)
			}
		}

		// Suppress logging key events if the debug screen itself is the active screener
		// to prevent the "infinite scroll" effect.
		shouldLogKeyEventToDebugScreen := true
		if activeScreener == tvP.debugScreen {
			shouldLogKeyEventToDebugScreen = false
		}

		if shouldLogKeyEventToDebugScreen {
			tvP.LogToDebugScreen("[KEY_HANDLE] Key: %s, Mod: %v (ActiveScreener: %s)",
				FormatEventKeyForLogging(event),
				event.Modifiers(),
				func() string {
					if activeScreener != nil {
						return activeScreener.Name()
					} else {
						return "None"
					}
				}())
		}

		if activeScreener != nil {
			if handler := activeScreener.InputHandler(); handler != nil {
				if shouldLogKeyEventToDebugScreen {
					tvP.LogToDebugScreen("[KEY_HANDLE] Passing event to screener %s InputHandler", activeScreener.Name())
				}
				returnedEvent := handler(event, func(p tview.Primitive) {
					if shouldLogKeyEventToDebugScreen {
						tvP.LogToDebugScreen("[KEY_HANDLE] Screener %s requests focus on %T", activeScreener.Name(), p)
					}
					if tvP.tviewApp != nil {
						tvP.tviewApp.SetFocus(p)
					}
				})
				if returnedEvent == nil { // Event was consumed by the screener's handler
					if shouldLogKeyEventToDebugScreen {
						tvP.LogToDebugScreen("[KEY_HANDLE] Event consumed by screener %s", activeScreener.Name())
					}
					return nil
				}
			}
		}

		// Global key bindings
		switch event.Key() {
		case tcell.KeyCtrlC:
			if shouldLogKeyEventToDebugScreen {
				tvP.LogToDebugScreen("[KEY_HANDLE] Ctrl+C pressed. Stopping app.")
			}
			if tvP.tviewApp != nil {
				tvP.tviewApp.Stop()
			}
			return nil
		case tcell.KeyTab:
			if shouldLogKeyEventToDebugScreen {
				tvP.LogToDebugScreen("[KEY_HANDLE] Tab pressed.")
			}
			tvP.dFocus(1)
			return nil
		case tcell.KeyBacktab:
			if shouldLogKeyEventToDebugScreen {
				tvP.LogToDebugScreen("[KEY_HANDLE] Shift+Tab pressed.")
			}
			tvP.dFocus(-1)
			return nil
		case tcell.KeyRune:
			if event.Rune() == '?' {
				if shouldLogKeyEventToDebugScreen {
					tvP.LogToDebugScreen("[KEY_HANDLE] '?' pressed. Switching to help.")
				}
				if tvP.helpScreenIndex >= 0 && tvP.helpScreenIndex < len(tvP.leftScreens) {
					tvP.setScreen(tvP.helpScreenIndex, true)
				}
				return nil // '?' is handled
			}
		case tcell.KeyCtrlB:
			if shouldLogKeyEventToDebugScreen {
				tvP.LogToDebugScreen("[KEY_HANDLE] Ctrl+B pressed.")
			}
			tvP.nextScreen(1, true)
			return nil
		case tcell.KeyCtrlN:
			if shouldLogKeyEventToDebugScreen {
				tvP.LogToDebugScreen("[KEY_HANDLE] Ctrl+N pressed.")
			}
			tvP.nextScreen(1, false)
			return nil
		}

		if shouldLogKeyEventToDebugScreen {
			tvP.LogToDebugScreen("[KEY_HANDLE] Event not handled by main keymap, passing through (%s).", FormatEventKeyForLogging(event))
		}
		return event
	}
	if tvP.tviewApp != nil { // Ensure tviewApp exists before setting input capture
		tvP.tviewApp.SetInputCapture(keyHandle)
	}
	tvP.LogToDebugScreen("[TUI_INIT] Global InputCapture function set.")

	tvP.dFocus(0) // Set initial focus and styles
	tvP.LogToDebugScreen("[TUI_INIT] Initial dFocus(0) called.")

	tvP.LogToDebugScreen("[TUI_INIT] Starting tview event loop (app.Run())...")
	var runErr error
	if tvP.tviewApp != nil && tvP.grid != nil { // Ensure root and app are set
		runErr = tvP.tviewApp.SetRoot(tvP.grid, true).Run()
	} else {
		runErr = fmt.Errorf("tviewApp or grid was nil before Run()")
		logError("Cannot start TUI", "error", runErr.Error())
	}

	if runErr != nil {
		tvP.LogToDebugScreen("[TUI_RUN_ERROR] %v", runErr)
		logError("tview.Application.Run() exited with error", "error", runErr)
		if interp := mainApp.GetInterpreter(); interp != nil && tvP.originalStdout != nil {
			interp.SetStdout(tvP.originalStdout)
		}
		return fmt.Errorf("tview application run error: %w", runErr)
	}

	logInfo("tview.Application.Run() exited normally.")
	tvP.LogToDebugScreen("[TUI_EXIT] tview.Application.Run() exited normally.")

	if interp := mainApp.GetInterpreter(); interp != nil && tvP.originalStdout != nil {
		interp.SetStdout(tvP.originalStdout)
	}
	return nil
}
