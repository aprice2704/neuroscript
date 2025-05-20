// NeuroScript Version: 0.4.0
// File version: 0.3.4
// Description: Main TUI entry point. Further review of Draw/QueueUpdateDraw calls.
// filename: pkg/neurogo/tview_tui.go
// nlines: 200 // Approximate
// risk_rating: MEDIUM
package neurogo

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// tviewAppPointers holds references to common tview components and application state.
type tviewAppPointers struct {
	tviewApp        *tview.Application
	grid            *tview.Grid
	localOutputView *tview.Pages    // Pane A
	aiOutputView    *tview.Pages    // Pane B
	localInputArea  *tview.TextArea // Pane C
	aiInputArea     *tview.TextArea // Pane D
	statusBar       *tview.TextView

	// Focus management
	focusablePrimitives    []tview.Primitive
	currentFocusIndex      int
	numFocusablePrimitives int
	paneAIndex             int
	paneBIndex             int
	paneCIndex             int
	paneDIndex             int

	app                 *App
	initialActivityText string

	// Screen management
	leftScreens  []PrimitiveScreener
	rightScreens []PrimitiveScreener
	chatScreen   *ChatConversationScreen

	leftShowing     int
	rightShowing    int
	helpScreenIndex int

	originalStdout io.Writer
}

// StartTviewTUI initializes and runs the tview-based Text User Interface.
func StartTviewTUI(mainApp *App, initialScriptPath string) error {
	logInfo := func(msg string, keyvals ...interface{}) {
		if mainApp != nil && mainApp.Log != nil {
			mainApp.Log.Info(msg, keyvals...)
		} else {
			fmt.Printf("INFO: %s %v\n", msg, keyvals)
		}
	}
	logError := func(msg string, keyvals ...interface{}) {
		if mainApp != nil && mainApp.Log != nil {
			mainApp.Log.Error(msg, keyvals...)
		} else {
			fmt.Printf("ERROR: %s %v\n", msg, keyvals)
		}
	}
	logDebug := func(msg string, keyvals ...interface{}) {
		if mainApp != nil && mainApp.Log != nil {
			mainApp.Log.Debug(msg, keyvals...)
		} else {
			fmt.Printf("DEBUG: %s %v\n", msg, keyvals)
		}
	}

	logInfo("StartTviewTUI initializing...")
	if mainApp == nil {
		fmt.Println("CRITICAL ERROR in StartTviewTUI: mainApp parameter is nil.")
		return fmt.Errorf("mainApp instance cannot be nil")
	}

	tvApp := tview.NewApplication()
	tvP := &tviewAppPointers{tviewApp: tvApp, app: mainApp}
	if mainApp.tui == nil {
		mainApp.tui = tvP
	}

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
				if tvP.app.activeChatInstance == nil {
					logError("Cannot send chat: No active chat instance.", "text", text)
					// Update status bar on the main thread
					tvP.tviewApp.QueueUpdateDraw(func() {
						tvP.statusBar.SetText("[red]No active chat. Select worker from AIWM (Ctrl+B > Select > Enter) then type here.[-]")
					})
					return nil
				}
				logInfo("Sending chat message from AI Input", "text", text)
				// Perform the LLM call in a separate goroutine to avoid blocking the TUI thread
				go func(msgToSend string) {
					ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
					defer cancel()
					_, err := tvP.app.SendChatMessageToActiveWorker(ctx, msgToSend)

					// Update UI elements on the main thread using QueueUpdateDraw
					tvP.tviewApp.QueueUpdateDraw(func() {
						if err != nil {
							logError("Error sending/processing chat message", "error", err)
							errorMsg := fmt.Sprintf("\n[red]Error: %s[-]\\n", EscapeTviewTags(err.Error()))
							if tvP.chatScreen != nil && tvP.chatScreen.textView != nil {
								fmt.Fprint(tvP.chatScreen.textView, errorMsg)
								tvP.chatScreen.textView.ScrollToEnd()
							}
						}
						if tvP.chatScreen != nil {
							tvP.chatScreen.UpdateConversation(tvP.app.GetActiveChatHistory())
						}
					})
				}(text)

				tvP.aiInputArea.SetText("", true) // Clear input area immediately
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
	logDebug("UI Primitives initialized.")

	helpStaticScreen := NewStaticPrimitiveScreen("Help", "Help", helpText)
	scriptOutputScreen := NewDynamicOutputScreen("ScriptOut", "Script Output")
	aiwmScreen := NewAIWMStatusScreen(tvP.app)
	tvP.chatScreen = NewChatConversationScreen(tvP.app)

	tvP.addScreen(scriptOutputScreen, true)
	tvP.addScreen(aiwmScreen, true)
	tvP.addScreen(helpStaticScreen, true)
	tvP.helpScreenIndex = 2

	tvP.addScreen(tvP.chatScreen, false)
	tvP.addScreen(helpStaticScreen, false)

	tvP.focusablePrimitives = []tview.Primitive{
		tvP.localInputArea,
		tvP.aiInputArea,
		tvP.aiOutputView,
		tvP.localOutputView,
	}
	tvP.numFocusablePrimitives = len(tvP.focusablePrimitives)
	tvP.paneCIndex = 0
	tvP.paneDIndex = 1
	tvP.paneBIndex = 2
	tvP.paneAIndex = 3
	tvP.currentFocusIndex = tvP.paneCIndex

	// Set initial screens *after* focusablePrimitives is initialized
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
		logInfo("Interpreter stdout redirected to script output screen (Pane A).")
	}

	if initialScriptPath != "" {
		logInfo("Executing initial TUI script...", "script", initialScriptPath)
		originalActivityText := tvP.initialActivityText
		tvP.initialActivityText = fmt.Sprintf("Running: %s...", filepath.Base(initialScriptPath))
		tvP.updateStatusText() // Update status to "Running..."

		// Execute script synchronously for now.
		err := mainApp.ExecuteScriptFile(context.Background(), initialScriptPath)
		scriptOutputScreen.FlushBufferToTextView() // Ensure content from script execution itself is in the TextView immediately

		// !! DIAGNOSTIC: Temporarily comment out the QueueUpdateDraw block !!
		/*
			tvP.tviewApp.QueueUpdateDraw(func() {
				if err != nil {
					errMsg := fmt.Sprintf("[red]Error in initial script '%s': %v[-]\\n", initialScriptPath, err)
					logError("Initial script execution error", "script", initialScriptPath, "error", err)
					fmt.Fprint(scriptOutputScreen, errMsg)
				} else {
					logInfo("Initial script completed.", "script", initialScriptPath)
					fmt.Fprintf(scriptOutputScreen, "[green]Initial script '%s' completed.[-]\\n", initialScriptPath)
				}
				tvP.initialActivityText = fmt.Sprintf("Finished: %s. %s", filepath.Base(initialScriptPath), strings.TrimSpace(originalActivityText))
				tvP.updateStatusText()
				scriptOutputScreen.FlushBufferToTextView()
			})
		*/
		// Synchronous fallback for script completion status for diagnostic purposes:
		if err != nil {
			// errMsg := fmt.Sprintf("[red]Error in initial script '%s': %v[-]\\n", initialScriptPath, err) // Prepared, but not Fprinting to avoid complex UI interaction
			logError("Initial script execution error (synchronous log path)", "script", initialScriptPath, "error", err)
			tvP.initialActivityText = fmt.Sprintf("[red]Script Error: %s[-]", filepath.Base(initialScriptPath))
			tvP.updateStatusText()
		} else {
			logInfo("Initial script completed (synchronous log path).", "script", initialScriptPath)
			tvP.initialActivityText = fmt.Sprintf("Finished: %s. %s", filepath.Base(initialScriptPath), strings.TrimSpace(originalActivityText))
			tvP.updateStatusText()
		}
	}

	tvP.grid = tview.NewGrid().
		SetRows(0, 5, 1).SetColumns(0, 0).SetBorders(false).SetGap(0, 0)
	tvP.grid.AddItem(tvP.localOutputView, 0, 0, 1, 1, 0, 0, false).
		AddItem(tvP.aiOutputView, 0, 1, 1, 1, 0, 0, false).
		AddItem(tvP.localInputArea, 1, 0, 1, 1, 0, 100, true).
		AddItem(tvP.aiInputArea, 1, 1, 1, 1, 0, 100, false).
		AddItem(tvP.statusBar, 2, 0, 1, 2, 0, 0, false)
	logDebug("Grid layout configured.")

	keyHandle := func(event *tcell.EventKey) *tcell.EventKey {
		var activeScreener PrimitiveScreener = nil
		// Check focus on Pages views themselves, not their content primitives directly here,
		// as the content primitives might not be directly in tvP.focusablePrimitives.
		if tvP.localOutputView.HasFocus() {
			_, prim := tvP.localOutputView.GetFrontPage()
			activeScreener, _ = tvP.getScreenerFromPrimitive(prim, true)
		} else if tvP.aiOutputView.HasFocus() {
			_, prim := tvP.aiOutputView.GetFrontPage()
			activeScreener, _ = tvP.getScreenerFromPrimitive(prim, false)
		}

		if activeScreener != nil {
			if handler := activeScreener.InputHandler(); handler != nil {
				returnedEvent := handler(event, func(p tview.Primitive) { tvP.tviewApp.SetFocus(p) })
				if returnedEvent == nil { // Event handled by the screen
					// If chat was started by AIWM screen, update UI.
					// This check + update should be on the main UI thread.
					if _, _, _, chatNowActive := tvP.app.GetActiveChatInstanceDetails(); chatNowActive {
						// No QueueUpdateDraw here if already on main thread (which keyHandle is)
						if tvP.chatScreen != nil {
							chatScreenIdx := tvP.getScreenIndex(tvP.chatScreen, false)
							if chatScreenIdx != -1 {
								currentChatPageName, _ := tvP.aiOutputView.GetFrontPage()
								expectedChatPageName := strconv.Itoa(chatScreenIdx)
								if currentChatPageName != expectedChatPageName {
									tvP.setScreen(chatScreenIdx, false)
								}
								tvP.chatScreen.UpdateConversation(tvP.app.GetActiveChatHistory())
							}
						}
					}
					return nil
				}
			}
		}

		switch event.Key() {
		case tcell.KeyCtrlC:
			logInfo("Ctrl-C pressed, stopping application.")
			tvP.tviewApp.Stop()
			return nil
		case tcell.KeyTab:
			tvP.dFocus(1)
			return nil
		case tcell.KeyBacktab:
			tvP.dFocus(-1)
			return nil
		case tcell.KeyRune:
			if event.Rune() == '?' {
				if tvP.helpScreenIndex >= 0 && tvP.helpScreenIndex < len(tvP.leftScreens) {
					// UI updates from key handlers (main thread) don't strictly need QueueUpdateDraw for themselves
					tvP.setScreen(tvP.helpScreenIndex, true)
				} else {
					logError("Help screen index for left pane is invalid.", "index", tvP.helpScreenIndex)
				}
				return nil
			}
		case tcell.KeyCtrlB: // Cycle Left Pane (A) screens
			tvP.nextScreen(1, true)
			return nil
		case tcell.KeyCtrlN: // Cycle Right Pane (B) screens
			tvP.nextScreen(1, false)
			return nil
		}
		return event
	}
	tvP.tviewApp.SetInputCapture(keyHandle)
	logDebug("Global InputCapture function set.")

	tvP.dFocus(0) // Apply initial focus styles

	logInfo("Starting tview event loop...")
	if err := tvP.tviewApp.SetRoot(tvP.grid, true).Run(); err != nil {
		logError("tview.Application.Run() exited with error", "error", err)
		if interp := mainApp.GetInterpreter(); interp != nil && tvP.originalStdout != nil {
			interp.SetStdout(tvP.originalStdout)
			logInfo("Restored interpreter's original stdout after TUI error.")
		}
		return fmt.Errorf("tview application run error: %w", err)
	}

	logInfo("tview.Application.Run() exited normally.")
	if interp := mainApp.GetInterpreter(); interp != nil && tvP.originalStdout != nil {
		interp.SetStdout(tvP.originalStdout)
		logInfo("Restored interpreter's original stdout after TUI exit.")
	}
	return nil
}

func (tvP *tviewAppPointers) onPanePageChange(pane *tview.Pages) {
	_, currentPrimitive := pane.GetFrontPage()
	if currentPrimitive == nil {
		return
	}

	isLeftPane := (pane == tvP.localOutputView)
	var screener PrimitiveScreener

	if isLeftPane {
		screener, _ = tvP.getScreenerFromPrimitive(currentPrimitive, true)
	} else {
		screener, _ = tvP.getScreenerFromPrimitive(currentPrimitive, false)
	}

	if screener != nil {
		// If the Pages view itself is the one that should have focus in the main cycle...
		if tvP.focusablePrimitives != nil && len(tvP.focusablePrimitives) > tvP.currentFocusIndex && tvP.currentFocusIndex >= 0 {
			if (isLeftPane && tvP.focusablePrimitives[tvP.currentFocusIndex] == tvP.localOutputView) ||
				(!isLeftPane && tvP.focusablePrimitives[tvP.currentFocusIndex] == tvP.aiOutputView) {
				// ...then try to pass focus to the content of the page if it's focusable.
				if screener.IsFocusable() {
					screener.OnFocus(func(p tview.Primitive) {
						tvP.tviewApp.SetFocus(p)
					})
				} else {
					// If the screen content isn't focusable, keep focus on the Pages view itself.
					tvP.tviewApp.SetFocus(pane)
				}
			}
		} else if tvP.app != nil && tvP.app.Log != nil {
			tvP.app.Log.Debug("onPanePageChange: focusablePrimitives not yet initialized or currentFocusIndex out of bounds during page change.")
		}

		// Update screen content
		if cs, ok := screener.(*ChatConversationScreen); ok {
			cs.UpdateConversation(tvP.app.GetActiveChatHistory())
			cs.Primitive() // Update title
		} else if dos, ok := screener.(*DynamicOutputScreen); ok {
			dos.FlushBufferToTextView()
		}
	}
	tvP.updateStatusText()
	// Let tview handle drawing naturally after state changes. No explicit Draw() here.
}
