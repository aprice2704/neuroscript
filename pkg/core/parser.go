// pkg/core/parser.go
package core

import (
	"bufio"
	"fmt"
	"io"
	"regexp" // Keep regexp import
	"strings"
)

// --- Parser Implementation ---

type parserState int

const (
	stateTopLevel      parserState = iota // Expecting DEFINE, comment, blank
	stateInProcedure                      // Expecting COMMENT, step, END procedure
	stateInDocstring                      // Expecting docstring line or END docstring
	stateProcedureDone                    // Just finished a procedure, expect DEFINE or EOF
	stateInBlock                          // Inside an IF/WHILE/FOR block, expecting steps or END block
	stateInElseBlock                      // Inside an ELSE block, expecting steps or END block
)

// BlockContext holds information about nested blocks
type BlockContext struct {
	BlockStep *Step       // The IF/WHILE/FOR step that started the block
	Steps     []Step      // Steps collected within this block
	StateType parserState // The state this block represents (stateInBlock, stateInElseBlock)
}

// ParseFile uses a state machine approach, handles line continuations (\), and multi-line blocks.
func ParseFile(r io.Reader) ([]Procedure, error) {
	scanner := bufio.NewScanner(r)
	const maxCapacity = 1024 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	procedures := []Procedure{}
	var currentProc *Procedure
	lineNum := 0
	var docstringLines []string
	state := stateTopLevel
	lastLineNumProcessed := 0

	var lineBuilder strings.Builder
	continuationActive := false

	var blockStack []*BlockContext

	for scanner.Scan() {
		lineNum++
		rawLine := scanner.Text()
		trimmedRawLine := strings.TrimSpace(rawLine)

		// Handle line continuation
		if strings.HasSuffix(trimmedRawLine, "\\") {
			lineToAppend := strings.TrimSpace(strings.TrimSuffix(trimmedRawLine, "\\"))
			lineBuilder.WriteString(lineToAppend) // Just concat, rely on spaces within lines or + operator
			continuationActive = true
			continue
		}

		var fullLine string
		if continuationActive {
			lineBuilder.WriteString(rawLine) // Append final line raw
			fullLine = lineBuilder.String()
			lineBuilder.Reset()
			continuationActive = false
		} else {
			fullLine = rawLine
		}

		currentLineNumForProcessing := lineNum
		lastLineNumProcessed = currentLineNumForProcessing

		// Trim comments AFTER joining lines
		trimmedLine := strings.TrimSpace(trimComments(fullLine))

		// Handle Docstring State separately first
		if state == stateInDocstring {
			if trimmedLine == "END" || strings.EqualFold(trimmedLine, "END COMMENT") {
				if currentProc == nil {
					return nil, fmt.Errorf("L%d: internal error: END docstring but no current proc", currentLineNumForProcessing)
				}
				doc, err := parseDocstringBlock(docstringLines)
				if err != nil {
					return nil, fmt.Errorf("L%d (around docstring): docstring block parsing error: %w", currentLineNumForProcessing, err)
				}
				currentProc.Docstring = doc
				state = stateInProcedure // Transition back
				docstringLines = nil
			} else {
				docstringLines = append(docstringLines, rawLine) // Append raw line
			}
			continue // Finished processing for docstring state
		}

		// Skip blank lines universally (now that docstring is handled)
		if trimmedLine == "" {
			continue
		}

		// Try parsing the line into a step (unless it's top level or needs special handling)
		var parsedStep Step
		var parseErr error
		needsParsing := true

		// --- Main State Machine ---
		//	processState: // Label to allow breaking to re-process END in parent block

		// Check for special keywords BEFORE calling ParseStep in relevant states
		switch state {
		case stateTopLevel, stateProcedureDone:
			if strings.HasPrefix(trimmedLine, "DEFINE PROCEDURE") {
				proc, err := parseProcedureHeader(fullLine)
				if err != nil {
					return nil, fmt.Errorf("L%d: header error: %w", currentLineNumForProcessing, err)
				}
				currentProc = &proc
				state = stateInProcedure
				needsParsing = false // Handled header, don't call ParseStep
			} else { // Any other non-blank/comment line is an error
				return nil, fmt.Errorf("L%d: unexpected statement outside procedure definition: %s", currentLineNumForProcessing, trimmedLine)
			}

		case stateInProcedure:
			if strings.HasPrefix(trimmedLine, "COMMENT:") {
				if currentProc == nil {
					return nil, fmt.Errorf("L%d: internal error: COMMENT: found but no current procedure", currentLineNumForProcessing)
				}
				if currentProc.Docstring.Purpose != "" {
					return nil, fmt.Errorf("L%d: duplicate COMMENT: block found in procedure '%s'", currentLineNumForProcessing, currentProc.Name)
				}
				if len(currentProc.Steps) > 0 {
					return nil, fmt.Errorf("L%d: COMMENT: block must appear before any steps in procedure '%s'", currentLineNumForProcessing, currentProc.Name)
				}
				state = stateInDocstring
				docstringLines = []string{}
				needsParsing = false // Handled COMMENT:, don't call ParseStep
			} else if trimmedLine == "END" {
				// This END terminates the procedure itself
				if currentProc == nil {
					return nil, fmt.Errorf("L%d: internal error: END found but no current procedure", currentLineNumForProcessing)
				}
				if err := validateDocstring(currentProc.Docstring); err != nil {
					return nil, fmt.Errorf("L%d (proc '%s'): docstring validation failed: %w", currentLineNumForProcessing, currentProc.Name, err)
				}
				procedures = append(procedures, *currentProc)
				currentProc = nil
				state = stateProcedureDone
				needsParsing = false // Handled END, don't call ParseStep
			}
			// If not COMMENT: or END, needsParsing remains true

		case stateInBlock: // Also covers stateInElseBlock conceptually now
			if trimmedLine == "END" {
				// This END terminates the current block
				needsParsing = true // Let ParseStep identify END_BLOCK type
			}
			// Otherwise, needsParsing remains true to parse the step inside the block
		}

		// Call ParseStep if needed for the current state and line content
		if needsParsing {
			parsedStep, parseErr = ParseStep(fullLine)
			if parseErr != nil {
				return nil, fmt.Errorf("L%d: step parse error ('%s'): %w", currentLineNumForProcessing, trimmedLine, parseErr)
			}
			// Ignore empty steps resulting from comment-only lines passed to ParseStep
			if parsedStep.Type == "" {
				continue
			}
		} else {
			// If needsParsing was false, skip the rest of the state processing for this line
			continue
		}

		// --- State Transitions based on Parsed Step (if parsing occurred) ---
		switch state {
		// stateTopLevel, stateProcedureDone were handled above if needsParsing was false
		// stateInDocstring was handled separately

		case stateInProcedure:
			// NeedsParsing was true, ParseStep was called.
			if parsedStep.Type == "IF" || parsedStep.Type == "WHILE" || parsedStep.Type == "FOR" {
				newContext := &BlockContext{BlockStep: &parsedStep, Steps: []Step{}, StateType: stateInBlock}
				blockStack = append(blockStack, newContext)
				state = stateInBlock
			} else if parsedStep.Type == "ELSE" {
				return nil, fmt.Errorf("L%d: unexpected ELSE statement (must follow IF block END)", currentLineNumForProcessing)
			} else if parsedStep.Type == "END_BLOCK" {
				// END should have been caught by the specific check for `trimmedLine == "END"` above
				return nil, fmt.Errorf("L%d: internal error: END_BLOCK received unexpectedly in stateInProcedure", currentLineNumForProcessing)
			} else { // Regular step
				if currentProc == nil {
					return nil, fmt.Errorf("L%d: internal error: trying to add step but no current procedure", currentLineNumForProcessing)
				}
				currentProc.Steps = append(currentProc.Steps, parsedStep)
			}

		case stateInBlock: // Covers nested blocks and ELSE blocks
			if parsedStep.Type == "END_BLOCK" { // Standalone END closes the current block
				if len(blockStack) == 0 {
					return nil, fmt.Errorf("L%d: unexpected END statement outside of block", currentLineNumForProcessing)
				}

				currentBlock := blockStack[len(blockStack)-1]
				blockStack = blockStack[:len(blockStack)-1] // Pop

				currentBlock.BlockStep.Value = currentBlock.Steps // Store []Step

				// Add the completed block step to the parent's context
				if len(blockStack) > 0 { // Parent is another block
					parentBlock := blockStack[len(blockStack)-1]
					parentBlock.Steps = append(parentBlock.Steps, *currentBlock.BlockStep)
					state = parentBlock.StateType // Use parent's specific state type
				} else { // Parent is the main procedure body
					if currentProc == nil {
						return nil, fmt.Errorf("L%d: internal error: closing block but no current procedure", currentLineNumForProcessing)
					}
					currentProc.Steps = append(currentProc.Steps, *currentBlock.BlockStep)
					state = stateInProcedure
				}
				// Future: Add stateMaybeElse logic here if currentBlock.BlockStep.Type == "IF"

			} else if parsedStep.Type == "IF" || parsedStep.Type == "WHILE" || parsedStep.Type == "FOR" {
				// Start a nested block
				if len(blockStack) == 0 {
					return nil, fmt.Errorf("L%d: internal error: block state inconsistency", currentLineNumForProcessing)
				}
				// Parent state is stateInBlock (or conceptually stateInElseBlock)
				//		parentState := blockStack[len(blockStack)-1].StateType
				newContext := &BlockContext{BlockStep: &parsedStep, Steps: []Step{}, StateType: stateInBlock} // Nested blocks are just stateInBlock
				blockStack = append(blockStack, newContext)
				state = stateInBlock // Enter nested block state

			} else if parsedStep.Type == "ELSE" {
				// Allow ELSE only if we just closed an IF block? No, handle that later.
				// For now, ELSE cannot be nested directly inside another block's steps.
				return nil, fmt.Errorf("L%d: unexpected ELSE statement inside block", currentLineNumForProcessing)
			} else { // Regular step inside a block
				if len(blockStack) == 0 {
					return nil, fmt.Errorf("L%d: internal error: block state inconsistency", currentLineNumForProcessing)
				}
				currentBlock := blockStack[len(blockStack)-1]
				currentBlock.Steps = append(currentBlock.Steps, parsedStep)
				// State remains the current block's state (e.g., stateInBlock)
			}
		} // End switch state for transitions

	} // End scanner loop

	// --- After Loop (EOF) ---
	if err := scanner.Err(); err != nil {
		if err == bufio.ErrTooLong {
			return nil, fmt.Errorf("scanner error: line exceeded buffer capacity (%d bytes)", maxCapacity)
		}
		return nil, fmt.Errorf("scanner error at end of file: %w", err)
	}
	if continuationActive {
		return nil, fmt.Errorf("L%d: unexpected EOF after line continuation character '\\'", lineNum)
	}
	if len(blockStack) > 0 {
		return nil, fmt.Errorf("L%d: unexpected EOF: missing 'END' for %s block", lastLineNumProcessed, blockStack[len(blockStack)-1].BlockStep.Type)
	}

	switch state {
	case stateInDocstring:
		return nil, fmt.Errorf("L%d: unexpected EOF: missing 'END' for docstring block", lastLineNumProcessed)
	case stateInProcedure:
		if currentProc != nil {
			return nil, fmt.Errorf("L%d: unexpected EOF: missing 'END' for procedure '%s'", lastLineNumProcessed, currentProc.Name)
		}
		return nil, fmt.Errorf("L%d: internal error: EOF in procedure state but no procedure active", lastLineNumProcessed)
	case stateTopLevel, stateProcedureDone:
		return procedures, nil
	default:
		return nil, fmt.Errorf("L%d: internal error: unknown parser state (%d) at EOF", lastLineNumProcessed, state)
	}
}

// --- Header Parsing --- (Unchanged)
func parseProcedureHeader(line string) (Procedure, error) {
	definePrefix := "DEFINE PROCEDURE"
	trimmedLine := trimComments(line)
	trimmedLine = strings.TrimSpace(trimmedLine)
	if !strings.HasPrefix(trimmedLine, definePrefix) {
		return Procedure{}, fmt.Errorf("invalid procedure header: does not start with 'DEFINE PROCEDURE': %q", line)
	}
	rest := strings.TrimSpace(trimmedLine[len(definePrefix):])
	openParen := strings.Index(rest, "(")
	closeParen := strings.LastIndex(rest, ")")
	if openParen == -1 {
		return Procedure{}, fmt.Errorf("invalid procedure header: missing '(' for parameter list in '%s'", rest)
	}
	if !strings.HasSuffix(strings.TrimSpace(rest), ")") {
		return Procedure{}, fmt.Errorf("invalid procedure header: missing or misplaced ')' after parameter list in '%s'", rest)
	}
	tempTrimmed := strings.TrimSpace(rest)
	if tempTrimmed[len(tempTrimmed)-1] != ')' {
		return Procedure{}, fmt.Errorf("invalid procedure header: content after closing parenthesis ')' in '%s'", rest)
	}
	closeParen = strings.LastIndex(rest, ")")
	if closeParen < openParen {
		return Procedure{}, fmt.Errorf("internal parser error: ')' appears before '(' despite checks in '%s'", rest)
	}
	name := strings.TrimSpace(rest[:openParen])
	if name == "" {
		return Procedure{}, fmt.Errorf("invalid procedure header: procedure name cannot be empty")
	}
	if !isValidIdentifier(name) {
		return Procedure{}, fmt.Errorf("invalid procedure header: invalid procedure name '%s'", name)
	}
	paramsPart := strings.TrimSpace(rest[openParen+1 : closeParen])
	var params []string
	if paramsPart != "" {
		params = splitParams(paramsPart)
		for i, p := range params {
			if !isValidIdentifier(p) {
				return Procedure{}, fmt.Errorf("invalid procedure header: invalid parameter name '%s' (index %d) in '%s'", p, i, paramsPart)
			}
		}
	} else {
		params = []string{}
	}
	return Procedure{Name: name, Params: params}, nil
}

// --- Docstring Parsing --- (Unchanged)
func parseDocstringBlock(lines []string) (Docstring, error) {
	doc := Docstring{Inputs: make(map[string]string)}
	currentSection := ""
	var sectionContent strings.Builder
	sectionHeaderRegex := regexp.MustCompile(`(?i)^\s*(PURPOSE|INPUTS|OUTPUT|ALGORITHM|CAVEATS|EXAMPLES)\s*:\s*(.*)$`)
	start := 0
	end := len(lines) - 1
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	for end >= start && strings.TrimSpace(lines[end]) == "" {
		end--
	}
	trimmedLines := lines[start : end+1]
	for i, line := range trimmedLines {
		match := sectionHeaderRegex.FindStringSubmatch(line)
		isNewSection := len(match) == 3
		if isNewSection {
			newSectionKey := strings.ToUpper(strings.TrimSpace(match[1]))
			restOfLine := match[2]
			if currentSection != "" {
				if err := saveSectionContent(currentSection, &sectionContent, &doc); err != nil {
					return doc, fmt.Errorf("saving section '%s' (around line %d): %w", currentSection, i, err)
				}
			}
			currentSection = newSectionKey
			sectionContent.Reset()
			sectionContent.WriteString(restOfLine)
		} else {
			if currentSection == "" {
				if strings.TrimSpace(line) != "" {
					return doc, fmt.Errorf("line %d: content before first section header: %q", i+1, line)
				}
			} else {
				if sectionContent.Len() > 0 {
					sectionContent.WriteString("\n")
				}
				sectionContent.WriteString(line)
			}
		}
	}
	if currentSection != "" {
		if err := saveSectionContent(currentSection, &sectionContent, &doc); err != nil {
			return doc, fmt.Errorf("saving last section '%s': %w", currentSection, err)
		}
	}
	return doc, nil
}

func saveSectionContent(section string, content *strings.Builder, doc *Docstring) error {
	if section == "" {
		return nil
	}
	rawContent := content.String()
	finalContent := strings.TrimSpace(rawContent)
	switch section {
	case "PURPOSE":
		if doc.Purpose != "" {
			return fmt.Errorf("duplicate PURPOSE section")
		}
		doc.Purpose = finalContent
	case "OUTPUT":
		if doc.Output != "" {
			return fmt.Errorf("duplicate OUTPUT section")
		}
		doc.Output = finalContent
	case "ALGORITHM":
		if doc.Algorithm != "" {
			return fmt.Errorf("duplicate ALGORITHM section")
		}
		doc.Algorithm = finalContent
	case "CAVEATS":
		if doc.Caveats != "" {
			return fmt.Errorf("duplicate CAVEATS section")
		}
		doc.Caveats = finalContent
	case "EXAMPLES":
		if doc.Examples != "" {
			return fmt.Errorf("duplicate EXAMPLES section")
		}
		doc.Examples = finalContent
	case "INPUTS":
		if len(doc.Inputs) > 0 {
			return fmt.Errorf("duplicate INPUTS section or content before parsing")
		}
		err := processInputBlock(rawContent, doc.Inputs)
		if err != nil {
			return fmt.Errorf("parsing INPUTS section: %w", err)
		}
	default:
		return fmt.Errorf("internal error: unknown section '%s'", section)
	}
	return nil
}

func processInputBlock(content string, inputsMap map[string]string) error {
	trimmedContent := strings.TrimSpace(content)
	if strings.EqualFold(trimmedContent, "none") {
		return nil
	}
	lines := strings.Split(content, "\n")
	var currentInputName string
	var currentInputDesc strings.Builder
	inputLineRegex := regexp.MustCompile(`^\s*-\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*(?:\([^)]+\))?\s*:\s*(.*)$`)
	foundAnyInput := false
	saveCurrentInput := func() error {
		if currentInputName != "" {
			desc := strings.TrimSpace(currentInputDesc.String())
			if _, exists := inputsMap[currentInputName]; exists {
				return fmt.Errorf("duplicate input parameter name '%s'", currentInputName)
			}
			if desc == "" {
				fmt.Printf("[Warning] Input parameter '%s' has no description.\n", currentInputName)
			}
			inputsMap[currentInputName] = desc
			currentInputName = ""
			currentInputDesc.Reset()
		}
		return nil
	}
	for i, line := range lines {
		match := inputLineRegex.FindStringSubmatch(line)
		if len(match) == 3 {
			if err := saveCurrentInput(); err != nil {
				return fmt.Errorf("line %d: %w", i+1, err)
			}
			foundAnyInput = true
			currentInputName = strings.TrimSpace(match[1])
			descPart := match[2]
			if !isValidIdentifier(currentInputName) {
				return fmt.Errorf("line %d: invalid input parameter name '%s'", i+1, currentInputName)
			}
			currentInputDesc.WriteString(descPart)
		} else if currentInputName != "" {
			if strings.TrimSpace(line) != "" && (len(line) > 0 && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t")) {
				if err := saveCurrentInput(); err != nil {
					return fmt.Errorf("line %d: %w", i+1, err)
				}
				return fmt.Errorf("line %d: malformed INPUTS content, expected new '- param: desc' or indented continuation, got: %q", i+1, line)
			} else if strings.TrimSpace(line) != "" {
				if currentInputDesc.Len() > 0 {
					currentInputDesc.WriteString("\n")
				}
				currentInputDesc.WriteString(line)
			}
		} else {
			if strings.TrimSpace(line) != "" {
				return fmt.Errorf("line %d: malformed INPUTS content, expected '- param: description', got: %q", i+1, line)
			}
		}
	}
	if err := saveCurrentInput(); err != nil {
		return err
	}
	if !foundAnyInput && trimmedContent != "" {
		return fmt.Errorf("INPUTS section has content but no valid '- param: description' lines found")
	}
	return nil
}

func validateDocstring(doc Docstring) error {
	missing := []string{}
	if doc.Purpose == "" {
		missing = append(missing, "PURPOSE")
	}
	if doc.Output == "" {
		missing = append(missing, "OUTPUT")
	}
	if doc.Algorithm == "" {
		missing = append(missing, "ALGORITHM")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required docstring section(s): %s", strings.Join(missing, ", "))
	}
	return nil
}
