// NeuroScript Version: 0.3.1
// File version: 0.1.0
// Defines ToolImplementation structs for list tools.
// filename: pkg/core/tooldefs_io.go

package core

// listToolsToRegister contains ToolImplementation definitions for list tools.
var listToolsToRegister = []ToolImplementation{
	// Existing tools...
	{Spec: ToolSpec{
		Name:        "ListLength",
		Description: "Returns the number of elements in a list.",
		Args:        []ArgSpec{{Name: "list", Type: ArgTypeSliceAny, Required: true}},
		ReturnType:  ArgTypeInt,
	}, Func: toolListLength},
	{Spec: ToolSpec{
		Name:        "ListAppend",
		Description: "Returns a *new* list with the given element added to the end.",
		Args: []ArgSpec{
			{Name: "list", Type: ArgTypeSliceAny, Required: true},
			// *** MODIFIED: Allow nil element by making it not strictly required for validation ***
			{Name: "element", Type: ArgTypeAny, Required: false},
		},
		ReturnType: ArgTypeSliceAny,
	}, Func: toolListAppend},
	{Spec: ToolSpec{
		Name:        "ListPrepend",
		Description: "Returns a *new* list with the given element added to the beginning.",
		Args: []ArgSpec{
			{Name: "list", Type: ArgTypeSliceAny, Required: true},
			// *** MODIFIED: Allow nil element by making it not strictly required for validation ***
			{Name: "element", Type: ArgTypeAny, Required: false},
		},
		ReturnType: ArgTypeSliceAny,
	}, Func: toolListPrepend},
	{Spec: ToolSpec{
		Name:        "ListGet",
		Description: "Safely gets the element at a specific index (0-based). Returns nil or the optional default value if the index is out of bounds.",
		Args: []ArgSpec{
			{Name: "list", Type: ArgTypeSliceAny, Required: true},
			{Name: "index", Type: ArgTypeInt, Required: true},
			{Name: "default", Type: ArgTypeAny, Required: false}, // Optional default value
		},
		ReturnType: ArgTypeAny,
	}, Func: toolListGet},
	{Spec: ToolSpec{
		Name:        "ListSlice",
		Description: "Returns a *new* list containing elements from the start index (inclusive) up to the end index (exclusive). Follows Go slice semantics (indices are clamped, invalid range returns empty list).",
		Args: []ArgSpec{
			{Name: "list", Type: ArgTypeSliceAny, Required: true},
			{Name: "start", Type: ArgTypeInt, Required: true},
			{Name: "end", Type: ArgTypeInt, Required: true},
		},
		ReturnType: ArgTypeSliceAny,
	}, Func: toolListSlice},
	{Spec: ToolSpec{
		Name:        "ListContains",
		Description: "Checks if a list contains a specific element (using deep equality comparison).",
		Args: []ArgSpec{
			{Name: "list", Type: ArgTypeSliceAny, Required: true},
			// *** MODIFIED: Allow nil element by making it not strictly required for validation ***
			{Name: "element", Type: ArgTypeAny, Required: false},
		},
		ReturnType: ArgTypeBool,
	}, Func: toolListContains},
	{Spec: ToolSpec{
		Name:        "ListReverse",
		Description: "Returns a *new* list with the elements in reverse order.",
		Args:        []ArgSpec{{Name: "list", Type: ArgTypeSliceAny, Required: true}},
		ReturnType:  ArgTypeSliceAny,
	}, Func: toolListReverse},
	{Spec: ToolSpec{
		Name:        "ListSort",
		Description: "Returns a *new* list with elements sorted. Restricted to lists containing only numbers (int/float) or only strings. Throws error for mixed types or non-sortable types (nil, bool, list, map).",
		Args:        []ArgSpec{{Name: "list", Type: ArgTypeSliceAny, Required: true}},
		ReturnType:  ArgTypeSliceAny, // Return type is always a slice
	}, Func: toolListSort},
	{Spec: ToolSpec{
		Name:        "ListHead",
		Description: "Returns the first element of the list, or nil if the list is empty.",
		Args:        []ArgSpec{{Name: "list", Type: ArgTypeSliceAny, Required: true}},
		ReturnType:  ArgTypeAny, // Can return any type or nil
	}, Func: toolListHead},
	{Spec: ToolSpec{
		Name:        "ListRest",
		Description: "Returns a *new* list containing all elements except the first. Returns an empty list if the input list has 0 or 1 element.",
		Args:        []ArgSpec{{Name: "list", Type: ArgTypeSliceAny, Required: true}},
		ReturnType:  ArgTypeSliceAny,
	}, Func: toolListRest},
	{Spec: ToolSpec{
		Name:        "ListTail",
		Description: "Returns a *new* list containing the last 'count' elements. Returns an empty list if count <= 0. Returns a copy of the whole list if count >= list length.",
		Args: []ArgSpec{
			{Name: "list", Type: ArgTypeSliceAny, Required: true},
			{Name: "count", Type: ArgTypeInt, Required: true},
		},
		ReturnType: ArgTypeSliceAny,
	}, Func: toolListTail},
	{Spec: ToolSpec{
		Name:        "ListIsEmpty",
		Description: "Returns true if the list has zero elements, false otherwise.",
		Args:        []ArgSpec{{Name: "list", Type: ArgTypeSliceAny, Required: true}},
		ReturnType:  ArgTypeBool,
	}, Func: toolListIsEmpty},
}
