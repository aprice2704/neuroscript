// NeuroScript Version: 0.3.1
// File version: 0.1.3
// Purpose: Populated Category, Example, ReturnHelp, and ErrorConditions for all list tool specs.
// filename: pkg/tool/list/tooldefs_list.go
// nlines: 148
// risk_rating: MEDIUM

package list

import "github.com/aprice2704/neuroscript/pkg/tool"

// listToolsToRegister contains ToolImplementation definitions for list tools.
var listToolsToRegister = []tool.ToolImplementation{
	{Spec: tool.ToolSpec{
		Name:            "List.Length",
		Description:     "Returns the number of elements in a list.",
		Category:        "List Operations",
		Args:            []tool.ArgSpec{{Name: "list", Type: tool.ArgTypeSliceAny, Required: true, Description: "The list to measure."}},
		ReturnType:      tool.ArgTypeInt,
		ReturnHelp:      "Returns an integer representing the number of elements in the provided list.",
		Example:         `tool.List.Length([1, 2, 3]) // returns 3`,
		ErrorConditions: "None expected, as input validation ensures 'list' is a slice. An empty list returns 0.",
	}, Func: toolListLength},
	{Spec: tool.ToolSpec{
		Name:        "List.Append",
		Description: "Returns a *new* list with the given element added to the end.",
		Category:    "List Operations",
		Args: []tool.ArgSpec{
			{Name: "list", Type: tool.ArgTypeSliceAny, Required: true, Description: "The list to append to."},
			{Name: "element", Type: tool.ArgTypeAny, Required: false, Description: "The element to append (can be nil)."},
		},
		ReturnType:      tool.ArgTypeSliceAny,
		ReturnHelp:      "Returns a new list with the 'element' added to the end of the input 'list'. The original list is not modified.",
		Example:         `tool.List.Append([1, 2], 3) // returns [1, 2, 3]`,
		ErrorConditions: "None expected, as input validation ensures 'list' is a slice. Appending 'nil' is allowed.",
	}, Func: toolListAppend},
	{Spec: tool.ToolSpec{
		Name:        "List.Prepend",
		Description: "Returns a *new* list with the given element added to the beginning.",
		Category:    "List Operations",
		Args: []tool.ArgSpec{
			{Name: "list", Type: tool.ArgTypeSliceAny, Required: true, Description: "The list to prepend to."},
			{Name: "element", Type: tool.ArgTypeAny, Required: false, Description: "The element to prepend (can be nil)."},
		},
		ReturnType:      tool.ArgTypeSliceAny,
		ReturnHelp:      "Returns a new list with the 'element' added to the beginning of the input 'list'. The original list is not modified.",
		Example:         `tool.List.Prepend([2, 3], 1) // returns [1, 2, 3]`,
		ErrorConditions: "None expected, as input validation ensures 'list' is a slice. Prepending 'nil' is allowed.",
	}, Func: toolListPrepend},
	{Spec: tool.ToolSpec{
		Name:        "List.Get",
		Description: "Safely gets the element at a specific index (0-based). Returns nil or the optional default value if the index is out of bounds.",
		Category:    "List Operations",
		Args: []tool.ArgSpec{
			{Name: "list", Type: tool.ArgTypeSliceAny, Required: true, Description: "The list to get from."},
			{Name: "index", Type: tool.ArgTypeInt, Required: true, Description: "The 0-based index."},
			{Name: "default", Type: tool.ArgTypeAny, Required: false, Description: "Optional default value if index is out of bounds."},
		},
		ReturnType:      tool.ArgTypeAny,
		ReturnHelp:      "Returns the element at the specified 0-based 'index' in the 'list'. If the index is out of bounds, it returns the provided 'default' value. If no 'default' is provided and the index is out of bounds, it returns nil.",
		Example:         `tool.List.Get(["a", "b", "c"], 1) // returns "b"\n tool.List.Get(["a"], 5, "default_val") // returns "default_val"`,
		ErrorConditions: "Returns nil or the default value if the index is out of bounds. No specific error type is returned for out-of-bounds access by design.",
	}, Func: toolListGet},
	{Spec: tool.ToolSpec{
		Name:        "List.Slice",
		Description: "Returns a *new* list containing elements from the start index (inclusive) up to the end index (exclusive). Follows Go slice semantics (indices are clamped, invalid range returns empty list).",
		Category:    "List Operations",
		Args: []tool.ArgSpec{
			{Name: "list", Type: tool.ArgTypeSliceAny, Required: true, Description: "The list to slice."},
			{Name: "start", Type: tool.ArgTypeInt, Required: true, Description: "The starting index (inclusive)."},
			{Name: "end", Type: tool.ArgTypeInt, Required: true, Description: "The ending index (exclusive)."},
		},
		ReturnType:      tool.ArgTypeSliceAny,
		ReturnHelp:      "Returns a new list containing elements from the 'start' index (inclusive) up to the 'end' index (exclusive). Adheres to Go's slice semantics: indices are clamped to valid ranges (0 to list length). If 'start' > 'end' after clamping, or if 'start' is out of bounds (e.g. beyond list length), an empty list is returned. The original list is not modified.",
		Example:         `tool.List.Slice([1, 2, 3, 4, 5], 1, 4) // returns [2, 3, 4]`,
		ErrorConditions: "Returns an empty list for invalid or out-of-bounds start/end indices. Does not return an error for range issues.",
	}, Func: toolListSlice},
	{Spec: tool.ToolSpec{
		Name:        "List.Contains",
		Description: "Checks if a list contains a specific element (using deep equality comparison).",
		Category:    "List Operations",
		Args: []tool.ArgSpec{
			{Name: "list", Type: tool.ArgTypeSliceAny, Required: true, Description: "The list to search within."},
			{Name: "element", Type: tool.ArgTypeAny, Required: false, Description: "The element to search for (can be nil)."},
		},
		ReturnType:      tool.ArgTypeBool,
		ReturnHelp:      "Returns true if the 'list' contains the specified 'element', using deep equality for comparison. Returns false otherwise.",
		Example:         `tool.List.Contains([1, "a", true], "a") // returns true`,
		ErrorConditions: "None expected. Comparison with 'nil' elements is handled.",
	}, Func: toolListContains},
	{Spec: tool.ToolSpec{
		Name:            "List.Reverse",
		Description:     "Returns a *new* list with the elements in reverse order.",
		Category:        "List Operations",
		Args:            []tool.ArgSpec{{Name: "list", Type: tool.ArgTypeSliceAny, Required: true, Description: "The list to reverse."}},
		ReturnType:      tool.ArgTypeSliceAny,
		ReturnHelp:      "Returns a new list with the elements of the input 'list' in reverse order. The original list is not modified.",
		Example:         `tool.List.Reverse([1, 2, 3]) // returns [3, 2, 1]`,
		ErrorConditions: "None expected.",
	}, Func: toolListReverse},
	{Spec: tool.ToolSpec{
		Name:            "List.Sort",
		Description:     "Returns a *new* list with elements sorted. Restricted to lists containing only numbers (int/float) or only strings. Throws error for mixed types or non-sortable types (nil, bool, list, map).",
		Category:        "List Operations",
		Args:            []tool.ArgSpec{{Name: "list", Type: tool.ArgTypeSliceAny, Required: true, Description: "The list to sort."}},
		ReturnType:      tool.ArgTypeSliceAny,
		ReturnHelp:      "Returns a new list with elements sorted. The list must contain either all numbers (integers or floats, which will be sorted numerically) or all strings (sorted lexicographically). The original list is not modified. Returns an empty list if the input list is empty.",
		Example:         `tool.List.Sort([3, 1, 2]) // returns [1, 2, 3]\ntool.List.Sort(["c", "a", "b"]) // returns ["a", "b", "c"]`,
		ErrorConditions: "Returns an error (ErrListCannotSortMixedTypes) if the list contains mixed types (e.g., numbers and strings), nil elements, or other non-sortable types like booleans, maps, or other lists.",
	}, Func: toolListSort},
	{Spec: tool.ToolSpec{
		Name:            "List.Head",
		Description:     "Returns the first element of the list, or nil if the list is empty.",
		Category:        "List Operations",
		Args:            []tool.ArgSpec{{Name: "list", Type: tool.ArgTypeSliceAny, Required: true, Description: "The list to get the head from."}},
		ReturnType:      tool.ArgTypeAny,
		ReturnHelp:      "Returns the first element of the 'list'. If the list is empty, it returns nil.",
		Example:         `tool.List.Head([1, 2, 3]) // returns 1\ntool.List.Head([]) // returns nil`,
		ErrorConditions: "None expected. Returns nil for an empty list.",
	}, Func: toolListHead},
	{Spec: tool.ToolSpec{
		Name:            "List.Rest",
		Description:     "Returns a *new* list containing all elements except the first. Returns an empty list if the input list has 0 or 1 element.",
		Category:        "List Operations",
		Args:            []tool.ArgSpec{{Name: "list", Type: tool.ArgTypeSliceAny, Required: true, Description: "The list to get the rest from."}},
		ReturnType:      tool.ArgTypeSliceAny,
		ReturnHelp:      "Returns a new list containing all elements of the input 'list' except the first. If the list has 0 or 1 element, it returns an empty list. The original list is not modified.",
		Example:         `tool.List.Rest([1, 2, 3]) // returns [2, 3]\ntool.List.Rest([1]) // returns []`,
		ErrorConditions: "None expected. Returns an empty list for lists with 0 or 1 element.",
	}, Func: toolListRest},
	{Spec: tool.ToolSpec{
		Name:        "List.Tail",
		Description: "Returns a *new* list containing the last 'count' elements. Returns an empty list if count <= 0. Returns a copy of the whole list if count >= list length.",
		Category:    "List Operations",
		Args: []tool.ArgSpec{
			{Name: "list", Type: tool.ArgTypeSliceAny, Required: true, Description: "The list to get the tail from."},
			{Name: "count", Type: tool.ArgTypeInt, Required: true, Description: "The number of elements to take from the end."},
		},
		ReturnType:      tool.ArgTypeSliceAny,
		ReturnHelp:      "Returns a new list containing the last 'count' elements from the input 'list'. If 'count' is less than or equal to 0, an empty list is returned. If 'count' is greater than or equal to the list length, a copy of the original list is returned. The original list is not modified.",
		Example:         `tool.List.Tail([1, 2, 3, 4, 5], 3) // returns [3, 4, 5]\ntool.List.Tail([1, 2], 5) // returns [1, 2]`,
		ErrorConditions: "None expected. Handles various 'count' values appropriately, returning an empty list or a copy of the whole list as applicable.",
	}, Func: toolListTail},
	{Spec: tool.ToolSpec{
		Name:            "List.IsEmpty",
		Description:     "Returns true if the list has zero elements, false otherwise.",
		Category:        "List Operations",
		Args:            []tool.ArgSpec{{Name: "list", Type: tool.ArgTypeSliceAny, Required: true, Description: "The list to check."}},
		ReturnType:      tool.ArgTypeBool,
		ReturnHelp:      "Returns true if the 'list' contains zero elements, and false otherwise.",
		Example:         `tool.List.IsEmpty([]) // returns true\ntool.List.IsEmpty([1]) // returns false`,
		ErrorConditions: "None expected.",
	}, Func: toolListIsEmpty},
}
