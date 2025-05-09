// NeuroScript Version: 0.3.1
// File version: 0.1.2
// Revert 'element' to Required: false for Append, Prepend, Contains.
// nlines: 100
// risk_rating: MEDIUM
// filename: pkg/core/tooldefs_list.go

package core

// listToolsToRegister contains ToolImplementation definitions for list tools.
var listToolsToRegister = []ToolImplementation{
	{Spec: ToolSpec{
		Name:        "List.Length",
		Description: "Returns the number of elements in a list.",
		Args:        []ArgSpec{{Name: "list", Type: ArgTypeSliceAny, Required: true, Description: "The list to measure."}},
		ReturnType:  ArgTypeInt,
	}, Func: toolListLength},
	{Spec: ToolSpec{
		Name:        "List.Append",
		Description: "Returns a *new* list with the given element added to the end.",
		Args: []ArgSpec{
			{Name: "list", Type: ArgTypeSliceAny, Required: true, Description: "The list to append to."},
			// Reverted: Allow nil element by making it not strictly required for validation.
			// The tool function itself might treat nil differently if needed.
			{Name: "element", Type: ArgTypeAny, Required: false, Description: "The element to append (can be nil)."},
		},
		ReturnType: ArgTypeSliceAny,
	}, Func: toolListAppend},
	{Spec: ToolSpec{
		Name:        "List.Prepend",
		Description: "Returns a *new* list with the given element added to the beginning.",
		Args: []ArgSpec{
			{Name: "list", Type: ArgTypeSliceAny, Required: true, Description: "The list to prepend to."},
			// Reverted: Allow nil element.
			{Name: "element", Type: ArgTypeAny, Required: false, Description: "The element to prepend (can be nil)."},
		},
		ReturnType: ArgTypeSliceAny,
	}, Func: toolListPrepend},
	{Spec: ToolSpec{
		Name:        "List.Get",
		Description: "Safely gets the element at a specific index (0-based). Returns nil or the optional default value if the index is out of bounds.",
		Args: []ArgSpec{
			{Name: "list", Type: ArgTypeSliceAny, Required: true, Description: "The list to get from."},
			{Name: "index", Type: ArgTypeInt, Required: true, Description: "The 0-based index."},
			{Name: "default", Type: ArgTypeAny, Required: false, Description: "Optional default value if index is out of bounds."},
		},
		ReturnType: ArgTypeAny,
	}, Func: toolListGet},
	{Spec: ToolSpec{
		Name:        "List.Slice",
		Description: "Returns a *new* list containing elements from the start index (inclusive) up to the end index (exclusive). Follows Go slice semantics (indices are clamped, invalid range returns empty list).",
		Args: []ArgSpec{
			{Name: "list", Type: ArgTypeSliceAny, Required: true, Description: "The list to slice."},
			{Name: "start", Type: ArgTypeInt, Required: true, Description: "The starting index (inclusive)."},
			{Name: "end", Type: ArgTypeInt, Required: true, Description: "The ending index (exclusive)."},
		},
		ReturnType: ArgTypeSliceAny,
	}, Func: toolListSlice},
	{Spec: ToolSpec{
		Name:        "List.Contains",
		Description: "Checks if a list contains a specific element (using deep equality comparison).",
		Args: []ArgSpec{
			{Name: "list", Type: ArgTypeSliceAny, Required: true, Description: "The list to search within."},
			// Reverted: Allow nil element.
			{Name: "element", Type: ArgTypeAny, Required: false, Description: "The element to search for (can be nil)."},
		},
		ReturnType: ArgTypeBool,
	}, Func: toolListContains},
	{Spec: ToolSpec{
		Name:        "List.Reverse",
		Description: "Returns a *new* list with the elements in reverse order.",
		Args:        []ArgSpec{{Name: "list", Type: ArgTypeSliceAny, Required: true, Description: "The list to reverse."}},
		ReturnType:  ArgTypeSliceAny,
	}, Func: toolListReverse},
	{Spec: ToolSpec{
		Name:        "List.Sort",
		Description: "Returns a *new* list with elements sorted. Restricted to lists containing only numbers (int/float) or only strings. Throws error for mixed types or non-sortable types (nil, bool, list, map).",
		Args:        []ArgSpec{{Name: "list", Type: ArgTypeSliceAny, Required: true, Description: "The list to sort."}},
		ReturnType:  ArgTypeSliceAny, // Return type is always a slice
	}, Func: toolListSort},
	{Spec: ToolSpec{
		Name:        "List.Head",
		Description: "Returns the first element of the list, or nil if the list is empty.",
		Args:        []ArgSpec{{Name: "list", Type: ArgTypeSliceAny, Required: true, Description: "The list to get the head from."}},
		ReturnType:  ArgTypeAny, // Can return any type or nil
	}, Func: toolListHead},
	{Spec: ToolSpec{
		Name:        "List.Rest",
		Description: "Returns a *new* list containing all elements except the first. Returns an empty list if the input list has 0 or 1 element.",
		Args:        []ArgSpec{{Name: "list", Type: ArgTypeSliceAny, Required: true, Description: "The list to get the rest from."}},
		ReturnType:  ArgTypeSliceAny,
	}, Func: toolListRest},
	{Spec: ToolSpec{
		Name:        "List.Tail",
		Description: "Returns a *new* list containing the last 'count' elements. Returns an empty list if count <= 0. Returns a copy of the whole list if count >= list length.",
		Args: []ArgSpec{
			{Name: "list", Type: ArgTypeSliceAny, Required: true, Description: "The list to get the tail from."},
			{Name: "count", Type: ArgTypeInt, Required: true, Description: "The number of elements to take from the end."},
		},
		ReturnType: ArgTypeSliceAny,
	}, Func: toolListTail},
	{Spec: ToolSpec{
		Name:        "List.IsEmpty",
		Description: "Returns true if the list has zero elements, false otherwise.",
		Args:        []ArgSpec{{Name: "list", Type: ArgTypeSliceAny, Required: true, Description: "The list to check."}},
		ReturnType:  ArgTypeBool,
	}, Func: toolListIsEmpty},
}
