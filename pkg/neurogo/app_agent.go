// filename: pkg/neurogo/app_agent.go
// UPDATED: Fix variable name in handleMultilineInput call.
package neurogo

import (
	"bufio" // Correctly imported
	"context"
	"errors" // For os.IsNotExist
	"fmt"    // Needed for strings.NewReader
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core"
	// Still needed for genai.Tool type below
)

// Constants (unchanged)
const (
	maxFunctionCallCycles = 5
	multiLineCommand      = "/m"
	syncCommand           = "/sync"
	defaultSyncDir        = "."
	// defaultSandboxDir defined but might be immediately overridden by startup script
)

// runAgentMode is the main entry point for interactive agent execution.
func (a *App) runAgentMode(ctx context.Context) error {
	a.InfoLog.Println("--- Starting NeuroGo in Agent Mode ---")

	// --- AgentContext and Startup Script Initialization ---
	a.DebugLog.Println("Initializing AgentContext...")
	a.agentCtx = NewAgentContext(a.GetInfoLogger()) // Create AgentContext, store on App

	agentCtxHandle, regErr := a.interpreter.RegisterHandle(a.agentCtx, HandlePrefixAgentContext)
	if regErr != nil {
		a.ErrorLog.Printf("CRITICAL: Failed to register AgentContext handle: %v", regErr)
		return fmt.Errorf("failed to register agent context handle: %w", regErr)
	}
	a.DebugLog.Printf("AgentContext registered with handle: %s", agentCtxHandle)

	if err := a.interpreter.SetVariable("AGENT_CONTEXT_HANDLE", agentCtxHandle); err != nil {
		a.ErrorLog.Printf("CRITICAL: Failed to set AGENT_CONTEXT_HANDLE variable: %v", err)
		return fmt.Errorf("failed set agent context handle variable: %w", err)
	}

	startupScriptPath := a.Config.StartupScript
	a.InfoLog.Printf("Attempting to load and execute startup script: %s", startupScriptPath)

	if _, statErr := os.Stat(startupScriptPath); errors.Is(statErr, os.ErrNotExist) {
		a.WarnLog.Printf("Startup script '%s' not found, skipping agent initialization script.", startupScriptPath)
	} else if statErr != nil {
		a.ErrorLog.Printf("Error accessing startup script '%s': %v", startupScriptPath, statErr)
		return fmt.Errorf("error accessing startup script %s: %w", startupScriptPath, statErr)
	} else {
		scriptContentBytes, readErr := os.ReadFile(startupScriptPath)
		if readErr != nil {
			a.ErrorLog.Printf("Failed to read startup script '%s': %v", startupScriptPath, readErr)
			return fmt.Errorf("failed to read startup script %s: %w", startupScriptPath, readErr)
		}
		scriptContent := string(scriptContentBytes)

		parseReader := strings.NewReader(scriptContent)
		parseOptions := core.ParseOptions{
			DebugAST: false,              // Or get from config/flag?
			Logger:   a.GetDebugLogger(), // Pass appropriate logger
		}
		parsedProcsSlice, fileVersion, parseErr := core.ParseNeuroScript(parseReader, startupScriptPath, parseOptions)
		if parseErr != nil {
			a.ErrorLog.Printf("Failed to parse startup script '%s':\n%v", startupScriptPath, parseErr)
			return fmt.Errorf("failed to parse startup script %s: %w", startupScriptPath, parseErr)
		}
		a.DebugLog.Printf("Startup script parsed successfully. File version: '%s'", fileVersion)

		mainExists := false
		for _, proc := range parsedProcsSlice {
			if addErr := a.interpreter.AddProcedure(proc); addErr != nil {
				if strings.Contains(addErr.Error(), "already defined") {
					a.WarnLog.Printf("Redefining procedure '%s' from startup script.", proc.Name)
				} else {
					a.ErrorLog.Printf("Error adding procedure '%s' from startup script: %v", proc.Name, addErr)
				}
			}
			if proc.Name == "main" {
				mainExists = true
			}
		}

		if !mainExists {
			a.WarnLog.Printf("Startup script '%s' does not contain a 'main' procedure, cannot execute.", startupScriptPath)
		} else {
			a.InfoLog.Printf("Executing 'main' procedure from startup script '%s'...", startupScriptPath)
			_, execErr := a.interpreter.RunProcedure("main")
			if execErr != nil {
				a.ErrorLog.Printf("Error executing startup script '%s': %v", startupScriptPath, execErr)
				return fmt.Errorf("error executing startup script %s: %w", startupScriptPath, execErr)
			}
			a.InfoLog.Printf("Startup script '%s' executed successfully.", startupScriptPath)
		}
	}
	// --- End AgentContext and Startup Script Initialization ---

	// --- Existing Agent Setup (Post-Startup Script) ---
	sandboxDir := a.agentCtx.GetSandboxDir()
	if sandboxDir == "" {
		a.WarnLog.Println("Startup script did not set sandbox directory, using default '.'")
		sandboxDir = "."
	}
	cleanSandboxDir := filepath.Clean(sandboxDir)
	a.DebugLog.Printf("Using effective sandbox dir (after startup script): %s", cleanSandboxDir)

	allowlistPath := a.agentCtx.GetAllowlistPath()
	allowlist, _ := loadToolListFromFile(allowlistPath)
	if allowlistPath != "" {
		a.InfoLog.Printf("Using tool allowlist: %s (%d tools allowed)", allowlistPath, len(allowlist))
	} else {
		a.InfoLog.Println("Tool allowlist disabled.")
	}
	denylistSet := make(map[string]bool)

	llmClient := a.GetLLMClient()
	if llmClient == nil || llmClient.Client() == nil {
		return fmt.Errorf("LLM Client not available for agent mode")
	}

	agentInterpreter := a.interpreter
	securityLayer := core.NewSecurityLayer(allowlist, denylistSet, cleanSandboxDir, agentInterpreter.ToolRegistry(), a.GetInfoLogger())
	toolDeclarations, genErr := securityLayer.GetToolDeclarations()
	if genErr != nil {
		a.ErrorLog.Printf("Failed to generate tool declarations: %v", genErr)
		return fmt.Errorf("failed to generate tool declarations: %w", genErr)
	}
	a.InfoLog.Printf("Initialized agent components. Sandbox: %s, Allowlist: %d, Tools: %d",
		cleanSandboxDir, len(allowlist), len(toolDeclarations))

	convoManager := core.NewConversationManager(a.GetInfoLogger())
	a.InfoLog.Println("Initial attachments should now be handled by the startup script via TOOL.AgentPinFile.")

	// --- Main Agent Loop ---
	a.InfoLog.Printf("Enter prompt (or '%s', '%s', '%s <dir> [filter]', 'quit'):", multiLineCommand, syncCommand, syncCommand)
	stdinScanner := bufio.NewScanner(os.Stdin) // Correctly initialized here
	turnCounter := 0

	for { // Main input loop
		turnCounter++
		a.InfoLog.Printf("--- Agent Conversation Turn %d ---", turnCounter)

		currentTurnFiles := a.agentCtx.GetURIsForNextContext()
		currentTurnURIs := make([]string, len(currentTurnFiles))
		for i, f := range currentTurnFiles {
			currentTurnURIs[i] = f.Name
		}
		a.DebugLog.Printf("Using %d URIs for turn %d context: %v", len(currentTurnURIs), turnCounter, currentTurnURIs)

		fmt.Printf("\nPrompt (or '%s', '%s', '%s <dir> [filter]', 'quit'): ", multiLineCommand, syncCommand, syncCommand)
		if !stdinScanner.Scan() {
			break
		}
		userInput := strings.TrimSpace(stdinScanner.Text())

		switch {
		case strings.ToLower(userInput) == "quit":
			a.InfoLog.Println("Quit command received.")
			return nil
		case userInput == multiLineCommand:
			// *** UPDATED Call: Use stdinScanner ***
			userInput = handleMultilineInput(a, stdinScanner) // Pass correct variable
			// *** END UPDATE ***
			if userInput == "" {
				continue
			}
		// --- Sync Command Handling (Simplified - Needs integration with AgentContext/Tools) ---
		case userInput == syncCommand:
			a.WarnLog.Println("Bare '/sync' command executed - state NOT YET reflected in AgentContext.")
			syncTargetDir := a.agentCtx.GetSandboxDir() // Default to sandbox
			if syncTargetDir == "" {
				syncTargetDir = defaultSyncDir
			}
			fmt.Printf("[AGENT] Starting sync for directory '%s'...\n", syncTargetDir)
			_, syncErr := core.SyncDirectoryUpHelper(ctx, syncTargetDir, "", false, llmClient.Client(), a.InfoLog, a.ErrorLog, a.DebugLog) // Assumes helper exists
			if syncErr != nil {
				fmt.Printf("[AGENT] Sync FAIL for '%s': %v\n", syncTargetDir, syncErr)
			} else {
				fmt.Printf("[AGENT] Sync OK for '%s'.\n", syncTargetDir)
				// TODO: Update AgentContext synced map
			}
			continue
		case strings.HasPrefix(userInput, syncCommand+" "):
			a.WarnLog.Println("'/sync <dir>' command executed - state NOT YET reflected in AgentContext.")
			parts := strings.Fields(userInput)
			if len(parts) < 2 {
				fmt.Println("[AGENT] Usage: /sync <directory> [optional_filter]")
				continue
			}
			syncDirArg := parts[1]
			syncFilterArg := ""
			if len(parts) > 2 {
				syncFilterArg = parts[2]
			}
			fmt.Printf("[AGENT] Starting sync for directory '%s'...\n", syncDirArg)
			_, syncErr := core.SyncDirectoryUpHelper(ctx, syncDirArg, syncFilterArg, false, llmClient.Client(), a.InfoLog, a.ErrorLog, a.DebugLog) // Assumes helper exists
			if syncErr != nil {
				fmt.Printf("[AGENT] Sync FAIL for '%s': %v\n", syncDirArg, syncErr)
			} else {
				fmt.Printf("[AGENT] Sync OK for '%s'.\n", syncDirArg)
				// TODO: Update AgentContext synced map
			}
			continue
		// --- End Sync Command Handling ---
		case userInput == "":
			continue
		default: // Process as LLM Prompt
			convoManager.AddUserMessage(userInput)
			a.DebugLog.Printf("Added user message to history: %q", userInput)

			// Use correct signature for handleAgentTurn found in handle_turn.go
			errTurn := a.handleAgentTurn(ctx, llmClient, convoManager, agentInterpreter, securityLayer, toolDeclarations, currentTurnURIs)

			if errTurn != nil {
				a.ErrorLog.Printf("Error agent turn %d: %v", turnCounter, errTurn)
				fmt.Printf("\n[AGENT] Error processing turn: %v\n", errTurn)
			}
		} // End switch
	} // End main input loop

	a.InfoLog.Println("--- Exiting Agent Mode ---")
	return nil
}

// handleMultilineInput requires scanner argument
func handleMultilineInput(a *App, scanner *bufio.Scanner) string { // Argument name matches call site
	a.InfoLog.Printf("Launching nsinput...")
	fmt.Println("Launching multi-line editor (nsinput)...")
	tempFile, err := os.CreateTemp("", "nsinput-*.txt")
	if err != nil {
		a.ErrorLog.Printf("Failed create temp file: %v", err)
		fmt.Println("[AGENT] Error creating temp file.")
		return ""
	}
	tempFilePath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempFilePath)
	cmd := exec.Command("nsinput", tempFilePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	runErr := cmd.Run()
	if runErr != nil {
		a.ErrorLog.Printf("nsinput error: %v", runErr)
		fmt.Printf("[AGENT] Multi-line input cancelled/error.\n")
		return ""
	}
	contentBytes, readErr := os.ReadFile(tempFilePath)
	if readErr != nil {
		a.ErrorLog.Printf("Failed read temp file %s: %v", tempFilePath, readErr)
		fmt.Println("[AGENT] Error reading input.")
		return ""
	}
	userInput := string(contentBytes)
	if userInput == "" {
		a.InfoLog.Println("Multi-line input empty.")
		fmt.Println("[AGENT] Multi-line input empty.")
	}
	return strings.TrimSpace(userInput)
}

// Assume loadToolListFromFile exists (likely in helpers.go)
// Assume SyncDirectoryUpHelper exists (likely in core/sync_logic.go or similar)

// loadToolListFromFile stub if needed for compilation standalone
// func loadToolListFromFile(filePath string) ([]string, error) {
// 	if filePath == "" {
// 		return nil, nil // No allowlist if path is empty
// 	}
// 	content, err := os.ReadFile(filePath)
// 	if err != nil {
// 		// Return error for now. App log should indicate failure.
// 		return nil, fmt.Errorf("failed to read allowlist %s: %w", filePath, err)
// 	}
// 	lines := strings.Split(string(content), "\n")
// 	list := make([]string, 0, len(lines))
// 	for _, line := range lines {
// 		trimmed := strings.TrimSpace(line)
// 		if trimmed != "" && !strings.HasPrefix(trimmed, "#") { // Ignore empty lines and comments
// 			list = append(list, trimmed)
// 		}
// 	}
// 	return list, nil
// }
