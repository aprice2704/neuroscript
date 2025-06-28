// NeuroScript Version: 0.4.0
// File version: 0.4.0
// Description: Updated to use the new load/run protocol for initial script execution.
// filename: pkg/neurogo/tview_tui.go
package neurogo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/atotto/clipboard"
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
	// Logging via this function has been commented out to reduce noise.
	// Use mainApp.Log for structured logging instead.
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
					logError("Cannot send chat: No active chat session.", "text", text)
					return nil
				}
				logInfo("Sending chat message from AI Input", "text", text, "sessionID", activeChatSession.SessionID)

				go func(msgToSend string, sessID string) {
					ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
					defer cancel()
					_, err := tvP.app.SendChatMessageToActiveSession(ctx, msgToSend)

					if tvP.tviewApp != nil {
						tvP.tviewApp.QueueUpdateDraw(func() {
							if screen, ok := tvP.chatScreenMap[sessID]; ok {
								if err != nil {
									errMsg := fmt.Sprintf("Error sending/processing chat: %v", err)
									logError("Error sending/processing chat message", "sessionID", sessID, "error", err)
									if screen.textView != nil {
										fmt.Fprintln(screen.textView, "[red]"+EscapeTviewTags(errMsg)+"[-]")
										screen.textView.ScrollToEnd()
									}
								}
								screen.UpdateConversation()
							} else {
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

	helpStaticScreen := NewStaticPrimitiveScreen("Help", "Help", helpText)
	aiwmScreen := NewAIWMStatusScreen(tvP.app)
	aiwmStringScreen := NewAIWMStringScreen(tvP.app)

	tvP.addScreen(scriptOutputScreen, true)
	tvP.addScreen(aiwmScreen, true)
	tvP.addScreen(helpStaticScreen, true)
	tvP.helpScreenIndex = 2

	tvP.addScreen(aiwmStringScreen, false)
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

	if interp := mainApp.GetInterpreter(); interp != nil {
		if tvP.originalStdout == nil {
			tvP.originalStdout = interp.Stdout()
		}
		interp.SetStdout(scriptOutputScreen)
	}

	if initialScriptPath != "" {
		originalActivityText := tvP.initialActivityText
		tvP.initialActivityText = fmt.Sprintf("Running: %s...", filepath.Base(initialScriptPath))
		tvP.updateStatusText()

		// NEW PROTOCOL: Read file via tool, load content, run main procedure.
		var execErr error
		func() { // Use anonymous function to handle errors cleanly with a single 'err' var
			interpreter := mainApp.Interpreter()
			if interpreter == nil {
				execErr = fmt.Errorf("interpreter is nil")
				return
			}
			// 1. Read file
			filepathArg, err := core.Wrap(initialScriptPath)
			if err != nil {
				execErr = fmt.Errorf("internal error wrapping script path: %w", err)
				return
			}
			toolArgs := map[string]core.Value{"filepath": filepathArg}
			contentValue, err := interpreter.ExecuteTool("TOOL.ReadFile", toolArgs)
			if err != nil {
				execErr = fmt.Errorf("failed to read initial script '%s': %w", initialScriptPath, err)
				return
			}
			scriptContent, ok := core.Unwrap(contentValue).(string)
			if !ok {
				execErr = fmt.Errorf("TOOL.ReadFile did not return a string")
				return
			}
			// 2. Load script
			if _, err := mainApp.LoadScriptString(context.Background(), scriptContent); err != nil {
				execErr = fmt.Errorf("failed to load initial script: %w", err)
				return
			}
			// 3. Run 'main' procedure
			if _, err := mainApp.RunProcedure(context.Background(), "main", nil); err != nil {
				var rErr *core.RuntimeError
				if errors.As(err, &rErr) && rErr.Code == core.ErrorCodeProcNotFound {
					logInfo("Initial script loaded. No 'main' procedure found to execute.", "script", initialScriptPath)
				} else {
					execErr = fmt.Errorf("error running main from initial script: %w", err)
				}
			}
		}()

		if execErr != nil {
			errMsg := fmt.Sprintf("Initial script error: %s", execErr.Error())
			logError("Initial script execution error", "script", initialScriptPath, "error", execErr)
			if scriptOutputScreen != nil {
				scriptOutputScreen.Write([]byte("[red]" + EscapeTviewTags(errMsg) + "[-]\n"))
			}
			tvP.initialActivityText = fmt.Sprintf("[red]Script Error: %s[-]", filepath.Base(initialScriptPath))
		} else {
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

		if event.Key() == tcell.KeyCtrlC {
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
			return nil
		}

		if event.Key() == tcell.KeyCtrlQ {
			tvP.LogToDebugScreen("[KEY_HANDLE] Ctrl+Q pressed. Stopping app.")
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
			tvP.LogToDebugScreen("[KEY_HANDLE] Tab pressed.")
			tvP.dFocus(1)
			return nil
		case tcell.KeyBacktab:
			tvP.LogToDebugScreen("[KEY_HANDLE] Shift+Tab pressed.")
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
					tvP.LogToDebugScreen("[KEY_HANDLE] '?' pressed (global). Switching to help on left.")
					if tvP.leftScreens != nil && tvP.helpScreenIndex >= 0 && tvP.helpScreenIndex < len(tvP.leftScreens) {
						tvP.setScreen(tvP.helpScreenIndex, true)
						if tvP.localOutputView != nil {
							tvP.currentFocusIndex = tvP.paneAIndex
							tvP.tviewApp.SetFocus(tvP.localOutputView)
						}
					} else {
						tvP.LogToDebugScreen("[KEY_HANDLE] Help screen index invalid or leftScreens nil for '?' key.")
					}
					return nil
				}
			}
		case tcell.KeyCtrlB:
			tvP.LogToDebugScreen("[KEY_HANDLE] Ctrl+B pressed. Next screen left.")
			tvP.nextScreen(1, true)
			return nil
		case tcell.KeyCtrlN:
			tvP.LogToDebugScreen("[KEY_HANDLE] Ctrl+N pressed. Next screen right.")
			tvP.nextScreen(1, false)
			return nil
		case tcell.KeyCtrlP:
			tvP.LogToDebugScreen("[KEY_HANDLE] Ctrl+P pressed. Prev screen right.")
			tvP.nextScreen(-1, false)
			return nil
		case tcell.KeyCtrlF:
			tvP.LogToDebugScreen("[KEY_HANDLE] Ctrl+F pressed. Next screen right.")
			tvP.nextScreen(1, false)
			return nil
		}

		return event
	}

	if tvP.tviewApp != nil {
		tvP.tviewApp.SetInputCapture(keyHandle)
	}

	tvP.dFocus(0)

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
