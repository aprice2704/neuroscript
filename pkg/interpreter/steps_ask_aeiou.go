// NeuroScript Version: 0.8.0
// File version: 11
// Purpose: Use lang.ErrToolDenied to break circular dependency with api package.
// filename: pkg/interpreter/steps_ask_aeiou.go
// nlines: 62
// risk_rating: HIGH

package interpreter

import (
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// executeAeiouTurn executes the 'ACTIONS' section of an AEIOU envelope.
// It uses a temporary, overridden HostContext to capture all 'emit' and 'whisper'
// statements for the host loop to process. It now captures emits as strings.
func executeAeiouTurn(i *Interpreter, env *aeiou.Envelope, actionEmits *[]string, actionWhispers *map[string]lang.Value) error {
	if env.Actions == "" {
		return nil // Nothing to execute
	}

	// This function operates directly on 'i', which is an ephemeral sandbox interpreter.
	// We temporarily modify its HostContext to intercept I/O.
	originalHostContext := i.hostContext
	turnHostContext := *i.hostContext // Create a shallow copy
	turnHostContext.EmitFunc = func(e lang.Value) {
		// THE FIX: Convert the emitted value to a string before capturing.
		s, _ := lang.ToString(e)
		*actionEmits = append(*actionEmits, s)
	}
	turnHostContext.WhisperFunc = func(handle, data lang.Value) {
		handleStr, _ := lang.ToString(handle)
		if handleStr != "" {
			(*actionWhispers)[handleStr] = data
		}
	}
	i.hostContext = &turnHostContext
	defer func() { i.hostContext = originalHostContext }()

	p, pErr := i.parser.Parse(env.Actions)
	if pErr != nil {
		return lang.NewRuntimeError(lang.ErrorCodeSyntax, "failed to parse ACTIONS block from AI response", pErr)
	}

	program, _, bErr := i.astBuilder.Build(p)
	if bErr != nil {
		return lang.NewRuntimeError(lang.ErrorCodeSyntax, "failed to build AST for ACTIONS block", bErr)
	}

	// Execute the parsed command(s). The custom emit and whisper functions
	// in turnHostContext will capture all relevant output.
	_, err := i.Execute(program)
	if err != nil {
		// Check for an internal policy error and wrap it in the public API's sentinel error.
		var rtErr *lang.RuntimeError
		if errors.As(err, &rtErr) && rtErr.Code == lang.ErrorCodePolicy {
			// Wrap the detailed internal error with the stable, public error type from lang.
			return fmt.Errorf("%w: %s", lang.ErrToolDenied, rtErr.Message)
		}
		return err // Propagate other runtime errors as they are
	}

	return nil
}
