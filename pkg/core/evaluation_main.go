// pkg/core/evaluation_main.go
package core

import (
	"fmt"
	"strings"
)

// evaluateExpression evaluates an AST node representing an expression.
// Returns the evaluated RAW value. Placeholders are only resolved via EvalNode.
func (i *Interpreter) evaluateExpression(node interface{}) (interface{}, error) {

	switch n := node.(type) {
	case StringLiteralNode:
		return n.Value, nil // Return RAW string value
	case NumberLiteralNode:
		return n.Value, nil
	case BooleanLiteralNode:
		return n.Value, nil
	case VariableNode:
		val, exists := i.variables[n.Name]
		if !exists {
			return nil, fmt.Errorf("variable '%s' not found", n.Name)
		}
		return val, nil // Return RAW variable value
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
			return nil, fmt.Errorf("variable '{{%s}}' referenced in placeholder not found", varName)
		}
		return refValue, nil // Return RAW referenced value
	case LastNode:
		return i.lastCallResult, nil // Return RAW last result
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
	case ConcatenationNode:
		// *** Refined: Explicit string check before fmt.Sprintf, STILL NO RESOLUTION ***
		var builder strings.Builder
		for iOp, operandNode := range n.Operands {
			evaluatedOperand, err := i.evaluateExpression(operandNode) // Gets raw value
			if err != nil {
				return nil, fmt.Errorf("evaluating operand %d for concatenation: %w", iOp, err)
			}

			// Convert raw evaluated value directly to string representation
			var operandStr string
			if evaluatedOperand == nil {
				operandStr = "" // Represent nil as empty string in concatenation
			} else if strVal, ok := evaluatedOperand.(string); ok {
				operandStr = strVal // Use string directly if already string
			} else {
				operandStr = fmt.Sprintf("%v", evaluatedOperand) // Use fmt for other types
			}

			// *** NO CALL to resolvePlaceholdersWithError ***
			builder.WriteString(operandStr)
		}
		return builder.String(), nil // Return concatenated raw strings

	// ... (List, Map, ElementAccess, passthrough, default cases as before) ...
	case ListLiteralNode:
		evaluatedElements := make([]interface{}, len(n.Elements))
		var err error
		for idx, elemNode := range n.Elements {
			evaluatedElements[idx], err = i.evaluateExpression(elemNode)
			if err != nil {
				return nil, fmt.Errorf("evaluating element %d: %w", idx, err)
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
				return nil, fmt.Errorf("evaluating value for key %q: %w", mapKey, err)
			}
		}
		return evaluatedMap, nil
	case ElementAccessNode:
		return i.evaluateElementAccess(n)
	case string, int64, float64, bool, nil, []interface{}, map[string]interface{}:
		return n, nil
	case ComparisonNode:
		return nil, fmt.Errorf("internal error: evaluateExpression called directly on ComparisonNode")
	default:
		return nil, fmt.Errorf("internal error: evaluateExpression encountered unhandled node type: %T", node)
	}
}
