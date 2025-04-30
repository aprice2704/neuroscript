// filename: pkg/core/evaluation_resolve.go
package core

import (
	"fmt"
	"regexp" // Import regexp for placeholder parsing
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

// --- Helper Function for Placeholder Substitution ---

// placeholderRegex finds {{variable_name}} occurrences. It captures the variable_name.
// Using \s* to allow optional whitespace inside the braces, e.g., {{ my_var }}
var placeholderRegex = regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}`)

// FIX: Renamed function to match calls in evaluation_main.go
// resolvePlaceholdersWithError processes a raw string, finding and replacing {{placeholders}}.
func (i *Interpreter) resolvePlaceholdersWithError(rawString string) (string, error) {
	var firstError error
	processedString := placeholderRegex.ReplaceAllStringFunc(rawString, func(match string) string {
		// If an error has already occurred, skip further processing for this string
		if firstError != nil {
			return match // Return the original placeholder text
		}

		// Extract variable name from the match (group 1)
		groups := placeholderRegex.FindStringSubmatch(match)
		if len(groups) < 2 {
			// This should not happen with the defined regex, but handle defensively
			i.Logger().Error("[ERROR EVAL] Regex match '%s' failed to capture group", match)
			firstError = fmt.Errorf("internal regex error processing placeholder '%s'", match)
			return match // Return original on internal error
		}
		varName := groups[1]

		// Look up the variable
		value, found := i.GetVariable(varName)
		if !found {
			errMsg := fmt.Sprintf("variable '%s' referenced in placeholder not found", varName)
			firstError = fmt.Errorf("%s: %w", errMsg, ErrVariableNotFound)
			return match // Return original if variable not found
		}

		// Convert the found value to string
		// Use fmt.Sprintf("%v") for general compatibility. Adjust if specific formatting is needed.
		strValue := fmt.Sprintf("%v", value)
		i.Logger().Debug("[DEBUG EVAL]     Substituting placeholder match '%s' with value: %q", match, strValue)

		return strValue // Return the substituted value
	})

	if firstError != nil {
		return "", firstError // Return only the first error encountered
	}

	return processedString, nil
}

// --- Placeholder Implementations (ensure these exist elsewhere) ---
// func (i *Interpreter) GetVariable(name string) (interface{}, bool)
// func (i *Interpreter) Logger() logging.Logger { ... }
// var ErrVariableNotFound = errors.New(...) // Assumed defined in errors.go
// AST Node definitions (VariableNode, StringLiteralNode, etc.) assumed in ast.go
