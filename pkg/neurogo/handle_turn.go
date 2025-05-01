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

// Define locally until added to core/llm_types.go
type llmRequestContext struct {
	History  []*core.ConversationTurn // Use core type
	FileURIs []string
}

// handleAgentTurn processes a single response from the LLM, potentially involving tool calls.
// It manages the inner loop of interacting with the LLM and tools.
func (a *App) handleAgentTurn(
	ctx context.Context,
	llmClient core.LLMClient, // Accept the interface type
	convoManager *core.ConversationManager,
	agentInterpreter *core.Interpreter, // Pass interpreter for tool execution context if needed
	securityLayer *core.SecurityLayer, // Pass security layer for validation
	toolDeclarations []*genai.Tool, // Pass the list of available tools
	initialAttachmentURIs []string, // Pass the file URIs for context
) error { // Return only error

	// Need logger inside this method
	logger := a.GetLogger()
	if logger == nil {
		fmt.Println("Error: Logger not available in handleAgentTurn")
		return errors.New("logger not available in handleAgentTurn")
	}

	currentURIs := initialAttachmentURIs // Start with initial URIs

	for cycle := 0; cycle < maxFunctionCallCycles; cycle++ {
		logger.Info("--- Agent Inner Loop Cycle ---", "cycle", cycle+1)

		// --- FIX: Convert History Format ---
		genaiHistory := convoManager.GetHistory()
		coreHistory := make([]*core.ConversationTurn, 0, len(genaiHistory))
		for _, content := range genaiHistory {
			if content == nil {
				logger.Warn("[CONVO] Skipping nil content during history conversion.")
				continue
			}
			turn := &core.ConversationTurn{
				Role: core.Role(content.Role), // Cast role
				// Simplified conversion: assumes first text part is main content
				// Ignores tool calls/results stored in genai.Content during this conversion for now
			}
			var textContent strings.Builder
			var toolCalls []*core.ToolCall // Collect core tool calls if present (unlikely in genai.Content usually)
			for _, part := range content.Parts {
				switch v := part.(type) {
				case genai.Text:
					textContent.WriteString(string(v))
				case genai.FunctionCall:
					// Convert genai.FunctionCall to core.ToolCall if needed
					toolCalls = append(toolCalls, &core.ToolCall{
						// ID is not directly available in genai.FunctionCall, generate or leave empty?
						Name:      v.Name,
						Arguments: v.Args,
					})
					logger.Warn("[CONVO] Converting genai.FunctionCall found in history part to core.ToolCall (ID missing).")
				case genai.FunctionResponse:
					// Convert genai.FunctionResponse to core.ToolResult if needed
					// TODO: Add conversion logic if ToolResult is needed in core.ConversationTurn
					logger.Warn("[CONVO] genai.FunctionResponse found in history part, conversion to core.ToolResult not implemented.")
				default:
					logger.Warn("[CONVO] Unknown part type in history during conversion.", "type", fmt.Sprintf("%T", v))
				}
			}
			turn.Content = textContent.String()
			turn.ToolCalls = toolCalls // Add converted/collected tool calls
			// turn.ToolResults = ... // Add converted tool results if needed
			coreHistory = append(coreHistory, turn)
		}
		// --- End History Conversion ---

		// Prepare request context for this cycle
		requestContext := llmRequestContext{ // Use local type definition
			History:  coreHistory, // Use converted history
			FileURIs: currentURIs,
		}

		if len(requestContext.History) == 0 {
			logger.Warn("Attempting agent call with empty history (after conversion).")
			return errors.New("cannot call LLM with empty history")
		}

		// Convert []*genai.Tool to []core.ToolDefinition
		coreToolDefs := make([]core.ToolDefinition, 0, len(toolDeclarations))
		for _, genaiTool := range toolDeclarations {
			if len(genaiTool.FunctionDeclarations) > 0 {
				decl := genaiTool.FunctionDeclarations[0]
				coreToolDefs = append(coreToolDefs, core.ToolDefinition{
					Name:        decl.Name,
					Description: decl.Description,
					InputSchema: decl.Parameters,
				})
			} else {
				logger.Warn("Tool declaration found with no FunctionDeclarations.", "tool", fmt.Sprintf("%+v", genaiTool))
			}
		}

		// Call LLM
		logger.Debug("Calling LLM.", "history_len", len(requestContext.History), "uri_count", len(requestContext.FileURIs), "tool_def_count", len(coreToolDefs))
		// --- FIX: Adapt to LLMClient return types ---
		llmResponseTurn, returnedToolCalls, err := llmClient.AskWithTools(ctx, requestContext.History, coreToolDefs)
		if err != nil {
			logger.Error("LLM API call failed.", "error", err)
			return fmt.Errorf("LLM API call failed: %w", err)
		}

		// Log the turn received from the LLM interface
		if llmResponseTurn != nil {
			logger.Info("[CONVO] Received Model Turn:", "role", llmResponseTurn.Role, "content_snippet", snippet(llmResponseTurn.Content, 50), "tool_calls_in_turn", len(llmResponseTurn.ToolCalls))
			// TODO: Convert llmResponseTurn back to genai.Content to add to history via convoManager, or update convoManager.
			// For now, we *don't* add this core.ConversationTurn back automatically.
			// The history update logic needs refinement based on core vs genai types.
		} else if len(returnedToolCalls) > 0 {
			logger.Info("[CONVO] LLMClient returned Tool Calls but nil ConversationTurn.")
		} else {
			logger.Info("[CONVO] LLMClient returned nil ConversationTurn and no Tool Calls.")
			// This might be a valid end state if the LLM just stops.
		}
		// --- End FIX ---

		// --- Process Response Parts ---
		foundFunctionCall := false
		var firstToolCallToExecute *core.ToolCall // Use core type
		var accumulatedText strings.Builder

		// --- FIX: Prioritize returnedToolCalls ---
		if len(returnedToolCalls) > 0 {
			logger.Info("Processing tool calls returned separately by LLMClient.", "count", len(returnedToolCalls))
			foundFunctionCall = true
			firstToolCallToExecute = returnedToolCalls[0] // Use core.ToolCall
			if len(returnedToolCalls) > 1 {
				logger.Warn("Multiple tool calls returned by LLMClient, only processing the first.", "ignored_count", len(returnedToolCalls)-1)
			}
		} else if llmResponseTurn != nil && len(llmResponseTurn.ToolCalls) > 0 {
			// Fallback: Check if tool calls are embedded in the turn object (less ideal)
			logger.Warn("Tool calls found within ConversationTurn response object. LLMClient should return them separately. Processing first.", "count", len(llmResponseTurn.ToolCalls))
			foundFunctionCall = true
			firstToolCallToExecute = llmResponseTurn.ToolCalls[0] // Use core.ToolCall
		}
		// --- End FIX ---

		// Process text content if no function call was prioritized, or if the turn also had text
		if !foundFunctionCall && llmResponseTurn != nil && llmResponseTurn.Content != "" {
			logger.Debug("LLM Response Turn: Text", "content", snippet(llmResponseTurn.Content, 50))
			accumulatedText.WriteString(llmResponseTurn.Content)
		}

		// --- Decide Next Action ---
		if foundFunctionCall && firstToolCallToExecute != nil {
			toolCall := *firstToolCallToExecute // Use the captured first call (core.ToolCall)

			// --- FIX: Convert core.ToolCall back to genai.FunctionCall for SecurityLayer ---
			// This highlights that SecurityLayer likely needs updating to use core types.
			genaiFC := genai.FunctionCall{
				Name: toolCall.Name,
				Args: toolCall.Arguments, // Assume map[string]any is compatible
			}
			logger.Info("Executing tool call.", "tool_name", genaiFC.Name)

			// Execute using genai.FunctionCall (as SecurityLayer expects it currently)
			funcResultPart, execErr := securityLayer.ExecuteToolCall(agentInterpreter, genaiFC) // Pass interpreter

			if execErr != nil {
				logger.Error("Tool execution failed.", "tool_name", genaiFC.Name, "error", execErr)
				funcResultPart = core.CreateErrorFunctionResultPart(genaiFC.Name, execErr)
			} else {
				logger.Info("Tool execution successful.", "tool_name", genaiFC.Name)
			}

			// Add the function result (genai.Part) back to the conversation manager (which expects genai.Part)
			if err := convoManager.AddFunctionResultMessage(funcResultPart); err != nil {
				logger.Error("Failed to add function result to conversation history.", "tool_name", genaiFC.Name, "error", err)
				return fmt.Errorf("failed to record function result for %s: %w", genaiFC.Name, err)
			}
			// --- End FIX ---

			logger.Debug("Function call result added to history. Continuing agent cycle.")
			accumulatedText.Reset() // Reset text since we handled a tool call
			continue                // Go to the next cycle

		} else {
			// No function call requested or processed, handle final text output
			finalText := accumulatedText.String()
			logger.Info("No function call requested or processed. Handling final text response.")
			logger.Debug("Final text response", "content", snippet(finalText, 100))

			// Check for patch pattern
			if a.patchHandler != nil { // Corrected: Use receiver 'a'
				trimmedResponse := strings.TrimSpace(finalText)
				if strings.HasPrefix(trimmedResponse, "@@@PATCH") && strings.HasSuffix(trimmedResponse, "@@@") {
					logger.Info("Detected patch pattern in response.")
					patchContent := strings.TrimPrefix(trimmedResponse, "@@@PATCH")
					patchContent = strings.TrimSuffix(patchContent, "@@@")
					patchContent = strings.TrimSpace(patchContent)

					patchErr := a.patchHandler.ApplyPatch(ctx, patchContent) // Corrected: Use receiver 'a'
					if patchErr != nil {
						logger.Error("Applying patch failed.", "error", patchErr)
						fmt.Printf("[AGENT] Error applying patch: %v\n", patchErr)
						if finalText != "" {
							fmt.Printf("\n[AGENT RESPONSE (Patch Failed)]\n%s\n\n", finalText)
						}
						return fmt.Errorf("patch application failed: %w", patchErr)
					} else {
						logger.Info("Patch applied successfully.")
						fmt.Println("[AGENT] Patch applied successfully.")
						return nil
					}
				}
			} else {
				logger.Warn("Patch handler is nil, cannot check for or apply patches.")
			}

			// Output final text if not handled as patch
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
