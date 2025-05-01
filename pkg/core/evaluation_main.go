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

	// Ensure node is not nil before proceeding (defensive check)
	if node == nil {
		// Or return a specific error if appropriate
		return nil, fmt.Errorf("internal error: evaluateExpression received nil node")
	}

	// --- MODIFIED: Switch cases now expect POINTERS (*NodeType) ---
	switch n := node.(type) {

	// --- Basic Value Nodes ---
	case *StringLiteralNode: // Pointer
		if n.IsRaw {
			resolvedStr, resolveErr := i.resolvePlaceholdersWithError(n.Value)
			if resolveErr != nil {
				maxLength := 20
				if len(n.Value) < maxLength {
					maxLength = len(n.Value)
				}
				return nil, fmt.Errorf("evaluating raw string literal '%s...': %w", n.Value[:maxLength], resolveErr)
			}
			return resolvedStr, nil
		}
		return n.Value, nil
	case *NumberLiteralNode: // Pointer
		return n.Value, nil
	case *BooleanLiteralNode: // Pointer
		return n.Value, nil
	case *VariableNode: // Pointer
		val, exists := i.variables[n.Name]
		if !exists {
			return nil, fmt.Errorf("%w: '%s'", ErrVariableNotFound, n.Name)
		}
		return val, nil
	case *PlaceholderNode: // Pointer
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
	case *LastNode: // Pointer
		return i.lastCallResult, nil

	// --- Collection Literals ---
	case *ListLiteralNode: // Pointer
		evaluatedElements := make([]interface{}, len(n.Elements))
		var err error
		for idx, elemNode := range n.Elements {
			evaluatedElements[idx], err = i.evaluateExpression(elemNode)
			if err != nil {
				return nil, fmt.Errorf("evaluating list literal element %d: %w", idx, err)
			}
		}
		return evaluatedElements, nil
	case *MapLiteralNode: // Pointer
		evaluatedMap := make(map[string]interface{})
		var err error
		for _, entry := range n.Entries {
			// Map keys are StringLiteralNodes, access their value directly
			mapKey := entry.Key.Value
			evaluatedMap[mapKey], err = i.evaluateExpression(entry.Value)
			if err != nil {
				return nil, fmt.Errorf("evaluating value for map key %q: %w", mapKey, err)
			}
		}
		return evaluatedMap, nil

	// --- Operations ---
	case *EvalNode: // Pointer
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
	case *UnaryOpNode: // Pointer
		operandVal, err := i.evaluateExpression(n.Operand)
		if err != nil {
			return nil, fmt.Errorf("evaluating operand for unary operator '%s': %w", n.Operator, err)
		}
		// Delegate to helper in evaluation_operators.go
		return evaluateUnaryOp(n.Operator, operandVal)
	case *BinaryOpNode: // Pointer
		// Use structured logging - CORRECTED
		i.Logger().Debug("[DEBUG-EVAL-BINOP] Evaluating BinaryOpNode", "operator", n.Operator)

		leftVal, errL := i.evaluateExpression(n.Left)
		if errL != nil {
			// Allow ==/!= comparison with potentially undefined variable (treat as nil)
			if (n.Operator == "==" || n.Operator == "!=") && errors.Is(errL, ErrVariableNotFound) {
				// Use structured logging - CORRECTED
				// Get string representation of the left node for logging context
				leftNodeStr := NodeToString(n.Left) // Assuming NodeToString exists
				i.Logger().Debug("[DEBUG-EVAL-BINOP] Left operand not found, treating as nil for comparison", "operand_str", leftNodeStr)
				leftVal = nil
			} else {
				// Use structured logging - CORRECTED
				i.Logger().Error("[DEBUG-EVAL-BINOP] Error evaluating left operand", "operator", n.Operator, "error", errL)
				return nil, fmt.Errorf("evaluating left operand for '%s': %w", n.Operator, errL)
			}
		}
		// Use structured logging - CORRECTED
		i.Logger().Debug("[DEBUG-EVAL-BINOP] Left operand evaluated", "value", leftVal, "type", fmt.Sprintf("%T", leftVal))

		// Handle short-circuiting for 'and' and 'or'
		if n.Operator == "and" {
			if !isTruthy(leftVal) {
				// Use structured logging - CORRECTED
				i.Logger().Debug("[DEBUG-EVAL-BINOP] Short-circuiting 'and' (left is falsey)")
				return false, nil
			}
		} else if n.Operator == "or" {
			if isTruthy(leftVal) {
				// Use structured logging - CORRECTED
				i.Logger().Debug("[DEBUG-EVAL-BINOP] Short-circuiting 'or' (left is truthy)")
				return true, nil
			}
		}

		rightVal, errR := i.evaluateExpression(n.Right)
		if errR != nil {
			// Allow ==/!= comparison with potentially undefined variable (treat as nil)
			if (n.Operator == "==" || n.Operator == "!=") && errors.Is(errR, ErrVariableNotFound) {
				// Use structured logging - CORRECTED
				// Get string representation of the right node for logging context
				rightNodeStr := NodeToString(n.Right) // Assuming NodeToString exists
				i.Logger().Debug("[DEBUG-EVAL-BINOP] Right operand not found, treating as nil for comparison", "operand_str", rightNodeStr)
				rightVal = nil
			} else {
				// Use structured logging - CORRECTED
				i.Logger().Error("[DEBUG-EVAL-BINOP] Error evaluating right operand", "operator", n.Operator, "error", errR)
				return nil, fmt.Errorf("evaluating right operand for '%s': %w", n.Operator, errR)
			}
		}
		// Use structured logging - CORRECTED
		i.Logger().Debug("[DEBUG-EVAL-BINOP] Right operand evaluated", "value", rightVal, "type", fmt.Sprintf("%T", rightVal))

		// Delegate actual operation (in evaluation_operators.go?)
		// Use structured logging - CORRECTED
		i.Logger().Debug("[DEBUG-EVAL-BINOP] Calling evaluateBinaryOp", "left_value", leftVal, "right_value", rightVal, "operator", n.Operator)
		result, err := evaluateBinaryOp(leftVal, rightVal, n.Operator)
		if err != nil {
			// Wrap the error with context if it's not already a RuntimeError
			if _, ok := err.(*RuntimeError); !ok {
				// Use ErrorCodeEvaluation for errors from the binary operation itself
				err = NewRuntimeError(ErrorCodeEvaluation, fmt.Sprintf("operation '%s' failed", n.Operator), err)
			}
			// Use structured logging - CORRECTED
			i.Logger().Error("[DEBUG-EVAL-BINOP] Error from evaluateBinaryOp", "operator", n.Operator, "error", err)
			return nil, err // Return the (potentially wrapped) error
		}
		// Use structured logging - CORRECTED
		i.Logger().Debug("[DEBUG-EVAL-BINOP] evaluateBinaryOp successful", "result", result, "type", fmt.Sprintf("%T", result))
		return result, nil

	case *CallableExprNode: // Pointer
		target := n.Target // Target is a value type CallTarget within the pointer node
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
			// Use structured logging - CORRECTED
			i.Logger().Debug("[DEBUG-EVAL] Calling Tool from expression", "tool_name", targetName)
			toolImpl, found := i.ToolRegistry().GetTool(targetName)
			if !found {
				errMsg := fmt.Sprintf("tool '%s' not found", targetName)
				return nil, NewRuntimeError(ErrorCodeToolNotFound, errMsg, fmt.Errorf("%s: %w", errMsg, ErrToolNotFound))
			}
			validatedArgs, validationErr := ValidateAndConvertArgs(toolImpl.Spec, evaluatedArgs)
			if validationErr != nil {
				code := ErrorCodeArgMismatch
				// Example check for specific validation error type if defined elsewhere
				// if errors.Is(validationErr, ErrValidationTypeMismatch) { code = ErrorCodeType }
				return nil, NewRuntimeError(code, fmt.Sprintf("args failed for tool '%s'", targetName), fmt.Errorf("validating args for %s: %w", targetName, validationErr))
			}
			toolResult, toolErr := toolImpl.Func(i, validatedArgs)
			if toolErr != nil {
				if re, ok := toolErr.(*RuntimeError); ok {
					return nil, re // Return existing RuntimeError
				}
				// Wrap non-RuntimeError
				return nil, NewRuntimeError(ErrorCodeToolSpecific, fmt.Sprintf("tool '%s' failed", targetName), fmt.Errorf("executing tool %s: %w", targetName, toolErr))
			}
			// Use structured logging - CORRECTED
			i.Logger().Debug("[DEBUG-EVAL] Tool call successful", "tool_name", targetName, "result_type", fmt.Sprintf("%T", toolResult))
			i.lastCallResult = toolResult
			return toolResult, nil

		} else {
			// --- User Procedure or Built-in Function Call ---
			// Use structured logging - CORRECTED
			i.Logger().Debug("[DEBUG-EVAL] Calling User Proc or Built-in from expression", "function_name", targetName)
			// Note: evaluateUserOrBuiltInFunction needs access to 'i' (Interpreter)
			result, err := i.evaluateUserOrBuiltInFunction(targetName, evaluatedArgs)
			if err != nil {
				return nil, err // Error should already be wrapped
			}
			// evaluateUserOrBuiltInFunction updates i.lastCallResult internally if needed (for user procs)
			return result, nil
		}
		// --- End Call Type Handling ---

	case *ElementAccessNode: // Pointer
		// Delegate to helper in evaluation_access.go
		// Pass the pointer 'n' directly
		return i.evaluateElementAccess(n)

	// --- Pass-through for already evaluated primitive/collection values ---
	// (This default case handles results from previous evaluations)
	default:
		switch node.(type) { // Nested switch still checks for value types
		case string, int64, float64, bool, nil, []interface{}, map[string]interface{}, []string:
			return node, nil // Return primitive types directly
		}
		// If it's not an AST node type handled above AND not a primitive/collection, it's an error
		return nil, fmt.Errorf("internal error: evaluateExpression unhandled node type: %T", node)
	}
}

// evaluateUserOrBuiltInFunction handles dispatching to built-ins or user procedures.
// (Needs to be defined, likely in evaluation_logic.go or similar)
func (i *Interpreter) evaluateUserOrBuiltInFunction(funcName string, args []interface{}) (interface{}, error) {
	// 1. Check if it's a built-in
	if isBuiltInFunction(funcName) { // Assumes isBuiltInFunction is defined correctly
		result, err := evaluateBuiltInFunction(funcName, args) // Call the actual built-in logic
		if err != nil {
			if _, ok := err.(*RuntimeError); !ok {
				// Wrap non-runtime errors from built-ins
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
		// Ensure errors from RunProcedure are RuntimeError
		if _, ok := procErr.(*RuntimeError); !ok {
			code := ErrorCodeGeneric
			wrapped := procErr
			errMsg := procErr.Error()
			// Check for specific underlying errors if needed
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

// --- Required Helper Implementations ---
// These functions are called by evaluateExpression and need to exist,
// likely in files like evaluation_logic.go, evaluation_operators.go,
// evaluation_helpers.go, evaluation_resolve.go, evaluation_access.go.

// isBuiltInFunction(funcName string) bool
// evaluateBuiltInFunction(funcName string, args []interface{}) (interface{}, error)
// evaluateUnaryOp(operator string, operand interface{}) (interface{}, error)
// evaluateBinaryOp(left, right interface{}, operator string) (interface{}, error)
// isTruthy(value interface{}) bool
// resolvePlaceholdersWithError(template string) (string, error)
// (evaluateElementAccess is in evaluation_access.go and takes *ElementAccessNode)
// NodeToString(node interface{}) string // Added assumption
