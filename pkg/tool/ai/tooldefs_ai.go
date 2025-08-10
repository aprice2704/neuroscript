// NeuroScript Version: 0.3.0
// File version: 3
// Purpose: Defines the toolset for the AI package. Aligned with new ToolImplementation spec.
// filename: pkg/tool/ai/tooldefs_ai.go
// nlines: 77
// risk_rating: LOW

package ai

import "github.com/aprice2704/neuroscript/pkg/tool"

const group = "ai"
const Group = group

// aiToolsToRegister contains ToolImplementation definitions for AI tools.
var aiToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "Validate",
			Group:       group,
			Description: "Validates a map against a shape definition, according to Shape-Lite spec.",
			Category:    "Data Validation",
			Args: []tool.ArgSpec{
				{Name: "value", Type: tool.ArgTypeMap, Required: true, Description: "The data map to validate."},
				{Name: "shape", Type: tool.ArgTypeMap, Required: true, Description: "The shape map to validate against."},
				{Name: "allow_extra", Type: tool.ArgTypeBool, Required: false, Description: "If true, allows extra keys in the value not present in the shape."},
			},
			ReturnType:      tool.ArgTypeBool,
			ReturnHelp:      "Returns true on success, otherwise returns a validation error.",
			Example:         `tool.ai.Validate(my_data, my_shape, false)`,
			ErrorConditions: "Returns 'ErrValidationRequiredArgMissing', 'ErrValidationTypeMismatch', or 'ErrInvalidArgument' on failure.",
		},
		Func: Validate,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Select",
			Group:       group,
			Description: "Selects a single value from a map or list using a path, according to Path-Lite spec.",
			Category:    "Data Selection",
			Args: []tool.ArgSpec{
				{Name: "value", Type: tool.ArgTypeAny, Required: true, Description: "The map or list to select from."},
				{Name: "path", Type: tool.ArgTypeAny, Required: true, Description: "The string or list path to the desired value."},
				{Name: "missing_ok", Type: tool.ArgTypeBool, Required: false, Description: "If true, returns nil if the path does not exist instead of failing."},
			},
			ReturnType:      tool.ArgTypeAny,
			ReturnHelp:      "Returns the value found at the specified path.",
			Example:         `tool.ai.Select(my_data, "user.name")`,
			ErrorConditions: "Returns 'ErrMapKeyNotFound', 'ErrListIndexOutOfBounds', or 'ErrInvalidPath' on failure.",
		},
		Func: Select,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "SelectMany",
			Group:       group,
			Description: "Selects multiple values from a map or list using a map of target keys to paths.",
			Category:    "Data Selection",
			Args: []tool.ArgSpec{
				{Name: "value", Type: tool.ArgTypeAny, Required: true, Description: "The map or list to select from."},
				{Name: "extracts", Type: tool.ArgTypeMap, Required: true, Description: "A map where keys are new names and values are the paths to extract."},
				{Name: "missing_ok", Type: tool.ArgTypeBool, Required: false, Description: "If true, keys for missing paths will be omitted from the result instead of failing."},
			},
			ReturnType:      tool.ArgTypeMap,
			ReturnHelp:      "Returns a new map containing the extracted key-value pairs.",
			Example:         `tool.ai.SelectMany(my_data, {"name": "user.name", "city": "user.address.city"})`,
			ErrorConditions: "Returns 'ErrMapKeyNotFound' or 'ErrInvalidPath' on failure.",
		},
		Func: SelectMany,
	},
}
