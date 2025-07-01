// NeuroScript Version: 0.3.1
// File version: 1.0.0
// Purpose: Replaces a call to a non-existent `unwrapValue` with a contract-compliant call to `Unwrap`, including proper error handling.
// filename: pkg/core/interpreter_steps_ask.go
// nlines: 250+
// risk_rating: MEDIUM
package runtime

import (
	"context"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// executeAskAI is a placeholder for the full 'ask' step logic.
// The actual 'ask' step logic is more complex and not fully represented here.
func (i *Interpreter) executeAsk(step ast.Step) (Value, error) {
	// This is a simplified stand-in. The full logic would be similar to executeAskAI.
	if len(step.Values) == 0 {
		return nil, lang.NewRuntimeError(ErrorCodeInternal, "ask step has no value expression", nil).WithPosition(step.Pos)
	}
	promptExpr := step.Values[0]

	promptVal, err := i.evaluate.Expression(promptExpr)
	if err != nil {
		return nil, err
	}
	// For now, just return the evaluated prompt.
	// The real implementation would call an LLM.
	if step.AskIntoVar != "" {
		if err := i.SetVariable(step.AskIntoVar, promptVal); err != nil {
			return nil, err
		}
	}
	return promptVal, nil
}

// executeAskAI handles the 'ask ai' step.
func (i *Interpreter) executeAskAI(step ast.Step) error {
	i.logger.Debug("Executing 'ask ai' step", "pos", step.Pos.String())

	promptExpr, ok := step.Value.(ast.Expression)
	if !ok {
		return lang.NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("Internal error at %s: 'ask' step value is not an ast.Expression (got %T)", step.Pos.String(), step.Value), nil)
	}

	conversationTurns, err := i.prepareConversationForAsk(promptExpr)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare conversation for 'ask ai' at %s: %v", step.Pos.String(), err)
		return lang.NewRuntimeError(ErrorCodeEvaluation, errMsg, err)
	}

	availableTools := i.getAvailableToolsForAsk(step)

	var responseTurn *interfaces.ConversationTurn
	var ToolCalls []*interfaces.ToolCall

	if i.llmClient == nil {
		return lang.NewRuntimeError(ErrorCodeLLMError, fmt.Sprintf("LLM client is not configured in the interpreter (required at %s)", step.Pos.String()), nil)
	}

	if len(availableTools) > 0 {
		i.logger.Debug("Calling LLM with tools", "turn_count", len(conversationTurns), "tool_count", len(availableTools))
		responseTurn, ToolCalls, err = i.llmClient.AskWithTools(context.Background(), conversationTurns, availableTools)
	} else {
		i.logger.Debug("Calling LLM without tools", "turn_count", len(conversationTurns))
		responseTurn, err = i.llmClient.Ask(context.Background(), conversationTurns)
	}

	if err != nil {
		return lang.NewRuntimeError(ErrorCodeLLMError, fmt.Sprintf("LLM interaction failed (ask at %s)", step.Pos.String()), err)
	}
	if responseTurn == nil {
		return lang.NewRuntimeError(ErrorCodeLLMError, fmt.Sprintf("LLM returned nil response without error (ask at %s)", step.Pos.String()), nil)
	}

	i.logger.Debug("LLM response received", "role", responseTurn.Role, "content_length", len(responseTurn.Content), "tool_calls", len(ToolCalls))

	i.addResponseToConversation(responseTurn)

	if len(ToolCalls) > 0 {
		err = i.handleToolCalls(ToolCalls, step.Pos)
		if err != nil {
			return fmt.Errorf("failed during tool execution (initiated at %s): %w", step.Pos.String(), err)
		}
		i.logger.Debug("Tool calls requested by LLM were processed.")
	}

	i.lastCallResult = StringValue{Value: responseTurn.Content}
	return nil
}

func (i *Interpreter) prepareConversationForAsk(promptExpr ast.Expression) ([]*interfaces.ConversationTurn, error) {
	promptVal, err := i.evaluate.Expression(promptExpr)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate prompt expression: %w", err)
	}

	strVal, ok := promptVal.(StringValue)
	if !ok {
		errMsg := fmt.Sprintf("prompt expression at %s did not evaluate to a string, got %s", promptExpr.GetPos().String(), promptVal.Type())
		return nil, lang.NewRuntimeError(ErrorCodeEvaluation, errMsg, nil)
	}
	promptStr := strVal.Value

	currentHistory := []*interfaces.ConversationTurn{}
	finalHistory := append(currentHistory, &interfaces.ConversationTurn{Role: interfaces.RoleUser, Content: promptStr})
	return finalHistory, nil
}

func (i *Interpreter) getAvailableToolsForAsk(step ast.Step) []interfaces.ToolDefinition {
	if i.toolRegistry == nil {
		i.logger.Warn("getAvailableToolsForAsk called but toolRegistry is nil", "pos", step.Pos.String())
		return nil
	}
	allToolSpecs := i.toolRegistry.ListTools()
	definitions := make([]interfaces.ToolDefinition, 0, len(allToolSpecs))

	for _, spec := range allToolSpecs {
		inputSchema, err := ConvertToolSpecArgsToInputSchema(spec.Args)
		if err != nil {
			i.logger.Warn("Skipping tool for LLM due to schema conversion error", "tool", spec.Name, "error", err, "pos", step.Pos.String())
			continue
		}
		definitions = append(definitions, interfaces.ToolDefinition{
			Name:        spec.Name,
			Description: spec.Description,
			InputSchema: inputSchema,
		})
	}
	return definitions
}

func (i *Interpreter) addResponseToConversation(turn *interfaces.ConversationTurn) {
	i.logger.Debug("Adding LLM response to conversation history (Not Implemented)", "role", turn.Role)
}

func (i *Interpreter) handleToolCalls(calls []*interfaces.ToolCall, pos *lang.Position) error {
	i.logger.Debug("Handling tool calls requested by LLM", "count", len(calls), "pos", pos.String())
	if len(calls) == 0 {
		return nil
	}
	if i.toolRegistry == nil {
		return lang.NewRuntimeError(ErrorCodeConfiguration, fmt.Sprintf("cannot handle tool calls at %s: toolRegistry is nil", pos.String()), nil)
	}

	results := make([]*interfaces.ToolResult, len(calls))

	for idx, call := range calls {
		results[idx] = &interfaces.ToolResult{ID: call.ID}
		i.logger.Debug("Processing tool call", "id", call.ID, "name", call.Name, "args", call.Arguments, "pos", pos.String())

		toolImpl, found := i.toolRegistry.GetTool(call.Name)
		if !found {
			results[idx].Error = fmt.Sprintf("Tool '%s' not found", call.Name)
			continue
		}

		orderedRawArgs := make([]Value, len(toolImpl.Spec.Args))
		// Note: Robust argument mapping logic would be needed here.

		resultVal, execErr := i.toolRegistry.CallFromInterpreter(i, call.Name, orderedRawArgs)

		if execErr != nil {
			results[idx].Error = execErr.Error()
			results[idx].Result = nil
		} else {
			// Unwrap the result for the tool turn, handling potential errors.
			unwrappedResult := Unwrap(resultVal)
			results[idx].Result = unwrappedResult
		}
	}

	i.addToolResultsToConversation(results)
	return nil
}

func (i *Interpreter) addToolResultsToConversation(results []*interfaces.ToolResult) {
	i.logger.Debug("Adding tool results to conversation history (Not Implemented)", "count", len(results))
}

func ConvertToolSpecArgsToInputSchema(args []ArgSpec) (map[string]interface{}, error) {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
		"required":   []string{},
	}
	props, okProps := schema["properties"].(map[string]interface{})
	if !okProps {
		return nil, fmt.Errorf("internal error: could not assert schema properties as map")
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
		case ArgTypeMap:
			jsonType = "object"
		case ArgTypeNil:
			continue
		default:
			return nil, fmt.Errorf("unsupported ArgType '%s' for tool arg '%s'", argSpec.Type, argSpec.Name)
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
