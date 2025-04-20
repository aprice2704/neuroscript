package neurogo

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/google/generative-ai-go/genai"
)

// handleAgentTurn processes a single response from the LLM.
// Returns true if the turn is complete (final text, patch handled, error), false if a tool call requires continuation.
func (a *App) handleAgentTurn(
	response *genai.GenerateContentResponse,
	convoManager *core.ConversationManager,
	agentInterpreter *core.Interpreter,
	securityLayer *core.SecurityLayer,
	cleanSandboxDir string,
) (turnComplete bool) {

	if response == nil {
		a.ErrorLog.Println("LLM API returned a nil response")
		fmt.Println("\n[AGENT] Received nil response from LLM.")
		return true
	} // End turn on nil response
	if len(response.Candidates) == 0 { /* ... safety block handling ... */
		a.InfoLog.Println("LLM returned no candidates.")
		blockMsg := "[AGENT] LLM returned no response."
		if response.PromptFeedback != nil {
			if response.PromptFeedback.BlockReason != genai.BlockReasonUnspecified {
				errMsg := fmt.Sprintf("Request blocked by safety filter: %s", response.PromptFeedback.BlockReason)
				a.ErrorLog.Printf("LLM Request Blocked: %s", errMsg)
				blockMsg = fmt.Sprintf("[AGENT] %s", errMsg)
			}
		}
		fmt.Println("\n" + blockMsg)
		return true
	} // End turn on no candidates

	candidate := response.Candidates[0]
	if err := convoManager.AddModelResponse(candidate); err != nil {
		a.ErrorLog.Printf("Failed to add model response to history: %v", err)
	} // Log error but continue processing
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		a.InfoLog.Println("LLM candidate had nil content or no parts.")
		fmt.Println("\n[AGENT] LLM returned an empty response part.")
		return true
	} // End turn on empty content

	part := candidate.Content.Parts[0]

	// --- Text Part Handling (including Patch Workaround) ---
	if txt, ok := part.(genai.Text); ok {
		textResponse := string(txt)
		trimmedResponse := strings.TrimSpace(textResponse)
		// Trim potential markdown fences robustly
		trimmedResponse = strings.TrimPrefix(trimmedResponse, "```json") // Handle json fence specifically
		trimmedResponse = strings.TrimPrefix(trimmedResponse, "```")
		trimmedResponse = strings.TrimSuffix(trimmedResponse, "```")
		trimmedResponse = strings.TrimSpace(trimmedResponse)

		// Check for Patch Pattern
		if strings.HasPrefix(trimmedResponse, applyPatchFunctionName+"(") && strings.HasSuffix(trimmedResponse, ")") {
			a.InfoLog.Printf("Agent detected Patch request via text pattern.")
			fmt.Printf("[AGENT] Received patch request (via text pattern).\n")
			startIndex := strings.Index(trimmedResponse, "(")
			endIndex := strings.LastIndex(trimmedResponse, ")")

			if startIndex != -1 && endIndex != -1 && endIndex > startIndex {
				extractedArg := strings.TrimSpace(trimmedResponse[startIndex+1 : endIndex])

				// --- MODIFIED: Simplified Unquoting Attempt ---
				// Try direct Unmarshal first, then try Unquote if that fails
				// This handles cases where LLM sends raw JSON vs escaped JSON within quotes
				var patchJSON string = extractedArg // Assume raw JSON initially

				// Check if it might be quoted+escaped (needs Unquote)
				// A simple check is if it starts/ends with quotes AND contains escapes
				if strings.HasPrefix(extractedArg, `"`) && strings.HasSuffix(extractedArg, `"`) && (strings.Contains(extractedArg, `\"`) || strings.Contains(extractedArg, `\\`)) {
					unquoted, errUnquote := strconv.Unquote(extractedArg)
					if errUnquote == nil {
						a.DebugLog.Printf("[DEBUG PATCH] Successfully unquoted extracted argument.")
						patchJSON = unquoted // Use unquoted version
					} else {
						// Log unquote error but proceed trying to unmarshal the original extractedArg
						a.ErrorLog.Printf("[WARN PATCH] Failed to unquote extracted patch argument, will try direct unmarshal. Error: %v. Extracted: %q", errUnquote, extractedArg)
						// patchJSON remains extractedArg
					}
				}
				// --- End Unquote Attempt ---

				// Call the patch handler
				patchErr := handleReceivedPatch(patchJSON, agentInterpreter, securityLayer, cleanSandboxDir, a.InfoLog, a.ErrorLog)

				// --- MODIFIED: Add Model Text Response & BREAK loop ---
				var resultMessage string
				if patchErr != nil {
					a.ErrorLog.Printf("Patch application failed: %v", patchErr)
					fmt.Printf("[AGENT] Patch application failed: %v\n", patchErr)
					resultMessage = fmt.Sprintf("Patch application failed: %v", patchErr)
				} else {
					a.InfoLog.Printf("Patch applied successfully.")
					fmt.Printf("[AGENT] Patch applied successfully.\n")
					resultMessage = "Patch applied successfully."
				}
				patchOutcomeContent := &genai.Content{Role: "model", Parts: []genai.Part{genai.Text(resultMessage)}}
				convoManager.History = append(convoManager.History, patchOutcomeContent)
				a.InfoLog.Printf("[CONVO] Added Patch Outcome Text: %q", resultMessage)
				return true // Patch handled, end the turn.

			} else { // Malformed parentheses
				a.InfoLog.Println("Agent treating text starting with pattern prefix but missing/malformed parens as regular text.")
				fmt.Printf("\n[AGENT] %s\n", textResponse)
				return true // Treat as final text response
			}

			// --- Handle Regular Text Response ---
		} else if textResponse != "" {
			a.InfoLog.Println("Agent received final Text response.")
			fmt.Printf("\n[AGENT] %s\n", textResponse)
			return true // Final response, end the turn.
		} else { // Empty text part
			a.InfoLog.Printf("Agent received empty text part.")
			fmt.Println("\n[AGENT] Received an unexpected or empty response part from LLM.")
			return true // End the turn.
		}
	} // --- End Text Part Handling ---

	// --- Handle Regular Function Calls ---
	if fc, ok := part.(genai.FunctionCall); ok {
		// IMPORTANT: Ensure applyPatchFunctionName is NOT handled here if using text workaround
		if fc.Name == applyPatchFunctionName {
			a.ErrorLog.Printf("ERROR: Received %s as FunctionCall but expected it via text pattern workaround!", applyPatchFunctionName)
			fmt.Printf("\n[AGENT] Internal Error: Unexpected FunctionCall format for patch.\n")
			return true // End turn due to unexpected state
		}

		// Handle regular TOOL.xxx calls
		a.InfoLog.Printf("Agent received FunctionCall request: %s", fc.Name)
		fmt.Printf("[AGENT] Requesting tool: %s\n", fc.Name)
		validatedArgs, validationErr := securityLayer.ValidateToolCall(fc.Name, fc.Args)
		var toolResult map[string]interface{}
		if validationErr != nil {
			a.ErrorLog.Printf("Tool call validation failed for %s: %v", fc.Name, validationErr)
			fmt.Printf("[AGENT] Tool validation failed: %v\n", validationErr)
			toolResult = formatErrorResponse(validationErr)
		} else {
			a.InfoLog.Printf("Executing tool %s with validated args...", fc.Name)
			toolOutput, execErr := executeAgentTool(fc.Name, validatedArgs, agentInterpreter)
			toolResult = formatToolResult(toolOutput, execErr)
			if execErr != nil {
				a.ErrorLog.Printf("Tool execution failed for %s: %v", fc.Name, execErr)
				fmt.Printf("[AGENT] Tool execution failed: %v\n", execErr)
			} else {
				a.InfoLog.Printf("Tool %s executed successfully. Result map: %v", fc.Name, toolResult)
			}
		}
		if err := convoManager.AddFunctionResponse(fc.Name, toolResult); err != nil {
			a.ErrorLog.Printf("Failed to add tool result response to history: %v", err)
			// If adding response fails, should probably end turn?
			return true
		}
		return false // Indicate loop should continue for LLM response to tool result
	} // --- End Function Call Handling ---

	// --- Handle Unexpected Part Types ---
	a.InfoLog.Printf("Agent received response part of unexpected type (%T).", part)
	fmt.Println("\n[AGENT] Received an unexpected response part from LLM.")
	return true // End turn, cannot process.
}
