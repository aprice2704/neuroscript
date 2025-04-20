// filename: pkg/core/interpreter_simple_steps.go
package core

import (
	"context"
	"errors"
	"fmt"
	"strings"
	// Added for logging if needed, ensure it's present
)

// --- Simple Statement Execution Helpers ---

// executeSet evaluates the expression node and assigns the result
func (i *Interpreter) executeSet(step Step, stepNum int) error {
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]      Executing SET for variable '%s'", step.Target)
	}
	targetVar := step.Target
	valueNode := step.Value

	if !isValidIdentifier(targetVar) {
		return fmt.Errorf("SET target '%s' is not a valid variable name", targetVar)
	}

	// Evaluate expression and check for errors
	finalValue, err := i.evaluateExpression(valueNode) // Use depth 0 for top-level call
	if err != nil {
		return fmt.Errorf("evaluating value for SET '%s': %w", targetVar, err) // Propagate error
	}

	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]        SET evaluated value: %v (%T)", finalValue, finalValue)
	}

	// Special handling for 'generated_code' remains
	if targetVar == "generated_code" {
		if finalStr, isStr := finalValue.(string); isStr {
			trimmedVal := trimCodeFences(finalStr) // Use helper
			if trimmedVal != finalStr {
				if i.logger != nil {
					i.logger.Printf("[DEBUG-INTERP]        Trimmed code fences for 'generated_code'")
				}
				finalValue = trimmedVal
			}
		}
	}

	i.variables[targetVar] = finalValue
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]        Stored var '%s' = %v (%T)", targetVar, finalValue, finalValue)
	}
	return nil // Return nil error on success
}

// executeCall evaluates argument nodes and performs the call
func (i *Interpreter) executeCall(step Step, stepNum int) (interface{}, error) {
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]      Executing CALL %q", step.Target)
	}
	target := step.Target
	argNodes := step.Args

	evaluatedArgs := make([]interface{}, len(argNodes))
	var err error // Declare err here for use in the loop
	for idx, argNode := range argNodes {
		// Evaluate each argument and check for errors
		evaluatedArgs[idx], err = i.evaluateExpression(argNode) // Use depth 0 for top-level call
		if err != nil {
			return nil, fmt.Errorf("evaluating argument %d for CALL %q: %w", idx, target, err) // Propagate error
		}
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        CALL Arg %d evaluated: %v (%T)", idx, evaluatedArgs[idx], evaluatedArgs[idx])
		}
	}

	var callResultValue interface{}
	var callErr error

	if strings.HasPrefix(target, "TOOL.") {
		toolName := strings.TrimPrefix(target, "TOOL.")
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        Calling TOOL.%s", toolName)
		}
		toolImpl, found := i.toolRegistry.GetTool(toolName)
		if !found {
			callErr = fmt.Errorf("unknown TOOL '%s'", toolName) // Set error
		} else {
			preparedArgs, validationErr := ValidateAndConvertArgs(toolImpl.Spec, evaluatedArgs)
			if validationErr != nil {
				callErr = fmt.Errorf("TOOL %s argument error: %w", toolName, validationErr) // Set error
			} else {
				if i.logger != nil {
					i.logger.Printf("[DEBUG-INTERP]          Prepared TOOL args: %+v", preparedArgs)
				}
				// Execute the tool function
				callResultValue, callErr = toolImpl.Func(i, preparedArgs) // Captures potential tool execution error
				if callErr == nil {                                       // Only log/set last result on SUCCESS
					if i.logger != nil {
						i.logger.Printf("[DEBUG-INTERP]          TOOL.%s Result: %v (%T)", toolName, callResultValue, callResultValue)
					}
					i.lastCallResult = callResultValue // Store successful result
				} else {
					// Propagate error from tool execution
					callErr = fmt.Errorf("TOOL %s execution failed: %w", toolName, callErr)
				}
			}
		}

	} else if target == "LLM" {
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        Calling LLM (stateless)")
		}
		if len(evaluatedArgs) != 1 {
			callErr = errors.New("CALL LLM expects 1 prompt arg") // Set error
		} else {
			prompt := fmt.Sprintf("%v", evaluatedArgs[0]) // Convert evaluated arg to string
			if i.logger != nil {
				i.logger.Printf("[DEBUG-INTERP]          LLM Prompt: %q", prompt)
			}

			// *** CORRECTED: Provide model name (using interpreter's or default) and use i.Logger() ***
			modelToUse := i.modelName // Use interpreter's configured model if set
			if modelToUse == "" {
				modelToUse = "gemini-1.5-pro-latest" // Fallback if not set on interpreter
				if i.logger != nil {
					i.logger.Printf("[WARN INTERP] Interpreter modelName not set for LLM call, using default: %s", modelToUse)
				}
			}
			// Create a temporary client using interpreter's logger
			llmClient := NewLLMClient("", modelToUse, i.Logger())
			if llmClient.client == nil { // Check if client init failed
				callErr = errors.New("failed to initialize LLM client for CALL LLM")
			} else {
				ctx := context.Background()
				response, llmErr := llmClient.CallLLM(ctx, prompt) // Use stateless call
				if llmErr != nil {
					callErr = fmt.Errorf("CALL LLM failed: %w", llmErr) // Propagate LLM error
				} else {
					callResultValue = response
					callErr = nil // Explicitly nil on success
					if i.logger != nil {
						i.logger.Printf("[DEBUG-INTERP]          LLM Result: %q", response)
					}
					i.lastCallResult = callResultValue // Store successful result
				}
			}
			// *** End Correction ***
		}

	} else { // Procedure Call
		procToCall := target
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        Calling Procedure %q", procToCall)
		}

		// Convert evaluated arguments to strings for procedure call signature
		stringArgs := make([]string, len(evaluatedArgs))
		for idx, val := range evaluatedArgs {
			stringArgs[idx] = fmt.Sprintf("%v", val)
		}
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]          Procedure args (as strings): %v", stringArgs)
		}

		// Recursively call RunProcedure
		procResultValue, procCallErr := i.RunProcedure(procToCall, stringArgs...)
		if procCallErr != nil {
			// Propagate error from nested procedure call
			// Error context is already added within RunProcedure/executeSteps
			callErr = procCallErr
		} else {
			callResultValue = procResultValue
			callErr = nil // Explicitly nil on success
			if i.logger != nil {
				i.logger.Printf("[DEBUG-INTERP]          Procedure %q Result: %v (%T)", procToCall, callResultValue, callResultValue)
			}
			i.lastCallResult = callResultValue // Store successful result
		}
	}

	// Return the captured result (if any) and the error status from the call
	return callResultValue, callErr
}

// executeReturn handles the RETURN statement, now checking expression evaluation error
func (i *Interpreter) executeReturn(step Step, stepNum int) (result interface{}, wasReturn bool, err error) {
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]      Executing RETURN")
	}
	valueNode := step.Value
	var returnValue interface{}
	var evalErr error

	if valueNode != nil {
		// Evaluate return value expression and check for errors
		returnValue, evalErr = i.evaluateExpression(valueNode) // Depth 0
		if evalErr != nil {
			// Return error immediately if return value evaluation fails
			return nil, true, fmt.Errorf("evaluating RETURN value: %w", evalErr) // Still signal RETURN intent
		}
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        RETURN evaluated value: %v (%T)", returnValue, returnValue)
		}
	} else {
		returnValue = nil // No value provided
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        RETURN with no value (implicit nil)")
		}
	}

	// Signal RETURN occurred, return the value (or nil) and nil error
	return returnValue, true, nil
}

// executeEmit handles the EMIT statement, now checking expression evaluation error
func (i *Interpreter) executeEmit(step Step, stepNum int) error {
	if i.logger != nil {
		i.logger.Printf("[DEBUG-INTERP]      Executing EMIT")
	}
	valueNode := step.Value
	var evaluatedValue interface{}
	var evalErr error

	if valueNode != nil {
		// Evaluate emit value expression and check for errors
		evaluatedValue, evalErr = i.evaluateExpression(valueNode) // Depth 0
		if evalErr != nil {
			return fmt.Errorf("evaluating EMIT value: %w", evalErr) // Return error
		}
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        EMIT evaluated value: %v (%T)", evaluatedValue, evaluatedValue)
		}
	} else {
		evaluatedValue = "" // Default to empty string if no value node
		if i.logger != nil {
			i.logger.Printf("[DEBUG-INTERP]        EMIT with no value (implicit empty string)")
		}
	}

	fmt.Printf("[EMIT] %v\n", evaluatedValue) // Direct print for EMIT

	return nil // Return nil error on success
}
