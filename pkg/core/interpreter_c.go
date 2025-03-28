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
// Handles single parts, concatenation. Assumes splitExpression is in interpreter_b.go
func (i *Interpreter) evaluateExpression(expr string) interface{} {
	trimmedExpr := strings.TrimSpace(expr)
	// Handle __last_call_result directly for efficiency if used as whole expression
	if trimmedExpr == "__last_call_result" {
		if i.lastCallResult != nil {
			return i.lastCallResult
		}
		fmt.Printf("  [Warn] Evaluating __last_call_result before CALL\n")
		return ""
	}

	parts := splitExpression(trimmedExpr) // Use splitExpression from interpreter_b.go

	// --- Handle Single Part ---
	if len(parts) == 1 {
		singlePart := parts[0]
		// Resolve var/placeholder/literal using resolveValue (from interpreter_b.go)
		resolvedValue := i.resolveValue(singlePart)

		// Resolve placeholders *within* the resolved value if it's a string
		resolvedValueStr, isStr := resolvedValue.(string)
		if isStr {
			finalStrValue := i.resolvePlaceholders(resolvedValueStr) // resolvePlaceholders from interpreter_b.go

			// If the original single part was a quoted literal, unquote the final resolved string
			trimmedOriginalPart := strings.TrimSpace(singlePart)
			if len(trimmedOriginalPart) >= 2 &&
				((trimmedOriginalPart[0] == '"' && trimmedOriginalPart[len(trimmedOriginalPart)-1] == '"') ||
					(trimmedOriginalPart[0] == '\'' && trimmedOriginalPart[len(trimmedOriginalPart)-1] == '\'')) {

				unquoted, err := strconv.Unquote(finalStrValue)
				if err == nil {
					fmt.Printf("      [Eval] Unquoted single literal %q -> %q\n", singlePart, unquoted)
					return unquoted
				} else { // Fallback manual strip if Unquote fails (e.g., internal invalid escapes)
					if len(finalStrValue) >= 2 && ((finalStrValue[0] == '"' && finalStrValue[len(finalStrValue)-1] == '"') || (finalStrValue[0] == '\'' && finalStrValue[len(finalStrValue)-1] == '\'')) {
						manualUnquote := finalStrValue[1 : len(finalStrValue)-1]
						fmt.Printf("      [Eval] Manually unquoted single literal %q -> %q\n", singlePart, manualUnquote)
						return manualUnquote
					}
					// If it wasn't actually quoted after resolution, return as is
					fmt.Printf("      [Eval] Failed to unquote single literal %q, using resolved value: %q\n", singlePart, finalStrValue)
					return finalStrValue
				}
			}
			// If original wasn't quoted literal, return the placeholder-resolved string
			fmt.Printf("      [Eval] Single part %q (var/placeholder/lit) resolved to: %q\n", singlePart, finalStrValue)
			return finalStrValue
		}
		// If not a string after initial resolveValue (e.g., direct __last_call_result was non-string)
		fmt.Printf("      [Eval] Single part %q resolved to non-string: %v (type %T)\n", singlePart, resolvedValue, resolvedValue)
		return resolvedValue // Return the non-string value as is
	}

	// --- Check for valid concatenation pattern (value + value + ...) ---
	isPotentialConcat := false
	if len(parts) > 1 {
		isValidConcat := true
		if len(parts)%2 == 0 { // Must be odd number of parts (value, +, value, ...)
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
				} // Value part cannot be '+'
			}
		}
		if isValidConcat {
			isPotentialConcat = true
		} else {
			// Check if '+' exists at all to decide if warning is needed
			hasPlus := false
			for _, p := range parts {
				if p == "+" {
					hasPlus = true
					break
				}
			}
			if hasPlus {
				fmt.Printf("  [Warn] Expression %q (parts: %v) looks like invalid concatenation pattern.\n", expr, parts)
			}

			// --- Fallback for invalid patterns ---
			// Resolve the original expression as a single literal/variable/placeholder string
			resolvedValue := i.resolveValue(expr) // Uses resolveValue from _b.go
			resolvedValueStr, isStr := resolvedValue.(string)
			finalStr := ""
			if isStr {
				finalStr = i.resolvePlaceholders(resolvedValueStr)
			} else {
				finalStr = fmt.Sprintf("%v", resolvedValue)
			} // Uses resolvePlaceholders from _b.go
			fmt.Printf("      [Eval] Invalid concat/pattern in %q, resolving whole expr -> %v\n", expr, finalStr)
			return finalStr
		}
	}

	// --- Concatenation Logic ---
	if isPotentialConcat {
		var builder strings.Builder
		fmt.Printf("      [Eval +] Attempting concat: %v\n", parts)
		for idx, part := range parts {
			if idx%2 == 1 {
				continue
			} // Skip '+' operator

			resolvedValue := i.resolveValue(part) // Resolve var/placeholder/literal using func from _b.go

			// Resolve placeholders *within* the resolved value if it's a string
			valueToAppendStr, isStr := resolvedValue.(string)
			if isStr {
				valueToAppendStr = i.resolvePlaceholders(valueToAppendStr)
			} else {
				valueToAppendStr = fmt.Sprintf("%v", resolvedValue)
			} // Convert non-string to string

			// Check if the original 'part' was a quoted literal
			trimmedOriginalPart := strings.TrimSpace(part)
			isOriginalLiteral := false
			if len(trimmedOriginalPart) >= 2 &&
				((trimmedOriginalPart[0] == '"' && trimmedOriginalPart[len(trimmedOriginalPart)-1] == '"') ||
					(trimmedOriginalPart[0] == '\'' && trimmedOriginalPart[len(trimmedOriginalPart)-1] == '\'')) {
				isOriginalLiteral = true
			}

			if isOriginalLiteral {
				// If original was literal, unquote the final resolved value string before appending
				unquoted, err := strconv.Unquote(valueToAppendStr)
				if err == nil {
					valueToAppendStr = unquoted
				} else { // Fallback manual strip
					if len(valueToAppendStr) >= 2 && ((valueToAppendStr[0] == '"' && valueToAppendStr[len(valueToAppendStr)-1] == '"') || (valueToAppendStr[0] == '\'' && valueToAppendStr[len(valueToAppendStr)-1] == '\'')) {
						valueToAppendStr = valueToAppendStr[1 : len(valueToAppendStr)-1]
					}
				}
				fmt.Printf("        > Appending literal part %q -> %q\n", part, valueToAppendStr)
			} else {
				fmt.Printf("        > Appending var/placeholder part %q -> %q\n", part, valueToAppendStr)
			}
			builder.WriteString(valueToAppendStr)
		}
		finalResult := builder.String()
		fmt.Printf("      [Eval +] Concatenated Result: %q\n", finalResult)
		return finalResult // Return concatenated string
	}

	// --- Fallback: Should not be easily reachable now ---
	fmt.Printf("      [Eval] Fallback: No valid pattern found for: %q\n", expr)
	resolvedValue := i.resolveValue(expr)
	resolvedValueStr, isStr := resolvedValue.(string)
	finalStr := ""
	if isStr {
		finalStr = i.resolvePlaceholders(resolvedValueStr)
	} else {
		finalStr = fmt.Sprintf("%v", resolvedValue)
	}
	return finalStr
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
