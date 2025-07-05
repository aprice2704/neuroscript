// NeuroScript Version: 0.3.0
// File version: 0.1.2
// Corrected compiler errors from typos and missing imports.
// filename: pkg/neurogo/handle_turn.go
package neurogo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/llm"
	"github.com/aprice2704/neuroscript/pkg/security"
	"github.com/google/generative-ai-go/genai"
)

const (
	applyPatchFunctionName = "_ApplyNeuroScriptPatch"
	maxFunctionCallCycles  = 5
)

type llmRequestContext struct {
	History  []*interfaces.ConversationTurn
	FileURIs []string
}

// createErrorFunctionResultPart is a helper to format a tool execution error.
func createErrorFunctionResultPart(toolName string, execErr error) genai.Part {
	errorContent := "An unknown error occurred."
	if execErr != nil {
		errorContent = execErr.Error()
	}
	return genai.FunctionResponse{
		Name:     toolName,
		Response: map[string]interface{}{"error": errorContent},
	}
}

func (a *App) handleAgentTurn(
	ctx context.Context,
	llmClient interfaces.LLMClient,
	convoManager *llm.ConversationManager,
	agentInterpreter *interpreter.Interpreter,
	securityLayer *security.SecurityLayer,
	toolDeclarations []*genai.Tool,
	initialAttachmentURIs []string,
) error {

	logger := a.GetLogger()
	if logger == nil {
		fmt.Println("Error: Logger not available in handleAgentTurn")
		return errors.New("logger not available in handleAgentTurn")
	}

	currentURIs := initialAttachmentURIs

	for cycle := 0; cycle < maxFunctionCallCycles; cycle++ {
		logger.Info("--- Agent Inner Loop Cycle ---", "cycle", cycle+1)

		genaiHistory := convoManager.GetHistory()
		coreHistory := make([]*interfaces.ConversationTurn, 0, len(genaiHistory))
		for _, content := range genaiHistory {
			if content == nil {
				logger.Warn("[CONVO] Skipping nil content during history conversion.")
				continue
			}
			turn := &interfaces.ConversationTurn{Role: interfaces.Role(content.Role)}
			var textContent strings.Builder
			var toolCalls []*interfaces.ToolCall
			for _, part := range content.Parts {
				switch v := part.(type) {
				case genai.Text:
					textContent.WriteString(string(v))
				case genai.FunctionCall:
					toolCalls = append(toolCalls, &interfaces.ToolCall{Name: v.Name, Arguments: v.Args})
					logger.Warn("[CONVO] Converting genai.FunctionCall found in history part to ToolCall (ID missing).")
				case genai.FunctionResponse:
					logger.Warn("[CONVO] genai.FunctionResponse found in history part, conversion to ToolResult not implemented.")
				default:
					logger.Warn("[CONVO] Unknown part type in history during conversion.", "type", fmt.Sprintf("%T", v))
				}
			}
			turn.Content = textContent.String()
			turn.ToolCalls = toolCalls
			coreHistory = append(coreHistory, turn)
		}

		requestContext := llmRequestContext{
			History:  coreHistory,
			FileURIs: currentURIs,
		}

		if len(requestContext.History) == 0 {
			logger.Warn("Attempting agent call with empty history (after conversion).")
			return errors.New("cannot call LLM with empty history")
		}

		coreToolDefs := make([]interfaces.ToolDefinition, 0, len(toolDeclarations))
		for _, genaiTool := range toolDeclarations {
			if len(genaiTool.FunctionDeclarations) > 0 {
				decl := genaiTool.FunctionDeclarations[0]
				coreToolDefs = append(coreToolDefs, interfaces.ToolDefinition{
					Name:        decl.Name,
					Description: decl.Description,
					InputSchema: decl.Parameters,
				})
			} else {
				logger.Warn("Tool declaration found with no FunctionDeclarations.", "tool", fmt.Sprintf("%+v", genaiTool))
			}
		}

		logger.Debug("Calling LLM.", "history_len", len(requestContext.History), "uri_count", len(requestContext.FileURIs), "tool_def_count", len(coreToolDefs))
		llmResponseTurn, returnedToolCalls, err := llmClient.AskWithTools(ctx, requestContext.History, coreToolDefs)
		if err != nil {
			logger.Error("LLM API call failed.", "error", err)
			return fmt.Errorf("LLM API call failed: %w", err)
		}

		if llmResponseTurn != nil {
			logger.Info("[CONVO] Received Model Turn:", "role", llmResponseTurn.Role, "content_snippet", snippet(llmResponseTurn.Content, 50), "tool_calls_in_turn", len(llmResponseTurn.ToolCalls))
			logger.Warn("Need to update ConversationManager to handle ConversationTurn or convert turn back to genai.Content")
		} else if len(returnedToolCalls) > 0 {
			logger.Info("[CONVO] LLMClient returned Tool Calls but nil ConversationTurn.")
		} else {
			logger.Info("[CONVO] LLMClient returned nil ConversationTurn and no Tool Calls.")
		}

		foundFunctionCall := false
		var firstToolCallToExecute *interfaces.ToolCall
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
			genaiFC := genai.FunctionCall{
				Name: toolCall.Name,
				Args: toolCall.Arguments,
			}
			logger.Info("Executing tool call.", "tool_name", genaiFC.Name)
			funcResultPart, execErr := securityLayer.ExecuteToolCall(agentInterpreter, genaiFC)

			if execErr != nil {
				logger.Error("Tool execution failed.", "tool_name", genaiFC.Name, "error", execErr)
				funcResultPart = createErrorFunctionResultPart(genaiFC.Name, execErr)
			} else {
				logger.Info("Tool execution successful.", "tool_name", genaiFC.Name)
			}

			if err := convoManager.AddFunctionResultMessage(funcResultPart); err != nil {
				logger.Error("Failed to add function result to conversation history.", "tool_name", genaiFC.Name, "error", err)
				return fmt.Errorf("failed to record function result for %s: %w", genaiFC.Name, err)
			}
			logger.Warn("Need to update ConversationManager to handle ToolResult or convert result back to genai.Part")
			logger.Debug("Function call result added to history. Continuing agent cycle.")
			accumulatedText.Reset()
			continue

		} else {
			finalText := accumulatedText.String()
			logger.Info("No function call requested or processed. Handling final text response.")
			logger.Debug("Final text response", "content", snippet(finalText, 100))

			if a.patchHandler != nil {
				trimmedResponse := strings.TrimSpace(finalText)
				if strings.HasPrefix(trimmedResponse, "@@@PATCH") && strings.HasSuffix(trimmedResponse, "@@@") {
					patchContent := strings.TrimPrefix(trimmedResponse, "@@@PATCH")
					patchContent = strings.TrimSuffix(patchContent, "@@@")
					patchContent = strings.TrimSpace(patchContent)

					logger.Info("Patch directive found. Deferring actual patch application.", "content_snippet", snippet(patchContent, 30))
					// patchErr := (*a.patchHandler).ApplyPatch(ctx, patchContent)
					// if patchErr != nil {
					// 	logger.Error("Patch application failed", "error", patchErr)
					// 	fmt.Fprintf(os.Stderr, "\n[AGENT] Error applying patch: %v\n", patchErr)
					// } else {
					// 	logger.Info("Patch applied successfully by deferred handler.")
					// 	fmt.Println("\n[AGENT] Patch applied successfully (simulated).")
					// }
					return nil
				}
			} else {
				logger.Warn("Patch handler is nil, cannot check for or apply patches.")
			}

			if finalText != "" {
				fmt.Printf("\n[AGENT RESPONSE]\n%s\n\n", finalText)
			} else {
				logger.Info("Agent provided no text response and no function call.")
				fmt.Println("\n[AGENT RESPONSE]\n(Agent provided no text response.)")
			}
			return nil
		}
	}

	errLoop := fmt.Errorf("exceeded max agent cycles (%d)", maxFunctionCallCycles)
	logger.Error("Agent turn failed: max cycles exceeded.", "max_cycles", maxFunctionCallCycles)
	fmt.Printf("\n[AGENT] Error: %v\n", errLoop)
	return errLoop
}
