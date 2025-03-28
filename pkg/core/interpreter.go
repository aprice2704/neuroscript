package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath" // Import reflect
	"strings"
	// NOTE: parseStep removed if it was here previously
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

func NewInterpreter() *Interpreter {
	return &Interpreter{
		variables:       make(map[string]interface{}),
		knownProcedures: make(map[string]Procedure),
		vectorIndex:     make(map[string][]float32),
		embeddingDim:    16, // Example dimension
	}
}

// LoadProcedures (Unchanged from original)
func (i *Interpreter) LoadProcedures(procs []Procedure) error {
	for _, p := range procs {
		if _, exists := i.knownProcedures[p.Name]; exists {
			fmt.Printf("[Warning] Reloading procedure: %s\n", p.Name)
		}
		i.knownProcedures[p.Name] = p
	}
	return nil
}

// RunProcedure (Unchanged from original)
func (i *Interpreter) RunProcedure(procName string, args ...string) (interface{}, error) {
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
	originalVars := i.variables
	originalProcName := i.currentProcName
	i.variables = localVars
	i.currentProcName = procName
	result, err := i.executeSteps(proc.Steps) // Execute using local scope
	i.variables = originalVars
	i.currentProcName = originalProcName

	logResult := fmt.Sprintf("%v", result)
	if len(logResult) > 70 {
		logResult = logResult[:67] + "..."
	}
	fmt.Printf("[Exec] Procedure '%s' finished. Returning: %q\n", procName, logResult)

	return result, err
}

// --- Step Execution ---
// ** MODIFIED TO HANDLE BLOCKS and FOR EACH on strings **
func (i *Interpreter) executeSteps(steps []Step) (interface{}, error) {
	skipElse := false // Flag to skip ELSE block if IF was true
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	for stepNum, step := range steps {
		// Check for empty/comment steps (shouldn't happen if parser filters)
		if step.Type == "" {
			continue
		}
		// Basic logging
		logTarget := step.Target
		if step.Type == "IF" || step.Type == "WHILE" {
			logTarget = step.Cond
		}
		if len(logTarget) > 40 {
			logTarget = logTarget[:37] + "..."
		}
		fmt.Printf("  [Step %d] %s %s ...\n", stepNum+1, step.Type, logTarget)

		// Handle skipping ELSE block
		if skipElse {
			if step.Type == "ELSE" {
				fmt.Printf("    [Skip] Skipping ELSE block because IF was true.\n")
				skipElse = false // Reset flag after skipping ELSE
				continue         // Skip this ELSE step
			}
			skipElse = false // If not ELSE, stop skipping
		}

		switch step.Type {
		case "SET":
			targetVar := step.Target
			valueExpr, ok := step.Value.(string)
			if !ok {
				return nil, fmt.Errorf("step %d: SET value is not a string (internal error: type %T)", stepNum+1, step.Value)
			}
			finalValue := i.evaluateExpression(valueExpr) // Evaluate expression (defined in _c.go potentially)
			if targetVar == "generated_code" {
				// Special handling for generated_code (utility function defined in _c.go potentially)
				if finalStr, isStr := finalValue.(string); isStr {
					trimmedVal := trimCodeFences(finalStr)
					if trimmedVal != finalStr {
						fmt.Printf("      [Fence Strip] Trimmed fences if present\n")
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
				evaluatedArg := i.evaluateExpression(argExpr) // Evaluate expression (defined in _c.go potentially)
				resolvedArgs[idx] = fmt.Sprintf("%v", evaluatedArg)
			}
			target := step.Target

			if strings.HasPrefix(target, "TOOL.") { // --- TOOL Call ---
				toolName := strings.TrimPrefix(target, "TOOL.")
				var toolErr error
				var toolResult interface{} = "OK"
				fmt.Printf("    [Exec] CALL TOOL: %s(%v)\n", toolName, resolvedArgs)
				// Use utility functions (defined in _c.go potentially)
				switch toolName {
				case "ReadFile":
					if len(resolvedArgs) != 1 {
						toolErr = fmt.Errorf("ReadFile expects 1 arg (filepath)")
						break
					}
					filePathArg := resolvedArgs[0]
					absPath, secErr := secureFilePath(filePathArg, cwd)
					if secErr != nil {
						toolErr = fmt.Errorf("ReadFile security/path error: %w", secErr)
						break
					}
					fmt.Printf("      [Sec Check OK] Attempting read from: %s\n", absPath)
					contentBytes, readErr := os.ReadFile(absPath)
					if readErr != nil {
						toolErr = fmt.Errorf("ReadFile failed for '%s': %w", absPath, readErr)
					} else {
						toolResult = string(contentBytes)
						fmt.Printf("      [Tool OK] Read %d bytes from %s\n", len(contentBytes), absPath)
					}
				case "WriteFile":
					if len(resolvedArgs) != 2 {
						toolErr = fmt.Errorf("WriteFile expects 2 args (filepath, content)")
						break
					}
					filePathArg, contentToWrite := resolvedArgs[0], resolvedArgs[1]
					absPath, secErr := secureFilePath(filePathArg, cwd)
					if secErr != nil {
						toolErr = fmt.Errorf("WriteFile security/path error: %w", secErr)
						break
					}
					fmt.Printf("      [Sec Check OK] Attempting write to: %s\n", absPath)
					dirPath := filepath.Dir(absPath)
					if dirErr := os.MkdirAll(dirPath, 0755); dirErr != nil {
						toolErr = fmt.Errorf("WriteFile mkdir failed '%s': %w", dirPath, dirErr)
						break
					}
					contentBytes := []byte(fmt.Sprintf("%v", contentToWrite))
					writeErr := os.WriteFile(absPath, contentBytes, 0644)
					if writeErr != nil {
						toolErr = fmt.Errorf("WriteFile failed for '%s': %w", absPath, writeErr)
					} else {
						fmt.Printf("      [Tool OK] Wrote %d bytes to %s\n", len(contentBytes), absPath)
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
						toolErr = fmt.Errorf("query embed fail: %w", embErr)
						break
					} // GenerateEmbedding likely in _c.go
					type Result struct {
						Path  string
						Score float64
					}
					results := []Result{}
					threshold := 0.5
					for path, storedEmb := range i.vectorIndex {
						score, simErr := cosineSimilarity(queryEmb, storedEmb) // cosineSimilarity likely in _c.go
						if simErr != nil {
							fmt.Printf("      [Warn] Cannot compare with '%s': %v\n", path, simErr)
							continue
						}
						if score >= threshold {
							results = append(results, Result{Path: path, Score: score})
							fmt.Printf("        Match: %s (Score: %.4f)\n", path, score)
						}
					}
					resultBytes, jsonErr := json.Marshal(results)
					if jsonErr != nil {
						toolErr = fmt.Errorf("marshal results fail: %w", jsonErr)
					} else {
						toolResult = string(resultBytes)
					}
					fmt.Printf("      [Tool OK] SearchSkills returning JSON array (or empty)\n")
				case "GitAdd":
					if len(resolvedArgs) != 1 {
						toolErr = fmt.Errorf("GitAdd expects 1 arg (filepath)")
						break
					}
					filePathArg := resolvedArgs[0]
					absPath, secErr := secureFilePath(filePathArg, cwd)
					if secErr != nil {
						toolErr = fmt.Errorf("GitAdd security/path error: %w", secErr)
						break
					}
					fmt.Printf("      [Tool] Running: git add %s\n", absPath)
					toolErr = runGitCommand("add", absPath) // runGitCommand likely in _c.go
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
					toolErr = runGitCommand("commit", "-m", message) // runGitCommand likely in _c.go
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
						toolErr = fmt.Errorf("VectorUpdate security/path error: %w", secErr)
						break
					}
					fmt.Printf("      [Tool] Updating vector index for: %s\n", absPath)
					contentBytes, readErr := os.ReadFile(absPath)
					if readErr != nil {
						toolErr = fmt.Errorf("VectorUpdate failed to read file '%s': %w", absPath, readErr)
						break
					}
					textToEmbed := string(contentBytes)
					embedding, embErr := i.GenerateEmbedding(textToEmbed) // GenerateEmbedding likely in _c.go
					if embErr != nil {
						toolErr = fmt.Errorf("embed fail for '%s': %w", absPath, embErr)
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
					fmt.Printf("      [Tool OK] SanitizeFilename result: %q\n", toolResult) // sanitizeFilename likely in _c.go
				default:
					toolErr = fmt.Errorf("unknown TOOL '%s'", toolName)
				}
				if toolErr != nil {
					return nil, fmt.Errorf("step %d: TOOL %s execution failed: %w", stepNum+1, toolName, toolErr)
				}
				i.lastCallResult = toolResult

			} else if target == "LLM" { // --- LLM Call ---
				if len(resolvedArgs) != 1 {
					return nil, fmt.Errorf("step %d: CALL LLM expects 1 prompt arg, got %d", stepNum+1, len(resolvedArgs))
				}
				prompt := resolvedArgs[0]
				fmt.Printf("    [Exec] CALL LLM prompt: %q\n", prompt)
				response, llmErr := CallLLMAPI(prompt) // Assumes CallLLMAPI is accessible (e.g., in llm.go)
				if llmErr != nil {
					return nil, fmt.Errorf("step %d: CALL LLM failed: %w", stepNum+1, llmErr)
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
				callResult, callErr := i.RunProcedure(procToCall, resolvedArgs...) // Recursive call
				if callErr != nil {
					return nil, fmt.Errorf("step %d: in procedure '%s', CALL to procedure '%s' failed: %w", stepNum+1, i.currentProcName, procToCall, callErr)
				}
				i.lastCallResult = callResult
			}

		case "IF":
			conditionStr := step.Cond
			fmt.Printf("    [Eval] IF Condition: %s\n", conditionStr)
			conditionResult, evalErr := i.evaluateCondition(conditionStr) // evaluateCondition in _b.go
			if evalErr != nil {
				fmt.Printf("      [Warn] IF condition evaluation error: %v (treating as false)\n", evalErr)
				conditionResult = false
			}
			fmt.Printf("      [Eval] Condition Result: %t\n", conditionResult)

			if conditionResult {
				// Check if Value is a block ([]Step)
				if blockBody, ok := step.Value.([]Step); ok {
					fmt.Printf("      [Exec IF Block] %d steps\n", len(blockBody))
					_, blockErr := i.executeSteps(blockBody) // Recursive call
					if blockErr != nil {
						return nil, blockErr
					} // Propagate error
					// RETURN is handled by recursive call returning immediately

				} else if step.Value == nil { // Handle empty block parsed correctly
					fmt.Printf("      [Skip] IF condition true, but block body is empty (Value is nil).\n")
				} else { // Should not happen with new parser if Value isn't []Step or nil
					return nil, fmt.Errorf("step %d: IF body has unexpected type %T (expected []Step or nil)", stepNum+1, step.Value)
				}
				skipElse = true // Skip next ELSE if IF condition was true
			}

		case "ELSE":
			// Check if Value is a block ([]Step)
			if blockBody, ok := step.Value.([]Step); ok {
				fmt.Printf("    [Exec ELSE Block] %d steps\n", len(blockBody))
				_, blockErr := i.executeSteps(blockBody) // Execute block
				if blockErr != nil {
					return nil, blockErr
				}
				// RETURN handled by recursive call

			} else if step.Value == nil { // Handle empty ELSE block parsed correctly
				fmt.Printf("      [Skip] ELSE encountered, but block body is empty (Value is nil).\n")
			} else {
				return nil, fmt.Errorf("step %d: ELSE body has unexpected type %T (expected []Step or nil)", stepNum+1, step.Value)
			}

		case "RETURN":
			valueExpr, ok := step.Value.(string)
			if !ok {
				return nil, fmt.Errorf("step %d: RETURN value is not a string (internal error)", stepNum+1)
			}
			returnValue := i.evaluateExpression(valueExpr) // evaluateExpression in _c.go
			// Logging done by RunProcedure wrapper
			return returnValue, nil // Return evaluated value - This terminates executeSteps

		case "FOR":
			loopVar := step.Target
			collectionExpr := step.Cond // Collection is stored in Cond by parser

			evaluatedCollection := i.evaluateExpression(collectionExpr) // evaluateExpression in _c.go
			fmt.Printf("    [Exec] FOR EACH %s IN %s (resolved collection type: %T)\n", loopVar, collectionExpr, evaluatedCollection)

			originalLoopVarValue, loopVarExists := i.variables[loopVar] // Save original value if exists
			loopExecuted := false                                       // Track if loop body ran at least once
			var bodyErr error                                           // Store error from body execution

			// --- Determine Iteration Method ---
			if collectionStr, okStr := evaluatedCollection.(string); okStr {
				// ** Iterate over String Characters (Runes) **
				fmt.Printf("      [Looping String Chars: %q]\n", collectionStr)
				if len(collectionStr) > 0 {
					loopExecuted = true
				}
				for itemNum, charRune := range collectionStr {
					charStr := string(charRune) // Convert rune to string
					i.variables[loopVar] = charStr
					fmt.Printf("      [Iter %d Char] SET %s = %q\n", itemNum, loopVar, charStr)

					// Execute Body (Block or Single Step) using helper
					bodyErr = i.executeLoopBody(step.Value, stepNum, itemNum)
					if bodyErr != nil {
						break
					} // Exit loop on error
					// RETURN inside body will cause executeLoopBody->executeSteps to return immediately
				}

			} else {
				// ** Iterate using Comma Split (Fallback/Default) **
				// TODO: Consider supporting slices/maps directly if interpreter handles them
				collectionValStr := fmt.Sprintf("%v", evaluatedCollection)
				items := []string{}
				if collectionValStr != "" {
					items = strings.Split(collectionValStr, ",")
				}
				fmt.Printf("      [Looping Comma-Split Items: %d]\n", len(items))
				if len(items) > 0 {
					loopExecuted = true
				}

				for itemNum, item := range items {
					trimmedItem := strings.TrimSpace(item)
					i.variables[loopVar] = trimmedItem
					fmt.Printf("      [Iter %d Item] SET %s = %q\n", itemNum, loopVar, trimmedItem)

					// Execute Body (Block or Single Step) using helper
					bodyErr = i.executeLoopBody(step.Value, stepNum, itemNum)
					if bodyErr != nil {
						break
					} // Exit loop on error
					// RETURN inside body will cause executeLoopBody->executeSteps to return immediately
				}
			}

			// Restore loop variable after loop finishes or breaks
			if loopVarExists {
				i.variables[loopVar] = originalLoopVarValue
			} else {
				delete(i.variables, loopVar)
			}

			// Handle errors from loop body execution
			if bodyErr != nil {
				return nil, bodyErr // Propagate error from body
			}

			if loopExecuted {
				fmt.Printf("      [Loop Finished]\n")
			}

		case "WHILE":
			conditionStr := step.Cond
			fmt.Printf("    [Exec] WHILE %s ...\n", conditionStr)

			loopCounter := 0
			maxLoops := 1000
			loopExecuted := false
			var bodyErr error // Store error from body

			for loopCounter < maxLoops {
				conditionResult, evalErr := i.evaluateCondition(conditionStr) // evaluateCondition in _b.go
				if evalErr != nil {
					fmt.Printf("      [Warn] WHILE condition evaluation error: %v (exiting loop)\n", evalErr)
					break
				}
				fmt.Printf("      [Loop Check %d] Condition %q -> %t\n", loopCounter, conditionStr, conditionResult)
				if !conditionResult {
					break
				}

				loopExecuted = true // Mark that loop body should execute

				// Execute Body (Block or Single Step) using helper
				bodyErr = i.executeLoopBody(step.Value, stepNum, loopCounter)
				if bodyErr != nil {
					break
				} // Exit loop on error
				// RETURN inside body will cause executeLoopBody->executeSteps to return immediately

				loopCounter++
			}

			// Handle errors / infinite loop break
			if bodyErr != nil {
				return nil, fmt.Errorf("error in WHILE loop body (iteration %d): %w", loopCounter, bodyErr)
			}
			if loopCounter >= maxLoops {
				return nil, fmt.Errorf("step %d: WHILE loop exceeded max iterations (%d)", stepNum+1, maxLoops)
			}

			if loopExecuted {
				fmt.Printf("      [Loop Finished]\n")
			}

		default:
			fmt.Printf("    [Warn] Unknown or unhandled step type '%s' encountered during execution.\n", step.Type)
		}
	}
	// Finished all steps without hitting RETURN
	return nil, nil
}

// executeLoopBody is a helper to execute the body of FOR/WHILE loops.
// Returns error if execution fails. RETURNs are handled by recursive executeSteps call.
func (i *Interpreter) executeLoopBody(bodyValue interface{}, stepNum int, iterNum int) error {
	if blockBody, ok := bodyValue.([]Step); ok {
		// fmt.Printf("      [Exec Loop Block] %d steps\n", len(blockBody)) // Logged by caller
		_, blockErr := i.executeSteps(blockBody) // Execute block recursively
		return blockErr                          // Propagate error (nil if successful, non-nil if error OR RETURN occurred?) - Need careful testing

	} else if bodyValue == nil {
		fmt.Printf("      [Skip] Loop body is empty (Value is nil).\n")
		return nil // Empty body, continue loop

	} else {
		// Fallback for single-line string bodies (maybe from old scripts?) - Treat as error now?
		// Let's require blocks for loops for simplicity with the new parser.
		// return fmt.Errorf("step %d: Loop body has unexpected type %T (iter %d, expected []Step or nil)", stepNum+1, bodyValue, iterNum)
		// OR try to parse it? Let's error for now. Interpreter expects blocks.
		return fmt.Errorf("step %d: Loop body has unexpected type %T (iter %d, expected []Step or nil)", stepNum+1, bodyValue, iterNum)
	}
}
