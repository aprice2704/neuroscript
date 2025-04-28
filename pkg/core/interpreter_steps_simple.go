// filename: pkg/core/interpreter_steps_simple.go
package core

import (
	"errors"
	"fmt"
	"strings"
	// Assuming NsError, RuntimeError, error codes, and ArgType constants/types
	// are defined in errors.go and tools_types.go
	// Assuming ValidateAndConvertArgs is defined in tools_validation.go
)

// executeSet handles the "set" step.
// It returns the value that was set, so it can potentially update the interpreter's last result state.
func (i *Interpreter) executeSet(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	i.Logger().Debug("[DEBUG-INTERP]   Executing SET: Target=%s", step.Target)
	value, err := i.evaluateExpression(step.Value) // Pass context flags if evaluateExpression needs them
	if err != nil {
		// Wrap underlying evaluation error
		return nil, NewRuntimeError(ErrorCodeGeneric, fmt.Sprintf("evaluating value for SET %s", step.Target), fmt.Errorf("evaluating value for set %s: %w", step.Target, err))
	}
	if isInHandler && (step.Target == "err_code" || step.Target == "err_msg") {
		// Wrap specific sentinel error (ErrReadOnlyViolation now defined in errors.go)
		errMsg := fmt.Sprintf("cannot assign to read-only variable '%s' within on_error handler", step.Target)
		return nil, NewRuntimeError(ErrorCodeReadOnly, errMsg, fmt.Errorf("%s: %w", errMsg, ErrReadOnlyViolation))
	}
	setErr := i.SetVariable(step.Target, value)
	if setErr != nil {
		// Wrap internal error
		return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("setting variable '%s'", step.Target), fmt.Errorf("setting variable %s: %w", step.Target, setErr))
	}
	// Return the value being set so it can become the "last result" in executeSteps
	return value, nil
}

// executeCall handles the "call" step (procedure or tool).
// It returns the result of the call.
func (i *Interpreter) executeCall(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	i.Logger().Debug("[DEBUG-INTERP]   Executing CALL: Target=%s", step.Target)

	evaluatedArgs := make([]interface{}, len(step.Args))
	for idx, arg := range step.Args {
		evaluatedArg, err := i.evaluateExpression(arg) // Pass context flags if evaluateExpression needs them
		if err != nil {
			// Wrap underlying evaluation error
			return nil, NewRuntimeError(ErrorCodeGeneric, fmt.Sprintf("evaluating arg %d for CALL %s", idx+1, step.Target), fmt.Errorf("evaluating arg %d for call %s: %w", idx+1, step.Target, err))
		}
		evaluatedArgs[idx] = evaluatedArg
	}

	// --- Actual call logic ---
	if strings.HasPrefix(step.Target, "tool.") {
		// --- Tool Call Logic ---
		parts := strings.SplitN(step.Target, ".", 2) // Split only on the first dot
		if len(parts) != 2 || parts[0] != "tool" || parts[1] == "" {
			errMsg := fmt.Sprintf("invalid tool call format: %s (must be tool.FunctionName)", step.Target)
			// Wrap generic error sentinel (ErrUnsupportedSyntax now defined in errors.go)
			return nil, NewRuntimeError(ErrorCodeSyntax, errMsg, fmt.Errorf("%s: %w", errMsg, ErrUnsupportedSyntax)) // Used ErrorCodeSyntax
		}
		funcName := parts[1]

		toolImpl, found := i.ToolRegistry().GetTool(funcName)
		if !found {
			errMsg := fmt.Sprintf("tool '%s' not found in registry", funcName)
			// Wrap specific sentinel error (ErrToolNotFound confirmed to exist in errors.go)
			return nil, NewRuntimeError(ErrorCodeToolNotFound, errMsg, fmt.Errorf("%s: %w", errMsg, ErrToolNotFound))
		}

		// --- Argument Validation & Conversion ---
		validatedAndConvertedArgs, validationErr := ValidateAndConvertArgs(toolImpl.Spec, evaluatedArgs)
		if validationErr != nil {
			code := ErrorCodeArgMismatch
			// Check for specific validation errors using errors.Is if needed (e.g., ErrValidationTypeMismatch)
			if errors.Is(validationErr, ErrValidationTypeMismatch) {
				code = ErrorCodeType // More specific code?
			} else if errors.Is(validationErr, ErrValidationArgCount) {
				code = ErrorCodeArgMismatch
			}
			// Use validationErr as the wrapped error directly. Add context.
			return nil, NewRuntimeError(code, fmt.Sprintf("argument validation failed for tool '%s'", funcName), fmt.Errorf("validating args for %s: %w", funcName, validationErr))
		}
		// --- End Argument Validation ---

		i.Logger().Debug("[DEBUG-INTERP]     Executing Tool '%s'...", funcName)
		// Execute the tool's function with VALIDATED and CONVERTED args
		toolResult, toolErr := toolImpl.Func(i, validatedAndConvertedArgs)

		// The interpreter loop (executeSteps) will handle setting the last result state.

		if toolErr != nil {
			// Ensure it's a RuntimeError, wrap if not
			if re, ok := toolErr.(*RuntimeError); ok {
				return nil, re // Already a RuntimeError
			}
			code := ErrorCodeToolSpecific // Default code for unwrapped tool errors
			// Optionally check for specific Go errors from the tool if needed
			// Wrap the original toolErr
			return nil, NewRuntimeError(code, fmt.Sprintf("tool '%s' execution failed", funcName), fmt.Errorf("executing tool %s: %w", funcName, toolErr))
		}
		i.Logger().Debug("[DEBUG-INTERP]     Tool '%s' execution successful.", funcName)
		return toolResult, nil
		// --- End Tool Call Logic ---

	} else {
		// Internal procedure call
		i.Logger().Debug("[DEBUG-INTERP]     Calling Internal Procedure '%s'...", step.Target)
		procResult, procErr := i.RunProcedure(step.Target, evaluatedArgs...)
		// The interpreter loop (executeSteps) will handle setting the last result state.
		if procErr != nil {
			// Ensure it's a RuntimeError, wrap if not
			if re, ok := procErr.(*RuntimeError); ok {
				return nil, re // Already a RuntimeError
			}
			code := ErrorCodeGeneric
			wrapped := procErr // Default wrap
			errMsg := procErr.Error()

			if errors.Is(procErr, ErrProcedureNotFound) {
				code = ErrorCodeProcNotFound
				wrapped = ErrProcedureNotFound // Use sentinel
				errMsg = fmt.Sprintf("procedure '%s' not found", step.Target)
			} else if errors.Is(procErr, ErrArgumentMismatch) {
				code = ErrorCodeArgMismatch
				wrapped = ErrArgumentMismatch // Use sentinel
				errMsg = fmt.Sprintf("argument mismatch calling procedure '%s'", step.Target)
			}
			// Wrap the error with context
			return nil, NewRuntimeError(code, errMsg, fmt.Errorf("calling procedure %s: %w", step.Target, wrapped))
		}
		i.Logger().Debug("[DEBUG-INTERP]     Procedure '%s' call successful.", step.Target)
		return procResult, nil
	}
}

// executeReturn handles the "return" step.
// It evaluates the return value(s) and signals the calling loop (executeSteps) that a return occurred.
func (i *Interpreter) executeReturn(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, bool, error) {
	// Note: isInHandler check is done in executeSteps before calling this.
	i.Logger().Debug("[DEBUG-INTERP]   Executing RETURN")
	rawValue := step.Value // This is usually []interface{} or nil

	if rawValue == nil {
		i.Logger().Debug("[DEBUG-INTERP]     Return has no value (implicit nil)")
		return nil, true, nil // Return nil value, signal return=true, no error
	}

	// Expect rawValue to be a slice of expression nodes from the AST builder
	if exprSlice, ok := rawValue.([]interface{}); ok {
		i.Logger().Debug("[DEBUG-INTERP]     Return has %d expression(s)", len(exprSlice))
		if len(exprSlice) == 0 {
			i.Logger().Debug("[DEBUG-INTERP]     Return has empty expression list (equivalent to nil)")
			return nil, true, nil
		}

		results := make([]interface{}, len(exprSlice))
		for idx, exprNode := range exprSlice {
			evaluatedValue, err := i.evaluateExpression(exprNode) // Pass context flags if evaluateExpression needs them
			if err != nil {
				// Wrap evaluation error
				return nil, true, NewRuntimeError(ErrorCodeGeneric, fmt.Sprintf("evaluating return expression %d", idx+1), fmt.Errorf("evaluating return expression %d: %w", idx+1, err))
			}
			results[idx] = evaluatedValue
		}

		// Return the slice as is. The RunProcedure caller or assignment logic will handle single vs multiple values.
		return results, true, nil // Return evaluated results slice, signal return=true, no error

	} else {
		// This case should ideally not happen if the AST builder always creates a slice for RETURN values.
		i.Logger().Error("[ERROR INTERP] RETURN step value was not a slice: %T", rawValue)
		// Return an internal error. Assume ErrInternal is defined.
		errMsg := fmt.Sprintf("internal error: RETURN step value was not []interface{}, but %T", rawValue)
		return nil, true, NewRuntimeError(ErrorCodeInternal, errMsg, fmt.Errorf("%s: %w", errMsg, ErrInternal))
	}
}

// executeEmit handles the "emit" step.
// It evaluates the value, calls the default print mechanism, and returns the emitted value.
func (i *Interpreter) executeEmit(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	i.Logger().Debug("[DEBUG-INTERP]   Executing EMIT")
	value, err := i.evaluateExpression(step.Value) // Pass context flags if evaluateExpression needs them
	if err != nil {
		// Wrap evaluation error
		return nil, NewRuntimeError(ErrorCodeGeneric, "evaluating value for EMIT", fmt.Errorf("evaluating emit value: %w", err))
	}
	// Removed undefined i.emitCallback. Use default print.
	fmt.Printf("EMIT: %v\n", value) // Consider using logger if available everywhere?

	// Return the emitted value so it can become the "last result"
	return value, nil
}

// executeMust handles "must" and "mustbe" steps.
// It evaluates the condition. If false or evaluation errors, returns a specific MUST error.
// If true, it returns the evaluated condition value.
func (i *Interpreter) executeMust(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (interface{}, error) {
	stepType := strings.ToLower(step.Type) // must or mustbe
	i.Logger().Debug("[DEBUG-INTERP]   Executing %s", strings.ToUpper(stepType))
	value, err := i.evaluateExpression(step.Value) // Pass context flags if evaluateExpression needs them

	// Consistently handle evaluation errors by wrapping ErrMustConditionFailed
	if err != nil {
		errMsg := fmt.Sprintf("error evaluating condition for %s", stepType)
		// Use ErrorCodeMustFailed and wrap ErrMustConditionFailed, using the original evaluation error for context.
		// Add specific detail about the evaluation error itself.
		wrappedErr := fmt.Errorf("%w: evaluation failed (%v)", ErrMustConditionFailed, err)
		return nil, NewRuntimeError(ErrorCodeMustFailed, errMsg, wrappedErr)
	}

	// Check truthiness of the successfully evaluated value
	if !isTruthy(value) {
		errMsg := ""
		// Try to reconstruct a meaningful message from the AST node if possible
		nodeStr := NodeToString(step.Value) // Assumes NodeToString helper exists

		if stepType == "mustbe" {
			errMsg = fmt.Sprintf("'mustbe %s' evaluated to false", nodeStr)
		} else { // "must"
			errMsg = fmt.Sprintf("'must %s' condition evaluated to false", nodeStr)
		}
		// Use ErrorCodeMustFailed and wrap ErrMustConditionFailed
		return nil, NewRuntimeError(ErrorCodeMustFailed, errMsg, ErrMustConditionFailed)
	}

	i.Logger().Debug("[DEBUG-INTERP]     %s condition TRUE.", strings.ToUpper(stepType))
	// Return the successfully evaluated condition value so it can become the "last result"
	return value, nil
}

// executeFail handles the "fail" step.
// It evaluates the optional error message/code and returns a specific FAIL error.
// It does not return a value for the "last result".
func (i *Interpreter) executeFail(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) error {
	i.Logger().Debug("[DEBUG-INTERP]   Executing FAIL")
	errCode := ErrorCodeFailStatement
	errMsg := "fail statement executed"
	var evalErr error = nil                 // To store any error during evaluation of the fail value
	var wrappedErr error = ErrFailStatement // The sentinel error for FAIL itself

	if step.Value != nil {
		failValue, err := i.evaluateExpression(step.Value) // Pass context flags if evaluateExpression needs them
		if err != nil {
			// If evaluation fails, note the error, but the primary error is still the FAIL.
			evalErr = err
			errMsg = fmt.Sprintf("fail statement executed (error evaluating message/code: %v)", err)
			// Keep errCode as ErrorCodeFailStatement, the intention was to fail.
		} else {
			// If evaluation succeeds, use the value to customize msg/code
			switch v := failValue.(type) {
			case string:
				errMsg = v // Use the string as the message
			case int:
				errCode = v // Use the int as the code
			case int64:
				errCode = int(v) // Use the int64 as the code
			case float64:
				// Disallow float codes? Truncate for now.
				errCode = int(v)
				errMsg = fmt.Sprintf("fail statement executed with code %d (from float %v)", errCode, v)
			default:
				errMsg = fmt.Sprintf("fail statement executed with value: %v", failValue)
			}
		}
	}
	// Create the final RuntimeError. Wrap the sentinel ErrFailStatement.
	// Include the evaluation error (if any) in the message for debugging.
	finalErrMsg := errMsg
	if evalErr != nil {
		finalErrMsg = fmt.Sprintf("%s [evaluation error: %v]", errMsg, evalErr)
	}
	return NewRuntimeError(errCode, finalErrMsg, wrappedErr)
}

// executeOnError handles the "on_error" step setup.
// It returns the handler step details for the interpreter loop.
// It does not return a value for the "last result".
func (i *Interpreter) executeOnError(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (*Step, error) {
	i.Logger().Debug("[DEBUG-INTERP]   Executing ON_ERROR - Handler now active for subsequent steps in this scope.")
	// This step only sets up the handler; the actual execution happens in the main loop.
	// It doesn't produce a result value itself.
	handlerStep := step
	return &handlerStep, nil // Return the handler step itself for the main loop to register
}

// executeClearError handles the "clear_error" step.
// It signals the interpreter loop to clear the active error state.
// It does not return a value for the "last result".
func (i *Interpreter) executeClearError(step Step, stepNum int, isInHandler bool, activeError *RuntimeError) (bool, error) {
	i.Logger().Debug("[DEBUG-INTERP]   Executing CLEAR_ERROR")
	if !isInHandler {
		// Wrap specific sentinel error (ErrClearViolation now defined in errors.go)
		errMsg := fmt.Sprintf("step %d: 'clear_error' can only be used inside an on_error block", stepNum+1)
		return false, NewRuntimeError(ErrorCodeClearViolation, errMsg, fmt.Errorf("%s: %w", errMsg, ErrClearViolation))
	}
	// This step signals to the main loop to clear the error; it doesn't produce a result value.
	return true, nil // Signal clear was called
}

// --- Placeholders / Assumed Helpers ---

// evaluateExpression needs to be defined (e.g., in evaluation_main.go)
// func (i *Interpreter) evaluateExpression(node interface{}) (interface{}, error) { ... }

// isTruthy needs to be defined (e.g., in evaluation_helpers.go)
// func isTruthy(value interface{}) bool { ... }

// ValidateAndConvertArgs is assumed defined in tools_validation.go
// func ValidateAndConvertArgs(spec ToolSpec, rawArgs []interface{}) ([]interface{}, error) { ... }

// NodeToString converts an AST node to a string representation for error messages.
// Placeholder implementation - needs actual AST node types.
func NodeToString(node interface{}) string {
	// Basic fallback using fmt
	str := fmt.Sprintf("%v", node)
	// Truncate long representations for brevity in error messages
	maxLen := 50
	if len(str) > maxLen {
		str = str[:maxLen-3] + "..."
	}
	return str
}

// Removed placeholder definitions for errors - they are now in errors.go
// Removed duplicate FunctionCallNode definition - it's defined in ast.go
