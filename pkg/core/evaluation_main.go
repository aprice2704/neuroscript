// filename: pkg/core/evaluation_main.go
package core

import (
	"errors"
	"fmt"
	// Keep math if other helpers need it eventually
	// Keep reflect if other helpers need it
)

// evaluateExpression evaluates an AST node representing an expression.
// Returns the evaluated RAW value.
func (i *Interpreter) evaluateExpression(node interface{}) (interface{}, error) {

	switch n := node.(type) {

	// --- Basic Value Nodes ---
	case StringLiteralNode:
		if n.IsRaw {
			resolvedStr, resolveErr := i.resolvePlaceholdersWithError(n.Value)
			if resolveErr != nil {
				// Use min helper if available, otherwise inline
				maxLength := 20
				if len(n.Value) < maxLength {
					maxLength = len(n.Value)
				}
				return nil, fmt.Errorf("evaluating raw string literal '%s...': %w", n.Value[:maxLength], resolveErr)
			}
			return resolvedStr, nil
		}
		return n.Value, nil
	case NumberLiteralNode:
		return n.Value, nil
	case BooleanLiteralNode:
		return n.Value, nil
	case VariableNode:
		val, exists := i.variables[n.Name]
		if !exists {
			return nil, fmt.Errorf("%w: '%s'", ErrVariableNotFound, n.Name)
		}
		return val, nil
	case PlaceholderNode:
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
			return nil, fmt.Errorf("%w: '{{%s}}' referenced in placeholder", ErrVariableNotFound, varName)
		}
		return refValue, nil
	case LastNode:
		return i.lastCallResult, nil

	// --- Collection Literals ---
	case ListLiteralNode:
		evaluatedElements := make([]interface{}, len(n.Elements))
		var err error
		for idx, elemNode := range n.Elements {
			evaluatedElements[idx], err = i.evaluateExpression(elemNode)
			if err != nil {
				return nil, fmt.Errorf("evaluating list literal element %d: %w", idx, err)
			}
		}
		return evaluatedElements, nil
	case MapLiteralNode:
		evaluatedMap := make(map[string]interface{})
		var err error
		for _, entry := range n.Entries {
			mapKey := entry.Key.Value
			evaluatedMap[mapKey], err = i.evaluateExpression(entry.Value)
			if err != nil {
				return nil, fmt.Errorf("evaluating value for map key %q: %w", mapKey, err)
			}
		}
		return evaluatedMap, nil

	// --- Operations ---
	case EvalNode:
		argValueRaw, err := i.evaluateExpression(n.Argument)
		if err != nil {
			return nil, fmt.Errorf("evaluating argument for EVAL: %w", err)
		}
		argStr := ""
		if argValueRaw != nil {
			argStr = fmt.Sprintf("%v", argValueRaw)
		}
		resolvedStr, resolveErr := i.resolvePlaceholdersWithError(argStr)
		if resolveErr != nil {
			return nil, fmt.Errorf("resolving placeholders during EVAL: %w", resolveErr)
		}
		return resolvedStr, nil
	case UnaryOpNode:
		operandVal, err := i.evaluateExpression(n.Operand)
		if err != nil {
			return nil, fmt.Errorf("evaluating operand for unary operator '%s': %w", n.Operator, err)
		}
		// Delegate to helper in evaluation_logic.go
		return evaluateUnaryOp(n.Operator, operandVal)
	case BinaryOpNode:
		leftVal, errL := i.evaluateExpression(n.Left)
		if errL != nil {
			if (n.Operator == "==" || n.Operator == "!=") && errors.Is(errL, ErrVariableNotFound) {
				leftVal = nil
			} else {
				return nil, fmt.Errorf("evaluating left operand for '%s': %w", n.Operator, errL)
			}
		}
		if n.Operator == "and" {
			if !isTruthy(leftVal) {
				return false, nil
			}
		} else if n.Operator == "or" {
			if isTruthy(leftVal) {
				return true, nil
			}
		}
		rightVal, errR := i.evaluateExpression(n.Right)
		if errR != nil {
			if (n.Operator == "==" || n.Operator == "!=") && errors.Is(errR, ErrVariableNotFound) {
				rightVal = nil
			} else {
				return nil, fmt.Errorf("evaluating right operand for '%s': %w", n.Operator, errR)
			}
		}
		// Delegate actual operation
		return evaluateBinaryOp(leftVal, rightVal, n.Operator)

	// MODIFIED: Handle CallableExprNode
	case CallableExprNode:
		target := n.Target
		targetName := target.Name

		// 1. Evaluate Arguments first
		evaluatedArgs := make([]interface{}, len(n.Arguments))
		var argErr error
		for idx, argNode := range n.Arguments {
			evaluatedArgs[idx], argErr = i.evaluateExpression(argNode)
			if argErr != nil {
				callDesc := targetName
				if target.IsTool {
					callDesc = "tool." + targetName
				}
				return nil, fmt.Errorf("evaluating arg %d for call to '%s': %w", idx+1, callDesc, argErr)
			}
		}

		// 2. Determine Call Type and Execute
		if target.IsTool {
			// --- Tool Call ---
			i.Logger().Debug("[DEBUG-EVAL]   Calling Tool '%s' from expression", targetName)
			toolImpl, found := i.ToolRegistry().GetTool(targetName)
			if !found {
				errMsg := fmt.Sprintf("tool '%s' not found", targetName)
				return nil, NewRuntimeError(ErrorCodeToolNotFound, errMsg, fmt.Errorf("%s: %w", errMsg, ErrToolNotFound))
			}
			validatedArgs, validationErr := ValidateAndConvertArgs(toolImpl.Spec, evaluatedArgs)
			if validationErr != nil {
				code := ErrorCodeArgMismatch
				if errors.Is(validationErr, ErrValidationTypeMismatch) {
					code = ErrorCodeType
				} else if errors.Is(validationErr, ErrValidationArgCount) {
					code = ErrorCodeArgMismatch
				}
				return nil, NewRuntimeError(code, fmt.Sprintf("args failed for tool '%s'", targetName), fmt.Errorf("validating args for %s: %w", targetName, validationErr))
			}
			toolResult, toolErr := toolImpl.Func(i, validatedArgs)
			if toolErr != nil {
				if re, ok := toolErr.(*RuntimeError); ok {
					return nil, re
				}
				return nil, NewRuntimeError(ErrorCodeToolSpecific, fmt.Sprintf("tool '%s' failed", targetName), fmt.Errorf("executing tool %s: %w", targetName, toolErr))
			}
			i.Logger().Debug("[DEBUG-EVAL]   Tool '%s' call successful (Result Type: %T)", targetName, toolResult)
			i.lastCallResult = toolResult
			return toolResult, nil

		} else {
			// --- User Procedure or Built-in Function Call ---
			// Delegate to a combined handler (can be in evaluation_logic.go)
			// This avoids duplicating the recursive call logic here vs built-in logic
			// Let's call it evaluateUserOrBuiltInFunction
			i.Logger().Debug("[DEBUG-EVAL]   Calling User Proc or Built-in '%s' from expression", targetName)
			// Note: evaluateUserOrBuiltInFunction needs access to 'i' (Interpreter)
			// to call RunProcedure recursively.
			result, err := i.evaluateUserOrBuiltInFunction(targetName, evaluatedArgs) // Need to implement this
			if err != nil {
				// Error should already be wrapped by the helper
				return nil, err
			}
			// Decide if user procs update LAST, but built-ins don't.
			// The helper can determine this. For now, assume helper updates i.lastCallResult internally if needed.
			// i.lastCallResult = result // Let helper handle this
			return result, nil
		}
		// --- End Call Type Handling ---

	case ElementAccessNode:
		// Delegate to helper in evaluation_access.go
		return i.evaluateElementAccess(n)

	// --- Pass-through for already evaluated values ---
	default:
		switch node.(type) {
		case string, int64, float64, bool, nil, []interface{}, map[string]interface{}, []string:
			return node, nil // Return primitive types directly
		}
		return nil, fmt.Errorf("internal error: evaluateExpression unhandled node type: %T", node)
	}
}

// --- REMOVED evaluateBuiltInFunction ---
// func evaluateBuiltInFunction(funcName string, args []interface{}) (interface{}, error) { ... }

// --- Helpers (Assume these exist or add them) ---

// --- ADDED Placeholder for combined user/built-in function evaluator ---
// Needs implementation, likely in evaluation_logic.go or a new file.
func (i *Interpreter) evaluateUserOrBuiltInFunction(funcName string, args []interface{}) (interface{}, error) {
	// 1. Check if it's a built-in
	if isBuiltInFunction(funcName) { // Assumes isBuiltInFunction is defined correctly
		result, err := evaluateBuiltInFunction(funcName, args) // Call the actual built-in logic
		if err != nil {
			if _, ok := err.(*RuntimeError); !ok {
				err = NewRuntimeError(ErrorCodeGeneric, fmt.Sprintf("built-in function '%s' failed", funcName), err)
			}
			return nil, err
		}
		// Do NOT update i.lastCallResult for built-ins
		return result, nil
	}

	// 2. If not built-in, assume it's a User Procedure
	procResult, procErr := i.RunProcedure(funcName, args...)
	if procErr != nil {
		if _, ok := procErr.(*RuntimeError); !ok {
			code := ErrorCodeGeneric
			wrapped := procErr
			errMsg := procErr.Error()
			if errors.Is(procErr, ErrProcedureNotFound) {
				code = ErrorCodeProcNotFound
				wrapped = ErrProcedureNotFound
				errMsg = fmt.Sprintf("procedure '%s' not found", funcName)
			} else if errors.Is(procErr, ErrArgumentMismatch) {
				code = ErrorCodeArgMismatch
				wrapped = ErrArgumentMismatch
				errMsg = fmt.Sprintf("argument mismatch calling procedure '%s'", funcName)
			}
			procErr = NewRuntimeError(code, errMsg, fmt.Errorf("calling procedure %s: %w", funcName, wrapped))
		}
		return nil, procErr
	}
	// DO update i.lastCallResult for user procedures
	i.lastCallResult = procResult
	return procResult, nil
}

// isBuiltInFunction, evaluateBuiltInFunction (implementation) needs to be defined, likely in evaluation_logic.go
// evaluateUnaryOp, evaluateBinaryOp, isTruthy are assumed to exist (evaluation_logic.go, evaluation_helpers.go)
// resolvePlaceholdersWithError is assumed to exist (evaluation_resolve.go)
// evaluateElementAccess is assumed to exist (evaluation_access.go)
// toFloat64 needs to be defined (e.g., in evaluation_helpers.go)
