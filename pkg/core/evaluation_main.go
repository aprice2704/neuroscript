// pkg/core/evaluation_main.go
package core

import (
	"fmt"
	"strings" // Keep for variable not found check
)

// evaluateExpression evaluates an AST node representing an expression.
// Returns the evaluated RAW value. Handles new expression types.
func (i *Interpreter) evaluateExpression(node interface{}) (interface{}, error) {

	switch n := node.(type) {

	// --- Basic Value Nodes ---
	case StringLiteralNode:
		return n.Value, nil // Return RAW string value
	case NumberLiteralNode:
		return n.Value, nil // Return raw int64 or float64
	case BooleanLiteralNode:
		return n.Value, nil
	case VariableNode:
		val, exists := i.variables[n.Name]
		if !exists {
			return nil, fmt.Errorf("variable '%s' not found", n.Name)
		}
		return val, nil // Return RAW variable value
	case PlaceholderNode: // Placeholders only have meaning within EVAL
		// Evaluating a standalone placeholder node returns an error or its name?
		// Let's return an error to enforce EVAL usage.
		// return nil, fmt.Errorf("cannot evaluate raw placeholder '{{%s}}'; use EVAL()", n.Name)
		// OR - lenient approach: return the raw value of the referenced var/LAST
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
			return nil, fmt.Errorf("variable '{{%s}}' referenced in placeholder not found", varName)
		}
		return refValue, nil // Return RAW referenced value (consistent with previous behavior)
	case LastNode:
		return i.lastCallResult, nil // Return RAW last result

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
			// Key is already validated as StringLiteralNode by builder
			mapKey := entry.Key.Value
			evaluatedMap[mapKey], err = i.evaluateExpression(entry.Value)
			if err != nil {
				return nil, fmt.Errorf("evaluating value for map key %q: %w", mapKey, err)
			}
		}
		return evaluatedMap, nil

	// --- Operations ---
	case EvalNode:
		// Evaluate argument, convert to string, THEN resolve placeholders
		argValueRaw, err := i.evaluateExpression(n.Argument)
		if err != nil {
			return nil, fmt.Errorf("evaluating argument for EVAL: %w", err)
		}
		argStr := ""
		if argValueRaw != nil {
			argStr = fmt.Sprintf("%v", argValueRaw)
		}
		resolvedStr, resolveErr := i.resolvePlaceholdersWithError(argStr) // RESOLUTION HAPPENS HERE ONLY
		if resolveErr != nil {
			return nil, fmt.Errorf("resolving placeholders during EVAL: %w", resolveErr)
		}
		return resolvedStr, nil

	case UnaryOpNode:
		operandVal, err := i.evaluateExpression(n.Operand)
		if err != nil {
			// Check if error is "variable not found" - treat operand as nil in that case for NOT?
			// For now, propagate error strictly.
			return nil, fmt.Errorf("evaluating operand for unary operator '%s': %w", n.Operator, err)
		}
		return evaluateUnaryOp(n.Operator, operandVal) // Call helper

	case BinaryOpNode:
		leftVal, errL := i.evaluateExpression(n.Left)
		if errL != nil {
			// Handle "variable not found" gracefully for comparisons?
			// Treat not found var as nil for == / != ?
			if (n.Operator == "==" || n.Operator == "!=") && strings.Contains(errL.Error(), "not found") {
				leftVal = nil // Treat as nil for comparison
			} else {
				return nil, fmt.Errorf("evaluating left operand for binary operator '%s': %w", n.Operator, errL)
			}
		}
		// Short-circuit AND/OR before evaluating right side
		if n.Operator == "AND" && !isTruthy(leftVal) {
			return false, nil
		}
		if n.Operator == "OR" && isTruthy(leftVal) {
			return true, nil
		}

		rightVal, errR := i.evaluateExpression(n.Right)
		if errR != nil {
			if (n.Operator == "==" || n.Operator == "!=") && strings.Contains(errR.Error(), "not found") {
				rightVal = nil // Treat as nil for comparison
			} else {
				return nil, fmt.Errorf("evaluating right operand for binary operator '%s': %w", n.Operator, errR)
			}
		}
		return evaluateBinaryOp(leftVal, rightVal, n.Operator) // Call helper

	case FunctionCallNode:
		evaluatedArgs := make([]interface{}, len(n.Arguments))
		var err error
		for idx, argNode := range n.Arguments {
			evaluatedArgs[idx], err = i.evaluateExpression(argNode)
			if err != nil {
				return nil, fmt.Errorf("evaluating argument %d for function '%s': %w", idx+1, n.FunctionName, err)
			}
		}
		return evaluateFunctionCall(n.FunctionName, evaluatedArgs) // Call helper

	case ElementAccessNode:
		return i.evaluateElementAccess(n) // Existing logic

	// --- Pass-through for already evaluated values ---
	case string, int64, float64, bool, nil, []interface{}, map[string]interface{}:
		return n, nil // Return value if node is already a primitive/collection type

	// --- Error for unexpected node types ---
	default:
		return nil, fmt.Errorf("internal error: evaluateExpression encountered unhandled node type: %T", node)
	}
}
