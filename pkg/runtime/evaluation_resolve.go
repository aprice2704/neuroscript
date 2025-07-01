// filename: pkg/core/evaluation_resolve.go
package runtime

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	// Assuming errors like ErrVariableNotFound are defined in errors.go
	// Assuming AST node types like ast.VariableNode, ast.Placeholder.Node, ast.StringLiteralNode etc. are defined in ast.go
)

// --- Value Resolution ---

// resolveValue handles resolving variable names, placeholders, and literals to their actual values.
// It's called by evaluate.Expression.
// Returns the resolved value or an error (e.g., variable not found).
func (i *Interpreter) resolveValue(node interface{}) (interface{}, error) {
	i.Logger().Debug("[DEBUG EVAL] Resolving value for node type: %T", node)

	switch n := node.(type) {
	case ast.VariableNode:
		val, found := i.GetVariable(n.Name)
		if !found {
			// Wrap ErrVariableNotFound
			errMsg := fmt.Sprintf("variable '%s' not found", n.Name)
			return nil, fmt.Errorf("%s: %w", errMsg, ErrVariableNotFound)
		}
		i.Logger().Debug("[DEBUG EVAL]   Resolved Variable '%s' to: %v (%T)", n.Name, val, val)
		return val, nil

	case ast.Placeholder.Node: // Assuming ast.Placeholder.Node is distinct, otherwise handled by ast.VariableNode
		// This might be used if placeholders have different lookup rules than variables.
		// For now, treat like ast.VariableNode.
		val, found := i.GetVariable(n.Name)
		if !found {
			// Wrap ErrVariableNotFound
			errMsg := fmt.Sprintf("placeholder variable '%s' not found", n.Name)
			return nil, fmt.Errorf("%s: %w", errMsg, ErrVariableNotFound)
		}
		i.Logger().Debug("[DEBUG EVAL]   Resolved Placeholder '%s' to: %v (%T)", n.Name, val, val)
		return val, nil

	case ast.EvalNode:
		// Return the result stored by the interpreter from the *previous step*
		// Use i.lastCallResult as determined previously
		i.Logger().Debug("[DEBUG EVAL]   Resolved LAST to: %v (%T)", i.lastCallResult, i.lastCallResult)
		return i.lastCallResult, nil

	// --- Literals ---
	case ast.StringLiteralNode:
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

	case ast.NumberLiteralNode:
		i.Logger().Debug("[DEBUG EVAL]   Resolved Number Literal: %v (%T)", n.Value, n.Value)
		return n.Value, nil // Value is already int64 or float64
	case ast.BooleanLiteralNode:
		i.Logger().Debug("[DEBUG EVAL]   Resolved Boolean Literal: %v", n.Value)
		return n.Value, nil
	case ast.ListLiteralNode:
		// Evaluate elements within the list *if necessary*?
		// Current assumption: list elements are already evaluated by the time they reach here,
		// or evaluate.Expression handles ast.ListLiteralNode directly.
		// If evaluate.Expression calls resolveValue *on the ast.ListLiteralNode itself*,
		// then resolveValue should return the node's Elements slice.
		i.Logger().Debug("[DEBUG EVAL]   Resolved List Literal (returning Elements slice)")
		// Let's assume evaluate.Expression handles evaluating elements.
		// This function just returns the literal value structure.
		return n.Elements, nil // Return the slice of (potentially unevaluated) elements
	case ast.MapLiteralNode:
		// Similar to ast.ListLiteralNode, assume evaluate.Expression handles evaluating keys/values.
		i.Logger().Debug("[DEBUG EVAL]   Resolved Map Literal (returning Entries slice)")
		return n.Entries, nil // Return the slice of (potentially unevaluated) entries

	// --- Other Node Types ---
	// If evaluate.Expression calls resolveValue for other node types that aren't simple values
	// (like ast.BinaryOpNode, FunctionCallNode), resolveValue might need to return them as-is
	// or return an error indicating they should be handled by evaluate.Expression directly.
	// For now, assume evaluate.Expression handles complex nodes directly.

	default:
		// This case handles values that are *already resolved* (e.g., results from previous operations)
		// being passed back into evaluation. Just return them.
		i.Logger().Debug("[DEBUG EVAL]   Node is already a resolved value: %v (%T)", node, node)
		return node, nil
		// Alternatively, if only specific AST nodes should be handled here:
		// return nil, fmt.Errorf("internal error: resolveValue received unexpected node type %T", node)
	}
}
