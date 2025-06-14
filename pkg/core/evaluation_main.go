// NeuroScript Version: 0.4.0
// File version: 18
// Purpose: Uses the bridge for tool calls and removes the redundant internal Wrap method.
// filename: pkg/core/evaluation_main.go
// nlines: 230
// risk_rating: HIGH

package core

import (
	"errors"
	"fmt"
)

// evaluateExpression evaluates an AST node representing an expression, ensuring the return is always a `Value` type.
func (i *Interpreter) evaluateExpression(node interface{}) (Value, error) {
	if node == nil {
		return NilValue{}, nil
	}

	var resultValue Value
	var resultErr error

	switch n := node.(type) {
	case *StringLiteralNode:
		if n.IsRaw {
			resolvedStr, resolveErr := i.resolvePlaceholdersWithError(n.Value)
			if resolveErr != nil {
				return nil, NewRuntimeError(ErrorCodeEvaluation, "evaluating raw string literal", resolveErr)
			}
			resultValue = StringValue{Value: resolvedStr}
		} else {
			resultValue = StringValue{Value: n.Value}
		}
	case *NumberLiteralNode:
		val, _ := toFloat64(n.Value)
		resultValue = NumberValue{Value: val}
	case *BooleanLiteralNode:
		resultValue = BoolValue{Value: n.Value}
	case *NilLiteralNode:
		resultValue = NilValue{}
	case *VariableNode:
		if val, exists := i.variables[n.Name]; exists {
			resultValue = val.(Value)
		} else if proc, procExists := i.knownProcedures[n.Name]; procExists {
			resultValue = FunctionValue{Value: *proc}
		} else if tool, toolExists := i.ToolRegistry().GetTool(n.Name); toolExists {
			resultValue = ToolValue{Value: tool}
		} else if typeVal, typeExists := i.GetTypeConstant(n.Name); typeExists {
			resultValue = StringValue{Value: typeVal}
		} else {
			resultErr = NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("variable '%s' not found", n.Name), ErrVariableNotFound)
		}
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
			resultErr = NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("variable '%s' not found", n.Name), ErrVariableNotFound)
		} else {
			resultValue = refValue.(Value)
		}
	case *LastNode:
		resultValue = i.lastCallResult.(Value)
	case *ListLiteralNode:
		evaluatedElements := make([]Value, len(n.Elements))
		for idx, elemNode := range n.Elements {
			evaluatedElements[idx], resultErr = i.evaluateExpression(elemNode)
			if resultErr != nil {
				return nil, resultErr
			}
		}
		resultValue = NewListValue(evaluatedElements)
	case *MapLiteralNode:
		evaluatedMap := make(map[string]Value)
		for _, entry := range n.Entries {
			mapKey := entry.Key.Value
			elemVal, err := i.evaluateExpression(entry.Value)
			if err != nil {
				return nil, err
			}
			evaluatedMap[mapKey] = elemVal
		}
		resultValue = NewMapValue(evaluatedMap)
	case *EvalNode:
		argValueRaw, err := i.evaluateExpression(n.Argument)
		if err != nil {
			return nil, err
		}
		argStr, _ := toString(argValueRaw)
		resolvedStr, resolveErr := i.resolvePlaceholdersWithError(argStr)
		if resolveErr != nil {
			return nil, NewRuntimeError(ErrorCodeEvaluation, "resolving placeholders during EVAL", resolveErr)
		}
		resultValue = StringValue{Value: resolvedStr}
	case *UnaryOpNode:
		operandVal, err := i.evaluateExpression(n.Operand)
		if err != nil {
			return nil, err
		}
		resultValue, resultErr = i.evaluateUnaryOp(n.Operator, operandVal)

	case *BinaryOpNode:
		leftVal, errL := i.evaluateExpression(n.Left)
		if errL != nil {
			if (n.Operator == "==" || n.Operator == "!=") && errors.Is(errL, ErrVariableNotFound) {
				leftVal = NilValue{}
			} else {
				return nil, errL
			}
		}
		rightVal, errR := i.evaluateExpression(n.Right)
		if errR != nil {
			if (n.Operator == "==" || n.Operator == "!=") && errors.Is(errR, ErrVariableNotFound) {
				rightVal = NilValue{}
			} else {
				return nil, errR
			}
		}
		resultValue, resultErr = i.evaluateBinaryOp(leftVal, rightVal, n.Operator)

	case *TypeOfNode:
		argValue, err := i.evaluateExpression(n.Argument)
		if err != nil {
			if errors.Is(err, ErrVariableNotFound) {
				argValue = NilValue{}
			} else {
				return nil, err
			}
		}
		resultValue = StringValue{Value: string(TypeOf(argValue))}

	case *CallableExprNode:
		target := n.Target
		targetName := target.Name
		evaluatedArgs := make([]Value, len(n.Arguments))
		for idx, argNode := range n.Arguments {
			evaluatedArgs[idx], resultErr = i.evaluateExpression(argNode)
			if resultErr != nil {
				return nil, resultErr
			}
		}
		if target.IsTool {
			resultValue, resultErr = i.toolRegistry.CallFromInterpreter(i, targetName, evaluatedArgs)
			if resultErr == nil {
				i.lastCallResult = resultValue
			}
		} else {
			resultValue, resultErr = i.evaluateUserOrBuiltInFunction(targetName, evaluatedArgs)
		}

	case *ElementAccessNode:
		resultValue, resultErr = i.evaluateElementAccess(n)
	default:
		// This case handles raw values passed in from tests or other Go code.
		resultValue, resultErr = Wrap(n)
	}

	if resultErr != nil {
		return nil, resultErr
	}

	return resultValue, nil
}

func (i *Interpreter) evaluateUserOrBuiltInFunction(funcName string, args []Value) (Value, error) {
	// Convert []Value to []interface{} for the call signatures
	interfaceArgs := make([]interface{}, len(args))
	for i, v := range args {
		interfaceArgs[i] = v
	}

	if isBuiltInFunction(funcName) {
		result, err := evaluateBuiltInFunction(funcName, interfaceArgs)
		if err != nil {
			if _, ok := err.(*RuntimeError); !ok {
				err = NewRuntimeError(ErrorCodeGeneric, fmt.Sprintf("built-in function '%s' failed", funcName), err)
			}
			return nil, err
		}
		// The result from a built-in function should already be a Value
		return result.(Value), nil
	}

	procResult, procErr := i.RunProcedure(funcName, interfaceArgs...)
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
	return procResult.(Value), nil
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
	case "TYPE_BYTES":
		return string(TypeBytes), true
	}
	return "", false
}
