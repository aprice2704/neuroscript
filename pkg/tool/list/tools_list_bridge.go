// filename: pkg/tool/list/tools_list_bridge.go
package list

import (
	"fmt"
)

// -----------------------------------------------------------------------------
// Bridge function between the interpreter (wrappers) and the List-Reverse tool.
// This is the ONLY place that unwraps ↓ and re-wraps ↑.
// -----------------------------------------------------------------------------

// CallListReverse is registered with the interpreter’s tool registry.
// Signature: wrappers in  -> wrapper out.
func CallListReverse(args []lang.lang.Value) (lang.lang.Value, error) {
	// 1. Unwrap interpreter args (they arrive as wrappers).
	rawArgs, err := lang.UnwrapSlice(args)
	if err != nil {
		return nil, err
	}

	// 2. Run validation on **primitives**.
	if err := validateListReverse(rawArgs); err != nil {
		return nil, err
	}

	// 3. Call the actual tool implementation (also primitives).
	//
	//    Expecting exactly one argument: the list.
	//    Adjust indexing if your tool signature differs.
	list, ok := rawArgs[0].([]any)
	if !ok {
		return nil, fmt.Errorf("list.reverse: expected []any, got %T", rawArgs[0])
	}
	out, err := listReverseImpl(list)
	if err != nil {
		return nil, err
	}

	// 4. Wrap result back for the interpreter.
	return lang.Wrap(out)
}

// -----------------------------------------------------------------------------
// Below are *place-holders* so the file compiles.  Replace with your real code.
// -----------------------------------------------------------------------------

// validateListReverse performs business-rule checks on primitives.
func validateListReverse(args []any) error {
	if len(args) != 1 {
		return fmt.Errorf("list.reverse: want 1 arg, got %d", len(args))
	}
	_, ok := args[0].([]any)
	if !ok {
		return fmt.Errorf("list.reverse: arg 0 must be list")
	}
	return nil
}

// listReverseImpl is the real tool logic, completely wrapper-free.
func listReverseImpl(in []any) ([]any, error) {
	n := len(in)
	out := make([]any, n)
	for i, v := range in {
		out[n-1-i] = v
	}
	return out, nil
}