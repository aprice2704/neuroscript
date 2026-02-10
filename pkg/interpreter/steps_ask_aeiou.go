// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 17
// :: description: Fixed Dirty Buffer anti-pattern by adding mutex protection to output capturing.
// :: latestChange: Wrapped EmitFunc and WhisperFunc in a mutex to prevent races on slice/map.
// :: filename: pkg/interpreter/steps_ask_aeiou.go
// :: serialization: go

package interpreter

import (
	"sync" // Added sync

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

	var mu sync.Mutex // Protective mutex for output capturing

	// This function operates directly on 'i', which is an ephemeral sandbox interpreter.
	// We temporarily modify its HostContext to intercept I/O.
	originalHostContext := i.hostContext
	turnHostContext := *i.hostContext // Create a shallow copy
	turnHostContext.EmitFunc = func(e lang.Value) {
		s, _ := lang.ToString(e)
		mu.Lock()
		defer mu.Unlock()
		*actionEmits = append(*actionEmits, s)
	}
	turnHostContext.WhisperFunc = func(handle, data lang.Value) {
		handleStr, _ := lang.ToString(handle)
		if handleStr != "" {
			mu.Lock()
			defer mu.Unlock()
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

	// Execute the parsed command(s).
	_, err := i.Execute(program)
	if err != nil {
		return err // Propagate other runtime errors as they are
	}

	return nil
}
