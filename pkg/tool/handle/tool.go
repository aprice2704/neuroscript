// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Implements the tool functions for handle inspection.
// filename: pkg/tool/handle/tools.go
// nlines: 48
// risk_rating: LOW

package handle

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// toolHandleType returns the Kind string of the handle.
func toolHandleType(rt tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Handle.Type: expected 1 argument (handle)", lang.ErrArgumentMismatch)
	}

	// The argument has been unwrapped to interfaces.HandleValue by the binder/validator
	// because we specified ArgTypeHandle in the spec.
	h, ok := args[0].(interfaces.HandleValue)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("Handle.Type: argument must be a handle, got %T", args[0]), lang.ErrArgumentMismatch)
	}

	return h.HandleKind(), nil
}

// toolHandleIsValid checks if the handle points to a live object in the registry.
func toolHandleIsValid(rt tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Handle.IsValid: expected 1 argument (handle)", lang.ErrArgumentMismatch)
	}

	h, ok := args[0].(interfaces.HandleValue)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("Handle.IsValid: argument must be a handle, got %T", args[0]), lang.ErrArgumentMismatch)
	}

	// Check registry existence
	reg := rt.HandleRegistry()
	if reg == nil {
		// If there is no registry, no handles are valid.
		return false, nil
	}

	_, err := reg.GetHandle(h.HandleID())
	return err == nil, nil
}
