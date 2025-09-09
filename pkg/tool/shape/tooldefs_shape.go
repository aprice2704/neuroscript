// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Defines the toolset for the json_lite package, under the 'shape' group.
// filename: pkg/tool/shape/tooldefs_shape.go
// nlines: 105
// risk_rating: LOW

package shape

import "github.com/aprice2704/neuroscript/pkg/tool"

const group = "shape"

// shapeToolsToRegister defines the ToolImplementation structs for the json_lite toolset.
var shapeToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "Validate",
			Group:       group,
			Description: "Validates a map against a Shape-Lite definition.",
			Category:    "Data Validation",
			Args: []tool.ArgSpec{
				{Name: "value", Type: tool.ArgTypeMap, Required: true, Description: "The data map to validate."},
				{Name: "shape", Type: tool.ArgTypeMap, Required: true, Description: "The Shape-Lite map to validate against."},
				{Name: "options", Type: tool.ArgTypeMap, Required: false, Description: "Options map, e.g., {\"allow_extra\": true, \"case_insensitive\": true}."},
			},
			ReturnType:      tool.ArgTypeBool,
			ReturnHelp:      "Returns true on success, otherwise returns a validation error.",
			Example:         `tool.shape.Validate(my_data, my_shape, {"allow_extra": true})`,
			ErrorConditions: "Returns validation errors (e.g., ErrValidationTypeMismatch) on failure.",
		},
		Func: toolShapeValidate,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Select",
			Group:       group,
			Description: "Selects a single value from a map or list using a Path-Lite expression.",
			Category:    "Data Selection",
			Args: []tool.ArgSpec{
				{Name: "value", Type: tool.ArgTypeAny, Required: true, Description: "The map or list to select from."},
				{Name: "path", Type: tool.ArgTypeAny, Required: true, Description: "The Path-Lite string or array-form list path."},
				{Name: "options", Type: tool.ArgTypeMap, Required: false, Description: "Options map, e.g., {\"case_insensitive\": true, \"missing_ok\": true}."},
			},
			ReturnType:      tool.ArgTypeAny,
			ReturnHelp:      "Returns the value found at the specified path.",
			Example:         `tool.shape.Select(my_data, "user.name", {"case_insensitive": true})`,
			ErrorConditions: "Returns selection errors (e.g., ErrMapKeyNotFound) on failure, unless 'missing_ok' is true.",
		},
		Func: toolShapeSelect,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "IsValidPath",
			Group:       group,
			Description: "Checks if a string is a syntactically valid Path-Lite expression.",
			Category:    "Data Validation",
			Args: []tool.ArgSpec{
				{Name: "path_string", Type: tool.ArgTypeString, Required: true, Description: "The Path-Lite string to check."},
			},
			ReturnType:      tool.ArgTypeBool,
			ReturnHelp:      "Returns true if the path has valid syntax, false otherwise.",
			Example:         `tool.shape.IsValidPath("a.b[0].c")`,
			ErrorConditions: "None. Always returns a boolean.",
		},
		Func: toolShapeIsValidPath,
	},
}
