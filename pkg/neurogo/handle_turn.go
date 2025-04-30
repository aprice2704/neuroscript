// filename: pkg/neurogo/handle_turn.go
package neurogo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	// Keep for placeholder ApiFileInfo if used here later
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/google/generative-ai-go/genai"
)

// Constants
const (
	applyPatchFunctionName = "_ApplyNeuroScriptPatch"
	maxFunctionCallCycles  = 5 // Limit recursive function calls
)

// handleAgentTurn processes a single response from the LLM, potentially involving tool calls.
// It manages the inner loop of interacting with the LLM and tools.
func (a *App) handleAgentTurn(
	ctx context.Context,
	llmClient core.LLMClient, // <<< FIX: Accept the interface type
	convoManager *core.ConversationManager,
	agentInterpreter *core.Interpreter, // Pass interpreter for tool execution context if needed
	securityLayer *core.SecurityLayer, // Pass security layer for validation
	toolDeclarations []*genai.Tool, // Pass the list of available tools
	initialAttachmentURIs []string, // Pass the file URIs for context
) error { // Return only error

	// Need logger inside this method
	logger := a.GetLogger()
	if logger == nil {
		// Fallback or return error if logger is essential
		fmt.Println("Error: Logger not available in handleAgentTurn")
		return errors.New("logger not available in handleAgentTurn")
	}

	currentURIs := initialAttachmentURIs // Start with initial URIs

	for cycle := 0; cycle < maxFunctionCallCycles; cycle++ {
		logger.Info("--- Agent Inner Loop Cycle ---", "cycle", cycle+1)

		// Prepare request context for this cycle
		requestContext := core.LLMRequestContext{
			History:  convoManager.GetHistory(), // Get current history
			FileURIs: currentURIs,               // Use URIs relevant for this cycle
		}

		if len(requestContext.History) == 0 {
			logger.Warn("Attempting agent call with empty history.")
			// Decide if this is an error or just needs skipping
			return errors.New("cannot call LLM with empty history")
		}

		// Call LLM
		logger.Debug("Calling LLM.", "history_len", len(requestContext.History), "uri_count", len(requestContext.FileURIs))
		llmResponse, err := llmClient.AskWithTools(ctx, requestContext.History, toolDeclarations) // Use AskWithTools
		if err != nil {
			// Log specific API error
			logger.Error("LLM API call failed.", "error", err)
			return fmt.Errorf("LLM API call failed: %w", err)
		}
		if llmResponse == nil || len(llmResponse.Candidates) == 0 {
			logger.Error("Invalid LLM response: nil or no candidates.")
			// Check safety ratings if available on llmResponse
			// ... safety rating check logic ...
			return errors.New("invalid LLM response (nil or no candidates)")
		}

		// Process first candidate
		candidate := llmResponse.Candidates[0]
		if err := convoManager.AddModelResponse(candidate); err != nil {
			// Log failure but might not be fatal for the turn
			logger.Error("Failed to add model response to conversation history.", "error", err)
		}
		if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
			logger.Info("LLM candidate has no content parts. Ending turn.")
			fmt.Println("\n[AGENT] LLM returned empty response content.")
			return nil // End turn successfully, nothing more to process
		}

		// --- Process Response Parts ---
		foundFunctionCall := false
		var firstFunctionCall *genai.FunctionCall
		var accumulatedText strings.Builder

		for _, part := range candidate.Content.Parts {
			switch v := part.(type) {
			case genai.Text:
				logger.Debug("LLM Response Part: Text", "content", snippet(string(v), 50))
				accumulatedText.WriteString(string(v))
			case genai.FunctionCall:
				logger.Debug("LLM Response Part: FunctionCall", "name", v.Name, "args", v.Args)
				if !foundFunctionCall {
					foundFunctionCall = true
					fcCopy := v // Make a copy since v is a loop variable
					firstFunctionCall = &fcCopy
					logger.Info("LLM requested tool call.", "tool_name", fcCopy.Name)
				} else {
					logger.Warn("Multiple function calls in one response, only processing the first.", "ignored_tool", v.Name)
				}
			default:
				logger.Warn("Unhandled LLM response part type.", "type", fmt.Sprintf("%T", v))
			}
		}
		// --- End Process Response Parts ---

		// --- Decide Next Action ---
		if foundFunctionCall && firstFunctionCall != nil {
			fc := *firstFunctionCall // Use the captured first call
			logger.Info("Executing tool call.", "tool_name", fc.Name)

			// TODO: Pass agentInterpreter correctly if tool execution needs it
			// funcResultPart, execErr := securityLayer.ExecuteToolCall(agentInterpreter, fc)
			// Placeholder execution - replace with actual call
			funcResultPart, execErr := securityLayer.ExecuteToolCall(fc) // Assuming ExecuteToolCall takes only FunctionCall now? Check definition

			if execErr != nil {
				logger.Error("Tool execution failed.", "tool_name", fc.Name, "error", execErr)
				// Format error as FunctionResponse for the LLM
				funcResultPart = core.CreateErrorFunctionResultPart(fc.Name, execErr)
				// Decide if tool execution error should stop the cycle or let LLM retry
			} else {
				logger.Info("Tool execution successful.", "tool_name", fc.Name)
			}

			// Add the function result back to the conversation
			if err := convoManager.AddFunctionResultMessage(funcResultPart); err != nil {
				// This is more critical, log and potentially return error
				logger.Error("Failed to add function result to conversation history.", "tool_name", fc.Name, "error", err)
				return fmt.Errorf("failed to record function result for %s: %w", fc.Name, err)
			}
			logger.Debug("Function call result added to history. Continuing agent cycle.")
			// Clear accumulated text for this cycle as we handled a function call
			accumulatedText.Reset()
			// URIs likely don't change based *only* on a tool call unless the tool itself updates context
			// currentURIs = ... update if tool modifies context state ...
			continue // Go to the next cycle, LLM will see the tool result

		} else {
			// No function call requested, process final text output
			finalText := accumulatedText.String()
			logger.Info("No function call requested by LLM. Processing final text response.")
			logger.Debug("Final text response", "content", snippet(finalText, 100))

			// Check for patch pattern (assuming handleReceivedPatch exists)
			if app.patchHandler != nil { // Check if patch handler is initialized
				trimmedResponse := strings.TrimSpace(finalText)
				// Simple check for patch prefix/suffix - refine as needed
				if strings.HasPrefix(trimmedResponse, "@@@PATCH") && strings.HasSuffix(trimmedResponse, "@@@") {
					logger.Info("Detected patch pattern in response.")
					// Extract patch content (needs robust extraction)
					patchContent := strings.TrimPrefix(trimmedResponse, "@@@PATCH")
					patchContent = strings.TrimSuffix(patchContent, "@@@")
					patchContent = strings.TrimSpace(patchContent)

					patchErr := app.patchHandler.ApplyPatch(ctx, patchContent)
					if patchErr != nil {
						logger.Error("Applying patch failed.", "error", patchErr)
						fmt.Printf("[AGENT] Error applying patch: %v\n", patchErr)
						// Let the user see the raw response that contained the failed patch
						if finalText != "" {
							fmt.Printf("\n[AGENT RESPONSE (Patch Failed)]\n%s\n\n", finalText)
						}
						return fmt.Errorf("patch application failed: %w", patchErr) // End turn with error
					} else {
						logger.Info("Patch applied successfully.")
						fmt.Println("[AGENT] Patch applied successfully.")
						return nil // End turn successfully after patch
					}
				}
			} else {
				logger.Warn("Patch handler is nil, cannot check for or apply patches.")
			}

			// If no patch detected or handler is nil, output the text
			if finalText != "" {
				fmt.Printf("\n[AGENT RESPONSE]\n%s\n\n", finalText)
			} else {
				logger.Info("Agent provided no text response and no function call.")
				fmt.Println("\n[AGENT RESPONSE]\n(Agent provided no text response.)")
			}
			return nil // End the turn successfully
		}
		// --- End Decide Next Action ---

	} // End inner loop cycles

	// If loop completes without returning, max cycles were exceeded
	errLoop := fmt.Errorf("exceeded max agent cycles (%d)", maxFunctionCallCycles)
	logger.Error("Agent turn failed: max cycles exceeded.", "max_cycles", maxFunctionCallCycles)
	fmt.Printf("\n[AGENT] Error: %v\n", errLoop)
	return errLoop
}
