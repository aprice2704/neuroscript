package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// --- Interpreter ---
type Interpreter struct {
	variables       map[string]interface{}
	knownProcedures map[string]Procedure
	lastCallResult  interface{}
	vectorIndex     map[string][]float32 // Simple in-memory vector store
	embeddingDim    int                  // Dimension for mock embeddings
	currentProcName string               // Added to track current procedure for errors
}

// NewInterpreter creates a new interpreter instance
func NewInterpreter() *Interpreter {
	return &Interpreter{
		variables:       make(map[string]interface{}),
		knownProcedures: make(map[string]Procedure),
		vectorIndex:     make(map[string][]float32),
		embeddingDim:    16, // Example dimension
	}
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
		localVars[paramName] = args[idx]
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
		fmt.Printf("[Exec] Procedure '%s' finished normally. Final Result (Implicit nil): %v\n", procName, result)
	}

	return result, err
}

// --- Step Execution ---
// Propagates return value and signal correctly
func (i *Interpreter) executeSteps(steps []Step) (result interface{}, wasReturn bool, err error) { // Named returns
	skipElse := false
	cwd, cwdErr := os.Getwd()
	if cwdErr != nil {
		return nil, false, fmt.Errorf("cwd error: %w", cwdErr)
	}

	for stepNum, step := range steps {
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

		if skipElse {
			if step.Type == "ELSE" {
				fmt.Printf("    [Skip] Skipping ELSE block...\n")
				skipElse = false
				continue
			}
			skipElse = false
		}

		switch step.Type {
		case "SET":
			targetVar := step.Target
			valueExpr, ok := step.Value.(string)
			if !ok {
				return nil, false, fmt.Errorf("step %d: SET value is not string (type %T)", stepNum+1, step.Value)
			}
			finalValue := i.evaluateExpression(valueExpr) // Assumes evaluateExpression in interpreter_c.go
			if targetVar == "generated_code" {
				if finalStr, isStr := finalValue.(string); isStr {
					trimmedVal := trimCodeFences(finalStr) // Assumes trimCodeFences in interpreter_c.go
					if trimmedVal != finalStr {
						fmt.Printf("      [Fence Strip] Trimmed fences\n")
						finalValue = trimmedVal
					}
				}
			}
			i.variables[targetVar] = finalValue
			logValueStr := fmt.Sprintf("%v", finalValue)
			if len(logValueStr) > 70 {
				logValueStr = logValueStr[:67] + "..."
			}
			fmt.Printf("    [Exec] SET %s = %q\n", targetVar, logValueStr)

		case "CALL":
			i.lastCallResult = nil
			resolvedArgs := make([]string, len(step.Args))
			for idx, argExpr := range step.Args {
				evaluatedArg := i.evaluateExpression(argExpr) // Assumes evaluateExpression in interpreter_c.go
				resolvedArgs[idx] = fmt.Sprintf("%v", evaluatedArg)
			}
			target := step.Target

			if strings.HasPrefix(target, "TOOL.") { // --- TOOL Call ---
				toolName := strings.TrimPrefix(target, "TOOL.")
				var toolErr error
				var toolResult interface{} = "OK" // Default result
				fmt.Printf("    [Exec] CALL TOOL: %s(%v)\n", toolName, resolvedArgs)
				// Assumes helper functions defined elsewhere (e.g., interpreter_c.go)
				switch toolName {
				case "ReadFile":
					if len(resolvedArgs) != 1 {
						toolErr = fmt.Errorf("ReadFile expects 1 arg (filepath)")
						break
					}
					filePathArg := resolvedArgs[0]
					absPath, secErr := secureFilePath(filePathArg, cwd)
					if secErr != nil {
						toolErr = fmt.Errorf("ReadFile path error: %w", secErr)
						break
					}
					contentBytes, readErr := os.ReadFile(absPath)
					if readErr != nil {
						toolErr = fmt.Errorf("ReadFile failed for '%s': %w", absPath, readErr)
					} else {
						toolResult = string(contentBytes)
						fmt.Printf("      [Tool OK] Read %d bytes from %s\n", len(contentBytes), filepath.Base(absPath))
					}
				case "WriteFile":
					if len(resolvedArgs) != 2 {
						toolErr = fmt.Errorf("WriteFile expects 2 args (filepath, content)")
						break
					}
					filePathArg, contentToWrite := resolvedArgs[0], resolvedArgs[1]
					absPath, secErr := secureFilePath(filePathArg, cwd)
					if secErr != nil {
						toolErr = fmt.Errorf("WriteFile path error: %w", secErr)
						break
					}
					dirPath := filepath.Dir(absPath)
					if dirErr := os.MkdirAll(dirPath, 0755); dirErr != nil {
						toolErr = fmt.Errorf("WriteFile mkdir fail: %w", dirErr)
						break
					}
					contentBytes := []byte(fmt.Sprintf("%v", contentToWrite))
					writeErr := os.WriteFile(absPath, contentBytes, 0644)
					if writeErr != nil {
						toolErr = fmt.Errorf("WriteFile failed for '%s': %w", absPath, writeErr)
					} else {
						fmt.Printf("      [Tool OK] Wrote %d bytes to %s\n", len(contentBytes), filepath.Base(absPath))
					}
				case "SearchSkills": // Mock Search
					if len(resolvedArgs) != 1 {
						toolErr = fmt.Errorf("SearchSkills expects 1 arg (query)")
						break
					}
					query := resolvedArgs[0]
					fmt.Printf("      [Tool] Searching skills for: %q\n", query)
					queryEmb, embErr := i.GenerateEmbedding(query)
					if embErr != nil {
						toolErr = fmt.Errorf("embed fail: %w", embErr)
						break
					}
					type Result struct {
						Path  string
						Score float64
					}
					results := []Result{}
					threshold := 0.5
					for path, storedEmb := range i.vectorIndex {
						score, simErr := cosineSimilarity(queryEmb, storedEmb)
						if simErr == nil && score >= threshold {
							results = append(results, Result{Path: path, Score: score})
						}
					}
					resultBytes, jsonErr := json.Marshal(results)
					if jsonErr != nil {
						toolErr = fmt.Errorf("marshal results fail: %w", jsonErr)
					} else {
						toolResult = string(resultBytes)
					}
					fmt.Printf("      [Tool OK] SearchSkills found %d potential matches\n", len(results))
				case "GitAdd":
					if len(resolvedArgs) != 1 {
						toolErr = fmt.Errorf("GitAdd expects 1 arg (filepath)")
						break
					}
					filePathArg := resolvedArgs[0]
					absPath, secErr := secureFilePath(filePathArg, cwd)
					if secErr != nil {
						toolErr = fmt.Errorf("GitAdd path error: %w", secErr)
						break
					}
					fmt.Printf("      [Tool] Running: git add %s\n", absPath)
					toolErr = runGitCommand("add", absPath)
					if toolErr == nil {
						fmt.Printf("      [Tool OK] GitAdd\n")
					}
				case "GitCommit":
					if len(resolvedArgs) != 1 {
						toolErr = fmt.Errorf("GitCommit expects 1 arg (message)")
						break
					}
					message := resolvedArgs[0]
					fmt.Printf("      [Tool] Running: git commit -m %q\n", message)
					toolErr = runGitCommand("commit", "-m", message)
					if toolErr == nil {
						fmt.Printf("      [Tool OK] GitCommit\n")
					}
				case "VectorUpdate": // Mock Update
					if len(resolvedArgs) != 1 {
						toolErr = fmt.Errorf("VectorUpdate expects 1 arg (filepath)")
						break
					}
					filePathArg := resolvedArgs[0]
					absPath, secErr := secureFilePath(filePathArg, cwd)
					if secErr != nil {
						toolErr = fmt.Errorf("VectorUpdate path error: %w", secErr)
						break
					}
					fmt.Printf("      [Tool] Updating vector index for: %s\n", filepath.Base(absPath))
					contentBytes, readErr := os.ReadFile(absPath)
					if readErr != nil {
						toolErr = fmt.Errorf("read fail: %w", readErr)
						break
					}
					textToEmbed := string(contentBytes)
					embedding, embErr := i.GenerateEmbedding(textToEmbed)
					if embErr != nil {
						toolErr = fmt.Errorf("embed fail: %w", embErr)
						break
					}
					i.vectorIndex[absPath] = embedding
					fmt.Printf("      [Tool OK] VectorUpdate - Index size: %d\n", len(i.vectorIndex))
				case "SanitizeFilename":
					if len(resolvedArgs) != 1 {
						toolErr = fmt.Errorf("SanitizeFilename expects 1 arg (string)")
						break
					}
					toolResult = sanitizeFilename(resolvedArgs[0])
					fmt.Printf("      [Tool OK] SanitizeFilename result: %q\n", toolResult)
				default:
					toolErr = fmt.Errorf("unknown TOOL '%s'", toolName)
				} // End tool switch
				if toolErr != nil {
					return nil, false, fmt.Errorf("step %d: TOOL %s failed: %w", stepNum+1, toolName, toolErr)
				} // Return error
				i.lastCallResult = toolResult // Store result

			} else if target == "LLM" { // --- LLM Call ---
				if len(resolvedArgs) != 1 {
					return nil, false, fmt.Errorf("step %d: CALL LLM expects 1 prompt arg", stepNum+1)
				}
				prompt := resolvedArgs[0]
				fmt.Printf("    [Exec] CALL LLM prompt (first 80 chars): %q\n", truncateString(prompt, 80))
				response, llmErr := CallLLMAPI(prompt) // Assumes CallLLMAPI in llm.go
				if llmErr != nil {
					return nil, false, fmt.Errorf("step %d: CALL LLM failed: %w", stepNum+1, llmErr)
				}
				i.lastCallResult = response
				logResp := response
				if len(logResp) > 70 {
					logResp = logResp[:67] + "..."
				}
				fmt.Printf("      [LLM OK] Response: %q stored\n", logResp)

			} else { // --- Procedure Call ---
				procToCall := target
				fmt.Printf("    [Exec] CALL Procedure: %s(%v)\n", procToCall, resolvedArgs)
				// RunProcedure returns (val, err). Assume RETURN inside it exits immediately.
				callResultValue, callErr := i.RunProcedure(procToCall, resolvedArgs...)
				if callErr != nil {
					return nil, false, fmt.Errorf("step %d: CALL to proc '%s' failed: %w", stepNum+1, procToCall, callErr)
				} // Return error
				i.lastCallResult = callResultValue // Store procedure's return value
			}

		case "IF": // --- IF ---
			conditionStr := step.Cond
			fmt.Printf("    [Eval] IF Condition: %s\n", conditionStr)
			conditionResult, evalErr := i.evaluateCondition(conditionStr) // Assumes in interpreter_b.go
			if evalErr != nil {
				fmt.Printf("      [Warn] IF condition error: %v\n", evalErr)
				conditionResult = false
			}
			fmt.Printf("      [Eval] Condition Result: %t\n", conditionResult)
			if conditionResult {
				if blockBody, ok := step.Value.([]Step); ok {
					fmt.Printf("      [Exec IF Block] %d steps\n", len(blockBody))
					blockResult, blockReturned, blockErr := i.executeSteps(blockBody) // Recursive call
					// IMMEDIATELY RETURN if inner block returned/errored
					if blockErr != nil {
						return nil, false, blockErr
					}
					if blockReturned {
						return blockResult, true, nil
					} // Return VALUE + signal
				} else if step.Value != nil { // Error if Value is not nil and not []Step
					return nil, false, fmt.Errorf("step %d: IF body unexpected type %T", stepNum+1, step.Value)
				}
				// If blockBody is nil or empty, execution just falls through here
				skipElse = true
			}

		case "ELSE": // --- ELSE ---
			if blockBody, ok := step.Value.([]Step); ok {
				fmt.Printf("    [Exec ELSE Block] %d steps\n", len(blockBody))
				blockResult, blockReturned, blockErr := i.executeSteps(blockBody) // Recursive call
				// IMMEDIATELY RETURN if inner block returned/errored
				if blockErr != nil {
					return nil, false, blockErr
				} // Return error
				if blockReturned {
					return blockResult, true, nil
				} // Return VALUE + signal
			} else if step.Value != nil { // Error if Value is not nil and not []Step
				return nil, false, fmt.Errorf("step %d: ELSE body unexpected type %T", stepNum+1, step.Value)
			}
			// If blockBody is nil or empty, execution just falls through here

		case "RETURN": // --- RETURN ---
			valueExpr, ok := step.Value.(string)
			if !ok {
				return nil, false, fmt.Errorf("step %d: RETURN value not string", stepNum+1)
			}
			returnValue := i.evaluateExpression(valueExpr) // Assumes in interpreter_c.go
			return returnValue, true, nil                  // Signal return occurred + VALUE

		case "WHILE": // --- WHILE ---
			conditionStr := step.Cond
			fmt.Printf("    [Exec] WHILE %s ...\n", conditionStr)
			loopCounter := 0
			maxLoops := 1000
			loopExecuted := false
			var bodyErr error
			var returnedFromBody bool
			var resultFromLoop interface{} // Capture result if body returns

			for loopCounter < maxLoops {
				conditionResult, evalErr := i.evaluateCondition(conditionStr) // Assumes in interpreter_b.go
				if evalErr != nil {
					fmt.Printf("      [Warn] WHILE condition error: %v (exiting)\n", evalErr)
					bodyErr = evalErr
					break
				} // Store error, break loop
				if !conditionResult {
					fmt.Printf("      [Loop Check %d] Cond %q -> false. Exiting.\n", loopCounter, conditionStr)
					break
				} // Condition false, exit loop
				fmt.Printf("      [Loop Check %d] Cond %q -> true. Continuing.\n", loopCounter, conditionStr)
				loopExecuted = true

				resultFromLoop, returnedFromBody, bodyErr = i.executeLoopBody(step.Value, stepNum, loopCounter) // Call helper
				if bodyErr != nil {
					break
				} // Error in body, break loop
				if returnedFromBody {
					break
				} // RETURN in body, break loop

				loopCounter++
			} // End while loop

			// Handle loop exit reason and RETURN VALUE
			if bodyErr != nil {
				return nil, false, fmt.Errorf("error in WHILE loop (iter %d): %w", loopCounter, bodyErr)
			} // Propagate error
			if loopCounter >= maxLoops {
				return nil, false, fmt.Errorf("step %d: WHILE loop max iterations (%d)", stepNum+1, maxLoops)
			} // Propagate error
			if returnedFromBody {
				return resultFromLoop, true, nil
			} // Return exit: return the captured value + signal
			if loopExecuted {
				fmt.Printf("      [Loop Finished Normally]\n")
			} // Normal exit

		case "FOR": // --- FOR EACH ---
			loopVar := step.Target
			collectionExpr := step.Cond
			evaluatedCollection := i.evaluateExpression(collectionExpr) // Assumes in interpreter_c.go
			fmt.Printf("    [Exec] FOR EACH %s IN %s (type: %T)\n", loopVar, collectionExpr, reflect.TypeOf(evaluatedCollection))
			originalLoopVarValue, loopVarExists := i.variables[loopVar]
			loopExecuted := false
			var bodyErr error
			var returnedFromBody bool
			var resultFromLoop interface{} // Capture result if body returns

			switch collection := evaluatedCollection.(type) {
			case string:
				// Prioritize Comma Split If Commas Exist
				if strings.Contains(collection, ",") {
					items := strings.Split(collection, ",")
					useSplit := len(items) > 1
					if !useSplit && len(items) == 1 {
						if strings.TrimSpace(items[0]) != collection {
							useSplit = true
						}
					}
					if useSplit {
						fmt.Printf("      [Looping Comma-Split Items from string: %d]\n", len(items))
						if len(items) > 0 {
							loopExecuted = true
						}
						for itemNum, item := range items {
							trimmedItem := strings.TrimSpace(item)
							i.variables[loopVar] = trimmedItem
							fmt.Printf("      [Iter %d Item] SET %s = %q\n", itemNum, loopVar, trimmedItem)
							resultFromLoop, returnedFromBody, bodyErr = i.executeLoopBody(step.Value, stepNum, itemNum)
							if bodyErr != nil || returnedFromBody {
								break
							} // Exit inner loop
						}
						goto PostForLoop // Jump past character iteration
					}
				}
				// Fallback to Character Iteration
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
					} // Exit inner loop
				}
			// --- Native List/Slice handling ---
			case []interface{}:
				fmt.Printf("      [Looping Native Slice: %d items]\n", len(collection))
				if len(collection) > 0 {
					loopExecuted = true
				}
				for itemNum, item := range collection {
					i.variables[loopVar] = item
					fmt.Printf("      [Iter %d Slice Item] SET %s = %v (%T)\n", itemNum, loopVar, item, item)
					resultFromLoop, returnedFromBody, bodyErr = i.executeLoopBody(step.Value, stepNum, itemNum)
					if bodyErr != nil || returnedFromBody {
						break
					}
				}
			case []string:
				fmt.Printf("      [Looping String Slice: %d items]\n", len(collection))
				if len(collection) > 0 {
					loopExecuted = true
				}
				for itemNum, item := range collection {
					i.variables[loopVar] = item
					fmt.Printf("      [Iter %d String Slice Item] SET %s = %q\n", itemNum, loopVar, item)
					resultFromLoop, returnedFromBody, bodyErr = i.executeLoopBody(step.Value, stepNum, itemNum)
					if bodyErr != nil || returnedFromBody {
						break
					}
				}
				// TODO: Add map iteration case here
			default: // Fallback comma split for other types
				collectionValStr := fmt.Sprintf("%v", evaluatedCollection)
				items := []string{}
				if collectionValStr != "" {
					items = strings.Split(collectionValStr, ",")
				}
				fmt.Printf("      [Looping Comma-Split Fallback from %T: %d]\n", evaluatedCollection, len(items))
				if len(items) > 0 {
					loopExecuted = true
				}
				for itemNum, item := range items {
					trimmedItem := strings.TrimSpace(item)
					i.variables[loopVar] = trimmedItem
					fmt.Printf("      [Iter %d Item] SET %s = %q\n", itemNum, loopVar, trimmedItem)
					resultFromLoop, returnedFromBody, bodyErr = i.executeLoopBody(step.Value, stepNum, itemNum)
					if bodyErr != nil || returnedFromBody {
						break
					}
				}
			} // End switch collection type

		PostForLoop: // Label for shared post-loop logic
			if loopVarExists {
				i.variables[loopVar] = originalLoopVarValue
			} else {
				delete(i.variables, loopVar)
			} // Restore loop var

			// Handle loop exit reason and RETURN VALUE
			if bodyErr != nil {
				return nil, false, fmt.Errorf("error in FOR EACH loop body: %w", bodyErr)
			} // Error exit
			if returnedFromBody {
				return resultFromLoop, true, nil
			} // Return exit: return value + signal
			if loopExecuted {
				fmt.Printf("      [Loop Finished Normally]\n")
			} // Normal exit

		default: // Unknown step type
			return nil, false, fmt.Errorf("unknown step type encountered: %s", step.Type)
		}
	} // End for loop over steps

	// Finished all steps without hitting RETURN or error
	return nil, false, nil
}

// executeLoopBody helper - Returns return value, signal, and error
func (i *Interpreter) executeLoopBody(bodyValue interface{}, stepNum int, iterNum int) (result interface{}, wasReturn bool, err error) { // Named returns
	if blockBody, ok := bodyValue.([]Step); ok {
		// Execute block recursively. Capture all return values.
		result, wasReturn, err = i.executeSteps(blockBody) // Capture all 3
		// Return results directly upwards
		return result, wasReturn, err

	} else if bodyValue == nil {
		return nil, false, nil // Empty body is valid, finished normally
	} else { // Invalid body type
		return nil, false, fmt.Errorf("step %d: Loop body unexpected type %T (iter %d)", stepNum+1, bodyValue, iterNum)
	}
}

// --- Helper Functions ---
// Helper to truncate strings for logging
func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

// Assume helpers from interpreter_b/c (evaluateCondition, evaluateExpression, tools etc.) are accessible in this package.
// Make sure necessary utility functions like trimCodeFences, sanitizeFilename, secureFilePath, runGitCommand, GenerateEmbedding, cosineSimilarity are defined (likely in interpreter_c.go)
// Make sure isValidIdentifier is defined (likely in parser_c.go or interpreter_b.go)
