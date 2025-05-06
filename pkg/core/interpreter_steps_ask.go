// NeuroScript Version: 0.3.1
// File version: 0.0.2 // Replace GetAllTools with ListTools.
// filename: pkg/core/interpreter_steps_ask.go
package core

import (
	"context"
	"errors"
	"fmt"
	// Added for tool name prefix handling
	// Ensure necessary types like Expression, ConversationTurn, ToolCall, etc. are accessible
)

// executeAskAI handles the 'ask ai' step.
func (i *Interpreter) executeAskAI(step Step) error {
	i.logger.Debug("Executing 'ask ai' step", "pos", step.Pos.String()) // Log position

	// --- Get Prompt Expression from step.Value ---
	promptExpr, ok := step.Value.(Expression)
	if !ok {
		// This indicates an AST building error or incorrect step structure
		return NewRuntimeError(ErrorCodeInternal,
			fmt.Sprintf("Internal error at %s: 'ask' step value is not an Expression (got %T)", step.Pos.String(), step.Value),
			nil, // No underlying error to wrap
		)
	}
	// --- End Get Prompt ---

	// 1. Prepare Conversation History
	conversationTurns, err := i.prepareConversationForAsk(promptExpr)
	if err != nil {
		// Use specific error code, include position in message
		errMsg := fmt.Sprintf("failed to prepare conversation for 'ask ai' at %s: %v", step.Pos.String(), err)
		// Wrap original error if it's not already a RuntimeError
		var runtimeErr *RuntimeError
		if errors.As(err, &runtimeErr) {
			// If already a runtime error, use its code if desired, otherwise wrap it simply
			return NewRuntimeError(ErrorCodeEvaluation, errMsg, err)
		} else {
			return NewRuntimeError(ErrorCodeEvaluation, errMsg, err) // Wrap original Go error
		}
	}

	// 2. Get Available Tools
	availableTools := i.getAvailableToolsForAsk(step) // Pass step for context

	var responseTurn *ConversationTurn
	var toolCalls []*ToolCall // Ensure this is the core.ToolCall type

	// 3. Call LLMClient using the interface methods
	if i.llmClient == nil {
		errMsg := fmt.Sprintf("LLM client is not configured in the interpreter (required at %s)", step.Pos.String())
		return NewRuntimeError(ErrorCodeLLMError, errMsg, nil)
	}

	if len(availableTools) > 0 {
		i.logger.Debug("Calling LLM with tools", "turn_count", len(conversationTurns), "tool_count", len(availableTools))
		responseTurn, toolCalls, err = i.llmClient.AskWithTools(context.Background(), conversationTurns, availableTools)
	} else {
		i.logger.Debug("Calling LLM without tools", "turn_count", len(conversationTurns))
		responseTurn, err = i.llmClient.Ask(context.Background(), conversationTurns)
	}

	// Handle LLM call errors
	if err != nil {
		errMsg := fmt.Sprintf("LLM interaction failed (ask at %s)", step.Pos.String())
		return NewRuntimeError(ErrorCodeLLMError, errMsg, err) // Use ErrorCode constant and wrap error
	}
	if responseTurn == nil {
		errMsg := fmt.Sprintf("LLM returned nil response without error (ask at %s)", step.Pos.String())
		return NewRuntimeError(ErrorCodeLLMError, errMsg, nil) // Use ErrorCode constant
	}

	i.logger.Debug("LLM response received", "role", responseTurn.Role, "content_length", len(responseTurn.Content), "tool_calls", len(toolCalls))

	// 4. Process Response
	//    Update conversation history (placeholder)
	i.addResponseToConversation(responseTurn)

	// Handle Tool Calls if any were returned
	if len(toolCalls) > 0 {
		err = i.handleToolCalls(toolCalls, step.Pos) // Execute tools and update history, pass position
		if err != nil {
			// If handleToolCalls returns an error, wrap it appropriately.
			// Add position info if not already present
			var runtimeErr *RuntimeError
			if errors.As(err, &runtimeErr) {
				// Already a runtime error, return as is
				return err
			} else {
				// Wrap standard Go error
				return fmt.Errorf("failed during tool execution (initiated at %s): %w", step.Pos.String(), err)
			}
		}
		i.logger.Info("Tool calls requested by LLM were processed.")
		// NOTE: Depending on the desired flow, after handling tool calls,
		// you might need to call the LLM *again* with the tool results
		// to get a final natural language response. This requires looping logic.
	}

	// 5. Store Result (Primary text response)
	i.lastCallResult = responseTurn.Content // Store the primary text response
	// Assignment to target variable (if `ask ... into myVar`) happens in interpreter_exec.go

	return nil
}

// --- Helper methods ---

// prepareConversationForAsk evaluates the prompt and constructs the turn list for the LLM.
func (i *Interpreter) prepareConversationForAsk(promptExpr Expression) ([]*ConversationTurn, error) {
	promptVal, err := i.evaluateExpression(promptExpr) // Assumes evaluateExpression exists
	if err != nil {
		// Wrap evaluation error, preserving original if it's a RuntimeError
		var runtimeErr *RuntimeError
		if errors.As(err, &runtimeErr) {
			return nil, fmt.Errorf("failed to evaluate prompt expression: %w", err)
		} else {
			// Add position from the expression node itself
			errMsg := fmt.Sprintf("failed to evaluate prompt expression at %s: %v", promptExpr.GetPos().String(), err)
			return nil, NewRuntimeError(ErrorCodeEvaluation, errMsg, err)
		}
	}
	promptStr, ok := promptVal.(string)
	if !ok {
		// Use evaluation error code for type mismatch during prompt prep
		errMsg := fmt.Sprintf("prompt expression at %s did not evaluate to a string, got %T", promptExpr.GetPos().String(), promptVal)
		return nil, NewRuntimeError(ErrorCodeEvaluation, errMsg, nil)
	}

	// Placeholder: Get actual conversation history if managed by Interpreter
	currentHistory := []*ConversationTurn{} // Assuming empty for now

	// Construct the turns to send
	finalHistory := append(currentHistory, &ConversationTurn{Role: RoleUser, Content: promptStr})

	return finalHistory, nil
}

// getAvailableToolsForAsk determines which tools (as ToolDefinition for the LLM) to offer.
func (i *Interpreter) getAvailableToolsForAsk(step Step) []ToolDefinition {
	if i.toolRegistry == nil {
		i.logger.Warn("getAvailableToolsForAsk called but toolRegistry is nil", "pos", step.Pos.String())
		return nil
	}
	// *** FIXED: Use ListTools() which returns []ToolSpec ***
	allToolSpecs := i.toolRegistry.ListTools()
	definitions := make([]ToolDefinition, 0, len(allToolSpecs)) // Use core.ToolDefinition

	for _, spec := range allToolSpecs {
		// Convert ToolSpec.Args ( []ArgSpec ) to ToolDefinition.InputSchema ( any / JSON schema map )
		inputSchema, err := ConvertToolSpecArgsToInputSchema(spec.Args) // Use existing helper
		if err != nil {
			// Log error and skip tool if conversion fails
			i.logger.Warn("Skipping tool for LLM due to schema conversion error",
				"tool", spec.Name, // Use name from ToolSpec
				"error", err,
				"pos", step.Pos.String(), // Log position of the ask step
			)
			continue // Skip this tool
		}

		definitions = append(definitions, ToolDefinition{ // Construct core.ToolDefinition
			Name:        spec.Name, // Use name from ToolSpec
			Description: spec.Description,
			InputSchema: inputSchema,
		})
	}
	return definitions
}

// addResponseToConversation updates the managed conversation history.
func (i *Interpreter) addResponseToConversation(turn *ConversationTurn) {
	// TODO: Implement conversation history management logic if needed
	i.logger.Debug("Adding LLM response to conversation history (Not Implemented)", "role", turn.Role)
}

// handleToolCalls executes requested tool calls and adds results to the conversation.
func (i *Interpreter) handleToolCalls(calls []*ToolCall, pos *Position) error { // Accept position for error context
	i.logger.Info("Handling tool calls requested by LLM", "count", len(calls), "pos", pos.String())
	if len(calls) == 0 {
		return nil
	}
	if i.toolRegistry == nil {
		// Return a runtime error if the registry isn't available
		errMsg := fmt.Sprintf("cannot handle tool calls at %s: toolRegistry is nil", pos.String())
		return NewRuntimeError(ErrorCodeConfiguration, errMsg, nil)
	}

	results := make([]*ToolResult, len(calls)) // Ensure this is core.ToolResult

	for idx, call := range calls {
		// Initialize result structure for this call
		results[idx] = &ToolResult{ID: call.ID}

		i.logger.Debug("Processing tool call", "id", call.ID, "name", call.Name, "args", call.Arguments, "pos", pos.String())

		// --- Tool Execution Logic ---
		// 1. Find the tool implementation by name
		toolImpl, found := i.toolRegistry.GetTool(call.Name) // Use the name directly from the call

		if !found {
			errMsg := fmt.Sprintf("Tool '%s' not found in registry", call.Name)
			results[idx].Error = errMsg // Record error in the result
			i.logger.Error("Tool execution failed", "id", call.ID, "name", call.Name, "error", errMsg, "pos", pos.String())
			continue // Process next call
		}

		// 2. Convert and Validate Arguments
		orderedRawArgs := make([]interface{}, len(toolImpl.Spec.Args))
		argConversionError := false
		for iSpec, argSpec := range toolImpl.Spec.Args {
			val, exists := call.Arguments[argSpec.Name]
			if !exists {
				if argSpec.Required {
					errMsg := fmt.Sprintf("Missing required argument '%s' for tool '%s'", argSpec.Name, call.Name)
					results[idx].Error = errMsg
					i.logger.Error("Tool execution failed: Missing required arg", "id", call.ID, "name", call.Name, "arg", argSpec.Name, "error", errMsg, "pos", pos.String())
					argConversionError = true
					break // Stop processing args for this call
				}
				orderedRawArgs[iSpec] = nil // Use nil for missing optional arg
			} else {
				// Basic type check/conversion might be needed here if call.Arguments are map[string]interface{}
				// For now, assume the types are compatible enough for ValidateAndConvertArgs
				orderedRawArgs[iSpec] = val
			}
		}
		if argConversionError {
			continue // Process next call
		}

		// Use ValidateAndConvertArgs which takes the spec and the ordered raw arguments
		validatedArgs, validationErr := ValidateAndConvertArgs(toolImpl.Spec, orderedRawArgs)
		if validationErr != nil {
			errMsg := fmt.Sprintf("Argument validation failed for tool '%s': %v", call.Name, validationErr)
			results[idx].Error = errMsg
			i.logger.Error("Tool execution failed: Validation", "id", call.ID, "name", call.Name, "error", errMsg, "pos", pos.String())
			continue // Process next call
		}

		// 3. Execute the tool's function
		resultVal, execErr := toolImpl.Func(i, validatedArgs) // Pass Interpreter `i`

		// 4. Record Result or Error
		if execErr != nil {
			// Check if it's already a RuntimeError
			var runtimeErr *RuntimeError
			var errMsg string
			if errors.As(execErr, &runtimeErr) {
				errMsg = fmt.Sprintf("Tool '%s' execution failed: %s (code %d)", call.Name, runtimeErr.Message, runtimeErr.Code)
			} else {
				errMsg = fmt.Sprintf("Tool '%s' execution failed: %v", call.Name, execErr)
			}
			results[idx].Error = errMsg     // Record error message in the result
			results[idx].Result = resultVal // Store potentially partial result
			i.logger.Error("Tool execution failed", "id", call.ID, "name", call.Name, "error", errMsg, "pos", pos.String())
			// Continue processing other calls
		} else {
			results[idx].Result = resultVal                                                                          // Store successful result
			i.logger.Debug("Tool execution successful", "id", call.ID, "name", call.Name /*, "result", resultVal */) // Avoid logging potentially large results
		}
		// --- End Tool Execution Logic ---
	}

	// Add tool results back to the conversation history
	i.addToolResultsToConversation(results) // Assumes this helper exists

	return nil // Indicate overall handling success (individual errors are in results)
}

// addToolResultsToConversation updates the managed conversation history with tool results.
func (i *Interpreter) addToolResultsToConversation(results []*ToolResult) {
	// TODO: Implement conversation history management logic for tool results
	i.logger.Debug("Adding tool results to conversation history (Not Implemented)", "count", len(results))
}

// --- Helper Function for Schema Conversion ---
// ConvertToolSpecArgsToInputSchema converts the []ArgSpec from a ToolImplementation
// into a JSON Schema-like map[string]interface{} suitable for ToolDefinition.InputSchema.
func ConvertToolSpecArgsToInputSchema(args []ArgSpec) (map[string]interface{}, error) {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
		"required":   []string{},
	}
	props, okProps := schema["properties"].(map[string]interface{})
	if !okProps {
		return nil, fmt.Errorf("internal error: could not assert schema properties as map[string]interface{}")
	}
	required, okReq := schema["required"].([]string)
	if !okReq {
		return nil, fmt.Errorf("internal error: could not assert schema required as []string")
	}

	for _, argSpec := range args {
		var jsonType string
		switch argSpec.Type {
		case ArgTypeString:
			jsonType = "string"
		case ArgTypeInt:
			jsonType = "integer"
		case ArgTypeFloat:
			jsonType = "number"
		case ArgTypeBool:
			jsonType = "boolean"
		case ArgTypeSlice, ArgTypeSliceAny, ArgTypeSliceString, ArgTypeSliceInt, ArgTypeSliceFloat, ArgTypeSliceBool, ArgTypeSliceMap:
			jsonType = "array"
			// TODO: Add "items": {"type": "..."} based on specific slice type if possible/needed
		case ArgTypeMap:
			jsonType = "object"
			// TODO: Add "properties": {} or "additionalProperties": true/false/schema ?
		case ArgTypeAny:
			// Represent 'any' loosely. Could be object or string? String is simpler.
			jsonType = "string" // Or omit type? JSON Schema allows omitting type.
		case ArgTypeNil:
			continue // Nil is not a valid input type for a schema property
		default:
			return nil, fmt.Errorf("unsupported ArgType '%s' for tool argument '%s' cannot be converted to JSON schema", argSpec.Type, argSpec.Name)
		}

		propSchema := map[string]interface{}{"type": jsonType}
		if argSpec.Description != "" {
			propSchema["description"] = argSpec.Description
		}
		props[argSpec.Name] = propSchema
		if argSpec.Required {
			required = append(required, argSpec.Name)
		}
	}
	schema["required"] = required
	schema["properties"] = props
	return schema, nil
}
