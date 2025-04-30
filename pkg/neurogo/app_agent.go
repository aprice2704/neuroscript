// filename: pkg/neurogo/app_agent.go
package neurogo

import (
	"bufio" // Correctly imported
	"context"
	"errors" // For os.IsNotExist
	"fmt"    // Needed for strings.NewReader which implements io.Reader
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	// Import time
	"github.com/aprice2704/neuroscript/pkg/core"
	// Ensure interfaces logger is imported if needed
	// Assume logger interface is here
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
	a.Logger.Info("--- Starting NeuroGo in Agent Mode ---")

	// --- AgentContext and Startup Script Initialization ---
	a.Logger.Debug("Initializing AgentContext...")
	// Assuming NewAgentContext is defined correctly in agent_context.go
	a.agentCtx = NewAgentContext(a.GetLogger()) // Create AgentContext, store on App

	// Check if interpreter is initialized
	if a.interpreter == nil {
		a.Logger.Error("CRITICAL: Interpreter not initialized before runAgentMode.")
		return fmt.Errorf("interpreter is not initialized")
	}

	agentCtxHandle, regErr := a.interpreter.RegisterHandle(a.agentCtx, HandlePrefixAgentContext)
	if regErr != nil {
		a.Logger.Error("CRITICAL: Failed to register AgentContext handle: %v", regErr)
		return fmt.Errorf("failed to register agent context handle: %w", regErr)
	}
	a.Logger.Debug("AgentContext registered with handle: %s", agentCtxHandle)

	if err := a.interpreter.SetVariable("AGENT_CONTEXT_HANDLE", agentCtxHandle); err != nil {
		a.Logger.Error("CRITICAL: Failed to set AGENT_CONTEXT_HANDLE variable: %v", err)
		return fmt.Errorf("failed set agent context handle variable: %w", err)
	}

	startupScriptPath := a.Config.StartupScript
	a.Logger.Info("Attempting to load and execute startup script: %s", startupScriptPath)

	if _, statErr := os.Stat(startupScriptPath); errors.Is(statErr, os.ErrNotExist) {
		a.Logger.Warn("Startup script '%s' not found, skipping agent initialization script.", startupScriptPath)
	} else if statErr != nil {
		a.Logger.Error("Error accessing startup script '%s': %v", startupScriptPath, statErr)
		return fmt.Errorf("error accessing startup script %s: %w", startupScriptPath, statErr)
	} else {
		// --- START: Updated Startup Script Parsing ---
		scriptContentBytes, readErr := os.ReadFile(startupScriptPath)
		if readErr != nil {
			a.Logger.Error("Failed to read startup script '%s': %v", startupScriptPath, readErr)
			return fmt.Errorf("failed to read startup script %s: %w", startupScriptPath, readErr)
		}
		scriptContent := string(scriptContentBytes)

		// 1. Create ParserAPI
		parserAPI := core.NewParserAPI(a.GetLogger()) // Use app's logger

		// 2. Parse the script - MODIFIED ERROR HANDLING
		antlrTree, parseErr := parserAPI.Parse(scriptContent) // Now returns single error
		if parseErr != nil {                                  // Check if error is non-nil
			// Log the single parsing error
			errMsg := fmt.Sprintf("failed to parse startup script %s: %v", startupScriptPath, parseErr)
			a.Logger.Error(errMsg)
			return fmt.Errorf("%s", errMsg) // Return the error directly
		}
		// --- END MODIFIED ERROR HANDLING ---
		if antlrTree == nil {
			// Handle case where parsing might succeed technically but yield no tree (shouldn't happen with valid grammar?)
			a.Logger.Error("Parsing startup script '%s' succeeded but resulted in a nil tree.", startupScriptPath)
			return fmt.Errorf("parsing startup script '%s' produced nil tree", startupScriptPath)
		}

		// 3. Create ASTBuilder
		astBuilder := core.NewASTBuilder(a.GetLogger()) // Use app's logger

		// 4. Build the AST (*core.Program)
		programAST, buildErr := astBuilder.Build(antlrTree)
		if buildErr != nil {
			a.Logger.Error("Failed to build AST for startup script '%s': %v", startupScriptPath, buildErr)
			return fmt.Errorf("failed to build AST for startup script '%s': %w", startupScriptPath, buildErr)
		}
		if programAST == nil {
			a.Logger.Error("AST build for startup script '%s' resulted in nil program without explicit error.", startupScriptPath)
			return fmt.Errorf("AST build for startup script '%s' resulted in nil program", startupScriptPath)
		}

		// --- MODIFIED: Access Metadata for file version ---
		fileVersion := "(not specified)" // Default value
		if version, ok := programAST.Metadata["file_version"]; ok {
			fileVersion = version
		}
		a.Logger.Debug("Startup script parsed and AST built successfully. File version: '%s'", fileVersion)
		// --- END MODIFICATION ---

		// --- Process Procedures from the built AST ---
		mainExists := false
		for _, proc := range programAST.Procedures { // Iterate over procedures in the AST
			procCopy := proc // Important: Create a copy as we pass it by value below.
			// --- FIXED: Pass procCopy by value, not pointer ---
			if addErr := a.interpreter.AddProcedure(procCopy); addErr != nil { // Pass value
				// Check if the error is due to redefinition (which might be intended)
				if strings.Contains(addErr.Error(), "already defined") {
					a.Logger.Warn("Redefining procedure '%s' from startup script.", proc.Name)
				} else {
					a.Logger.Error("Error adding procedure '%s' from startup script: %v", proc.Name, addErr)
					// Decide if this should be fatal
				}
			} else {
				a.Logger.Debug("Added procedure '%s' from startup script.", proc.Name)
			}
			if proc.Name == "main" {
				mainExists = true
			}
		}
		// --- END: Updated Startup Script Parsing ---

		if !mainExists {
			a.Logger.Warn("Startup script '%s' does not contain a 'main' procedure, cannot execute.", startupScriptPath)
		} else {
			a.Logger.Info("Executing 'main' procedure from startup script '%s'...", startupScriptPath)
			// Execute the main procedure using the interpreter instance
			_, execErr := a.interpreter.RunProcedure("main") // Assuming RunProcedure("main") is correct API
			if execErr != nil {
				a.Logger.Error("Error executing startup script '%s': %v", startupScriptPath, execErr)
				return fmt.Errorf("error executing startup script %s: %w", startupScriptPath, execErr)
			}
			a.Logger.Info("Startup script '%s' executed successfully.", startupScriptPath)
		}
	}
	// --- End AgentContext and Startup Script Initialization ---

	// --- Existing Agent Setup (Post-Startup Script) ---
	sandboxDir := a.agentCtx.GetSandboxDir()
	if sandboxDir == "" {
		a.Logger.Warn("Startup script did not set sandbox directory, using default '.'")
		sandboxDir = "."
	}
	cleanSandboxDir := filepath.Clean(sandboxDir)
	a.Logger.Debug("Using effective sandbox dir (after startup script): %s", cleanSandboxDir)

	allowlistPath := a.agentCtx.GetAllowlistPath()
	// Use helper function from helpers.go (definition removed from this file)
	allowlist, loadErr := loadToolListFromFile(allowlistPath) // Assuming this helper exists
	if loadErr != nil {
		a.Logger.Error("Failed to load tool allowlist from '%s': %v. Proceeding without allowlist.", allowlistPath, loadErr)
		allowlist = nil // Ensure allowlist is nil on error
	}

	if allowlistPath != "" && loadErr == nil {
		a.Logger.Info("Using tool allowlist: %s (%d tools allowed)", allowlistPath, len(allowlist))
	} else if allowlistPath == "" {
		a.Logger.Info("Tool allowlist path not set.")
	}
	denylistSet := make(map[string]bool) // Assuming denylist isn't loaded from file currently

	llmClient := a.GetLLMClient()
	if llmClient == nil || llmClient.Client() == nil {
		a.Logger.Error("LLM Client or its underlying client is nil, cannot proceed in agent mode.")
		return fmt.Errorf("LLM Client not available for agent mode")
	}

	agentInterpreter := a.interpreter
	if agentInterpreter == nil || agentInterpreter.ToolRegistry() == nil {
		a.Logger.Error("Interpreter or its ToolRegistry is nil, cannot initialize SecurityLayer.")
		return fmt.Errorf("interpreter tool registry is not initialized")
	}
	securityLayer := core.NewSecurityLayer(allowlist, denylistSet, cleanSandboxDir, agentInterpreter.ToolRegistry(), a.GetLogger())
	toolDeclarations, genErr := securityLayer.GetToolDeclarations()
	if genErr != nil {
		a.Logger.Error("Failed to generate tool declarations: %v", genErr)
		return fmt.Errorf("failed to generate tool declarations: %w", genErr)
	}
	a.Logger.Info("Initialized agent components. Sandbox: %s, Allowlist: %d, Tools: %d",
		cleanSandboxDir, len(allowlist), len(toolDeclarations))

	convoManager := core.NewConversationManager(a.GetLogger())
	a.Logger.Info("Initial attachments should now be handled by the startup script via TOOL.AgentPin.")

	// --- Main Agent Loop ---
	a.Logger.Info("Enter prompt (or '%s', '%s', '%s <dir> [filter]', 'quit'):", multiLineCommand, syncCommand, syncCommand)
	stdinScanner := bufio.NewScanner(os.Stdin)
	turnCounter := 0

	for { // Main input loop
		turnCounter++
		a.Logger.Info("--- Agent Conversation Turn %d ---", turnCounter)

		// Get current attachments from AgentContext (URIs)
		currentTurnFiles := a.agentCtx.GetURIsForNextContext() // Assume returns []core.FileInfo
		currentTurnURIs := make([]string, 0, len(currentTurnFiles))
		if currentTurnFiles != nil {
			for _, f := range currentTurnFiles {
				// Ensure FileInfo has a field like URI or Name containing the URI string
				// Let's assume it's URI for clarity
				if f.URI != "" {
					currentTurnURIs = append(currentTurnURIs, f.URI)
				} else {
					a.Logger.Warn("AgentContext returned FileInfo with empty URI.", "display_name", f.DisplayName)
				}
			}
		}
		a.Logger.Debug("Using %d URIs for turn %d context: %v", len(currentTurnURIs), turnCounter, currentTurnURIs)

		fmt.Printf("\nPrompt (or '%s', '%s', '%s <dir> [filter]', 'quit'): ", multiLineCommand, syncCommand, syncCommand)
		if !stdinScanner.Scan() {
			if err := stdinScanner.Err(); err != nil {
				a.Logger.Error("Error reading stdin: %v", err)
				return fmt.Errorf("error reading input: %w", err)
			}
			a.Logger.Info("EOF detected on stdin, exiting agent mode.")
			break
		}
		userInput := strings.TrimSpace(stdinScanner.Text())

		switch {
		case strings.ToLower(userInput) == "quit":
			a.Logger.Info("Quit command received.")
			return nil
		case userInput == multiLineCommand:
			userInput = handleMultilineInput(a, stdinScanner)
			if userInput == "" {
				continue // Skip if multiline input was cancelled or empty
			}
			// Fallthrough intended: process the received multiline input as a prompt
			fallthrough
		case strings.HasPrefix(userInput, syncCommand): // Combined sync handling
			a.Logger.Warn("Sync command executed - state changes are NOT YET fully reflected back into AgentContext after sync.")
			parts := strings.Fields(userInput)
			syncTargetDir := a.agentCtx.GetSandboxDir() // Default to agent sandbox
			syncFilterArg := ""
			if len(parts) > 1 {
				syncTargetDir = parts[1]
			}
			if len(parts) > 2 {
				syncFilterArg = parts[2] // Use the third part as filter
			}
			// Ensure syncTargetDir is never empty
			if syncTargetDir == "" {
				a.Logger.Warn("Sync target directory resolved to empty, using default '.'")
				syncTargetDir = defaultSyncDir
			}
			absSyncDir, pathErr := filepath.Abs(syncTargetDir)
			if pathErr != nil {
				a.Logger.Error("Invalid sync directory path '%s': %v", syncTargetDir, pathErr)
				fmt.Printf("[AGENT] Error: Invalid directory path for sync: %s\n", syncTargetDir)
				continue
			}

			fmt.Printf("[AGENT] Starting sync for directory '%s' (Filter: '%s')...\n", absSyncDir, syncFilterArg)
			// Assuming SyncDirectoryUpHelper is in core package and handles File API client
			_, syncErr := core.SyncDirectoryUpHelper(ctx, absSyncDir, syncFilterArg, false, llmClient.Client(), a.Logger)
			if syncErr != nil {
				fmt.Printf("[AGENT] Sync FAIL for '%s': %v\n", absSyncDir, syncErr)
			} else {
				fmt.Printf("[AGENT] Sync OK for '%s'.\n", absSyncDir)
				// TODO: Update AgentContext's internal map of synced files/URIs after successful sync.
			}
			continue // Sync command doesn't proceed to LLM turn
		case userInput == "":
			continue // Ignore empty input
		default: // Process as LLM Prompt
			convoManager.AddUserMessage(userInput)
			a.Logger.Debug("Added user message to history", "message", userInput)

			// Call handleAgentTurn - ensure it exists and compiles
			// It should now take currentTurnURIs as context
			errTurn := a.handleAgentTurn(ctx, llmClient, convoManager, agentInterpreter, securityLayer, toolDeclarations, currentTurnURIs)

			if errTurn != nil {
				a.Logger.Error("Error during agent turn", "turn", turnCounter, "error", errTurn)
				fmt.Printf("\n[AGENT] Error processing turn: %v\n", errTurn)
				// Decide if error should break loop or just report
			}
		} // End switch
	} // End main input loop

	a.Logger.Info("--- Exiting Agent Mode (End of Input) ---")
	return nil
}

// handleMultilineInput requires scanner argument
func handleMultilineInput(a *App, scanner *bufio.Scanner) string { // Argument name matches call site
	a.Logger.Info("Launching nsinput...")
	fmt.Println("Launching multi-line editor (nsinput)... End with empty line or Ctrl+D.") // Info for user
	tempFile, err := os.CreateTemp("", "nsinput-*.txt")
	if err != nil {
		a.Logger.Error("Failed create temp file for nsinput", "error", err)
		fmt.Println("[AGENT] Error creating temp file for multi-line input.")
		return ""
	}
	tempFilePath := tempFile.Name()
	tempFile.Close() // Close file before editor uses it
	defer func() {
		err := os.Remove(tempFilePath)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			a.Logger.Warn("Failed to remove nsinput temp file", "path", tempFilePath, "error", err)
		}
	}()

	nsinputPath, err := exec.LookPath("nsinput")
	if err != nil {
		a.Logger.Error("nsinput command not found in PATH", "error", err)
		fmt.Println("[AGENT] Error: 'nsinput' command not found. Cannot use multi-line input.")
		return ""
	}

	cmd := exec.Command(nsinputPath, tempFilePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	runErr := cmd.Run()
	if runErr != nil {
		// Log as warning, could be user cancellation
		a.Logger.Warn("nsinput command finished with error (e.g., cancelled)", "error", runErr)
		fmt.Printf("[AGENT] Multi-line input cancelled or editor exited with error.\n")
		// Return empty to indicate cancellation or error
		return ""
	}

	// Read content AFTER editor is closed
	contentBytes, readErr := os.ReadFile(tempFilePath)
	if readErr != nil {
		a.Logger.Error("Failed read temp file after nsinput", "path", tempFilePath, "error", readErr)
		fmt.Println("[AGENT] Error reading input after editor closed.")
		return ""
	}

	userInput := string(contentBytes)
	trimmedInput := strings.TrimSpace(userInput)
	if trimmedInput == "" {
		a.Logger.Info("Multi-line input resulted in empty content.")
		fmt.Println("[AGENT] Multi-line input empty.")
	} else {
		a.Logger.Debug("Multi-line input received", "bytes", len(userInput))
	}
	// Return the potentially empty (if user saved nothing) but trimmed input
	return trimmedInput
}

// loadToolListFromFile should be defined in helpers.go or similar
// func loadToolListFromFile(filePath string) ([]string, error) { ... }
