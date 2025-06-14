// NeuroScript Version: 0.4.1
// File version: 10
// Purpose: Refactored to expect primitives, per the value-wrapping contract. Replaced generic validator with specific funcs.
// nlines: 71
// risk_rating: HIGH
// filename: pkg/core/tools_validation.go

package core

import (
	"fmt"
)

// --- Tool Argument Convention (NEW) ---
//
// As per the value-wrapping contract, all validation functions operate on raw
// Go primitives (e.g., []any, string, float64). They are the boundary between
// the adapter/bridge layer and the tool implementation. They should NEVER import
// the 'core' package or handle 'core.Value' wrapper types.

// validateListLength checks if args are valid for the 'List.Length' tool.
// It expects: [ list ]
func validateListLength(args []any) error {
	if len(args) != 1 {
		return fmt.Errorf("%w: tool 'List.Length' expected 1 argument, got %d", ErrValidationArgCount, len(args))
	}
	if args[0] == nil {
		return fmt.Errorf("%w: argument 'list' cannot be nil", ErrValidationRequiredArgNil)
	}
	if _, ok := args[0].([]any); !ok {
		return fmt.Errorf("%w: expected argument 'list' to be a list, but got %T", ErrValidationTypeMismatch, args[0])
	}
	return nil
}

// validateListAppend checks if args are valid for the 'List.Append' tool.
// It expects: [ list, element ]
func validateListAppend(args []any) error {
	if len(args) != 2 {
		return fmt.Errorf("%w: tool 'List.Append' expected 2 arguments, got %d", ErrValidationArgCount, len(args))
	}
	if args[0] == nil {
		return fmt.Errorf("%w: argument 'list' cannot be nil", ErrValidationRequiredArgNil)
	}
	if _, ok := args[0].([]any); !ok {
		return fmt.Errorf("%w: expected argument 'list' to be a list, but got %T", ErrValidationTypeMismatch, args[0])
	}
	// The element to append (args[1]) can be any type, so no further validation is needed here.
	return nil
}

// validateListGet checks if args are valid for the 'List.Get' tool.
// It expects: [ list, index, (optional) default ]
func validateListGet(args []any) error {
	if len(args) < 2 || len(args) > 3 {
		return fmt.Errorf("%w: tool 'List.Get' expected 2 or 3 arguments, got %d", ErrValidationArgCount, len(args))
	}
	if args[0] == nil {
		return fmt.Errorf("%w: argument 'list' cannot be nil", ErrValidationRequiredArgNil)
	}
	if _, ok := args[0].([]any); !ok {
		return fmt.Errorf("%w: expected argument 'list' to be a list, but got %T", ErrValidationTypeMismatch, args[0])
	}
	if args[1] == nil {
		return fmt.Errorf("%w: argument 'index' cannot be nil", ErrValidationRequiredArgNil)
	}
	// We could check if index is a number, but coercion can handle that.
	// This layer is for basic shape validation.
	return nil
}
