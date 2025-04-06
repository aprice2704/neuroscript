// pkg/neurodata/checklist/checklist.go
package checklist

import (
	"fmt"
	"strings"

	// Import core package for Tool types and Interpreter interface
	"github.com/aprice2704/neuroscript/pkg/core"

	"github.com/antlr4-go/antlr/v4"
	// Adjust import path if your generated code location differs
	generated "github.com/aprice2704/neuroscript/pkg/neurodata/checklist/generated"
)

// === PARSER LOGIC (from previous parser.go) ===

// ChecklistItem represents a single parsed item.
type ChecklistItem struct {
	Text   string
	Status string // "pending" or "done"
}

// checklistListener is an ANTLR listener specifically for extracting checklist items.
type checklistListener struct {
	*generated.BaseNeuroDataChecklistListener                 // Embed the generated base listener
	Items                                     []ChecklistItem // Slice to store parsed items
	err                                       error           // To capture parsing errors within the listener
}

// newChecklistListener creates a new listener instance.
func newChecklistListener() *checklistListener {
	return &checklistListener{
		Items: make([]ChecklistItem, 0),
	}
}

// EnterItemLine is called by the ANTLR walker when entering an itemLine rule.
func (l *checklistListener) EnterItemLine(ctx *generated.ItemLineContext) {
	if l.err != nil {
		return // Don't process if an error already occurred
	}
	fmt.Println("[DEBUG CL Parser] EnterItemLine:", ctx.GetText()) // Debug output

	markToken := ctx.MARK()
	textToken := ctx.TEXT()

	if markToken == nil || textToken == nil {
		l.err = fmt.Errorf("internal parser error: missing MARK or TEXT in itemLine context: %q", ctx.GetText())
		fmt.Println("[ERROR CL Parser] Listener:", l.err) // Debug output
		return
	}

	mark := markToken.GetText()
	text := textToken.GetText()

	status := "pending"
	if strings.ToLower(mark) == "x" {
		status = "done"
	}

	newItem := ChecklistItem{
		Text:   strings.TrimSpace(text),
		Status: status,
	}
	l.Items = append(l.Items, newItem)
	fmt.Printf("[DEBUG CL Parser] Parsed Item: %+v\n", newItem) // Debug output
}

// parseChecklistANTLR uses the generated ANTLR parser to parse checklist content.
// It's kept unexported as it's only called by the tool function within this package.
// Returns a slice of maps matching the old tool's output format, or an error.
func parseChecklistANTLR(content string) ([]map[string]interface{}, error) {
	fmt.Println("[DEBUG CL Parser] Starting parseChecklistANTLR") // Debug output
	// Setup ANTLR input stream and lexer
	inputStream := antlr.NewInputStream(content)
	lexer := generated.NewNeuroDataChecklistLexer(inputStream)

	// TODO: Replace with a custom error listener that collects errors.
	errorListener := antlr.NewDiagnosticErrorListener(true)
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(errorListener)

	// Create token stream and parser
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	parser := generated.NewNeuroDataChecklistParser(tokenStream)
	parser.RemoveErrorListeners()
	parser.AddErrorListener(errorListener)

	// Parse the input content starting from the checklistFile rule
	tree := parser.ChecklistFile()

	// TODO: Check errors collected by a custom error listener here.

	// Create our custom listener and walk the parse tree
	listener := newChecklistListener()
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)

	// Check for errors captured *during* the walk by our listener
	if listener.err != nil {
		return nil, fmt.Errorf("error during parse tree walk: %w", listener.err)
	}

	// Convert []ChecklistItem to []map[string]interface{}
	resultMaps := make([]map[string]interface{}, len(listener.Items))
	for i, item := range listener.Items {
		resultMaps[i] = map[string]interface{}{
			"text":   item.Text,
			"status": item.Status,
		}
	}

	fmt.Printf("[DEBUG CL Parser] parseChecklistANTLR finished. Found %d items.\n", len(resultMaps)) // Debug output
	return resultMaps, nil
}

// === TOOL LOGIC (from previous tool.go) ===

// RegisterChecklistTools adds checklist-specific tools to the core registry.
// This function is called from gonsi/main.go
func RegisterChecklistTools(registry *core.ToolRegistry) {
	registry.RegisterTool(core.ToolImplementation{
		Spec: core.ToolSpec{
			Name:        "ChecklistParse", // New tool name
			Description: "Parses a string formatted as a NeuroData checklist (lines starting with '- [ ]' or '- [x]') into a list of maps using ANTLR. Each map contains 'text' and 'status' ('pending' or 'done'). Ignores non-item lines.",
			Args: []core.ArgSpec{
				{Name: "content", Type: core.ArgTypeString, Required: true, Description: "The string containing the checklist items."},
			},
			ReturnType: core.ArgTypeSliceAny, // Returns []map[string]interface{}
		},
		Func: toolChecklistParse, // Calls the function below
	})
	// Add other checklist-related tools here in the future
}

// toolChecklistParse is the implementation for the TOOL.ChecklistParse function.
func toolChecklistParse(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	// Validation handled by core.ValidateAndConvertArgs
	content := args[0].(string)

	logger := interpreter.Logger() // Get logger safely
	if logger != nil {
		logSnippet := content
		if len(logSnippet) > 100 {
			logSnippet = logSnippet[:100] + "..."
		}
		logger.Printf("[DEBUG TOOL] Calling TOOL.ChecklistParse on content (snippet): %q", logSnippet)
	}

	// Call the ANTLR parsing function (now in the same file)
	parsedItems, err := parseChecklistANTLR(content)

	if err != nil {
		errMsg := fmt.Sprintf("ChecklistParse failed: %s", err.Error())
		if logger != nil {
			logger.Printf("[ERROR TOOL] TOOL.ChecklistParse failed: %s", err.Error())
		}
		// Return the error message string for the NeuroScript caller
		return errMsg, nil
	}

	if logger != nil {
		logger.Printf("[DEBUG TOOL] TOOL.ChecklistParse successful. Found %d items.", len(parsedItems))
	}
	// Return the slice of maps on success
	return parsedItems, nil
}
