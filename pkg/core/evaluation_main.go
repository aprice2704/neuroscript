// pkg/core/evaluation_main.go
package core

import (
	"errors"
	"fmt"
)

// evaluateExpression evaluates an AST node representing an expression.
// Returns the evaluated RAW value. Handles new expression types.
func (i *Interpreter) evaluateExpression(node interface{}) (interface{}, error) {

	switch n := node.(type) {

	// --- Basic Value Nodes ---
	case StringLiteralNode:
		// RAW strings (```...```) potentially need placeholder evaluation by default.
		// Regular strings ('...', "...") need EVAL().
		// Let's handle this based on the IsRaw flag.
		if n.IsRaw {
			// Design doc says ``` strings evaluate placeholders by default.
			resolvedStr, resolveErr := i.resolvePlaceholdersWithError(n.Value)
			if resolveErr != nil {
				return nil, fmt.Errorf("evaluating raw string literal '%s...': %w", n.Value[:min(len(n.Value), 20)], resolveErr)
			}
			return resolvedStr, nil
		}
		// Regular strings are returned raw, need EVAL() for placeholders.
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
		if n.Operator == "and" && !isTruthy(leftVal) {
			return false, nil
		}
		if n.Operator == "or" && isTruthy(leftVal) {
			return true, nil
		}
		rightVal, errR := i.evaluateExpression(n.Right)
		if errR != nil {
			if (n.Operator == "==" || n.Operator == "!=") && errors.Is(errR, ErrVariableNotFound) {
				rightVal = nil
			} else {
				return nil, fmt.Errorf("evaluating right operand for '%s': %w", n.Operator, errR)
			}
		}
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
		return evaluateFunctionCall(n.FunctionName, evaluatedArgs)
	case ElementAccessNode:
		return i.evaluateElementAccess(n)

	// --- Pass-through ---
	case string, int64, float64, bool, nil, []interface{}, map[string]interface{}, []string:
		return n, nil

	default:
		return nil, fmt.Errorf("internal error: evaluateExpression unhandled node type: %T", node)
	}
}
