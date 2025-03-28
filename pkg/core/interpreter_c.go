package core

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// evaluateExpression - Central Evaluator. Returns final value (interface{}).
// ** FIX: Ensure placeholders within variable values are resolved **
func (i *Interpreter) evaluateExpression(expr string) interface{} {
	trimmedExpr := strings.TrimSpace(expr)
	// Handle __last_call_result directly for efficiency
	if trimmedExpr == "__last_call_result" {
		if i.lastCallResult != nil {
			return i.lastCallResult
		}
		fmt.Printf("  [Warn] Evaluating __last_call_result before CALL\n")
		return "" // Return empty string if not set
	}

	parts := splitExpression(trimmedExpr) // Use trimmed expression for splitting

	// If only one part after splitting, evaluate that part directly
	if len(parts) == 1 {
		singlePart := parts[0]
		// Resolve var/placeholder/literal using resolveValue
		resolvedValue := i.resolveValue(singlePart)

		// ** FIX **: Now, resolve placeholders *within* the resolved value if it's a string
		resolvedValueStr, isStr := resolvedValue.(string)
		if isStr {
			finalStrValue := i.resolvePlaceholders(resolvedValueStr) // Resolve internal placeholders

			// If the original single part was a quoted literal, unquote the final resolved string
			trimmedOriginalPart := strings.TrimSpace(singlePart) // Use the part *before* resolveValue
			if len(trimmedOriginalPart) >= 2 &&
				((trimmedOriginalPart[0] == '"' && trimmedOriginalPart[len(trimmedOriginalPart)-1] == '"') ||
					(trimmedOriginalPart[0] == '\'' && trimmedOriginalPart[len(trimmedOriginalPart)-1] == '\'')) {

				unquoted, err := strconv.Unquote(finalStrValue) // Attempt to unquote the final string
				if err == nil {
					fmt.Printf("      [Eval] Unquoted single literal %q -> %q\n", singlePart, unquoted)
					return unquoted // Return unquoted string
				} else {
					// Fallback manual strip
					if len(finalStrValue) >= 2 && ((finalStrValue[0] == '"' && finalStrValue[len(finalStrValue)-1] == '"') || (finalStrValue[0] == '\'' && finalStrValue[len(finalStrValue)-1] == '\'')) {
						manualUnquote := finalStrValue[1 : len(finalStrValue)-1]
						fmt.Printf("      [Eval] Manually unquoted single literal %q -> %q\n", singlePart, manualUnquote)
						return manualUnquote
					}
					fmt.Printf("      [Eval] Failed to unquote single literal %q, using resolved value: %q\n", singlePart, finalStrValue)
					return finalStrValue // Return final resolved (potentially still quoted) string
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
	// (Concatenation logic remains the same, relies on resolved values from resolveValue+resolvePlaceholders)
	isPotentialConcat := false
	if len(parts) > 1 {
		isValidConcat := true
		if len(parts)%2 == 0 { // Must be odd number of parts (value, +, value, ...)
			isValidConcat = false
		} else {
			for idx, part := range parts {
				isOperatorPart := (idx%2 == 1)
				if isOperatorPart && part != "+" {
					isValidConcat = false // Operator part must be '+'
					break
				}
				if !isOperatorPart && part == "+" {
					isValidConcat = false // Value part cannot be '+'
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
				fmt.Printf("  [Warn] Expression %q (parts: %v) looks like invalid concatenation pattern.\n", expr, parts)
			}
			// Fallback: treat original expr as single value
			resolvedValue := i.resolveValue(expr) // Get potentially unresolved value
			resolvedValueStr, isStr := resolvedValue.(string)
			finalStr := ""
			if isStr { // Resolve placeholders if it's a string
				finalStr = i.resolvePlaceholders(resolvedValueStr)
			} else {
				finalStr = fmt.Sprintf("%v", resolvedValue)
			}
			fmt.Printf("      [Eval] Invalid concat pattern in %q, resolving whole expr -> %v\n", expr, finalStr)
			return finalStr // Return placeholder-resolved string
		}
	}

	if isPotentialConcat { // --- Concatenation Logic ---
		var builder strings.Builder
		fmt.Printf("      [Eval +] Attempting concat: %v\n", parts)
		for idx, part := range parts {
			if idx%2 == 1 {
				continue
			} // Skip '+' operator

			resolvedValue := i.resolveValue(part) // Resolves var/placeholder/literal

			// Resolve placeholders *within* the resolved value if it's a string
			valueToAppendStr, isStr := resolvedValue.(string)
			if isStr {
				valueToAppendStr = i.resolvePlaceholders(valueToAppendStr)
			} else {
				valueToAppendStr = fmt.Sprintf("%v", resolvedValue) // Convert non-string to string
			}

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

	// --- Fallback: Should not be reached if logic above is complete ---
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
// (trimCodeFences, sanitizeFilename, runGitCommand, secureFilePath, GenerateEmbedding, cosineSimilarity unchanged)

// trimCodeFences removes leading/trailing code fences (``` or ```neuroscript)
func trimCodeFences(code string) string {
	trimmed := strings.TrimSpace(code)
	lines := strings.Split(trimmed, "\n")
	if len(lines) < 1 {
		return code // Should not happen with TrimSpace, but safer
	}

	// Check first line for start fence
	firstLineTrimmed := strings.TrimSpace(lines[0])
	startFenceFound := false
	if firstLineTrimmed == "```neuroscript" || firstLineTrimmed == "```" {
		startFenceFound = true
		lines = lines[1:] // Remove first line
	}

	// Check last line for end fence
	endFenceFound := false
	if len(lines) > 0 {
		lastLineTrimmed := strings.TrimSpace(lines[len(lines)-1])
		if lastLineTrimmed == "```" {
			endFenceFound = true
			lines = lines[:len(lines)-1] // Remove last line
		}
	}

	// Only return modified string if at least one fence was found
	if startFenceFound || endFenceFound {
		// Rejoin the remaining lines and trim potential whitespace again
		return strings.TrimSpace(strings.Join(lines, "\n"))
	}

	// If no fences seemed present, return original trimmed string
	return trimmed
}

// sanitizeFilename creates a safe filename component.
func sanitizeFilename(name string) string {
	// 1. Replace common separators with underscore
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")

	// 2. Remove characters invalid in most filesystems
	// Keep '.', allow '_' and '-'
	removeChars := regexp.MustCompile(`[<>:"|?*']`) // Removed / \ space
	name = removeChars.ReplaceAllString(name, "")

	// 3. Remove leading/trailing dots or underscores/hyphens
	name = strings.Trim(name, "._-")

	// 4. Collapse multiple underscores/hyphens
	name = regexp.MustCompile(`_{2,}`).ReplaceAllString(name, "_")
	name = regexp.MustCompile(`-{2,}`).ReplaceAllString(name, "-")

	// 5. Basic check for path traversal attempts - not foolproof
	name = strings.ReplaceAll(name, "..", "") // Remove ".." sequences

	// 6. Limit length
	const maxLength = 100
	if len(name) > maxLength {
		// Find last valid character boundary if possible (e.g., underscore)
		lastSep := strings.LastIndexAny(name[:maxLength], "_-.")
		if lastSep > maxLength/2 { // Only truncate at boundary if reasonable
			name = name[:lastSep]
		} else {
			name = name[:maxLength]
		}
		name = strings.Trim(name, "._-") // Trim again after potential cut
	}

	// 7. Ensure non-empty/dangerous name
	if name == "" || name == "." || name == ".." {
		name = "default_skill_name"
	}
	return name
}

// runGitCommand executes a git command.
func runGitCommand(args ...string) error {
	cmd := exec.Command("git", args...)
	// Prevent git commands from running outside CWD for safety?
	// cmd.Dir = "." // Explicitly run in current dir
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

	// Ensure allowedDir is absolute and clean first
	absAllowedDir, err := filepath.Abs(allowedDir)
	if err != nil {
		return "", fmt.Errorf("could not get absolute path for allowed directory '%s': %w", allowedDir, err)
	}
	absAllowedDir = filepath.Clean(absAllowedDir)

	// Join the allowed dir with the potentially relative filePath
	joinedPath := filepath.Join(absAllowedDir, filePath)

	// Clean the potentially unsafe joined path (resolves "..")
	cleanedPath := filepath.Clean(joinedPath)

	// Check if the cleaned, absolute path starts with the allowed directory + separator
	// Or if it *is* the allowed directory itself. This prevents escaping via "..".
	if !strings.HasPrefix(cleanedPath, absAllowedDir+string(os.PathSeparator)) && cleanedPath != absAllowedDir {
		return "", fmt.Errorf("path '%s' resolves to '%s' which is outside the allowed directory '%s'", filePath, cleanedPath, absAllowedDir)
	}

	return cleanedPath, nil // Return the safe, absolute, cleaned path
}

// --- Mock Embeddings --- (Keep for testing if needed)

// GenerateEmbedding creates a mock deterministic embedding.
func (i *Interpreter) GenerateEmbedding(text string) ([]float32, error) {
	// Simple mock embedding based on text length and random numbers
	embedding := make([]float32, i.embeddingDim)
	// Use a fixed seed derived from text for deterministic mock results
	var seed int64
	for _, r := range text {
		seed = (seed*31 + int64(r)) % (1<<31 - 1) // Simple hash for seed
	}
	rng := rand.New(rand.NewSource(seed))

	norm := float32(0.0)
	for d := 0; d < i.embeddingDim; d++ {
		val := rng.Float32()*2 - 1 // Random value between -1 and 1
		embedding[d] = val
		norm += val * val
	}
	norm = float32(math.Sqrt(float64(norm)))
	if norm > 1e-6 { // Normalize vector, avoid division by zero
		for d := range embedding {
			embedding[d] /= norm
		}
	} else {
		// Handle zero vector case if necessary (e.g., return error or a default vector)
		// For mock, just leave it as near-zero
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
		// Handle zero vectors - similarity is undefined or 0? Return 0.
		if mag1 < 1e-9 && mag2 < 1e-9 {
			return 1.0, nil // Or 0.0? Let's say two zero vectors are perfectly similar.
		}
		return 0.0, nil // One zero vector, similarity is 0.
	}

	similarity := dotProduct / (mag1 * mag2)

	// Clamp result to [-1, 1] due to potential floating point inaccuracies
	if similarity > 1.0 {
		similarity = 1.0
	} else if similarity < -1.0 {
		similarity = -1.0
	}

	return similarity, nil
}
