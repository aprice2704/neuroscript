// NeuroScript Version: 0.4.0
// File version: 0.1.14
// filename: pkg/neurogo/tview_tui.go
// nlines: 250 // Approximate
// risk_rating: MEDIUM // Changed focus update mechanism

package neurogo

import (
	"fmt"

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
}

var nprims, Aidx, Bidx, Cidx, Didx int

func getPrimitiveName(p tview.Primitive, tvP *tviewAppPointers) string {
	if p == nil {
		if tvP.focusablePrimitives != nil && len(tvP.focusablePrimitives) > 0 &&
			tvP.currentFocusIndex >= 0 && tvP.currentFocusIndex < len(tvP.focusablePrimitives) {
			p = tvP.focusablePrimitives[tvP.currentFocusIndex]
		} else {
			return "Unknown (no focusable)"
		}
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
	return "Unnamed"
}

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

	logInfo("StartTviewTUI entered (v0.1.14)")
	if mainApp == nil {
		fmt.Println("CRITICAL ERROR in StartTviewTUI: mainApp parameter is nil.")
		return fmt.Errorf("mainApp instance cannot be nil")
	}

	tvApp := tview.NewApplication()
	tvP := &tviewAppPointers{tviewApp: tvApp, app: mainApp}
	if mainApp.tui == nil {
		mainApp.tui = tvP
	}
	logDebug("tview.Application and tviewAppPointers created.")

	tvP.localOutputView = tview.NewTextView().SetDynamicColors(true).SetScrollable(true).SetRegions(true)
	tvP.localOutputView.SetChangedFunc(func() { tvP.tviewApp.Draw() }) // Explicit Draw on change
	tvP.localOutputView.SetBorder(true).SetTitle("A: Local Output")

	tvP.aiOutputView = tview.NewTextView().SetDynamicColors(true).SetScrollable(true).SetRegions(true)
	tvP.aiOutputView.SetChangedFunc(func() { tvP.tviewApp.Draw() }) // Explicit Draw on change
	tvP.aiOutputView.SetBorder(true).SetTitle("B: AI Output")

	tvP.localInputArea = tview.NewTextArea().SetWrap(false).SetWordWrap(false)
	tvP.localInputArea.SetBorder(true).SetTitle("C: Local Input")

	tvP.aiInputArea = tview.NewTextArea().SetWrap(true).SetWordWrap(true)
	tvP.aiInputArea.SetBorder(true).SetTitle("D: AI Input")

	tvP.initialActivityText = "NeuroScript TUI Ready"
	if initialScriptPath != "" && mainApp.Config.StartupScript == initialScriptPath {
		tvP.initialActivityText = fmt.Sprintf("Script: %s | Ready", initialScriptPath)
	}
	tvP.statusBar = tview.NewTextView().SetDynamicColors(true)
	logDebug("UI Primitives created.")

	tvP.focusablePrimitives = []tview.Primitive{
		tvP.localInputArea, tvP.aiInputArea, tvP.aiOutputView, tvP.localOutputView,
	}
	nprims = len(tvP.focusablePrimitives)
	Aidx = 3
	Bidx = 2
	Cidx = 0
	Didx = 1
	tvP.currentFocusIndex = 0
	logDebug("Focusable primitives defined.")

	tvP.updateStatusText(tvP.initialActivityText, getPrimitiveName(tvP.focusablePrimitives[tvP.currentFocusIndex], tvP))
	logDebug("Initial status bar text set directly.")

	tvP.grid = tview.NewGrid().
		SetRows(0, 5, 1).SetColumns(0, 0).SetBorders(true).SetGap(0, 1)
	tvP.grid.AddItem(tvP.localOutputView, 0, 0, 1, 1, 0, 0, false).
		AddItem(tvP.aiOutputView, 0, 1, 1, 1, 0, 0, false).
		AddItem(tvP.localInputArea, 1, 0, 1, 1, 0, 30, true).
		AddItem(tvP.aiInputArea, 1, 1, 1, 1, 0, 30, false).
		AddItem(tvP.statusBar, 2, 0, 1, 2, 0, 0, false)
	logDebug("Grid layout configured.")

	if mainApp.GetInterpreter() != nil {
		stdoutWriter := &tviewWriter{app: tvP.tviewApp, textView: tvP.localOutputView}
		if mainApp.GetInterpreter().Stdout() != nil {
			mainApp.originalStdout = mainApp.GetInterpreter().Stdout()
		}
		mainApp.GetInterpreter().SetStdout(stdoutWriter)
		logDebug("Interpreter Stdout redirected.")
	}

	dFocus := func(df int) {
		oldFocus := tvP.focusablePrimitives[tvP.currentFocusIndex]
		tvP.currentFocusIndex = (tvP.currentFocusIndex + df) % nprims
		if tvP.currentFocusIndex < 0 {
			tvP.currentFocusIndex += nprims
		}
		nextFocus := tvP.focusablePrimitives[tvP.currentFocusIndex]
		logDebug("Tab: Queuing SetFocus", "targetPrim", getPrimitiveName(nextFocus, tvP))
		tvP.updateStatusText(tvP.initialActivityText, getPrimitiveName(nextFocus, tvP))
		switch v := nextFocus.(type) {
		case *tview.TextView:
			v.SetBorder(true)
			v.SetBorderColor(tcell.ColorRed)
		case *tview.TextArea:
			v.SetBorder(true)
			v.SetBorderColor(tcell.ColorRed)
		}
		switch v := oldFocus.(type) {
		case *tview.TextView:
			v.SetBorder(false)
		case *tview.TextArea:
			v.SetBorder(false)
		}
		tvP.tviewApp.SetFocus(nextFocus)
	}

	keyHandle := func(event *tcell.EventKey) *tcell.EventKey {
		logDebug("InputCapture", "keyName", event.Name(), "key", event.Key(), "rune", event.Rune())
		switch event.Key() {
		case tcell.KeyCtrlC:
			logInfo("Ctrl-C pressed, stopping application.")
			tvP.tviewApp.Stop()
			return nil
		case tcell.KeyTab:
			dFocus(1)
			fmt.Printf("> ")
			return nil
		case tcell.KeyBacktab:
			dFocus(-1)
			fmt.Printf("< ")
			return nil
		}
		return event
	}

	tvP.tviewApp.SetInputCapture(keyHandle)

	logDebug("Global InputCapture function set.")

	logInfo("Setting root primitive and starting tview event loop...")
	if err := tvP.tviewApp.SetRoot(tvP.grid, true).SetFocus(tvP.localInputArea).Run(); err != nil {
		logError("tview.Application.Run() exited with error", "error", err)
		if mainApp.Log == nil {
			fmt.Printf("FATAL: tview.Application.Run() error: %v\n", err)
		}
		return fmt.Errorf("tview application run error: %w", err)
	}

	logInfo("tview.Application.Run() exited normally.")
	if mainApp.GetInterpreter() != nil && mainApp.originalStdout != nil {
		mainApp.GetInterpreter().SetStdout(mainApp.originalStdout)
		logInfo("Restored interpreter's original stdout.")
	}
	return nil
}

type tviewWriter struct {
	app      *tview.Application
	textView *tview.TextView
}

func (tw *tviewWriter) Write(p []byte) (n int, err error) {
	n, err = tw.textView.Write(p)
	if tw.app != nil {
		tw.app.QueueUpdateDraw(func() {})
	}
	return
}

func (tvP *tviewAppPointers) updateStatusText(activity string, focusedPrimitiveName string) {
	statusText := activity
	if focusedPrimitiveName != "" {
		statusText = fmt.Sprintf("%s | Focus: %s", activity, focusedPrimitiveName)
	}
	statusText = fmt.Sprintf("%s | Tab: Next | Shift-Tab: Prev | Ctrl-C: Quit", statusText)

	if tvP.statusBar != nil {
		tvP.statusBar.SetText(statusText)
	}
}
