// NeuroScript Version: 0.3.0
// File version: 0.1.0 // Updated version
// Minimal changes for compilation after refactor. Needs AI WM integration.
// filename: pkg/neurogo/handle_turn.go
package neurogo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/google/generative-ai-go/genai" // Still uses genai types heavily
)

// Constants remain the same
const (
	applyPatchFunctionName = "_ApplyNeuroScriptPatch"
	maxFunctionCallCycles  = 5
)

// llmRequestContext remains the same for now
type llmRequestContext struct {
	History  []*core.ConversationTurn
	FileURIs []string
}

// handleAgentTurn processes a single response from the LLM, potentially involving tool calls.
// NOTE: This function needs significant refactoring to work properly with the AI Worker Manager.
// It currently assumes a single LLM interaction flow and mixes core/genai types.
func (a *App) handleAgentTurn(
	ctx context.Context,
	llmClient core.LLMClient, // Accept the interface type
	convoManager *core.ConversationManager, // Assumes a single convo manager exists
	agentInterpreter *core.Interpreter, // Pass interpreter for tool execution context if needed
	securityLayer *core.SecurityLayer, // Pass security layer for validation
	toolDeclarations []*genai.Tool, // Pass the list of available tools
	initialAttachmentURIs []string, // Pass the file URIs for context
) error { // Return only error

	logger := a.GetLogger() // Use safe getter
	if logger == nil {
		// Should not happen if logger initialized correctly
		fmt.Println("Error: Logger not available in handleAgentTurn")
		return errors.New("logger not available in handleAgentTurn")
	}

	currentURIs := initialAttachmentURIs

	for cycle := 0; cycle < maxFunctionCallCycles; cycle++ {
		logger.Info("--- Agent Inner Loop Cycle ---", "cycle", cycle+1)

		// --- History Conversion (KEEPING FOR NOW - highlights type mismatch issue) ---
		genaiHistory := convoManager.GetHistory()
		coreHistory := make([]*core.ConversationTurn, 0, len(genaiHistory))
		// (Conversion logic remains the same for now)
		for _, content := range genaiHistory {
			if content == nil {
				logger.Warn("[CONVO] Skipping nil content during history conversion.")
				continue
			}
			turn := &core.ConversationTurn{Role: core.Role(content.Role)}
			var textContent strings.Builder
			var toolCalls []*core.ToolCall
			for _, part := range content.Parts {
				switch v := part.(type) {
				case genai.Text:
					textContent.WriteString(string(v))
				case genai.FunctionCall:
					toolCalls = append(toolCalls, &core.ToolCall{Name: v.Name, Arguments: v.Args})
					logger.Warn("[CONVO] Converting genai.FunctionCall found in history part to core.ToolCall (ID missing).")
				case genai.FunctionResponse:
					logger.Warn("[CONVO] genai.FunctionResponse found in history part, conversion to core.ToolResult not implemented.")
				default:
					logger.Warn("[CONVO] Unknown part type in history during conversion.", "type", fmt.Sprintf("%T", v))
				}
			}
			turn.Content = textContent.String()
			turn.ToolCalls = toolCalls
			coreHistory = append(coreHistory, turn)
		}
		// --- End History Conversion ---

		requestContext := llmRequestContext{
			History:  coreHistory,
			FileURIs: currentURIs,
		}

		if len(requestContext.History) == 0 {
			logger.Warn("Attempting agent call with empty history (after conversion).")
			return errors.New("cannot call LLM with empty history")
		}

		// Convert tool declarations (genai -> core)
		coreToolDefs := make([]core.ToolDefinition, 0, len(toolDeclarations))
		// (Conversion logic remains the same)
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
		llmResponseTurn, returnedToolCalls, err := llmClient.AskWithTools(ctx, requestContext.History, coreToolDefs) // Uses core types now
		if err != nil {
			logger.Error("LLM API call failed.", "error", err)
			return fmt.Errorf("LLM API call failed: %w", err)
		}

		if llmResponseTurn != nil {
			logger.Info("[CONVO] Received Model Turn:", "role", llmResponseTurn.Role, "content_snippet", snippet(llmResponseTurn.Content, 50), "tool_calls_in_turn", len(llmResponseTurn.ToolCalls))
			// Add response turn to history - Requires convoManager to accept core.ConversationTurn or conversion
			// convoManager.AddModelMessage(llmResponseTurn.Content) // Placeholder - needs update
			// TODO: Update convoManager or convert turn back to genai.Content
			logger.Warn("Need to update ConversationManager to handle core.ConversationTurn or convert turn back to genai.Content")
		} else if len(returnedToolCalls) > 0 {
			logger.Info("[CONVO] LLMClient returned Tool Calls but nil ConversationTurn.")
		} else {
			logger.Info("[CONVO] LLMClient returned nil ConversationTurn and no Tool Calls.")
		}

		foundFunctionCall := false
		var firstToolCallToExecute *core.ToolCall
		var accumulatedText strings.Builder

		if len(returnedToolCalls) > 0 {
			logger.Info("Processing tool calls returned separately by LLMClient.", "count", len(returnedToolCalls))
			foundFunctionCall = true
			firstToolCallToExecute = returnedToolCalls[0]
		} else if llmResponseTurn != nil && len(llmResponseTurn.ToolCalls) > 0 {
			logger.Warn("Tool calls found within ConversationTurn response object. Processing first.", "count", len(llmResponseTurn.ToolCalls))
			foundFunctionCall = true
			firstToolCallToExecute = llmResponseTurn.ToolCalls[0]
		}

		if !foundFunctionCall && llmResponseTurn != nil && llmResponseTurn.Content != "" {
			logger.Debug("LLM Response Turn: Text", "content", snippet(llmResponseTurn.Content, 50))
			accumulatedText.WriteString(llmResponseTurn.Content)
		}

		if foundFunctionCall && firstToolCallToExecute != nil {
			toolCall := *firstToolCallToExecute

			// --- Execute Tool Call (Requires Security Layer update) ---
			// SecurityLayer currently expects genai.FunctionCall. Needs update.
			// Temporary conversion:
			genaiFC := genai.FunctionCall{
				Name: toolCall.Name,
				Args: toolCall.Arguments,
			}
			logger.Info("Executing tool call.", "tool_name", genaiFC.Name)
			funcResultPart, execErr := securityLayer.ExecuteToolCall(agentInterpreter, genaiFC)
			// --- End Temp Conversion ---

			if execErr != nil {
				logger.Error("Tool execution failed.", "tool_name", genaiFC.Name, "error", execErr)
				funcResultPart = core.CreateErrorFunctionResultPart(genaiFC.Name, execErr)
			} else {
				logger.Info("Tool execution successful.", "tool_name", genaiFC.Name)
			}

			// Add result back to history (Requires convoManager update or conversion)
			if err := convoManager.AddFunctionResultMessage(funcResultPart); err != nil {
				logger.Error("Failed to add function result to conversation history.", "tool_name", genaiFC.Name, "error", err)
				return fmt.Errorf("failed to record function result for %s: %w", genaiFC.Name, err)
			}
			logger.Warn("Need to update ConversationManager to handle core.ToolResult or convert result back to genai.Part")
			// --- End History Update ---

			logger.Debug("Function call result added to history. Continuing agent cycle.")
			accumulatedText.Reset()
			continue

		} else {
			// No function call, handle final text
			finalText := accumulatedText.String()
			logger.Info("No function call requested or processed. Handling final text response.")
			logger.Debug("Final text response", "content", snippet(finalText, 100))

			// Patch Handling (Remains the same for now)
			if a.patchHandler != nil {
				trimmedResponse := strings.TrimSpace(finalText)
				if strings.HasPrefix(trimmedResponse, "@@@PATCH") && strings.HasSuffix(trimmedResponse, "@@@") {
					// (Patch logic remains the same)
					patchContent := strings.TrimPrefix(trimmedResponse, "@@@PATCH")
					patchContent = strings.TrimSuffix(patchContent, "@@@")
					patchContent = strings.TrimSpace(patchContent)
					patchErr := a.patchHandler.ApplyPatch(ctx, patchContent)
					if patchErr != nil {
						// ... error handling ...
						return fmt.Errorf("patch application failed: %w", patchErr)
					} else {
						// ... success handling ...
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
	}

	errLoop := fmt.Errorf("exceeded max agent cycles (%d)", maxFunctionCallCycles)
	logger.Error("Agent turn failed: max cycles exceeded.", "max_cycles", maxFunctionCallCycles)
	fmt.Printf("\n[AGENT] Error: %v\n", errLoop)
	return errLoop
}
