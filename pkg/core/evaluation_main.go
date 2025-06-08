// NeuroScript Version: 0.3.5
// File version: 0.1.5
// Purpose: Removed premature short-circuiting for and/or to allow for fuzzy logic evaluation.
// filename: pkg/core/evaluation_main.go
// nlines: 270
// risk_rating: HIGH

package core

import (
	"errors"
	"fmt"
)

// evaluateExpression evaluates an AST node representing an expression.
func (i *Interpreter) evaluateExpression(node interface{}) (interface{}, error) {
	if node == nil {
		return nil, NewRuntimeError(ErrorCodeInternal, "evaluateExpression received nil node", nil)
	}

	var currentPosStr string = "[unknown_pos]"
	if posProvider, ok := node.(interface {
		Position() interface{ String() string }
	}); ok {
		posResult := posProvider.Position()
		if posResult != nil {
			currentPosStr = posResult.String()
		}
	}

	switch n := node.(type) {
	// ... (cases for StringLiteralNode, NumberLiteralNode, etc. are unchanged) ...
	case *StringLiteralNode:
		if n.IsRaw {
			resolvedStr, resolveErr := i.resolvePlaceholdersWithError(n.Value)
			if resolveErr != nil {
				maxLength := 20
				if len(n.Value) < maxLength {
					maxLength = len(n.Value)
				}
				return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("evaluating raw string literal '%s...' at %s", n.Value[:maxLength], currentPosStr), resolveErr)
			}
			return resolvedStr, nil
		}
		return n.Value, nil
	case *NumberLiteralNode:
		return n.Value, nil
	case *BooleanLiteralNode:
		return n.Value, nil
	case *NilLiteralNode:
		return nil, nil
	case *VariableNode:
		val, exists := i.variables[n.Name]
		if !exists {
			if typeVal, typeExists := i.GetTypeConstant(n.Name); typeExists {
				return typeVal, nil
			}
			return nil, NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("variable '%s' not found at %s", n.Name, currentPosStr), ErrVariableNotFound)
		}
		return val, nil
	case *PlaceholderNode:
		var refValue interface{}
		var exists bool
		if n.Name == "LAST" {
			refValue = i.lastCallResult
			exists = true
		} else {
			refValue, exists = i.variables[n.Name]
		}
		if !exists {
			return nil, NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("variable '%s' not found at %s", n.Name, currentPosStr), ErrVariableNotFound)
		}
		return refValue, nil
	case *LastNode:
		return i.lastCallResult, nil
	case *ListLiteralNode:
		evaluatedElements := make([]interface{}, len(n.Elements))
		var err error
		for idx, elemNode := range n.Elements {
			evaluatedElements[idx], err = i.evaluateExpression(elemNode)
			if err != nil {
				return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("evaluating list literal element %d at %s", idx, currentPosStr), err)
			}
		}
		return evaluatedElements, nil
	case *MapLiteralNode:
		evaluatedMap := make(map[string]interface{})
		var err error
		for _, entry := range n.Entries {
			mapKey := entry.Key.Value // Assuming Key is a StringLiteralNode
			evaluatedMap[mapKey], err = i.evaluateExpression(entry.Value)
			if err != nil {
				return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("evaluating value for map key %q at %s", mapKey, currentPosStr), err)
			}
		}
		return evaluatedMap, nil
	case *EvalNode:
		argValueRaw, err := i.evaluateExpression(n.Argument)
		if err != nil {
			return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("evaluating argument for EVAL at %s", currentPosStr), err)
		}
		argStr := ""
		if argValueRaw != nil {
			argStr = fmt.Sprintf("%v", argValueRaw)
		}
		resolvedStr, resolveErr := i.resolvePlaceholdersWithError(argStr)
		if resolveErr != nil {
			return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("resolving placeholders during EVAL at %s", currentPosStr), resolveErr)
		}
		return resolvedStr, nil
	case *UnaryOpNode:
		operandVal, err := i.evaluateExpression(n.Operand)
		if err != nil {
			return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("evaluating operand for unary operator '%s' at %s", n.Operator, currentPosStr), err)
		}
		return evaluateUnaryOp(n.Operator, operandVal)

	case *BinaryOpNode:
		i.Logger().Debug("[DEBUG-EVAL-BINOP] Evaluating BinaryOpNode", "operator", n.Operator, "pos", currentPosStr)
		leftVal, errL := i.evaluateExpression(n.Left)
		if errL != nil {
			if (n.Operator == "==" || n.Operator == "!=") && errors.Is(errL, ErrVariableNotFound) {
				leftVal = nil
			} else {
				return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("evaluating left operand for '%s' at %s", n.Operator, currentPosStr), errL)
			}
		}

		// CORRECTED: The short-circuiting logic is REMOVED from this function.
		// All operator logic, including short-circuiting for both boolean and fuzzy types,
		// is now handled exclusively by evaluateBinaryOp.

		rightVal, errR := i.evaluateExpression(n.Right)
		if errR != nil {
			if (n.Operator == "==" || n.Operator == "!=") && errors.Is(errR, ErrVariableNotFound) {
				rightVal = nil
			} else {
				return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("evaluating right operand for '%s' at %s", n.Operator, currentPosStr), errR)
			}
		}

		return evaluateBinaryOp(leftVal, rightVal, n.Operator)

	case *TypeOfNode:
		i.Logger().Debug("[DEBUG-EVAL] Evaluating TypeOfNode", "pos", currentPosStr)
		if n.Argument == nil {
			return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("TypeOfNode has nil Argument at %s", currentPosStr), nil)
		}
		argValue, err := i.evaluateExpression(n.Argument)
		if err != nil {
			if !errors.Is(err, ErrVariableNotFound) {
				return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("evaluating argument for typeof at %s", currentPosStr), err)
			}
			argValue = nil
		}
		return i.TypeOf(argValue), nil

	case *CallableExprNode:
		// ... (This logic remains unchanged) ...
		target := n.Target
		targetName := target.Name
		callablePosStr := currentPosStr
		evaluatedArgs := make([]interface{}, len(n.Arguments))
		var argErr error
		for idx, argNode := range n.Arguments {
			argPosStr := "[unknown_arg_pos]"
			if pArgNode, ok := argNode.(interface {
				Position() interface{ String() string }
			}); ok {
				posResult := pArgNode.Position()
				if posResult != nil {
					argPosStr = posResult.String()
				}
			}
			evaluatedArgs[idx], argErr = i.evaluateExpression(argNode)
			if argErr != nil {
				return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("evaluating arg %d for call to '%s' at %s (arg at %s)", idx+1, targetName, callablePosStr, argPosStr), argErr)
			}
		}
		if target.IsTool {
			toolImpl, found := i.ToolRegistry().GetTool(targetName)
			if !found {
				return nil, NewRuntimeError(ErrorCodeToolNotFound, fmt.Sprintf("tool '%s' not found at %s", targetName, callablePosStr), ErrToolNotFound)
			}
			validatedArgs, validationErr := ValidateAndConvertArgs(toolImpl.Spec, evaluatedArgs)
			if validationErr != nil {
				return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("argument validation failed for tool '%s' at %s: %v", targetName, callablePosStr, validationErr), validationErr)
			}
			toolResult, toolErr := toolImpl.Func(i, validatedArgs)
			if toolErr != nil {
				if re, ok := toolErr.(*RuntimeError); ok {
					return nil, re
				}
				return nil, NewRuntimeError(ErrorCodeToolExecutionFailed, fmt.Sprintf("tool '%s' execution failed at %s: %v", targetName, callablePosStr, toolErr), toolErr)
			}
			i.lastCallResult = toolResult
			return toolResult, nil
		} else {
			result, err := i.evaluateUserOrBuiltInFunction(targetName, evaluatedArgs)
			return result, err
		}

	case *ElementAccessNode:
		return i.evaluateElementAccess(n)
	default:
		switch node.(type) {
		case string, int64, float64, bool, nil, []interface{}, map[string]interface{}:
			return node, nil
		}
		i.Logger().Error("[DEBUG-EVAL] Unhandled node type in evaluateExpression", "type", fmt.Sprintf("%T", node), "pos", currentPosStr)
		return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("evaluateExpression unhandled node type: %T at %s", node, currentPosStr), nil)
	}
}

// ... (evaluateUserOrBuiltInFunction and GetTypeConstant are unchanged) ...
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

func (i *Interpreter) GetTypeConstant(name string) (string, bool) {
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
