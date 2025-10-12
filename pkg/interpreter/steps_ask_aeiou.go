// NeuroScript Version: 0.8.0
// File version: 5.0.0
// Purpose: Creates a temporary HostContext to correctly capture emit/whisper outputs from an AEIOU turn.
// filename: pkg/interpreter/steps_ask_aeiou.go
// nlines: 60
// risk_rating: HIGH

package interpreter

import (
	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

// executeAeiouTurn executes the 'ACTIONS' section of an AEIOU envelope.
// It uses a temporary, overridden HostContext to capture all 'emit' and 'whisper'
// statements for the host loop to process.
func executeAeiouTurn(i *Interpreter, env *aeiou.Envelope, actionEmits *[]string, actionWhispers *map[string]lang.Value) error {
	if env.Actions == "" {
		return nil // Nothing to execute
	}

	// Fork the interpreter to create an isolated execution environment.
	execInterp := i.fork()

	// Create a new, temporary HostContext for this specific turn.
	// This allows us to intercept I/O without affecting the parent or other interpreters.
	turnHostContext := *i.hostContext // Create a shallow copy
	turnHostContext.EmitFunc = func(e lang.Value) {
		s, _ := lang.ToString(e)
		*actionEmits = append(*actionEmits, s)
	}
	turnHostContext.WhisperFunc = func(handle, data lang.Value) {
		handleStr, _ := lang.ToString(handle)
		if handleStr != "" {
			(*actionWhispers)[handleStr] = data
		}
	}
	// Apply the temporary context to our forked interpreter.
	execInterp.hostContext = &turnHostContext

	parserAPI := parser.NewParserAPI(execInterp.Logger())
	p, pErr := parserAPI.Parse(env.Actions)
	if pErr != nil {
		return lang.NewRuntimeError(lang.ErrorCodeSyntax, "failed to parse ACTIONS block from AI response", pErr)
	}

	program, _, bErr := parser.NewASTBuilder(execInterp.Logger()).Build(p)
	if bErr != nil {
		return lang.NewRuntimeError(lang.ErrorCodeSyntax, "failed to build AST for ACTIONS block", bErr)
	}

	// Execute the parsed command(s). The custom emit and whisper functions
	// in turnHostContext will capture all relevant output.
	_, err := execInterp.Execute(program)
	if err != nil {
		return err // Propagate runtime errors from the executed code
	}

	return nil
}
