// NeuroScript Version: 0.3.1
// File version: 3
// Purpose: Reviewed for compliance. Manages raw Go objects via handles, assuming Interpreter.objectCache is map[string]interface{}.
// filename: pkg/core/interpreter_handles.go
// nlines: 60
// risk_rating: LOW

package core

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func (i *Interpreter) RegisterHandle(obj interface{}, typePrefix string) (string, error) {
	if typePrefix == "" {
		return "", fmt.Errorf("%w: handle type prefix cannot be empty", ErrInvalidArgument)
	}
	if strings.Contains(typePrefix, handleSeparator) {
		return "", fmt.Errorf("%w: handle type prefix '%s' cannot contain separator '%s'", ErrInvalidArgument, typePrefix, handleSeparator)
	}
	if i.objectCache == nil {
		i.objectCache = make(map[string]interface{})
	}
	handleIDPart := uuid.NewString()
	fullHandle := fmt.Sprintf("%s%s%s", typePrefix, handleSeparator, handleIDPart)
	i.objectCache[fullHandle] = obj
	return fullHandle, nil
}

func (i *Interpreter) GetHandleValue(handle string, expectedTypePrefix string) (interface{}, error) {
	if expectedTypePrefix == "" {
		return nil, fmt.Errorf("%w: expected handle type prefix cannot be empty", ErrInvalidArgument)
	}
	if handle == "" {
		return nil, fmt.Errorf("%w: handle cannot be empty", ErrInvalidArgument)
	}
	parts := strings.SplitN(handle, handleSeparator, 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("%w: invalid handle format", ErrInvalidArgument)
	}
	actualPrefix := parts[0]

	if actualPrefix != expectedTypePrefix {
		return nil, fmt.Errorf("%w: expected prefix '%s', got '%s'", ErrHandleWrongType, expectedTypePrefix, actualPrefix)
	}

	if i.objectCache == nil {
		return nil, fmt.Errorf("%w: internal error: object cache is not initialized", ErrInternal)
	}
	obj, found := i.objectCache[handle]
	if !found {
		return nil, fmt.Errorf("%w: handle '%s'", ErrHandleNotFound, handle)
	}
	return obj, nil
}

func (i *Interpreter) RemoveHandle(handle string) bool {
	if i.objectCache == nil {
		return false
	}
	_, found := i.objectCache[handle]
	if found {
		delete(i.objectCache, handle)
	}
	return found
}
