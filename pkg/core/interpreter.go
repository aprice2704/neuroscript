package core

import (
	"bytes"         // For exec command output/error
	"encoding/json" // For potentially saving/loading index later
	"fmt"
	"math"      // For cosine similarity
	"math/rand" // For dummy embeddings
	"os"
	"os/exec" // For running Git commands
	"regexp"
	"strings"
)

// Interpreter holds the state during execution
type Interpreter struct {
	variables       map[string]interface{}
	knownProcedures map[string]Procedure
	lastCallResult  interface{}
	vectorIndex     map[string][]float32 // Simple in-memory index: filepath -> embedding
	embeddingDim    int                  // Dimension for dummy embeddings
}

// NewInterpreter creates a new interpreter instance
func NewInterpreter() *Interpreter {
	// Initialize with an empty vector index
	// We could potentially load an index from a file here later
	return &Interpreter{
		variables:       make(map[string]interface{}),
		knownProcedures: make(map[string]Procedure),
		vectorIndex:     make(map[string][]float32),
		embeddingDim:    16, // Small dimension for dummy embeddings
	}
}

// LoadProcedures adds procedures to the interpreter's known set
func (i *Interpreter) LoadProcedures(procs []Procedure) error { /* ... unchanged ... */
	for _, p := range procs {
		if _, exists := i.knownProcedures[p.Name]; exists {
			fmt.Printf("[Warning] Reloading procedure: %s\n", p.Name)
		}
		i.knownProcedures[p.Name] = p
	}
	return nil
}

// RunProcedure executes a given procedure by name, mapping args
func (i *Interpreter) RunProcedure(procName string, args ...string) (interface{}, error) { /* ... unchanged ... */
	proc, exists := i.knownProcedures[procName]
	if !exists {
		return nil, fmt.Errorf("procedure '%s' not defined", procName)
	}
	localVars := make(map[string]interface{})
	if len(args) != len(proc.Params) {
		return nil, fmt.Errorf("proc '%s' expects %d args (%v), got %d", procName, len(proc.Params), proc.Params, len(args))
	}
	for idx, paramName := range proc.Params {
		localVars[paramName] = args[idx]
		fmt.Printf("  [Arg Init] SET %s = %q\n", paramName, args[idx])
	}
	fmt.Printf("[Exec] Running Procedure: %s\n", procName)
	originalVars := i.variables
	i.variables = localVars
	result, err := i.executeSteps(proc.Steps)
	i.variables = originalVars
	return result, err
}

// --- Vector / Embedding Helpers (Phase 3 Stubs/Placeholders) ---

// GenerateEmbedding creates a dummy embedding for text. Replace with real model later.
func (i *Interpreter) GenerateEmbedding(text string) ([]float32, error) {
	// Very simple placeholder: fixed vector based on text length modulo dimensions
	// OR just random vector for now
	embedding := make([]float32, i.embeddingDim)
	// Use a deterministic seed based on text? No, just random for now.
	// Seed with time maybe? Or constant seed for predictability? Constant.
	rng := rand.New(rand.NewSource(int64(len(text)))) // Seed based on text length
	for d := 0; d < i.embeddingDim; d++ {
		embedding[d] = rng.Float32()*2 - 1 // Random value between -1 and 1
	}
	// Normalize the vector (important for cosine similarity)
	norm := float32(0.0)
	for _, val := range embedding {
		norm += val * val
	}
	norm = float32(math.Sqrt(float64(norm)))
	if norm > 0 {
		for d := range embedding {
			embedding[d] /= norm
		}
	}
	// fmt.Printf("      [Embed] Generated dummy embedding for text (len %d)\n", len(text))
	return embedding, nil
}

// cosineSimilarity calculates similarity between two vectors. Assumes normalized vectors.
func cosineSimilarity(v1, v2 []float32) (float64, error) {
	if len(v1) != len(v2) {
		return 0, fmt.Errorf("vector dimensions mismatch (%d vs %d)", len(v1), len(v2))
	}
	dotProduct := float64(0.0)
	for i := range v1 {
		dotProduct += float64(v1[i] * v2[i])
	}
	// For already normalized vectors, dot product is the cosine similarity
	return dotProduct, nil
}

// --- Git Command Helper ---
func runGitCommand(args ...string) error {
	cmd := exec.Command("git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run() // Use Run instead of Output/CombinedOutput as we don't need stdout
	if err != nil {
		return fmt.Errorf("git command 'git %s' failed: %v\nStderr: %s", strings.Join(args, " "), err, stderr.String())
	}
	return nil
}

// executeSteps handles the step-by-step execution logic (Updated CALL TOOL section)
func (i *Interpreter) executeSteps(steps []Step) (interface{}, error) { /* ... SET logic unchanged ... */
	for stepNum, step := range steps {
		fmt.Printf("  [Step %d] %s %s ...\n", stepNum+1, step.Type, step.Target)
		switch step.Type {
		case "SET":
			targetVar := step.Target
			valueToSet := ""
			if valStr, ok := step.Value.(string); ok {
				if valStr == "__last_call_result" {
					if i.lastCallResult != nil {
						valueToSet = fmt.Sprintf("%v", i.lastCallResult)
					} else {
						fmt.Printf("  [Warn] SET using __last_call_result before CALL\n")
						valueToSet = ""
					}
					i.lastCallResult = nil
				} else {
					valueToSet = i.resolveVariables(valStr)
				}
			} else {
				valueToSet = fmt.Sprintf("%v", step.Value)
			}
			i.variables[targetVar] = valueToSet
			logValue := valueToSet
			if len(logValue) > 70 {
				logValue = logValue[:67] + "..."
			}
			fmt.Printf("    [Exec] SET %s = %q\n", targetVar, logValue)
		case "CALL":
			i.lastCallResult = nil
			resolvedArgs := make([]string, len(step.Args))
			for idx, arg := range step.Args {
				resolvedArgs[idx] = i.resolveVariables(arg)
			}
			if strings.HasPrefix(step.Target, "TOOL.") {
				toolName := strings.TrimPrefix(step.Target, "TOOL.")
				var err error
				var toolResult interface{} = nil
				fmt.Printf("    [Exec] CALL TOOL: %s(%v)\n", toolName, resolvedArgs)
				switch toolName {
				// --- Existing Tools ---
				case "ReadFile":
					if len(resolvedArgs) != 1 {
						err = fmt.Errorf("ReadFile expects 1 arg (filepath)")
					} else {
						filePath := resolvedArgs[0]
						contentBytes, readErr := os.ReadFile(filePath)
						if readErr != nil {
							err = fmt.Errorf("ReadFile failed for '%s': %w", filePath, readErr)
						} else {
							toolResult = string(contentBytes)
							fmt.Printf("      [Tool OK] Read %d bytes from %s\n", len(contentBytes), filePath)
						}
					}
				case "WriteFile":
					if len(resolvedArgs) != 2 {
						err = fmt.Errorf("WriteFile expects 2 args (filepath, content)")
					} else {
						filePath, content := resolvedArgs[0], resolvedArgs[1]
						writeErr := os.WriteFile(filePath, []byte(content), 0644)
						if writeErr != nil {
							err = fmt.Errorf("WriteFile failed for '%s': %w", filePath, writeErr)
						} else {
							fmt.Printf("      [Tool OK] Wrote %d bytes to %s\n", len(content), filePath)
						}
					}
				// --- Updated Git/Vector Tools (Phase 3) ---
				case "SearchSkills":
					if len(resolvedArgs) != 1 {
						err = fmt.Errorf("SearchSkills expects 1 arg (query)")
					} else {
						query := resolvedArgs[0]
						fmt.Printf("      [Tool] Searching skills for: %q\n", query)
						queryEmb, embErr := i.GenerateEmbedding(query)
						if embErr != nil {
							err = fmt.Errorf("failed to generate query embedding: %w", embErr)
							break
						}

						type Result struct {
							Path  string
							Score float64
						}
						var results []Result
						threshold := 0.5 // Example similarity threshold

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
						// Return results as a JSON string for simplicity in NeuroScript
						resultBytes, jsonErr := json.Marshal(results)
						if jsonErr != nil {
							err = fmt.Errorf("failed to marshal search results: %w", jsonErr)
							break
						}
						toolResult = string(resultBytes)
						fmt.Printf("      [Tool OK] SearchSkills returning: %s\n", toolResult)
					}
				case "GitAdd":
					if len(resolvedArgs) != 1 {
						err = fmt.Errorf("GitAdd expects 1 arg (filepath)")
					} else {
						filePath := resolvedArgs[0]
						fmt.Printf("      [Tool] Running: git add %s\n", filePath)
						err = runGitCommand("add", filePath) // Execute actual git command
						if err == nil {
							fmt.Printf("      [Tool OK] GitAdd\n")
							toolResult = "OK"
						}
					}
				case "GitCommit":
					if len(resolvedArgs) != 1 {
						err = fmt.Errorf("GitCommit expects 1 arg (message)")
					} else {
						message := resolvedArgs[0]
						fmt.Printf("      [Tool] Running: git commit -m %q\n", message)
						err = runGitCommand("commit", "-m", message) // Execute actual git command
						if err == nil {
							fmt.Printf("      [Tool OK] GitCommit\n")
							toolResult = "OK"
						}
					}
				case "VectorUpdate":
					if len(resolvedArgs) != 1 {
						err = fmt.Errorf("VectorUpdate expects 1 arg (filepath)")
					} else {
						filePath := resolvedArgs[0]
						fmt.Printf("      [Tool] Updating vector index for: %s\n", filePath)
						// Read file content (or relevant parts) to generate embedding
						// In reality, parse the file to get PURPOSE etc. For MVP, use whole content? Or just filename?
						// Let's use filename + path for simplicity for now.
						textToEmbed := filePath // Use file path as proxy for content for MVP
						embedding, embErr := i.GenerateEmbedding(textToEmbed)
						if embErr != nil {
							err = fmt.Errorf("failed to generate embedding for '%s': %w", filePath, embErr)
							break
						}
						i.vectorIndex[filePath] = embedding // Store in map
						toolResult = "OK"
						fmt.Printf("      [Tool OK] VectorUpdate - Index size: %d\n", len(i.vectorIndex))
					}
				default:
					err = fmt.Errorf("unknown TOOL '%s'", toolName)
				}
				if err != nil {
					return nil, fmt.Errorf("step %d: TOOL %s: %w", stepNum+1, toolName, err)
				}
				i.lastCallResult = toolResult

			} else if step.Target == "LLM" { /* ... LLM Call unchanged ... */
				if len(resolvedArgs) != 1 {
					return nil, fmt.Errorf("step %d: CALL LLM expects 1 prompt arg, got %d", stepNum+1, len(resolvedArgs))
				}
				prompt := resolvedArgs[0]
				fmt.Printf("    [Exec] CALL LLM prompt: %q\n", prompt)
				response, llmErr := CallLLMAPI(prompt)
				if llmErr != nil {
					return nil, fmt.Errorf("step %d: CALL LLM failed: %w", stepNum+1, llmErr)
				}
				i.lastCallResult = response
				fmt.Printf("      [LLM OK] Response stored\n")
			} else { /* ... Procedure Call unchanged ... */
				procToCall := step.Target
				fmt.Printf("    [Exec] CALL Procedure: %s(%v)\n", procToCall, resolvedArgs)
				callResult, callErr := i.RunProcedure(procToCall, resolvedArgs...)
				if callErr != nil {
					return nil, fmt.Errorf("step %d: CALL Proc '%s' failed: %w", stepNum+1, procToCall, callErr)
				}
				i.lastCallResult = callResult
				fmt.Printf("      [Call OK] Proc '%s' returned, result stored\n", procToCall)
			}
		case "RETURN": /* ... unchanged ... */
			returnValue := ""
			if retValStr, ok := step.Value.(string); ok {
				if retValStr == "__last_call_result" {
					if i.lastCallResult != nil {
						returnValue = fmt.Sprintf("%v", i.lastCallResult)
					}
				} else {
					returnValue = i.resolveVariables(retValStr)
				}
			} else {
				returnValue = fmt.Sprintf("%v", step.Value)
			}
			fmt.Printf("    [Exec] RETURN %q\n", returnValue)
			return returnValue, nil
		default:
		}
	}
	return nil, nil // Finished steps without RETURN
}

func (i *Interpreter) resolveVariables(input string) string { /* ... unchanged ... */
	resolved := input
	re := regexp.MustCompile(`\{\{([a-zA-Z_][a-zA-Z0-9_]*)\}\}`)
	isSimpleVar := !strings.Contains(input, "{{") && !strings.Contains(input, "}}") && re.MatchString("^"+input+"$")
	if val, exists := i.variables[input]; exists && isSimpleVar {
		return fmt.Sprintf("%v", val)
	}
	matches := re.FindAllStringSubmatch(resolved, -1)
	for _, match := range matches {
		if len(match) == 2 {
			varName := match[1]
			placeholder := match[0]
			if value, exists := i.variables[varName]; exists {
				resolved = strings.ReplaceAll(resolved, placeholder, fmt.Sprintf("%v", value))
			} else {
				fmt.Printf("  [Warn] Variable '%s' not found for substitution in %q\n", varName, input)
			}
		}
	}
	return resolved
}

// ... (Keep llm.go content unchanged) ...
