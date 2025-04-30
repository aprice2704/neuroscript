// filename: pkg/core/utils.go
package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/term" // Keep for ReadPassword function
)

// Function to read file content, used by various tools
// FIX: Renamed to ReadFileContent to export it
func ReadFileContent(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading file %s: %w", path, err)
	}
	// Basic check for binary content (presence of null bytes)
	if bytes.Contains(content, []byte{0}) {
		// Handle binary file or return error - currently returns error
		return "", fmt.Errorf("file %s appears to be binary or contains null bytes: %w", path, ErrSkippedBinaryFile)

	}
	return string(content), nil
}

// ReadPassword securely reads a password from the terminal.
func ReadPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	fmt.Println() // Add a newline after password input
	return string(bytePassword), nil
}

// isValidIdentifier checks if a string is a valid NeuroScript identifier.
// Updated to reflect current keyword list (lowercase) and rules.
func isValidIdentifier(name string) bool {
	if name == "" {
		return false
	}
	// Check if the name matches any known keyword (case-insensitive)
	lowerName := strings.ToLower(name)
	keywords := map[string]bool{
		"define":      true, // DEPRECATED: remove if truly gone
		"procedure":   true, // DEPRECATED: remove if truly gone
		"end":         true, // DEPRECATED: remove if truly gone
		"func":        true,
		"endfunc":     true,
		"needs":       true,
		"optional":    true,
		"returns":     true,
		"means":       true,
		"set":         true,
		"call":        true,
		"return":      true,
		"emit":        true,
		"fail":        true,
		"if":          true,
		"then":        true, // DEPRECATED? Check grammar
		"else":        true,
		"endif":       true,
		"while":       true,
		"endwhile":    true,
		"do":          true, // DEPRECATED? Check grammar
		"for":         true,
		"each":        true, // DEPRECATED? Check grammar
		"in":          true,
		"endfor":      true,
		"try":         true, // DEPRECATED by on_error? Check grammar
		"catch":       true, // DEPRECATED by on_error? Check grammar
		"finally":     true, // DEPRECATED by on_error? Check grammar
		"endtry":      true, // DEPRECATED by on_error? Check grammar
		"on_error":    true,
		"endon":       true,
		"clear_error": true,
		"must":        true,
		"mustbe":      true,
		"tool":        true,
		"llm":         true, // DEPRECATED? Check grammar
		"last":        true,
		"eval":        true, // DEPRECATED? Check grammar
		"true":        true,
		"false":       true,
		"nil":         true,
		"and":         true,
		"or":          true,
		"not":         true,
		"no":          true,
		"some":        true,
		// Built-in functions are NOT keywords
	}
	if keywords[lowerName] {
		return false // It's a keyword
	}

	// Check character rules
	for idx, r := range name {
		if idx == 0 {
			// Must start with a letter or underscore
			if !unicode.IsLetter(r) && r != '_' {
				return false
			}
		} else {
			// Subsequent characters can be letters, digits, or underscores
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				return false
			}
		}
	}

	return true // Passed keyword and character checks
}

// normalizePath resolves ., .., and symlinks, returning an absolute path.
// It checks if the final path is within the allowed sandbox directory.
func normalizePath(sandboxDir, inputPath string) (string, error) {
	if !filepath.IsAbs(sandboxDir) {
		var err error
		sandboxDir, err = filepath.Abs(sandboxDir)
		if err != nil {
			return "", fmt.Errorf("failed to make sandbox path absolute: %w", err)
		}
	}
	sandboxDir = filepath.Clean(sandboxDir) // Clean the sandbox path itself

	absPath := inputPath
	if !filepath.IsAbs(inputPath) {
		absPath = filepath.Join(sandboxDir, inputPath)
	}

	// Clean the path (resolves ., .., etc.)
	cleanedPath := filepath.Clean(absPath)

	// Check if the cleaned path starts with the sandbox directory prefix.
	// Need to ensure it matches exactly or with a path separator.
	if cleanedPath != sandboxDir && !strings.HasPrefix(cleanedPath, sandboxDir+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: path '%s' (resolved to '%s') is outside sandbox '%s'", ErrPathViolation, inputPath, cleanedPath, sandboxDir)
	}

	// Optionally, check for symlinks leading outside the sandbox (more complex)
	// For now, rely on the prefix check after cleaning.

	return cleanedPath, nil
}

// ConvertToJSON attempts to convert various Go types to a JSON string representation.
func ConvertToJSON(data interface{}) (string, error) {
	if data == nil {
		return "null", nil
	}

	// Handle specific types that might not marshal well or need custom representation
	switch v := data.(type) {
	case string:
		// Marshal string correctly to include quotes and escape sequences
		bytes, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("failed to marshal string for JSON: %w", err)
		}
		return string(bytes), nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		// Use fmt for simple numerics (json.Marshal also works but adds overhead)
		return fmt.Sprintf("%v", v), nil
	case []interface{}, map[string]interface{}:
		// Standard complex types that json.Marshal handles well
		bytes, err := json.MarshalIndent(v, "", "  ") // Use indent for readability
		if err != nil {
			return "", fmt.Errorf("failed to marshal data to JSON: %w", err)
		}
		return string(bytes), nil
	// Add cases for other specific types if needed (e.g., custom structs)
	default:
		// For unknown types, attempt default marshaling
		// Check if it's a slice or map of potentially marshalable types
		rt := reflect.TypeOf(v)
		switch rt.Kind() {
		case reflect.Slice, reflect.Map:
			bytes, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				// Fallback to string representation if marshaling fails
				return fmt.Sprintf("%q", fmt.Sprintf("%v", v)), nil // Quoted string representation
			}
			return string(bytes), nil
		default:
			// Fallback for other types: return quoted string representation
			return fmt.Sprintf("%q", fmt.Sprintf("%v", v)), nil
		}
	}
}

// --- Glob Matching ---
// Borrowed from https://github.com/ryanuber/go-glob/blob/master/glob.go (MIT License)
// Original License:
// The MIT License (MIT)
// Copyright (c) 2015 Ryan Uber
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions: The above copyright
// notice and this permission notice shall be included in all copies or
// substantial portions of the Software.
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Glob converts a glob pattern to a regex pattern.
func Glob(glob string) string {
	var regex strings.Builder
	regex.WriteRune('^') // Anchor at the beginning

	inGroup := false
	inClass := false

	for i := 0; i < len(glob); i++ {
		r := glob[i]

		switch r {
		case '\\': // Escape next character
			i++
			if i < len(glob) {
				regex.WriteByte(glob[i])
			}
		case '*': // Match zero or more characters
			regex.WriteString(".*")
		case '?': // Match exactly one character
			regex.WriteRune('.')
		case '[': // Character class start
			inClass = true
			regex.WriteRune('[')
		case ']': // Character class end
			inClass = false
			regex.WriteRune(']')
		case '{': // Group start
			inGroup = true
			regex.WriteRune('(')
		case '}': // Group end
			inGroup = false
			regex.WriteRune(')')
		case ',': // Group separator
			if inGroup {
				regex.WriteRune('|')
			} else {
				regex.WriteRune(',')
			}
		case '!': // Negation within character class
			if inClass {
				regex.WriteRune('^')
			} else {
				regex.WriteRune('!')
			}
		default:
			// Escape regex metacharacters if not in a class
			if !inClass && (r == '$' || r == '^' || r == '.' || r == '(' || r == ')' || r == '|' || r == '+') {
				regex.WriteRune('\\')
			}
			regex.WriteByte(byte(r)) // Write the character itself
		}
	}

	regex.WriteRune('$') // Anchor at the end
	return regex.String()
}

// GlobMatch checks if a name matches a glob pattern.
func GlobMatch(pattern, name string) (bool, error) {
	regexPattern := Glob(pattern)
	return regexp.MatchString(regexPattern, name)
}
