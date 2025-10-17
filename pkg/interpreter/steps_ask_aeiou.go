// NeuroScript Version: 0.8.0
// File version: 16
// Purpose: Correctly propagate all runtime errors as-is, removing special (and incorrect) wrapping for policy errors.
// filename: pkg/interpreter/steps_ask_aeiou.go
// nlines: 75
// risk_rating: MEDIUM

package interpreter

import (
	"fmt"
	"os"

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

	// DEBUG: Log the context on the interpreter *before* execution.
	fmt.Fprintf(os.Stderr, "[DEBUG] executeAeiouTurn: Entered with interpreter %s\n", i.id)
	// --- FIX: 'i' is a concrete type, not an interface. Call method directly. ---
	ctx := i.GetTurnContext()
	if ctx != nil {
		fmt.Fprintf(os.Stderr, "[DEBUG] executeAeiouTurn: Context on entry is %p\n", ctx)
		if sid, ok := ctx.Value(AeiouSessionIDKey).(string); ok {
			fmt.Fprintf(os.Stderr, "[DEBUG] executeAeiouTurn: Found SID '%s' on entry.\n", sid)
		} else {
			fmt.Fprintf(os.Stderr, "[DEBUG] executeAeiouTurn: WARNING! SID not found on entry.\n")
		}
	} else {
		fmt.Fprintf(os.Stderr, "[DEBUG] executeAeiouTurn: WARNING! GetTurnContext() returned nil.\n")
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
	fmt.Fprintf(os.Stderr, "[DEBUG] executeAeiouTurn: Calling i.Execute() on interpreter %s\n", i.id) // DEBUG
	_, err := i.Execute(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[DEBUG] executeAeiouTurn: i.Execute() FAILED: %v\n", err) // DEBUG
		// THE FIX: The 'tool.aeiou.magic' tool no longer exists, so we don't
		// need to check for a policy error related to it. We just propagate
		// all other errors as-is.

		// This block was wrapping policy errors incorrectly, causing them to
		// be re-wrapped as InternalErrors upstream.
		/*
			var rtErr *lang.RuntimeError
			if errors.As(err, &rtErr) && rtErr.Code == lang.ErrorCodePolicy {
				return fmt.Errorf("%w: %s", lang.ErrToolDenied, rtErr.Message)
			}
		*/
		return err // Propagate other runtime errors as they are
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] executeAeiouTurn: i.Execute() succeeded.\n") // DEBUG

	return nil
}
