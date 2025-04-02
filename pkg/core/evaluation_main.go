// pkg/core/evaluation_main.go
// Contains the main evaluateExpression function.
package core

import (
	"fmt"
	"strings"
)

// evaluateExpression evaluates an AST node representing an expression.
// Returns the evaluated value and an error if evaluation fails.
func (i *Interpreter) evaluateExpression(node interface{}) (interface{}, error) {

	switch n := node.(type) {
	case StringLiteralNode:
		resolved, err := i.resolvePlaceholdersWithError(n.Value)
		if err != nil {
			return nil, fmt.Errorf("resolving placeholders in string literal %q: %w", n.Value, err)
		}
		return resolved, nil
	case NumberLiteralNode:
		return n.Value, nil
	case BooleanLiteralNode:
		return n.Value, nil
	case VariableNode:
		val, exists := i.variables[n.Name]
		if !exists {
			return nil, fmt.Errorf("variable '%s' not found", n.Name)
		}
		if strVal, ok := val.(string); ok {
			resolved, err := i.resolvePlaceholdersWithError(strVal)
			if err != nil {
				return nil, fmt.Errorf("resolving placeholders in variable '%s' value %q: %w", n.Name, strVal, err)
			}
			return resolved, nil
		}
		return val, nil // Return raw value (list, map, number, bool, nil)
	case PlaceholderNode:
		var refValue interface{}
		if n.Name == "__last_call_result" {
			refValue = i.lastCallResult
		} else {
			val, exists := i.variables[n.Name]
			if !exists {
				return nil, fmt.Errorf("variable '{{%s}}' referenced in placeholder not found", n.Name)
			}
			refValue = val
		}
		if strVal, ok := refValue.(string); ok {
			resolved, err := i.resolvePlaceholdersWithError(strVal)
			if err != nil {
				return nil, fmt.Errorf("resolving placeholders in placeholder '{{%s}}' value %q: %w", n.Name, strVal, err)
			}
			return resolved, nil
		}
		return refValue, nil // Return raw value
	case LastCallResultNode:
		if strVal, ok := i.lastCallResult.(string); ok {
			resolved, err := i.resolvePlaceholdersWithError(strVal)
			if err != nil {
				return nil, fmt.Errorf("resolving placeholders in __last_call_result value %q: %w", strVal, err)
			}
			return resolved, nil
		}
		return i.lastCallResult, nil // Return raw value
	case ConcatenationNode:
		var builder strings.Builder
		for iOp, operandNode := range n.Operands {
			evaluatedOperand, err := i.evaluateExpression(operandNode)
			if err != nil {
				return nil, fmt.Errorf("evaluating operand %d for concatenation: %w", iOp, err)
			}
			if evaluatedOperand == nil {
				builder.WriteString("") // Treat nil as empty string
			} else {
				builder.WriteString(fmt.Sprintf("%v", evaluatedOperand))
			}
		}
		return builder.String(), nil
	case ListLiteralNode:
		evaluatedElements := make([]interface{}, len(n.Elements))
		for idx, elemNode := range n.Elements {
			var err error
			evaluatedElements[idx], err = i.evaluateExpression(elemNode)
			if err != nil {
				return nil, fmt.Errorf("evaluating element %d in list literal: %w", idx, err)
			}
		}
		return evaluatedElements, nil
	case MapLiteralNode:
		evaluatedMap := make(map[string]interface{})
		for _, entry := range n.Entries {
			mapKey := entry.Key.Value
			mapValue, err := i.evaluateExpression(entry.Value)
			if err != nil {
				return nil, fmt.Errorf("evaluating value for key %q in map literal: %w", mapKey, err)
			}
			evaluatedMap[mapKey] = mapValue
		}
		return evaluatedMap, nil

	case ElementAccessNode:
		// Delegate to the helper function in evaluation_access.go
		return i.evaluateElementAccess(n)

	// --- Handle Raw Go Types (Pass-through if needed) ---
	case string:
		resolved, err := i.resolvePlaceholdersWithError(n)
		if err != nil {
			return nil, fmt.Errorf("resolving placeholders in raw string %q: %w", n, err)
		}
		return resolved, nil
	case int64, float64, bool, nil, []interface{}, map[string]interface{}:
		return n, nil // Pass through basic Go types and collections

	// --- Error Cases / Invalid Nodes ---
	case ComparisonNode:
		return nil, fmt.Errorf("internal error: evaluateExpression called directly on ComparisonNode")
	default:
		return nil, fmt.Errorf("internal error: evaluateExpression encountered unhandled node type: %T", node)
	}
}
