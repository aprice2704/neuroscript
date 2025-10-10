// NeuroScript Version: 0.7.0
// File version: 6
// Purpose: FIX: Uses the GetTurnContext() method instead of the removed turnCtx field to access the turn's context.
// filename: pkg/interpreter/interpreter_steps_ask_aeiou.go
// nlines: 71
// risk_rating: HIGH

package interpreter

import (
	"context"
	"fmt"
	"os"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

// executeAeiouTurn executes the 'ACTIONS' section of a parsed AEIOU envelope
// within a cloned interpreter instance to isolate its state. It captures all
// 'emit' and 'whisper' statements for the host loop to process.
func executeAeiouTurn(i *Interpreter, env *aeiou.Envelope, actionEmits *[]string, actionWhispers *map[string]lang.Value) error {
	// --- NEW DEBUGGING ---
	turnCtx := i.GetTurnContext()
	if turnCtx != nil && turnCtx != context.Background() {
		sid, _ := turnCtx.Value(aeiou.SessionIDKey).(string)
		turn, _ := turnCtx.Value(aeiou.TurnIndexKey).(int)
		fmt.Fprintf(os.Stderr, "[DEBUG executeAeiouTurn] START. Interp ID: %s, SID: %q, Turn: %d\n", i.id, sid, turn)
	} else {
		fmt.Fprintf(os.Stderr, "[DEBUG executeAeiouTurn] START. Interp ID: %s, Context is NIL or empty\n", i.id)
	}
	// --- END NEW DEBUGGING ---

	if env.Actions == "" {
		return nil // Nothing to execute
	}

	// Capture all emits from this execution for the OUTPUT section of the next turn.
	i.SetEmitFunc(func(e lang.Value) {
		s, _ := lang.ToString(e)
		*actionEmits = append(*actionEmits, s)
	})

	// Capture all whispers for the SCRATCHPAD section of the next turn.
	i.SetWhisperFunc(func(handle, data lang.Value) {
		handleStr, _ := lang.ToString(handle)
		if handleStr != "" {
			(*actionWhispers)[handleStr] = data
		}
	})

	parserAPI := parser.NewParserAPI(i.GetLogger())
	p, pErr := parserAPI.Parse(env.Actions)
	if pErr != nil {
		// If parsing the agent's response fails, it's a runtime error,
		// as it indicates a malformed response from the AI.
		return lang.NewRuntimeError(lang.ErrorCodeSyntax, "failed to parse ACTIONS block from AI response", pErr)
	}

	program, _, bErr := parser.NewASTBuilder(i.GetLogger()).Build(p)
	if bErr != nil {
		return lang.NewRuntimeError(lang.ErrorCodeSyntax, "failed to build AST for ACTIONS block", bErr)
	}

	// Execute the parsed command(s) from the ACTIONS block. The custom emit
	// and whisper functions registered above will capture all relevant output.
	_, err := i.Execute(program)
	if err != nil {
		return err // Propagate runtime errors from the executed code
	}

	return nil
}
