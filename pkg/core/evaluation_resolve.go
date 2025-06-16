// filename: pkg/core/evaluation_resolve.go
package core

import (
	"fmt"
	// Assuming errors like ErrVariableNotFound are defined in errors.go
	// Assuming AST node types like VariableNode, PlaceholderNode, StringLiteralNode etc. are defined in ast.go
)

// --- Value Resolution ---

// resolveValue handles resolving variable names, placeholders, and literals to their actual values.
// It's called by evaluateExpression.
// Returns the resolved value or an error (e.g., variable not found).
func (i *Interpreter) resolveValue(node interface{}) (interface{}, error) {
	i.Logger().Debug("[DEBUG EVAL] Resolving value for node type: %T", node)

	switch n := node.(type) {
	case VariableNode:
		val, found := i.GetVariable(n.Name)
		if !found {
			// Wrap ErrVariableNotFound
			errMsg := fmt.Sprintf("variable '%s' not found", n.Name)
			return nil, fmt.Errorf("%s: %w", errMsg, ErrVariableNotFound)
		}
		i.Logger().Debug("[DEBUG EVAL]   Resolved Variable '%s' to: %v (%T)", n.Name, val, val)
		return val, nil

	case PlaceholderNode: // Assuming PlaceholderNode is distinct, otherwise handled by VariableNode
		// This might be used if placeholders have different lookup rules than variables.
		// For now, treat like VariableNode.
		val, found := i.GetVariable(n.Name)
		if !found {
			// Wrap ErrVariableNotFound
			errMsg := fmt.Sprintf("placeholder variable '%s' not found", n.Name)
			return nil, fmt.Errorf("%s: %w", errMsg, ErrVariableNotFound)
		}
		i.Logger().Debug("[DEBUG EVAL]   Resolved Placeholder '%s' to: %v (%T)", n.Name, val, val)
		return val, nil

	case LastNode:
		// Return the result stored by the interpreter from the *previous step*
		// Use i.lastCallResult as determined previously
		i.Logger().Debug("[DEBUG EVAL]   Resolved LAST to: %v (%T)", i.lastCallResult, i.lastCallResult)
		return i.lastCallResult, nil

	// --- Literals ---
	case StringLiteralNode:
		// Implement placeholder substitution for raw strings
		if n.IsRaw {
			i.Logger().Debug("[DEBUG EVAL]   Resolving Raw String Literal, evaluating placeholders...")
			// FIX: Call the renamed function
			substitutedValue, err := i.resolvePlaceholdersWithError(n.Value)
			if err != nil {
				// Wrap error for context
				return nil, fmt.Errorf("evaluating placeholders in raw string: %w", err)
			}
			i.Logger().Debug("[DEBUG EVAL]     Raw string after substitution: %q", substitutedValue)
			return substitutedValue, nil
		} else {
			// Normal (double-quoted) string, return as is
			i.Logger().Debug("[DEBUG EVAL]   Resolved String Literal: %q", n.Value)
			return n.Value, nil
		}

	case NumberLiteralNode:
		i.Logger().Debug("[DEBUG EVAL]   Resolved Number Literal: %v (%T)", n.Value, n.Value)
		return n.Value, nil // Value is already int64 or float64
	case BooleanLiteralNode:
		i.Logger().Debug("[DEBUG EVAL]   Resolved Boolean Literal: %v", n.Value)
		return n.Value, nil
	case ListLiteralNode:
		// Evaluate elements within the list *if necessary*?
		// Current assumption: list elements are already evaluated by the time they reach here,
		// or evaluateExpression handles ListLiteralNode directly.
		// If evaluateExpression calls resolveValue *on the ListLiteralNode itself*,
		// then resolveValue should return the node's Elements slice.
		i.Logger().Debug("[DEBUG EVAL]   Resolved List Literal (returning Elements slice)")
		// Let's assume evaluateExpression handles evaluating elements.
		// This function just returns the literal value structure.
		return n.Elements, nil // Return the slice of (potentially unevaluated) elements
	case MapLiteralNode:
		// Similar to ListLiteralNode, assume evaluateExpression handles evaluating keys/values.
		i.Logger().Debug("[DEBUG EVAL]   Resolved Map Literal (returning Entries slice)")
		return n.Entries, nil // Return the slice of (potentially unevaluated) entries

	// --- Other Node Types ---
	// If evaluateExpression calls resolveValue for other node types that aren't simple values
	// (like BinaryOpNode, FunctionCallNode), resolveValue might need to return them as-is
	// or return an error indicating they should be handled by evaluateExpression directly.
	// For now, assume evaluateExpression handles complex nodes directly.

	default:
		// This case handles values that are *already resolved* (e.g., results from previous operations)
		// being passed back into evaluation. Just return them.
		i.Logger().Debug("[DEBUG EVAL]   Node is already a resolved value: %v (%T)", node, node)
		return node, nil
		// Alternatively, if only specific AST nodes should be handled here:
		// return nil, fmt.Errorf("internal error: resolveValue received unexpected node type %T", node)
	}
}
