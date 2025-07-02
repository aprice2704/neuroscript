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
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// --- List Tool Implementations (Primitive-Aware) ---
// These functions are called by the bridge and receive unwrapped, native Go types.
// They return native Go types, which the bridge will then wrap.

func toolListLength(_ tool.RunTime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "len expects a list", lang.ErrArgumentMismatch)
	}
	return float64(len(list)), nil
}

func toolListAppend(_ tool.RunTime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "append expects a list for the first argument", lang.ErrArgumentMismatch)
	}
	element := args[1]
	return append(list, element), nil
}

func toolListPrepend(_ tool.RunTime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "prepend expects a list for the first argument", lang.ErrArgumentMismatch)
	}
	element := args[1]
	return append([]interface{}{element}, list...), nil
}

func toolListGet(_ tool.RunTime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "get expects a list", lang.ErrArgumentMismatch)
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
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "get expects an integer for index", lang.ErrArgumentMismatch)
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

func toolListSlice(_ tool.RunTime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "slice expects a list", lang.ErrArgumentMismatch)
	}
	start, ok := args[1].(int64)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "slice expects an integer for start index", lang.ErrArgumentMismatch)
	}
	end, ok := args[2].(int64)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "slice expects an integer for end index", lang.ErrArgumentMismatch)
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

func toolListContains(_ tool.RunTime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "contains expects a list", lang.ErrArgumentMismatch)
	}
	element := args[1]
	for _, item := range list {
		if reflect.DeepEqual(item, element) {
			return true, nil
		}
	}
	return false, nil
}

func toolListReverse(_ tool.RunTime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "reverse expects a list", lang.ErrArgumentMismatch)
	}
	listLen := len(list)
	newList := make([]interface{}, listLen)
	for i := 0; i < listLen; i++ {
		newList[i] = list[listLen-1-i]
	}
	return newList, nil
}

func toolListSort(_ tool.RunTime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "sort expects a list", lang.ErrArgumentMismatch)
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
		na, naOK := lang.ToFloat64(a)
		nb, nbOK := lang.ToFloat64(b)
		if naOK && nbOK {
			return na < nb
		}

		sortErr = fmt.Errorf("cannot sort mixed types: %T and %T", a, b)
		return false
	})
	if sortErr != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, sortErr.Error(), lang.ErrListCannotSortMixedTypes)
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
			if num, ok := lang.ToFloat64(v); ok {
				newList[i] = num
			}
		}
	}

	return newList, nil
}

func toolListHead(_ tool.RunTime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "head expects a list", lang.ErrArgumentMismatch)
	}
	if len(list) == 0 {
		return nil, nil
	}
	return list[0], nil
}

func toolListRest(_ tool.RunTime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "rest expects a list", lang.ErrArgumentMismatch)
	}
	if len(list) <= 1 {
		return []interface{}{}, nil
	}
	return list[1:], nil
}

func toolListTail(_ tool.RunTime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "tail expects a list", lang.ErrArgumentMismatch)
	}
	count, ok := args[1].(int64)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "tail expects an integer for count", lang.ErrArgumentMismatch)
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

func toolListIsEmpty(_ tool.RunTime, args []interface{}) (interface{}, error) {
	list, ok := args[0].([]interface{})
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "isEmpty expects a list", lang.ErrArgumentMismatch)
	}
	return len(list) == 0, nil
}
