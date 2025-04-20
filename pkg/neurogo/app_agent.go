// filename: pkg/neurogo/app_agent.go
package neurogo

import (
	"bufio"
	"context" // Added for json check within handleAgentTurn workaround
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/neurodata/blocks"
	checklist "github.com/aprice2704/neuroscript/pkg/neurodata/checklist"
)

const (
	applyPatchFunctionName = "_ApplyNeuroScriptPatch"
	maxFunctionCallCycles  = 5
)

// runAgentMode handles the agent setup and the main user input loop.
func (a *App) runAgentMode(ctx context.Context) error {
	a.InfoLog.Println("--- Starting NeuroGo in Agent Mode ---")

	// --- Load Config / Initialize Components ---
	allowlist, errAllow := loadToolListFromFile(a.Config.AllowlistFile)
	if errAllow != nil {
		a.ErrorLog.Printf("Failed to load agent allowlist from %s: %v", a.Config.AllowlistFile, errAllow)
		a.ErrorLog.Println("CRITICAL: Proceeding with EMPTY allowlist. Agent will likely have no tools.")
		allowlist = []string{}
	} else {
		a.InfoLog.Printf("Loaded %d tools from allowlist: %s", len(allowlist), a.Config.AllowlistFile)
	}
	denylistSet := make(map[string]bool)
	mandatoryDenyFile := "agent_denylist.ndtl.txt"
	mandatoryDenied, errMandatoryDeny := loadToolListFromFile(mandatoryDenyFile)
	if errMandatoryDeny != nil {
		if !os.IsNotExist(errMandatoryDeny) {
			a.ErrorLog.Printf("Warning: Could not read mandatory denylist file %s: %v", mandatoryDenyFile, errMandatoryDeny)
		} else {
			a.InfoLog.Printf("Mandatory denylist file %s not found, none loaded.", mandatoryDenyFile)
		}
	} else {
		a.InfoLog.Printf("Loaded %d tools from mandatory denylist: %s", len(mandatoryDenied), mandatoryDenyFile)
		for _, tool := range mandatoryDenied {
			denylistSet[tool] = true
		}
	}
	for _, denyFile := range a.Config.DenylistFiles {
		optionalDenied, errOptionalDeny := loadToolListFromFile(denyFile)
		if errOptionalDeny != nil {
			a.ErrorLog.Printf("Warning: Could not read optional denylist file %s: %v", denyFile, errOptionalDeny)
		} else {
			a.InfoLog.Printf("Loaded %d tools from optional denylist: %s", len(optionalDenied), denyFile)
			for _, tool := range optionalDenied {
				denylistSet[tool] = true
			}
		}
	}
	a.InfoLog.Printf("Total unique denied tools: %d", len(denylistSet))
	cleanSandboxDir := filepath.Clean(a.Config.SandboxDir)
	a.InfoLog.Printf("Agent sandbox directory set to: %s", cleanSandboxDir)
	llmClient := core.NewLLMClient(a.Config.APIKey, a.Config.ModelName, a.LLMLog)
	convoManager := core.NewConversationManager(a.InfoLog)
	agentInterpreter := core.NewInterpreter(a.DebugLog)
	if err := agentInterpreter.SetModelName(a.Config.ModelName); err != nil {
		a.ErrorLog.Printf("Warning: Failed to set interpreter model name from config ('%s'): %v. Interpreter/Tools may use default model.", a.Config.ModelName, err)
	}
	coreRegistry := agentInterpreter.ToolRegistry()
	if coreRegistry == nil {
		return fmt.Errorf("internal error: Interpreter's ToolRegistry is nil after creation")
	}
	core.RegisterCoreTools(coreRegistry)
	if err := blocks.RegisterBlockTools(coreRegistry); err != nil {
		a.ErrorLog.Printf("CRITICAL: Failed to register blocks tools: %v", err)
	} else {
		a.DebugLog.Println("Registered blocks tools.")
	}
	if err := checklist.RegisterChecklistTools(coreRegistry); err != nil {
		a.ErrorLog.Printf("CRITICAL: Failed to register checklist tools: %v", err)
	} else {
		a.DebugLog.Println("Registered checklist tools.")
	}
	securityLayer := core.NewSecurityLayer(allowlist, denylistSet, cleanSandboxDir, coreRegistry, a.InfoLog)
	// --- End Initialization ---

	// --- Agent Interaction Loop ---
	a.InfoLog.Println("\nEnter your prompt for the agent (or type 'quit'):")
	stdinScanner := bufio.NewScanner(os.Stdin)
	for stdinScanner.Scan() {
		userInput := stdinScanner.Text()
		if strings.ToLower(userInput) == "quit" {
			break
		}
		if userInput == "" {
			continue
		}

		convoManager.AddUserMessage(userInput)

		// Call the turn handler function within a loop for function calls
		turnCompleted := false // Flag to see if the turn finished naturally (text response or handled action)
		for i := 0; i < maxFunctionCallCycles; i++ {
			a.InfoLog.Printf("--- Agent Turn %d ---", i+1)
			llmResponse, err := llmClient.CallLLMAgent(ctx, core.LLMRequestContext{History: convoManager.GetHistory(), FileURIs: nil}, nil) // Pass nil tools for now
			if err != nil {
				a.ErrorLog.Printf("LLM API call failed: %v", err)
				fmt.Printf("\n[AGENT] Error communicating with LLM: %v\n", err)
				turnCompleted = true // End this turn due to error
				break
			}

			// Delegate processing the response to the helper function
			// handleAgentTurn returns true if the turn should end (final text, handled patch, error), false if loop should continue (tool call)
			shouldEndTurn := a.handleAgentTurn(llmResponse, convoManager, agentInterpreter, securityLayer, cleanSandboxDir)
			if shouldEndTurn {
				turnCompleted = true
				break // Exit the function call cycle
			}
		} // End inner loop (function call cycle)

		if !turnCompleted {
			a.ErrorLog.Printf("Agent turn exceeded maximum function call cycles (%d).", maxFunctionCallCycles)
			fmt.Println("[AGENT] Error: Too many function call cycles.")
		}

		a.InfoLog.Println("\nEnter your prompt for the agent (or type 'quit'):")

	} // End outer loop (stdin scanner)

	if err := stdinScanner.Err(); err != nil {
		a.ErrorLog.Printf("Input scanner error: %v", err)
		return fmt.Errorf("error reading user input: %w", err)
	}
	a.InfoLog.Println("--- Exiting Agent Mode ---")
	return nil
}
