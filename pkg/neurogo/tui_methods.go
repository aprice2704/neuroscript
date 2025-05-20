// NeuroScript Version: 0.4.0
// File version: 0.1.0 // Initial creation, combines tui_layout.go and tview_tui.go methods
// Description: Contains methods for the tviewAppPointers struct, managing TUI logic.
// filename: pkg/neurogo/tui_methods.go
package neurogo

import (
	"fmt"
	"log" // For fallback logging if debug screen isn't ready
	"strconv"
	"strings"

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

// addScreen adds a PrimitiveScreener to the specified pane (left or right).
func (tvP *tviewAppPointers) addScreen(s PrimitiveScreener, onLeft bool) {
	var pageManager *tview.Pages
	var screenList *[]PrimitiveScreener

	if onLeft {
		pageManager = tvP.localOutputView
		screenList = &tvP.leftScreens
	} else {
		pageManager = tvP.aiOutputView
		screenList = &tvP.rightScreens
	}

	*screenList = append(*screenList, s)
	// Page names are their index in the list. This assumes screens are not reordered
	// or removed in a way that would break this indexing for SwitchToPage.
	pageNumStr := strconv.Itoa(len(*screenList) - 1)
	pageManager.AddPage(pageNumStr, s.Primitive(), true, false)

	// Initialize background of newly added screen primitive.
	setPrimitiveBackgroundColor(s.Primitive(), tcell.ColorBlack)
	if tvP.app != nil && tvP.app.tui == tvP { // Ensure tvP is the main TUI controller for app
		tvP.LogToDebugScreen("[ADD_SCREEN] Added screen '%s' to %s pane. Page name: %s. Total screens: %d",
			s.Name(), map[bool]string{true: "Left", false: "Right"}[onLeft], pageNumStr, len(*screenList))
	}
}

// nextScreen cycles to the next/previous screen in the specified pane.
func (tvP *tviewAppPointers) nextScreen(d int, onLeft bool) {
	var screens []PrimitiveScreener
	var currentIndex int

	if onLeft {
		screens = tvP.leftScreens
		currentIndex = tvP.leftShowing
	} else {
		screens = tvP.rightScreens
		currentIndex = tvP.rightShowing
	}

	numScreens := len(screens)
	if numScreens == 0 {
		return
	}
	nextIndex := posmod(currentIndex+d, numScreens) // posmod is in tui_utils.go
	tvP.setScreen(nextIndex, onLeft)
}

// setScreen switches the visible screen in the specified pane.
func (tvP *tviewAppPointers) setScreen(sIndex int, onLeft bool) {
	paneName := "Right"
	if onLeft {
		paneName = "Left"
	}
	tvP.LogToDebugScreen("[SET_SCREEN] Attempting to set screen in %s pane to index %d.", paneName, sIndex)

	var targetPages *tview.Pages
	var screensList []PrimitiveScreener
	var showingIndexToUpdate *int
	var paneIdentifier tview.Primitive

	if onLeft {
		targetPages = tvP.localOutputView
		screensList = tvP.leftScreens
		showingIndexToUpdate = &tvP.leftShowing
		paneIdentifier = tvP.localOutputView
	} else {
		targetPages = tvP.aiOutputView
		screensList = tvP.rightScreens
		showingIndexToUpdate = &tvP.rightShowing
		paneIdentifier = tvP.aiOutputView
	}

	if targetPages == nil {
		tvP.LogToDebugScreen("[SET_SCREEN] TargetPages is nil for %s pane. Aborting.", paneName)
		return
	}
	if sIndex < 0 || sIndex >= len(screensList) {
		tvP.LogToDebugScreen("[SET_SCREEN] sIndex %d out of bounds for %s pane (len %d). Aborting.", sIndex, paneName, len(screensList))
		return
	}

	// Handle screen being switched AWAY from
	_, oldPagePrimitive := targetPages.GetFrontPage()
	if oldPagePrimitive != nil {
		setPrimitiveBackgroundColor(oldPagePrimitive, tcell.ColorBlack)
		oldScreener, oldScreenerExists := tvP.getScreenerFromPrimitive(oldPagePrimitive, onLeft)
		if oldScreenerExists {
			tvP.LogToDebugScreen("[SET_SCREEN] Calling OnBlur for old screen: %s", oldScreener.Name())
			oldScreener.OnBlur()
		}
	}

	// Switch to the new screen
	newPageNameStr := strconv.Itoa(sIndex)
	tvP.LogToDebugScreen("[SET_SCREEN] Switching %s pane to page name: %s (index %d)", paneName, newPageNameStr, sIndex)
	targetPages.SwitchToPage(newPageNameStr) // This triggers onPanePageChange via SetChangedFunc
	*showingIndexToUpdate = sIndex

	// Handle screen being switched TO
	newlyVisibleScreen := screensList[sIndex]
	newlyVisiblePrimitive := newlyVisibleScreen.Primitive()
	tvP.LogToDebugScreen("[SET_SCREEN] New visible screen in %s pane: %s", paneName, newlyVisibleScreen.Name())

	paneHasFocus := false
	if tvP.tviewApp != nil && tvP.tviewApp.GetFocus() == paneIdentifier {
		paneHasFocus = true
	} else if tvP.focusablePrimitives != nil && len(tvP.focusablePrimitives) > 0 && tvP.currentFocusIndex >= 0 && tvP.currentFocusIndex < len(tvP.focusablePrimitives) {
		paneHasFocus = (tvP.focusablePrimitives[tvP.currentFocusIndex] == paneIdentifier)
	}
	tvP.LogToDebugScreen("[SET_SCREEN] Pane %s has focus: %v", paneName, paneHasFocus)

	if paneHasFocus {
		setPrimitiveBackgroundColor(newlyVisiblePrimitive, tcell.ColorDarkBlue)
	} else {
		setPrimitiveBackgroundColor(newlyVisiblePrimitive, tcell.ColorBlack)
	}
	// onPanePageChange (called by SwitchToPage) will handle OnFocus for the new screen's content
	// and update the status text.
}

// getScreenerFromPrimitive finds the PrimitiveScreener associated with a given tview.Primitive.
func (tvP *tviewAppPointers) getScreenerFromPrimitive(p tview.Primitive, isLeftPane bool) (PrimitiveScreener, bool) {
	var screens []PrimitiveScreener
	if isLeftPane {
		screens = tvP.leftScreens
	} else {
		screens = tvP.rightScreens
	}
	for _, s := range screens {
		if s.Primitive() == p {
			return s, true
		}
	}
	return nil, false
}

// getScreenIndex finds the index of a PrimitiveScreener in its pane's list.
func (tvP *tviewAppPointers) getScreenIndex(s PrimitiveScreener, onLeft bool) int {
	var list []PrimitiveScreener
	if onLeft {
		list = tvP.leftScreens
	} else {
		list = tvP.rightScreens
	}
	for i, item := range list {
		if item == s {
			return i
		}
	}
	tvP.LogToDebugScreen("[GET_SCREEN_INDEX] Screen %s not found in %s pane list.", s.Name(), map[bool]string{true: "Left", false: "Right"}[onLeft])
	return -1
}

// dFocus handles cycling focus among major UI components and updating styles.
func (tvP *tviewAppPointers) dFocus(df int) {
	if tvP.numFocusablePrimitives == 0 {
		tvP.LogToDebugScreen("[DFOCUS] No focusable primitives.")
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
			return
		}
	}
	oldFocusPrimitive := tvP.focusablePrimitives[tvP.currentFocusIndex]
	// tvP.LogToDebugScreen("[DFOCUS] Old focus: %T (%p)", oldFocusPrimitive, oldFocusPrimitive) // Already logged in SetFocus if needed

	if oldPagesView, ok := oldFocusPrimitive.(*tview.Pages); ok {
		_, oldPageContent := oldPagesView.GetFrontPage()
		if oldPageContent != nil {
			isLeftOld := (oldPagesView == tvP.localOutputView) // Determine if it was left for logging
			if oldScreener, exists := tvP.getScreenerFromPrimitive(oldPageContent, isLeftOld); exists {
				tvP.LogToDebugScreen("[DFOCUS] Calling OnBlur for old focused pane screener: %s", oldScreener.Name())
				oldScreener.OnBlur()
			}
		}
	}

	tvP.currentFocusIndex = posmod(tvP.currentFocusIndex+df, tvP.numFocusablePrimitives)
	newFocusTarget := tvP.focusablePrimitives[tvP.currentFocusIndex]
	// tvP.LogToDebugScreen("[DFOCUS] New focus target: %T (%p), index: %d", newFocusTarget, newFocusTarget, tvP.currentFocusIndex)

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

	// Call OnFocus for the screener of the NEW focused primitive if it's a pane
	if newPagesView, ok := newFocusTarget.(*tview.Pages); ok {
		pageName, newPageContent := newPagesView.GetFrontPage()
		if newPageContent != nil {
			isLeftNew := (newPagesView == tvP.localOutputView) // Define isLeftNew here
			newScreener, exists := tvP.getScreenerFromPrimitive(newPageContent, isLeftNew)
			if exists {
				paneSideName := "Right"
				if isLeftNew { // Use isLeftNew
					paneSideName = "Left"
				}
				// Corrected the usage of isLeft, using paneSideName for clarity in the log
				tvP.LogToDebugScreen("[DFOCUS] New focus is pane %s (Page: %s, Screener: %s). IsFocusable: %v",
					paneSideName, pageName, newScreener.Name(), newScreener.IsFocusable())
				if newScreener.IsFocusable() {
					newScreener.OnFocus(func(primToFocus tview.Primitive) {
						tvP.LogToDebugScreen("[DFOCUS] Screener %s delegated focus to %T (%p)", newScreener.Name(), primToFocus, primToFocus)
						primitiveToActuallySetFocusOnTview = primToFocus
					})
				}
			} else {
				paneSideName := "Right"
				if isLeftNew { // Use isLeftNew
					paneSideName = "Left"
				}
				tvP.LogToDebugScreen("[DFOCUS] New focus is pane %s (Page: %s), but no screener found for its content.", paneSideName, pageName)
			}
		} else {
			// Determine if it was left or right for logging purposes
			paneSideName := "Unknown"
			if newPagesView == tvP.localOutputView {
				paneSideName = "Left"
			}
			if newPagesView == tvP.aiOutputView {
				paneSideName = "Right"
			}
			tvP.LogToDebugScreen("[DFOCUS] New focus is pane %s (Page: %s), but page content is nil.", paneSideName, pageName)
		}
	}

	if tvP.tviewApp != nil {
		tvP.LogToDebugScreen("[DFOCUS] tviewApp.SetFocus on: %T (%p)", primitiveToActuallySetFocusOnTview, primitiveToActuallySetFocusOnTview)
		tvP.tviewApp.SetFocus(primitiveToActuallySetFocusOnTview)
	}
	tvP.updateStatusText()
}

// updateStatusText updates the status bar content.
func (tvP *tviewAppPointers) updateStatusText() {
	if tvP.statusBar == nil || tvP.localOutputView == nil || tvP.aiOutputView == nil {
		return
	}

	statusBarHighlightStyle := "[yellow]"
	normalTextStyle := "[-]"
	dimmedTextStyle := "[gray]%s[-]"
	separator := " [white]|[-] "

	// Left Pane Screens
	actualLeftShowing := -1
	if tvP.localOutputView.GetPageCount() > 0 {
		leftCurrentPageName, _ := tvP.localOutputView.GetFrontPage() // This is the page *name*, e.g., "0", "1"
		if leftCurrentPageName != "" {
			idx, err := strconv.Atoi(leftCurrentPageName)
			if err == nil && idx >= 0 && idx < len(tvP.leftScreens) {
				actualLeftShowing = idx
			}
		}
	}
	var leftScreenDisplayParts []string
	for i, screen := range tvP.leftScreens {
		name := EscapeTviewTags(screen.Name())
		if i == actualLeftShowing {
			leftScreenDisplayParts = append(leftScreenDisplayParts, fmt.Sprintf("%s%s%s", statusBarHighlightStyle, name, normalTextStyle))
		} else {
			leftScreenDisplayParts = append(leftScreenDisplayParts, fmt.Sprintf(dimmedTextStyle, name))
		}
	}
	leftText := strings.Join(leftScreenDisplayParts, separator)

	// Right Pane Screens
	actualRightShowing := -1
	if tvP.aiOutputView.GetPageCount() > 0 {
		rightCurrentPageName, _ := tvP.aiOutputView.GetFrontPage()
		if rightCurrentPageName != "" {
			idx, err := strconv.Atoi(rightCurrentPageName)
			if err == nil && idx >= 0 && idx < len(tvP.rightScreens) {
				actualRightShowing = idx
			}
		}
	}
	var rightScreenDisplayParts []string
	for i, screen := range tvP.rightScreens {
		var nameToDisplay string
		if cs, ok := screen.(*ChatConversationScreen); ok {
			nameToDisplay = cs.Title() // Chat title is dynamic and can have colors
		} else {
			nameToDisplay = EscapeTviewTags(screen.Name())
		}
		if i == actualRightShowing {
			rightScreenDisplayParts = append(rightScreenDisplayParts, fmt.Sprintf("%s%s%s", statusBarHighlightStyle, nameToDisplay, normalTextStyle))
		} else {
			rightScreenDisplayParts = append(rightScreenDisplayParts, fmt.Sprintf(dimmedTextStyle, nameToDisplay))
		}
	}
	rightText := strings.Join(rightScreenDisplayParts, separator)

	var statusBarActualWidth int
	if tvP.statusBar != nil {
		_, _, rectWidth, _ := tvP.statusBar.GetInnerRect()
		statusBarActualWidth = rectWidth
	}
	if statusBarActualWidth <= 0 { // Fallback if GetInnerRect is not yet valid
		statusBarActualWidth = 120
	}

	leftTextVisibleWidth := tview.TaggedStringWidth(leftText)
	rightTextVisibleWidth := tview.TaggedStringWidth(rightText)
	paddingSize := statusBarActualWidth - leftTextVisibleWidth - rightTextVisibleWidth
	if paddingSize < 1 { // Ensure at least one space if there's room, or prevent negative
		paddingSize = 1
	}
	padding := strings.Repeat(" ", paddingSize)
	finalStatusText := leftText + padding + rightText

	if tvP.statusBar != nil {
		tvP.statusBar.SetText(finalStatusText)
	}
}

// --- Methods from former tview_tui.go (event/app specific) ---

// LogToDebugScreen appends a message to the debug screen.
func (tvP *tviewAppPointers) LogToDebugScreen(format string, args ...interface{}) {
	if tvP.debugScreen == nil {
		log.Printf("DEBUG_SCREEN_NIL_FALLBACK: "+format, args...) // Standard log fallback
		return
	}
	message := fmt.Sprintf(format+"\n", args...)     // Add newline for Fprintln behavior
	_, err := tvP.debugScreen.Write([]byte(message)) // DynamicOutputScreen.Write uses textView.Write
	if err != nil {
		log.Printf("ERROR_WRITING_TO_DEBUG_SCREEN: %v | Original message: %s", err, message)
	}
}

// onPanePageChange is called when a tview.Pages view (a pane) switches its front page.
func (tvP *tviewAppPointers) onPanePageChange(pane *tview.Pages) {
	pageName, currentPrimitive := pane.GetFrontPage()
	tvP.LogToDebugScreen("[PAGE_CHANGE] Pane page changed. New page name: '%s'", pageName)
	if currentPrimitive == nil {
		tvP.LogToDebugScreen("[PAGE_CHANGE] Current primitive is nil for page '%s'.", pageName)
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

		// If the Pages view itself is the one that should have focus in the main cycle...
		if tvP.focusablePrimitives != nil && len(tvP.focusablePrimitives) > tvP.currentFocusIndex && tvP.currentFocusIndex >= 0 {
			currentFocusedElementInCycle := tvP.focusablePrimitives[tvP.currentFocusIndex]
			paneIsTheFocusedElementInCycle := (isLeftPane && currentFocusedElementInCycle == tvP.localOutputView) ||
				(!isLeftPane && currentFocusedElementInCycle == tvP.aiOutputView)

			if paneIsTheFocusedElementInCycle { // If the pane itself is supposed to be focused
				if screener.IsFocusable() {
					tvP.LogToDebugScreen("[PAGE_CHANGE] Pane is focused, screener %s is focusable. Calling OnFocus.", screener.Name())
					screener.OnFocus(func(p tview.Primitive) { // Pass tviewApp's SetFocus
						if tvP.tviewApp != nil {
							tvP.tviewApp.SetFocus(p)
						}
					})
				} else { // If screen content not focusable, focus the Pages view itself
					tvP.LogToDebugScreen("[PAGE_CHANGE] Pane is focused, but screener %s is NOT focusable. Focusing pane itself.", screener.Name())
					if tvP.tviewApp != nil {
						tvP.tviewApp.SetFocus(pane)
					}
				}
			}
		} else {
			tvP.LogToDebugScreen("[PAGE_CHANGE] focusablePrimitives not fully initialized or currentFocusIndex out of bounds.")
		}

		// Update screen content if necessary
		if cs, ok := screener.(*ChatConversationScreen); ok {
			tvP.LogToDebugScreen("[PAGE_CHANGE] Updating ChatConversationScreen: %s", cs.Name())
			cs.UpdateConversation() // Fetches history for its sessionID
			cs.Primitive()          // Updates title
		} else if dos, ok := screener.(*DynamicOutputScreen); ok {
			// DynamicOutputScreen's Write method now directly updates the TextView.
			// A Flush might only be needed if content was buffered elsewhere by this screener type.
			// For now, assume Write is sufficient.
			tvP.LogToDebugScreen("[PAGE_CHANGE] DynamicOutputScreen %s became visible. Content should be current via its Write method.", dos.Name())
			// dos.FlushBufferToTextView() // May not be needed if Write is used consistently
		}
	} else {
		tvP.LogToDebugScreen("[PAGE_CHANGE] No screener found for current primitive on page '%s'", pageName)
	}
	tvP.updateStatusText() // Update status bar after page change
}
