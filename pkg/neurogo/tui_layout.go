// NeuroScript Version: 0.4.0
// File version: 0.1.2
// Description: TUI layout, focus, and screen management logic.
//
//	Pane content backgrounds update on focus (dark blue for focused pane, black otherwise).
//	Screen becoming non-visible has its background set to black.
//	Status bar highlights visible screen names with yellow text.
//
// filename: pkg/neurogo/tui_layout.go
package neurogo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Helper to set background color of a primitive if it's a *tview.Box or embeds it.
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

func (tvP *tviewAppPointers) addScreen(s PrimitiveScreener, onLeft bool) {
	var pageManager *tview.Pages
	var screenList *[]PrimitiveScreener

	if onLeft {
		pageManager = tvP.localOutputView // Pane A
		screenList = &tvP.leftScreens
	} else {
		pageManager = tvP.aiOutputView // Pane B
		screenList = &tvP.rightScreens
	}

	*screenList = append(*screenList, s)
	pageNumStr := strconv.Itoa(len(*screenList) - 1)
	pageManager.AddPage(pageNumStr, s.Primitive(), true, false)
	// Initialize background of newly added screen primitive to black.
	// It will be updated if/when it becomes visible and its pane is focused.
	setPrimitiveBackgroundColor(s.Primitive(), tcell.ColorBlack)
}

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

// setScreen handles switching screens within a pane (A or B)
func (tvP *tviewAppPointers) setScreen(sIndex int, onLeft bool) {
	logDebug := func(msg string, keyvals ...interface{}) {}
	if tvP.app != nil && tvP.app.GetLogger() != nil {
		logDebug = tvP.app.GetLogger().Debug
	}

	var targetPages *tview.Pages        // The Pages view (Pane A or B)
	var screensList []PrimitiveScreener // The list of all screens for this pane
	var showingIndexToUpdate *int       // Pointer to tvP.leftShowing or tvP.rightShowing
	var paneIdentifier tview.Primitive  // The tview.Pages primitive itself (tvP.localOutputView or tvP.aiOutputView)

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

	if targetPages == nil { // Should not happen with current setup
		logDebug("setScreen: targetPages is nil")
		return
	}
	if sIndex < 0 || sIndex >= len(screensList) {
		logDebug("setScreen: sIndex out of bounds", "sIndex", sIndex, "len", len(screensList))
		return
	}

	// 1. Handle the screen being switched AWAY FROM
	_, oldPagePrimitive := targetPages.GetFrontPage()
	if oldPagePrimitive != nil {
		// Set its background to black as it's becoming non-visible.
		setPrimitiveBackgroundColor(oldPagePrimitive, tcell.ColorBlack)
		// Call OnBlur for its screener
		// We need to get the screener based on the primitive, not just name, as names can be "0", "1"...
		// getScreenerFromPrimitive is better here if oldPagePrimitive is reliable.
		oldScreener, oldScreenerExists := tvP.getScreenerFromPrimitive(oldPagePrimitive, onLeft)
		if oldScreenerExists {
			oldScreener.OnBlur()
		}
	}

	// 2. Switch to the new screen in the Pages view
	newPageNameStr := strconv.Itoa(sIndex)
	targetPages.SwitchToPage(newPageNameStr) // This triggers onPanePageChange
	*showingIndexToUpdate = sIndex           // Update our tracking index for which screen is "showing"

	// 3. Handle the screen being switched TO (the new visible screen)
	newlyVisibleScreen := screensList[sIndex]               // This is the PrimitiveScreener
	newlyVisiblePrimitive := newlyVisibleScreen.Primitive() // This is the tview.Primitive

	// Determine if the parent pane (targetPages, e.g., tvP.localOutputView) currently has the TUI focus
	paneHasFocus := false
	if tvP.tviewApp != nil && tvP.tviewApp.GetFocus() == paneIdentifier { // Check if the tview.Pages object itself has focus
		paneHasFocus = true
	} else if len(tvP.focusablePrimitives) > tvP.currentFocusIndex && tvP.currentFocusIndex >= 0 { // Fallback check on our list
		paneHasFocus = (tvP.focusablePrimitives[tvP.currentFocusIndex] == paneIdentifier)
	}

	if paneHasFocus {
		setPrimitiveBackgroundColor(newlyVisiblePrimitive, tcell.ColorDarkBlue)
	} else {
		setPrimitiveBackgroundColor(newlyVisiblePrimitive, tcell.ColorBlack)
	}

	// onPanePageChange (triggered by SwitchToPage) will call updateStatusText.
	// It also handles calling OnFocus for the newlyVisibleScreen if the pane itself is focused
	// and the screen is focusable.
	logDebug("setScreen done", "newScreen", newlyVisibleScreen.Name(), "paneHasFocus", paneHasFocus)
}

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
	// Fallback for chat screen if it wasn't found by iterating rightScreens (shouldn't happen if setup correctly)
	if !isLeftPane && tvP.chatScreen != nil && p == tvP.chatScreen.Primitive() {
		return tvP.chatScreen, true
	}
	return nil, false
}

func (tvP *tviewAppPointers) switchToChatViewAndUpdate() {
	if tvP.chatScreen == nil {
		return
	}
	chatScreenIndex := tvP.getScreenIndex(tvP.chatScreen, false)

	if chatScreenIndex != -1 {
		currentPageName, _ := tvP.aiOutputView.GetFrontPage()
		currentIdx, err := strconv.Atoi(currentPageName)
		// If not on chat screen, switch to it.
		if err != nil || currentIdx != chatScreenIndex {
			tvP.setScreen(chatScreenIndex, false)
		} else { // Already on chat screen, just ensure updates
			tvP.tviewApp.QueueUpdateDraw(func() {
				if tvP.chatScreen != nil {
					tvP.chatScreen.Primitive() // Updates title based on AI status
				}
				tvP.updateStatusText() // Refresh status bar
			})
		}
	} else {
		if tvP.app != nil && tvP.app.Log != nil {
			tvP.app.Log.Warn("switchToChatViewAndUpdate: ChatScreen not found in rightScreens list.")
		}
	}
}

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
	return -1
}

// dFocus handles cycling focus AND updating pane/input area backgrounds.
func (tvP *tviewAppPointers) dFocus(df int) {
	if tvP.numFocusablePrimitives == 0 {
		return
	}

	focusedPaneContentBackgroundColor := tcell.ColorDarkBlue
	unfocusedPaneContentBackgroundColor := tcell.ColorBlack
	defaultInputStyle := tcell.StyleDefault.Background(unfocusedPaneContentBackgroundColor).Foreground(tcell.ColorWhite)
	focusedInputStyle := tcell.StyleDefault.Background(tcell.ColorDarkSlateGray).Foreground(tcell.ColorYellow) // Distinct input focus

	// --- Get old focused primitive and call OnBlur if it was a pane's content ---
	oldFocusPrimitive := tvP.focusablePrimitives[tvP.currentFocusIndex]
	if oldPagesView, ok := oldFocusPrimitive.(*tview.Pages); ok {
		_, oldPageContent := oldPagesView.GetFrontPage()
		if oldPageContent != nil {
			isLeft := (oldPagesView == tvP.localOutputView)
			if oldScreener, exists := tvP.getScreenerFromPrimitive(oldPageContent, isLeft); exists {
				oldScreener.OnBlur()
			}
		}
	}

	// --- Cycle focus index ---
	tvP.currentFocusIndex = posmod(tvP.currentFocusIndex+df, tvP.numFocusablePrimitives)
	newFocusTarget := tvP.focusablePrimitives[tvP.currentFocusIndex] // This is one of localInputArea, aiInputArea, localOutputView, aiOutputView

	primitiveToActuallySetFocusOnTview := newFocusTarget

	// --- Style Input Areas based on whether they are the newFocusTarget ---
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

	// --- Style Pane Content Backgrounds ---
	// Pane A (LocalOutputView)
	if tvP.localOutputView != nil {
		_, pageContentA := tvP.localOutputView.GetFrontPage()
		if newFocusTarget == tvP.localOutputView {
			setPrimitiveBackgroundColor(pageContentA, focusedPaneContentBackgroundColor)
		} else {
			setPrimitiveBackgroundColor(pageContentA, unfocusedPaneContentBackgroundColor)
		}
	}
	// Pane B (AIOutputView)
	if tvP.aiOutputView != nil {
		_, pageContentB := tvP.aiOutputView.GetFrontPage()
		if newFocusTarget == tvP.aiOutputView {
			setPrimitiveBackgroundColor(pageContentB, focusedPaneContentBackgroundColor)
		} else {
			setPrimitiveBackgroundColor(pageContentB, unfocusedPaneContentBackgroundColor)
		}
	}

	// --- Call OnFocus for the screener of the NEW focused primitive if it's a pane ---
	// And allow it to delegate the tview.SetFocus call
	if newPagesView, ok := newFocusTarget.(*tview.Pages); ok {
		_, newPageContent := newPagesView.GetFrontPage()
		if newPageContent != nil {
			isLeft := (newPagesView == tvP.localOutputView)
			newScreener, exists := tvP.getScreenerFromPrimitive(newPageContent, isLeft)
			if exists && newScreener.IsFocusable() {
				newScreener.OnFocus(func(primToFocus tview.Primitive) { // This callback receives the primitive the screener wants focused
					primitiveToActuallySetFocusOnTview = primToFocus
				})
			}
			// If screener is not focusable or doesn't delegate, primitiveToActuallySetFocusOnTview remains newPagesView
		}
	}

	if tvP.tviewApp != nil { // Ensure app is not nil before calling SetFocus
		tvP.tviewApp.SetFocus(primitiveToActuallySetFocusOnTview)
	}
	tvP.updateStatusText()
}

// updateStatusText highlights visible screen names with yellow text.
func (tvP *tviewAppPointers) updateStatusText() {
	if tvP.statusBar == nil || tvP.localOutputView == nil || tvP.aiOutputView == nil {
		return
	}

	statusBarHighlightStyle := "[yellow]"
	normalTextStyle := "[-]"
	dimmedTextStyle := "[gray]%s[-]" // Using Sprintf format for consistency
	separator := " [white]|[-] "

	// Get current VISIBLE screen index for left pane
	actualLeftShowing := -1 // Default: no screen highlighted if conditions not met
	if tvP.localOutputView.GetPageCount() > 0 {
		leftCurrentPageName, _ := tvP.localOutputView.GetFrontPage()
		if leftCurrentPageName != "" {
			idx, err := strconv.Atoi(leftCurrentPageName)
			if err == nil {
				// Crucial check: is the parsed index valid for the current tvP.leftScreens slice?
				if idx >= 0 && idx < len(tvP.leftScreens) {
					actualLeftShowing = idx
				}
			}
		}
	}

	// Get current VISIBLE screen index for right pane
	actualRightShowing := -1
	if tvP.aiOutputView.GetPageCount() > 0 {
		rightCurrentPageName, _ := tvP.aiOutputView.GetFrontPage()
		if rightCurrentPageName != "" {
			idx, err := strconv.Atoi(rightCurrentPageName)
			if err == nil {
				// Crucial check: is the parsed index valid for the current tvP.rightScreens slice?
				if idx >= 0 && idx < len(tvP.rightScreens) {
					actualRightShowing = idx
				}
			}
		}
	}

	var leftScreenDisplayParts []string
	if len(tvP.leftScreens) > 0 {
		for i, screen := range tvP.leftScreens {
			name := EscapeTviewTags(screen.Name())
			if i == actualLeftShowing {
				leftScreenDisplayParts = append(leftScreenDisplayParts, fmt.Sprintf("%s%s%s", statusBarHighlightStyle, name, normalTextStyle))
			} else {
				leftScreenDisplayParts = append(leftScreenDisplayParts, fmt.Sprintf(dimmedTextStyle, name))
			}
		}
	}
	leftText := strings.Join(leftScreenDisplayParts, separator)

	var rightScreenDisplayParts []string
	if len(tvP.rightScreens) > 0 {
		for i, screen := range tvP.rightScreens {
			var nameToDisplay string
			if cs, ok := screen.(*ChatConversationScreen); ok {
				nameToDisplay = cs.Title() // Chat title can have its own colors
			} else {
				nameToDisplay = EscapeTviewTags(screen.Name())
			}

			if i == actualRightShowing {
				// Apply yellow text to the whole name/title. If title has internal colors,
				// tview's tag nesting rules will apply (often the innermost wins for foreground).
				rightScreenDisplayParts = append(rightScreenDisplayParts, fmt.Sprintf("%s%s%s", statusBarHighlightStyle, nameToDisplay, normalTextStyle))
			} else {
				rightScreenDisplayParts = append(rightScreenDisplayParts, fmt.Sprintf(dimmedTextStyle, nameToDisplay))
			}
		}
	}
	rightText := strings.Join(rightScreenDisplayParts, separator)

	// --- Combine and Pad for Layout ---
	var statusBarActualWidth int
	_, _, rectWidth, _ := tvP.statusBar.GetInnerRect()
	if rectWidth > 0 {
		statusBarActualWidth = rectWidth
	} else {
		statusBarActualWidth = 120 // Fallback width if GetInnerRect returns 0 (e.g., before first draw)
	}

	leftTextVisibleWidth := tview.TaggedStringWidth(leftText)
	rightTextVisibleWidth := tview.TaggedStringWidth(rightText)
	paddingSize := statusBarActualWidth - leftTextVisibleWidth - rightTextVisibleWidth
	if paddingSize < 0 {
		paddingSize = 0 // No space for padding if text is wider than the bar.
	}
	padding := strings.Repeat(" ", paddingSize)
	finalStatusText := leftText + padding + rightText
	tvP.statusBar.SetText(finalStatusText)
}
