// NeuroScript Version: 0.4.0
// File version: 0.1.2 // Corrected LogToDebugScreen for compiler error and scrolling.
// Description: Contains methods for the tviewAppPointers struct, managing TUI logic.
// filename: pkg/neurogo/tui_methods.go
package neurogo

import (
	"fmt"
	"log" // For fallback logging if debug screen isn't ready

	"github.com/gdamore/tcell/v2" // Keep for setPrimitiveBackgroundColor, dFocus colors
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
// ADD THE fmt.Println STATEMENTS HERE AS PER THE PREVIOUS RESPONSE FOR HANG DIAGNOSIS
func (tvP *tviewAppPointers) dFocus(df int) {
	fmt.Println("[STDOUT_DFOCUS_ENTRY] dFocus called with df:", df)
	if tvP.numFocusablePrimitives == 0 {
		tvP.LogToDebugScreen("[DFOCUS] No focusable primitives.")
		fmt.Println("[STDOUT_DFOCUS_EXIT] No focusable primitives, dFocus exiting.")
		return
	}

	focusedPaneContentBackgroundColor := tcell.ColorDarkBlue
	unfocusedPaneContentBackgroundColor := tcell.ColorBlack
	defaultInputStyle := tcell.StyleDefault.Background(unfocusedPaneContentBackgroundColor).Foreground(tcell.ColorWhite)
	focusedInputStyle := tcell.StyleDefault.Background(tcell.ColorDarkSlateGray).Foreground(tcell.ColorYellow)

	if tvP.currentFocusIndex < 0 || tvP.currentFocusIndex >= len(tvP.focusablePrimitives) {
		tvP.LogToDebugScreen("[DFOCUS] currentFocusIndex %d was out of bounds (0-%d). Resetting to 0.", tvP.currentFocusIndex, tvP.numFocusablePrimitives-1)
		if tvP.numFocusablePrimitives > 0 {
			tvP.currentFocusIndex = 0
		} else {
			fmt.Println("[STDOUT_DFOCUS_EXIT] currentFocusIndex invalid and no primitives, dFocus exiting.")
			return
		}
	}
	oldFocusPrimitive := tvP.focusablePrimitives[tvP.currentFocusIndex]

	if oldPagesView, ok := oldFocusPrimitive.(*tview.Pages); ok {
		_, oldPageContent := oldPagesView.GetFrontPage()
		if oldPageContent != nil {
			isLeftOld := (oldPagesView == tvP.localOutputView)
			if oldScreener, exists := tvP.getScreenerFromPrimitive(oldPageContent, isLeftOld); exists {
				fmt.Printf("[STDOUT_DFOCUS] Attempting to call OnBlur for old screener: %s (type %T)\n", oldScreener.Name(), oldScreener)
				tvP.LogToDebugScreen("[DFOCUS] Calling OnBlur for old focused pane screener: %s", oldScreener.Name())
				oldScreener.OnBlur()
				fmt.Printf("[STDOUT_DFOCUS] OnBlur for old screener %s completed.\n", oldScreener.Name())
			}
		}
	}

	tvP.currentFocusIndex = posmod(tvP.currentFocusIndex+df, tvP.numFocusablePrimitives)
	newFocusTarget := tvP.focusablePrimitives[tvP.currentFocusIndex]
	fmt.Printf("[STDOUT_DFOCUS] New focus target index: %d, type: %T\n", tvP.currentFocusIndex, newFocusTarget)

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
			setPrimitiveBackgroundColor(pageContentA, focusedPaneContentBackgroundColor)
		} else {
			setPrimitiveBackgroundColor(pageContentA, unfocusedPaneContentBackgroundColor)
		}
	}
	if tvP.aiOutputView != nil {
		_, pageContentB := tvP.aiOutputView.GetFrontPage()
		if newFocusTarget == tvP.aiOutputView {
			setPrimitiveBackgroundColor(pageContentB, focusedPaneContentBackgroundColor)
		} else {
			setPrimitiveBackgroundColor(pageContentB, unfocusedPaneContentBackgroundColor)
		}
	}
	fmt.Println("[STDOUT_DFOCUS] Styling complete.")

	// Call OnFocus for the screener of the NEW focused primitive if it's a pane
	if newPagesView, ok := newFocusTarget.(*tview.Pages); ok {
		pageName, newPageContent := newPagesView.GetFrontPage()
		fmt.Printf("[STDOUT_DFOCUS] New focus target is Pages. Current front page name: %s, content type: %T\n", pageName, newPageContent)

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
					fmt.Printf("[STDOUT_DFOCUS] Attempting to call OnFocus for new screener: %s (Page Name: %s, Screener Type: %T)\n", newScreener.Name(), pageName, newScreener)
					newScreener.OnFocus(func(primToFocus tview.Primitive) {
						fmt.Printf("[STDOUT_DFOCUS_ONFOCUS_CALLBACK] OnFocus callback executed for %s. Primitive to focus: %T\n", newScreener.Name(), primToFocus)
						tvP.LogToDebugScreen("[DFOCUS] Screener %s delegated focus to %T (%p)", newScreener.Name(), primToFocus, primToFocus)
						primitiveToActuallySetFocusOnTview = primToFocus
					})
					fmt.Printf("[STDOUT_DFOCUS] OnFocus call for new screener %s completed/returned.\n", newScreener.Name())
				} else {
					fmt.Printf("[STDOUT_DFOCUS] Screener %s (Page: %s) is not focusable.\n", newScreener.Name(), pageName)
				}
			} else {
				fmt.Printf("[STDOUT_DFOCUS] No screener found for newFocusTarget's front page content (Page: %s, Content Type: %T).\n", pageName, newPageContent)
			}
		} else {
			fmt.Printf("[STDOUT_DFOCUS] New focus target is Pages (Page Name: %s), but its front page content is nil.\n", pageName)
		}
	}

	if tvP.tviewApp != nil {
		fmt.Printf("[STDOUT_DFOCUS] Attempting tvP.tviewApp.SetFocus on: %T\n", primitiveToActuallySetFocusOnTview)
		tvP.LogToDebugScreen("[DFOCUS] tviewApp.SetFocus on: %T (%p)", primitiveToActuallySetFocusOnTview, primitiveToActuallySetFocusOnTview)
		tvP.tviewApp.SetFocus(primitiveToActuallySetFocusOnTview)
		fmt.Printf("[STDOUT_DFOCUS] tvP.tviewApp.SetFocus completed.\n")
	}
	tvP.updateStatusText() // This method is defined in tui_layout.go on *tviewAppPointers
	fmt.Println("[STDOUT_DFOCUS_EXIT] dFocus exiting.")
}

// --- Methods from former tview_tui.go (event/app specific) ---

// LogToDebugScreen appends a message to the debug screen and scrolls to the end.
func (tvP *tviewAppPointers) LogToDebugScreen(format string, args ...interface{}) {
	if tvP.debugScreen == nil { // tvP.debugScreen is *DynamicOutputScreen (which is a PrimitiveScreener)
		log.Printf("DEBUG_SCREEN_NIL_FALLBACK: "+format, args...)
		return
	}

	message := fmt.Sprintf(format+"\n", args...)

	// Call Write directly on tvP.debugScreen (*DynamicOutputScreen)
	// Assuming *DynamicOutputScreen has a Write method.
	// If not, this will be a compiler error, and DynamicOutputScreen needs a Write method.
	_, err := tvP.debugScreen.Write([]byte(message))
	if err != nil {
		log.Printf("ERROR_WRITING_TO_DEBUG_SCREEN (via debugScreen.Write): %v | Original message: %s", err, message)
	}

	// Scroll the underlying TextView to the end
	var actualTextView *tview.TextView
	// Get the primitive from the screen wrapper
	debugPrimitive := tvP.debugScreen.Primitive()
	if debugPrimitive != nil {
		if textView, ok := debugPrimitive.(*tview.TextView); ok {
			actualTextView = textView
		} else {
			log.Printf("LogToDebugScreen: tvP.debugScreen.Primitive() is not a *tview.TextView. Type is %T. Cannot scroll.", debugPrimitive)
		}
	} else {
		log.Printf("LogToDebugScreen: tvP.debugScreen.Primitive() is nil. Cannot scroll.")
	}

	if actualTextView != nil && tvP.tviewApp != nil {
		tvP.tviewApp.QueueUpdate(func() { // Use QueueUpdate for UI state changes like scrolling
			if actualTextView != nil { // Re-check in closure as a good practice
				actualTextView.ScrollToEnd()
			}
		})
	}
}

// onPanePageChange is called when a tview.Pages view (a pane) switches its front page.
// ADD THE fmt.Println STATEMENTS HERE AS PER THE PREVIOUS RESPONSE FOR HANG DIAGNOSIS
func (tvP *tviewAppPointers) onPanePageChange(pane *tview.Pages) {
	pageName, currentPrimitive := pane.GetFrontPage()
	// Attempt to get a unique identifier for the pane if possible, otherwise use its pointer.
	// For now, let's assume pane pointers are distinct enough for logging here.
	fmt.Printf("[STDOUT_ONPANEPAGECHANGE_ENTRY] onPanePageChange called. Pane Addr: %p, New Page Name: '%s'\n", pane, pageName)

	tvP.LogToDebugScreen("[PAGE_CHANGE] Pane page changed. New page name: '%s'", pageName)
	if currentPrimitive == nil {
		tvP.LogToDebugScreen("[PAGE_CHANGE] Current primitive is nil for page '%s'.", pageName)
		fmt.Printf("[STDOUT_ONPANEPAGECHANGE_EXIT] Current primitive nil for page '%s', onPanePageChange exiting.\n", pageName)
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
		tvP.LogToDebugScreen("[PAGE_CHANGE] Active screener in %s pane: %s", paneType, screener.Name())
		fmt.Printf("[STDOUT_ONPANEPAGECHANGE] Active screener in %s pane: %s (Type: %T)\n", paneType, screener.Name(), screener)

		if tvP.focusablePrimitives != nil && len(tvP.focusablePrimitives) > tvP.currentFocusIndex && tvP.currentFocusIndex >= 0 {
			currentFocusedElementInCycle := tvP.focusablePrimitives[tvP.currentFocusIndex]
			paneIsTheFocusedElementInCycle := (isLeftPane && currentFocusedElementInCycle == tvP.localOutputView) ||
				(!isLeftPane && currentFocusedElementInCycle == tvP.aiOutputView)

			if paneIsTheFocusedElementInCycle { // If the pane itself (e.g. localOutputView) is the one that should be "active" in the focus cycle
				if screener.IsFocusable() {
					fmt.Printf("[STDOUT_ONPANEPAGECHANGE] Pane is focused, screener %s is focusable. Attempting to call OnFocus.\n", screener.Name())
					tvP.LogToDebugScreen("[PAGE_CHANGE] Pane is focused, screener %s is focusable. Calling OnFocus.", screener.Name())
					screener.OnFocus(func(p tview.Primitive) { // p is the primitive returned by screener's OnFocus's callback argument
						fmt.Printf("[STDOUT_ONPANEPAGECHANGE_ONFOCUS_CALLBACK] OnFocus callback executed for %s. Primitive to focus: %T\n", screener.Name(), p)
						if tvP.tviewApp != nil {
							tvP.LogToDebugScreen("[PAGE_CHANGE] OnFocus callback: Setting tviewApp focus to primitive from %s (%T)", screener.Name(), p)
							tvP.tviewApp.SetFocus(p)
						}
					})
					fmt.Printf("[STDOUT_ONPANEPAGECHANGE] OnFocus call for %s completed/returned.\n", screener.Name())
				} else {
					tvP.LogToDebugScreen("[PAGE_CHANGE] Pane is focused, but screener %s is NOT focusable. Focusing pane itself.", screener.Name())
					if tvP.tviewApp != nil {
						fmt.Printf("[STDOUT_ONPANEPAGECHANGE] Screener %s not focusable. Attempting tvP.tviewApp.SetFocus on pane itself (%T).\n", screener.Name(), pane)
						tvP.tviewApp.SetFocus(pane) // Focus the tview.Pages primitive
						fmt.Printf("[STDOUT_ONPANEPAGECHANGE] tvP.tviewApp.SetFocus on pane completed.\n")
					}
				}
			} else {
				fmt.Printf("[STDOUT_ONPANEPAGECHANGE] Pane changed (%s), but this pane is NOT the one currently designated for focus in the dFocus cycle.\n", paneType)
			}
		} else {
			tvP.LogToDebugScreen("[PAGE_CHANGE] focusablePrimitives not fully initialized or currentFocusIndex out of bounds.")
			fmt.Println("[STDOUT_ONPANEPAGECHANGE] focusablePrimitives issue or currentFocusIndex out of bounds.")
		}

		// Update screen content if necessary
		if cs, ok := screener.(*ChatConversationScreen); ok {
			tvP.LogToDebugScreen("[PAGE_CHANGE] Updating ChatConversationScreen: %s", cs.Name())
			fmt.Printf("[STDOUT_ONPANEPAGECHANGE] Updating ChatConversationScreen: %s\n", cs.Name())
			cs.UpdateConversation()
			cs.Primitive()
		} else if dos, ok := screener.(*DynamicOutputScreen); ok {
			tvP.LogToDebugScreen("[PAGE_CHANGE] DynamicOutputScreen %s became visible. Content should be current via its Write method.", dos.Name())
			fmt.Printf("[STDOUT_ONPANEPAGECHANGE] DynamicOutputScreen %s became visible.\n", dos.Name())
		}
	} else {
		tvP.LogToDebugScreen("[PAGE_CHANGE] No screener found for current primitive on page '%s'", pageName)
		fmt.Printf("[STDOUT_ONPANEPAGECHANGE] No screener found for current primitive on page '%s' (Primitive type: %T).\n", pageName, currentPrimitive)
	}
	tvP.updateStatusText() // This method is defined in tui_layout.go on *tviewAppPointers
	fmt.Printf("[STDOUT_ONPANEPAGECHANGE_EXIT] onPanePageChange exiting for page '%s'.\n", pageName)
}
