// ast_builder_nslistener.go - Concrete listener that converts the ANTLR parse
// tree into an in-memory AST.
//
// Overview
// --------
// This type (`neuroScriptListenerImpl`) implements the ANTLR-generated
// `NeuroScriptListener` interface and overrides the callbacks we care about while the
// parser walks the grammar. It owns the two runtime stacks described in
// ast.go:
//   - valueStack      – LIFO []interface{} for every temporary AST value.
//   - blockStepStack  – LIFO []*[]Step tracking nested block bodies.
//
// These stacks _must_ be empty when `ExitProgram` returns; any residual entries
// are recorded as internal-error diagnostics.
//
// The file also provides helper methods for error reporting, debug logging, and
// simple stack access that are reused across the specialized *ast_builder_*.go
// files.
//
// All code below is private to the builder package; the public entry point is
// `Build(programText []byte, logger interfaces.Logger)` in ast_builder_main.go.
package core

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// neuroScriptListenerImpl implements the ANTLR listener and incrementally builds
// a Program AST while walking the parse tree.
type neuroScriptListenerImpl struct {
	// *gen.BaseNeuroScriptListener // Intentionally commented out to enforce full interface implementation.
	program              *Program
	fileMetadata         map[string]string // alias for program.Metadata
	procedures           []*Procedure      // temporary slice before final map assembly
	events               []*OnEventDecl    // temporary slice of event declarations
	currentProc          *Procedure
	currentSteps         *[]Step
	blockStepStack       []*[]Step     // LIFO of step slices for nested blocks
	valueStack           []interface{} // LIFO of AST fragments
	currentMapKey        *StringLiteralNode
	logger               interfaces.Logger
	debugAST             bool
	errors               []error
	loopDepth            int
	blockValueDepthStack []int
}

// Sentinel to enforce implementing whole interface
var _ gen.NeuroScriptListener = &neuroScriptListenerImpl{}

// newNeuroScriptListener constructs a fresh builder listener.
func newNeuroScriptListener(logger interfaces.Logger, debugAST bool) *neuroScriptListenerImpl {
	prog := &Program{
		Metadata:   make(map[string]string),
		Procedures: make(map[string]*Procedure),
		Events:     make([]*OnEventDecl, 0),
		Pos:        nil, // assigned in EnterProgram
	}
	return &neuroScriptListenerImpl{
		// BaseNeuroScriptListener: &gen.BaseNeuroScriptListener{},
		program:        prog,
		fileMetadata:   prog.Metadata,
		procedures:     make([]*Procedure, 0, 10),
		events:         make([]*OnEventDecl, 0, 5),
		blockStepStack: make([]*[]Step, 0, 5),
		valueStack:     make([]interface{}, 0, 20),
		logger:         logger,
		debugAST:       debugAST,
		errors:         make([]error, 0),
		loopDepth:      0,
	}
}

// ---------- Error-handling helpers ----------

func (l *neuroScriptListenerImpl) addError(ctx antlr.ParserRuleContext, format string, args ...interface{}) {
	var startToken antlr.Token
	if ctx != nil {
		startToken = ctx.GetStart()
	}
	pos := tokenToPosition(startToken)
	errMsg := fmt.Sprintf(format, args...)
	err := fmt.Errorf("AST build error at %s: %s", pos.String(), errMsg)
	for _, existing := range l.errors {
		if existing.Error() == err.Error() {
			l.logger.Debug("AST Build Error (Duplicate)", "position", pos.String(), "message", errMsg)
			return
		}
	}
	l.errors = append(l.errors, err)
	l.logger.Error("AST Build Error", "position", pos.String(), "message", errMsg)
}

func (l *neuroScriptListenerImpl) addErrorf(token antlr.Token, format string, args ...interface{}) {
	pos := tokenToPosition(token)
	errMsg := fmt.Sprintf(format, args...)
	err := fmt.Errorf("AST build error near %s: %s", pos.String(), errMsg)
	for _, existing := range l.errors {
		if existing.Error() == err.Error() {
			l.logger.Debug("AST Build Error (Duplicate)", "position", pos.String(), "message", errMsg)
			return
		}
	}
	l.errors = append(l.errors, err)
	l.logger.Error("AST Build Error", "position", pos.String(), "message", errMsg)
}

// ---------- Tiny stack utilities used by various callbacks ----------

func (l *neuroScriptListenerImpl) pop() interface{} {
	if len(l.valueStack) == 0 {
		return nil
	}
	idx := len(l.valueStack) - 1
	v := l.valueStack[idx]
	l.valueStack = l.valueStack[:idx]
	return v
}

func (l *neuroScriptListenerImpl) peek() interface{} {
	if len(l.valueStack) == 0 {
		return nil
	}
	return l.valueStack[len(l.valueStack)-1]
}

// ---------- Public accessors (used only in tests) ----------

func (l *neuroScriptListenerImpl) GetFileMetadata() map[string]string {
	if l.program != nil {
		return l.program.Metadata
	}
	l.logger.Warn("GetFileMetadata called with program == nil")
	return map[string]string{}
}

func (l *neuroScriptListenerImpl) GetResult() []*Procedure {
	l.logger.Warn("GetResult called; this returns the temporary slice, not the final map")
	return l.procedures
}

// ---------- Internal logging helper ----------

func (l *neuroScriptListenerImpl) logDebugAST(format string, v ...interface{}) {
	if l.debugAST {
		l.logger.Debug(fmt.Sprintf(format, v...))
	}
}

// ---------- Simple context helpers ----------

func (l *neuroScriptListenerImpl) isInsideLoop() bool { return l.loopDepth > 0 }

// ---------- ANTLR listener overrides (root rule only in this file) ----------

func (l *neuroScriptListenerImpl) EnterProgram(ctx *gen.ProgramContext) {
	l.logDebugAST(">>> Enter Program")

	// Reset state for a fresh build.
	if l.program == nil {
		l.program = &Program{Metadata: make(map[string]string), Procedures: make(map[string]*Procedure), Events: make([]*OnEventDecl, 0)}
	}
	l.program.Pos = tokenToPosition(ctx.GetStart())

	l.fileMetadata = l.program.Metadata
	l.procedures = l.procedures[:0]
	l.events = l.events[:0]
	l.errors = l.errors[:0]
	l.valueStack = l.valueStack[:0]
	l.blockStepStack = l.blockStepStack[:0]
	l.currentProc = nil
	l.currentSteps = nil
	l.currentMapKey = nil
	l.loopDepth = 0
}

func (l *neuroScriptListenerImpl) ExitProgram(ctx *gen.ProgramContext) {
	if l.program != nil {
		l.program.Events = l.events
	}

	l.logDebugAST("<<< Exit Program (meta=%d procs=%d handlers=%d errors=%d stack=%d)", len(l.fileMetadata), len(l.program.Procedures), len(l.program.Events), len(l.errors), len(l.valueStack))

	// Stack sanity checks.
	if len(l.valueStack) != 0 {
		l.errors = append(l.errors, fmt.Errorf("internal AST builder error: value stack size is %d at end of program", len(l.valueStack)))
		l.logger.Error("ExitProgram: value stack not empty", "size", len(l.valueStack))
	}
	if len(l.blockStepStack) != 0 {
		l.errors = append(l.errors, fmt.Errorf("internal AST builder error: block step stack size is %d at end of program", len(l.blockStepStack)))
		l.logger.Error("ExitProgram: blockStepStack not empty", "size", len(l.blockStepStack))
	}
}

// EnterEveryRule is called when entering any rule. Required to satisfy the interface.
func (l *neuroScriptListenerImpl) EnterEveryRule(ctx antlr.ParserRuleContext) {
	// This method can be used for very detailed, rule-by-rule debug logging.
	// It is intentionally left empty for now to avoid excessive log spam.
}

// ExitEveryRule is called when exiting any rule. Required to satisfy the interface.
func (l *neuroScriptListenerImpl) ExitEveryRule(ctx antlr.ParserRuleContext) {
	// This method can be used for very detailed, rule-by-rule debug logging.
	// It is intentionally left empty for now to avoid excessive log spam.
}

// VisitTerminal is called when a terminal node is visited.
func (l *neuroScriptListenerImpl) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (l *neuroScriptListenerImpl) VisitErrorNode(node antlr.ErrorNode) {}
