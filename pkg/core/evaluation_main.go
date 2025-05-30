// NeuroScript Version: 0.3.5
// File version: 0.1.0 // Added TypeOfNode and NilLiteralNode handling
// Purpose: Main expression evaluation logic for the NeuroScript interpreter.
// filename: pkg/core/evaluation_main.go
// nlines: 280 // Approximate, adjust after pasting
// risk_rating: HIGH

package core

import (
	"errors"
	"fmt"
	// "math" // Not directly used in this snippet, keep if other parts of file need it
	// "reflect" // Not directly used in this snippet, keep if other parts of file need it
)

// evaluateExpression evaluates an AST node representing an expression.
// Returns the evaluated RAW Go value.
func (i *Interpreter) evaluateExpression(node interface{}) (interface{}, error) {
	if node == nil {
		return nil, NewRuntimeError(ErrorCodeInternal, "evaluateExpression received nil node", nil)
	}

	switch n := node.(type) {
	case *StringLiteralNode:
		if n.IsRaw {
			resolvedStr, resolveErr := i.resolvePlaceholdersWithError(n.Value)
			if resolveErr != nil {
				maxLength := 20
				if len(n.Value) < maxLength {
					maxLength = len(n.Value)
				}
				return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("evaluating raw string literal '%s...'", n.Value[:maxLength]), resolveErr)
			}
			return resolvedStr, nil
		}
		return n.Value, nil
	case *NumberLiteralNode:
		return n.Value, nil
	case *BooleanLiteralNode:
		return n.Value, nil
	case *NilLiteralNode: // Added case for NilLiteralNode
		return nil, nil
	case *VariableNode:
		val, exists := i.variables[n.Name]
		if !exists {
			// Check if it's a predefined constant like a type name
			// This is a preliminary step for making type constants available.
			// This specific location might need refinement based on how globals/built-ins are structured.
			if typeVal, typeExists := i.GetTypeConstant(n.Name); typeExists {
				return typeVal, nil
			}
			return nil, NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("variable '%s' not found", n.Name), ErrVariableNotFound)
		}
		return val, nil
	case *PlaceholderNode:
		var refValue interface{}
		var varName string
		var exists bool
		if n.Name == "LAST" {
			refValue = i.lastCallResult
			varName = "LAST"
			exists = true
		} else {
			refValue, exists = i.variables[n.Name]
			varName = n.Name
		}
		if !exists {
			return nil, NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("variable '%s' not found", n.Name), ErrVariableNotFound)
		}
		_ = varName
		return refValue, nil
	case *LastNode:
		return i.lastCallResult, nil
	case *ListLiteralNode:
		evaluatedElements := make([]interface{}, len(n.Elements))
		var err error
		for idx, elemNode := range n.Elements {
			evaluatedElements[idx], err = i.evaluateExpression(elemNode)
			if err != nil {
				return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("evaluating list literal element %d", idx), err)
			}
		}
		return evaluatedElements, nil
	case *MapLiteralNode:
		evaluatedMap := make(map[string]interface{})
		var err error
		for _, entry := range n.Entries {
			mapKey := entry.Key.Value
			evaluatedMap[mapKey], err = i.evaluateExpression(entry.Value)
			if err != nil {
				return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("evaluating value for map key %q", mapKey), err)
			}
		}
		return evaluatedMap, nil
	case *EvalNode:
		argValueRaw, err := i.evaluateExpression(n.Argument)
		if err != nil {
			return nil, NewRuntimeError(ErrorCodeEvaluation, "evaluating argument for EVAL", err)
		}
		argStr := ""
		if argValueRaw != nil {
			argStr = fmt.Sprintf("%v", argValueRaw)
		}
		resolvedStr, resolveErr := i.resolvePlaceholdersWithError(argStr)
		if resolveErr != nil {
			return nil, NewRuntimeError(ErrorCodeEvaluation, "resolving placeholders during EVAL", resolveErr)
		}
		return resolvedStr, nil
	case *UnaryOpNode:
		operandVal, err := i.evaluateExpression(n.Operand)
		if err != nil {
			return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("evaluating operand for unary operator '%s'", n.Operator), err)
		}
		return evaluateUnaryOp(n.Operator, operandVal)
	case *BinaryOpNode:
		i.Logger().Debug("[DEBUG-EVAL-BINOP] Evaluating BinaryOpNode", "operator", n.Operator)
		leftVal, errL := i.evaluateExpression(n.Left)
		if errL != nil {
			if (n.Operator == "==" || n.Operator == "!=") && errors.Is(errL, ErrVariableNotFound) {
				leftNodeStr := NodeToString(n.Left)
				i.Logger().Debug("[DEBUG-EVAL-BINOP] Left operand not found, treating as nil for comparison", "operand_str", leftNodeStr)
				leftVal = nil
			} else {
				i.Logger().Error("[DEBUG-EVAL-BINOP] Error evaluating left operand", "operator", n.Operator, "error", errL)
				return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("evaluating left operand for '%s'", n.Operator), errL)
			}
		}
		i.Logger().Debug("[DEBUG-EVAL-BINOP] Left operand evaluated", "value", leftVal, "type", fmt.Sprintf("%T", leftVal))

		if n.Operator == "and" {
			if !isTruthy(leftVal) {
				i.Logger().Debug("[DEBUG-EVAL-BINOP] Short-circuiting 'and' (left is falsey)")
				return false, nil
			}
		} else if n.Operator == "or" {
			if isTruthy(leftVal) {
				i.Logger().Debug("[DEBUG-EVAL-BINOP] Short-circuiting 'or' (left is truthy)")
				return true, nil
			}
		}

		rightVal, errR := i.evaluateExpression(n.Right)
		if errR != nil {
			if (n.Operator == "==" || n.Operator == "!=") && errors.Is(errR, ErrVariableNotFound) {
				rightNodeStr := NodeToString(n.Right)
				i.Logger().Debug("[DEBUG-EVAL-BINOP] Right operand not found, treating as nil for comparison", "operand_str", rightNodeStr)
				rightVal = nil
			} else {
				i.Logger().Error("[DEBUG-EVAL-BINOP] Error evaluating right operand", "operator", n.Operator, "error", errR)
				return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("evaluating right operand for '%s'", n.Operator), errR)
			}
		}
		i.Logger().Debug("[DEBUG-EVAL-BINOP] Right operand evaluated", "value", rightVal, "type", fmt.Sprintf("%T", rightVal))
		i.Logger().Debug("[DEBUG-EVAL-BINOP] Calling evaluateBinaryOp", "left_value", leftVal, "right_value", rightVal, "operator", n.Operator)
		result, err := evaluateBinaryOp(leftVal, rightVal, n.Operator)
		if err != nil {
			if _, ok := err.(*RuntimeError); !ok {
				err = NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("operation '%s' failed", n.Operator), err)
			}
			i.Logger().Error("[DEBUG-EVAL-BINOP] Error from evaluateBinaryOp", "operator", n.Operator, "error", err)
			return nil, err
		}
		i.Logger().Debug("[DEBUG-EVAL-BINOP] evaluateBinaryOp successful", "result", result, "type", fmt.Sprintf("%T", result))
		return result, nil

	case *TypeOfNode: // Added case for TypeOfNode
		i.Logger().Debug("[DEBUG-EVAL] Evaluating TypeOfNode")
		if n.Argument == nil {
			i.Logger().Error("[DEBUG-EVAL] TypeOfNode has nil Argument")
			return nil, NewRuntimeError(ErrorCodeInternal, "TypeOfNode has nil Argument", nil)
		}

		argValue, err := i.evaluateExpression(n.Argument)
		if err != nil {
			// If the argument evaluation failed (e.g. variable not found),
			// typeof should operate on the effective value, which would be nil in such cases.
			// We only propagate the error if it's not ErrVariableNotFound.
			if !errors.Is(err, ErrVariableNotFound) {
				i.Logger().Error("[DEBUG-EVAL] Error evaluating argument for TypeOfNode", "error", err)
				return nil, NewRuntimeError(ErrorCodeEvaluation, "evaluating argument for typeof", err)
			}
			argValue = nil // Treat as nil if variable not found for typeof's purpose
		}

		// Determine type and return the string name using constants from type_names.go
		switch argValue.(type) {
		case string:
			return string(TypeString), nil
		case int, int8, int16, int32, int64, float32, float64:
			return string(TypeNumber), nil
		case bool:
			return string(TypeBoolean), nil
		case []interface{}:
			return string(TypeList), nil
		case map[string]interface{}:
			return string(TypeMap), nil
		case nil:
			return string(TypeNil), nil
		// TODO: Add cases for *ProcedureDefinition and tool types if they are distinct types.
		// For example:
		// case *ProcedureDefinition:
		//     return string(TypeFunction), nil
		// Check if argValue could be a Tool structure or similar:
		// case Tool: // Assuming 'Tool' is the struct type for registered tools
		//     return string(TypeTool), nil
		default:
			i.Logger().Warn("[DEBUG-EVAL] TypeOfNode encountered an unhandled Go type for evaluated argument", "type", fmt.Sprintf("%T", argValue))
			return string(TypeUnknown), nil
		}

	case *CallableExprNode:
		target := n.Target
		targetName := target.Name
		evaluatedArgs := make([]interface{}, len(n.Arguments))
		var argErr error
		for idx, argNode := range n.Arguments {
			evaluatedArgs[idx], argErr = i.evaluateExpression(argNode)
			if argErr != nil {
				callDesc := targetName
				if target.IsTool {
					callDesc = "tool." + targetName
				}
				return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("evaluating arg %d for call to '%s'", idx+1, callDesc), argErr)
			}
		}

		if target.IsTool {
			i.Logger().Debug("[DEBUG-EVAL] Calling Tool from expression", "tool_name", targetName)
			toolImpl, found := i.ToolRegistry().GetTool(targetName)
			if !found {
				errMsg := fmt.Sprintf("tool '%s' not found", targetName)
				return nil, NewRuntimeError(ErrorCodeToolNotFound, errMsg, ErrToolNotFound)
			}
			validatedArgs, validationErr := ValidateAndConvertArgs(toolImpl.Spec, evaluatedArgs)
			if validationErr != nil {
				return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("args failed for tool '%s'", targetName), validationErr)
			}
			toolResult, toolErr := toolImpl.Func(i, validatedArgs) // Pass interpreter 'i'
			if toolErr != nil {
				if re, ok := toolErr.(*RuntimeError); ok {
					return nil, re
				}
				return nil, NewRuntimeError(ErrorCodeToolExecutionFailed, fmt.Sprintf("tool '%s' execution failed", targetName), toolErr)
			}
			i.Logger().Debug("[DEBUG-EVAL] Tool call successful", "tool_name", targetName, "result_type", fmt.Sprintf("%T", toolResult))
			i.lastCallResult = toolResult
			return toolResult, nil
		} else {
			i.Logger().Debug("[DEBUG-EVAL] Calling User Proc or Built-in from expression", "function_name", targetName)
			result, err := i.evaluateUserOrBuiltInFunction(targetName, evaluatedArgs)
			if err != nil {
				return nil, err
			}
			return result, nil
		}
	case *ElementAccessNode:
		return i.evaluateElementAccess(n)
	default:
		// This default case handles values that might already be evaluated (e.g. primitives passed around)
		// or truly unhandled AST node types.
		switch node.(type) {
		case string, int64, float64, bool, nil, []interface{}, map[string]interface{}:
			return node, nil
		}
		i.Logger().Error("[DEBUG-EVAL] Unhandled node type in evaluateExpression", "type", fmt.Sprintf("%T", node))
		return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("evaluateExpression unhandled node type: %T", node), nil)
	}
}

// evaluateUserOrBuiltInFunction handles dispatching to built-ins or user procedures.
func (i *Interpreter) evaluateUserOrBuiltInFunction(funcName string, args []interface{}) (interface{}, error) {
	if isBuiltInFunction(funcName) {
		result, err := evaluateBuiltInFunction(funcName, args)
		if err != nil {
			if _, ok := err.(*RuntimeError); !ok {
				err = NewRuntimeError(ErrorCodeGeneric, fmt.Sprintf("built-in function '%s' failed", funcName), err)
			}
			return nil, err
		}
		return result, nil
	}

	procResult, procErr := i.RunProcedure(funcName, args...)
	if procErr != nil {
		if _, ok := procErr.(*RuntimeError); !ok {
			code := ErrorCodeGeneric
			wrappedErr := procErr
			errMsg := procErr.Error()
			if errors.Is(procErr, ErrProcedureNotFound) {
				code = ErrorCodeProcNotFound
				wrappedErr = ErrProcedureNotFound
				errMsg = fmt.Sprintf("procedure '%s' not found", funcName)
			} else if errors.Is(procErr, ErrArgumentMismatch) {
				code = ErrorCodeArgMismatch
				wrappedErr = ErrArgumentMismatch
				errMsg = fmt.Sprintf("argument mismatch calling procedure '%s'", funcName)
			}
			procErr = NewRuntimeError(code, errMsg, wrappedErr)
		}
		return nil, procErr
	}
	i.lastCallResult = procResult
	return procResult, nil
}

// GetTypeConstant checks if a name matches a predefined type constant.
// This is a helper for making type constants accessible.
func (i *Interpreter) GetTypeConstant(name string) (string, bool) {
	// This assumes type constants are exposed as global variables with specific names.
	// If they are namespaced like 'types.STRING', this logic would need to change.
	switch name {
	case "TYPE_STRING":
		return string(TypeString), true
	case "TYPE_NUMBER":
		return string(TypeNumber), true
	case "TYPE_BOOLEAN":
		return string(TypeBoolean), true
	case "TYPE_LIST":
		return string(TypeList), true
	case "TYPE_MAP":
		return string(TypeMap), true
	case "TYPE_NIL":
		return string(TypeNil), true
	case "TYPE_FUNCTION":
		return string(TypeFunction), true
	case "TYPE_TOOL":
		return string(TypeTool), true
	case "TYPE_UNKNOWN":
		return string(TypeUnknown), true
	}
	return "", false
}
