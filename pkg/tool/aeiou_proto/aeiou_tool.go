// NeuroScript Version: 0.7.0
// File version: 7
// Purpose: Implements the 'magic' tool function and restores diagnostic debug logging.
// filename: pkg/tool/aeiou_proto/aeiou_tool.go
// nlines: 135
// risk_rating: MEDIUM

package aeiou_proto

import (
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/aeiou"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

const handlePrefix = "aeiou_env"

func getEnvelopeFromHandle(interpreter tool.Runtime, handle string) (*aeiou.Envelope, error) {
	fmt.Printf("[DEBUG] getEnvelopeFromHandle: received handle %q\n", handle) // DEBUG
	val, err := interpreter.GetHandleValue(handle, handlePrefix)
	if err != nil {
		fmt.Printf("[DEBUG] getEnvelopeFromHandle: GetHandleValue returned error: %v\n", err) // DEBUG
		return nil, err
	}
	env, ok := val.(*aeiou.Envelope)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("internal error: handle '%s' has wrong type %T", handle, val), lang.ErrHandleWrongType)
	}
	return env, nil
}

func toolAeiouNew(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	env := &aeiou.Envelope{}
	handle, err := interpreter.RegisterHandle(env, handlePrefix)
	if err != nil {
		return nil, err
	}
	fmt.Printf("[DEBUG] toolAeiouNew: created handle %q\n", handle) // DEBUG
	return handle, nil
}

func toolAeiouParse(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.ErrArgumentMismatch
	}
	payload, ok := args[0].(string)
	if !ok {
		return nil, lang.ErrArgumentMismatch
	}

	env, err := aeiou.RobustParse(payload)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeSyntax, "failed to parse envelope", err)
	}

	handle, err := interpreter.RegisterHandle(env, handlePrefix)
	if err != nil {
		return nil, err
	}
	return handle, nil
}

func toolAeiouGetSection(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, lang.ErrArgumentMismatch
	}
	handle, okH := args[0].(string)
	sectionName, okS := args[1].(string)
	if !okH || !okS {
		return nil, lang.ErrArgumentMismatch
	}

	env, err := getEnvelopeFromHandle(interpreter, handle)
	if err != nil {
		return nil, err
	}

	switch aeiou.SectionType(strings.ToUpper(sectionName)) {
	case aeiou.SectionActions:
		return env.Actions, nil
	case aeiou.SectionEvents:
		return env.Events, nil
	case aeiou.SectionImplementations:
		return env.Implementations, nil
	case aeiou.SectionOrchestration:
		return env.Orchestration, nil
	case aeiou.SectionUserData:
		return env.UserData, nil
	default:
		return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, fmt.Sprintf("unknown section: %s", sectionName), lang.ErrMapKeyNotFound)
	}
}

func toolAeiouSetSection(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 3 {
		return nil, lang.ErrArgumentMismatch
	}
	handle, okH := args[0].(string)
	sectionName, okS := args[1].(string)
	content, okC := args[2].(string)
	if !okH || !okS || !okC {
		return nil, lang.ErrArgumentMismatch
	}

	env, err := getEnvelopeFromHandle(interpreter, handle)
	if err != nil {
		return nil, err
	}

	switch aeiou.SectionType(strings.ToUpper(sectionName)) {
	case aeiou.SectionActions:
		env.Actions = content
	case aeiou.SectionEvents:
		env.Events = content
	case aeiou.SectionImplementations:
		env.Implementations = content
	case aeiou.SectionOrchestration:
		env.Orchestration = content
	case aeiou.SectionUserData:
		env.UserData = content
	default:
		return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, fmt.Sprintf("unknown section: %s", sectionName), lang.ErrMapKeyNotFound)
	}
	return nil, nil
}

func toolAeiouCompose(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.ErrArgumentMismatch
	}
	handle, ok := args[0].(string)
	if !ok {
		return nil, lang.ErrArgumentMismatch
	}

	env, err := getEnvelopeFromHandle(interpreter, handle)
	if err != nil {
		return nil, err
	}
	return env.Compose()
}

func toolAeiouValidate(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.ErrArgumentMismatch
	}
	handle, ok := args[0].(string)
	if !ok {
		return nil, lang.ErrArgumentMismatch
	}

	env, err := getEnvelopeFromHandle(interpreter, handle)
	if err != nil {
		return nil, err
	}

	errs := env.Validate()
	errorStrings := make([]string, len(errs))
	for i, e := range errs {
		errorStrings[i] = e.Error()
	}
	return errorStrings, nil
}

func toolAeiouMagic(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, lang.ErrArgumentMismatch
	}
	sectionTypeStr, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, "first argument 'type' must be a string", lang.ErrArgumentMismatch)
	}

	var payload interface{}
	if len(args) == 2 {
		payload = args[1]
	}

	magicString, err := aeiou.Wrap(aeiou.SectionType(strings.ToUpper(sectionTypeStr)), payload)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeGeneric, "failed to create magic string", err)
	}
	return magicString, nil
}
