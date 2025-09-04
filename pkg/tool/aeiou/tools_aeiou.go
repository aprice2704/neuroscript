// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Implements the Go function for the 'tool.aeiou.ComposeEnvelope' tool.
// filename: pkg/tool/aeiou/tools_aeiou.go
// nlines: 36
// risk_rating: LOW

package aeiou

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func envelopeToolFunc(rt tool.Runtime, args []any) (any, error) {
	userdata, _ := args[0].(string)
	actions, _ := args[1].(string)
	scratchpad, _ := args[2].(string)
	output, _ := args[3].(string)

	if actions == "" {
		// Ensure a minimal valid command block if none is provided.
		actions = "command\nendcommand"
	}

	env := &aeiou.Envelope{
		UserData:   userdata,
		Actions:    actions,
		Scratchpad: scratchpad,
		Output:     output,
	}

	composed, err := env.Compose()
	if err != nil {
		return nil, fmt.Errorf("failed to compose envelope: %w", err)
	}
	return composed, nil
}
