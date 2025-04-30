// filename: pkg/core/interpreter_steps_ask.go
package core

import (
	"context"
	"fmt"
	// Ensure necessary types like Expression, ConversationTurn, ToolCall, etc. are accessible
	// Imports might be needed depending on where those types are defined.
)

// executeAskAI handles the 'ask ai' step.
func (i *Interpreter) executeAskAI(step *AskAIStep) error {
	i.logger.Debug("Executing 'ask ai' step")

	// 1. Prepare Conversation History
	//    This involves evaluating the prompt expression and potentially retrieving
	//    existing conversation history managed elsewhere.
	conversationTurns, err := i.prepareConversationForAsk(step.Prompt)
	if err != nil {
		return fmt.Errorf("failed to prepare conversation for 'ask ai': %w", err)
	}

	// 2. Get Available Tools
	//    Determine which tools should be sent to the LLM for this specific call.
	availableTools := i.getAvailableToolsForAsk() // Helper to get currently relevant tools

	var responseTurn *ConversationTurn
	var toolCalls []*ToolCall

	// 3. Call LLMClient using the interface methods
	if i.llmClient == nil {
		// Use specific error type if defined
		return NewRuntimeError(ErrLLMError, "LLM client is not configured in the interpreter", step.GetPos())
	}

	if len(availableTools) > 0 {
		i.logger.Debug("Calling LLM with tools", "turn_count", len(conversationTurns), "tool_count", len(availableTools))
		// Call AskWithTools directly on the interface
		responseTurn, toolCalls, err = i.llmClient.AskWithTools(context.Background(), conversationTurns, availableTools)
	} else {
		i.logger.Debug("Calling LLM without tools", "turn_count", len(conversationTurns))
		// Call Ask directly on the interface
		responseTurn, err = i.llmClient.Ask(context.Background(), conversationTurns)
	}

	// Handle LLM call errors
	if err != nil {
		// Ensure NewRuntimeError and ErrLLMError are defined (e.g., in errors.go)
		return NewRuntimeError(ErrLLMError, fmt.Sprintf("LLM interaction failed: %v", err), step.GetPos())
	}
	if responseTurn == nil {
		// Defensive check, should ideally be covered by the error return
		return NewRuntimeError(ErrLLMError, "LLM returned nil response without error", step.GetPos())
	}

	i.logger.Debug("LLM response received", "role", responseTurn.Role, "content_length", len(responseTurn.Content), "tool_calls", len(toolCalls))

	// 4. Process Response
	//    Update conversation history (implementation needed)
	i.addResponseToConversation(responseTurn)

	// Handle Tool Calls if any were returned
	if len(toolCalls) > 0 {
		err = i.handleToolCalls(toolCalls) // Execute tools and update history
		if err != nil {
			// Decide if tool execution errors should halt the script
			// For now, return the error. Could potentially log and continue.
			return fmt.Errorf("failed during tool execution: %w", err)
		}
		i.logger.Info("Tool calls requested by LLM were processed.")
		// NOTE: Depending on the desired flow, after handling tool calls,
		// you might need to call the LLM *again* with the tool results
		// to get a final natural language response. This requires looping logic.
	}

	// 5. Store Result
	//    The result of 'ask ai' is typically the assistant's text content.
	//    If the step assigns to a variable ('ask ai ... into myVar'), that logic
	//    would happen in the main executeStep function after this returns.
	i.lastCallResult = responseTurn.Content // Store the primary text response

	return nil
}

// --- Helper methods (Placeholders - require actual implementation or confirmation) ---

// prepareConversationForAsk evaluates the prompt and constructs the turn list for the LLM.
// Needs access to evaluateExpression and potentially conversation history management.
func (i *Interpreter) prepareConversationForAsk(promptExpr Expression) ([]*ConversationTurn, error) {
	promptVal, err := i.evaluateExpression(promptExpr) // Assumes evaluateExpression exists
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate prompt expression: %w", err)
	}
	promptStr, ok := promptVal.(string)
	if !ok {
		return nil, fmt.Errorf("prompt expression did not evaluate to a string, got %T", promptVal)
	}

	// Placeholder: Get actual conversation history if needed
	currentHistory := []*ConversationTurn{}

	// Construct the turns to send
	// May need to add system prompts, etc.
	finalHistory := append(currentHistory, &ConversationTurn{Role: RoleUser, Content: promptStr})

	return finalHistory, nil
}

// getAvailableToolsForAsk determines which tools to offer the LLM.
// Needs access to the ToolRegistry.
func (i *Interpreter) getAvailableToolsForAsk() []ToolDefinition {
	if i.toolRegistry == nil {
		i.logger.Warn("getAvailableToolsForAsk called but toolRegistry is nil")
		return nil
	}
	// Assumes ToolRegistry has a method to get all definitions
	return i.toolRegistry.GetAllToolDefinitions()
}

// addResponseToConversation updates the managed conversation history.
// Implementation depends on how conversation state is stored.
func (i *Interpreter) addResponseToConversation(turn *ConversationTurn) {
	// TODO: Implement conversation history management logic
	i.logger.Debug("Adding LLM response to conversation history (Not Implemented)", "role", turn.Role)
}

// handleToolCalls executes requested tool calls and adds results to the conversation.
// Needs access to the ToolRegistry and conversation management.
func (i *Interpreter) handleToolCalls(calls []*ToolCall) error {
	i.logger.Info("Handling tool calls requested by LLM", "count", len(calls))
	if len(calls) == 0 {
		return nil
	}
	if i.toolRegistry == nil {
		return fmt.Errorf("cannot handle tool calls: toolRegistry is nil")
	}

	results := make([]*ToolResult, len(calls))
	for idx, call := range calls {
		i.logger.Debug("Executing tool call", "id", call.ID, "name", call.Name, "args", call.Arguments)
		// Assumes ToolRegistry has ExecuteTool method
		resultVal, err := i.toolRegistry.ExecuteTool(context.Background(), call.Name, call.Arguments)

		// Create result struct regardless of error
		results[idx] = &ToolResult{
			ID:     call.ID,
			Result: resultVal, // Store result even if error occurred
		}
		if err != nil {
			errMsg := fmt.Sprintf("Tool '%s' execution failed: %v", call.Name, err)
			results[idx].Error = errMsg // Record error message in the result
			i.logger.Error("Tool execution failed", "id", call.ID, "name", call.Name, "error", err)
			// Continue processing other calls, errors are recorded in results
		} else {
			i.logger.Debug("Tool execution successful", "id", call.ID, "name", call.Name /*, "result", resultVal */) // Avoid logging potentially large results by default
		}
	}

	// Add tool results back to the conversation history
	i.addToolResultsToConversation(results) // Assumes this helper exists

	// Check if any tool failed critically (optional - could return combined error)
	// for _, res := range results {
	// 	if res.Error != "" {
	// 		return fmt.Errorf("one or more tool executions failed")
	// 	}
	// }

	return nil // Indicate overall handling success (individual errors are in results)
}

// addToolResultsToConversation updates the managed conversation history with tool results.
// Implementation depends on how conversation state is stored.
func (i *Interpreter) addToolResultsToConversation(results []*ToolResult) {
	// TODO: Implement conversation history management logic for tool results
	i.logger.Debug("Adding tool results to conversation history (Not Implemented)", "count", len(results))
}

// --- Assumed helper function/type definitions (ensure they exist) ---
// func (i *Interpreter) evaluateExpression(expr Expression) (interface{}, error) // Defined in evaluation_main.go or similar
// func (tr *ToolRegistry) GetAllToolDefinitions() []ToolDefinition // Defined in tools_registry.go
// func (tr *ToolRegistry) ExecuteTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) // Defined in tools_registry.go
// type Expression interface { ... } // Defined in ast.go
// type AskAIStep struct { Prompt Expression; StepPos } // Defined in ast.go
// type StepPos interface { GetPos() *Position } // Defined in ast.go
// type Position struct { ... } // Defined in ast.go
// type RuntimeError struct { ... } // Defined in errors.go
// func NewRuntimeError(code ErrorCode, message string, pos *Position) *RuntimeError // Defined in errors.go
// var ErrLLMError ErrorCode // Defined in errors.go
