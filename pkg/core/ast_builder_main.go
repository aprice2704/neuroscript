// filename: pkg/core/ast_builder_main.go
package core

import (
	"fmt"     // Needed for Unquote if used elsewhere
	"strings" // Needed for metadata parsing

	"github.com/antlr4-go/antlr/v4" // Import antlr
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// --- ASTBuilder (Exported Constructor and Build Method) ---

// ASTBuilder encapsulates the logic for building the NeuroScript AST using a listener.
type ASTBuilder struct {
	logger   logging.Logger
	debugAST bool // Option to enable detailed AST construction logging
}

// NewASTBuilder creates a new ASTBuilder instance.
func NewASTBuilder(logger logging.Logger) *ASTBuilder {
	if logger == nil {
		logger = &coreNoOpLogger{}
		logger.Warn("ASTBuilder created with nil logger, using NoOpLogger.")
	}
	return &ASTBuilder{
		logger:   logger,
		debugAST: false, // Default to false, could be configurable
	}
}

// Build takes an ANTLR parse tree and constructs the NeuroScript Program AST (*core.Program).
func (b *ASTBuilder) Build(tree antlr.Tree) (*Program, error) {
	if tree == nil {
		return nil, fmt.Errorf("cannot build AST from nil parse tree")
	}
	b.logger.Debug("Starting AST build process using Listener.")

	// Create the listener instance.
	listener := newNeuroScriptListener(b.logger, b.debugAST)

	// Walk the parse tree with the listener.
	walker := antlr.NewParseTreeWalker()
	walker.Walk(listener, tree)

	// Assemble Program AST using collected metadata and procedures.
	programAST := &Program{
		Metadata:   listener.GetFileMetadata(), // Get collected metadata
		Procedures: listener.GetResult(),
	}

	b.logger.Debug("AST build process completed successfully.")
	// TODO: Add error collection/checking to the listener if needed. Currently returns nil error.
	return programAST, nil
}

// --- neuroScriptListenerImpl (Internal Listener Implementation) ---

// neuroScriptListenerImpl builds the AST using the Listener pattern.
type neuroScriptListenerImpl struct {
	*gen.BaseNeuroScriptListener
	fileMetadata   map[string]string // For file-level metadata
	procedures     []Procedure
	currentProc    *Procedure
	currentSteps   *[]Step
	blockStepStack []*[]Step
	valueStack     []interface{}
	currentMapKey  *StringLiteralNode // Assuming this is used for map literal parsing elsewhere
	blockSteps     map[antlr.ParserRuleContext][]Step
	// isInFileMetadataBlock bool // REMOVED: No longer needed with file_header rule
	logger   logging.Logger
	debugAST bool
}

// newNeuroScriptListener creates a new listener instance.
func newNeuroScriptListener(logger logging.Logger, debugAST bool) *neuroScriptListenerImpl {
	return &neuroScriptListenerImpl{
		fileMetadata:   make(map[string]string),
		procedures:     make([]Procedure, 0),
		blockStepStack: make([]*[]Step, 0),
		valueStack:     make([]interface{}, 0, 10),
		blockSteps:     make(map[antlr.ParserRuleContext][]Step),
		// isInFileMetadataBlock: false, // REMOVED
		logger:   logger,
		debugAST: debugAST,
	}
}

// --- Listener Getters ---

// GetFileMetadata returns the collected file-level metadata.
func (l *neuroScriptListenerImpl) GetFileMetadata() map[string]string {
	return l.fileMetadata
}

func (l *neuroScriptListenerImpl) GetResult() []Procedure { return l.procedures }

// --- Listener Stack Helpers (Unchanged) ---

func (l *neuroScriptListenerImpl) pushValue(v interface{}) {
	l.valueStack = append(l.valueStack, v)
	l.logDebugAST("      Pushed Value: %T %+v (Stack size: %d)", v, v, len(l.valueStack))
}

func (l *neuroScriptListenerImpl) popValue() (interface{}, bool) {
	if len(l.valueStack) == 0 {
		l.logger.Error("AST Builder: Pop from empty value stack!")
		return nil, false
	}
	index := len(l.valueStack) - 1
	value := l.valueStack[index]
	l.valueStack = l.valueStack[:index]
	l.logDebugAST("      Popped Value: %T %+v (Stack size: %d)", value, value, len(l.valueStack))
	return value, true
}

func (l *neuroScriptListenerImpl) popNValues(n int) ([]interface{}, bool) {
	if len(l.valueStack) < n {
		l.logger.Error("AST Builder: Stack underflow pop %d, have %d.", n, len(l.valueStack))
		return nil, false
	}
	startIndex := len(l.valueStack) - n
	values := make([]interface{}, n)
	if len(l.valueStack) >= startIndex+n {
		copy(values, l.valueStack[startIndex:])
	} else {
		l.logger.Error("AST Builder: Slice bounds out of range during popNValues.", "n", n, "stack_size", len(l.valueStack), "start_index", startIndex)
		return nil, false
	}
	l.valueStack = l.valueStack[:startIndex]
	l.logDebugAST("      Popped %d Values (Stack size: %d)", n, len(l.valueStack))
	return values, true
}

// --- Listener Logging Helper (Unchanged) ---

func (l *neuroScriptListenerImpl) logDebugAST(format string, v ...interface{}) {
	if l.debugAST {
		l.logger.Debug(format, v...)
	}
}

// --- Listener ANTLR Method Implementations ---

func (l *neuroScriptListenerImpl) EnterProgram(ctx *gen.ProgramContext) {
	l.logDebugAST(">>> Enter Program")
	l.procedures = make([]Procedure, 0)
	l.fileMetadata = make(map[string]string) // Reset file metadata
}

func (l *neuroScriptListenerImpl) ExitProgram(ctx *gen.ProgramContext) {
	l.logDebugAST("<<< Exit Program (Metadata Keys: %v)", MapKeysListener(l.fileMetadata))
}

// --- MODIFIED: Metadata Handling via file_header ---

// EnterFile_header processes all metadata lines found at the start of the file.
func (l *neuroScriptListenerImpl) EnterFile_header(ctx *gen.File_headerContext) {
	l.logDebugAST("  >> Enter File Header")
	// Iterate through all children of the header context
	for _, child := range ctx.GetChildren() {
		// Check if the child is a METADATA_LINE terminal node
		if termNode, ok := child.(antlr.TerminalNode); ok && termNode.GetSymbol().GetTokenType() == gen.NeuroScriptLexerMETADATA_LINE {
			lineText := termNode.GetText()
			l.logDebugAST("   - Processing File Metadata Line: %s", lineText)
			// Parse the line
			lineText = strings.TrimSpace(lineText)
			if strings.HasPrefix(lineText, "::") {
				trimmedLine := strings.TrimSpace(lineText[2:])
				parts := strings.SplitN(trimmedLine, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1]) // TODO: Handle potential quoting/escaping
					if key != "" {
						l.fileMetadata[key] = value
						l.logDebugAST("     Stored File Metadata: '%s' = '%s'", key, value)
					} else {
						l.logger.Warn("Ignoring file metadata line with empty key", "line", lineText)
					}
				} else {
					l.logger.Warn("Ignoring malformed file metadata line (missing or misplaced ':'?)", "line", lineText)
				}
			} else {
				// Should not happen if lexer rule is correct
				l.logger.Warn("Unexpected line format in file_header (missing '::'?)", "line", lineText)
			}
		}
		// Ignore NEWLINE tokens within the file_header
	}
}

func (l *neuroScriptListenerImpl) ExitFile_header(ctx *gen.File_headerContext) {
	l.logDebugAST("  << Exit File Header")
}

// REMOVED: Enter/ExitMetadata_block are no longer needed for file-level metadata
// func (l *neuroScriptListenerImpl) EnterMetadata_block(ctx *gen.Metadata_blockContext) { ... }
// func (l *neuroScriptListenerImpl) ExitMetadata_block(ctx *gen.Metadata_blockContext) { ... }
// Procedure-level metadata is handled within EnterProcedure_definition now.

// --- END MODIFIED Methods ---

// MapKeysListener is a helper function (consider moving to utils)
func MapKeysListener(m map[string]string) []string {
	if m == nil {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// IMPORTANT: Ensure that listener methods for procedure definitions, statements,
// and expressions are implemented correctly in other ast_builder_*.go files
// or within this file.
