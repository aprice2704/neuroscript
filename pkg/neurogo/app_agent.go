// filename: pkg/neurogo/app_agent.go
package neurogo

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec" // Added
	"path/filepath"

	// "strconv" // Not needed here anymore
	"strings"
	// +++ ADDED: For checking exit status +++
	"github.com/aprice2704/neuroscript/pkg/core"
)

// Constants
const (
	// applyPatchFunctionName = "_ApplyNeuroScriptPatch" // Keep if needed
	maxFunctionCallCycles = 5
	multiLineCommand      = "/m" // Changed trigger command
)

// runAgentMode handles the agent setup and the main user input loop.
func (a *App) runAgentMode(ctx context.Context) error {
	a.InfoLog.Println("--- Starting NeuroGo in Agent Mode ---")

	// --- Load Config / Initialize Components (Same as before) ---
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

	a.InfoLog.Printf("Enter your prompt (or type '%s' for multi-line, 'quit' to exit):", multiLineCommand)
	stdinScanner := bufio.NewScanner(os.Stdin)
	turnCounter := 0

	for { // Main input loop
		turnCounter++
		a.InfoLog.Printf("--- Agent Conversation Turn %d ---", turnCounter)

		fmt.Printf("\nPrompt (or '%s' for multi-line): ", multiLineCommand)
		if !stdinScanner.Scan() {
			if err := stdinScanner.Err(); err != nil {
				a.ErrorLog.Printf("Input scanner error: %v", err)
				return fmt.Errorf("error reading input: %w", err)
			}
			a.InfoLog.Println("Input stream closed.")
			break // Exit loop
		}
		userInput := strings.TrimSpace(stdinScanner.Text())

		// --- Check for Commands ---
		if strings.ToLower(userInput) == "quit" {
			a.InfoLog.Println("Quit command received.")
			break
		}

		if userInput == multiLineCommand {
			a.InfoLog.Printf("Multi-line input command ('%s') received. Launching nsinput...", multiLineCommand)
			fmt.Println("Launching multi-line editor (nsinput)...")

			// --- Create Temp File ---
			tempFile, err := os.CreateTemp("", "nsinput-*.txt")
			if err != nil {
				a.ErrorLog.Printf("Error creating temp file for nsinput: %v", err)
				fmt.Printf("[AGENT] Error: Could not create temporary file for input.\n")
				continue // Ask for input again
			}
			tempFilePath := tempFile.Name()
			tempFile.Close() // Close the file handle immediately, nsinput will open/write it
			a.DebugLog.Printf("Created temp file: %s", tempFilePath)
			// --- Defer removal ---
			defer func() {
				a.DebugLog.Printf("Removing temp file: %s", tempFilePath)
				removeErr := os.Remove(tempFilePath)
				if removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) { // Don't log error if already gone
					a.ErrorLog.Printf("Error removing temp file %s: %v", tempFilePath, removeErr)
				}
			}()
			// ---

			// --- Execute nsinput, connecting terminal ---
			cmd := exec.Command("nsinput", tempFilePath) // Pass temp file path as argument
			cmd.Stdin = os.Stdin                         // Inherit neurogo's stdin
			cmd.Stdout = os.Stdout                       // Inherit neurogo's stdout (Bubble Tea needs this)
			cmd.Stderr = os.Stderr                       // Inherit neurogo's stderr

			runErr := cmd.Run() // Use Run(), not Output()

			if runErr != nil {
				// Check if it was just a cancellation (Ctrl+C in nsinput might cause non-zero exit)
				// Or if the command failed for other reasons (not found, internal error)
				a.ErrorLog.Printf("Error running 'nsinput': %v", runErr)
				// Check for specific exit errors if possible, otherwise print generic message
				fmt.Printf("[AGENT] Multi-line input editor exited with error or was cancelled.\n")
				// Check if temp file exists and is empty - might indicate cancellation
				contentBytes, readErr := os.ReadFile(tempFilePath)
				if readErr == nil && len(contentBytes) == 0 {
					a.InfoLog.Println("nsinput exited and left empty file - likely cancelled.")
				} else if readErr != nil {
					a.ErrorLog.Printf("Error reading temp file %s after nsinput exit: %v", tempFilePath, readErr)
				}
				continue // Ask for input again
			}
			// ---

			// --- Read result from Temp File ---
			contentBytes, readErr := os.ReadFile(tempFilePath)
			if readErr != nil {
				a.ErrorLog.Printf("Error reading temp file %s after nsinput success: %v", tempFilePath, readErr)
				fmt.Printf("[AGENT] Error: Could not read input from editor.\n")
				continue // Ask for input again
			}
			userInput = string(contentBytes)
			// ---

			// Check if cancelled (nsinput writes empty file on cancel)
			if userInput == "" {
				a.InfoLog.Println("Multi-line input cancelled (empty temp file).")
				fmt.Println("[AGENT] Multi-line input cancelled.")
				continue // Ask for input again
			}

			// Trim trailing newline that might be added by editor/file write
			userInput = strings.TrimSpace(userInput)
			fmt.Printf("[AGENT] Multi-line input received (%d characters).\n", len(userInput))

		} else if userInput == "" {
			a.InfoLog.Println("Empty input received, waiting for next prompt.")
			continue
		}
		// --- End Input Handling ---

		// --- Process the input (single line or from nsinput) ---
		// Add user message BEFORE calling handleAgentTurn
		convoManager.AddUserMessage(userInput)
		a.DebugLog.Printf("Added user message to history: %q", userInput)

		// Call the turn handler function
		errTurn := a.handleAgentTurn(ctx, llmClient, convoManager, agentInterpreter, securityLayer, toolDeclarations)
		if errTurn != nil {
			a.ErrorLog.Printf("Error during agent turn %d: %v", turnCounter, errTurn)
			fmt.Printf("\n[AGENT] Error processing turn: %v\n", errTurn)
		}
		// --- End Turn Processing ---

	} // End main input loop

	// Check scanner error after loop (e.g., if Scan returned false due to error)
	if err := stdinScanner.Err(); err != nil && err != io.EOF {
		a.ErrorLog.Printf("Input scanner error after loop: %v", err)
		return fmt.Errorf("error reading user input: %w", err)
	}

	a.InfoLog.Println("--- Exiting Agent Mode ---")
	return nil
}
