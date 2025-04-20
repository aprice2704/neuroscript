// filename: pkg/neurogo/handle_turn.go
package neurogo

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/google/generative-ai-go/genai"
)

// Constants for patch detection and loop control
const (
	applyPatchFunctionName = "_ApplyNeuroScriptPatch"
)

// handleAgentTurn processes a single response from the LLM.
func (a *App) handleAgentTurn(
	ctx context.Context,
	llmClient *core.LLMClient,
	convoManager *core.ConversationManager,
	agentInterpreter *core.Interpreter,
	securityLayer *core.SecurityLayer,
	toolDeclarations []*genai.Tool,
) error {

	for cycle := 0; cycle < maxFunctionCallCycles; cycle++ {
		a.InfoLog.Printf("--- Agent Inner Loop Cycle %d ---", cycle+1)

		requestContext := core.LLMRequestContext{History: convoManager.GetHistory()}
		if len(requestContext.History) == 0 {
			return errors.New("internal error: conversation history is empty before LLM call")
		}

		llmResponse, err := llmClient.CallLLMAgent(ctx, requestContext, toolDeclarations)
		// --- Handle LLM call errors / safety blocks (Same as before) ---
		if err != nil {
			return fmt.Errorf("LLM API call failed: %w", err)
		}
		if llmResponse == nil || len(llmResponse.Candidates) == 0 {
			blockReason := genai.BlockReasonUnspecified
			if llmResponse != nil && llmResponse.PromptFeedback != nil {
				blockReason = llmResponse.PromptFeedback.BlockReason
			}
			if blockReason != genai.BlockReasonUnspecified {
				errMsg := fmt.Sprintf("request blocked by safety filter: %s", blockReason)
				a.ErrorLog.Printf("LLM Request Blocked: %s", errMsg)
				fmt.Printf("\n[AGENT] %s\n", errMsg)
				return fmt.Errorf(errMsg)
			}
			a.ErrorLog.Println("LLM response was nil or had no valid candidates.")
			fmt.Println("\n[AGENT] Received empty or invalid response from LLM.")
			return errors.New("received empty or invalid LLM response")
		}
		// ---

		candidate := llmResponse.Candidates[0]
		if err := convoManager.AddModelResponse(candidate); err != nil {
			a.ErrorLog.Printf("Failed to add model response to history: %v", err)
		}

		if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
			a.InfoLog.Println("LLM candidate had nil content or no parts. Turn ends.")
			fmt.Println("\n[AGENT] LLM returned an empty response part.")
			return nil // End turn normally
		}

		// --- Process Response Parts ---
		foundFunctionCall := false
		var firstFunctionCall *genai.FunctionCall
		var accumulatedText strings.Builder

		for _, part := range candidate.Content.Parts { // Process parts (same as before)
			switch v := part.(type) {
			case genai.Text:
				accumulatedText.WriteString(string(v))
			case genai.FunctionCall:
				if !foundFunctionCall {
					foundFunctionCall = true
					fcCopy := v
					firstFunctionCall = &fcCopy
					a.InfoLog.Printf("LLM requested Function Call: %s", fcCopy.Name)
					a.DebugLog.Printf("Function Call Args: %v", fcCopy.Args)
				} else {
					a.InfoLog.Printf("Ignoring subsequent function call in same response: %s", v.Name)
				}
			default:
				a.DebugLog.Printf("LLM response part: Unexpected type %T", v)
			}
		} // End part processing loop

		// --- Decide Next Action ---
		if foundFunctionCall && firstFunctionCall != nil {
			// Execute the function call (same as before)
			fc := *firstFunctionCall
			funcResultPart, execErr := securityLayer.ExecuteToolCall(agentInterpreter, fc)
			if execErr != nil {
				a.ErrorLog.Printf("Tool execution error for '%s': %v", fc.Name, execErr)
			} else {
				a.InfoLog.Printf("Successfully executed tool: %s", fc.Name)
			}
			if err := convoManager.AddFunctionResultMessage(funcResultPart); err != nil {
				a.ErrorLog.Printf("Error adding func result message: %v", err)
				return fmt.Errorf("failed to record func result: %w", err)
			}
			a.DebugLog.Printf("Added function result to history for %s.", fc.Name)
			a.DebugLog.Println("Function call processed, continuing inner loop cycle.")
			continue // Go to next iteration

		} else {
			// No function call. Process text / patch.
			finalText := accumulatedText.String()
			trimmedResponse := strings.TrimSpace(finalText)

			// --- MODIFIED: Trim BOTH triple and single backticks ---
			trimmedResponse = strings.TrimPrefix(trimmedResponse, "```json")
			trimmedResponse = strings.TrimPrefix(trimmedResponse, "```")
			trimmedResponse = strings.TrimSuffix(trimmedResponse, "```")
			trimmedResponse = strings.TrimPrefix(trimmedResponse, "`") // Trim leading single backtick
			trimmedResponse = strings.TrimSuffix(trimmedResponse, "`") // Trim trailing single backtick
			trimmedResponse = strings.TrimSpace(trimmedResponse)       // Trim again just in case
			// --- END MODIFIED ---

			// Check for Patch Pattern
			if strings.HasPrefix(trimmedResponse, applyPatchFunctionName+"(") && strings.HasSuffix(trimmedResponse, ")") {
				a.InfoLog.Printf("Agent detected Patch request via text pattern.")
				fmt.Printf("[AGENT] Received patch request (via text pattern).\n")
				startIndex := strings.Index(trimmedResponse, "(")
				endIndex := strings.LastIndex(trimmedResponse, ")")

				if startIndex != -1 && endIndex > startIndex {
					extractedArg := strings.TrimSpace(trimmedResponse[startIndex+1 : endIndex])
					patchJSON := extractedArg

					// Unquoting Logic (unchanged)
					if strings.HasPrefix(extractedArg, `"`) && strings.HasSuffix(extractedArg, `"`) && (strings.Contains(extractedArg, `\"`) || strings.Contains(extractedArg, `\\`)) {
						unquoted, errUnquote := strconv.Unquote(extractedArg)
						if errUnquote == nil {
							a.DebugLog.Printf("[DEBUG PATCH] Successfully unquoted extracted argument.")
							patchJSON = unquoted
						} else {
							a.ErrorLog.Printf("[WARN PATCH] Failed to unquote, trying direct unmarshal. Err: %v. Arg: %q", errUnquote, extractedArg)
						}
					}

					// Call patch handler
					sandboxDir := securityLayer.SandboxRoot()
					patchErr := handleReceivedPatch(patchJSON, agentInterpreter, securityLayer, sandboxDir, a.InfoLog, a.ErrorLog)

					if patchErr != nil {
						a.ErrorLog.Printf("Patch application failed: %v", patchErr)
						fmt.Printf("[AGENT] Patch application failed: %v\n", patchErr)
						return fmt.Errorf("patch application failed: %w", patchErr)
					} else {
						a.InfoLog.Printf("Patch applied successfully.")
						fmt.Printf("[AGENT] Patch applied successfully.\n")
						return nil // Patch handled, end the turn successfully.
					}
				} else { // Malformed parentheses
					a.InfoLog.Println("Agent treating text matching pattern prefix but malformed parens as regular text.")
					fmt.Printf("\n[AGENT RESPONSE]\n%s\n\n", finalText) // Print original accumulated text
					return nil                                          // End the turn normally
				}
			} else if finalText != "" {
				// Regular text response
				a.InfoLog.Println("Agent received final Text response.")
				fmt.Printf("\n[AGENT RESPONSE]\n%s\n\n", finalText)
				return nil // End the turn successfully
			} else {
				// Empty text part and no function call
				a.InfoLog.Printf("Agent received empty text part and no function call.")
				fmt.Println("\n[AGENT RESPONSE]\n(Agent provided no text response for this turn.)\n")
				return nil // End turn normally
			}
		} // End handling text/patch
	} // End inner loop

	// If loop finishes, we exceeded max cycles
	errLoop := fmt.Errorf("exceeded maximum function call cycles (%d)", maxFunctionCallCycles)
	a.ErrorLog.Printf("Agent turn failed: %v", errLoop)
	fmt.Printf("\n[AGENT] Error: %v. The agent may be stuck in a loop.\n", errLoop)
	return errLoop
}

// --- Helpers (executeAgentTool, formatToolResult, formatErrorResponse, loadToolListFromFile) ---
// Assume these exist correctly in helpers.go or elsewhere in pkg/neurogo
