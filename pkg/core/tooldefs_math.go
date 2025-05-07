// NeuroScript Version: 0.3.1
// File version: 0.1.0
// Defines ToolImplementation structs for Math tools.
// filename: pkg/core/tooldefs_math.go

package core

// mathToolsToRegister contains ToolImplementation definitions for Math tools.
// These definitions are based on the mathTools slice previously in tools_math.go.
var mathToolsToRegister = []ToolImplementation{
	{
		Spec: ToolSpec{
			Name:        "Add",
			Description: "Calculates the sum of two numbers (integers or decimals). Strings convertible to numbers are accepted.",
			Args: []ArgSpec{
				{Name: "num1", Type: ArgTypeFloat, Required: true, Description: "The first number (or numeric string) to add."},
				{Name: "num2", Type: ArgTypeFloat, Required: true, Description: "The second number (or numeric string) to add."},
			},
			ReturnType: ArgTypeFloat,
		},
		Func: toolAdd, // From tools_math.go
	},
	{
		Spec: ToolSpec{
			Name:        "Subtract",
			Description: "Calculates the difference between two numbers (num1 - num2). Strings convertible to numbers are accepted.",
			Args: []ArgSpec{
				{Name: "num1", Type: ArgTypeFloat, Required: true, Description: "The number to subtract from."},
				{Name: "num2", Type: ArgTypeFloat, Required: true, Description: "The number to subtract."},
			},
			ReturnType: ArgTypeFloat,
		},
		Func: toolSubtract, // From tools_math.go
	},
	{
		Spec: ToolSpec{
			Name:        "Multiply",
			Description: "Calculates the product of two numbers. Strings convertible to numbers are accepted.",
			Args: []ArgSpec{
				{Name: "num1", Type: ArgTypeFloat, Required: true, Description: "The first number."},
				{Name: "num2", Type: ArgTypeFloat, Required: true, Description: "The second number."},
			},
			ReturnType: ArgTypeFloat,
		},
		Func: toolMultiply, // From tools_math.go
	},
	{
		Spec: ToolSpec{
			Name:        "Divide",
			Description: "Calculates the division of two numbers (num1 / num2). Returns float. Handles division by zero.",
			Args: []ArgSpec{
				{Name: "num1", Type: ArgTypeFloat, Required: true, Description: "The dividend."},
				{Name: "num2", Type: ArgTypeFloat, Required: true, Description: "The divisor."},
			},
			ReturnType: ArgTypeFloat,
		},
		Func: toolDivide, // From tools_math.go
	},
	{
		Spec: ToolSpec{
			Name:        "Modulo",
			Description: "Calculates the modulo (remainder) of two integers (num1 % num2). Handles division by zero.",
			Args: []ArgSpec{
				{Name: "num1", Type: ArgTypeInt, Required: true, Description: "The dividend (must be integer)."},
				{Name: "num2", Type: ArgTypeInt, Required: true, Description: "The divisor (must be integer)."},
			},
			ReturnType: ArgTypeInt,
		},
		Func: toolModulo, // From tools_math.go
	},
	// Add other math tools like Power, Sqrt, Abs, Max, Min, Round, Floor, Ceil, Random here
	// if their tool... functions and ToolImplementation structs are defined.
}
