// NeuroScript Version: 0.4.0
// File version: 7
// Purpose: Corrected toolListGet to handle float64 and other numeric types for the index argument.
// filename: pkg/tool/list/tools_list_impl.go
// nlines: 237
// risk_rating: LOW

package list

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// --- List Tool Implementations (Primitive-Aware) ---
// These functions are called by the bridge and receive unwrapped, native Go types.
// They return native Go types, which the bridge will then wrap.

func toolListLength(_ *Interpreter, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "len expects a list", ErrArgumentMismatch)
	}
	return float64(len(list)), nil
}

func toolListAppend(_ *Interpreter, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "append expects a list for the first argument", ErrArgumentMismatch)
	}
	element := args[1]
	return append(list, element), nil
}

func toolListPrepend(_ *Interpreter, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "prepend expects a list for the first argument", ErrArgumentMismatch)
	}
	element := args[1]
	return append([]interface{}{element}, list...), nil
}

func toolListGet(_ *Interpreter, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "get expects a list", ErrArgumentMismatch)
	}

	// FIX: Handle multiple numeric types for the index, as the interpreter unwraps to float64.
	var index int
	switch v := args[1].(type) {
	case float64:
		index = int(v)
	case int:
		index = v
	case int64:
		index = int(v)
	default:
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "get expects an integer for index", ErrArgumentMismatch)
	}

	var defaultValue interface{} = nil
	if len(args) > 2 {
		defaultValue = args[2]
	}

	if index < 0 || index >= len(list) {
		return defaultValue, nil
	}
	return list[index], nil
}

func toolListSlice(_ *Interpreter, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "slice expects a list", ErrArgumentMismatch)
	}
	start, ok := args[1].(int64)
	if !ok {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "slice expects an integer for start index", ErrArgumentMismatch)
	}
	end, ok := args[2].(int64)
	if !ok {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "slice expects an integer for end index", ErrArgumentMismatch)
	}

	listLen := len(list)
	s := int(start)
	e := int(end)

	if s < 0 {
		s = 0
	}
	if e > listLen {
		e = listLen
	}
	if s > e || s >= listLen {
		return []interface{}{}, nil
	}

	return list[s:e], nil
}

func toolListContains(_ *Interpreter, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "contains expects a list", ErrArgumentMismatch)
	}
	element := args[1]
	for _, item := range list {
		if reflect.DeepEqual(item, element) {
			return true, nil
		}
	}
	return false, nil
}

func toolListReverse(_ *Interpreter, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "reverse expects a list", ErrArgumentMismatch)
	}
	listLen := len(list)
	newList := make([]interface{}, listLen)
	for i := 0; i < listLen; i++ {
		newList[i] = list[listLen-1-i]
	}
	return newList, nil
}

func toolListSort(_ *Interpreter, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "sort expects a list", ErrArgumentMismatch)
	}
	if len(list) == 0 {
		return []interface{}{}, nil
	}

	/* ---------- stable sort ---------- */

	newList := make([]interface{}, len(list))
	copy(newList, list)

	var sortErr error
	sort.SliceStable(newList, func(i, j int) bool {
		a, b := newList[i], newList[j]

		// both strings → lexicographic
		sa, saOK := a.(string)
		sb, sbOK := b.(string)
		if saOK && sbOK {
			return sa < sb
		}

		// both numeric → numeric order
		na, naOK := toFloat64(a)
		nb, nbOK := toFloat64(b)
		if naOK && nbOK {
			return na < nb
		}

		sortErr = fmt.Errorf("cannot sort mixed types: %T and %T", a, b)
		return false
	})
	if sortErr != nil {
		return nil, lang.NewRuntimeError(ErrorCodeType, sortErr.Error(), ErrListCannotSortMixedTypes)
	}

	/* ---------- decide whether to coerce numbers ---------- */

	allNumeric := true
	for _, v := range newList {
		switch v.(type) {
		case int, int64, float64:
			// ok
		default:
			allNumeric = false
			break
		}
	}

	if allNumeric {
		for i, v := range newList {
			if num, ok := toFloat64(v); ok {
				newList[i] = num
			}
		}
	}

	return newList, nil
}

func toolListHead(_ *Interpreter, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "head expects a list", ErrArgumentMismatch)
	}
	if len(list) == 0 {
		return nil, nil
	}
	return list[0], nil
}

func toolListRest(_ *Interpreter, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "rest expects a list", ErrArgumentMismatch)
	}
	if len(list) <= 1 {
		return []interface{}{}, nil
	}
	return list[1:], nil
}

func toolListTail(_ *Interpreter, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "tail expects a list", ErrArgumentMismatch)
	}
	count, ok := args[1].(int64)
	if !ok {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "tail expects an integer for count", ErrArgumentMismatch)
	}

	listLen := len(list)
	c := int(count)

	if c <= 0 {
		return []interface{}{}, nil
	}
	if c >= listLen {
		return list, nil
	}
	return list[listLen-c:], nil
}

func toolListIsEmpty(_ *Interpreter, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, "isEmpty expects a list", ErrArgumentMismatch)
	}
	return len(list) == 0, nil
}