// NeuroScript Version: 0.7.4
// File version: 2
// Purpose: Implements the ax.Registry interface. FIX: Commented out unimplemented methods.
// filename: pkg/api/ax_bridge.go
// nlines: 26
// risk_rating: LOW

package api

import (
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/ax"
)

var _ ax.Registry = (*Interpreter)(nil)

// RegisterBuiltin provides a way for extensions to register built-in functions.
// TODO: Uncomment when this is implemented on the internal interpreter.
func (i *Interpreter) RegisterBuiltin(name string, fn any) error {
	return errors.New("RegisterBuiltin is not yet implemented")
	// return i.internal.RegisterBuiltin(name, fn)
}

// RegisterType provides a way for extensions to register custom types.
// TODO: Uncomment when this is implemented on the internal interpreter.
func (i *Interpreter) RegisterType(name string, factory any) error {
	return errors.New("RegisterType is not yet implemented")
	// return i.internal.RegisterType(name, factory)
}

func (i *Interpreter) Use(exts ...ax.Extension) error {
	for _, e := range exts {
		if err := e.Register(i); err != nil {
			return fmt.Errorf("ax extension %q: %w", e.Name(), err)
		}
	}
	return nil
}
