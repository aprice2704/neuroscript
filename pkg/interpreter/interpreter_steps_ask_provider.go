// NeuroScript Version: 0.7.0
// File version: 4
// Purpose: FIX: Passes the correct, stateful turn context to the AI provider instead of context.Background().
// filename: pkg/interpreter/interpreter_steps_ask_provider.go
// nlines: 66
// risk_rating: MEDIUM

package interpreter

import (
	"fmt"
	"os"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/provider"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func callAIProvider(i *Interpreter, model types.AgentModel, withOpts *lang.MapValue, prompt string, pos *types.Position) (*provider.AIResponse, error) {
	apiKey := ""
	if model.AccountName != "" {
		apiKey = os.Getenv(model.AccountName)
	}

	prov, provExists := i.GetProvider(model.Provider)
	if !provExists {
		return nil, lang.NewRuntimeError(lang.ErrorCodeProviderNotFound, fmt.Sprintf("provider '%s' for AgentModel '%s' not found", model.Provider, model.Name), nil).WithPosition(pos)
	}

	req := provider.AIRequest{
		ModelName: model.Model,
		Prompt:    prompt,
		APIKey:    apiKey,
	}

	// Apply 'with' options, overriding defaults from the AgentModel
	if tempVal, ok := withOpts.Value["temperature"]; ok {
		if tempFloat, isFloat := lang.ToFloat64(tempVal); isFloat {
			req.Temperature = tempFloat
		}
	}

	// Add other 'with' options here as needed...

	// FIX: Pass the interpreter's current turn context to the provider.
	resp, err := prov.Chat(i.GetTurnContext(), req)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeExternal, "AI provider call failed", err).WithPosition(pos)
	}
	return resp, nil
}
