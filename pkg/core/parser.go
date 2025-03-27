package core

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"
)

// --- AST Definitions ---
type Docstring struct {
	Purpose   string
	Inputs    map[string]string
	Output    string
	Algorithm string
	Caveats   string
	Examples  string
}
type Procedure struct {
	Name      string
	Params    []string
	Docstring Docstring
	Steps     []Step
}
type Step struct {
	Type   string
	Target string
	Value  interface{}
	Args   []string
	Cond   string
}

// --- Parser Implementation ---

// Final ParseFile Refactor
func ParseFile(r io.Reader) ([]Procedure, error) {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	procedures := []Procedure{}
	var currentProc *Procedure
	lineNum := 0
	var docstringLines []string // Buffer used only when collecting docstring
	collectingDocstring := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Are we collecting lines for a docstring?
		if collectingDocstring {
			trimmedLine := strings.TrimSpace(line)
			if trimmedLine == "END" { // End of docstring block
				if currentProc == nil {
					return nil, fmt.Errorf("L%d: internal error: END docstring but no current proc", lineNum)
				}
				doc, err := parseDocstringBlock(docstringLines)
				if err != nil {
					startLine := lineNum - len(docstringLines) - 1
					if startLine < 1 {
						startLine = 1
					}
					return nil, fmt.Errorf("L%d-%d: docstring block error: %w", startLine, lineNum, err)
				}
				currentProc.Docstring = doc
				collectingDocstring = false // Reset flag
				docstringLines = nil        // Clear buffer
			} else {
				docstringLines = append(docstringLines, line) // Collect line
			}
			continue // Always continue to next line when collecting/ending docstring
		}

		// --- Not collecting docstring ---
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "#") || strings.HasPrefix(trimmedLine, "--") {
			continue
		} // Skip blank/comment

		// Is it a new procedure definition?
		if strings.HasPrefix(trimmedLine, "DEFINE PROCEDURE") {
			if currentProc != nil { // Finalize previous if any
				if err := validateDocstring(currentProc.Docstring); err != nil {
					return nil, fmt.Errorf("~L%d(before %s): doc validation: %w", lineNum, trimmedLine, err)
				}
				procedures = append(procedures, *currentProc)
			}
			proc, err := parseProcedureHeader(trimmedLine)
			if err != nil {
				return nil, fmt.Errorf("L%d: header error: %w", lineNum, err)
			}
			currentProc = &proc // Start new proc
			continue
		}

		// If not starting a new procedure, we MUST be inside one already
		if currentProc == nil {
			if trimmedLine != "END" { // Ignore stray ENDs
				return nil, fmt.Errorf("L%d: unexpected statement outside procedure: %s", lineNum, trimmedLine)
			}
			continue // Ignore stray END
		}

		// Inside a procedure - check for COMMENT start
		if strings.HasPrefix(trimmedLine, "COMMENT:") && len(currentProc.Steps) == 0 {
			collectingDocstring = true  // Start collecting mode
			docstringLines = []string{} // Init buffer
			// Content on COMMENT: line itself is ignored by spec/parser. Next lines are collected.
			continue
		}

		// Check for Procedure END
		if trimmedLine == "END" {
			if err := validateDocstring(currentProc.Docstring); err != nil {
				return nil, fmt.Errorf("~L%d (proc %s) doc validation: %w", lineNum, currentProc.Name, err)
			}
			procedures = append(procedures, *currentProc)
			currentProc = nil // Reset state
			continue
		}

		// Check for ELSE (ignore)
		if strings.ToUpper(trimmedLine) == "ELSE" {
			continue
		}

		// Otherwise, parse as Step
		step, err := parseStep(trimmedLine)
		if err != nil {
			return nil, fmt.Errorf("L%d: step parse error ('%s'): %w", lineNum, trimmedLine, err)
		}
		if step.Type != "" {
			currentProc.Steps = append(currentProc.Steps, step)
		}

	} // End scanner loop

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}
	if collectingDocstring {
		return nil, fmt.Errorf("L%d: EOF missing 'END' for docstring block", lineNum)
	} // Check state at EOF
	if currentProc != nil { // Finalize last procedure at EOF
		if err := validateDocstring(currentProc.Docstring); err != nil {
			return nil, fmt.Errorf("EOF: proc %s docstring validation: %w", currentProc.Name, err)
		}
		procedures = append(procedures, *currentProc)
	}
	return procedures, nil
}

func parseProcedureHeader(line string) (Procedure, error) { /* ... unchanged ... */
	definePrefix := "DEFINE PROCEDURE"
	rest := strings.TrimSpace(line[len(definePrefix):])
	openParen := strings.Index(rest, "(")
	closeParen := strings.LastIndex(rest, ")")
	if openParen == -1 && closeParen == -1 {
		if len(strings.Fields(rest)) == 1 && rest != "" {
			return Procedure{}, fmt.Errorf("invalid procedure header: missing '()' for parameters (e.g., %s())", rest)
		}
		if rest == "" {
			return Procedure{}, fmt.Errorf("invalid procedure header: missing procedure name and '()'")
		}
	}
	if openParen == -1 {
		return Procedure{}, fmt.Errorf("invalid procedure header: missing '(' for parameter list")
	}
	if closeParen == -1 {
		return Procedure{}, fmt.Errorf("invalid procedure header: missing ')' after parameter list")
	}
	if closeParen < openParen {
		return Procedure{}, fmt.Errorf("invalid procedure header: ')' appears before '(' or parentheses mismatch")
	}
	name := strings.TrimSpace(rest[:openParen])
	if name == "" {
		return Procedure{}, fmt.Errorf("invalid procedure header: procedure name cannot be empty")
	}
	paramsPart := strings.TrimSpace(rest[openParen+1 : closeParen])
	var params []string
	if paramsPart != "" {
		params = splitParams(paramsPart)
	} else {
		params = []string{}
	}
	return Procedure{Name: name, Params: params}, nil
}
func splitParams(paramStr string) []string { /* ... unchanged ... */
	if strings.TrimSpace(paramStr) == "" {
		return []string{}
	}
	parts := strings.Split(paramStr, ",")
	params := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmedParam := strings.TrimSpace(p)
		if trimmedParam != "" {
			params = append(params, trimmedParam)
		}
	}
	return params
}

// Updated parseDocstringBlock - Logic moved from old parseDocstring
func parseDocstringBlock(lines []string) (Docstring, error) {
	doc := Docstring{Inputs: make(map[string]string)}
	currentSection := ""
	var sectionContent strings.Builder
	sectionHeaderRegex := regexp.MustCompile(`(?i)^\s*(PURPOSE|INPUTS|OUTPUT|ALGORITHM|CAVEATS|EXAMPLES)\s*:\s*(.*)$`)
	// baseIndent/getIndent removed - using simple TrimSpace in saveSectionContent

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		// Skip truly empty lines between sections? Or keep for formatting? Keep for now.
		// if trimmedLine == "" { continue }

		match := sectionHeaderRegex.FindStringSubmatch(line)
		isNewSection := len(match) == 3

		if isNewSection {
			newSectionKey := strings.ToUpper(strings.TrimSpace(match[1]))
			restOfLine := match[2]
			// Save previous section before switching
			if err := saveSectionContent(currentSection, &sectionContent, &doc); err != nil {
				return doc, fmt.Errorf("line %d: section '%s': %w", i+1, currentSection, err)
			}
			currentSection = newSectionKey
			sectionContent.Reset()
			sectionContent.WriteString(restOfLine) // Start raw
		} else { // Continuation line
			// Error if content appears before the *first* section header is found
			if currentSection == "" && trimmedLine != "" {
				return doc, fmt.Errorf("line %d: content before first section header: %q", i+1, line)
			}
			// Only append if we are *inside* a section
			if currentSection != "" {
				// Add newline only if buffer isn't empty (avoid leading newline)
				if sectionContent.Len() > 0 {
					sectionContent.WriteString("\n")
				}
				sectionContent.WriteString(line) // Add raw line content
			}
		}
	}
	// Save the very last section after the loop finishes
	if err := saveSectionContent(currentSection, &sectionContent, &doc); err != nil {
		return doc, fmt.Errorf("last section '%s': %w", currentSection, err)
	}

	return doc, nil
}

// Updated saveSectionContent - Simplest TrimSpace ONLY
func saveSectionContent(section string, content *strings.Builder, doc *Docstring, baseIndent ...int) error { // baseIndent ignored
	if section == "" {
		return nil
	} // No section active, nothing to save
	rawContent := content.String()
	finalContent := strings.TrimSpace(rawContent) // Simplest processing: trim all leading/trailing space

	switch section {
	case "PURPOSE":
		doc.Purpose = finalContent
	case "OUTPUT":
		doc.Output = finalContent
	case "ALGORITHM":
		doc.Algorithm = finalContent
	case "CAVEATS":
		doc.Caveats = finalContent
	case "EXAMPLES":
		doc.Examples = finalContent
	case "INPUTS":
		// Input parsing uses raw content for structure
		err := processInputBlock(rawContent, doc.Inputs)
		if err != nil {
			return fmt.Errorf("parsing INPUTS: %w", err)
		}
	default:
		return fmt.Errorf("internal: unknown section '%s'", section)
	}
	return nil
}

func processInputBlock(content string, inputsMap map[string]string) error { /* ... unchanged ... */
	trimmedContent := strings.TrimSpace(content)
	if strings.EqualFold(trimmedContent, "none") {
		return nil
	}
	lines := strings.Split(content, "\n")
	var currentInputName string
	var currentInputDesc strings.Builder
	inputLineRegex := regexp.MustCompile(`^\s*-\s*([^(: M]+?)\s*(?:\(([^)]+)\))?\s*:\s*(.*)$`)
	foundInputLine := false
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			continue
		}
		match := inputLineRegex.FindStringSubmatch(line)
		if len(match) == 4 {
			foundInputLine = true
			if currentInputName != "" {
				inputsMap[currentInputName] = strings.TrimSpace(currentInputDesc.String())
			}
			currentInputName = strings.TrimSpace(match[1])
			currentInputDesc.Reset()
			descPart := strings.TrimSpace(match[3])
			if descPart != "" {
				currentInputDesc.WriteString(descPart)
			}
		} else if currentInputName != "" && (strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")) {
			if currentInputDesc.Len() > 0 {
				currentInputDesc.WriteString("\n")
			}
			currentInputDesc.WriteString(trimmedLine)
		} else if currentInputName != "" {
			inputsMap[currentInputName] = strings.TrimSpace(currentInputDesc.String())
			currentInputName = ""
			currentInputDesc.Reset()
			if !strings.HasPrefix(trimmedLine, "-") {
				return fmt.Errorf("malformed INPUTS content (expected '- ...' or indent): %q", line)
			}
		} else if !strings.HasPrefix(trimmedLine, "-") && !strings.EqualFold(trimmedLine, "none") {
			return fmt.Errorf("malformed INPUTS content (expected '- ...'): %q", line)
		} else if strings.HasPrefix(trimmedLine, "-") {
			return fmt.Errorf("malformed INPUTS definition line (check format): %q", line)
		}
	}
	if currentInputName != "" {
		inputsMap[currentInputName] = strings.TrimSpace(currentInputDesc.String())
	}
	if !foundInputLine && !strings.EqualFold(trimmedContent, "none") && trimmedContent != "" {
		return fmt.Errorf("malformed INPUTS content: expected '- name: description' format")
	}
	return nil
}
func validateDocstring(doc Docstring) error { /* ... unchanged ... */
	missing := []string{}
	if doc.Purpose == "" {
		missing = append(missing, "PURPOSE")
	}
	if doc.Inputs == nil {
		missing = append(missing, "INPUTS (internal map error)")
	}
	if doc.Output == "" {
		missing = append(missing, "OUTPUT")
	}
	if doc.Algorithm == "" {
		missing = append(missing, "ALGORITHM")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required section(s): %s", strings.Join(missing, ", "))
	}
	return nil
}
func parseStep(line string) (Step, error) { /* ... unchanged ... */
	tokens := splitTokens(line)
	if len(tokens) == 0 {
		return Step{}, nil
	}
	keyword := strings.ToUpper(tokens[0])
	emptyStep := Step{}
	switch keyword {
	case "SET":
		return parseSetStep(tokens)
	case "CALL":
		return parseCallStep(tokens)
	case "RETURN":
		return parseReturnStep(tokens)
	case "IF":
		return parseIfStep(line)
	case "WHILE":
		return parseWhileStep(line)
	case "FOR":
		return parseForStep(line)
	case "DEFINE", "COMMENT:", "END", "ELSE":
		return emptyStep, nil
	default:
		return emptyStep, fmt.Errorf("unknown statement keyword: '%s'", tokens[0])
	}
}

// Updated parseIfStep - Error Logic Refined
func parseIfStep(line string) (Step, error) {
	ifRegex := regexp.MustCompile(`(?i)^\s*IF\s+(.+?)\s+THEN\s+(.*?)(?:\s+END\s*)?$`)
	match := ifRegex.FindStringSubmatch(line)
	if len(match) == 3 {
		cond := strings.TrimSpace(match[1])
		body := strings.TrimSpace(match[2])
		body = strings.TrimSuffix(strings.TrimSpace(body), "END")
		body = strings.TrimSpace(body)
		return Step{Type: "IF", Cond: cond, Value: body}, nil
	}
	// Regex failed diagnose:
	upperLine := strings.ToUpper(strings.TrimSpace(line))
	if !strings.HasPrefix(upperLine, "IF ") {
		return Step{}, fmt.Errorf("malformed IF (no IF)")
	}
	thenIndex := strings.Index(upperLine, " THEN ")
	// **Check 1: Missing THEN?**
	if thenIndex == -1 {
		return Step{}, fmt.Errorf("missing THEN keyword in IF statement")
	}
	// **Check 2: Missing Condition?** (Condition part is empty between IF and THEN)
	conditionPart := ""
	if len("IF ") <= thenIndex {
		conditionPart = strings.TrimSpace(upperLine[len("IF "):thenIndex])
	}
	if conditionPart == "" {
		return Step{}, fmt.Errorf("missing condition after IF keyword")
	}
	// **Check 3: Missing Body?** (Nothing meaningful after THEN)
	bodyPart := ""
	if len(upperLine) > thenIndex+len(" THEN ") {
		bodyPart = strings.TrimSpace(upperLine[thenIndex+len(" THEN "):])
		bodyPart = strings.TrimSuffix(strings.TrimSpace(bodyPart), " END")
		bodyPart = strings.TrimSpace(bodyPart)
	}
	// ** This is the crucial check **
	if bodyPart == "" {
		return Step{}, fmt.Errorf("missing statement/body after THEN in IF statement")
	}
	// Fallback if none of the above matched the error condition
	return Step{}, fmt.Errorf("malformed IF statement structure")
}

// Updated parseWhileStep - Error Logic Refined
func parseWhileStep(line string) (Step, error) {
	whileRegex := regexp.MustCompile(`(?i)^\s*WHILE\s+(.+?)\s+DO\s+(.*?)(?:\s+END\s*)?$`)
	match := whileRegex.FindStringSubmatch(line)
	if len(match) == 3 {
		cond := strings.TrimSpace(match[1])
		body := strings.TrimSpace(match[2])
		body = strings.TrimSuffix(strings.TrimSpace(body), "END")
		body = strings.TrimSpace(body)
		return Step{Type: "WHILE", Cond: cond, Value: body}, nil
	}
	// Regex failed diagnose:
	upperLine := strings.ToUpper(strings.TrimSpace(line))
	if !strings.HasPrefix(upperLine, "WHILE ") {
		return Step{}, fmt.Errorf("malformed WHILE (no WHILE)")
	}
	// **Check 1: Missing DO?**
	doIndex := strings.Index(upperLine, " DO ")
	if doIndex == -1 {
		return Step{}, fmt.Errorf("missing DO keyword in WHILE statement")
	}
	// **Check 2: Missing Condition?**
	conditionPart := ""
	if len("WHILE ") <= doIndex {
		conditionPart = strings.TrimSpace(upperLine[len("WHILE "):doIndex])
	}
	if conditionPart == "" {
		return Step{}, fmt.Errorf("missing condition after WHILE keyword")
	}
	// **Check 3: Missing Body?**
	bodyPart := ""
	if len(upperLine) > doIndex+len(" DO ") {
		bodyPart = strings.TrimSpace(upperLine[doIndex+len(" DO "):])
		bodyPart = strings.TrimSuffix(strings.TrimSpace(bodyPart), " END")
		bodyPart = strings.TrimSpace(bodyPart)
	}
	if bodyPart == "" {
		return Step{}, fmt.Errorf("missing statement/body after DO in WHILE statement")
	}
	return Step{}, fmt.Errorf("malformed WHILE statement structure") // Fallback
}

func parseForStep(line string) (Step, error) { /* ... unchanged ... */
	forRegex := regexp.MustCompile(`(?i)^\s*FOR\s+EACH\s+([a-zA-Z_][a-zA-Z0-9_]*)\s+IN\s+(.+?)\s*(?:DO)?\s*$`)
	match := forRegex.FindStringSubmatch(line)
	if len(match) != 3 {
		upperLine := strings.ToUpper(strings.TrimSpace(line))
		if !strings.HasPrefix(upperLine, "FOR EACH ") {
			return Step{}, fmt.Errorf("malformed FOR EACH statement, must start with 'FOR EACH'")
		}
		if !strings.Contains(upperLine, " IN ") {
			return Step{}, fmt.Errorf("invalid FOR EACH syntax, missing 'IN' keyword")
		}
		return Step{}, fmt.Errorf("invalid FOR EACH syntax, expected 'FOR EACH variable IN collection [DO]'")
	}
	loopVar := strings.TrimSpace(match[1])
	collectionExpr := strings.TrimSpace(match[2])
	if loopVar == "" {
		return Step{}, fmt.Errorf("missing loop variable name")
	}
	if collectionExpr == "" {
		return Step{}, fmt.Errorf("missing collection expression")
	}
	return Step{Type: "FOR", Target: loopVar, Value: collectionExpr}, nil
}

// Updated parseSetStep - Final check on error logic
func parseSetStep(tokens []string) (Step, error) {
	if len(tokens) < 2 {
		return Step{}, fmt.Errorf("invalid SET syntax, expected 'SET variable = value'")
	}
	// ** Check for SET = value first **
	if tokens[1] == "=" {
		return Step{}, fmt.Errorf("invalid SET syntax: invalid variable name '='")
	}

	variableName := tokens[1]
	value := ""
	assignmentFound := false
	if len(tokens) >= 3 && tokens[2] == "=" { // Case: SET var = value...
		assignmentFound = true
		if len(tokens) > 3 {
			value = strings.Join(tokens[3:], " ")
		} else {
			value = ""
		}
	} else if len(tokens) >= 2 { // Case: SET var=value... OR SET var (no equals)
		// Try splitting token 1 on equals
		parts := strings.SplitN(tokens[1], "=", 2)
		if len(parts) == 2 && parts[0] != "" { // Found var=value in first token
			variableName = parts[0]
			assignmentFound = true
			value = parts[1]
			if len(tokens) > 2 {
				value = value + " " + strings.Join(tokens[2:], " ")
			}
		} else if len(parts) == 1 && len(tokens) >= 3 { // Looks like SET var value... (missing equals)
			return Step{}, fmt.Errorf("missing '=' after variable '%s'", tokens[1])
		} else if len(parts) == 1 && len(tokens) == 2 { // Just SET var
			return Step{}, fmt.Errorf("missing '=' and value after variable '%s'", tokens[1])
		}
	}

	if !assignmentFound {
		return Step{}, fmt.Errorf("invalid SET syntax, expected 'SET variable = value'")
	} // Fallback

	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
			value = value[1 : len(value)-1]
		}
	}
	return Step{Type: "SET", Target: variableName, Value: value}, nil
}

func parseCallStep(tokens []string) (Step, error) { /* ... unchanged ... */
	if len(tokens) < 2 {
		return Step{}, fmt.Errorf("invalid CALL statement")
	}
	callRest := strings.Join(tokens[1:], " ")
	openParen := strings.Index(callRest, "(")
	closeParen := strings.LastIndex(callRest, ")")
	var target string
	var args []string
	if openParen != -1 && closeParen > openParen && strings.HasSuffix(strings.TrimSpace(callRest), ")") {
		target = strings.TrimSpace(callRest[:openParen])
		argsStr := callRest[openParen+1 : closeParen]
		args = parseCallArgs(argsStr)
	} else {
		if openParen != -1 && (closeParen <= openParen || !strings.HasSuffix(strings.TrimSpace(callRest), ")")) {
			return Step{}, fmt.Errorf("mismatched parentheses: '%s'", callRest)
		}
		if openParen == -1 && strings.Contains(callRest, ")") {
			return Step{}, fmt.Errorf("found ')' without '(': '%s'", callRest)
		}
		if openParen == -1 {
			return Step{}, fmt.Errorf("missing '()': '%s()'", callRest)
		}
		return Step{}, fmt.Errorf("invalid CALL structure: '%s'", callRest)
	}
	if target == "" {
		return Step{}, fmt.Errorf("missing CALL target")
	}
	return Step{Type: "CALL", Target: target, Args: args}, nil
}
func parseCallArgs(argsStr string) []string { /* ... unchanged ... */
	trimmedArgsStr := strings.TrimSpace(argsStr)
	if trimmedArgsStr == "" {
		return []string{}
	}
	var args []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)
	escapeNext := false
	for _, ch := range trimmedArgsStr {
		if escapeNext {
			current.WriteRune(ch)
			escapeNext = false
			continue
		}
		if ch == '\\' {
			escapeNext = true
			continue
		}
		switch {
		case (ch == '"' || ch == '\'') && !inQuotes:
			inQuotes = true
			quoteChar = ch
		case ch == quoteChar && inQuotes:
			inQuotes = false
			quoteChar = rune(0)
		case ch == ',' && !inQuotes:
			args = append(args, strings.TrimSpace(current.String()))
			current.Reset()
		case !inQuotes && (ch == ' ' || ch == '\t'):
			if current.Len() > 0 {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}
	lastArg := strings.TrimSpace(current.String())
	if lastArg != "" || strings.HasSuffix(argsStr, ",") || (len(args) == 0 && argsStr != "") {
		args = append(args, lastArg)
	} else if argsStr != "" && len(args) == 0 && current.Len() == 0 {
		args = append(args, "")
	} else if argsStr != "" && len(args) == 0 {
		args = append(args, strings.TrimSpace(current.String()))
	}
	for i := range args {
		args[i] = strings.TrimSpace(args[i])
	}
	return args
}
func parseReturnStep(tokens []string) (Step, error) { /* ... unchanged ... */
	if len(tokens) < 2 {
		return Step{}, fmt.Errorf("invalid RETURN format")
	}
	value := strings.Join(tokens[1:], " ")
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
			value = value[1 : len(value)-1]
		}
	}
	return Step{Type: "RETURN", Value: value}, nil
}
func splitTokens(line string) []string { /* ... unchanged (simple space split) ... */
	commentIdx := -1
	inQuotes := false
	quoteChar := rune(0)
	escapeNext := false
	for i, ch := range line {
		if escapeNext {
			escapeNext = false
			continue
		}
		if ch == '\\' {
			escapeNext = true
			continue
		}
		if (ch == '"' || ch == '\'') && !inQuotes {
			inQuotes = true
			quoteChar = ch
		} else if ch == quoteChar && inQuotes {
			inQuotes = false
			quoteChar = rune(0)
		} else if !inQuotes && (ch == '#' || (ch == '-' && i+1 < len(line) && line[i+1] == '-')) {
			commentIdx = i
			break
		}
	}
	if commentIdx != -1 {
		line = line[:commentIdx]
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return []string{}
	}
	var tokens []string
	var current strings.Builder
	inQuotes = false
	quoteChar = rune(0)
	escapeNext = false
	for _, ch := range line {
		if escapeNext {
			current.WriteRune(ch)
			escapeNext = false
			continue
		}
		if ch == '\\' {
			escapeNext = true
			current.WriteRune(ch)
			continue
		}
		if (ch == '"' || ch == '\'') && !inQuotes {
			inQuotes = true
			quoteChar = ch
			current.WriteRune(ch)
		} else if ch == quoteChar && inQuotes {
			inQuotes = false
			quoteChar = rune(0)
			current.WriteRune(ch)
		} else if !inQuotes && unicode.IsSpace(ch) {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(ch)
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}
