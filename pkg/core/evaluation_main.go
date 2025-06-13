// NeuroScript Version: 0.3.5
// File version: 8
// Purpose: Corrected evaluation calls to reflect recent refactoring of helpers/operators.
// filename: pkg/core/evaluation_main.go
// nlines: 275
// risk_rating: LOW

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
		GetPos() *Position
	}); ok {
		posResult := posProvider.GetPos()
		if posResult != nil {
			currentPosStr = posResult.String()
		}
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
				return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("evaluating raw string literal '%s...' at %s", n.Value[:maxLength], currentPosStr), resolveErr)
			}
			return StringValue{Value: resolvedStr}, nil
		}
		return StringValue{Value: n.Value}, nil
	case *NumberLiteralNode:
		// Normalize all incoming numeric types from the parser to float64
		// for our internal NumberValue representation.
		switch v := n.Value.(type) {
		case float64:
			return NumberValue{Value: v}, nil
		case int:
			return NumberValue{Value: float64(v)}, nil
		case int64:
			return NumberValue{Value: float64(v)}, nil
		default:
			return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("unhandled number literal type %T", n.Value), nil)
		}
	case *BooleanLiteralNode:
		return BoolValue{Value: n.Value}, nil
	case *NilLiteralNode:
		return NilValue{}, nil
	case *VariableNode:
		val, exists := i.variables[n.Name]
		if !exists {
			if typeVal, typeExists := i.GetTypeConstant(n.Name); typeExists {
				return StringValue{Value: typeVal}, nil // Constants should also be wrapped
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
		evaluatedElements := make([]Value, len(n.Elements))
		for idx, elemNode := range n.Elements {
			elemVal, err := i.evaluateExpression(elemNode)
			if err != nil {
				return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("evaluating list literal element %d at %s", idx, currentPosStr), err)
			}
			valAsValue, ok := elemVal.(Value)
			if !ok {
				// This block indicates an issue where evaluation returns a raw type.
				// For now, we attempt to wrap it, but ideally all evaluation paths should return a Value type.
				if elemVal == nil {
					valAsValue = NilValue{}
				} else {
					// This should become increasingly rare as we refactor.
					// For now, it's a critical signal of a logic error.
					return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("list element expression evaluated to a non-Value type: %T", elemVal), nil)
				}
			}
			evaluatedElements[idx] = valAsValue
		}
		return NewListValue(evaluatedElements), nil
	case *MapLiteralNode:
		evaluatedMap := make(map[string]Value)
		for _, entry := range n.Entries {
			mapKey := entry.Key.Value // Assuming Key is a StringLiteralNode
			elemVal, err := i.evaluateExpression(entry.Value)
			if err != nil {
				return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("evaluating value for map key %q at %s", mapKey, currentPosStr), err)
			}
			valAsValue, ok := elemVal.(Value)
			if !ok {
				if elemVal == nil {
					valAsValue = NilValue{}
				} else {
					return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("map value expression for key '%s' evaluated to a non-Value type: %T", mapKey, elemVal), nil)
				}
			}
			evaluatedMap[mapKey] = valAsValue
		}
		return NewMapValue(evaluatedMap), nil
	case *EvalNode:
		argValueRaw, err := i.evaluateExpression(n.Argument)
		if err != nil {
			return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("evaluating argument for EVAL at %s", currentPosStr), err)
		}
		argStr, _ := toString(argValueRaw)
		resolvedStr, resolveErr := i.resolvePlaceholdersWithError(argStr)
		if resolveErr != nil {
			return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("resolving placeholders during EVAL at %s", currentPosStr), resolveErr)
		}
		return StringValue{Value: resolvedStr}, nil
	case *UnaryOpNode:
		operandVal, err := i.evaluateExpression(n.Operand)
		if err != nil {
			return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("evaluating operand for unary operator '%s' at %s", n.Operator, currentPosStr), err)
		}
		// FIX: Call as method on interpreter instance 'i'
		return i.evaluateUnaryOp(n.Operator, operandVal)

	case *BinaryOpNode:
		i.Logger().Debug("[DEBUG-EVAL-BINOP] Evaluating BinaryOpNode", "operator", n.Operator, "pos", currentPosStr)

		// Short-circuiting logic for AND/OR is now handled within evaluateBinaryOp.
		// We still need to evaluate left and potentially right operands.
		leftVal, errL := i.evaluateExpression(n.Left)
		if errL != nil {
			// For equality checks, a missing variable should be treated as nil, not an error.
			if (n.Operator == "==" || n.Operator == "!=") && errors.Is(errL, ErrVariableNotFound) {
				leftVal = NilValue{}
			} else {
				return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("evaluating left operand for '%s' at %s", n.Operator, currentPosStr), errL)
			}
		}

		// The 'evaluateBinaryOp' function now contains the short-circuit logic.
		// It will decide whether it even needs to evaluate the right-hand side.
		// To do this, it needs the unevaluated right node, not the value.
		// However, for simplicity and to avoid redesigning the call chain,
		// we'll evaluate right here and let the boolean logic inside evaluateBinaryOp handle it.
		// This is slightly inefficient for short-circuited cases but correct.
		rightVal, errR := i.evaluateExpression(n.Right)
		if errR != nil {
			if (n.Operator == "==" || n.Operator == "!=") && errors.Is(errR, ErrVariableNotFound) {
				rightVal = NilValue{}
			} else {
				return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("evaluating right operand for '%s' at %s", n.Operator, currentPosStr), errR)
			}
		}

		// FIX: Call as method on interpreter instance 'i'
		return i.evaluateBinaryOp(leftVal, rightVal, n.Operator)

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
			argValue = NilValue{}
		}
		// FIX: Call as standalone helper function, not method
		return StringValue{Value: string(TypeOf(argValue))}, nil

	case *CallableExprNode:
		target := n.Target
		targetName := target.Name
		callablePosStr := currentPosStr
		evaluatedArgs := make([]interface{}, len(n.Arguments))
		var argErr error
		for idx, argNode := range n.Arguments {
			argPosStr := "[unknown_arg_pos]"
			if pArgNode, ok := argNode.(interface {
				GetPos() *Position
			}); ok {
				posResult := pArgNode.GetPos()
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
			ToolResult, toolErr := toolImpl.Func(i, validatedArgs)
			if toolErr != nil {
				if re, ok := toolErr.(*RuntimeError); ok {
					return nil, re
				}
				return nil, NewRuntimeError(ErrorCodeToolExecutionFailed, fmt.Sprintf("tool '%s' execution failed at %s: %v", targetName, callablePosStr, toolErr), toolErr)
			}
			i.lastCallResult = ToolResult
			return ToolResult, nil
		} else {
			result, err := i.evaluateUserOrBuiltInFunction(targetName, evaluatedArgs)
			return result, err
		}

	case *ElementAccessNode:
		return i.evaluateElementAccess(n)
	default:
		// This case handles values that are already evaluated (or should be).
		if _, ok := node.(Value); ok {
			return node, nil
		}
		// Fallback for raw types that might still be flowing through during refactoring.
		if _, ok := node.(string); ok {
			return node, nil
		}
		if _, ok := node.(int64); ok {
			return node, nil
		}
		if _, ok := node.(float64); ok {
			return node, nil
		}
		if _, ok := node.(bool); ok {
			return node, nil
		}
		if node == nil {
			return nil, nil
		}

		i.Logger().Error("[DEBUG-EVAL] Unhandled node type in evaluateExpression", "type", fmt.Sprintf("%T", node), "pos", currentPosStr)
		return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("evaluateExpression unhandled node type: %T at %s", node, currentPosStr), nil)
	}
}

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
	case "TYPE_ERROR":
		return string(TypeError), true
	case "TYPE_EVENT":
		return string(TypeEvent), true
	case "TYPE_TIMEDATE":
		return string(TypeTimedate), true
	case "TYPE_FUZZY":
		return string(TypeFuzzy), true
	case "TYPE_UNKNOWN":
		return string(TypeUnknown), true
	}
	return "", false
}
