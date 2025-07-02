// NeuroScript Version: 0.3.1
// File version: 0.0.1 // Removed init() function; registration handled by tooldefs_io/zz_registrar.
// nlines: 44 // Approximate
// risk_rating: LOW
// filename: pkg/tool/io/tools_io.go

package io

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// toolInput implements the Input tool.
func toolInput(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	prompt := ""
	if len(args) > 0 {
		if p, ok := args[0].(string); ok {
			prompt = p
		} else if args[0] != nil {
			// Handle non-string, non-nil prompt argument if necessary, or error out
			// For now, we only accept string prompts or nil.
			return "", lang.NewRuntimeError(ErrorCodeType, fmt.Sprintf("Input: prompt argument must be a string or null, got %T", args[0]), ErrInvalidArgument)
		}
	}

	// Print prompt directly to stdout if provided
	if prompt != "" {
		fmt.Print(prompt)	// Use Print, not Println, so input is on the same line
	}

	// Read input from standard input
	reader := bufio.NewReader(os.Stdin)
	// TODO: Consider context cancellation or timeouts if input needs to be interruptible.
	line, err := reader.ReadString('\n')
	if err != nil {
		// Log the error, but often for interactive input, returning empty string might be acceptable.
		// Or return a specific error if EOF or other issues are critical.
		interpreter.Logger().Warn("Tool: Input read error", "error", err)
		// Use ErrorCodeIOFailed for read errors
		return "", lang.NewRuntimeError(ErrorCodeIOFailed, "failed to read input", errors.Join(ErrIOFailed, err))
	}

	// Trim trailing newline characters (\n or \r\n)
	line = strings.TrimRight(line, "\r\n")

	interpreter.Logger().Debug("Tool: Input read line successfully")	// Avoid logging the actual input content
	return line, nil
}

// toolPrint implements the Print tool.
func toolPrint(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// The spec defines one arg "values" of type Any.
	// This implementation will handle if that arg is a single value or a slice.
	if len(args) != 1 {
		// This shouldn't happen if argument validation works correctly based on the spec.
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "Print: tool implementation expected exactly 1 argument ('values')", ErrArgumentMismatch)
	}

	valuesToPrint := []interface{}{}
	valuesArg := args[0]

	// Check if the single argument is itself a slice
	if slice, ok := valuesArg.([]interface{}); ok {
		valuesToPrint = slice
	} else {
		// Treat the single argument as the only value to print
		valuesToPrint = append(valuesToPrint, valuesArg)
	}

	// Convert slice of interface{} to slice of string for logging (optional)
	// stringValues := make([]string, len(valuesToPrint))
	// for i, v := range valuesToPrint {
	// 	stringValues[i] = fmt.Sprint(v) // Simple conversion
	// }
	// interpreter.Logger().Debug("Tool: Print executing", "values", strings.Join(stringValues, " "))

	// Use fmt.Println which handles different types and adds spaces + newline
	fmt.Println(valuesToPrint...)

	// Print tool has no meaningful return value
	return nil, nil
}