// NeuroScript Version: 0.8.0
// File version: 12
// Purpose: BUGFIX: Passes line prefix length into formatExpression to enable context-aware line wrapping.
// filename: pkg/nsfmt/format.go
// nlines: 295

package nsfmt

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

const indentString = "    "
const maxLineLength = 80 // Global line length for wrapping

// Format takes raw NeuroScript source code, parses it, and returns the
// canonically formatted version.
func Format(source []byte) ([]byte, error) {
	logger := logging.NewNoOpLogger() // We don't want the formatter to log errors, just return them.

	// 1. Parse the source into an ANTLR tree.
	parserAPI := parser.NewParserAPI(logger)
	antlrTree, tokenStream, pErr := parserAPI.ParseAndGetStream("format.ns", string(source))
	if pErr != nil {
		return nil, fmt.Errorf("syntax error, cannot format: %w", pErr)
	}

	// 2. Build the high-level NeuroScript AST.
	astBuilder := parser.NewASTBuilder(logger)
	program, _, bErr := astBuilder.BuildFromParseResult(antlrTree, tokenStream)
	if bErr != nil {
		return nil, fmt.Errorf("ast build error, cannot format: %w", bErr)
	}

	// 3. Create a formatter and walk the AST to pretty-print.
	f := &formatter{
		builder: &bytes.Buffer{},
		indent:  0,
	}

	f.formatProgram(program)

	return f.builder.Bytes(), nil
}

// formatter holds the state for the recursive pretty-printer.
type formatter struct {
	builder *bytes.Buffer
	indent  int
}

func (f *formatter) write(s string) {
	f.builder.WriteString(s)
}

func (f *formatter) writeLine(s string) {
	if s == "" { // Write a blank line
		f.builder.WriteString("\n")
		return
	}

	// FIX: Handle multi-line strings (from formatExpression) correctly.
	lines := strings.Split(s, "\n")

	// Write the first line *with* the formatter's indent
	for i := 0; i < f.indent; i++ {
		f.builder.WriteString(indentString)
	}
	f.builder.WriteString(lines[0])
	f.builder.WriteString("\n")

	// Write all *subsequent* lines *without* the formatter's indent,
	// as they are already indented by formatExpression.
	if len(lines) > 1 {
		for _, line := range lines[1:] {
			f.builder.WriteString(line)
			f.builder.WriteString("\n")
		}
	}
}

func (f *formatter) formatProgram(prog *ast.Program) {
	// 1. Format top-level comments (These are usually file headers)
	f.formatComments(prog.Comments, true)

	// 2. Format Metadata
	f.formatMetadata(prog.Metadata)
	if len(prog.Metadata) > 0 {
		f.writeLine("")
	}

	// 3. Format Procedures
	if len(prog.Procedures) > 0 {
		procNames := make([]string, 0, len(prog.Procedures))
		for name := range prog.Procedures {
			procNames = append(procNames, name)
		}
		sort.Strings(procNames)
		for i, name := range procNames {
			if i > 0 {
				f.writeLine("")
			}
			f.formatProcedure(prog.Procedures[name])
		}
	}

	// 4. Format Commands
	if len(prog.Commands) > 0 {
		if len(prog.Procedures) > 0 {
			f.writeLine("")
		}
		for i, cmd := range prog.Commands {
			if i > 0 {
				f.writeLine("")
			}
			f.formatCommand(cmd)
		}
	}

	// 5. Format Events
	if len(prog.Events) > 0 {
		if len(prog.Procedures) > 0 || len(prog.Commands) > 0 {
			f.writeLine("")
		}
		for i, event := range prog.Events {
			if i > 0 {
				f.writeLine("")
			}
			f.formatEvent(event)
		}
	}
}

func (f *formatter) formatMetadata(meta map[string]string) {
	if len(meta) == 0 {
		return
	}
	keys := make([]string, 0, len(meta))
	for k := range meta {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		f.writeLine(fmt.Sprintf(":: %s: %s", k, meta[k]))
	}
}

func (f *formatter) formatComments(comments []*ast.Comment, standalone bool) {
	for _, c := range comments {
		if standalone {
			f.writeLine(c.Text)
		} else {
			f.write(c.Text) // Trailing comment
		}
	}
}

// formatProcedureSignature formats the (needs ..., optional ..., returns ...) part
// of a function signature, applying smart multi-line formatting.
func (f *formatter) formatProcedureSignature(proc *ast.Procedure) string {
	var parts []string
	if len(proc.RequiredParams) > 0 {
		parts = append(parts, fmt.Sprintf("needs %s", strings.Join(proc.RequiredParams, ", ")))
	}
	if len(proc.OptionalParams) > 0 {
		optNames := make([]string, len(proc.OptionalParams))
		for i, p := range proc.OptionalParams {
			optNames[i] = p.Name // Note: We don't format default values here yet
		}
		parts = append(parts, fmt.Sprintf("optional %s", strings.Join(optNames, ", ")))
	}
	if len(proc.ReturnVarNames) > 0 {
		parts = append(parts, fmt.Sprintf("returns %s", strings.Join(proc.ReturnVarNames, ", ")))
	}

	if len(parts) == 0 {
		return ""
	}

	// 1. Try single-line format
	singleLineParams := strings.Join(parts, " ")

	// *** BUGFIX ***
	// We must check the length of the *entire* line, not just the params.
	// This includes indentation, "func ", name, "()", and " means".
	indentLen := f.indent * len(indentString)
	fullSingleLineLen := indentLen + len("func ") + len(proc.Name()) + len("(") + len(singleLineParams) + len(") means")

	if fullSingleLineLen <= maxLineLength {
		return singleLineParams // Return *only* the params part
	}

	// 2. Build multi-line format
	var bMulti strings.Builder
	// Indent for the parameters themselves (one level deeper than the func)
	itemIndentStr := strings.Repeat(indentString, f.indent+1)
	// Indent for the closing parenthesis (same level as the func)
	closingIndentStr := strings.Repeat(indentString, f.indent)

	bMulti.WriteString(" \\\n") // Start with line continuation
	for i, part := range parts {
		bMulti.WriteString(itemIndentStr)
		bMulti.WriteString(part)
		if i < len(parts)-1 {
			bMulti.WriteString(" \\\n")
		} else {
			bMulti.WriteString(" \\\n") // Also add to last line before closing paren
		}
	}
	bMulti.WriteString(closingIndentStr)
	return bMulti.String()
}

func (f *formatter) formatProcedure(proc *ast.Procedure) {
	f.formatComments(proc.Comments, true)

	// Build signature
	sig := f.formatProcedureSignature(proc)

	if sig != "" {
		// The `writeLine` function will correctly handle both single-line
		// and multi-line signatures (which start with " \n").
		f.writeLine(fmt.Sprintf("func %s(%s) means", proc.Name(), sig))
	} else {
		f.writeLine(fmt.Sprintf("func %s() means", proc.Name()))
	}

	f.indent++
	f.formatMetadata(proc.Metadata)
	f.formatStepList(proc.Steps)
	// Format error handlers at the end of the step list
	for _, eh := range proc.ErrorHandlers {
		f.formatErrorHandler(eh)
	}
	f.indent--
	f.writeLine("endfunc")
}

func (f *formatter) formatCommand(cmd *ast.CommandNode) {
	f.formatComments(cmd.Comments, true)
	f.writeLine("command")
	f.indent++
	f.formatMetadata(cmd.Metadata)
	f.formatStepList(cmd.Body)
	// Format error handlers at the end of the step list
	for _, eh := range cmd.ErrorHandlers {
		f.formatErrorHandler(eh)
	}
	f.indent--
	f.writeLine("endcommand")
}

func (f *formatter) formatEvent(event *ast.OnEventDecl) {
	f.formatComments(event.Comments, true)
	var sig strings.Builder
	// Pass prefixLen=0; event name is just a string, not a wrapping list
	sig.WriteString(f.formatExpression(event.EventNameExpr, 0))
	if event.HandlerName != "" {
		sig.WriteString(fmt.Sprintf(" named %q", event.HandlerName))
	}
	if event.EventVarName != "" {
		sig.WriteString(fmt.Sprintf(" as %s", event.EventVarName))
	}

	f.writeLine(fmt.Sprintf("on event %s do", sig.String()))
	f.indent++
	f.formatMetadata(event.Metadata)
	f.formatStepList(event.Body)
	f.indent--
	f.writeLine("endon")
}

func (f *formatter) formatErrorHandler(handler *ast.Step) {
	f.formatComments(handler.Comments, true)
	f.writeLine("on error do")
	f.indent++
	f.formatStepList(handler.Body)
	f.indent--
	f.writeLine("endon")
}

func (f *formatter) formatStepList(steps []ast.Step) {
	for i, step := range steps {
		// Add a blank line before comments that precede a step
		// (unless it's the first step in the block)
		if i > 0 && len(step.Comments) > 0 {
			f.writeLine("")
		}
		f.formatStep(&step)
	}
}

func (f *formatter) formatStep(step *ast.Step) {
	f.formatComments(step.Comments, true)
	indentLen := f.indent * len(indentString)

	switch step.Type {
	case "set":
		lvals := make([]string, len(step.LValues))
		for i, lv := range step.LValues {
			lvals[i] = f.formatExpression(lv, 0) // prefix 0 for lvals
		}
		prefixStr := fmt.Sprintf("set %s = ", strings.Join(lvals, ", "))
		prefixLen := indentLen + len(prefixStr)
		rval := f.formatExpression(step.Values[0], prefixLen)
		f.writeLine(fmt.Sprintf("%s%s", prefixStr, rval))

	case "if":
		prefixStr := "if "
		prefixLen := indentLen + len(prefixStr)
		condStr := f.formatExpression(step.Cond, prefixLen)
		f.writeLine(fmt.Sprintf("%s%s", prefixStr, condStr))

		f.indent++
		f.formatStepList(step.Body)
		f.indent--
		if step.ElseBody != nil && len(step.ElseBody) > 0 {
			f.writeLine("else")
			f.indent++
			f.formatStepList(step.ElseBody)
			f.indent--
		}
		f.writeLine("endif")

	case "for":
		prefixStr := fmt.Sprintf("for each %s in ", step.LoopVarName)
		prefixLen := indentLen + len(prefixStr)
		collStr := f.formatExpression(step.Collection, prefixLen)
		f.writeLine(fmt.Sprintf("%s%s", prefixStr, collStr))

		f.indent++
		f.formatStepList(step.Body)
		f.indent--
		f.writeLine("endfor")

	case "while":
		prefixStr := "while "
		prefixLen := indentLen + len(prefixStr)
		condStr := f.formatExpression(step.Cond, prefixLen)
		f.writeLine(fmt.Sprintf("%s%s", prefixStr, condStr))

		f.indent++
		f.formatStepList(step.Body)
		f.indent--
		f.writeLine("endwhile")

	case "emit":
		prefixStr := "emit "
		prefixLen := indentLen + len(prefixStr)
		valStr := f.formatExpression(step.Values[0], prefixLen)
		f.writeLine(fmt.Sprintf("%s%s", prefixStr, valStr))

	case "return":
		vals := make([]string, len(step.Values))
		for i, v := range step.Values {
			vals[i] = f.formatExpression(v, 0) // prefix 0 for sub-expressions
		}
		if len(vals) > 0 {
			f.writeLine(fmt.Sprintf("return %s", strings.Join(vals, ", ")))
		} else {
			f.writeLine("return")
		}

	case "call":
		prefixStr := "call "
		prefixLen := indentLen + len(prefixStr)
		callStr := f.formatExpression(step.Call, prefixLen)
		f.writeLine(fmt.Sprintf("%s%s", prefixStr, callStr))

	case "must":
		prefixStr := "must "
		prefixLen := indentLen + len(prefixStr)
		condStr := f.formatExpression(step.Cond, prefixLen)
		f.writeLine(fmt.Sprintf("%s%s", prefixStr, condStr))

	case "fail":
		if step.Values != nil && len(step.Values) > 0 {
			prefixStr := "fail "
			prefixLen := indentLen + len(prefixStr)
			valStr := f.formatExpression(step.Values[0], prefixLen)
			f.writeLine(fmt.Sprintf("%s%s", prefixStr, valStr))
		} else {
			f.writeLine("fail")
		}

	case "ask":
		line := fmt.Sprintf("ask %s, %s", f.formatExpression(step.AskStmt.AgentModelExpr, 0), f.formatExpression(step.AskStmt.PromptExpr, 0))
		if step.AskStmt.WithOptions != nil {
			withPrefix := " with "
			// Calculate prefix for the 'with' expression
			prefixLen := indentLen + len(line) + len(withPrefix)
			withStr := f.formatExpression(step.AskStmt.WithOptions, prefixLen)
			line += fmt.Sprintf("%s%s", withPrefix, withStr)
		}
		if step.AskStmt.IntoTarget != nil {
			line += fmt.Sprintf(" into %s", f.formatExpression(step.AskStmt.IntoTarget, 0))
		}
		f.writeLine(line)

	case "promptuser":
		// prefix 0 for prompt, it's just a string lit or var
		promptStr := f.formatExpression(step.PromptUserStmt.PromptExpr, 0)
		// prefix 0 for target, it's just an identifier
		intoStr := f.formatExpression(step.PromptUserStmt.IntoTarget, 0)
		line := fmt.Sprintf("promptuser %s into %s", promptStr, intoStr)
		f.writeLine(line)

	case "whisper":
		// prefix 0 for handle, it's just a string lit or var
		handleStr := f.formatExpression(step.WhisperStmt.Handle, 0)
		prefixStr := fmt.Sprintf("whisper %s, ", handleStr)
		prefixLen := indentLen + len(prefixStr)
		// The value could be a long list, so it needs the prefix
		valueStr := f.formatExpression(step.WhisperStmt.Value, prefixLen)
		f.writeLine(fmt.Sprintf("%s%s", prefixStr, valueStr))

	case "clear_error":
		f.writeLine("clear_error")
	case "break":
		f.writeLine("break")
	case "continue":
		f.writeLine("continue")

	case "on_error":
		f.formatErrorHandler(step)

	default:
		f.writeLine(fmt.Sprintf("# FIXME: Unhandled step type in formatter: %s", step.Type))
	}
}
