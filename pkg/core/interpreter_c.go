package core

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	// NOTE: No reflect needed here
)

// evaluateExpression - Central Evaluator. Returns final value (interface{}).
// Handles simple variables (raw), literals/placeholders (string+resolved), and concatenation.
func (i *Interpreter) evaluateExpression(expr string) interface{} {
	trimmedExpr := strings.TrimSpace(expr)

	// --- Check 1: Direct variable/keyword lookup ---
	// Use resolveValue to get raw value ONLY if expr is JUST that var/keyword
	rawValue, found := i.resolveValue(trimmedExpr)
	if found {
		// If it was found directly by resolveValue (simple var or __last_call_result)
		// return the raw value. Placeholders within strings are NOT resolved here.
		fmt.Printf("      [EvalExpr] Simple var/keyword '%s' found, returning raw value (type %T).\n", trimmedExpr, rawValue)
		return rawValue
	}
	// --- END Check 1 ---

	// --- Expression needs string processing ---
	// It's a literal ("abc"), uses placeholders ("{{v}}"), concatenation ("a"+"b"),
	// or an unknown identifier ("xyz").
	// Resolve placeholders throughout the *entire original expression string*.
	resolvedExprStr := i.resolvePlaceholders(trimmedExpr)
	fmt.Printf("      [EvalExpr] Expression '%s' resolved via resolvePlaceholders to: %q\n", trimmedExpr, resolvedExprStr)

	// --- Check for Concatenation on the *placeholder-resolved* string ---
	parts := splitExpression(resolvedExprStr) // Split the potentially modified string

	if len(parts) > 1 { // Potential concatenation
		isPotentialConcat := false
		/* ... existing concatenation pattern check logic on 'parts' ... */
		isValidConcat := true
		if len(parts)%2 == 0 {
			isValidConcat = false
		} else {
			for idx, part := range parts {
				isOperatorPart := (idx%2 == 1)
				if isOperatorPart && part != "+" {
					isValidConcat = false
					break
				}
				if !isOperatorPart && part == "+" {
					isValidConcat = false
					break
				}
			}
		}
		if isValidConcat {
			isPotentialConcat = true
		} else {
			hasPlus := false
			for _, p := range parts {
				if p == "+" {
					hasPlus = true
					break
				}
			}
			if hasPlus {
				fmt.Printf("  [Warn] Expression %q after resolve (parts: %v) looks like invalid concat.\n", resolvedExprStr, parts)
			}
		}

		if isPotentialConcat {
			// Perform concatenation: Evaluate each *part* (which should now be literals or resolved values after the initial resolvePlaceholders)
			// and stringify results for joining.
			var builder strings.Builder
			for idx, part := range parts {
				if idx%2 == 1 {
					continue
				} // Skip '+'

				// Evaluate the *part* recursively. Since resolvePlaceholders ran initially,
				// parts should mostly be literals or already resolved simple values.
				// This handles cases like "a" + ("b" + "c")
				resolvedValue := i.evaluateExpression(part)
				valueToAppendStr := fmt.Sprintf("%v", resolvedValue) // Stringify result for concat

				// Unquote if the *part* was a literal string
				trimmedOriginalPart := strings.TrimSpace(part) // Check the part itself
				isOriginalLiteral := len(trimmedOriginalPart) >= 2 && ((trimmedOriginalPart[0] == '"' && trimmedOriginalPart[len(trimmedOriginalPart)-1] == '"') || (trimmedOriginalPart[0] == '\'' && trimmedOriginalPart[len(trimmedOriginalPart)-1] == '\''))
				if isOriginalLiteral {
					unquoted, err := strconv.Unquote(valueToAppendStr) // Unquote the *stringified* value
					if err == nil {
						valueToAppendStr = unquoted
					} else { /* ... fallback manual strip ... */
						if len(valueToAppendStr) >= 2 && ((valueToAppendStr[0] == '"' && valueToAppendStr[len(valueToAppendStr)-1] == '"') || (valueToAppendStr[0] == '\'' && valueToAppendStr[len(valueToAppendStr)-1] == '\'')) {
							valueToAppendStr = valueToAppendStr[1 : len(valueToAppendStr)-1]
						}
					}
				}
				builder.WriteString(valueToAppendStr)
			}
			finalResult := builder.String()
			fmt.Printf("      [EvalExpr +] Concatenated Result: %q\n", finalResult)
			return finalResult
		}
		// If not valid concatenation, fall through to treat resolvedExprStr as a single unit.
	}

	// --- Treat as Single Unit (Literal or Resolved Placeholder String) ---
	// If it wasn't a simple variable, and wasn't valid concatenation,
	// the result is the placeholder-resolved string `resolvedExprStr`.
	// Perform final unquoting if the *original* expression looked like a quoted literal.
	trimmedOriginalExpr := strings.TrimSpace(expr) // Check original expression
	if len(trimmedOriginalExpr) >= 2 &&
		((trimmedOriginalExpr[0] == '"' && trimmedOriginalExpr[len(trimmedOriginalExpr)-1] == '"') ||
			(trimmedOriginalExpr[0] == '\'' && trimmedOriginalExpr[len(trimmedOriginalExpr)-1] == '\'')) {
		unquoted, err := strconv.Unquote(resolvedExprStr)
		if err == nil {
			fmt.Printf("      [EvalExpr] Unquoted single literal/resolved %q -> %q\n", expr, unquoted)
			return unquoted
		} else { /* ... fallback manual strip ... */
			if len(resolvedExprStr) >= 2 && ((resolvedExprStr[0] == '"' && resolvedExprStr[len(resolvedExprStr)-1] == '"') || (resolvedExprStr[0] == '\'' && resolvedExprStr[len(resolvedExprStr)-1] == '\'')) {
				resolvedExprStr = resolvedExprStr[1 : len(resolvedExprStr)-1]
				fmt.Printf("      [EvalExpr] Manual unquote single literal/resolved %q -> %q\n", expr, resolvedExprStr)
			} else {
				fmt.Printf("      [EvalExpr] Failed unquote single literal/resolved %q -> %q\n", expr, resolvedExprStr)
			}
		}
	}

	fmt.Printf("      [EvalExpr] Final result for single part '%s': %q\n", expr, resolvedExprStr)
	return resolvedExprStr // Return the (potentially unquoted) resolved string
}

// --- Utility Helpers ---

// trimCodeFences removes leading/trailing code fences (``` or ```lang)
func trimCodeFences(code string) string {
	trimmed := strings.TrimSpace(code)
	lines := strings.Split(trimmed, "\n")
	if len(lines) < 1 {
		return code
	}
	firstLineTrimmed := strings.TrimSpace(lines[0])
	startFenceFound := false
	// More general check for ``` optionally followed by language hint
	if strings.HasPrefix(firstLineTrimmed, "```") {
		// Check if it's ONLY ``` or ``` plus non-space chars
		restOfLine := strings.TrimSpace(firstLineTrimmed[3:])
		if len(restOfLine) == 0 || !strings.ContainsAny(restOfLine, " \t") { // Allow ``` or ```lang, but not ``` lang with space
			startFenceFound = true
			lines = lines[1:]
		}
	}
	// if firstLineTrimmed == "```neuroscript" || firstLineTrimmed == "```" { startFenceFound = true; lines = lines[1:] } // Old version
	endFenceFound := false
	if len(lines) > 0 {
		lastLineTrimmed := strings.TrimSpace(lines[len(lines)-1])
		if lastLineTrimmed == "```" {
			endFenceFound = true
			lines = lines[:len(lines)-1]
		}
	}
	if startFenceFound || endFenceFound {
		return strings.TrimSpace(strings.Join(lines, "\n"))
	}
	return trimmed // Return original trimmed if no fences found
}

// sanitizeFilename creates a safe filename component.
func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	// Allow alphanumeric, underscore, hyphen, dot. Remove others.
	removeChars := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	name = removeChars.ReplaceAllString(name, "")
	// Remove leading/trailing dots, underscores, hyphens more carefully
	name = strings.TrimLeft(name, "._-")
	name = strings.TrimRight(name, "._-")
	// Collapse multiple underscores/hyphens/dots
	name = regexp.MustCompile(`_{2,}`).ReplaceAllString(name, "_")
	name = regexp.MustCompile(`-{2,}`).ReplaceAllString(name, "-")
	name = regexp.MustCompile(`\.{2,}`).ReplaceAllString(name, ".") // Avoid .. in middle
	name = strings.ReplaceAll(name, "..", "_")                      // Replace remaining ".." just in case

	const maxLength = 100
	if len(name) > maxLength {
		lastSep := strings.LastIndexAny(name[:maxLength], "_-.")
		if lastSep > maxLength/2 {
			name = name[:lastSep]
		} else {
			name = name[:maxLength]
		}
		name = strings.TrimRight(name, "._-") // Trim again after potential cut
	}
	if name == "" {
		name = "default_skill_name"
	} // Ensure non-empty
	// Avoid OS reserved names (Windows mainly) - simplistic check
	reserved := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "LPT1"}
	upperName := strings.ToUpper(name)
	for _, r := range reserved {
		if upperName == r {
			name = name + "_"
			break
		}
	}

	return name
}

// runGitCommand executes a git command.
func runGitCommand(args ...string) error {
	cmd := exec.Command("git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("git command 'git %s' failed: %v\nStderr: %s", strings.Join(args, " "), err, stderr.String())
	}
	return nil
}

// secureFilePath cleans and ensures the path is within the allowed directory (cwd).
func secureFilePath(filePath, allowedDir string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}
	// Basic check for null bytes
	if strings.Contains(filePath, "\x00") {
		return "", fmt.Errorf("file path contains null byte")
	}

	absAllowedDir, err := filepath.Abs(allowedDir)
	if err != nil {
		return "", fmt.Errorf("could not get absolute path for allowed directory '%s': %w", allowedDir, err)
	}
	absAllowedDir = filepath.Clean(absAllowedDir)

	// Clean the input path itself first to handle relative traversals better
	cleanedInputPath := filepath.Clean(filePath)
	// Prevent absolute paths in the input 'filePath' argument if allowedDir is meant as root
	if filepath.IsAbs(cleanedInputPath) {
		// Allow if it's within allowedDir? Or disallow always? Let's disallow absolute inputs for now.
		return "", fmt.Errorf("input file path '%s' must be relative", filePath)
	}

	joinedPath := filepath.Join(absAllowedDir, cleanedInputPath)

	// Final clean on the joined path
	absCleanedPath := filepath.Clean(joinedPath)

	// Check if the final absolute path is within the allowed directory
	if !strings.HasPrefix(absCleanedPath, absAllowedDir) {
		return "", fmt.Errorf("path '%s' resolves to '%s' which is outside the allowed directory '%s'", filePath, absCleanedPath, absAllowedDir)
	}
	// Additional check: Ensure it's not EXACTLY the allowed dir if filePath wasn't empty
	if absCleanedPath == absAllowedDir && filePath != "." && filePath != "" {
		return "", fmt.Errorf("path '%s' resolves to the allowed directory root '%s'", filePath, absCleanedPath)
	}

	return absCleanedPath, nil // Return the safe, absolute, cleaned path
}

// --- Mock Embeddings ---

// GenerateEmbedding creates a mock deterministic embedding.
func (i *Interpreter) GenerateEmbedding(text string) ([]float32, error) {
	embedding := make([]float32, i.embeddingDim)
	var seed int64
	for _, r := range text {
		seed = (seed*31 + int64(r)) % (1<<31 - 1)
	}
	rng := rand.New(rand.NewSource(seed))
	norm := float32(0.0)
	for d := 0; d < i.embeddingDim; d++ {
		val := rng.Float32()*2 - 1
		embedding[d] = val
		norm += val * val
	}
	norm = float32(math.Sqrt(float64(norm)))
	if norm > 1e-6 {
		for d := range embedding {
			embedding[d] /= norm
		}
	}
	return embedding, nil
}

// cosineSimilarity calculates similarity between two vectors.
func cosineSimilarity(v1, v2 []float32) (float64, error) {
	if len(v1) == 0 || len(v2) == 0 {
		return 0, fmt.Errorf("vectors cannot be empty")
	}
	if len(v1) != len(v2) {
		return 0, fmt.Errorf("vector dimensions mismatch (%d vs %d)", len(v1), len(v2))
	}
	dotProduct := 0.0
	norm1 := 0.0
	norm2 := 0.0
	for i := range v1 {
		dotProduct += float64(v1[i] * v2[i])
		norm1 += float64(v1[i] * v1[i])
		norm2 += float64(v2[i] * v2[i])
	}
	mag1 := math.Sqrt(norm1)
	mag2 := math.Sqrt(norm2)
	if mag1 < 1e-9 || mag2 < 1e-9 {
		if mag1 < 1e-9 && mag2 < 1e-9 {
			return 1.0, nil
		}
		return 0.0, nil
	}
	similarity := dotProduct / (mag1 * mag2)
	if similarity > 1.0 {
		similarity = 1.0
	} else if similarity < -1.0 {
		similarity = -1.0
	}
	return similarity, nil
}
