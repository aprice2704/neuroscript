package core

import (
	"errors"
	"fmt"

	"github.com/antlr4-go/antlr/v4"
	gen "github.com/aprice2704/neuroscript/pkg/core/generated"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// --- neuroScriptListenerImpl (Internal Listener Implementation) ---
type neuroScriptListenerImpl struct {
	*gen.BaseNeuroScriptListener
	program        *Program
	fileMetadata   map[string]string // Points to program.Metadata
	procedures     []*Procedure      // Temporary list of procedures built
	currentProc    *Procedure
	currentSteps   *[]Step
	blockStepStack []*[]Step
	valueStack     []interface{} // Stack for AST nodes (Expression, Step, etc.)
	currentMapKey  *StringLiteralNode
	logger         logging.Logger
	debugAST       bool
	errors         []error // For collecting parse/build errors
	loopDepth      int
}

// newNeuroScriptListener creates a new listener instance.
func newNeuroScriptListener(logger logging.Logger, debugAST bool) *neuroScriptListenerImpl {
	prog := &Program{
		Metadata:   make(map[string]string),
		Procedures: make(map[string]*Procedure),
		Pos:        nil, // Position set later in EnterProgram
	}
	return &neuroScriptListenerImpl{
		BaseNeuroScriptListener: &gen.BaseNeuroScriptListener{},
		program:                 prog,
		fileMetadata:            prog.Metadata, // Directly use program's metadata map
		procedures:              make([]*Procedure, 0, 10),
		blockStepStack:          make([]*[]Step, 0, 5),
		valueStack:              make([]interface{}, 0, 20),
		logger:                  logger,
		debugAST:                debugAST,
		errors:                  make([]error, 0),
		loopDepth:               0,
	}
}

// --- Listener Error Handling ---
func (l *neuroScriptListenerImpl) addError(ctx antlr.ParserRuleContext, format string, args ...interface{}) {
	var startToken antlr.Token
	if ctx != nil {
		startToken = ctx.GetStart()
	}
	pos := tokenToPosition(startToken) // tokenToPosition handles nil token
	errMsg := fmt.Sprintf(format, args...)
	// Create a ParseError type if you have one, or just a standard error.
	// Assuming ParseError is not the standard error type for listener.errors
	err := fmt.Errorf("AST build error at %s: %s", pos.String(), errMsg)
	isDuplicate := false
	for _, existingErr := range l.errors {
		if existingErr.Error() == err.Error() {
			isDuplicate = true
			break
		}
	}
	if !isDuplicate {
		l.errors = append(l.errors, err)
		l.logger.Error("AST Build Error", "position", pos.String(), "message", errMsg)
	} else {
		l.logger.Debug("AST Build Error (Duplicate)", "position", pos.String(), "message", errMsg)
	}
}
func (l *neuroScriptListenerImpl) addErrorf(token antlr.Token, format string, args ...interface{}) {
	pos := tokenToPosition(token) // tokenToPosition handles nil token
	errMsg := fmt.Sprintf(format, args...)
	err := fmt.Errorf("AST build error near %s: %s", pos.String(), errMsg)
	isDuplicate := false
	for _, existingErr := range l.errors {
		if existingErr.Error() == err.Error() {
			isDuplicate = true
			break
		}
	}
	if !isDuplicate {
		l.errors = append(l.errors, err)
		l.logger.Error("AST Build Error", "position", pos.String(), "message", errMsg)
	} else {
		l.logger.Debug("AST Build Error (Duplicate)", "position", pos.String(), "message", errMsg)
	}
}

// --- Listener Getters ---
func (l *neuroScriptListenerImpl) GetFileMetadata() map[string]string {
	if l.program != nil && l.program.Metadata != nil {
		return l.program.Metadata
	}
	// This should not be reached if program and its metadata are initialized in newNeuroScriptListener
	l.logger.Warn("GetFileMetadata called when listener.program.Metadata is nil.")
	return make(map[string]string) // Return empty map to avoid nil issues
}
func (l *neuroScriptListenerImpl) GetResult() []*Procedure { // This seems to be for the Build method's final assembly
	l.logger.Warn("GetResult called on listener; this returns the temporary slice for final assembly, not the final program map.")
	return l.procedures
}

// --- Listener Logging Helper ---
func (l *neuroScriptListenerImpl) logDebugAST(format string, v ...interface{}) {
	if l.debugAST {
		l.logger.Debug(fmt.Sprintf(format, v...))
	}
}

// --- ADDED: Loop Context Helper ---
func (l *neuroScriptListenerImpl) isInsideLoop() bool {
	return l.loopDepth > 0
}

// --- Listener ANTLR Method Implementations ---
func (l *neuroScriptListenerImpl) EnterProgram(ctx *gen.ProgramContext) {
	l.logDebugAST(">>> Enter Program")
	// Initialize ProgramNode or set its token
	if l.program == nil { // Should be initialized by newNeuroScriptListener
		l.program = &Program{
			Metadata:   make(map[string]string),
			Procedures: make(map[string]*Procedure),
		}
		l.logger.Warn("neuroScriptListenerImpl.program was nil in EnterProgram, re-initialized.")
	}
	l.program.Pos = tokenToPosition(ctx.GetStart())
	l.fileMetadata = l.program.Metadata // Ensure fileMetadata points to the program's map

	// Reset fields for potential re-use (though usually a new listener is made per build)
	l.procedures = make([]*Procedure, 0, 10)
	l.errors = make([]error, 0)
	l.valueStack = make([]interface{}, 0, 20)
	l.blockStepStack = make([]*[]Step, 0, 5)
	l.currentProc = nil
	l.currentSteps = nil
	l.currentMapKey = nil
	l.loopDepth = 0
}

func (l *neuroScriptListenerImpl) ExitProgram(ctx *gen.ProgramContext) {
	finalProcCount := 0
	if l.program != nil && l.program.Procedures != nil {
		finalProcCount = len(l.program.Procedures)
	} else if l.program == nil {
		l.logger.Error("ExitProgram: l.program is nil!")
	} else {
		l.logger.Error("ExitProgram: l.program.Procedures is nil!")
	}
	metaCount := 0
	if l.fileMetadata != nil {
		metaCount = len(l.fileMetadata)
	} else {
		l.logger.Error("ExitProgram: l.fileMetadata is nil!")
	}
	l.logDebugAST("<<< Exit Program (Metadata Count: %d, Final Procedure Count: %d, Listener Errors: %d, Final Stack Size: %d)",
		metaCount, finalProcCount, len(l.errors), len(l.valueStack))

	if len(l.valueStack) != 0 {
		errMsg := fmt.Sprintf("internal AST builder error: value stack size is %d at end of program", len(l.valueStack))
		l.logger.Error("ExitProgram: Value stack not empty!", "size", len(l.valueStack), "top_value_type", fmt.Sprintf("%T", l.valueStack[len(l.valueStack)-1]))
		l.errors = append(l.errors, errors.New(errMsg))
	}
	if len(l.blockStepStack) != 0 {
		errMsg := fmt.Sprintf("internal AST builder error: block step stack size is %d at end of program", len(l.blockStepStack))
		l.logger.Error("ExitProgram: Block step stack not empty!", "size", len(l.blockStepStack))
		l.errors = append(l.errors, errors.New(errMsg))
	}
}
