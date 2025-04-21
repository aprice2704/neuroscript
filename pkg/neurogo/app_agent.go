// filename: pkg/neurogo/app_agent.go
package neurogo

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core"
)

// Constants, runAgentMode setup... (unchanged)
const (
	maxFunctionCallCycles = 5
	multiLineCommand      = "/m"
	syncCommand           = "/sync"
	defaultSyncDir        = "."
	defaultSandboxDir     = "."
)

func (a *App) runAgentMode(ctx context.Context) error {
	// ... Initialization ... (unchanged)
	a.InfoLog.Println("--- Starting NeuroGo in Agent Mode ---")
	allowlist, _ := loadToolListFromFile(a.Config.AllowlistFile)
	denylistSet := make(map[string]bool)
	cleanSandboxDir := filepath.Clean(a.Config.SandboxDir)
	a.DebugLog.Printf("Using cleaned sandbox dir: %s", cleanSandboxDir)
	llmClient := a.llmClient
	if llmClient == nil || llmClient.Client() == nil {
		return fmt.Errorf("LLM Client not initialized")
	}
	convoManager := core.NewConversationManager(a.InfoLog)
	agentInterpreter := core.NewInterpreter(a.DebugLog, llmClient)
	coreRegistry := agentInterpreter.ToolRegistry()
	core.RegisterCoreTools(coreRegistry)
	securityLayer := core.NewSecurityLayer(allowlist, denylistSet, cleanSandboxDir, coreRegistry, a.InfoLog)
	toolDeclarations, genErr := securityLayer.GetToolDeclarations()
	if genErr != nil {
		a.ErrorLog.Printf("Failed tool declarations: %v", genErr)
		return fmt.Errorf("failed gen tools: %w", genErr)
	}
	a.InfoLog.Printf("Initialized agent components. Sandbox: %s, Allowlist: %d, Tools: %d", cleanSandboxDir, len(allowlist), len(toolDeclarations))

	// ... Initial Attachments ... (unchanged, includes fmt.Printf feedback)
	initialAttachmentURIs := []string{}
	if len(a.Config.InitialAttachments) > 0 {
		a.InfoLog.Printf("Processing %d initial attachments...", len(a.Config.InitialAttachments))
		fmt.Printf("[AGENT] Processing %d initial attachments...\n", len(a.Config.InitialAttachments))
		if llmClient.Client() != nil {
			for _, attachPath := range a.Config.InitialAttachments {
				absAttachPath, secErr := core.ResolveAndSecurePath(attachPath, cleanSandboxDir)
				if secErr != nil {
					a.ErrorLog.Printf("Skip attach %q sec err: %v", attachPath, secErr)
					fmt.Printf("[AGENT] Skipping attachment %q: %v\n", attachPath, secErr)
					continue
				}
				displayPath := attachPath
				relPath, relErr := filepath.Rel(cleanSandboxDir, absAttachPath)
				if relErr == nil {
					displayPath = filepath.ToSlash(relPath)
				}
				a.InfoLog.Printf("Uploading initial attachment: %s (Abs: %s, DisplayName: %s)", attachPath, absAttachPath, displayPath)
				fmt.Printf("[AGENT] Attaching file: %s ... ", attachPath)
				uploadCtx := context.Background()
				apiFile, uploadErr := core.HelperUploadAndPollFile(uploadCtx, absAttachPath, displayPath, llmClient.Client(), a.LLMLog)
				if uploadErr != nil {
					fmt.Printf("FAILED (%v)\n", uploadErr)
					a.ErrorLog.Printf("Failed attach %s: %v", attachPath, uploadErr)
				} else if apiFile != nil {
					fmt.Printf("OK (URI: %s)\n", apiFile.URI)
					a.InfoLog.Printf("Initial attach OK: %s -> URI: %s", attachPath, apiFile.URI)
					initialAttachmentURIs = append(initialAttachmentURIs, apiFile.URI)
				} else {
					fmt.Printf("FAILED (Unknown error)\n")
					a.ErrorLog.Printf("Unknown error attaching file %s", attachPath)
				}
			}
			a.InfoLog.Printf("Finished initial attachments. URIs: %d: %v", len(initialAttachmentURIs), initialAttachmentURIs)
		} else {
			a.ErrorLog.Println("Cannot process initial attachments: LLM Client is nil.")
		}
	}

	accumulatedContextURIs := []string{}
	accumulatedContextURIs = append(accumulatedContextURIs, initialAttachmentURIs...)
	a.InfoLog.Printf("Enter prompt (or '%s', '%s', '%s <dir> [filter]', 'quit'):", multiLineCommand, syncCommand, syncCommand)
	stdinScanner := bufio.NewScanner(os.Stdin)
	turnCounter := 0

	for { // Main input loop
		turnCounter++
		a.InfoLog.Printf("--- Agent Conversation Turn %d ---", turnCounter)
		currentTurnURIs := make([]string, len(accumulatedContextURIs))
		copy(currentTurnURIs, accumulatedContextURIs)
		a.DebugLog.Printf("Using %d URIs for turn %d context: %v", len(currentTurnURIs), turnCounter, currentTurnURIs)
		fmt.Printf("\nPrompt (or '%s', '%s', '%s <dir> [filter]', 'quit'): ", multiLineCommand, syncCommand, syncCommand)
		if !stdinScanner.Scan() {
			break
		} // Simplified EOF/error handling
		userInput := strings.TrimSpace(stdinScanner.Text())

		switch {
		case strings.ToLower(userInput) == "quit":
			a.InfoLog.Println("Quit command received.")
			return nil
		case userInput == multiLineCommand:
			userInput = handleMultilineInput(a, stdinScanner)
			if userInput == "" {
				continue
			}
		case userInput == syncCommand: // Bare /sync
			a.InfoLog.Printf("Bare '%s' command received.", syncCommand)
			syncTargetDir := defaultSyncDir
			syncDirSource := "default"
			// Logic to determine syncTargetDir (unchanged)
			if a.Config.SyncDir != defaultSyncDir {
				syncTargetDir = a.Config.SyncDir
				syncDirSource = "-sync-dir flag"
			} else if cleanSandboxDir != defaultSandboxDir {
				syncTargetDir = a.Config.SandboxDir
				syncDirSource = "-sandbox flag"
			}
			a.InfoLog.Printf("Determined sync target directory '%s' (Source: %s)", syncTargetDir, syncDirSource)
			fmt.Printf("[AGENT] Starting sync for directory '%s' (using %s)...\n", syncTargetDir, syncDirSource)
			absSyncDir, secErr := core.ResolveAndSecurePath(syncTargetDir, cleanSandboxDir) // Use new validator
			if secErr != nil {
				a.ErrorLog.Printf("Sync FAIL: Invalid target %q: %v", syncTargetDir, secErr)
				fmt.Printf("[AGENT] Sync FAIL: Invalid target %q: %v\n", syncTargetDir, secErr)
				continue
			}
			dirInfo, statErr := os.Stat(absSyncDir)
			if statErr != nil || !dirInfo.IsDir() {
				a.ErrorLog.Printf("Sync FAIL: Cannot access %q: %v", syncTargetDir, statErr)
				fmt.Printf("[AGENT] Sync FAIL: Cannot access %q\n", syncTargetDir)
				continue
			}
			stats, syncErr := core.SyncDirectoryUpHelper(ctx, absSyncDir, a.Config.SyncFilter, a.Config.SyncIgnoreGitignore, llmClient.Client(), a.InfoLog, a.ErrorLog, a.DebugLog)
			if syncErr != nil {
				a.ErrorLog.Printf("Sync command (bare) FAIL: %v", syncErr)
				fmt.Printf("[AGENT] Sync FAIL for '%s': %v\n", syncTargetDir, syncErr)
			} else {
				a.InfoLog.Printf("Sync command (bare) OK. Stats: %+v", stats)
				fmt.Printf("[AGENT] Sync OK for '%s'.\n", syncTargetDir)
				// Update context using the validated absolute path
				updateAccumulatedURIs(ctx, a, llmClient, absSyncDir, a.Config.SyncFilter, &accumulatedContextURIs)
			}
			continue
		case strings.HasPrefix(userInput, syncCommand+" "): // /sync <dir> [filter]
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
			a.InfoLog.Printf("Sync command args: Dir='%s', Filter='%s'", syncDirArg, syncFilterArg)
			fmt.Printf("[AGENT] Starting sync for directory '%s'...\n", syncDirArg)
			absSyncDir, secErr := core.ResolveAndSecurePath(syncDirArg, cleanSandboxDir) // Use new validator
			if secErr != nil {
				a.ErrorLog.Printf("Sync FAIL: Invalid dir %q: %v", syncDirArg, secErr)
				fmt.Printf("[AGENT] Sync FAIL: Invalid dir %q: %v\n", syncDirArg, secErr)
				continue
			}
			dirInfo, statErr := os.Stat(absSyncDir)
			if statErr != nil || !dirInfo.IsDir() {
				a.ErrorLog.Printf("Sync FAIL: Cannot access %q: %v", syncDirArg, statErr)
				fmt.Printf("[AGENT] Sync FAIL: Cannot access %q\n", syncDirArg)
				continue
			}
			stats, syncErr := core.SyncDirectoryUpHelper(ctx, absSyncDir, syncFilterArg, false, llmClient.Client(), a.InfoLog, a.ErrorLog, a.DebugLog)
			if syncErr != nil {
				a.ErrorLog.Printf("Sync command (args) FAIL: %v", syncErr)
				fmt.Printf("[AGENT] Sync FAIL for '%s': %v\n", syncDirArg, syncErr)
			} else {
				a.InfoLog.Printf("Sync command (args) OK. Stats: %+v", stats)
				fmt.Printf("[AGENT] Sync OK for '%s'.\n", syncDirArg)
				// Update context using the validated absolute path
				updateAccumulatedURIs(ctx, a, llmClient, absSyncDir, syncFilterArg, &accumulatedContextURIs)
			}
			continue
		case userInput == "":
			continue
		default: // Process as LLM Prompt
			convoManager.AddUserMessage(userInput)
			a.DebugLog.Printf("Added user message to history: %q", userInput)
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

// handleMultilineInput uses nsinput... (unchanged)
func handleMultilineInput(a *App, scanner *bufio.Scanner) string { /* ... unchanged ... */
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
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
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
