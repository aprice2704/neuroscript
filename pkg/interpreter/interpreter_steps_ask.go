// NeuroScript Version: 0.5.2
// File version: 11.0.0
// Purpose: Corrected the argument preparation logic in handleToolCalls to properly build the positional argument slice, fixing the final test failure.
// filename: pkg/interpreter/interpreter_steps_ask.go
// nlines: 260
// risk_rating: HIGH
package interpreter

import (
	"context"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// Ask satisfies the tool.Runtime interface. It's a simplified version for tools.
func (i *Interpreter) Ask(prompt string) string {
	if i.aiWorker == nil {
		i.Logger().Error("Ask called but LLM client (aiWorker) is not configured", "prompt", prompt)
		return ""
	}
	turns := []*interfaces.ConversationTurn{{Role: interfaces.RoleUser, Content: prompt}}

	responseTurn, err := i.aiWorker.Ask(context.Background(), turns)
	if err != nil {
		i.Logger().Error("LLM interaction failed in Ask", "error", err, "prompt", prompt)
		return ""
	}
	if responseTurn == nil {
		i.Logger().Error("LLM returned nil response without error in Ask", "prompt", prompt)
		return ""
	}
	return responseTurn.Content
}

// askWithTools is the internal, more powerful version of Ask used by the interpreter.
func (i *Interpreter) askWithTools(ctx context.Context, turns []*interfaces.ConversationTurn, tools []interfaces.ToolDefinition) (*interfaces.ConversationTurn, []*interfaces.ToolCall, error) {
	if i.aiWorker == nil {
		return nil, nil, lang.NewRuntimeError(lang.ErrorCodeLLMError, "LLM client (aiWorker) is not configured in the interpreter", nil)
	}

	var responseTurn *interfaces.ConversationTurn
	var toolCalls []*interfaces.ToolCall
	var err error

	if len(tools) > 0 {
		i.logger.Debug("Calling LLM with tools", "turn_count", len(turns), "tool_count", len(tools))
		responseTurn, toolCalls, err = i.aiWorker.AskWithTools(ctx, turns, tools)
	} else {
		i.logger.Debug("Calling LLM without tools", "turn_count", len(turns))
		responseTurn, err = i.aiWorker.Ask(ctx, turns)
	}

	if err != nil {
		return nil, nil, lang.NewRuntimeError(lang.ErrorCodeLLMError, "LLM interaction failed", err)
	}
	if responseTurn == nil {
		return nil, nil, lang.NewRuntimeError(lang.ErrorCodeLLMError, "LLM returned nil response without error", nil)
	}

	return responseTurn, toolCalls, nil
}

func (i *Interpreter) executeAsk(step ast.Step) (lang.Value, error) {
	if len(step.Values) == 0 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "ask step has no value expression", nil).WithPosition(step.GetPos())
	}
	promptExpr := step.Values[0]

	promptVal, err := i.evaluate.Expression(promptExpr)
	if err != nil {
		return nil, err
	}
	promptStr, _ := lang.ToString(promptVal)

	if i.aiWorker == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeLLMError, "LLM client (aiWorker) is not configured in the interpreter", lang.ErrLLMNotConfigured).WithPosition(step.GetPos())
	}

	turns := []*interfaces.ConversationTurn{{Role: interfaces.RoleUser, Content: promptStr}}
	responseTurn, err := i.aiWorker.Ask(context.Background(), turns)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeLLMError, "LLM interaction failed", err).WithPosition(step.GetPos())
	}

	responseVal := lang.StringValue{Value: responseTurn.Content}

	if step.AskIntoVar != "" {
		if err := i.SetVariable(step.AskIntoVar, responseVal); err != nil {
			return nil, err
		}
	}
	return responseVal, nil
}

// executeAskAI handles the 'ask ai' step.
func (i *Interpreter) executeAskAI(step ast.Step) error {
	i.logger.Debug("Executing 'ask ai' step", "pos", step.GetPos().String())

	if len(step.Values) == 0 {
		return lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("Internal error at %s: 'ask' step has no value expression", step.GetPos().String()), nil)
	}
	promptExpr, ok := step.Values[0].(ast.Expression)
	if !ok {
		return lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("Internal error at %s: 'ask' step value is not an ast.Expression (got %T)", step.GetPos().String(), step.Values[0]), nil)
	}

	conversationTurns, err := i.prepareConversationForAsk(promptExpr)
	if err != nil {
		errMsg := fmt.Sprintf("failed to prepare conversation for 'ask ai' at %s: %v", step.GetPos().String(), err)
		return lang.NewRuntimeError(lang.ErrorCodeEvaluation, errMsg, err)
	}

	availableTools := i.getAvailableToolsForAsk(step)
	responseTurn, toolCalls, err := i.askWithTools(context.Background(), conversationTurns, availableTools)
	if err != nil {
		return err
	}

	i.logger.Debug("LLM response received", "role", responseTurn.Role, "content_length", len(responseTurn.Content), "tool_calls", len(toolCalls))

	i.addResponseToConversation(responseTurn)

	if len(toolCalls) > 0 {
		err = i.handleToolCalls(toolCalls, step.GetPos())
		if err != nil {
			return fmt.Errorf("failed during tool execution (initiated at %s): %w", step.GetPos().String(), err)
		}
		i.logger.Debug("Tool calls requested by LLM were processed.")
	}

	i.lastCallResult = lang.StringValue{Value: responseTurn.Content}
	return nil
}

func (i *Interpreter) prepareConversationForAsk(promptExpr ast.Expression) ([]*interfaces.ConversationTurn, error) {
	promptVal, err := i.evaluate.Expression(promptExpr)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate prompt expression: %w", err)
	}

	strVal, ok := promptVal.(lang.StringValue)
	if !ok {
		errMsg := fmt.Sprintf("prompt expression at %s did not evaluate to a string, got %s", promptExpr.GetPos().String(), promptVal.Type())
		return nil, lang.NewRuntimeError(lang.ErrorCodeEvaluation, errMsg, nil)
	}
	promptStr := strVal.Value

	currentHistory := []*interfaces.ConversationTurn{}
	finalHistory := append(currentHistory, &interfaces.ConversationTurn{Role: interfaces.RoleUser, Content: promptStr})
	return finalHistory, nil
}

func (i *Interpreter) getAvailableToolsForAsk(step ast.Step) []interfaces.ToolDefinition {
	allToolSpecs := i.ToolRegistry().ListTools()
	definitions := make([]interfaces.ToolDefinition, 0, len(allToolSpecs))

	for _, spec := range allToolSpecs {
		inputSchema, err := ConvertToolSpecArgsToInputSchema(spec.Args)
		if err != nil {
			i.logger.Warn("Skipping tool for LLM due to schema conversion error", "tool", spec.Name, "error", err, "pos", step.GetPos().String())
			continue
		}
		fullname := types.MakeFullName(string(spec.Group), string(spec.Name))
		definitions = append(definitions, interfaces.ToolDefinition{
			Name:        fullname,
			Description: spec.Description,
			InputSchema: inputSchema,
		})
	}
	return definitions
}

func (i *Interpreter) addResponseToConversation(turn *interfaces.ConversationTurn) {
	i.logger.Debug("Adding LLM response to conversation history (Not Implemented)", "role", turn.Role)
}

func (i *Interpreter) handleToolCalls(calls []*interfaces.ToolCall, pos *types.Position) error {
	i.logger.Debug("Handling tool calls requested by LLM", "count", len(calls), "pos", pos.String())
	if len(calls) == 0 {
		return nil
	}

	results := make([]*interfaces.ToolResult, len(calls))

	for idx, call := range calls {
		results[idx] = &interfaces.ToolResult{ID: call.ID}
		i.logger.Debug("Processing tool call", "id", call.ID, "name", call.Name, "args", call.Arguments, "pos", pos.String())

		tool, found := i.ToolRegistry().GetTool(call.Name)
		if !found {
			results[idx].Error = fmt.Sprintf("Tool '%s' not found", call.Name)
			continue
		}

		// FIX: Correctly build the positional argument slice from the named map.
		positionalArgs := make([]interface{}, len(tool.Spec.Args))
		for i, argSpec := range tool.Spec.Args {
			if val, ok := call.Arguments[argSpec.Name]; ok {
				positionalArgs[i] = val
			} else {
				positionalArgs[i] = nil // Or handle default/required logic
			}
		}

		// Directly invoke the tool's function with the prepared arguments.
		result, execErr := tool.Func(i, positionalArgs)
		if execErr != nil {
			results[idx].Error = execErr.Error()
		} else {
			results[idx].Result = result
		}
	}

	i.addToolResultsToConversation(results)
	return nil
}

func (i *Interpreter) addToolResultsToConversation(results []*interfaces.ToolResult) {
	i.logger.Debug("Adding tool results to conversation history (Not Implemented)", "count", len(results))
}

func ConvertToolSpecArgsToInputSchema(args []tool.ArgSpec) (map[string]interface{}, error) {
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
		case "string":
			jsonType = "string"
		case "int":
			jsonType = "integer"
		case "float":
			jsonType = "number"
		case "bool":
			jsonType = "boolean"
		case "slice", "slice_any", "slice_string", "slice_int", "slice_float", "slice_bool", "slice_map":
			jsonType = "array"
		case "map":
			jsonType = "object"
		case "nil":
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
