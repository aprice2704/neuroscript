package core

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ParseFile reads a .ns file and returns parsed Procedures
func ParseFile(filename string) ([]Procedure, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	procedures := []Procedure{}
	var currentProc *Procedure

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines/comments
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}

		// Start of a new procedure
		if strings.HasPrefix(line, "DEFINE PROCEDURE") {
			proc, err := parseProcedureHeader(line)
			if err != nil {
				return nil, err
			}
			currentProc = &proc
			continue
		}

		// Docstring block
		if strings.HasPrefix(line, "COMMENT:") {
			doc, err := parseDocstring(scanner)
			if err != nil {
				return nil, err
			}
			currentProc.Docstring = doc
			continue
		}

		// Parse steps (SET, CALL, etc.)
		if currentProc != nil {
			step, err := parseStep(line)
			if err != nil {
				return nil, err
			}
			currentProc.Steps = append(currentProc.Steps, step)
		}
	}

	// Add the last procedure if it exists
	if currentProc != nil {
		procedures = append(procedures, *currentProc)
	}

	return procedures, nil
}

// Helper: Parse "DEFINE PROCEDURE Multiply(a, b)"
func parseProcedureHeader(line string) (Procedure, error) {
	parts := strings.Split(line, " ")
	if len(parts) < 3 {
		return Procedure{}, fmt.Errorf("invalid procedure header: %s", line)
	}

	name := strings.Split(parts[2], "(")[0]
	params := strings.TrimSuffix(strings.Split(parts[2], "(")[1], ")")
	paramList := strings.Split(params, ",")
	for i := range paramList {
		paramList[i] = strings.TrimSpace(paramList[i])
	}

	return Procedure{
		Name:   name,
		Params: paramList,
	}, nil
}

// parseDocstring extracts and validates docstring sections
func parseDocstring(scanner *bufio.Scanner) (Docstring, error) {
	doc := Docstring{
		Inputs: make(map[string]string),
	}
	currentSection := ""
	var sectionContent strings.Builder
	requiredSections := []string{"PURPOSE", "INPUTS", "OUTPUT", "ALGORITHM"}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "END" {
			break
		}

		// Check for section headers (e.g., "PURPOSE:")
		if strings.HasSuffix(line, ":") {
			// Save previous section content
			if currentSection != "" {
				content := strings.TrimSpace(sectionContent.String())
				switch currentSection {
				case "PURPOSE":
					doc.Purpose = content
				case "INPUTS":
					// Already processed line-by-line
				case "OUTPUT":
					doc.Output = content
				case "ALGORITHM":
					doc.Algorithm = content
				case "CAVEATS":
					doc.Caveats = content
				case "EXAMPLES":
					doc.Examples = content
				}
				sectionContent.Reset()
			}
			currentSection = strings.TrimSuffix(line, ":")
			continue
		}

		// Process section content
		switch currentSection {
		case "INPUTS":
			// Format: "- argName (type): description"
			if strings.HasPrefix(line, "- ") {
				parts := strings.SplitN(line[len("- "):], ":", 2)
				if len(parts) == 2 {
					argParts := strings.SplitN(parts[0], "(", 2)
					argName := strings.TrimSpace(argParts[0])
					doc.Inputs[argName] = strings.TrimSpace(parts[1])
				}
			}
		default:
			sectionContent.WriteString(line + "\n")
		}
	}

	// Save the last section
	if currentSection != "" {
		content := strings.TrimSpace(sectionContent.String())
		switch currentSection {
		case "PURPOSE":
			doc.Purpose = content
		case "OUTPUT":
			doc.Output = content
		case "ALGORITHM":
			doc.Algorithm = content
		case "CAVEATS":
			doc.Caveats = content
		case "EXAMPLES":
			doc.Examples = content
		}
	}

	// Validate required sections
	for _, section := range requiredSections {
		switch section {
		case "PURPOSE":
			if doc.Purpose == "" {
				return doc, fmt.Errorf("missing required PURPOSE section in docstring")
			}
		case "INPUTS":
			if len(doc.Inputs) == 0 {
				return doc, fmt.Errorf("missing required INPUTS section in docstring")
			}
		case "OUTPUT":
			if doc.Output == "" {
				return doc, fmt.Errorf("missing required OUTPUT section in docstring")
			}
		case "ALGORITHM":
			if doc.Algorithm == "" {
				return doc, fmt.Errorf("missing required ALGORITHM section in docstring")
			}
		}
	}

	return doc, nil
}
func parseStep(line string) (Step, error) {
	tokens := splitTokens(line)
	if len(tokens) == 0 {
		return Step{}, nil // Skip empty lines
	}

	switch tokens[0] {
	case "SET":
		return parseSetStep(tokens)
	case "CALL":
		return parseCallStep(tokens)
	case "RETURN":
		return parseReturnStep(tokens)
	case "IF":
		return parseIfStep(tokens)
	case "FOR":
		return parseForStep(tokens)
	case "WHILE":
		return parseWhileStep(tokens)
	case "COMMENT":
		return Step{Type: "COMMENT"}, nil // Inline comments
	default:
		return Step{}, fmt.Errorf("unknown statement: %s", tokens[0])
	}
}

// parseWhileStep handles WHILE condition DO ... END statements
func parseWhileStep(tokens []string) (Step, error) {
	// Minimum valid: "WHILE x DO END"
	if len(tokens) < 4 || tokens[len(tokens)-1] != "END" {
		return Step{}, fmt.Errorf("invalid WHILE syntax, expected 'WHILE condition DO ... END'")
	}

	// Find the "DO" separator
	doIndex := -1
	for i, token := range tokens {
		if token == "DO" {
			doIndex = i
			break
		}
	}
	if doIndex == -1 {
		return Step{}, fmt.Errorf("missing DO in WHILE statement")
	}

	// Extract condition (everything between WHILE and DO)
	condition := strings.Join(tokens[1:doIndex], " ")

	// Extract body (everything between DO and END)
	body := strings.Join(tokens[doIndex+1:len(tokens)-1], " ")

	return Step{
		Type:  "WHILE",
		Cond:  condition,
		Value: body, // The body to execute while condition is true
	}, nil
}

// New: Parse FOR EACH item IN collection
func parseForStep(tokens []string) (Step, error) {
	if len(tokens) < 6 || tokens[1] != "EACH" || tokens[3] != "IN" {
		return Step{}, fmt.Errorf("invalid FOR EACH syntax")
	}
	return Step{
		Type:   "FOR",
		Target: tokens[2], // Loop variable (e.g., "item")
		Value:  tokens[4], // Collection (e.g., "ListOfNumbers")
	}, nil
}

// Helper: Parse SET var = value
func parseSetStep(tokens []string) (Step, error) {
	if len(tokens) < 4 || tokens[2] != "=" {
		return Step{}, fmt.Errorf("invalid SET format, expected 'SET var = value'")
	}

	value := strings.Join(tokens[3:], " ")
	// Remove surrounding quotes if present
	if len(value) >= 2 && (value[0] == '"' && value[len(value)-1] == '"') {
		value = value[1 : len(value)-1]
	}

	return Step{
		Type:   "SET",
		Target: tokens[1],
		Value:  value,
	}, nil
}

// Helper: Parse CALL Proc(args) or CALL LLM("prompt")
func parseCallStep(tokens []string) (Step, error) {
	if len(tokens) < 2 {
		return Step{}, fmt.Errorf("invalid CALL, expected 'CALL target(args)'")
	}

	// Extract target and args
	targetAndArgs := strings.Join(tokens[1:], " ")
	openParen := strings.Index(targetAndArgs, "(")
	if openParen == -1 {
		return Step{}, fmt.Errorf("missing '(' in CALL")
	}
	if !strings.HasSuffix(targetAndArgs, ")") {
		return Step{}, fmt.Errorf("missing ')' in CALL")
	}

	target := strings.TrimSpace(targetAndArgs[:openParen])
	argsStr := targetAndArgs[openParen+1 : len(targetAndArgs)-1]
	args := parseCallArgs(argsStr)

	return Step{
		Type:   "CALL",
		Target: target,
		Args:   args,
	}, nil
}

// Helper: Parse comma-separated call arguments (handles quoted strings)
func parseCallArgs(argsStr string) []string {
	var args []string
	var current strings.Builder
	inQuotes := false

	for _, ch := range argsStr {
		switch {
		case ch == '"':
			inQuotes = !inQuotes
			current.WriteRune(ch)
		case ch == ',' && !inQuotes:
			args = append(args, strings.TrimSpace(current.String()))
			current.Reset()
		default:
			current.WriteRune(ch)
		}
	}

	// Add the last argument
	if current.Len() > 0 {
		args = append(args, strings.TrimSpace(current.String()))
	}

	// Remove surrounding quotes from each arg
	for i, arg := range args {
		if len(arg) >= 2 && arg[0] == '"' && arg[len(arg)-1] == '"' {
			args[i] = arg[1 : len(arg)-1]
		}
	}

	return args
}

// Helper: Parse RETURN value
func parseReturnStep(tokens []string) (Step, error) {
	value := strings.Join(tokens[1:], " ")
	if len(value) >= 2 && (value[0] == '"' && value[len(value)-1] == '"') {
		value = value[1 : len(value)-1]
	}

	return Step{
		Type:  "RETURN",
		Value: value,
	}, nil
}

// Helper: Parse IF condition THEN ...
func parseIfStep(tokens []string) (Step, error) {
	thenIdx := -1
	for i, token := range tokens {
		if token == "THEN" {
			thenIdx = i
			break
		}
	}
	if thenIdx == -1 {
		return Step{}, fmt.Errorf("invalid IF, missing THEN clause")
	}

	cond := strings.Join(tokens[1:thenIdx], " ")
	action := strings.Join(tokens[thenIdx+1:], " ")

	return Step{
		Type:  "IF",
		Cond:  cond,
		Value: action, // The action to take if true
	}, nil
}

// Helper: Split line into tokens while preserving quoted strings
func splitTokens(line string) []string {
	var tokens []string
	var current strings.Builder
	inQuotes := false

	for _, ch := range line {
		switch {
		case ch == '"':
			inQuotes = !inQuotes
			current.WriteRune(ch)
		case !inQuotes && (ch == ' ' || ch == '\t'):
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}
