// NeuroScript Version: 0.4.0
// File version: 0.1.4 // Removed local switchToChatViewAndUpdate, all tvP.chatScreen refs
// Description: TUI layout, focus, and screen management logic.
// filename: pkg/neurogo/tui_layout.go
package neurogo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

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
	// Page names are simply their index in the list for now.
	// This assumes screens are not reordered or removed in a way that invalidates these.
	pageNumStr := strconv.Itoa(len(*screenList) - 1)
	pageManager.AddPage(pageNumStr, s.Primitive(), true, false)

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
	nextIndex := posmod(currentIndex+d, numScreens)
	tvP.setScreen(nextIndex, onLeft)
}

func (tvP *tviewAppPointers) setScreen(sIndex int, onLeft bool) {
	logDebug := func(msg string, keyvals ...interface{}) {}
	if tvP.app != nil && tvP.app.GetLogger() != nil {
		logDebug = tvP.app.GetLogger().Debug
	}

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
		logDebug("setScreen: targetPages is nil")
		return
	}
	if sIndex < 0 || sIndex >= len(screensList) {
		logDebug("setScreen: sIndex out of bounds", "sIndex", sIndex, "len", len(screensList))
		return
	}

	_, oldPagePrimitive := targetPages.GetFrontPage()
	if oldPagePrimitive != nil {
		setPrimitiveBackgroundColor(oldPagePrimitive, tcell.ColorBlack)
		oldScreener, oldScreenerExists := tvP.getScreenerFromPrimitive(oldPagePrimitive, onLeft)
		if oldScreenerExists {
			oldScreener.OnBlur()
		}
	}

	newPageNameStr := strconv.Itoa(sIndex)
	targetPages.SwitchToPage(newPageNameStr)
	*showingIndexToUpdate = sIndex

	newlyVisibleScreen := screensList[sIndex]
	newlyVisiblePrimitive := newlyVisibleScreen.Primitive()

	paneHasFocus := false
	if tvP.tviewApp != nil && tvP.tviewApp.GetFocus() == paneIdentifier {
		paneHasFocus = true
	} else if tvP.focusablePrimitives != nil && len(tvP.focusablePrimitives) > 0 && tvP.currentFocusIndex >= 0 && tvP.currentFocusIndex < len(tvP.focusablePrimitives) {
		paneHasFocus = (tvP.focusablePrimitives[tvP.currentFocusIndex] == paneIdentifier)
	}

	if paneHasFocus {
		setPrimitiveBackgroundColor(newlyVisiblePrimitive, tcell.ColorDarkBlue)
	} else {
		setPrimitiveBackgroundColor(newlyVisiblePrimitive, tcell.ColorBlack)
	}
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
	return nil, false
}

// switchToChatViewAndUpdate method has been REMOVED from tui_layout.go.
// Its authoritative version is in tview_tui.go.

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

func (tvP *tviewAppPointers) updateStatusText() {
	if tvP.statusBar == nil || tvP.localOutputView == nil || tvP.aiOutputView == nil {
		return
	}

	statusBarHighlightStyle := "[yellow]"
	normalTextStyle := "[-]"
	dimmedTextStyle := "[gray]%s[-]"
	separator := " [white]|[-] "

	actualLeftShowing := -1
	if tvP.localOutputView.GetPageCount() > 0 {
		leftCurrentPageName, _ := tvP.localOutputView.GetFrontPage()
		if leftCurrentPageName != "" {
			idx, err := strconv.Atoi(leftCurrentPageName)
			if err == nil {
				if idx >= 0 && idx < len(tvP.leftScreens) {
					actualLeftShowing = idx
				}
			}
		}
	}

	actualRightShowing := -1
	if tvP.aiOutputView.GetPageCount() > 0 {
		rightCurrentPageName, _ := tvP.aiOutputView.GetFrontPage()
		if rightCurrentPageName != "" {
			idx, err := strconv.Atoi(rightCurrentPageName)
			if err == nil {
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
				nameToDisplay = cs.Title()
			} else {
				nameToDisplay = EscapeTviewTags(screen.Name())
			}

			if i == actualRightShowing {
				rightScreenDisplayParts = append(rightScreenDisplayParts, fmt.Sprintf("%s%s%s", statusBarHighlightStyle, nameToDisplay, normalTextStyle))
			} else {
				rightScreenDisplayParts = append(rightScreenDisplayParts, fmt.Sprintf(dimmedTextStyle, nameToDisplay))
			}
		}
	}
	rightText := strings.Join(rightScreenDisplayParts, separator)

	var statusBarActualWidth int
	if tvP.statusBar != nil {
		_, _, rectWidth, _ := tvP.statusBar.GetInnerRect()
		if rectWidth > 0 {
			statusBarActualWidth = rectWidth
		} else {
			statusBarActualWidth = 120
		}
	} else {
		statusBarActualWidth = 120
	}

	leftTextVisibleWidth := tview.TaggedStringWidth(leftText)
	rightTextVisibleWidth := tview.TaggedStringWidth(rightText)
	paddingSize := statusBarActualWidth - leftTextVisibleWidth - rightTextVisibleWidth
	if paddingSize < 0 {
		paddingSize = 0
	}
	padding := strings.Repeat(" ", paddingSize)
	finalStatusText := leftText + padding + rightText
	if tvP.statusBar != nil {
		tvP.statusBar.SetText(finalStatusText)
	}
}
