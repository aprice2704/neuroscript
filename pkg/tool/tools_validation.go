// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Removes variadic argument logic, enforcing fixed argument counts.
// filename: pkg/tool/tools_validation.go
// nlines: 88
// risk_rating: MEDIUM

package tool

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// validateAndCoerceArgs checks raw arguments against a tool's specification,
// performs type coercion, and handles required/optional rules.
// It returns the coerced arguments suitable for passing to the tool's Go function,
// or a combined error detailing all validation failures.
func validateAndCoerceArgs(fullName types.FullName, rawArgs []any, spec ToolSpec) ([]any, error) {
	numRawArgs := len(rawArgs)
	numSpecArgs := len(spec.Args)
	minRequiredArgs := 0
	for _, argSpec := range spec.Args {
		if argSpec.Required {
			minRequiredArgs++
		}
	}

	// 1. Check overall argument count (NO VARIADIC)
	if numRawArgs < minRequiredArgs || numRawArgs > numSpecArgs {
		var expected string
		if minRequiredArgs == numSpecArgs {
			expected = fmt.Sprintf("%d", minRequiredArgs)
		} else {
			expected = fmt.Sprintf("%d to %d", minRequiredArgs, numSpecArgs)
		}
		err := lang.NewRuntimeError(lang.ErrorCodeArgMismatch,
			fmt.Sprintf("tool '%s': expected %s arguments, got %d", fullName, expected, numRawArgs),
			lang.ErrArgumentMismatch) // Use sentinel error
		return nil, err
	}

	// 2. Validate and coerce individual arguments specified in the spec
	coercedArgs := make([]any, numSpecArgs) // Size to the exact number of defined args
	var validationErrors []string

	for i, argSpec := range spec.Args {
		var rawArg any
		if i < numRawArgs {
			rawArg = rawArgs[i] // Argument provided by caller
		} else {
			rawArg = nil // Argument not provided by caller (must be optional)
		}

		// Check if required argument is missing (nil or not provided)
		if argSpec.Required && rawArg == nil {
			validationErrors = append(validationErrors, fmt.Sprintf("argument '%s' is required but was not provided or was nil", argSpec.Name))
			continue // Skip coercion if required arg is missing
		}

		// Attempt coercion (coerceArg handles nil input correctly for optional args)
		coercedVal, coerceErr := coerceArg(rawArg, argSpec.Type)
		if coerceErr != nil {
			// Prepend context to the coercion error
			validationErrors = append(validationErrors, fmt.Sprintf("argument '%s': %v", argSpec.Name, coerceErr))
		} else {
			coercedArgs[i] = coercedVal
		}
	}

	// 3. Combine errors if any occurred
	if len(validationErrors) > 0 {
		combinedMessage := fmt.Sprintf("tool '%s' argument validation failed: %s", fullName, strings.Join(validationErrors, "; "))
		err := lang.NewRuntimeError(lang.ErrorCodeArgMismatch, combinedMessage, lang.ErrInvalidArgument) // Use sentinel error
		return nil, err
	}

	// 4. Return the coerced arguments (fixed length matching spec)
	return coercedArgs, nil
}
