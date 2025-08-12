// NeuroScript Version: 0.6.0
// File version: 2.1.0
// Purpose: Corrected lang.Wrap calls to handle two return values.
// filename: pkg/tool/agentmodel/tools_agentmodel.go
// nlines: 100
// risk_rating: HIGH

package agentmodel

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// getInterpreter is a helper to safely cast the runtime to our interpreter instance.
func getInterpreter(rt tool.Runtime) (*interpreter.Interpreter, error) {
	interp, ok := rt.(*interpreter.Interpreter)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeConfiguration, "agentmodel tools require a direct interpreter instance", lang.ErrConfiguration)
	}
	return interp, nil
}

func toolRegisterAgentModel(rt tool.Runtime, args []interface{}) (interface{}, error) {
	interp, err := getInterpreter(rt)
	if err != nil {
		return false, err
	}
	name, _ := args[0].(string)
	configMap, ok := args[1].(map[string]interface{})
	if !ok {
		return false, lang.NewRuntimeError(lang.ErrorCodeType, "config argument must be a map", lang.ErrInvalidArgument)
	}
	langValueMap := make(map[string]lang.Value)
	for k, v := range configMap {
		langValueMap[k], _ = lang.Wrap(v) // Corrected assignment
	}
	err = interp.RegisterAgentModel(types.AgentModelName(name), langValueMap)
	return err == nil, err
}

func toolUpdateAgentModel(rt tool.Runtime, args []interface{}) (interface{}, error) {
	interp, err := getInterpreter(rt)
	if err != nil {
		return false, err
	}
	name, _ := args[0].(string)
	updatesMap, ok := args[1].(map[string]interface{})
	if !ok {
		return false, lang.NewRuntimeError(lang.ErrorCodeType, "updates argument must be a map", lang.ErrInvalidArgument)
	}
	langValueMap := make(map[string]lang.Value)
	for k, v := range updatesMap {
		langValueMap[k], _ = lang.Wrap(v) // Corrected assignment
	}
	err = interp.UpdateAgentModel(types.AgentModelName(name), langValueMap)
	return err == nil, err
}

func toolDeleteAgentModel(rt tool.Runtime, args []interface{}) (interface{}, error) {
	interp, err := getInterpreter(rt)
	if err != nil {
		return false, err
	}
	name, _ := args[0].(string)
	if !interp.DeleteAgentModel(types.AgentModelName(name)) {
		return false, lang.NewRuntimeError(lang.ErrorCodeToolExecutionFailed, fmt.Sprintf("failed to delete agent model '%s'", name), lang.ErrNotFound)
	}
	return true, nil
}

func toolListAgentModels(rt tool.Runtime, args []interface{}) (interface{}, error) {
	interp, err := getInterpreter(rt)
	if err != nil {
		return nil, err
	}
	return interp.ListAgentModels(), nil
}

func toolSelectAgentModel(rt tool.Runtime, args []interface{}) (interface{}, error) {
	interp, err := getInterpreter(rt)
	if err != nil {
		return "", err
	}
	models := interp.ListAgentModels()
	if len(models) == 0 {
		return "", lang.NewRuntimeError(lang.ErrorCodeToolNotFound, "no agent models are registered", lang.ErrNotFound)
	}
	return models[0], nil
}
