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
	"github.com/google/generative-ai-go/genai"
)

// Constants
const (
	// applyPatchFunctionName = "_ApplyNeuroScriptPatch"
	maxFunctionCallCycles = 5
	multiLineCommand      = "/m"
	syncCommand           = "/sync"
	defaultSyncDir        = "." // Default value for Config.SyncDir
	defaultSandboxDir     = "." // Default value for Config.SandboxDir
)

// runAgentMode handles the agent setup and the main user input loop.
func (a *App) runAgentMode(ctx context.Context) error {
	a.InfoLog.Println("--- Starting NeuroGo in Agent Mode ---")

	// --- Load Config / Initialize Components ---
	allowlist, _ := loadToolListFromFile(a.Config.AllowlistFile)
	denylistSet := make(map[string]bool)
	// Clean the sandbox dir path provided by the flag ONCE
	cleanSandboxDir := filepath.Clean(a.Config.SandboxDir)
	a.DebugLog.Printf("Using cleaned sandbox directory for security: %s", cleanSandboxDir)

	llmClient := a.llmClient
	if llmClient == nil || llmClient.Client() == nil {
		return fmt.Errorf("LLM Client not initialized")
	}

	convoManager := core.NewConversationManager(a.InfoLog)
	agentInterpreter := core.NewInterpreter(a.DebugLog, llmClient)
	coreRegistry := agentInterpreter.ToolRegistry()
	core.RegisterCoreTools(coreRegistry)
	// SECURITY LAYER USES cleanSandboxDir
	securityLayer := core.NewSecurityLayer(allowlist, denylistSet, cleanSandboxDir, coreRegistry, a.InfoLog)
	toolDeclarations, genErr := securityLayer.GetToolDeclarations()
	if genErr != nil {
		a.ErrorLog.Printf("Failed tool declarations: %v", genErr)
		return fmt.Errorf("failed gen tools: %w", genErr)
	}
	a.InfoLog.Printf("Initialized agent components. Sandbox: %s, Allowlist: %d, Tools: %d", cleanSandboxDir, len(allowlist), len(toolDeclarations))
	// --- End Initialization ---

	// --- Process Initial Attachments (Validate against cleanSandboxDir) ---
	initialAttachmentURIs := []string{}
	if len(a.Config.InitialAttachments) > 0 {
		a.InfoLog.Printf("Processing %d initial file attachments...", len(a.Config.InitialAttachments))
		if llmClient.Client() != nil {
			for _, attachPath := range a.Config.InitialAttachments {
				// VALIDATE against the actual security sandbox root
				absAttachPath, secErr := core.SecureFilePath(attachPath, cleanSandboxDir)
				if secErr != nil {
					a.ErrorLog.Printf("Skip attach %q sec err: %v", attachPath, secErr)
					continue
				}

				displayPath := attachPath
				relPath, relErr := filepath.Rel(cleanSandboxDir, absAttachPath)
				if relErr == nil {
					displayPath = filepath.ToSlash(relPath)
				}

				a.InfoLog.Printf("Uploading initial attachment: %s (Abs: %s, DisplayName: %s)", attachPath, absAttachPath, displayPath)
				uploadCtx := context.Background()
				apiFile, uploadErr := core.HelperUploadAndPollFile(uploadCtx, absAttachPath, displayPath, llmClient.Client(), a.LLMLog)
				if uploadErr != nil {
					a.ErrorLog.Printf("Failed attach %s: %v", attachPath, uploadErr)
				} else if apiFile != nil {
					a.InfoLog.Printf("Initial attach OK: %s -> URI: %s", attachPath, apiFile.URI)
					initialAttachmentURIs = append(initialAttachmentURIs, apiFile.URI)
				}
			}
			a.InfoLog.Printf("Finished initial attachments. URIs: %d: %v", len(initialAttachmentURIs), initialAttachmentURIs)
		} else {
			a.ErrorLog.Println("Cannot process initial attachments: LLM Client is nil.")
		}
	}
	// --- End Initial Attachments ---

	accumulatedContextURIs := []string{}
	accumulatedContextURIs = append(accumulatedContextURIs, initialAttachmentURIs...)

	// --- Agent Interaction Loop ---
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
		if !stdinScanner.Scan() { /* ... handle EOF/error ... */
			break
		}
		userInput := strings.TrimSpace(stdinScanner.Text())

		// --- Check for Meta Commands ---
		switch {
		case strings.ToLower(userInput) == "quit":
			a.InfoLog.Println("Quit command received.")
			return nil // Normal exit

		case userInput == multiLineCommand:
			// --- Handle /m (Multi-line Input via nsinput) ---
			userInput = handleMultilineInput(a, stdinScanner) // Extracted logic
			if userInput == "" {
				continue
			} // Re-prompt if cancelled/error
			// Fall through to process the multi-line input as the prompt below

		// +++ MODIFIED: Handle bare '/sync' command with Sandbox awareness +++
		case userInput == syncCommand:
			a.InfoLog.Printf("Bare '%s' command received.", syncCommand)

			// Determine sync target: Prioritize -sync-dir, then -sandbox, then default "."
			syncTargetDir := defaultSyncDir // Start with overall default
			syncDirSource := "default"

			// Check if -sync-dir was explicitly set (i.e., not the default ".")
			if a.Config.SyncDir != defaultSyncDir {
				syncTargetDir = a.Config.SyncDir
				syncDirSource = "-sync-dir flag"
			} else if cleanSandboxDir != defaultSandboxDir { // Check if -sandbox was explicitly set
				// Use the *original* path provided to -sandbox, not the cleaned one,
				// as the target, but validate against the cleaned one.
				syncTargetDir = a.Config.SandboxDir // Use the potentially relative path from flag
				syncDirSource = "-sandbox flag"
			}

			a.InfoLog.Printf("Determined sync target directory '%s' (Source: %s)", syncTargetDir, syncDirSource)
			fmt.Printf("[AGENT] Starting sync for directory '%s' (using %s)...\n", syncTargetDir, syncDirSource)

			// --- Call SyncDirectoryUpHelper Directly ---
			// Validate the chosen target directory against the security sandbox root
			absSyncDir, secErr := core.SecureFilePath(syncTargetDir, cleanSandboxDir)
			if secErr != nil {
				a.ErrorLog.Printf("Sync failed: Invalid target directory path %q (from %s): %v", syncTargetDir, syncDirSource, secErr)
				fmt.Printf("[AGENT] Sync failed: Invalid target directory path %q: %v\n", syncTargetDir, secErr)
				continue // Re-prompt
			}
			dirInfo, statErr := os.Stat(absSyncDir)
			if statErr != nil || !dirInfo.IsDir() {
				a.ErrorLog.Printf("Sync failed: Cannot access target directory %q (Abs: %s): %v", syncTargetDir, absSyncDir, statErr)
				fmt.Printf("[AGENT] Sync failed: Cannot access target directory %q\n", syncTargetDir)
				continue // Re-prompt
			}

			// Call the helper directly using determined/validated paths and config flags
			stats, syncErr := core.SyncDirectoryUpHelper(
				ctx,
				absSyncDir,                   // Use the validated absolute path for the helper
				a.Config.SyncFilter,          // Use filter from config
				a.Config.SyncIgnoreGitignore, // Use ignore flag from config
				llmClient.Client(),
				a.InfoLog, a.ErrorLog, a.DebugLog,
			)
			// --- End Sync Call ---

			if syncErr != nil {
				a.ErrorLog.Printf("Sync command (bare) failed: %v", syncErr)
				fmt.Printf("[AGENT] Sync failed for '%s': %v\n", syncTargetDir, syncErr)
			} else {
				a.InfoLog.Printf("Sync command (bare) successful. Stats: %+v", stats)
				fmt.Printf("[AGENT] Sync successful for '%s'.\n", syncTargetDir)
				// Sync succeeded, update accumulated URIs for the next turn
				updateAccumulatedURIs(ctx, a, llmClient, absSyncDir, a.Config.SyncFilter, &accumulatedContextURIs) // Extracted logic
			}
			// Sync command handled, re-prompt user
			continue

		// --- Handle /sync <dir> [filter] (Mostly Unchanged) ---
		case strings.HasPrefix(userInput, syncCommand+" "):
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
			a.InfoLog.Printf("Sync command with args received: Dir='%s', Filter='%s'", syncDirArg, syncFilterArg)
			fmt.Printf("[AGENT] Starting sync for directory '%s'...\n", syncDirArg)

			// Validate path against the security sandbox root
			absSyncDir, secErr := core.SecureFilePath(syncDirArg, cleanSandboxDir)
			if secErr != nil {
				a.ErrorLog.Printf("Sync failed: Invalid dir %q: %v", syncDirArg, secErr)
				fmt.Printf("[AGENT] Sync failed: Invalid dir %q: %v\n", syncDirArg, secErr)
				continue
			}
			dirInfo, statErr := os.Stat(absSyncDir)
			if statErr != nil || !dirInfo.IsDir() {
				a.ErrorLog.Printf("Sync failed: Cannot access dir %q (Abs: %s): %v", syncDirArg, absSyncDir, statErr)
				fmt.Printf("[AGENT] Sync failed: Cannot access dir %q\n", syncDirArg)
				continue
			}

			// Call Sync Helper (use default ignoreGitignore=false for interactive)
			stats, syncErr := core.SyncDirectoryUpHelper(ctx, absSyncDir, syncFilterArg, false, llmClient.Client(), a.InfoLog, a.ErrorLog, a.DebugLog)

			if syncErr != nil {
				a.ErrorLog.Printf("Sync command (args) failed: %v", syncErr)
				fmt.Printf("[AGENT] Sync failed for '%s': %v\n", syncDirArg, syncErr)
			} else {
				a.InfoLog.Printf("Sync command (args) successful. Stats: %+v", stats)
				fmt.Printf("[AGENT] Sync successful for '%s'.\n", syncDirArg)
				// Update context URIs based on this specific sync op
				updateAccumulatedURIs(ctx, a, llmClient, absSyncDir, syncFilterArg, &accumulatedContextURIs) // Use extracted helper
			}
			// Sync command handled, re-prompt user
			continue

		case userInput == "":
			continue // Skip empty lines, re-prompt

		// --- Default: Process as LLM Prompt ---
		default:
			convoManager.AddUserMessage(userInput)
			a.DebugLog.Printf("Added user message to history: %q", userInput)
			// Pass currentTurnURIs (from start of loop)
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

// handleMultilineInput uses nsinput to get multiline input. Returns the input or empty string on error/cancel.
func handleMultilineInput(a *App, scanner *bufio.Scanner) string {
	a.InfoLog.Printf("Launching nsinput for multi-line input...")
	fmt.Println("Launching multi-line editor (nsinput)...")
	tempFile, err := os.CreateTemp("", "nsinput-*.txt")
	if err != nil {
		a.ErrorLog.Printf("Failed to create temp file for nsinput: %v", err)
		fmt.Println("[AGENT] Error creating temp file for multi-line input.")
		return ""
	}
	tempFilePath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempFilePath)

	cmd := exec.Command("nsinput", tempFilePath)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	runErr := cmd.Run()
	if runErr != nil {
		a.ErrorLog.Printf("nsinput command finished with error: %v", runErr)
		fmt.Printf("[AGENT] Multi-line input command finished (may have been cancelled).\n")
		return "" // Indicate cancellation/error
	}

	contentBytes, readErr := os.ReadFile(tempFilePath)
	if readErr != nil {
		a.ErrorLog.Printf("Failed read temp file %s after nsinput: %v", tempFilePath, readErr)
		fmt.Println("[AGENT] Error reading input from multi-line editor.")
		return "" // Indicate error
	}
	userInput := string(contentBytes)
	if userInput == "" {
		a.InfoLog.Println("Multi-line input was empty.")
		fmt.Println("[AGENT] Multi-line input was empty.")
		// Return empty string, main loop will re-prompt
	}
	return strings.TrimSpace(userInput)
}

// updateAccumulatedURIs lists files after a sync and updates the shared URI list.
func updateAccumulatedURIs(
	ctx context.Context,
	a *App,
	llmClient *core.LLMClient,
	absSyncedDir string, // Absolute path of the directory that was just synced
	syncFilter string, // The filter used for the sync
	accumulatedContextURIs *[]string, // Pointer to the slice to update
) {
	a.InfoLog.Println("Listing API files to update context for next turn...")

	// We need the sandbox root to calculate relative display names for filtering
	cleanSandboxDir := filepath.Clean(a.Config.SandboxDir)
	syncDirRel := ""
	relPath, relErr := filepath.Rel(cleanSandboxDir, absSyncedDir)
	if relErr != nil {
		a.ErrorLog.Printf("Cannot get relative path for synced dir %s for context update: %v", absSyncedDir, relErr)
		// Proceed without prefix filtering if rel path fails? Or warn user?
		// Let's warn and proceed cautiously.
		fmt.Println("[AGENT] Warning: Could not determine relative path for context update filtering.")
	} else {
		syncDirRel = filepath.ToSlash(relPath)
	}
	if syncDirRel == "." {
		syncDirRel = ""
	} // Treat "." as root

	apiFiles, listErr := core.HelperListApiFiles(ctx, llmClient.Client(), a.DebugLog)
	if listErr != nil {
		a.ErrorLog.Printf("Failed to list API files after sync: %v", listErr)
		fmt.Println("[AGENT] Warning: Could not list files after sync to update context URIs.")
		return
	}

	urisCollected := 0
	newURIs := []string{}
	prefix := ""
	if syncDirRel != "" {
		prefix = syncDirRel + "/"
	} // Need trailing slash for prefix check

	for _, file := range apiFiles {
		if file.DisplayName == "" || file.State != genai.FileStateActive || file.URI == "" {
			continue
		}

		// Check if file is within the synced directory (using prefix)
		if strings.HasPrefix(file.DisplayName, prefix) {
			// Check filter pattern if provided
			if syncFilter != "" {
				match, _ := filepath.Match(syncFilter, filepath.Base(file.DisplayName))
				if !match {
					continue
				}
			}
			newURIs = append(newURIs, file.URI)
			urisCollected++
		}
	}

	// Add the *new* URIs found to the *accumulated* list, avoiding duplicates.
	uriSet := make(map[string]bool)
	for _, uri := range *accumulatedContextURIs {
		uriSet[uri] = true
	} // Dereference pointer
	for _, uri := range newURIs {
		uriSet[uri] = true
	}
	// Clear and rebuild the slice pointed to
	*accumulatedContextURIs = (*accumulatedContextURIs)[:0]
	for uri := range uriSet {
		*accumulatedContextURIs = append(*accumulatedContextURIs, uri)
	}

	a.InfoLog.Printf("Collected %d URIs from sync (Dir: '%s', Filter: '%s'). Total accumulated URIs: %d", urisCollected, absSyncedDir, syncFilter, len(*accumulatedContextURIs))
	fmt.Printf("[AGENT] Context updated with %d files from '%s'.\n", urisCollected, filepath.Base(absSyncedDir)) // Show base name for clarity
}

// Assume helpers loadToolListFromFile exist
// Assume handleAgentTurn exists
