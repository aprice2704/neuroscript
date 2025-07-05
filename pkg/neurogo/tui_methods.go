// NeuroScript Version: 0.4.0
// File version: 0.1.5
// Corrected logging and focus management logic.
// Description: Contains methods for the tviewAppPointers struct, managing TUI logic.
// filename: pkg/neurogo/tui_methods.go
package neurogo

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// --- Methods from former tui_layout.go ---

// setPrimitiveBackgroundColor sets the background color of a primitive if it's a *tview.Box or embeds it.
func setPrimitiveBackgroundColor(p tview.Primitive, color tcell.Color) {
	if p == nil {
		return
	}
	if box, ok := p.(interface {
		SetBackgroundColor(c tcell.Color) *tview.Box
	}); ok {
		box.SetBackgroundColor(color)
	}
}

// dFocus handles cycling focus among major UI components and updating styles.
func (tvP *tviewAppPointers) dFocus(df int) {
	if tvP.numFocusablePrimitives == 0 {
		tvP.LogToDebugScreen("[DFOCUS] No focusable primitives.")
		return
	}
	defaultInputStyle := tcell.StyleDefault.Background(blurBackground).Foreground(tcell.ColorWhite)
	focusedInputStyle := tcell.StyleDefault.Background(focusBackground).Foreground(tcell.ColorYellow)

	if tvP.currentFocusIndex < 0 || tvP.currentFocusIndex >= len(tvP.focusablePrimitives) {
		tvP.LogToDebugScreen("[DFOCUS] currentFocusIndex %d was out of bounds (0-%d). Resetting to 0.", tvP.currentFocusIndex, tvP.numFocusablePrimitives-1)
		if tvP.numFocusablePrimitives > 0 {
			tvP.currentFocusIndex = 0
		} else {
			return
		}
	}
	oldFocusPrimitive := tvP.focusablePrimitives[tvP.currentFocusIndex]

	if oldPagesView, ok := oldFocusPrimitive.(*tview.Pages); ok {
		_, oldPageContent := oldPagesView.GetFrontPage()
		if oldPageContent != nil {
			isLeftOld := (oldPagesView == tvP.localOutputView)
			if oldScreener, exists := tvP.getScreenerFromPrimitive(oldPageContent, isLeftOld); exists {
				tvP.LogToDebugScreen("[DFOCUS] Calling OnBlur for old focused pane screener: %s (type %T)", oldScreener.Name(), oldScreener)
				oldScreener.OnBlur()
				tvP.LogToDebugScreen("[DFOCUS] OnBlur for old screener %s completed.", oldScreener.Name())
			}
		}
	}

	tvP.currentFocusIndex = posmod(tvP.currentFocusIndex+df, tvP.numFocusablePrimitives)
	newFocusTarget := tvP.focusablePrimitives[tvP.currentFocusIndex]
	tvP.LogToDebugScreen("[DFOCUS] New focus target index: %d, type: %T", tvP.currentFocusIndex, newFocusTarget)

	primitiveToActuallySetFocusOnTview := newFocusTarget

	// Style Input Areas
	if tvP.localInputArea != nil {
		if newFocusTarget == tvP.localInputArea {
			tvP.localInputArea.SetTextStyle(focusedInputStyle)
		} else {
			tvP.localInputArea.SetTextStyle(defaultInputStyle)
		}
	}
	if tvP.aiInputArea != nil {
		if newFocusTarget == tvP.aiInputArea {
			tvP.aiInputArea.SetTextStyle(focusedInputStyle)
		} else {
			tvP.aiInputArea.SetTextStyle(defaultInputStyle)
		}
	}

	// Style Pane Content Backgrounds
	if tvP.localOutputView != nil {
		_, pageContentA := tvP.localOutputView.GetFrontPage()
		if newFocusTarget == tvP.localOutputView {
			setPrimitiveBackgroundColor(pageContentA, focusBackground)
		} else {
			setPrimitiveBackgroundColor(pageContentA, blurBackground)
		}
	}
	if tvP.aiOutputView != nil {
		_, pageContentB := tvP.aiOutputView.GetFrontPage()
		if newFocusTarget == tvP.aiOutputView {
			setPrimitiveBackgroundColor(pageContentB, focusBackground)
		} else {
			setPrimitiveBackgroundColor(pageContentB, blurBackground)
		}
	}
	tvP.LogToDebugScreen("[DFOCUS] Styling complete.")

	if newPagesView, ok := newFocusTarget.(*tview.Pages); ok {
		pageName, newPageContent := newPagesView.GetFrontPage()
		tvP.LogToDebugScreen("[DFOCUS] New focus target is Pages. Current front page name: %s, content type: %T", pageName, newPageContent)

		if newPageContent != nil {
			isLeftNew := (newPagesView == tvP.localOutputView)
			newScreener, exists := tvP.getScreenerFromPrimitive(newPageContent, isLeftNew)
			if exists {
				paneSideName := "Right"
				if isLeftNew {
					paneSideName = "Left"
				}
				tvP.LogToDebugScreen("[DFOCUS] New focus is pane %s (Page: %s, Screener: %s). IsFocusable: %v",
					paneSideName, pageName, newScreener.Name(), newScreener.IsFocusable())
				if newScreener.IsFocusable() {
					tvP.LogToDebugScreen("[DFOCUS] Attempting to call OnFocus for new screener: %s (Page Name: %s, Screener Type: %T)", newScreener.Name(), pageName, newScreener)
					newScreener.OnFocus(func(primToFocus tview.Primitive) {
						tvP.LogToDebugScreen("[DFOCUS_ONFOCUS_CALLBACK] OnFocus callback executed for %s. Primitive to focus: %T", newScreener.Name(), primToFocus)
						tvP.LogToDebugScreen("[DFOCUS] Screener %s delegated focus to %T (%p)", newScreener.Name(), primToFocus, primToFocus)
						primitiveToActuallySetFocusOnTview = primToFocus
					})
					tvP.LogToDebugScreen("[DFOCUS] OnFocus call for new screener %s completed/returned.", newScreener.Name())
				} else {
					tvP.LogToDebugScreen("[DFOCUS] Screener %s (Page: %s) is not focusable.", newScreener.Name(), pageName)
				}
			} else {
				tvP.LogToDebugScreen("[DFOCUS] No screener found for newFocusTarget's front page content (Page: %s, Content Type: %T).", pageName, newPageContent)
			}
		} else {
			tvP.LogToDebugScreen("[DFOCUS] New focus target is Pages (Page Name: %s), but its front page content is nil.", pageName)
		}
	}

	if tvP.tviewApp != nil {
		tvP.LogToDebugScreen("[DFOCUS] Attempting tvP.tviewApp.SetFocus on: %T (%p)", primitiveToActuallySetFocusOnTview, primitiveToActuallySetFocusOnTview)
		tvP.tviewApp.SetFocus(primitiveToActuallySetFocusOnTview)
		tvP.LogToDebugScreen("[DFOCUS] tvP.tviewApp.SetFocus completed.")
	}
	tvP.updateStatusText()
	tvP.LogToDebugScreen("[DFOCUS_EXIT] dFocus exiting.")
}

// --- Methods from former tview_tui.go (event/app specific) ---

func (tvP *tviewAppPointers) LogToDebugScreen(format string, args ...interface{}) {
	if tvP.app != nil && tvP.app.Log != nil {
		tvP.app.Log.Debug(format, args...)
	} else {
		log.Printf("[TUI_LOG_FALLBACK] "+format, args...)
	}
}

// onPanePageChange is called when a tview.Pages view (a pane) switches its front page.
func (tvP *tviewAppPointers) onPanePageChange(pane *tview.Pages) {
	pageName, currentPrimitive := pane.GetFrontPage()
	tvP.LogToDebugScreen("[PAGE_CHANGE_ENTRY] onPanePageChange called. Pane Addr: %p, New Page Name: '%s'", pane, pageName)

	if currentPrimitive == nil {
		tvP.LogToDebugScreen("[PAGE_CHANGE] Current primitive is nil for page '%s'.", pageName)
		tvP.LogToDebugScreen("[PAGE_CHANGE_EXIT] Current primitive nil for page '%s', onPanePageChange exiting.", pageName)
		return
	}
	isLeftPane := (pane == tvP.localOutputView)
	var screener PrimitiveScreener
	screener, _ = tvP.getScreenerFromPrimitive(currentPrimitive, isLeftPane)

	if screener != nil {
		paneType := "Right"
		if isLeftPane {
			paneType = "Left"
		}
		tvP.LogToDebugScreen("[PAGE_CHANGE] Active screener in %s pane: %s (Type: %T)", paneType, screener.Name(), screener)

		if tvP.focusablePrimitives != nil && len(tvP.focusablePrimitives) > tvP.currentFocusIndex && tvP.currentFocusIndex >= 0 {
			currentFocusedElementInCycle := tvP.focusablePrimitives[tvP.currentFocusIndex]
			paneIsTheFocusedElementInCycle := (isLeftPane && currentFocusedElementInCycle == tvP.localOutputView) ||
				(!isLeftPane && currentFocusedElementInCycle == tvP.aiOutputView)

			if paneIsTheFocusedElementInCycle {
				if screener.IsFocusable() {
					tvP.LogToDebugScreen("[PAGE_CHANGE] Pane is focused, screener %s is focusable. Attempting to call OnFocus.", screener.Name())
					screener.OnFocus(func(p tview.Primitive) {
						tvP.LogToDebugScreen("[PAGE_CHANGE_ONFOCUS_CALLBACK] OnFocus callback executed for %s. Primitive to focus: %T", screener.Name(), p)
						if tvP.tviewApp != nil {
							tvP.LogToDebugScreen("[PAGE_CHANGE] OnFocus callback: Setting tviewApp focus to primitive from %s (%T)", screener.Name(), p)
							tvP.tviewApp.SetFocus(p)
						}
					})
					tvP.LogToDebugScreen("[PAGE_CHANGE] OnFocus call for %s completed/returned.", screener.Name())
				} else {
					tvP.LogToDebugScreen("[PAGE_CHANGE] Pane is focused, but screener %s is NOT focusable. Focusing pane itself.", screener.Name())
					if tvP.tviewApp != nil {
						tvP.LogToDebugScreen("[PAGE_CHANGE] Screener %s not focusable. Attempting tvP.tviewApp.SetFocus on pane itself (%T).", screener.Name(), pane)
						tvP.tviewApp.SetFocus(pane)
						tvP.LogToDebugScreen("[PAGE_CHANGE] tvP.tviewApp.SetFocus on pane completed.")
					}
				}
			} else {
				tvP.LogToDebugScreen("[PAGE_CHANGE] Pane changed (%s), but this pane is NOT the one currently designated for focus in the dFocus cycle.", paneType)
			}
		} else {
			tvP.LogToDebugScreen("[PAGE_CHANGE] focusablePrimitives not fully initialized or currentFocusIndex out of bounds.")
		}

		if cs, ok := screener.(*ChatConversationScreen); ok {
			tvP.LogToDebugScreen("[PAGE_CHANGE] Updating ChatConversationScreen: %s", cs.Name())
			cs.UpdateConversation()
			cs.Primitive()
		} else if dos, ok := screener.(*DynamicOutputScreen); ok {
			tvP.LogToDebugScreen("[PAGE_CHANGE] DynamicOutputScreen %s became visible. Content should be current via its Write method.", dos.Name())
		}
	} else {
		tvP.LogToDebugScreen("[PAGE_CHANGE] No screener found for current primitive on page '%s' (Primitive type: %T).", pageName, currentPrimitive)
	}
	tvP.updateStatusText()
	tvP.LogToDebugScreen("[PAGE_CHANGE_EXIT] onPanePageChange exiting for page '%s'.", pageName)
}
