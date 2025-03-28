package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	// Removed parseStep definition from here
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

func (i *Interpreter) LoadProcedures(procs []Procedure) error {
	for _, p := range procs {
		if _, exists := i.knownProcedures[p.Name]; exists {
			fmt.Printf("[Warning] Reloading procedure: %s\n", p.Name)
		}
		i.knownProcedures[p.Name] = p
		// TODO: Maybe index docstring/code for search here?
	}
	return nil
}

func (i *Interpreter) RunProcedure(procName string, args ...string) (interface{}, error) {
	proc, exists := i.knownProcedures[procName]
	if !exists {
		return nil, fmt.Errorf("procedure '%s' not defined", procName)
	}

	// Create a local scope for this procedure run
	localVars := make(map[string]interface{})
	// Copy global vars? Or allow access? Let's keep it simple: only args are local.
	// If we want globals:
	// for k, v := range i.variables { localVars[k] = v }

	if len(args) != len(proc.Params) {
		return nil, fmt.Errorf("procedure '%s' expects %d argument(s) (%v), but received %d", procName, len(proc.Params), proc.Params, len(args))
	}
	for idx, paramName := range proc.Params {
		// Args are passed as strings, store them as such initially.
		localVars[paramName] = args[idx]
		fmt.Printf("  [Arg Init] SET %s = %q\n", paramName, args[idx])
	}

	fmt.Printf("[Exec] Running Procedure: %s\n", procName)
	// --- Scope Management ---
	originalVars := i.variables               // Save outer scope
	originalProcName := i.currentProcName     // Save outer proc name context
	i.variables = localVars                   // Set current scope to local
	i.currentProcName = procName              // Set current proc name context
	result, err := i.executeSteps(proc.Steps) // Execute using local scope
	i.variables = originalVars                // Restore outer scope
	i.currentProcName = originalProcName      // Restore outer proc name context
	// --- End Scope Management ---

	// Log result before returning
	logResult := fmt.Sprintf("%v", result)
	if len(logResult) > 70 {
		logResult = logResult[:67] + "..."
	}
	fmt.Printf("[Exec] Procedure '%s' finished. Returning: %q\n", procName, logResult)

	return result, err
}

// --- Step Execution ---
// --- Step Execution ---
func (i *Interpreter) executeSteps(steps []Step) (interface{}, error) {
	skipElse := false      // Flag to skip ELSE block if IF was true
	cwd, err := os.Getwd() // Get current working directory for file safety checks
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	for stepNum, step := range steps {
		// Check for empty/comment steps parsed (should have Type=="")
		if step.Type == "" {
			continue
		}
		fmt.Printf("  [Step %d] %s %s ...\n", stepNum+1, step.Type, step.Target)

		// Handle skipping ELSE block
		if skipElse {
			if step.Type == "ELSE" {
				fmt.Printf("    [Skip] Skipping ELSE block because IF was true.\n")
				skipElse = false // Reset flag after skipping ELSE
				continue         // Skip this ELSE step
			}
			// If it wasn't ELSE, we stop skipping for subsequent steps
			skipElse = false
		}

		switch step.Type {
		case "SET":
			targetVar := step.Target
			// Value from parser should be the raw string representation
			valueExpr, ok := step.Value.(string)
			if !ok {
				// Should not happen if parser is correct
				return nil, fmt.Errorf("step %d: SET value is not a string (internal error)", stepNum+1)
			}

			// Evaluate the expression string (handles placeholders, vars, literals, concat)
			finalValue := i.evaluateExpression(valueExpr)

			// Special handling for specific variable names if needed
			// This check should happen AFTER evaluation
			if targetVar == "generated_code" {
				// Check if finalValue is a string before trimming
				if finalStr, isStr := finalValue.(string); isStr {
					trimmedVal := trimCodeFences(finalStr)
					if trimmedVal != finalStr {
						fmt.Printf("      [Fence Strip] Trimmed fences if present\n")
						finalValue = trimmedVal
					}
				}
			}

			i.variables[targetVar] = finalValue // Store evaluated result in current scope

			// Log potentially truncated value
			logValueStr := fmt.Sprintf("%v", finalValue) // Convert finalValue to string for logging
			if len(logValueStr) > 70 {
				logValueStr = logValueStr[:67] + "..."
			}
			fmt.Printf("    [Exec] SET %s = %q\n", targetVar, logValueStr)

		case "CALL":
			i.lastCallResult = nil // Reset last result before call
			// Evaluate all arguments *before* making the call
			resolvedArgs := make([]string, len(step.Args))
			for idx, argExpr := range step.Args {
				// evaluateExpression returns the evaluated value (string usually)
				evaluatedArg := i.evaluateExpression(argExpr)
				// Ensure resolvedArgs stores strings
				resolvedArgs[idx] = fmt.Sprintf("%v", evaluatedArg)
			}

			target := step.Target // Procedure name, LLM, or TOOL.xxx

			if strings.HasPrefix(target, "TOOL.") { // --- TOOL Call ---
				toolName := strings.TrimPrefix(target, "TOOL.")
				var toolErr error
				var toolResult interface{} = "OK" // Default result unless tool returns something else

				fmt.Printf("    [Exec] CALL TOOL: %s(%v)\n", toolName, resolvedArgs)
				switch toolName {
				case "ReadFile":
					if len(resolvedArgs) != 1 {
						toolErr = fmt.Errorf("ReadFile expects 1 arg (filepath)")
						break
					}
					filePathArg := resolvedArgs[0]
					// Security check should happen here, before reading
					absPath, secErr := secureFilePath(filePathArg, cwd)
					if secErr != nil {
						toolErr = fmt.Errorf("ReadFile security/path error: %w", secErr)
						break
					}

					fmt.Printf("      [Sec Check OK] Attempting read from: %s\n", absPath)
					contentBytes, readErr := os.ReadFile(absPath) // Read from absolute path
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
					// contentToWrite is already evaluated, convert to []byte
					contentBytes := []byte(fmt.Sprintf("%v", contentToWrite))
					writeErr := os.WriteFile(absPath, contentBytes, 0644)
					if writeErr != nil {
						toolErr = fmt.Errorf("WriteFile failed for '%s': %w", absPath, writeErr)
					} else {
						fmt.Printf("      [Tool OK] Wrote %d bytes to %s\n", len(contentBytes), absPath)
						// toolResult = "OK" // Already default
					}
				case "SearchSkills": // Mock vector search
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
					}
					type Result struct {
						Path  string
						Score float64
					}
					results := []Result{}
					threshold := 0.5
					for path, storedEmb := range i.vectorIndex {
						score, simErr := cosineSimilarity(queryEmb, storedEmb)
						if simErr != nil {
							fmt.Printf("      [Warn] Cannot compare with '%s': %v\n", path, simErr)
							continue
						}
						if score >= threshold {
							results = append(results, Result{Path: path, Score: score})
							fmt.Printf("        Match: %s (Score: %.4f)\n", path, score)
						}
					}
					// Sort results?
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
					absPath, secErr := secureFilePath(filePathArg, cwd) // Secure the path first
					if secErr != nil {
						toolErr = fmt.Errorf("GitAdd security/path error: %w", secErr)
						break
					}
					fmt.Printf("      [Tool] Running: git add %s\n", absPath) // Use absolute path
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
				case "VectorUpdate": // Mock update
					if len(resolvedArgs) != 1 {
						toolErr = fmt.Errorf("VectorUpdate expects 1 arg (filepath)")
						break
					}
					filePathArg := resolvedArgs[0]
					absPath, secErr := secureFilePath(filePathArg, cwd) // Ensure we index secure paths
					if secErr != nil {
						toolErr = fmt.Errorf("VectorUpdate security/path error: %w", secErr)
						break
					}
					fmt.Printf("      [Tool] Updating vector index for: %s\n", absPath)
					// Read file content to embed, not just path?
					contentBytes, readErr := os.ReadFile(absPath)
					if readErr != nil {
						toolErr = fmt.Errorf("VectorUpdate failed to read file '%s': %w", absPath, readErr)
						break
					}
					textToEmbed := string(contentBytes) // Embed file content
					embedding, embErr := i.GenerateEmbedding(textToEmbed)
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
					fmt.Printf("      [Tool OK] SanitizeFilename result: %q\n", toolResult)
				default:
					toolErr = fmt.Errorf("unknown TOOL '%s'", toolName)
				}
				// Handle tool error
				if toolErr != nil {
					return nil, fmt.Errorf("step %d: TOOL %s execution failed: %w", stepNum+1, toolName, toolErr)
				}
				i.lastCallResult = toolResult // Store result

			} else if target == "LLM" { // --- LLM Call ---
				if len(resolvedArgs) != 1 {
					return nil, fmt.Errorf("step %d: CALL LLM expects 1 prompt arg, got %d", stepNum+1, len(resolvedArgs))
				}
				prompt := resolvedArgs[0] // Already evaluated and converted to string
				fmt.Printf("    [Exec] CALL LLM prompt: %q\n", prompt)

				// Use CallLLMAPI from llm.go (core package)
				response, llmErr := CallLLMAPI(prompt)
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
				// Recursive call handles scope creation/restoration
				callResult, callErr := i.RunProcedure(procToCall, resolvedArgs...)
				if callErr != nil {
					// Use currentProcName member for context
					return nil, fmt.Errorf("step %d: in procedure '%s', CALL to procedure '%s' failed: %w", stepNum+1, i.currentProcName, procToCall, callErr)
				}
				i.lastCallResult = callResult // Store result
				// Logging already done by the called procedure's return path
			}

		case "IF":
			// ** FIX **: Direct assignment since step.Cond is string
			conditionStr := step.Cond
			// Body is stored in Value by the parser
			bodyStr, okBody := step.Value.(string)
			// ** REMOVED **: Redundant okCond check
			if !okBody {
				return nil, fmt.Errorf("step %d: IF body is not a string (internal error)", stepNum+1)
			}

			fmt.Printf("    [Eval] IF Condition: %s\n", conditionStr)
			conditionResult, evalErr := i.evaluateCondition(conditionStr)
			if evalErr != nil {
				fmt.Printf("      [Warn] IF condition evaluation error: %v (treating as false)\n", evalErr)
				conditionResult = false // Treat eval errors as false
			}
			fmt.Printf("      [Eval] Condition Result: %t\n", conditionResult)

			if conditionResult {
				if bodyStr == "" {
					fmt.Printf("      [Skip] IF condition true, but body is empty.\n")
					skipElse = true // Still need to skip potential ELSE
					continue
				}
				fmt.Printf("      [Exec IF Body] %q\n", bodyStr)
				// Parse body string into a step using the now public ParseStep
				bodyStep, parseErr := ParseStep(bodyStr)
				if parseErr != nil {
					return nil, fmt.Errorf("step %d: failed to parse IF body step '%s': %w", stepNum+1, bodyStr, parseErr)
				}
				// Execute the single body step IN THE CURRENT SCOPE
				bodyResult, bodyErr := i.executeSteps([]Step{bodyStep})
				if bodyErr != nil {
					return nil, bodyErr
				} // Propagate error
				if bodyStep.Type == "RETURN" { // Handle return from IF body
					fmt.Printf("    [Exec] Returning from IF block\n")
					return bodyResult, nil
				}
				skipElse = true // Set flag to skip the next ELSE
			}

		case "ELSE":
			// This step only executes if the preceding IF was false (skipElse is false)
			bodyStr, okBody := step.Value.(string)
			if !okBody {
				return nil, fmt.Errorf("step %d: ELSE body is not a string (internal error)", stepNum+1)
			}

			if bodyStr == "" {
				fmt.Printf("      [Skip] ELSE encountered, but body is empty.\n")
				continue
			}
			fmt.Printf("    [Exec ELSE Body] %q\n", bodyStr)
			// Parse body string using the public ParseStep
			bodyStep, parseErr := ParseStep(bodyStr)
			if parseErr != nil {
				return nil, fmt.Errorf("step %d: failed to parse ELSE body step '%s': %w", stepNum+1, bodyStr, parseErr)
			}
			// Execute body step IN CURRENT SCOPE
			bodyResult, bodyErr := i.executeSteps([]Step{bodyStep})
			if bodyErr != nil {
				return nil, bodyErr
			} // Propagate error
			if bodyStep.Type == "RETURN" { // Handle return from ELSE body
				fmt.Printf("    [Exec] Returning from ELSE block\n")
				return bodyResult, nil
			}

		case "RETURN":
			valueExpr, ok := step.Value.(string)
			if !ok {
				return nil, fmt.Errorf("step %d: RETURN value is not a string (internal error)", stepNum+1)
			}
			returnValue := i.evaluateExpression(valueExpr)
			// Logging done by RunProcedure wrapper
			return returnValue, nil // Return evaluated value

		case "FOR": // Basic FOR EACH loop
			loopVar := step.Target
			// Parser stores collection expression in Cond, body string in Value
			// ** FIX **: Direct assignment since step.Cond is string
			collectionExpr := step.Cond
			bodyStr, okBody := step.Value.(string)
			// ** REMOVED **: Redundant okCond check
			if !okBody {
				return nil, fmt.Errorf("step %d: FOR body is not string (internal error)", stepNum+1)
			}

			evaluatedCollectionExpr := i.evaluateExpression(collectionExpr)
			fmt.Printf("    [Exec] FOR EACH %s IN %s (resolved collection: %q)\n", loopVar, collectionExpr, evaluatedCollectionExpr)

			// Convert the evaluated collection to a string for splitting
			collectionStr := fmt.Sprintf("%v", evaluatedCollectionExpr)

			// Simple comma split - consider JSON array parsing later?
			items := strings.Split(collectionStr, ",")
			if collectionStr == "" {
				items = []string{}
			} // Handle empty collection case

			// Parse body step ONCE, if body exists
			var bodyStep Step
			var parseErr error
			if bodyStr != "" {
				bodyStep, parseErr = ParseStep(bodyStr)
				if parseErr != nil {
					return nil, fmt.Errorf("step %d: failed to parse FOR body step '%s': %w", stepNum+1, bodyStr, parseErr)
				}
			} else {
				fmt.Printf("      [Warn] FOR loop body is empty.\n")
				continue // Skip loop execution if body is empty
			}

			fmt.Printf("      [Looping %d items]\n", len(items))
			originalLoopVarValue, loopVarExists := i.variables[loopVar] // Save original value if exists

			for itemNum, item := range items {
				trimmedItem := strings.TrimSpace(item)
				i.variables[loopVar] = trimmedItem // Set loop variable in CURRENT SCOPE
				fmt.Printf("      [Iter %d] SET %s = %q\n", itemNum, loopVar, trimmedItem)

				// Execute body (only if bodyStep was successfully parsed)
				bodyResult, bodyErr := i.executeSteps([]Step{bodyStep})
				if bodyErr != nil { // Restore loop var before returning error
					if loopVarExists {
						i.variables[loopVar] = originalLoopVarValue
					} else {
						delete(i.variables, loopVar)
					}
					return nil, fmt.Errorf("error in FOR loop body (item %d): %w", itemNum, bodyErr)
				}
				if bodyStep.Type == "RETURN" { // Handle return from FOR body
					fmt.Printf("    [Exec] Returning from FOR loop (item %d)\n", itemNum)
					// Restore loop var before returning value
					if loopVarExists {
						i.variables[loopVar] = originalLoopVarValue
					} else {
						delete(i.variables, loopVar)
					}
					return bodyResult, nil
				}
				// TODO: Add BREAK/CONTINUE support?
			}
			// Restore loop variable after loop finishes
			if loopVarExists {
				i.variables[loopVar] = originalLoopVarValue
			} else {
				delete(i.variables, loopVar)
			}
			fmt.Printf("      [Loop Finished]\n")

		case "WHILE":
			// ** FIX **: Direct assignment since step.Cond is string
			conditionStr := step.Cond
			// Body stored in Value
			bodyStr, okBody := step.Value.(string)
			// ** REMOVED **: Redundant okCond check
			if !okBody {
				return nil, fmt.Errorf("step %d: WHILE body is not string (internal error)", stepNum+1)
			}

			fmt.Printf("    [Exec] WHILE %s ...\n", conditionStr)

			// Parse body step ONCE, if body exists
			var bodyStep Step
			var parseErr error
			if bodyStr != "" {
				bodyStep, parseErr = ParseStep(bodyStr)
				if parseErr != nil {
					return nil, fmt.Errorf("step %d: failed to parse WHILE body step '%s': %w", stepNum+1, bodyStr, parseErr)
				}
			} else {
				fmt.Printf("      [Warn] WHILE loop body is empty.\n")
				continue // Treat as no-op if body is empty
			}

			loopCounter := 0
			maxLoops := 1000 // Safety break

			for loopCounter < maxLoops {
				conditionResult, evalErr := i.evaluateCondition(conditionStr)
				if evalErr != nil {
					fmt.Printf("      [Warn] WHILE condition evaluation error: %v (exiting loop)\n", evalErr)
					break // Exit loop on condition evaluation error
				}
				fmt.Printf("      [Loop Check %d] Condition %q -> %t\n", loopCounter, conditionStr, conditionResult)
				if !conditionResult {
					break
				} // Exit loop if condition is false

				// Execute body (only if bodyStep was parsed)
				fmt.Printf("      [Exec WHILE Body %d] %q\n", loopCounter, bodyStr)
				bodyResult, bodyErr := i.executeSteps([]Step{bodyStep})
				if bodyErr != nil {
					return nil, fmt.Errorf("error in WHILE loop body (iteration %d): %w", loopCounter, bodyErr)
				}
				if bodyStep.Type == "RETURN" { // Handle return from WHILE body
					fmt.Printf("    [Exec] Returning from WHILE loop (iteration %d)\n", loopCounter)
					return bodyResult, nil
				}
				loopCounter++
			}
			if loopCounter >= maxLoops {
				return nil, fmt.Errorf("step %d: WHILE loop exceeded max iterations (%d)", stepNum+1, maxLoops)
			}
			fmt.Printf("      [Loop Finished]\n")

		default:
			// This case should ideally not be hit if parser filters empty/comment steps
			fmt.Printf("    [Warn] Unknown or unhandled step type '%s' encountered during execution.\n", step.Type)
		}
	}
	// If the procedure finishes without hitting a RETURN statement
	fmt.Printf("  [Exec] Procedure '%s' reached end without explicit RETURN.\n", i.currentProcName)
	return nil, nil // Default return nil if no explicit RETURN was hit
}
