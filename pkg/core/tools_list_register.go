// filename: pkg/core/tools_list_register.go
package core

import "fmt" // Keep fmt

// registerListTools adds list manipulation tools to the registry.
// Implementations are in tools_list_impl.go
// *** MODIFIED: Returns error ***
func registerListTools(registry *ToolRegistry) error {
	tools := []ToolImplementation{
		{Spec: ToolSpec{Name: "ListLength", Description: "Returns the number of elements in a list.", Args: []ArgSpec{{Name: "list", Type: ArgTypeSliceAny, Required: true}}, ReturnType: ArgTypeInt}, Func: toolListLength},
		{Spec: ToolSpec{Name: "ListAppend", Description: "Returns a *new* list with the given element added to the end.", Args: []ArgSpec{{Name: "list", Type: ArgTypeSliceAny, Required: true}, {Name: "element", Type: ArgTypeAny, Required: true}}, ReturnType: ArgTypeSliceAny}, Func: toolListAppend},
		{Spec: ToolSpec{Name: "ListPrepend", Description: "Returns a *new* list with the given element added to the beginning.", Args: []ArgSpec{{Name: "list", Type: ArgTypeSliceAny, Required: true}, {Name: "element", Type: ArgTypeAny, Required: true}}, ReturnType: ArgTypeSliceAny}, Func: toolListPrepend},
		{Spec: ToolSpec{Name: "ListGet", Description: "Safely gets the element at a specific index...", Args: []ArgSpec{{Name: "list", Type: ArgTypeSliceAny, Required: true}, {Name: "index", Type: ArgTypeInt, Required: true}, {Name: "default", Type: ArgTypeAny, Required: false}}, ReturnType: ArgTypeAny}, Func: toolListGet},
		{Spec: ToolSpec{Name: "ListSlice", Description: "Returns a *new* list containing elements...", Args: []ArgSpec{{Name: "list", Type: ArgTypeSliceAny, Required: true}, {Name: "start", Type: ArgTypeInt, Required: true}, {Name: "end", Type: ArgTypeInt, Required: true}}, ReturnType: ArgTypeSliceAny}, Func: toolListSlice},
		{Spec: ToolSpec{Name: "ListContains", Description: "Checks if a list contains a specific element (using deep equality).", Args: []ArgSpec{{Name: "list", Type: ArgTypeSliceAny, Required: true}, {Name: "element", Type: ArgTypeAny, Required: true}}, ReturnType: ArgTypeBool}, Func: toolListContains},
		{Spec: ToolSpec{Name: "ListReverse", Description: "Returns a *new* list with the elements in reverse order.", Args: []ArgSpec{{Name: "list", Type: ArgTypeSliceAny, Required: true}}, ReturnType: ArgTypeSliceAny}, Func: toolListReverse},
		{Spec: ToolSpec{Name: "ListSort", Description: "Returns a *new* list with elements sorted. Restricted...", Args: []ArgSpec{{Name: "list", Type: ArgTypeSliceAny, Required: true}}, ReturnType: ArgTypeAny}, Func: toolListSort},
		{Spec: ToolSpec{Name: "ListHead", Description: "Returns the first element of the list, or nil if the list is empty.", Args: []ArgSpec{{Name: "list", Type: ArgTypeSliceAny, Required: true}}, ReturnType: ArgTypeAny}, Func: toolListHead},
		{Spec: ToolSpec{Name: "ListRest", Description: "Returns a *new* list containing all elements except the first...", Args: []ArgSpec{{Name: "list", Type: ArgTypeSliceAny, Required: true}}, ReturnType: ArgTypeSliceAny}, Func: toolListRest},
		{Spec: ToolSpec{Name: "ListTail", Description: "Returns a *new* list containing the last 'count' elements...", Args: []ArgSpec{{Name: "list", Type: ArgTypeSliceAny, Required: true}, {Name: "count", Type: ArgTypeInt, Required: true}}, ReturnType: ArgTypeSliceAny}, Func: toolListTail},
		{Spec: ToolSpec{Name: "ListIsEmpty", Description: "Returns true if the list is nil or has zero elements, false otherwise.", Args: []ArgSpec{{Name: "list", Type: ArgTypeSliceAny, Required: true}}, ReturnType: ArgTypeBool}, Func: toolListIsEmpty},
	}
	for _, tool := range tools {
		// *** Check error from RegisterTool ***
		if err := registry.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register List tool %s: %w", tool.Spec.Name, err)
		}
	}
	return nil // Success
}
