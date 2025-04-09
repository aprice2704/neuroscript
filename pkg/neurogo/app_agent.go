// filename: pkg/neurogo/app_agent.go
package neurogo

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/neurodata/blocks"
)

// runAgentMode handles the agent initialization and interaction loop.
func (a *App) runAgentMode(ctx context.Context) error {
	a.InfoLog.Println("--- Starting NeuroGo in Agent Mode ---")

	// 1. Load Config / Initialize Components

	// --- Load Allowlist ---
	allowlist, errAllow := loadToolListFromFile(a.Config.AllowlistFile)
	if errAllow != nil {
		// Treat allowlist error as potentially serious - maybe default to empty?
		a.ErrorLog.Printf("Failed to load agent allowlist from %s: %v", a.Config.AllowlistFile, errAllow)
		a.ErrorLog.Println("CRITICAL: Proceeding with EMPTY allowlist. Agent will likely have no tools.")
		allowlist = []string{} // Default to empty allowlist on error
	} else {
		a.InfoLog.Printf("Loaded %d tools from allowlist: %s", len(allowlist), a.Config.AllowlistFile)
	}

	// --- Load Denylists ---
	denylistSet := make(map[string]bool) // Use a map for efficient denial checks
	// Mandatory denylist (ignore if not found)
	mandatoryDenyFile := "agent_denylist.ndtl.txt" // Or configure this name?
	mandatoryDenied, errMandatoryDeny := loadToolListFromFile(mandatoryDenyFile)
	if errMandatoryDeny != nil {
		if !os.IsNotExist(errMandatoryDeny) { // Log error only if it's not 'file not found'
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
	// Optional denylists from flags
	for _, denyFile := range a.Config.DenylistFiles {
		optionalDenied, errOptionalDeny := loadToolListFromFile(denyFile)
		if errOptionalDeny != nil {
			a.ErrorLog.Printf("Warning: Could not read optional denylist file %s: %v", denyFile, errOptionalDeny)
			// Continue processing other denylists if one fails
		} else {
			a.InfoLog.Printf("Loaded %d tools from optional denylist: %s", len(optionalDenied), denyFile)
			for _, tool := range optionalDenied {
				denylistSet[tool] = true
			}
		}
	}
	a.InfoLog.Printf("Total unique denied tools: %d", len(denylistSet))

	// --- Initialize Other Components ---
	cleanSandboxDir := filepath.Clean(a.Config.SandboxDir)
	a.InfoLog.Printf("Agent sandbox directory set to: %s", cleanSandboxDir)
	// TODO: Consider creating the sandbox dir or checking access/permissions

	llmClient := core.NewLLMClient(a.Config.APIKey, a.InfoLog)
	convoManager := core.NewConversationManager(a.InfoLog)
	agentInterpreter := core.NewInterpreter(a.DebugLog)
	core.RegisterCoreTools(agentInterpreter.ToolRegistry())
	blocks.RegisterBlockTools(agentInterpreter.ToolRegistry())

	// Pass allowlist AND denylist to SecurityLayer
	securityLayer := core.NewSecurityLayer(allowlist, denylistSet, cleanSandboxDir, agentInterpreter.ToolRegistry(), a.InfoLog)

	// --- Agent Interaction Loop ---
	a.InfoLog.Println("Enter your prompt for the agent (or type 'quit'):")
	stdinScanner := bufio.NewScanner(os.Stdin) // Renamed variable
	for stdinScanner.Scan() {
		userInput := stdinScanner.Text()
		if strings.ToLower(userInput) == "quit" {
			break
		}
		if userInput == "" {
			continue
		}

		convoManager.AddUserMessage(userInput)

		// Loop for potential function call cycles
		for i := 0; i < 5; i++ {
			a.InfoLog.Printf("--- Agent Turn %d ---", i+1)

			// Generate tool declarations (passing the original allowlist, not the denied set)
			toolDeclarations := core.GenerateToolDeclarations(agentInterpreter.ToolRegistry(), allowlist)

			// Call LLM
			response, llmErr := llmClient.CallLLMAgent(ctx, convoManager.GetHistory(), toolDeclarations)
			if llmErr != nil {
				a.ErrorLog.Printf("LLM API call failed: %v", llmErr)
				fmt.Printf("\n[AGENT] Error communicating with LLM: %v\n", llmErr)
				break // Break inner loop
			}

			if len(response.Candidates) == 0 {
				// Handle no candidates / safety blocks
				a.InfoLog.Println("LLM returned no candidates.")
				blockMsg := "[AGENT] LLM returned no response."
				if response.PromptFeedback != nil && response.PromptFeedback.BlockReason != "" {
					errMsg := fmt.Sprintf("Request blocked by safety filter: %s (%s)", response.PromptFeedback.BlockReason, response.PromptFeedback.BlockReasonMessage)
					a.ErrorLog.Printf("LLM Request Blocked: %s", errMsg)
					blockMsg = fmt.Sprintf("[AGENT] %s", errMsg)
				}
				fmt.Println("\n" + blockMsg)
				break // Break inner loop
			}

			candidate := response.Candidates[0]
			convoManager.AddModelResponse(candidate) // Add model response first

			if len(candidate.Content.Parts) == 0 {
				a.InfoLog.Println("LLM candidate had no parts.")
				fmt.Println("\n[AGENT] LLM returned an empty response part.")
				break
			}

			part := candidate.Content.Parts[0]

			if part.FunctionCall != nil {
				// Handle Function Call
				fc := part.FunctionCall
				a.InfoLog.Printf("Agent received FunctionCall request: %s", fc.Name)
				fmt.Printf("[AGENT] Requesting tool: %s\n", fc.Name)

				validatedArgs, validationErr := securityLayer.ValidateToolCall(fc.Name, fc.Args) // Validate first
				var toolResult map[string]interface{}

				if validationErr != nil {
					a.ErrorLog.Printf("Tool call validation failed for %s: %v", fc.Name, validationErr)
					fmt.Printf("[AGENT] Tool validation failed: %v\n", validationErr)
					toolResult = formatErrorResponse(validationErr)
				} else {
					// Execute only if validation passed
					a.InfoLog.Printf("Executing tool %s with validated args...", fc.Name)
					toolOutput, execErr := executeAgentTool(fc.Name, validatedArgs, agentInterpreter)
					toolResult = formatToolResult(toolOutput, execErr)
					if execErr != nil {
						a.ErrorLog.Printf("Tool execution failed for %s: %v", fc.Name, execErr)
						fmt.Printf("[AGENT] Tool execution failed: %v\n", execErr)
					} else {
						a.InfoLog.Printf("Tool %s executed successfully. Result map: %v", fc.Name, toolResult)
						// Optionally provide more user feedback about tool success/output?
						// fmt.Printf("[AGENT] Tool %s executed.\n", fc.Name)
					}
				}

				convoManager.AddFunctionResponse(fc.Name, toolResult)
				continue // Continue inner loop to send result back to LLM

			} else if part.Text != "" {
				// Handle Text Response
				a.InfoLog.Println("Agent received final Text response.")
				fmt.Printf("\n[AGENT] %s\n", part.Text)
				break // Final response for this user input, break inner loop
			} else {
				a.InfoLog.Println("Agent received response part with no text or function call.")
				fmt.Println("\n[AGENT] Received an empty response part from LLM.")
				break
			}
		} // End inner loop

		a.InfoLog.Println("\nEnter your prompt for the agent (or type 'quit'):")

	} // End outer loop (stdin scanner)

	if err := stdinScanner.Err(); err != nil {
		a.ErrorLog.Printf("Input scanner error: %v", err)
		return fmt.Errorf("error reading user input: %w", err)
	}

	a.InfoLog.Println("--- Exiting Agent Mode ---")
	return nil
}

// loadToolListFromFile reads tool names (one per line) from a file.
// Renamed from loadAllowlist for generic use.
func loadToolListFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		// Let caller handle os.IsNotExist specifically if needed
		return nil, fmt.Errorf("error opening tool list file '%s': %w", filePath, err)
	}
	defer file.Close()

	var tools []string
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "--") {
			continue
		}
		// Basic validation - could check TOOL. prefix etc. here
		if line == "" {
			// This check is redundant due to TrimSpace above, but harmless
			continue
		}
		tools = append(tools, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading tool list file '%s': %w", filePath, err)
	}
	return tools, nil
}

// executeAgentTool (unchanged)
func executeAgentTool(toolName string, args map[string]interface{}, interp *core.Interpreter) (interface{}, error) {
	interp.Logger().Printf("[AGENT TOOL] Attempting execution for tool '%s'", toolName)
	toolImpl, found := interp.ToolRegistry().GetTool(toolName)
	if !found {
		return nil, fmt.Errorf("internal error: agent tool '%s' not found in registry", toolName)
	}
	orderedArgs := make([]interface{}, len(toolImpl.Spec.Args))
	argIndexMap := make(map[string]int)
	for i, argSpec := range toolImpl.Spec.Args {
		argIndexMap[argSpec.Name] = i
		val, exists := args[argSpec.Name]
		if !exists {
			if argSpec.Required {
				return nil, fmt.Errorf("internal error: required argument '%s' missing for tool '%s' after validation", argSpec.Name, toolName)
			}
			orderedArgs[i] = nil
		} else {
			orderedArgs[i] = val
		}
	}
	if len(args) > len(toolImpl.Spec.Args) {
		extraArgs := []string{}
		for name := range args {
			if _, specExists := argIndexMap[name]; !specExists {
				extraArgs = append(extraArgs, name)
			}
		}
		if len(extraArgs) > 0 {
			interp.Logger().Printf("[WARN AGENT TOOL] Tool '%s' called with extra arguments not in spec: %v", toolName, extraArgs)
		}
	}
	interp.Logger().Printf("[AGENT TOOL] Calling %s with ordered args: %v", toolName, orderedArgs)
	toolOutput, execErr := toolImpl.Func(interp, orderedArgs)
	return toolOutput, execErr
}

// formatToolResult (unchanged)
func formatToolResult(toolOutput interface{}, execErr error) map[string]interface{} {
	resultMap := make(map[string]interface{})
	if execErr != nil {
		resultMap["success"] = false
		resultMap["error"] = execErr.Error()
		if toolOutput != nil {
			outputStr := fmt.Sprintf("%v", toolOutput)
			resultMap["partial_output"] = outputStr
		}
	} else {
		resultMap["success"] = true
		resultMap["result"] = toolOutput
	}
	return resultMap
}

// formatErrorResponse (unchanged)
func formatErrorResponse(err error) map[string]interface{} {
	return map[string]interface{}{"success": false, "error": err.Error()}
}
