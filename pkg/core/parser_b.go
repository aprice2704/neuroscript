// pkg/core/parser_b.go
package core

import (
	"fmt"
	"strings"
	"unicode"
)

// ParseStep parses a single line into a Step struct.
// It's made public ('P'arseStep) to be callable by the interpreter and the main parser loop.
// For block starters (IF, WHILE, FOR), it parses the header and expects ParseFile to handle the body.
func ParseStep(line string) (Step, error) {
	trimmedLine := trimComments(line) // Use helper to remove comments first
	trimmedLine = strings.TrimSpace(trimmedLine)
	if trimmedLine == "" {
		return Step{}, nil // Empty or comment-only line results in an empty step
	}

	// Get first word as keyword
	keyword := ""
	firstSpace := strings.IndexFunc(trimmedLine, unicode.IsSpace)
	if firstSpace == -1 {
		keyword = strings.ToUpper(trimmedLine) // Line is a single word
	} else {
		keyword = strings.ToUpper(trimmedLine[:firstSpace])
	}

	emptyStep := Step{}
	switch keyword {
	case "SET":
		return parseSetStep(line) // Pass original line
	case "CALL":
		return parseCallStep(line) // Pass original line
	case "RETURN":
		return parseReturnStep(line) // Pass original line
	case "IF":
		// Parses only "IF cond THEN", leaves Value=nil to signal block start
		return parseIfHeader(line)
	case "WHILE":
		// Parses only "WHILE cond DO", leaves Value=nil to signal block start
		return parseWhileHeader(line)
	case "FOR":
		// Check for "FOR EACH" specifically
		if strings.HasPrefix(strings.ToUpper(trimmedLine), "FOR EACH ") {
			// Parses only "FOR EACH var IN collection DO", leaves Value=nil to signal block start
			return parseForHeader(line)
		}
		return emptyStep, fmt.Errorf("unknown statement keyword: '%s' (expected 'FOR EACH')", keyword)

	case "ELSE":
		// This needs careful handling in the main ParseFile state machine,
		// as it signals the start of an ELSE block following an IF.
		// For now, let ParseStep just identify it. The main loop needs context.
		// It implicitly starts a block.
		if trimmedLine == "ELSE" {
			return Step{Type: "ELSE", Value: nil}, nil // Signal ELSE block start
		}
		// Allow "ELSE IF..."? Not currently, treat as error or separate IF.
		return emptyStep, fmt.Errorf("malformed ELSE statement (expected 'ELSE' on its own line): %q", line)

	// Keywords handled by the main ParseFile state machine or ignored within blocks
	case "DEFINE", "PROCEDURE", "COMMENT:":
		// These shouldn't be parsed by ParseStep when *inside* a procedure body normally.
		// If encountered, likely a syntax error unless ParseFile handles them specifically.
		return emptyStep, fmt.Errorf("unexpected keyword '%s' within procedure body", keyword)
	case "END":
		// A standalone "END" might terminate a block or the procedure.
		// ParseFile needs to handle this based on context (state).
		// ParseStep can return a specific type or nil? Let's return nil and let ParseFile check.
		if trimmedLine == "END" {
			return Step{Type: "END_BLOCK"}, nil // Special type for ParseFile state machine
		}
		return emptyStep, fmt.Errorf("unexpected content after END keyword: %q", line)

	default:
		// Check if it looks like a docstring keyword misplaced outside COMMENT block
		docstringKeywords := []string{"PURPOSE:", "INPUTS:", "OUTPUT:", "ALGORITHM:", "CAVEATS:", "EXAMPLES:"}
		upperTrimmed := strings.ToUpper(trimmedLine)
		for _, dk := range docstringKeywords {
			if strings.HasPrefix(upperTrimmed, dk) {
				return emptyStep, fmt.Errorf("unexpected docstring keyword outside COMMENT block: %s", trimmedLine)
			}
		}
		// If not a known keyword or docstring keyword, it's an unknown statement
		return emptyStep, fmt.Errorf("unknown statement keyword: '%s' in line: %q", keyword, line)
	}
}

// --- Step Parsing Helpers ---

// parseSetStep parses "SET var = value" (Unchanged)
func parseSetStep(originalLine string) (Step, error) {
	line := strings.TrimSpace(trimComments(originalLine))
	if !strings.HasPrefix(strings.ToUpper(line), "SET ") {
		return Step{}, fmt.Errorf("invalid SET syntax, must start with 'SET ': %q", originalLine)
	}

	eqIndex := findCharOutsideQuotes(line, '=')

	if eqIndex == -1 {
		parts := strings.Fields(line)
		if len(parts) == 2 {
			varName := parts[1]
			if isValidIdentifier(varName) {
				fmt.Printf("[Warn] SET statement '%s' missing '=', assuming assignment to empty string.\n", originalLine)
				return Step{Type: "SET", Target: varName, Value: ""}, nil
			}
		}
		return Step{}, fmt.Errorf("invalid SET syntax, missing '=' assignment operator: %q", originalLine)
	}

	prefixLen := len("SET ")
	if eqIndex <= prefixLen {
		return Step{}, fmt.Errorf("invalid SET syntax, missing variable name before '=': %q", originalLine)
	}
	variableName := strings.TrimSpace(line[prefixLen:eqIndex])
	if !isValidIdentifier(variableName) {
		return Step{}, fmt.Errorf("invalid SET syntax: invalid variable name '%s' in '%s'", variableName, originalLine)
	}

	value := ""
	if eqIndex+1 < len(line) {
		value = line[eqIndex+1:]
		value = strings.TrimSpace(value)
	} else {
		value = ""
	}

	return Step{Type: "SET", Target: variableName, Value: value}, nil
}

// parseCallStep parses "CALL target(arg1, arg2,...)" (Unchanged)
func parseCallStep(originalLine string) (Step, error) {
	line := strings.TrimSpace(trimComments(originalLine))
	if !strings.HasPrefix(strings.ToUpper(line), "CALL ") {
		return Step{}, fmt.Errorf("invalid CALL statement: must start with 'CALL ': %q", originalLine)
	}

	openParenIndex := findCharOutsideQuotes(line, '(')

	if openParenIndex == -1 {
		parts := strings.Fields(line)
		if len(parts) == 2 {
			targetName := parts[1]
			if isValidCallTarget(targetName) {
				fmt.Printf("[Warn] CALL statement '%s' missing '()', assuming zero arguments.\n", originalLine)
				return Step{Type: "CALL", Target: targetName, Args: []string{}}, nil
			}
		}
		return Step{}, fmt.Errorf("invalid CALL syntax: missing '(' for arguments: %q", originalLine)
	}

	prefixLen := len("CALL ")
	if openParenIndex <= prefixLen {
		return Step{}, fmt.Errorf("missing CALL target before '(': %q", originalLine)
	}
	target := strings.TrimSpace(line[prefixLen:openParenIndex])
	if target == "" {
		return Step{}, fmt.Errorf("missing CALL target before '(': %q", originalLine)
	}
	if !isValidCallTarget(target) {
		return Step{}, fmt.Errorf("invalid CALL target name: '%s' in '%s'", target, originalLine)
	}

	closeParenIndex := findMatchingParen(line, openParenIndex)

	if closeParenIndex == -1 {
		return Step{}, fmt.Errorf("invalid CALL syntax: missing or mismatched closing ')' for arguments: %q", originalLine)
	}

	if strings.TrimSpace(line[closeParenIndex+1:]) != "" {
		return Step{}, fmt.Errorf("unexpected content after closing ')' in CALL statement: %q", originalLine)
	}

	argsStr := line[openParenIndex+1 : closeParenIndex]
	args, err := parseCallArgs(argsStr)
	if err != nil {
		return Step{}, fmt.Errorf("error parsing CALL arguments for '%s' in '%s': %w", target, originalLine, err)
	}

	return Step{Type: "CALL", Target: target, Args: args}, nil
}

// parseReturnStep parses "RETURN [value]" (Unchanged)
func parseReturnStep(originalLine string) (Step, error) {
	line := strings.TrimSpace(trimComments(originalLine))
	upperLine := strings.ToUpper(line)

	if upperLine == "RETURN" {
		fmt.Printf("[Warn] RETURN statement '%s' missing value, assuming RETURN \"\".\n", originalLine)
		return Step{Type: "RETURN", Value: ""}, nil
	}

	if !strings.HasPrefix(upperLine, "RETURN ") {
		return Step{}, fmt.Errorf("invalid RETURN format: must be 'RETURN' or start with 'RETURN ': %q", originalLine)
	}

	value := line[len("RETURN "):]
	value = strings.TrimSpace(value)

	return Step{Type: "RETURN", Value: value}, nil
}

// parseIfHeader parses only "IF cond THEN", returns Step{Value: nil}
func parseIfHeader(originalLine string) (Step, error) {
	line := strings.TrimSpace(trimComments(originalLine))
	upperLine := strings.ToUpper(line)

	if !strings.HasPrefix(upperLine, "IF ") {
		return Step{}, fmt.Errorf("malformed IF statement (must start with 'IF '): %q", originalLine)
	}

	// Find ' THEN' (case-insensitive) respecting quotes - must be end of line now
	thenIndex := findKeywordIndex(line, "THEN", len("IF "))

	if thenIndex == -1 {
		// Allow THEN on next line? No, spec implies IF cond THEN starts the block.
		return Step{}, fmt.Errorf("missing ' THEN' keyword at the end of IF line: %q", originalLine)
	}

	// Check if anything follows THEN on the same line
	if strings.TrimSpace(line[thenIndex+len(" THEN"):]) != "" {
		return Step{}, fmt.Errorf("unexpected content after ' THEN' on IF line (body must be on subsequent lines): %q", originalLine)
	}

	condition := strings.TrimSpace(line[len("IF "):thenIndex])
	if condition == "" {
		return Step{}, fmt.Errorf("missing condition after IF keyword: %q", originalLine)
	}

	// Return Step with Type "IF", the Condition, and Value=nil to signal block start
	return Step{Type: "IF", Cond: condition, Value: nil}, nil
}

// parseWhileHeader parses only "WHILE cond DO", returns Step{Value: nil}
func parseWhileHeader(originalLine string) (Step, error) {
	line := strings.TrimSpace(trimComments(originalLine))
	upperLine := strings.ToUpper(line)

	if !strings.HasPrefix(upperLine, "WHILE ") {
		return Step{}, fmt.Errorf("malformed WHILE statement (must start with 'WHILE '): %q", originalLine)
	}

	// Find ' DO' respecting quotes - must be end of line now
	doIndex := findKeywordIndex(line, "DO", len("WHILE "))

	if doIndex == -1 {
		return Step{}, fmt.Errorf("missing ' DO' keyword at the end of WHILE line: %q", originalLine)
	}

	// Check if anything follows DO on the same line
	if strings.TrimSpace(line[doIndex+len(" DO"):]) != "" {
		return Step{}, fmt.Errorf("unexpected content after ' DO' on WHILE line (body must be on subsequent lines): %q", originalLine)
	}

	condition := strings.TrimSpace(line[len("WHILE "):doIndex])
	if condition == "" {
		return Step{}, fmt.Errorf("missing condition after WHILE keyword: %q", originalLine)
	}

	// Return Step with Type "WHILE", the Condition, and Value=nil to signal block start
	return Step{Type: "WHILE", Cond: condition, Value: nil}, nil
}
