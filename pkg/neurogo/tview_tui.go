// NeuroScript Version: 0.4.0
// File version: 0.2.1 // Based on your 0.2.0, adding initial script display fix
// filename: pkg/neurogo/tview_tui.go
// nlines: 280 // Approximate
// risk_rating: MEDIUM
// Changes:
// - Integrated DynamicOutputScreen to capture and display initial script output.
// - Initial script (-script flag) is now run synchronously during TUI setup
//   after interpreter's stdout is redirected to DynamicOutputScreen's buffer.
// - DynamicOutputScreen.Contents() provides the buffered output to localOutputView
//   when the screen is first set.
// - No live refresh from DynamicOutputScreen.Write() in this version (deferred).

package neurogo

import (
	"context" // For context.Background()
	"fmt"
	"io"
	"path/filepath" // For filepath.Base()
	"strings"       // For strings.TrimSpace in input handlers, if not already imported in user's original

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type tviewAppPointers struct {
	tviewApp        *tview.Application
	grid            *tview.Grid
	localOutputView *tview.TextView
	aiOutputView    *tview.TextView
	localInputArea  *tview.TextArea
	aiInputArea     *tview.TextArea
	statusBar       *tview.TextView

	focusablePrimitives []tview.Primitive
	currentFocusIndex   int

	app                 *App
	initialActivityText string

	leftScreens  []Screener
	rightScreens []Screener
	leftShowing  int
	rightShowing int
	helpScreen   int // This is an index for tvP.leftScreens

	// localOutputWriter is used by the original v0.2.0.
	// The interpreter's primary stdout will be DynamicOutputScreen.
	// This can remain if other direct writes to localOutputView are needed.
	localOutputWriter *tviewWriter
	originalStdout    io.Writer // To store the interpreter's original stdout
}

func (tvP *tviewAppPointers) addScreen(s Screener, onLeft bool) {
	if onLeft {
		tvP.leftScreens = append(tvP.leftScreens, s)
	} else {
		tvP.rightScreens = append(tvP.rightScreens, s)
	}
}

func (tvP *tviewAppPointers) nextScreen(d int, onLeft bool) {
	screens := tvP.rightScreens
	cur := tvP.rightShowing
	if onLeft {
		screens = tvP.leftScreens
		cur = tvP.leftShowing
	}
	n := len(screens)
	if n < 2 { // If less than 2 screens, no cycling possible
		return
	}
	nxt := posmod(cur+d, n)
	tvP.setScreen(nxt, onLeft)
}

func (tvP *tviewAppPointers) setScreen(sIndex int, onLeft bool) {
	// Logger from mainApp should be used if available, otherwise skip logging from here
	logDebug := func(msg string, keyvals ...interface{}) {}
	if tvP.app != nil && tvP.app.GetLogger() != nil { // Check if logger is available
		logDebug = tvP.app.GetLogger().Debug
	}

	var targetView *tview.TextView
	var screens []Screener
	paneName := "Right"
	if onLeft {
		targetView = tvP.localOutputView
		screens = tvP.leftScreens
		paneName = "Left"
	} else {
		targetView = tvP.aiOutputView
		screens = tvP.rightScreens
	}

	if sIndex < 0 || sIndex >= len(screens) {
		logDebug("setScreen: index out of bounds", "pane", paneName, "index", sIndex, "numScreens", len(screens))
		return
	}
	if targetView == nil {
		logDebug("setScreen: targetView is nil", "pane", paneName)
		return
	}

	activeScreen := screens[sIndex]
	targetView.SetTitle(activeScreen.Title())
	// Key change: For DynamicOutputScreen, Contents() will return its buffered output.
	// For StaticScreen, it returns its static content.
	targetView.SetText(activeScreen.Contents())

	// Scroll to the end for DynamicOutputScreen to see the latest, otherwise to the beginning.
	if _, ok := activeScreen.(*DynamicOutputScreen); ok {
		targetView.ScrollToEnd()
	} else {
		targetView.ScrollToBeginning()
	}

	if onLeft {
		tvP.leftShowing = sIndex
	} else {
		tvP.rightShowing = sIndex
	}
	logDebug("setScreen successful", "screenName", activeScreen.Name(), "index", sIndex, "onLeft", onLeft)
	tvP.updateStatusText() // Assumes updateStatusText is safe to call
}

// Jump around a cycle of numbers, always >=0
func posmod(a, b int) (c int) {
	c = a % b
	if c < 0 {
		c += b
	}
	return c
}

// Global vars from user's v0.2.0 file. These should ideally be part of tvP or managed differently.
var nprims, Aidx, Bidx, Cidx, Didx int

// getPrimitiveName from user's v0.2.0 file.
func getPrimitiveName(p tview.Primitive, tvP *tviewAppPointers) string {
	if p == nil {
		if tvP.focusablePrimitives != nil && len(tvP.focusablePrimitives) > 0 &&
			tvP.currentFocusIndex >= 0 && tvP.currentFocusIndex < len(tvP.focusablePrimitives) {
			p = tvP.focusablePrimitives[tvP.currentFocusIndex]
		} else {
			return "Unknown (no focusable)"
		}
	}
	if p == nil { // Still nil after attempt to get current focus
		return "Unknown (p is nil)"
	}
	switch p {
	case tvP.localInputArea:
		return "C:Local Input"
	case tvP.aiInputArea:
		return "D:AI Input"
	case tvP.aiOutputView:
		return "B:AI Output"
	case tvP.localOutputView:
		return "A:Local Output"
	}
	if titled, ok := p.(interface{ GetTitle() string }); ok {
		name := titled.GetTitle()
		if name != "" {
			return name
		}
	}
	return "Unnamed Primitive" // More descriptive fallback
}

// StartTviewTUI is based on the user's v0.2.0 structure, with minimal changes
// to integrate DynamicOutputScreen for initial script output.
func StartTviewTUI(mainApp *App, initialScriptPath string) error {

	// Logger setup from user's v0.2.0
	logInfo := func(msg string, keyvals ...interface{}) {
		if mainApp != nil && mainApp.Log != nil { // mainApp.Log is the logger instance
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

	logInfo("StartTviewTUI entered (NeuroGo TUI System)") // Updated version string
	if mainApp == nil {
		// Use fmt.Printf directly as logger might not be available if mainApp is nil
		fmt.Println("CRITICAL ERROR in StartTviewTUI: mainApp parameter is nil.")
		return fmt.Errorf("mainApp instance cannot be nil")
	}

	tvApp := tview.NewApplication()
	tvP := &tviewAppPointers{tviewApp: tvApp, app: mainApp}
	if mainApp.tui == nil { // As per user's v0.2.0
		mainApp.tui = tvP
	}
	logDebug("tview.Application and tviewAppPointers created.")

	// --- 1. Create Core UI Components (as in user's v0.2.0) ---
	tvP.localOutputView = tview.NewTextView().SetDynamicColors(true).SetScrollable(true).SetRegions(true)
	// SetChangedFunc can be useful for live updates later, keep if desired.
	// tvP.localOutputView.SetChangedFunc(func() { tvP.tviewApp.Draw() })
	tvP.localOutputView.SetBorder(false).SetTitle("A: Local Output") // Default title, will be overwritten by setScreen

	tvP.aiOutputView = tview.NewTextView().SetDynamicColors(true).SetScrollable(true).SetRegions(true)
	// tvP.aiOutputView.SetChangedFunc(func() { tvP.tviewApp.Draw() })
	tvP.aiOutputView.SetBorder(false).SetTitle("B: AI Output")

	tvP.localInputArea = tview.NewTextArea().SetWrap(false).SetWordWrap(false) // User's settings
	tvP.localInputArea.SetBorder(false).SetTitle("C: Local Input")

	tvP.aiInputArea = tview.NewTextArea().SetWrap(true).SetWordWrap(true) // User's settings
	tvP.aiInputArea.SetBorder(false).SetTitle("D: AI Input")

	tvP.initialActivityText = "NeuroScript TUI Ready"
	if initialScriptPath != "" && mainApp.Config != nil && mainApp.Config.StartupScript == initialScriptPath {
		tvP.initialActivityText = fmt.Sprintf("Script: %s | Ready", initialScriptPath)
	}
	tvP.statusBar = tview.NewTextView().SetDynamicColors(true)
	logDebug("UI Primitives created.")

	// --- 2. Setup DynamicOutputScreen & Redirect Interpreter Output ---
	// This DynamicOutputScreen (from tui_screens.go v0.4.0) only buffers.
	scriptOutputScreen := NewDynamicOutputScreen("Script Output", "A: Script Output")
	// Add it first to tvP.leftScreens so it's at index 0.
	tvP.addScreen(scriptOutputScreen, true)

	if mainApp.GetInterpreter() != nil {
		// Store original stdout if not already stored, then set new stdout
		if mainApp.GetInterpreter().Stdout() != nil && tvP.originalStdout == nil {
			tvP.originalStdout = mainApp.GetInterpreter().Stdout()
		}
		mainApp.GetInterpreter().SetStdout(scriptOutputScreen) // scriptOutputScreen is an io.Writer
		logInfo("Interpreter stdout redirected to DynamicOutputScreen buffer.")
	} else {
		logError("Interpreter is nil; cannot redirect stdout. Script output may go to console.")
	}

	// --- 3. Execute Initial Script (if any) Synchronously ---
	// Output from this script will be captured in scriptOutputScreen.builder.
	if initialScriptPath != "" {
		logInfo("Executing initial script (output to DynamicOutputScreen buffer)", "script", initialScriptPath)
		ctx := context.Background()
		baseScript := filepath.Base(initialScriptPath)  // Requires "path/filepath"
		originalActivityText := tvP.initialActivityText // For restoring status later
		tvP.initialActivityText = fmt.Sprintf("Running: %s...", baseScript)
		// updateStatusText not called here yet; will be called by setScreen

		if err := mainApp.ExecuteScriptFile(ctx, initialScriptPath); err != nil {
			errMsg := fmt.Sprintf("[red]Error executing initial script '%s': %v[-]", initialScriptPath, err)
			logError("Error executing initial script", "script", initialScriptPath, "error", err)
			if _, writeErr := fmt.Fprintln(scriptOutputScreen, errMsg); writeErr != nil {
				logError("Failed to write script execution error to DynamicOutputScreen buffer", "error", writeErr)
			}
		} else {
			successMsg := fmt.Sprintf("[green]Initial script '%s' completed successfully.[-]", initialScriptPath)
			logInfo("Initial script completed successfully", "script", initialScriptPath)
			if _, writeErr := fmt.Fprintln(scriptOutputScreen, successMsg); writeErr != nil {
				logError("Failed to write script success message to DynamicOutputScreen buffer", "error", writeErr)
			}
		}
		tvP.initialActivityText = fmt.Sprintf("Finished: %s. %s", baseScript, strings.TrimSpace(originalActivityText))
		// At this point, scriptOutputScreen.builder contains all output from the initial script.
	}

	// --- 4. Add Other Standard Screens (as in user's v0.2.0) ---
	hs := &StaticScreen{title: "Help", contents: helpText, name: "Help"} // Assumes helpText from tui_screens.go
	tvP.addScreen(hs, true)
	// Correctly determine helpScreen index after adding DynamicOutputScreen first.
	// If DynamicOutputScreen is [0], then Help is [1].
	tvP.helpScreen = 1 // Or loop to find by name if order isn't guaranteed

	blnk := &StaticScreen{title: "Blank", contents: " ", name: "Blank"}
	tvP.addScreen(blnk, true)
	tvP.addScreen(hs, false) // Help screen also on the right, as per user's v0.2.0
	tvP.addScreen(NewAIWMStatusScreen("AIWM", "AIWM Status", tvP.app), true)

	if len(tvP.leftScreens) > 0 {
		tvP.setScreen(0, true) // Activates DynamicOutputScreen (index 0)
	}
	if len(tvP.rightScreens) > 0 {
		tvP.setScreen(0, false) // Activates default right screen (e.g., Help)
	}

	// --- 6. Focusable Primitives & Grid (as in user's v0.2.0) ---
	tvP.focusablePrimitives = []tview.Primitive{
		tvP.localInputArea, tvP.aiInputArea, tvP.aiOutputView, tvP.localOutputView,
	}
	nprims = len(tvP.focusablePrimitives) // User's global var
	// User's global indices, ensure they match the order in focusablePrimitives:
	Cidx = 0                     // localInputArea
	Didx = 1                     // aiInputArea
	Bidx = 2                     // aiOutputView
	Aidx = 3                     // localOutputView
	tvP.currentFocusIndex = Cidx // Start focus on localInputArea

	tvP.updateStatusText() // Call after focus setup and initial screens.
	logDebug("Initial status bar text set directly by updateStatusText.")

	tvP.grid = tview.NewGrid().
		SetRows(0, 5, 1).SetColumns(0, 0).SetBorders(true).SetGap(0, 1) // User's settings
	tvP.grid.AddItem(tvP.localOutputView, 0, 0, 1, 1, 0, 0, false).
		AddItem(tvP.aiOutputView, 0, 1, 1, 1, 0, 0, false).
		AddItem(tvP.localInputArea, 1, 0, 1, 1, 0, 30, true). // Initial focus here
		AddItem(tvP.aiInputArea, 1, 1, 1, 1, 0, 30, false).
		AddItem(tvP.statusBar, 2, 0, 1, 2, 0, 0, false)
	logDebug("Grid layout configured.")

	// tvP.localOutputWriter setup (from user's v0.2.0) - can be kept if needed for direct writes
	// not related to interpreter stdout, otherwise DynamicOutputScreen handles interpreter stdout.
	tvP.localOutputWriter = &tviewWriter{app: tvP.tviewApp, textView: tvP.localOutputView}

	// --- 7. Input Capture Logic (dFocus and keyHandle from user's v0.2.0) ---
	focusInput := tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorYellow)
	blurInput := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite) // Use tcell.StyleDefault

	dFocus := func(df int) { // This is from user's v0.2.0
		oldFocus := tvP.focusablePrimitives[tvP.currentFocusIndex]
		tvP.currentFocusIndex = posmod(tvP.currentFocusIndex+df, nprims)
		nextFocus := tvP.focusablePrimitives[tvP.currentFocusIndex]
		// logDebug("Tab: Queuing SetFocus", "targetPrim", getPrimitiveName(nextFocus, tvP)) // getPrimitiveName might need tvP passed if it's global
		// The getPrimitiveName in user's code already takes tvP.
		logDebug("Focus change requested", "direction", df, "newIndex", tvP.currentFocusIndex, "newFocusItem", getPrimitiveName(nextFocus, tvP))

		// Visual styling for focus (from user's v0.2.0)
		switch v := nextFocus.(type) {
		case *tview.TextView:
			v.SetBackgroundColor(tcell.ColorDarkBlue) // Example focus style
		case *tview.TextArea:
			v.SetTextStyle(focusInput) // Use defined focus style
		}
		switch v := oldFocus.(type) {
		case *tview.TextView:
			v.SetBackgroundColor(tcell.ColorBlack) // Example blur style
		case *tview.TextArea:
			v.SetTextStyle(blurInput) // Use defined blur style
		}
		tvP.tviewApp.SetFocus(nextFocus)
		tvP.updateStatusText() // Update status after focus change
	}

	keyHandle := func(event *tcell.EventKey) *tcell.EventKey { // From user's v0.2.0, adapted
		retEv := event // Default to returning the event
		// logDebug("InputCapture", "keyName", event.Name(), "key", event.Key(), "rune", event.Rune()) // Can be verbose

		switch event.Key() {
		case tcell.KeyCtrlC:
			logInfo("Ctrl-C pressed, stopping application.")
			tvP.tviewApp.Stop()
			retEv = nil // Consume the event
		case tcell.KeyTab:
			dFocus(1)
			retEv = nil
		case tcell.KeyBacktab:
			dFocus(-1)
			retEv = nil
		case tcell.KeyRune:
			if event.Rune() == '?' {
				// Ensure helpScreen index is valid
				if tvP.helpScreen >= 0 && tvP.helpScreen < len(tvP.leftScreens) {
					tvP.setScreen(tvP.helpScreen, true)
				} else {
					logError("Help screen index invalid or help screen not found", "index", tvP.helpScreen)
				}
				retEv = nil
			}
		case tcell.KeyCtrlB:
			tvP.nextScreen(1, true)
			retEv = nil
		case tcell.KeyCtrlN:
			tvP.nextScreen(1, false)
			retEv = nil
		}
		// tvP.updateStatusText() // Called within dFocus and setScreen, might be redundant here unless other keys change state.
		return retEv
	}

	tvP.tviewApp.SetInputCapture(keyHandle)
	logDebug("Global InputCapture function set.")

	// --- 8. Start TUI ---
	logInfo("Setting root primitive and starting tview event loop...")
	// Initial focus is set by AddItem in the grid if `true` is passed as last arg,
	// or explicitly here if needed. Grid AddItem for localInputArea has `true`.
	if err := tvP.tviewApp.SetRoot(tvP.grid, true).SetFocus(tvP.localInputArea).Run(); err != nil {
		logError("tview.Application.Run() exited with error", "error", err)
		if mainApp.Log == nil { // User's v0.2.0 check
			fmt.Printf("FATAL: tview.Application.Run() error: %v\n", err)
		}
		// Restore original stdout on error as well
		if mainApp.GetInterpreter() != nil && tvP.originalStdout != nil { // Use tvP.originalStdout
			mainApp.GetInterpreter().SetStdout(tvP.originalStdout)
			logInfo("Restored interpreter's original stdout after TUI error.")
		}
		return fmt.Errorf("tview application run error: %w", err)
	}

	logInfo("tview.Application.Run() exited normally.")
	if mainApp.GetInterpreter() != nil && tvP.originalStdout != nil { // Use tvP.originalStdout
		mainApp.GetInterpreter().SetStdout(tvP.originalStdout)
		logInfo("Restored interpreter's original stdout.")
	}
	return nil
}

// tviewWriter struct from user's v0.2.0
type tviewWriter struct {
	app      *tview.Application
	textView *tview.TextView
}

// Write method from user's v0.2.0
func (tw *tviewWriter) Write(p []byte) (n int, err error) {
	if tw.textView == nil { // Added nil check for textView
		return 0, fmt.Errorf("tviewWriter.textView is nil")
	}
	n, err = tw.textView.Write(p)
	if tw.app != nil {
		tw.app.QueueUpdateDraw(func() {})
	}
	return
}

// updateStatusText from user's v0.2.0, slightly adapted for safety
func (tvP *tviewAppPointers) updateStatusText() {
	if tvP.statusBar == nil {
		return
	}
	// Ensure logger is available for debug messages from this function
	logDebug := func(msg string, keyvals ...interface{}) {} // No-op by default
	if tvP.app != nil && tvP.app.GetLogger() != nil {
		logDebug = tvP.app.GetLogger().Debug
	}

	// Use user's existing logic for "screens" string
	screens := "no screens yet"
	// Corrected condition: check tvP.rightScreens for the second part
	if len(tvP.leftScreens) > 0 && len(tvP.rightScreens) > 0 &&
		tvP.leftShowing >= 0 && tvP.leftShowing < len(tvP.leftScreens) &&
		tvP.rightShowing >= 0 && tvP.rightShowing < len(tvP.rightScreens) {
		screens = fmt.Sprintf("%s / %s | %s / %s",
			tvP.leftScreens[tvP.leftShowing].Name(), "Local input", // Placeholder "Local input"
			tvP.rightScreens[tvP.rightShowing].Name(), "AI input", // Placeholder "AI input"
		)
	} else if len(tvP.leftScreens) > 0 && tvP.leftShowing >= 0 && tvP.leftShowing < len(tvP.leftScreens) {
		screens = fmt.Sprintf("%s / %s | No right screen active", tvP.leftScreens[tvP.leftShowing].Name(), "Local input")
	} else if len(tvP.rightScreens) > 0 && tvP.rightShowing >= 0 && tvP.rightShowing < len(tvP.rightScreens) {
		screens = fmt.Sprintf("No left screen active | %s / %s", tvP.rightScreens[tvP.rightShowing].Name(), "AI input")
	}

	statusText := fmt.Sprintf("NS TUI: FocusIdx: %d | L: %d/%d | R: %d/%d | Screens: %s | Act: %s",
		tvP.currentFocusIndex,
		tvP.leftShowing, len(tvP.leftScreens), // Displaying 0-based internal index for now
		tvP.rightShowing, len(tvP.rightScreens), // Displaying 0-based internal index
		screens,
		strings.TrimSpace(tvP.initialActivityText), // Include initialActivityText
	)
	logDebug("Updating status bar", "text", statusText)
	tvP.statusBar.SetText(statusText)
}
