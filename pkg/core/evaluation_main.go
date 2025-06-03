// NeuroScript Version: 0.3.5
// File version: 0.1.4 // Adjusted anonymous interface for Position() to resolve 'undefined: Pos'.
// Purpose: Main expression evaluation logic for the NeuroScript interpreter, with robust 'typeof' and corrected call dispatch.
// filename: pkg/core/evaluation_main.go
// nlines: 290 // Approximate, adjust after pasting
// risk_rating: HIGH

package core

import (
	"errors"
	"fmt"
	// Ensure 'reflect' is imported for TypeOfNode
	// "math"
)

// evaluateExpression evaluates an AST node representing an expression.
// Returns the evaluated RAW Go value.
func (i *Interpreter) evaluateExpression(node interface{}) (interface{}, error) {
	if node == nil {
		return nil, NewRuntimeError(ErrorCodeInternal, "evaluateExpression received nil node", nil)
	}

	var currentPosStr string = "[unknown_pos]"
	// Try to get position information.
	// The anonymous interface now defines that Position() returns any type
	// that has a String() string method. This avoids needing 'Pos' to be defined in this exact context
	// if the compiler has issues resolving it from ast.go for the anonymous interface.
	// Your actual ASTNode types (from ast.go) which implement Position() returning a 'Pos' type
	// (where 'Pos' itself has String()) will satisfy this.
	if posProvider, ok := node.(interface {
		Position() interface{ String() string } // MODIFIED to address "undefined: Pos"
	}); ok {
		posResult := posProvider.Position()
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
		return evaluateUnaryOp(n.Operator, operandVal) // evaluateUnaryOp is in evaluation_logic.go
	case *BinaryOpNode:
		i.Logger().Debug("[DEBUG-EVAL-BINOP] Evaluating BinaryOpNode", "operator", n.Operator, "pos", currentPosStr)
		leftVal, errL := i.evaluateExpression(n.Left)
		if errL != nil {
			if (n.Operator == "==" || n.Operator == "!=") && errors.Is(errL, ErrVariableNotFound) {
				leftNodeStr := NodeToString(n.Left) // Assumes NodeToString helper exists
				i.Logger().Debug("[DEBUG-EVAL-BINOP] Left operand not found, treating as nil for comparison", "operand_str", leftNodeStr)
				leftVal = nil
			} else {
				i.Logger().Error("[DEBUG-EVAL-BINOP] Error evaluating left operand", "operator", n.Operator, "error", errL)
				return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("evaluating left operand for '%s' at %s", n.Operator, currentPosStr), errL)
			}
		}
		i.Logger().Debug("[DEBUG-EVAL-BINOP] Left operand evaluated", "value", leftVal, "type", fmt.Sprintf("%T", leftVal))

		if n.Operator == "and" {
			if !isTruthy(leftVal) { // isTruthy is likely in evaluation_helpers.go or evaluation_logic.go
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
				return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("evaluating right operand for '%s' at %s", n.Operator, currentPosStr), errR)
			}
		}
		i.Logger().Debug("[DEBUG-EVAL-BINOP] Right operand evaluated", "value", rightVal, "type", fmt.Sprintf("%T", rightVal))
		i.Logger().Debug("[DEBUG-EVAL-BINOP] Calling evaluateBinaryOp", "left_value", leftVal, "right_value", rightVal, "operator", n.Operator)
		result, err := evaluateBinaryOp(leftVal, rightVal, n.Operator) // evaluateBinaryOp is in evaluation_logic.go
		if err != nil {
			if _, ok := err.(*RuntimeError); !ok { // Ensure it's a RuntimeError
				err = NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("operation '%s' failed at %s", n.Operator, currentPosStr), err)
			}
			i.Logger().Error("[DEBUG-EVAL-BINOP] Error from evaluateBinaryOp", "operator", n.Operator, "error", err)
			return nil, err
		}
		i.Logger().Debug("[DEBUG-EVAL-BINOP] evaluateBinaryOp successful", "result", result, "type", fmt.Sprintf("%T", result))
		return result, nil

		// In pkg/core/evaluation_main.go
		// Inside func (i *Interpreter) evaluateExpression(node interface{}) (interface{}, error)
		// ...
	case *TypeOfNode:
		i.Logger().Debug("[DEBUG-EVAL] Evaluating TypeOfNode", "pos", currentPosStr)
		if n.Argument == nil {
			i.Logger().Error("[DEBUG-EVAL] TypeOfNode has nil Argument", "pos", currentPosStr)
			return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("TypeOfNode has nil Argument at %s", currentPosStr), nil)
		}

		argValue, err := i.evaluateExpression(n.Argument)
		if err != nil {
			if !errors.Is(err, ErrVariableNotFound) { // Keep this specific error handling for typeof argument
				i.Logger().Error("[DEBUG-EVAL] Error evaluating argument for TypeOfNode", "error", err, "pos", currentPosStr)
				return nil, NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("evaluating argument for typeof at %s", currentPosStr), err)
			}
			argValue = nil // Treat as nil if variable not found, typeof operates on this nil.
		}

		// Call the centralized TypeOf method from evaluation_logic.go
		return i.TypeOf(argValue), nil // <<< THIS IS THE SIMPLIFIED CALL

	case *CallableExprNode:
		target := n.Target // This is an *IdentifierNode
		targetName := target.Name
		callablePosStr := currentPosStr // Position of the callable expression itself

		evaluatedArgs := make([]interface{}, len(n.Arguments))
		var argErr error
		for idx, argNode := range n.Arguments {
			argPosStr := "[unknown_arg_pos]"
			// Use the same modified position fetching logic for argNode
			if pArgNode, ok := argNode.(interface {
				Position() interface{ String() string } // <<< MODIFIED HERE
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

		// Dispatch based on IdentifierNode.IsTool (from your original structure)
		if target.IsTool { // Assumes IdentifierNode (n.Target) has IsTool field
			i.Logger().Debug("[DEBUG-EVAL] Calling Tool from expression", "tool_name", targetName, "pos", callablePosStr)
			toolImpl, found := i.ToolRegistry().GetTool(targetName)
			if !found {
				return nil, NewRuntimeError(ErrorCodeToolNotFound, fmt.Sprintf("tool '%s' not found at %s", targetName, callablePosStr), ErrToolNotFound)
			}
			// Assuming ValidateAndConvertArgs exists (e.g., in tools_validation.go)
			validatedArgs, validationErr := ValidateAndConvertArgs(toolImpl.Spec, evaluatedArgs)
			if validationErr != nil {
				return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("argument validation failed for tool '%s' at %s: %v", targetName, callablePosStr, validationErr), validationErr)
			}
			toolResult, toolErr := toolImpl.Func(i, validatedArgs)
			if toolErr != nil {
				if re, ok := toolErr.(*RuntimeError); ok { // Preserve existing RuntimeError
					return nil, re
				}
				// Wrap other errors as RuntimeError
				return nil, NewRuntimeError(ErrorCodeToolExecutionFailed, fmt.Sprintf("tool '%s' execution failed at %s: %v", targetName, callablePosStr, toolErr), toolErr)
			}
			i.Logger().Debug("[DEBUG-EVAL] Tool call successful", "tool_name", targetName, "result_type", fmt.Sprintf("%T", toolResult))
			i.lastCallResult = toolResult // As per your original logic
			return toolResult, nil
		} else {
			// Not a tool, so it's a user-defined procedure or a built-in function
			i.Logger().Debug("[DEBUG-EVAL] Calling User Proc or Built-in from expression", "function_name", targetName, "pos", callablePosStr)
			// evaluateUserOrBuiltInFunction already handles error wrapping to RuntimeError
			result, err := i.evaluateUserOrBuiltInFunction(targetName, evaluatedArgs)
			// No need to re-wrap 'err' here as evaluateUserOrBuiltInFunction ensures it's a RuntimeError or nil
			return result, err
		}

	case *ElementAccessNode:
		return i.evaluateElementAccess(n) // evaluateElementAccess is likely in evaluation_access.go
	default:
		// This default case handles values that might already be evaluated (e.g. primitives passed around)
		// or truly unhandled AST node types.
		switch node.(type) {
		case string, int64, float64, bool, nil, []interface{}, map[string]interface{}:
			return node, nil
		}
		i.Logger().Error("[DEBUG-EVAL] Unhandled node type in evaluateExpression", "type", fmt.Sprintf("%T", node), "pos", currentPosStr)
		return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("evaluateExpression unhandled node type: %T at %s", node, currentPosStr), nil)
	}
}

// evaluateUserOrBuiltInFunction handles dispatching to built-ins or user procedures.
// (This function should remain as it was in your provided file)
func (i *Interpreter) evaluateUserOrBuiltInFunction(funcName string, args []interface{}) (interface{}, error) {
	if isBuiltInFunction(funcName) { // isBuiltInFunction is in evaluation_logic.go
		result, err := evaluateBuiltInFunction(funcName, args) // evaluateBuiltInFunction is in evaluation_logic.go
		if err != nil {
			if _, ok := err.(*RuntimeError); !ok { // Ensure it's a RuntimeError
				err = NewRuntimeError(ErrorCodeGeneric, fmt.Sprintf("built-in function '%s' failed", funcName), err)
			}
			return nil, err
		}
		return result, nil
	}

	// Call user-defined procedure
	procResult, procErr := i.RunProcedure(funcName, args...)
	if procErr != nil {
		// Ensure error is RuntimeError, adapting from your original logic
		if _, ok := procErr.(*RuntimeError); !ok {
			code := ErrorCodeGeneric // Default code
			wrappedErr := procErr
			errMsg := procErr.Error() // Default message

			if errors.Is(procErr, ErrProcedureNotFound) {
				code = ErrorCodeProcNotFound
				wrappedErr = ErrProcedureNotFound // Preserve sentinel
				errMsg = fmt.Sprintf("procedure '%s' not found", funcName)
			} else if errors.Is(procErr, ErrArgumentMismatch) {
				code = ErrorCodeArgMismatch
				wrappedErr = ErrArgumentMismatch // Preserve sentinel
				errMsg = fmt.Sprintf("argument mismatch calling procedure '%s'", funcName)
			}
			procErr = NewRuntimeError(code, errMsg, wrappedErr)
		}
		return nil, procErr
	}
	i.lastCallResult = procResult // Update last result for procedures as well
	return procResult, nil
}

// GetTypeConstant checks if a name matches a predefined type constant.
// (This function should remain as it was in your provided file)
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
