// NeuroScript Version: 0.3.0
// File version: 0.0.11 // Adjusted output content height for explicit title string.
// filename: pkg/neurogo/update_helpers.go
// nlines: 205 // Approximate, please update after changes
// risk_rating: MEDIUM
package neurogo

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *model) addMessage(sender, text string) {
	if sender == "emit" {
		m.emittedLines = append(m.emittedLines, text)
		if len(m.emittedLines) > maxEmitBufferLines {
			m.emittedLines = m.emittedLines[len(m.emittedLines)-maxEmitBufferLines:]
		}
		if m.ready {
			m.localOutput.SetContent(strings.Join(m.emittedLines, "\n"))
			m.localOutput.GotoBottom()
		}
		return
	}
	if len(m.messages) > 1000 {
		m.messages = m.messages[len(m.messages)-500:]
	}
	m.messages = append(m.messages, message{sender: sender, text: text})
	if m.ready {
		m.aiOutput.SetContent(m.renderMessages())
		m.aiOutput.GotoBottom()
	}
}

func (m *model) renderMessages() string {
	var content strings.Builder
	maxMessages := 200
	start := 0
	if len(m.messages) > maxMessages {
		start = len(m.messages) - maxMessages
	}
	visibleMessages := m.messages[start:]
	for _, msg := range visibleMessages {
		var style lipgloss.Style
		switch msg.sender {
		case "You":
			style = userStyle
		case "AI":
			style = aiStyle
		case "System":
			style = systemStyle
		case "Debug":
			style = systemStyle.Copy().Italic(true).Foreground(lipgloss.Color("13"))
		case "Command":
			style = systemStyle.Copy().Foreground(lipgloss.Color("14"))
		default:
			style = systemStyle
		}
		processedText := strings.ReplaceAll(msg.text, "\r\n", "\n")
		content.WriteString(style.Render(fmt.Sprintf("%s: %s\n", msg.sender, processedText)))
	}
	finalContent := content.String()
	if strings.TrimSpace(finalContent) == "" {
		return " "
	}
	return finalContent
}

func (m *model) renderStatusBar(width int) string {
	if !m.ready {
		return ""
	}
	modelInfo := fmt.Sprintf("AI: %s", m.aiModelName)
	left := modelInfo
	activity := ""
	spinnerView := ""
	if m.isWaitingForAI || m.isSyncing || m.patchStatus != "" || (m.initialScriptRunning && m.currentActivity != "") {
		spinnerView = m.spinner.View() + " "
	}
	if m.isWaitingForAI {
		activity = fmt.Sprintf("%sWaiting for AI...", spinnerView)
	} else if m.isSyncing {
		activity = fmt.Sprintf("%s%s", spinnerView, m.currentActivity)
	} else if m.patchStatus != "" {
		activity = fmt.Sprintf("%s%s", spinnerView, m.patchStatus)
	} else if m.currentActivity != "" {
		activity = fmt.Sprintf("%s%s", spinnerView, m.currentActivity)
	}
	errorMsg := ""
	if m.lastError != nil {
		errorMsg = errorStyle.Render(fmt.Sprintf("Error: %v", m.lastError))
	}
	right := ""
	if errorMsg != "" {
		right = errorMsg
	} else {
		right = activity
	}
	hPadding := statusBarSyle.GetHorizontalPadding()
	availableWidthForText := width - hPadding*2
	maxLeftWidth := availableWidthForText / 3
	if lipgloss.Width(left) > maxLeftWidth {
		if maxLeftWidth > 1 {
			left = left[:maxLeftWidth-1] + "…"
		} else if maxLeftWidth > 0 {
			left = left[:maxLeftWidth]
		} else {
			left = ""
		}
	}
	separatorWidth := availableWidthForText - lipgloss.Width(left) - lipgloss.Width(right)
	if separatorWidth < 0 {
		maxRightWidth := availableWidthForText - lipgloss.Width(left)
		if lipgloss.Width(right) > maxRightWidth {
			if maxRightWidth > 1 {
				right = right[:maxRightWidth-1] + "…"
			} else if maxRightWidth > 0 {
				right = right[:maxRightWidth]
			} else {
				right = ""
			}
		}
		separatorWidth = max(0, availableWidthForText-lipgloss.Width(left)-lipgloss.Width(right))
	}
	separator := strings.Repeat(" ", separatorWidth)
	finalStatusText := lipgloss.JoinHorizontal(lipgloss.Top, left, separator, right)
	return statusBarSyle.Copy().Width(width).Render(finalStatusText)
}

func (m *model) runSyncCmd() tea.Cmd {
	return func() tea.Msg {
		logger := m.app.GetLogger()
		if logger == nil {
			return errMsg{fmt.Errorf("logger not available")}
		}
		if m.app == nil {
			return errMsg{fmt.Errorf("app not available")}
		}
		syncDir := m.app.GetSyncDir()
		if syncDir == "" {
			return errMsg{fmt.Errorf("sync dir not configured")}
		}
		interp := m.app.GetInterpreter()
		if interp == nil {
			return errMsg{fmt.Errorf("interpreter not available")}
		}
		fileAPI := interp.FileAPI()
		if fileAPI == nil {
			return errMsg{fmt.Errorf("FileAPI not available")}
		}
		absSyncDir, secErr := fileAPI.ResolvePath(syncDir)
		if secErr != nil {
			return errMsg{fmt.Errorf("invalid sync dir '%s': %w", syncDir, secErr)}
		}
		dirInfo, statErr := os.Stat(absSyncDir)
		if statErr != nil {
			if os.IsNotExist(statErr) {
				return errMsg{fmt.Errorf("sync dir does not exist: %s", absSyncDir)}
			}
			return errMsg{fmt.Errorf("cannot access sync dir %s: %w", absSyncDir, statErr)}
		}
		if !dirInfo.IsDir() {
			return errMsg{fmt.Errorf("sync path is not a directory: %s", absSyncDir)}
		}
		stats, syncErr := core.SyncDirectoryUpHelper(context.Background(), absSyncDir, m.app.GetSyncFilter(), m.app.GetSyncIgnoreGitignore(), interp)
		return syncCompleteMsg{stats: stats, err: syncErr}
	}
}

func (m *model) setSizes(width, height int) {
	m.help.Width = width

	const inputTextAreaHeight = 5
	inputTotalHeightWithBorders := inputTextAreaHeight + 2

	statusBarHeight := lipgloss.Height(m.renderStatusBar(width))
	helpHeight := 0
	if m.helpVisible {
		helpHeight = lipgloss.Height(m.help.View(m.keyMap))
	}

	mainApplicationAreaHeight := height - statusBarHeight - helpHeight
	if mainApplicationAreaHeight < 0 {
		mainApplicationAreaHeight = 0
	}

	const titleStringHeight = 1 // Height of the explicit title string for output panes
	const outputPaneBordersHeight = 2
	const minOutputContentHeight = 1
	minOutputTotalHeightWithBordersAndTitle := minOutputContentHeight + outputPaneBordersHeight + titleStringHeight

	outputPaneContainerHeight := mainApplicationAreaHeight - inputTotalHeightWithBorders
	if outputPaneContainerHeight < minOutputTotalHeightWithBordersAndTitle {
		outputPaneContainerHeight = minOutputTotalHeightWithBordersAndTitle
	}

	// Content height for viewports is container height minus borders minus explicit title string height
	outputContentHeight := outputPaneContainerHeight - outputPaneBordersHeight - titleStringHeight
	if outputContentHeight < minOutputContentHeight {
		outputContentHeight = minOutputContentHeight
	}

	m.localInput.SetHeight(inputTextAreaHeight)
	m.aiInput.SetHeight(inputTextAreaHeight)
	m.localOutput.Height = outputContentHeight
	m.aiOutput.Height = outputContentHeight

	columnWidth := width / 2
	rightColumnWidth := width - columnWidth

	// Width for input textareas (inside their borders)
	// The SetWidth on textarea is for its content area.
	// The style (e.g., localInputFocusedStyle) defines the border around it.
	// So, the textarea's view will be columnWidth wide in total.
	m.localInput.SetWidth(columnWidth - localInputFocusedStyle.GetHorizontalFrameSize())
	m.aiInput.SetWidth(rightColumnWidth - aiInputFocusedStyle.GetHorizontalFrameSize())

	// Width for output viewports (inside their container's borders)
	// The localOutputContainerStyle defines the border around the title + viewport content.
	// The viewport itself needs to be narrower to fit inside these borders.
	m.localOutput.Width = columnWidth - localOutputFocusedStyle.GetHorizontalFrameSize() // Assuming localOutputFocusedStyle is the container
	m.aiOutput.Width = rightColumnWidth - aiOutputFocusedStyle.GetHorizontalFrameSize()  // Assuming aiOutputFocusedStyle is the container

	minComponentWidth := 1
	if m.localInput.Width() < minComponentWidth {
		m.localInput.SetWidth(minComponentWidth)
	}
	if m.aiInput.Width() < minComponentWidth {
		m.aiInput.SetWidth(minComponentWidth)
	}
	if m.localOutput.Width < minComponentWidth {
		m.localOutput.Width = minComponentWidth
	}
	if m.aiOutput.Width < minComponentWidth {
		m.aiOutput.Width = minComponentWidth
	}

	if m.ready {
		m.aiOutput.SetContent(m.renderMessages())
		m.aiOutput.GotoBottom()
		m.localOutput.SetContent(strings.Join(m.emittedLines, "\n"))
		m.localOutput.GotoBottom()
	}
}
