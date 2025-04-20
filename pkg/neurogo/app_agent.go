// filename: pkg/neurogo/app_agent.go
package neurogo

import (
	"bufio"
	"context"

	// "encoding/json" // Not needed directly here

	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	// "strconv" // Not needed directly here
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core"
)

// Constants (unchanged)
const (
	// applyPatchFunctionName = "_ApplyNeuroScriptPatch"
	maxFunctionCallCycles = 5
	multiLineCommand      = "/m"
)

// runAgentMode handles the agent setup and the main user input loop.
func (a *App) runAgentMode(ctx context.Context) error {
	a.InfoLog.Println("--- Starting NeuroGo in Agent Mode ---")

	// --- Load Config / Initialize Components ---
	allowlist, _ := loadToolListFromFile(a.Config.AllowlistFile)
	denylistSet := make(map[string]bool) // Assume loaded
	cleanSandboxDir := filepath.Clean(a.Config.SandboxDir)
	llmClient := a.llmClient
	if llmClient == nil || llmClient.Client() == nil {
		return fmt.Errorf("LLM Client not initialized")
	}
	convoManager := core.NewConversationManager(a.InfoLog)
	agentInterpreter := core.NewInterpreter(a.DebugLog, llmClient)
	coreRegistry := agentInterpreter.ToolRegistry()
	core.RegisterCoreTools(coreRegistry) // Add other registrations
	securityLayer := core.NewSecurityLayer(allowlist, denylistSet, cleanSandboxDir, coreRegistry, a.InfoLog)
	toolDeclarations, _ := securityLayer.GetToolDeclarations()
	// --- End Initialization ---

	// +++ ADDED: Process Initial Attachments from Flags +++
	initialAttachmentURIs := []string{}
	if len(a.Config.InitialAttachments) > 0 {
		a.InfoLog.Printf("Processing %d initial file attachments from -attach flags...", len(a.Config.InitialAttachments))
		if llmClient.Client() == nil {
			a.ErrorLog.Println("Cannot process attachments: LLM Client (GenAI Client) is nil.")
		} else {
			for _, attachPath := range a.Config.InitialAttachments {
				// Validate path against sandbox
				absAttachPath, secErr := core.SecureFilePath(attachPath, cleanSandboxDir)
				if secErr != nil {
					a.ErrorLog.Printf("Skipping attachment: Invalid path %q: %v", attachPath, secErr)
					continue
				}
				// Determine display name (use relative path from sandbox)
				displayPath := attachPath // Default to original relative path
				relPath, relErr := filepath.Rel(cleanSandboxDir, absAttachPath)
				if relErr == nil {
					displayPath = filepath.ToSlash(relPath)
				}
				a.InfoLog.Printf("Uploading attachment: %s (Display Name: %s)", attachPath, displayPath)

				// Call the refactored upload helper
				apiFile, uploadErr := core.HelperUploadAndPollFile(ctx, absAttachPath, displayPath, llmClient.Client(), a.LLMLog) // Pass LLM Log
				if uploadErr != nil {
					a.ErrorLog.Printf("Failed to upload attachment %s: %v", attachPath, uploadErr)
					// Optionally decide if this should be a fatal error for the agent startup
				} else if apiFile != nil {
					a.InfoLog.Printf("Attachment successful: %s -> URI: %s", attachPath, apiFile.URI)
					initialAttachmentURIs = append(initialAttachmentURIs, apiFile.URI)
				}
			}
			a.InfoLog.Printf("Finished processing initial attachments. %d URIs collected.", len(initialAttachmentURIs))
		}
	}
	// --- END ADDED ---

	// --- Agent Interaction Loop ---
	a.InfoLog.Printf("Enter your prompt (or type '%s' for multi-line, 'quit' to exit):", multiLineCommand)
	stdinScanner := bufio.NewScanner(os.Stdin)
	turnCounter := 0

	for { // Main input loop
		turnCounter++
		a.InfoLog.Printf("--- Agent Conversation Turn %d ---", turnCounter)

		fmt.Printf("\nPrompt (or '%s' for multi-line): ", multiLineCommand)
		if !stdinScanner.Scan() { /* handle EOF/error */
			break
		}
		userInput := strings.TrimSpace(stdinScanner.Text())
		if strings.ToLower(userInput) == "quit" {
			a.InfoLog.Println("Quit command received.")
			break
		}

		// --- Handle /m for nsinput (Unchanged from previous version) ---
		if userInput == multiLineCommand {
			a.InfoLog.Printf("Multi-line input command ('%s') received. Launching nsinput...", multiLineCommand)
			// ... (logic to create temp file, run nsinput, read temp file) ...
			tempFile, err := os.CreateTemp("", "nsinput-*.txt")
			if err != nil { /* handle error */
				continue
			}
			tempFilePath := tempFile.Name()
			tempFile.Close()
			defer func() { os.Remove(tempFilePath) }() // Ensure cleanup
			cmd := exec.Command("nsinput", tempFilePath)
			cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
			runErr := cmd.Run()
			if runErr != nil {
				a.ErrorLog.Printf("Error running 'nsinput': %v", runErr)
				fmt.Printf("[AGENT] Multi-line input editor error/cancelled.\n")
				continue
			}
			contentBytes, readErr := os.ReadFile(tempFilePath)
			if readErr != nil { /* handle error */
				continue
			}
			userInput = string(contentBytes)
			if userInput == "" {
				a.InfoLog.Println("Multi-line input cancelled.")
				fmt.Println("[AGENT] Multi-line input cancelled.")
				continue
			}
			userInput = strings.TrimSpace(userInput)
			fmt.Printf("[AGENT] Multi-line input received (%d characters).\n", len(userInput))
		} else if userInput == "" {
			continue
		}
		// --- End /m handling ---

		// --- Process the input (single line or from nsinput) ---
		convoManager.AddUserMessage(userInput)
		a.DebugLog.Printf("Added user message to history: %q", userInput)

		// +++ MODIFIED: Pass initialAttachmentURIs to handleAgentTurn +++
		errTurn := a.handleAgentTurn(ctx, llmClient, convoManager, agentInterpreter, securityLayer, toolDeclarations, initialAttachmentURIs)
		if errTurn != nil {
			a.ErrorLog.Printf("Error during agent turn %d: %v", turnCounter, errTurn)
			fmt.Printf("\n[AGENT] Error processing turn: %v\n", errTurn)
		}
		// --- End Turn Processing ---

	} // End main input loop

	if err := stdinScanner.Err(); err != nil && err != io.EOF {
		a.ErrorLog.Printf("Scanner error: %v", err)
		return fmt.Errorf("input error: %w", err)
	}
	a.InfoLog.Println("--- Exiting Agent Mode ---")
	return nil
}

// handleAgentTurn (Signature needs update, logic to use fileURIs needs update)
// ... (Needs modification as outlined in Step 4) ...
