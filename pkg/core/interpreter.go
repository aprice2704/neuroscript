package core

import (
	"fmt"
	"os" // Keep for potential future use
	"strings"
)

// --- Interpreter ---
type Interpreter struct {
	variables       map[string]interface{}
	knownProcedures map[string]Procedure
	lastCallResult  interface{}
	vectorIndex     map[string][]float32 // Simple in-memory vector store - TODO: Move out?
	embeddingDim    int                  // TODO: Move out?
	currentProcName string               // Added to track current procedure for errors
	toolRegistry    *ToolRegistry        // +++ ADDED +++ Tool registry
}

// NewInterpreter creates a new interpreter instance
func NewInterpreter() *Interpreter {
	interp := &Interpreter{ // Create instance first
		variables:       make(map[string]interface{}),
		knownProcedures: make(map[string]Procedure),
		vectorIndex:     make(map[string][]float32), // Keep mock DB for now
		embeddingDim:    16,                         // Keep mock DB for now
		toolRegistry:    NewToolRegistry(),          // +++ ADDED +++ Initialize registry
	}
	registerCoreTools(interp.toolRegistry) // +++ ADDED +++ Register tools (expects function in tools_register.go)
	return interp
}

// LoadProcedures loads procedures into the interpreter's known map
func (i *Interpreter) LoadProcedures(procs []Procedure) error {
	for _, p := range procs {
		if _, exists := i.knownProcedures[p.Name]; exists {
			fmt.Printf("[Warning] Reloading procedure: %s\n", p.Name)
		}
		i.knownProcedures[p.Name] = p
	}
	return nil
}

// RunProcedure executes a procedure by name with given arguments
func (i *Interpreter) RunProcedure(procName string, args ...string) (result interface{}, err error) { // Named returns
	proc, exists := i.knownProcedures[procName]
	if !exists {
		return nil, fmt.Errorf("procedure '%s' not defined", procName)
	}

	localVars := make(map[string]interface{})
	if len(args) != len(proc.Params) {
		return nil, fmt.Errorf("procedure '%s' expects %d argument(s) (%v), but received %d", procName, len(proc.Params), proc.Params, len(args))
	}
	for idx, paramName := range proc.Params {
		localVars[paramName] = args[idx] // Store initial args as strings
		fmt.Printf("  [Arg Init] SET %s = %q\n", paramName, args[idx])
	}

	fmt.Printf("[Exec] Running Procedure: %s\n", procName)
	originalVars := i.variables           // Save outer scope vars
	originalProcName := i.currentProcName // Save outer proc name context

	// --- Scope Management ---
	i.variables = localVars                             // Set current scope to local
	i.currentProcName = procName                        // Set current proc name context
	var wasReturn bool                                  // Need to capture wasReturn from executeSteps
	result, wasReturn, err = i.executeSteps(proc.Steps) // Execute using local scope
	i.currentProcName = originalProcName                // Restore outer proc name context
	i.variables = originalVars                          // Restore outer scope vars
	// --- End Scope Management ---

	// Log result before returning
	logResult := fmt.Sprintf("%v", result)
	if len(logResult) > 70 {
		logResult = logResult[:67] + "..."
	}
	if err != nil {
		fmt.Printf("[Exec] Procedure '%s' finished with error.\n", procName)
	} else if wasReturn {
		fmt.Printf("[Exec] Procedure '%s' returned explicitly: %q\n", procName, logResult)
	} else {
		// If a procedure doesn't explicitly RETURN, its result is implicitly the lastCallResult (if any)
		// This might need refinement based on desired semantics. For now, let's return nil if no explicit RETURN.
		// result = i.lastCallResult // Alternative: return lastCallResult implicitly
		fmt.Printf("[Exec] Procedure '%s' finished normally. Final Result (Implicit nil): %v\n", procName, result)
	}

	return result, err
}

// --- Step Execution ---
// Propagates return value and signal correctly
func (i *Interpreter) executeSteps(steps []Step) (result interface{}, wasReturn bool, err error) { // Named returns
	skipElse := false
	_, cwdErr := os.Getwd()
	if cwdErr != nil {
		return nil, false, fmt.Errorf("cwd error: %w", cwdErr)
	}

	// Outer loop iterates through the defined steps
	for stepNum, step := range steps {

		// +++ DEBUG LOGGING START +++
		fmt.Printf("\n[DEBUG] === Executing Step %d ===\n", stepNum+1)
		fmt.Printf("[DEBUG] Type:   %s\n", step.Type)
		fmt.Printf("[DEBUG] Target: %s\n", step.Target)
		fmt.Printf("[DEBUG] Cond:   %s\n", step.Cond)
		// Be careful logging Value, could be large ([]Step)
		if body, ok := step.Value.([]Step); ok {
			fmt.Printf("[DEBUG] Value:  (Block with %d steps)\n", len(body))
		} else {
			logValueStr := fmt.Sprintf("%v", step.Value)
			if len(logValueStr) > 100 {
				logValueStr = logValueStr[:97] + "..."
			}
			fmt.Printf("[DEBUG] Value:  %s\n", logValueStr)
		}
		fmt.Printf("[DEBUG] Args:   %v\n", step.Args)
		// +++ DEBUG LOGGING END +++

		// --- Existing logging ---
		logTarget := step.Target
		if step.Type == "IF" || step.Type == "WHILE" || step.Type == "FOR" {
			logTarget = step.Cond
		}
		if step.Type == "RETURN" {
			logTarget = fmt.Sprintf("%v", step.Value)
		}
		if len(logTarget) > 40 {
			logTarget = logTarget[:37] + "..."
		}
		fmt.Printf("  [Step %d] %s %s ...\n", stepNum+1, step.Type, logTarget)
		// --- End Existing logging ---

		if skipElse {
			if step.Type == "ELSE" {
				fmt.Printf("    [Skip] Skipping ELSE block...\n")
				skipElse = false
				continue
			}
			skipElse = false // Reset skipElse if the current step wasn't ELSE
		}

		switch step.Type {
		case "SET":
			targetVar := step.Target
			valueExpr, ok := step.Value.(string)
			if !ok {
				return nil, false, fmt.Errorf("step %d: SET value is not string (type %T)", stepNum+1, step.Value)
			}
			// Check if target is a valid identifier before assignment
			if !isValidIdentifier(targetVar) {
				// Allow internal vars like __last_call_result? No, SET should only be for user vars.
				return nil, false, fmt.Errorf("step %d: SET target '%s' is not a valid variable name", stepNum+1, targetVar)
			}

			finalValue := i.evaluateExpression(valueExpr) // Evaluate the RHS expression

			// Handle special case for generated code fence stripping
			if targetVar == "generated_code" {
				if finalStr, isStr := finalValue.(string); isStr {
					trimmedVal := trimCodeFences(finalStr) // trimCodeFences from interpreter_c.go
					if trimmedVal != finalStr {
						fmt.Printf("      [Fence Strip] Trimmed fences\n")
						finalValue = trimmedVal
					}
				}
			}
			i.variables[targetVar] = finalValue // Assign evaluated value to variable
			logValueStr := fmt.Sprintf("%v", finalValue)
			if len(logValueStr) > 70 {
				logValueStr = logValueStr[:67] + "..."
			}
			fmt.Printf("    [Exec] SET %s = %q\n", targetVar, logValueStr)

		case "CALL":
			i.lastCallResult = nil          // Clear previous result
			target := step.Target           // Target procedure/tool/LLM
			callErr := error(nil)           // Error specifically for this CALL step
			var callResultValue interface{} // Result specifically from this CALL step

			// --- Argument Evaluation ---
			// Evaluate all arguments first, preserving their types in evaluatedArgs
			evaluatedArgs := make([]interface{}, len(step.Args))
			for idx, argExpr := range step.Args {
				evaluatedArgs[idx] = i.evaluateExpression(argExpr) // Evaluate each arg expression
				// fmt.Printf("      [Arg Eval] Arg %d ('%s') evaluated to: %v (Type: %T)\n", idx, argExpr, evaluatedArgs[idx], evaluatedArgs[idx]) // Detailed log if needed
			}
			// --- End Argument Evaluation ---

			// --- Dispatch CALL ---
			if strings.HasPrefix(target, "TOOL.") { // --- TOOL Call ---
				toolName := strings.TrimPrefix(target, "TOOL.")
				fmt.Printf("    [Exec] Attempting CALL TOOL: %s\n", toolName)

				// 1. Lookup tool in registry
				toolImpl, found := i.toolRegistry.GetTool(toolName)
				if !found {
					callErr = fmt.Errorf("unknown TOOL '%s'", toolName)
				} else {
					// 2. Validate and Convert Args using Spec
					// This step modifies evaluatedArgs based on spec.Type!
					preparedArgs, validationErr := ValidateAndConvertArgs(toolImpl.Spec, evaluatedArgs)
					if validationErr != nil {
						callErr = fmt.Errorf("step %d: TOOL %s argument error: %w", stepNum+1, toolName, validationErr)
					} else {
						// 3. Call the registered ToolFunc
						fmt.Printf("      [Tool Dispatch] Calling func for %s with prepared args\n", toolName)
						callResultValue, callErr = toolImpl.Func(i, preparedArgs) // Pass interpreter and prepared args
						if callErr == nil {
							logStr := fmt.Sprintf("%v", callResultValue)
							if len(logStr) > 70 {
								logStr = logStr[:67] + "..."
							}
							fmt.Printf("      [Tool OK] %s result: %s\n", toolName, logStr)
						}
					}
				}
				// Error handling happens after the switch

			} else if target == "LLM" { // --- LLM Call ---
				// LLM expects exactly one string argument
				if len(evaluatedArgs) != 1 {
					callErr = fmt.Errorf("step %d: CALL LLM expects 1 prompt arg, got %d", stepNum+1, len(evaluatedArgs))
				} else {
					// Ensure the argument is a string
					prompt := fmt.Sprintf("%v", evaluatedArgs[0]) // Convert arg to string

					fmt.Printf("    [Exec] CALL LLM prompt (first 80 chars): %q\n", truncateString(prompt, 80))
					response, llmErr := CallLLMAPI(prompt) // CallLLMAPI from llm.go
					if llmErr != nil {
						callErr = fmt.Errorf("step %d: CALL LLM failed: %w", stepNum+1, llmErr)
					} else {
						callResultValue = response // Assign LLM result
						logResp := response
						if len(logResp) > 70 {
							logResp = logResp[:67] + "..."
						}
						fmt.Printf("      [LLM OK] Response: %q stored\n", logResp)
					}
				}

			} else { // --- Procedure Call ---
				// Procedures currently expect string args passed via RunProcedure
				stringArgs := make([]string, len(evaluatedArgs))
				for idx, val := range evaluatedArgs {
					stringArgs[idx] = fmt.Sprintf("%v", val) // Convert evaluated args to strings for proc call
				}

				procToCall := target
				fmt.Printf("    [Exec] CALL Procedure: %s(%v)\n", procToCall, stringArgs)
				// Note: RunProcedure recursively calls executeSteps, creating a new scope
				procResultValue, procCallErr := i.RunProcedure(procToCall, stringArgs...) // Use distinct vars
				if procCallErr != nil {
					callErr = fmt.Errorf("step %d: CALL to proc '%s' failed: %w", stepNum+1, procToCall, procCallErr)
				} else {
					callResultValue = procResultValue // Assign procedure result
				}
			}

			// --- Post CALL Processing ---
			if callErr != nil {
				fmt.Printf("[DEBUG] --- CALL step %d FAILED (Error: %v) ---\n", stepNum+1, callErr)
				return nil, false, callErr // Return error immediately if any call type failed
			}

			// Set lastCallResult *ONCE* after successful call
			i.lastCallResult = callResultValue
			fmt.Printf("[DEBUG] --- Finished CALL step %d successfully (lastCallResult type: %T) ---\n", stepNum+1, i.lastCallResult)
			// --- End Post CALL ---

		case "IF":
			conditionStr := step.Cond
			fmt.Printf("    [Eval] IF Condition: %s\n", conditionStr)
			conditionResult, evalErr := i.evaluateCondition(conditionStr) // evaluateCondition from interpreter_b.go
			if evalErr != nil {
				fmt.Printf("      [Warn] IF condition error: %v\n", evalErr)
				conditionResult = false // Treat error as false condition
			}
			fmt.Printf("      [Eval] Condition Result: %t\n", conditionResult)
			if conditionResult {
				if blockBody, ok := step.Value.([]Step); ok {
					fmt.Printf("      [Exec IF Block] %d steps\n", len(blockBody))
					// --- Recursive Call ---
					blockResult, blockReturned, blockErr := i.executeSteps(blockBody)
					// Propagate results immediately if block returned or errored
					if blockErr != nil {
						return nil, false, blockErr
					}
					if blockReturned {
						return blockResult, true, nil
					}
					// --- End Recursive Call ---
				} else if step.Value != nil { // Should be nil or []Step
					return nil, false, fmt.Errorf("step %d: IF body unexpected type %T", stepNum+1, step.Value)
				}
				// If condition was true, skip any potential ELSE part of this IF
				skipElse = true
			}

		case "ELSE": // Currently requires parser support for associating ELSE with IF
			// This logic assumes ELSE can only follow an IF that evaluated to false (skipElse is false)
			if blockBody, ok := step.Value.([]Step); ok {
				fmt.Printf("    [Exec ELSE Block] %d steps\n", len(blockBody))
				// --- Recursive Call ---
				blockResult, blockReturned, blockErr := i.executeSteps(blockBody)
				// Propagate results immediately
				if blockErr != nil {
					return nil, false, blockErr
				}
				if blockReturned {
					return blockResult, true, nil
				}
				// --- End Recursive Call ---
			} else if step.Value != nil { // Should be nil or []Step
				return nil, false, fmt.Errorf("step %d: ELSE body unexpected type %T", stepNum+1, step.Value)
			}

		case "RETURN":
			valueExpr, ok := step.Value.(string)
			if !ok {
				// Allow RETURN without value? Assign nil? For now, require expression string.
				return nil, false, fmt.Errorf("step %d: RETURN value must be an expression string (type %T)", stepNum+1, step.Value)
			}
			returnValue := i.evaluateExpression(valueExpr) // Evaluate the return expression
			return returnValue, true, nil                  // Signal return occurred + VALUE

		case "WHILE":
			conditionStr := step.Cond
			fmt.Printf("    [Exec] WHILE %s ...\n", conditionStr)
			loopCounter := 0
			maxLoops := 1000 // Prevent infinite loops
			loopExecuted := false
			var bodyErr error              // Error from loop body execution
			var returnedFromBody bool      // Did the body execute RETURN?
			var resultFromLoop interface{} // Value returned from body

			for loopCounter < maxLoops {
				conditionResult, evalErr := i.evaluateCondition(conditionStr) // Check condition before each iteration
				if evalErr != nil {
					fmt.Printf("      [Warn] WHILE condition error: %v (exiting loop)\n", evalErr)
					bodyErr = evalErr // Store error and break
					break
				}
				if !conditionResult {
					fmt.Printf("      [Loop Check %d] Cond %q -> false. Exiting.\n", loopCounter, conditionStr)
					break // Exit loop if condition is false
				}

				fmt.Printf("      [Loop Check %d] Cond %q -> true. Entering body.\n", loopCounter, conditionStr)
				loopExecuted = true

				// Execute loop body using helper
				resultFromLoop, returnedFromBody, bodyErr = i.executeLoopBody(step.Value, stepNum, loopCounter)

				// Check results from body execution
				if bodyErr != nil {
					break
				} // Exit loop on error
				if returnedFromBody {
					break
				} // Exit loop on RETURN

				loopCounter++
			} // End while loop

			// Handle loop exit reason and potential RETURN value propagation
			if bodyErr != nil {
				return nil, false, fmt.Errorf("error in WHILE loop (iter %d): %w", loopCounter, bodyErr)
			}
			if loopCounter >= maxLoops {
				return nil, false, fmt.Errorf("step %d: WHILE loop exceeded max iterations (%d)", stepNum+1, maxLoops)
			}
			if returnedFromBody {
				return resultFromLoop, true, nil
			} // Propagate RETURN from body
			if loopExecuted {
				fmt.Printf("      [Loop Finished Normally]\n")
			} // Normal exit after condition became false

		case "FOR": // FOR EACH variable IN collection DO ... END
			loopVar := step.Target
			collectionExpr := step.Cond
			evaluatedCollection := i.evaluateExpression(collectionExpr) // Evaluate the collection expression
			fmt.Printf("    [Exec] FOR EACH %s IN %s (evaluated type: %T)\n", loopVar, collectionExpr, evaluatedCollection)

			// Save original value of loop variable (if exists) to restore later
			originalLoopVarValue, loopVarExists := i.variables[loopVar]
			loopExecuted := false
			var bodyErr error              // Error from loop body
			var returnedFromBody bool      // RETURN from loop body?
			var resultFromLoop interface{} // Value returned from body

			// --- Iteration Logic ---
			switch collection := evaluatedCollection.(type) {
			case []string: // Iterate over native string slice first
				fmt.Printf("      [Looping Native String Slice: %d items]\n", len(collection))
				if len(collection) > 0 {
					loopExecuted = true
				}
				for itemNum, item := range collection {
					i.variables[loopVar] = item // Assign element to loop var
					fmt.Printf("      [Iter %d String Slice Item] SET %s = %q\n", itemNum, loopVar, item)
					resultFromLoop, returnedFromBody, bodyErr = i.executeLoopBody(step.Value, stepNum, itemNum)
					if bodyErr != nil || returnedFromBody {
						break
					} // Exit inner loop
				}
			case []interface{}: // Iterate over native interface slice
				fmt.Printf("      [Looping Native Interface Slice: %d items]\n", len(collection))
				if len(collection) > 0 {
					loopExecuted = true
				}
				for itemNum, item := range collection {
					i.variables[loopVar] = item // Assign element
					fmt.Printf("      [Iter %d Interface Slice Item] SET %s = %v (%T)\n", itemNum, loopVar, item, item)
					resultFromLoop, returnedFromBody, bodyErr = i.executeLoopBody(step.Value, stepNum, itemNum)
					if bodyErr != nil || returnedFromBody {
						break
					} // Exit inner loop
				}
			case string: // Iterate over string (potentially comma-split or characters)
				shouldCommaSplit := false // Heuristic check for comma splitting
				if strings.Contains(collection, ",") {
					parts := strings.Split(collection, ",")
					if len(parts) > 1 {
						shouldCommaSplit = true
					}
					if len(parts) == 1 && strings.TrimSpace(parts[0]) != collection {
						shouldCommaSplit = true
					}
				}

				if shouldCommaSplit { // Comma Split Path
					items := strings.Split(collection, ",")
					fmt.Printf("      [Looping Comma-Split String: %d items]\n", len(items))
					if len(items) > 0 {
						loopExecuted = true
					}
					for itemNum, item := range items {
						trimmedItem := strings.TrimSpace(item)
						i.variables[loopVar] = trimmedItem
						fmt.Printf("      [Iter %d Comma Item] SET %s = %q\n", itemNum, loopVar, trimmedItem)
						resultFromLoop, returnedFromBody, bodyErr = i.executeLoopBody(step.Value, stepNum, itemNum)
						if bodyErr != nil || returnedFromBody {
							break
						}
					}
				} else { // Character Iteration Path
					fmt.Printf("      [Looping String Chars: %q]\n", collection)
					if len(collection) > 0 {
						loopExecuted = true
					}
					for itemNum, charRune := range collection {
						charStr := string(charRune)
						i.variables[loopVar] = charStr
						fmt.Printf("      [Iter %d Char] SET %s = %q\n", itemNum, loopVar, charStr)
						resultFromLoop, returnedFromBody, bodyErr = i.executeLoopBody(step.Value, stepNum, itemNum)
						if bodyErr != nil || returnedFromBody {
							break
						}
					}
				}
			// TODO: Add map iteration case here
			default: // Fallback for unhandled types
				fmt.Printf("      [Looping Fallback - Unhandled Type: %T]\n", evaluatedCollection)
				bodyErr = fmt.Errorf("cannot iterate over type %T in FOR EACH", evaluatedCollection)
			} // End switch collection type
			// --- End Iteration Logic ---

			// Restore loop variable state after loop finishes or breaks
			if loopVarExists {
				i.variables[loopVar] = originalLoopVarValue
			} else {
				delete(i.variables, loopVar)
			}

			// Handle loop exit reason and potential RETURN value propagation
			if bodyErr != nil {
				return nil, false, fmt.Errorf("error in FOR EACH loop body: %w", bodyErr)
			}
			if returnedFromBody {
				return resultFromLoop, true, nil
			} // Propagate RETURN from body
			if loopExecuted {
				fmt.Printf("      [Loop Finished Normally]\n")
			}

		default: // Unknown step type
			return nil, false, fmt.Errorf("unknown step type encountered in step %d: %s", stepNum+1, step.Type)
		} // End switch step.Type

	} // End for loop over steps

	// Finished all steps without hitting RETURN or error
	return nil, false, nil
} // End executeSteps

// executeLoopBody helper - Executes steps within a loop iteration (WHILE/FOR)
// Returns: result (if body returned), wasReturn signal, error
func (i *Interpreter) executeLoopBody(bodyValue interface{}, stepNum int, iterNum int) (result interface{}, wasReturn bool, err error) { // Named returns
	blockBody, ok := bodyValue.([]Step)
	if !ok {
		// Allow nil body (empty loop body)
		if bodyValue == nil {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("step %d: Loop body unexpected type %T (iter %d)", stepNum+1, bodyValue, iterNum)
	}

	if len(blockBody) == 0 {
		return nil, false, nil // Empty block is valid
	}

	// --- Recursive Call to execute steps for the block ---
	// This inherits the current scope (including the loop variable)
	result, wasReturn, err = i.executeSteps(blockBody) // Capture all 3 return values
	// Return results directly upwards - the caller (WHILE/FOR) handles propagation
	return result, wasReturn, err
	// --- End Recursive Call ---
}

// --- Helper Functions ---
// Helper to truncate strings for logging
func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

// Assume other helpers (evaluateCondition, evaluateExpression, tools, trimCodeFences, etc.)
// are defined in interpreter_b.go, interpreter_c.go, tools_*.go
