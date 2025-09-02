// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Contains helper methods for the Interpreter. Corrected LoadAndRun signature and removed use of deleted 'accounts' field.
// filename: pkg/interpreter/interpreter_helpers.go
// nlines: 88
// risk_rating: MEDIUM

package interpreter

import (
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// defaultWhisperFunc is the built-in whisper implementation.
func (i *Interpreter) defaultWhisperFunc(handle, data lang.Value) {
	i.bufferManager.Write(handle.String(), data.String()+"\n")
}

// clone creates a new interpreter instance for sandboxing.
func (i *Interpreter) clone() *Interpreter {
	clone := NewInterpreter(
		WithLogger(i.logger),
		WithStdout(i.stdout),
		WithStdin(i.stdin),
		WithStderr(i.stderr),
		WithSandboxDir(i.state.sandboxDir),
	)
	clone.tools = i.tools
	clone.ExecPolicy = i.ExecPolicy
	clone.modelStore = i.modelStore
	// clone.accounts = i.accounts // This field was removed
	clone.turnCtx = i.turnCtx
	clone.aiTranscript = i.aiTranscript
	clone.transientPrivateKey = i.transientPrivateKey

	clone.customEmitFunc = i.customEmitFunc
	clone.customWhisperFunc = i.customWhisperFunc

	rootInterpreter := i
	if i.root != nil {
		rootInterpreter = i.root
	}
	clone.root = rootInterpreter

	clone.state.knownProcedures = i.state.knownProcedures

	rootInterpreter.state.variablesMu.RLock()
	defer rootInterpreter.state.variablesMu.RUnlock()

	for name := range rootInterpreter.state.globalVarNames {
		if val, ok := rootInterpreter.state.variables[name]; ok {
			clone.SetVariable(name, val)
			clone.state.globalVarNames[name] = true
		}
	}

	return clone
}

// AddProcedure programmatically adds a single procedure to the interpreter's registry.
func (i *Interpreter) AddProcedure(proc ast.Procedure) error {
	if i.state.knownProcedures == nil {
		i.state.knownProcedures = make(map[string]*ast.Procedure)
	}
	if proc.Name() == "" {
		return errors.New("cannot add procedure with empty name")
	}
	if _, exists := i.state.knownProcedures[proc.Name()]; exists {
		return fmt.Errorf("%w: '%s'", lang.ErrProcedureExists, proc.Name())
	}
	i.state.knownProcedures[proc.Name()] = &proc
	return nil
}

// GetAllVariables returns a copy of all variables in the current scope.
func (i *Interpreter) GetAllVariables() (map[string]lang.Value, error) {
	i.state.variablesMu.RLock()
	defer i.state.variablesMu.RUnlock()
	clone := make(map[string]lang.Value)
	for k, v := range i.state.variables {
		clone[k] = v
	}
	return clone, nil
}

// LoadAndRun is a convenience method to load a program and run its main procedure.
func (i *Interpreter) LoadAndRun(program *ast.Program, mainProcName string, args ...lang.Value) (lang.Value, error) {
	tree := &interfaces.Tree{Root: program}
	if err := i.Load(tree); err != nil {
		return nil, fmt.Errorf("failed to load program: %w", err)
	}
	return i.Run(mainProcName, args...)
}
