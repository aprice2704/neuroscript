// NeuroScript Version: 0.3.0
// File version: 0.0.2 // Corrected helper errors, defined cmd-returning methods on model.
// filename: pkg/neurogo/update_helpers.go
// nlines: 330 // Approximate
// risk_rating: MEDIUM
package neurogo

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// Constants for layout calculations, mirroring view.go logic if possible
const (
	// statusBarHeight is defined in model.go or view.go, assuming 1
	// defaultInputAreaHeight is also conceptual, let's use a fixed slot for calculation
	inputAreaSlotCalcHeight = 3 // A conceptual slot height for input areas in calculations
)

// --- Focus Management Helpers ---

func (m *model) cycleFocus(reverse bool) tea.Cmd {
	var cmds []tea.Cmd
	activeLeftS := m.getActiveLeftScreen()
	activeRightS := m.getActiveRightScreen()

	switch m.focusTarget {
	case FocusLeftInput:
		if activeLeftS != nil {
			cmds = append(cmds, activeLeftS.Blur(m.app))
		}
		m.leftInputArea.Blur()
	case FocusRightInput:
		if activeRightS != nil {
			cmds = append(cmds, activeRightS.Blur(m.app))
		}
		m.rightInputArea.Blur()
	case FocusLeftPane:
		if activeLeftS != nil {
			cmds = append(cmds, activeLeftS.Blur(m.app))
		}
	case FocusRightPane:
		if activeRightS != nil {
			cmds = append(cmds, activeRightS.Blur(m.app))
		}
	}

	if reverse {
		m.focusTarget = (m.focusTarget - 1 + totalFocusTargets) % totalFocusTargets
	} else {
		m.focusTarget = (m.focusTarget + 1) % totalFocusTargets
	}
	cmds = append(cmds, m.updateFocusStates())
	return tea.Batch(cmds...)
}

func (m *model) updateFocusStates() tea.Cmd {
	var cmds []tea.Cmd
	m.leftInputArea.Blur()
	m.rightInputArea.Blur()

	activeLeftScreen := m.getActiveLeftScreen()
	activeRightScreen := m.getActiveRightScreen()
	activityMsg := "Focus updated"

	switch m.focusTarget {
	case FocusLeftInput:
		if activeLeftScreen != nil {
			m.syncInputAreaWithScreen(&m.leftInputArea, activeLeftScreen)
			cmds = append(cmds, activeLeftScreen.Focus(m.app), m.leftInputArea.Focus())
			activityMsg = "Focus: Left Input (" + activeLeftScreen.Name() + ")"
		} else {
			m.syncInputAreaWithScreen(&m.leftInputArea, nil)
			cmds = append(cmds, m.leftInputArea.Focus())
			activityMsg = "Focus: Left Input"
		}
	case FocusRightInput:
		if activeRightScreen != nil {
			m.syncInputAreaWithScreen(&m.rightInputArea, activeRightScreen)
			cmds = append(cmds, activeRightScreen.Focus(m.app), m.rightInputArea.Focus())
			activityMsg = "Focus: Right Input (" + activeRightScreen.Name() + ")"
		} else {
			m.syncInputAreaWithScreen(&m.rightInputArea, nil)
			cmds = append(cmds, m.rightInputArea.Focus())
			activityMsg = "Focus: Right Input"
		}
	case FocusLeftPane:
		if activeLeftScreen != nil {
			cmds = append(cmds, activeLeftScreen.Focus(m.app))
			activityMsg = "Focus: Left Pane (" + activeLeftScreen.Name() + ")"
		} else {
			activityMsg = "Focus: Left Pane (No active screen)"
		}
	case FocusRightPane:
		if activeRightScreen != nil {
			cmds = append(cmds, activeRightScreen.Focus(m.app))
			activityMsg = "Focus: Right Pane (" + activeRightScreen.Name() + ")"
		} else {
			activityMsg = "Focus: Right Pane (No active screen)"
		}
	}
	m.currentActivity = activityMsg // Update status bar text
	return tea.Batch(cmds...)
}

func (m *model) getFocusedGlobalInputArea() *textarea.Model {
	if m.focusTarget == FocusLeftInput {
		return &m.leftInputArea
	}
	if m.focusTarget == FocusRightInput {
		return &m.rightInputArea
	}
	return nil
}

func (m *model) getScreenForFocusedInput() Screen {
	if m.focusTarget == FocusLeftInput {
		return m.getActiveLeftScreen()
	}
	if m.focusTarget == FocusRightInput {
		return m.getActiveRightScreen()
	}
	return nil
}

func (m *model) syncInputAreaWithScreen(globalInput *textarea.Model, screen Screen) {
	if screen == nil {
		globalInput.SetValue("")
		globalInput.Placeholder = "(No active screen)"
		globalInput.Prompt = "   "
		return
	}
	screenBubble := screen.GetInputBubble()
	if screenBubble != nil {
		globalInput.SetValue(screenBubble.Value())
		globalInput.Placeholder = screenBubble.Placeholder
		globalInput.Prompt = screenBubble.Prompt
		globalInput.KeyMap = screenBubble.KeyMap
	} else {
		globalInput.SetValue("")
		globalInput.Placeholder = "(No input for " + screen.Name() + ")"
		globalInput.Prompt = "   "
	}
}

// --- Command Execution Helpers ---

func (m *model) handleSystemCommand(inputValue string) tea.Cmd {
	var cmds []tea.Cmd
	m.systemMessages = append(m.systemMessages, message{"System Command", inputValue}) // Use m.addSystemMessage helper later
	parts := strings.Fields(inputValue)
	if len(parts) == 0 {
		return nil
	}
	sysCmd := parts[0]
	args := parts[1:]

	switch sysCmd {
	case "//chat":
		cmds = append(cmds, m.executeChatCommand(args))
	case "//run":
		cmds = append(cmds, m.executeRunCommand(args))
	case "//sync":
		cmds = append(cmds, m.executeSyncCommand())
	case "//q", "//quit", "//exit":
		m.quitting = true
		cmds = append(cmds, tea.Quit)
	default:
		m.systemMessages = append(m.systemMessages, message{"System", fmt.Sprintf("Unknown system command: %s", sysCmd)})
	}
	return tea.Batch(cmds...)
}

func (m *model) executeChatCommand(args []string) tea.Cmd {
	if len(args) < 1 {
		m.systemMessages = append(m.systemMessages, message{"System", "Usage: //chat <worker_base36_num>"})
		return nil
	}
	workerNumStr := args[0]
	workerIdx, err := base36ToIndex(workerNumStr)
	if err != nil {
		m.systemMessages = append(m.systemMessages, message{"System", fmt.Sprintf("Invalid worker number format: '%s'. Error: %v", workerNumStr, err)})
		return nil
	}
	if workerIdx < 0 || workerIdx >= len(m.lastDisplayedWMDefinitions) {
		maxVisibleIdx := "none"
		if len(m.lastDisplayedWMDefinitions) > 0 {
			maxVisibleIdx = indexToBase36(len(m.lastDisplayedWMDefinitions) - 1)
		}
		m.systemMessages = append(m.systemMessages, message{"System", fmt.Sprintf("Worker number '%s' (index %d) out of range. Available: 0-%s.", workerNumStr, workerIdx, maxVisibleIdx)})
		return nil
	}
	targetDef := m.lastDisplayedWMDefinitions[workerIdx]
	if targetDef == nil {
		m.systemMessages = append(m.systemMessages, message{"System", "Selected worker definition not found (nil pointer)."})
		return nil
	}
	isChatCapable := false
	for _, im := range targetDef.InteractionModels {
		if im == core.InteractionModelConversational || im == core.InteractionModelBoth {
			isChatCapable = true
			break
		}
	}
	if !isChatCapable {
		m.systemMessages = append(m.systemMessages, message{"System", fmt.Sprintf("Worker '%s' not chat capable.", targetDef.Name)})
		return nil
	}
	aiWM := m.app.GetAIWorkerManager()
	if aiWM == nil {
		m.systemMessages = append(m.systemMessages, message{"System", "AI Worker Manager not available."})
		return nil
	}
	instance, err := aiWM.SpawnWorkerInstance(targetDef.DefinitionID, nil, nil)
	if err != nil {
		m.systemMessages = append(m.systemMessages, message{"System", fmt.Sprintf("Error spawning worker '%s': %v", targetDef.Name, err)})
		return nil
	}

	// Calculate dimensions for the new chat screen
	helpHeight := 0
	if m.helpVisible {
		helpHeight = strings.Count(m.help.View(m.keyMap), "\n") + 1
	}

	// Use a fixed slot height for input area calculation, similar to view.go
	// This needs to be consistent with how view.go calculates screenHeight
	inputContainerVPadding := m.leftInputArea.BlurredStyle.Base.GetVerticalFrameSize() // Or inputPaneBlurredStyle if that's global
	screenContainerVPadding := screenPaneContainerStyle.GetVerticalFrameSize()

	screenHeight := m.height - statusBarHeight - helpHeight - (inputAreaSlotCalcHeight + inputContainerVPadding) - screenContainerVPadding
	if screenHeight < 1 {
		screenHeight = 1
	}

	rightPaneWidth := m.width - (m.width / 2)
	chatScreenWidth := rightPaneWidth - screenPaneContainerStyle.GetHorizontalFrameSize()
	if chatScreenWidth < 0 {
		chatScreenWidth = 0
	}
	// Assuming targetDef is *core.AIWorkerDefinition and instance is *core.AIWorkerInstance
	// You'll need to decide on a screenName, targetDef.Name is a good candidate.
	screenName := fmt.Sprintf("Chat: %s", targetDef.Name) // Example screen name
	chatScreen := NewChatScreen(m.app, chatScreenWidth, screenHeight, targetDef.DefinitionID, instance.InstanceID, screenName)
	//	chatScreen := NewChatScreen(m.app, targetDef, instance.InstanceID, chatScreenWidth, screenHeight)
	m.rightScreens = append(m.rightScreens, chatScreen)
	m.currentRightScreenIdx = len(m.rightScreens) - 1
	var cmds []tea.Cmd
	cmds = append(cmds, chatScreen.Init(m.app))
	m.focusTarget = FocusRightInput
	cmds = append(cmds, m.updateFocusStates())
	m.systemMessages = append(m.systemMessages, message{"System", fmt.Sprintf("Chat started with %s.", targetDef.Name)})
	return tea.Batch(cmds...)
}

func (m *model) executeRunCommand(args []string) tea.Cmd {
	if len(args) < 1 {
		m.systemMessages = append(m.systemMessages, message{"System", "Usage: //run <script_path>"})
		return nil
	}
	scriptPath := args[0]
	m.systemMessages = append(m.systemMessages, message{"System", fmt.Sprintf("Executing script: %s", scriptPath)})
	m.initialScriptRunning = true
	m.currentActivity = fmt.Sprintf("Running: %s", filepath.Base(scriptPath))
	// Call the model's method that returns tea.Cmd
	return tea.Batch(m.spinner.Tick, m.modelExecuteSpecificScriptCmd(scriptPath))
}

func (m *model) executeSyncCommand() tea.Cmd {
	if !m.isSyncing {
		m.isSyncing = true
		m.currentActivity = "Syncing files..."
		m.systemMessages = append(m.systemMessages, message{"System", m.currentActivity})
		// Call the model's method that returns tea.Cmd
		return tea.Batch(m.spinner.Tick, m.modelRunSyncCmd())
	}
	m.systemMessages = append(m.systemMessages, message{"System", "Sync already in progress."})
	return nil
}

// --- Screen Cycling Helper ---
func (m *model) cycleScreen(isLeftPane bool) tea.Cmd {
	var cmds []tea.Cmd
	var targetScreens *[]Screen
	var currentIndex *int
	var paneFocusTargetForInput FocusTarget
	// var paneFocusTargetForPane FocusTarget // Not directly used here, but for context
	var globalInputArea *textarea.Model
	paneName := "Left"

	if isLeftPane {
		targetScreens = &m.leftScreens
		currentIndex = &m.currentLeftScreenIdx
		paneFocusTargetForInput = FocusLeftInput
		// paneFocusTargetForPane = FocusLeftPane
		globalInputArea = &m.leftInputArea
	} else {
		targetScreens = &m.rightScreens
		currentIndex = &m.currentRightScreenIdx
		paneFocusTargetForInput = FocusRightInput
		// paneFocusTargetForPane = FocusRightPane
		globalInputArea = &m.rightInputArea
		paneName = "Right"
	}

	if len(*targetScreens) == 0 {
		m.systemMessages = append(m.systemMessages, message{"System", fmt.Sprintf("No screens available for %s pane.", paneName)})
		return nil
	}

	currentScreen := (*targetScreens)[*currentIndex]
	cmds = append(cmds, currentScreen.Blur(m.app))
	*currentIndex = (*currentIndex + 1) % len(*targetScreens)
	newScreen := (*targetScreens)[*currentIndex]
	cmds = append(cmds, newScreen.Init(m.app))
	m.syncInputAreaWithScreen(globalInputArea, newScreen)

	if m.focusTarget == paneFocusTargetForInput || (isLeftPane && m.focusTarget == FocusLeftPane) || (!isLeftPane && m.focusTarget == FocusRightPane) {
		cmds = append(cmds, newScreen.Focus(m.app))
		if m.focusTarget == paneFocusTargetForInput {
			cmds = append(cmds, globalInputArea.Focus())
		}
	}
	m.systemMessages = append(m.systemMessages, message{"System", fmt.Sprintf("%s Pane: %s", paneName, newScreen.Name())})
	return tea.Batch(cmds...)
}

// --- Model methods that return tea.Cmd for script/sync ---

// modelExecuteSpecificScriptCmd creates a command to run a script file.
// This is the actual logic that was in update.go's m.executeSpecificScriptCmd
func (m *model) modelExecuteSpecificScriptCmd(scriptPath string) tea.Cmd {
	return func() tea.Msg {
		if m.app == nil {
			return initialScriptDoneMsg{Path: scriptPath, Err: fmt.Errorf("application access not available to run script")}
		}
		logger := m.app.GetLogger()
		logger.Debug("Executing script via TUI command...", "path", scriptPath)
		ctxToUse := context.Background()
		if m.app.Context() != nil {
			ctxToUse = m.app.Context()
		}
		err := m.app.ExecuteScriptFile(ctxToUse, scriptPath)
		return initialScriptDoneMsg{Path: scriptPath, Err: err}
	}
}

// modelRunSyncCmd creates a command to run the sync operation.
// This is the actual logic that was in update.go's m.runSyncCmd
func (m *model) modelRunSyncCmd() tea.Cmd {
	return func() tea.Msg {
		if m.app == nil {
			return syncCompleteMsg{err: fmt.Errorf("application access not available for sync")}
		}
		logger := m.app.GetLogger()
		logger.Info("Starting sync process from TUI...")

		syncDir := m.app.Config.SyncDir // Use Config directly from app
		if syncDir == "" {
			return syncCompleteMsg{err: fmt.Errorf("sync dir not configured in app config")}
		}
		interp := m.app.GetInterpreter()
		if interp == nil {
			return syncCompleteMsg{err: fmt.Errorf("interpreter not available for sync")}
		}
		// Note: core.SyncDirectoryUpHelper was a placeholder.
		// Assuming a similar function exists or this needs to call a tool.
		// For now, simulate the call as before.
		// This part needs to be replaced with actual sync logic invocation.
		logger.Warn("Placeholder sync logic in modelRunSyncCmd. Implement actual sync call.")
		time.Sleep(1 * time.Second) // Simulate work
		// stats, syncErr := core.SyncDirectoryUpHelper(context.Background(), absSyncDir, m.app.GetSyncFilter(), m.app.GetSyncIgnoreGitignore(), interp)
		// return syncCompleteMsg{stats: stats, err: syncErr}
		return syncCompleteMsg{stats: map[string]interface{}{"info": "Simulated sync"}, err: nil}
	}
}
