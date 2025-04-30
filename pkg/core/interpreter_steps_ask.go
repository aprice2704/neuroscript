// filename: pkg/core/interpreter_steps_ask.go
package core

import (
	"context"
	"errors"
	"fmt"
	"strings" // Added for tool name prefix handling
	// Ensure necessary types like Expression, ConversationTurn, ToolCall, etc. are accessible
)

// executeAskAI handles the 'ask ai' step.
// *** MODIFIED: Accepts step Step instead of *AskAIStep ***
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
		// *** MODIFIED: Use ErrorCodeEvaluation, format message with position ***
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
		// Use correct error code constant
		// *** MODIFIED: Format message with position ***
		errMsg := fmt.Sprintf("LLM client is not configured in the interpreter (required at %s)", step.Pos.String())
		return NewRuntimeError(ErrorCodeLLMError, errMsg, nil)
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
		// Use correct error code constant and wrap original error
		// *** MODIFIED: Remove extra step.Pos argument, add pos to message ***
		errMsg := fmt.Sprintf("LLM interaction failed (ask at %s)", step.Pos.String())
		return NewRuntimeError(ErrorCodeLLMError, errMsg, err) // Corrected: Use ErrorCode constant and wrap error
	}
	if responseTurn == nil {
		// Use correct error code constant
		// *** MODIFIED: Remove extra step.Pos argument, add pos to message ***
		errMsg := fmt.Sprintf("LLM returned nil response without error (ask at %s)", step.Pos.String())
		return NewRuntimeError(ErrorCodeLLMError, errMsg, nil) // Corrected: Use ErrorCode constant
	}

	i.logger.Debug("LLM response received", "role", responseTurn.Role, "content_length", len(responseTurn.Content), "tool_calls", len(toolCalls))

	// 4. Process Response
	//    Update conversation history (implementation needed)
	i.addResponseToConversation(responseTurn)

	// Handle Tool Calls if any were returned
	if len(toolCalls) > 0 {
		err = i.handleToolCalls(toolCalls, step.Pos) // Execute tools and update history, pass position
		if err != nil {
			// If handleToolCalls returns an error, wrap it appropriately.
			// Assuming handleToolCalls returns a *RuntimeError or wrapped error:
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
			// Use evaluation error code if appropriate, or wrap generic error
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
// *** MODIFIED: Accepts step Step instead of *AskAIStep ***
func (i *Interpreter) getAvailableToolsForAsk(step Step) []ToolDefinition { // Return core.ToolDefinition
	if i.toolRegistry == nil {
		i.logger.Warn("getAvailableToolsForAsk called but toolRegistry is nil", "pos", step.Pos.String())
		return nil
	}
	// Corrected: Get ToolImplementation map, then convert Spec to Definition
	allToolsImpl := i.toolRegistry.GetAllTools()
	definitions := make([]ToolDefinition, 0, len(allToolsImpl)) // Use core.ToolDefinition

	for _, impl := range allToolsImpl {
		// Convert ToolSpec.Args ( []ArgSpec ) to ToolDefinition.InputSchema ( any / JSON schema map )
		inputSchema, err := ConvertToolSpecArgsToInputSchema(impl.Spec.Args)
		if err != nil {
			// Log error and skip tool if conversion fails
			i.logger.Warn("Skipping tool for LLM due to schema conversion error",
				"tool", impl.Spec.Name,
				"error", err,
				"pos", step.Pos.String(), // Log position of the ask step
			)
			continue // Skip this tool
		}

		definitions = append(definitions, ToolDefinition{ // Construct core.ToolDefinition
			Name:        impl.Spec.Name, // Use base name from ToolSpec
			Description: impl.Spec.Description,
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
		// *** MODIFIED: Use ErrorCodeConfiguration, format message, pass nil error ***
		errMsg := fmt.Sprintf("cannot handle tool calls at %s: toolRegistry is nil", pos.String())
		return NewRuntimeError(ErrorCodeConfiguration, errMsg, nil) // Corrected
	}

	results := make([]*ToolResult, len(calls)) // Ensure this is core.ToolResult

	for idx, call := range calls {
		// Initialize result structure for this call
		results[idx] = &ToolResult{ID: call.ID}

		i.logger.Debug("Processing tool call", "id", call.ID, "name", call.Name, "args", call.Arguments, "pos", pos.String())

		// --- Corrected Tool Execution Logic ---
		// 1. Find the tool implementation
		// *** MODIFIED: Use ToolPrefix constant ***
		baseToolName := strings.TrimPrefix(call.Name, ToolPrefix) // Assumes ToolPrefix = "TOOL."
		// Handle cases where LLM might return base name directly
		implName := baseToolName
		toolImpl, found := i.toolRegistry.GetTool(implName)
		// If not found with base name, try original name (in case prefix wasn't used)
		if !found && implName != call.Name {
			implName = call.Name
			toolImpl, found = i.toolRegistry.GetTool(implName)
		}

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
				orderedRawArgs[iSpec] = val // Use provided value
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
		// *** MODIFIED: Remove context argument from call ***
		resultVal, execErr := toolImpl.Func(i, validatedArgs) // Pass Interpreter `i`

		// 4. Record Result or Error
		if execErr != nil {
			// Check if it's already a RuntimeError
			runtimeErr, isRuntimeErr := execErr.(*RuntimeError)
			var errMsg string
			if isRuntimeErr {
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
		// --- End Corrected Tool Execution Logic ---
	}

	// Add tool results back to the conversation history
	i.addToolResultsToConversation(results) // Assumes this helper exists

	// Check if any tool failed critically. For now, errors are recorded in results.
	// If strict handling needed, loop through results and return first RuntimeError.
	// for _, res := range results {
	//  if res.Error != "" {
	//      // Maybe wrap the first error encountered
	//      return NewRuntimeError(ErrorCodeToolExecution, fmt.Sprintf("Tool call %s failed: %s", res.ID, res.Error), nil) // No original error to wrap easily here
	//  }
	// }

	return nil // Indicate overall handling success
}

// addToolResultsToConversation updates the managed conversation history with tool results.
func (i *Interpreter) addToolResultsToConversation(results []*ToolResult) {
	// TODO: Implement conversation history management logic for tool results
	i.logger.Debug("Adding tool results to conversation history (Not Implemented)", "count", len(results))
}

// --- Helper Function for Schema Conversion (Copied from previous correction - Consider moving) ---
// ConvertToolSpecArgsToInputSchema converts the []ArgSpec from a ToolImplementation
// into a JSON Schema-like map[string]interface{} suitable for ToolDefinition.InputSchema.
func ConvertToolSpecArgsToInputSchema(args []ArgSpec) (map[string]interface{}, error) {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
		"required":   []string{},
	}
	// Type assertion needed because map values are interface{}
	props, okProps := schema["properties"].(map[string]interface{})
	if !okProps {
		return nil, fmt.Errorf("internal error: could not assert schema properties as map[string]interface{}")
	}
	// Type assertion needed
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
		case ArgTypeList, ArgTypeSliceAny, ArgTypeSliceString:
			jsonType = "array"
		case ArgTypeMap:
			jsonType = "object"
		case ArgTypeAny:
			jsonType = "string" // Defaulting 'any' to string for schema
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

// --- Assumed Type Definitions (ensure they exist in relevant files) ---
// type Step struct { Pos *Position; Type string; Target string; Cond Expression; Value interface{}; ElseValue interface{}; Metadata map[string]string } // Defined in ast.go
// func (s Step) GetPos() *Position { return s.Pos } // Add this method to Step struct if needed elsewhere
// ... (other type definitions listed previously) ...
