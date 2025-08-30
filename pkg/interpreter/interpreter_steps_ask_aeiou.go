// NeuroScript Version: 0.7.0
// File version: 3.0.0
// Purpose: Updated to capture 'whisper' commands in addition to 'emits' for ask loop state.
// filename: pkg/interpreter/interpreter_steps_ask_aeiou.go
// nlines: 60
// risk_rating: MEDIUM

package interpreter

import (
	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
)

// executeAeiouTurn executes the 'ACTIONS' section of a parsed AEIOU envelope
// within a cloned interpreter instance to isolate its state. It captures any
// 'emit' and 'whisper' statements for loop control and state passing.
func executeAeiouTurn(i *Interpreter, env *aeiou.Envelope, actionEmits *[]string, actionWhispers *map[string]lang.Value) error {
	if env.Actions == "" {
		return nil // Nothing to execute
	}

	// We need to capture emits from this execution.
	i.SetEmitFunc(func(e lang.Value) {
		s, _ := lang.ToString(e)
		*actionEmits = append(*actionEmits, s)
	})

	// We also need to capture whispers to populate the scratchpad for the next turn.
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

	// Execute the parsed command from the ACTIONS block.
	// Since executeWhisper is a placeholder, this will log but not populate actionWhispers yet.
	// When whisper is implemented, this capture mechanism will work without further changes here.
	_, err := i.Execute(program)
	if err != nil {
		return err // Propagate runtime errors from the executed code
	}

	return nil
}
