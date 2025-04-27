// filename: pkg/core/evaluation_main.go
package core

import (
	"errors"
	"fmt"
)

// evaluateExpression evaluates an AST node representing an expression.
// Returns the evaluated RAW value. Handles new expression types.
// UPDATED: Implemented short-circuiting for AND/OR
func (i *Interpreter) evaluateExpression(node interface{}) (interface{}, error) {

	switch n := node.(type) {

	// --- Basic Value Nodes ---
	case StringLiteralNode:
		// RAW strings (```...```) potentially need placeholder evaluation by default.
		// Regular strings ('...', "...") need EVAL().
		if n.IsRaw {
			resolvedStr, resolveErr := i.resolvePlaceholdersWithError(n.Value)
			if resolveErr != nil {
				return nil, fmt.Errorf("evaluating raw string literal '%s...': %w", n.Value[:min(len(n.Value), 20)], resolveErr)
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
			// Handle variable not found specifically for conditions if needed? No, let it error normally first.
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
			// Evaluate the key expression first
			keyValRaw, keyErr := i.evaluateExpression(entry.Key)
			if keyErr != nil {
				return nil, fmt.Errorf("evaluating map key expression: %w", keyErr)
			}
			// Convert evaluated key to string
			mapKey := fmt.Sprintf("%v", keyValRaw)

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
		// Evaluate left operand first
		leftVal, errL := i.evaluateExpression(n.Left)
		if errL != nil {
			// Special handling for ==/!= where var not found becomes nil
			if (n.Operator == "==" || n.Operator == "!=") && errors.Is(errL, ErrVariableNotFound) {
				leftVal = nil // Treat not found as nil for comparison
			} else {
				// For other operators or errors, propagate the error
				return nil, fmt.Errorf("evaluating left operand for '%s': %w", n.Operator, errL)
			}
		}

		// *** Implement Short-Circuiting for AND/OR HERE ***
		if n.Operator == "and" || n.Operator == "AND" {
			leftBool := isTruthy(leftVal)
			if !leftBool {
				return false, nil // Short-circuit: false AND anything is false
			}
			// Don't return yet, need right side
		} else if n.Operator == "or" || n.Operator == "OR" {
			leftBool := isTruthy(leftVal)
			if leftBool {
				return true, nil // Short-circuit: true OR anything is true
			}
			// Don't return yet, need right side
		}

		// If not short-circuited (or not AND/OR), evaluate right operand
		rightVal, errR := i.evaluateExpression(n.Right)
		if errR != nil {
			// Special handling for ==/!= where var not found becomes nil
			if (n.Operator == "==" || n.Operator == "!=") && errors.Is(errR, ErrVariableNotFound) {
				rightVal = nil // Treat not found as nil for comparison
			} else {
				// For other operators or errors, propagate the error
				return nil, fmt.Errorf("evaluating right operand for '%s': %w", n.Operator, errR)
			}
		}

		// Delegate the actual operation (including non-short-circuited AND/OR)
		// to the helper in evaluation_logic.go
		return evaluateBinaryOp(leftVal, rightVal, n.Operator)

	case FunctionCallNode:
		evaluatedArgs := make([]interface{}, len(n.Arguments))
		var err error
		for idx, argNode := range n.Arguments {
			evaluatedArgs[idx], err = i.evaluateExpression(argNode)
			if err != nil {
				return nil, fmt.Errorf("evaluating arg %d for func '%s': %w", idx+1, n.FunctionName, err)
			}
		}
		// Delegate to helper in evaluation_logic.go
		return evaluateFunctionCall(n.FunctionName, evaluatedArgs)
	case ElementAccessNode:
		// Delegate to helper in evaluation_access.go
		return i.evaluateElementAccess(n)

	// --- Pass-through for already evaluated values ---
	// (string, int64, etc., handled previously)
	default:
		// Check if it's a simple value type that doesn't need evaluation
		switch node.(type) {
		case string, int64, float64, bool, nil, []interface{}, map[string]interface{}, []string:
			return node, nil // Return primitive types directly
		}
		// Otherwise, it's an unhandled node type
		return nil, fmt.Errorf("internal error: evaluateExpression unhandled node type: %T", node)
	}
}
