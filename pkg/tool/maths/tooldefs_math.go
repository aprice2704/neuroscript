// NeuroScript Version: 0.3.1
// File version: 0.1.1
// Purpose: Populated Category, Example, ReturnHelp, and ErrorConditions for existing math tool specs.
// filename: pkg/tool/maths/tooldefs_math.go
// nlines: 80
// risk_rating: MEDIUM

package maths

import "github.com/aprice2704/neuroscript/pkg/tool"

const group = "math"
const Group = group

// mathToolsToRegister contains ToolImplementation definitions for Math tools.
// These definitions are based on the mathTools slice previously in tools_math.go.
var mathToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "Add",
			Group:       group,
			Description: "Calculates the sum of two numbers (integers or decimals). Strings convertible to numbers are accepted.",
			Category:    "Math Operations",
			Args: []tool.ArgSpec{
				{Name: "num1", Type: tool.ArgTypeFloat, Required: true, Description: "The first number (or numeric string) to add."},
				{Name: "num2", Type: tool.ArgTypeFloat, Required: true, Description: "The second number (or numeric string) to add."},
			},
			ReturnType:      tool.ArgTypeFloat,
			ReturnHelp:      "Returns the sum of num1 and num2 as a float64. Both inputs are expected to be (or be coercible to) numbers.",
			Example:         `tool.Add(5, 3.5) // returns 8.5`,
			ErrorConditions: "Returns an 'ErrInternalTool' if arguments cannot be processed as float64 (this scenario should ideally be caught by input validation before the tool function is called).",
		},
		Func: toolAdd, // From tools_math.go
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Subtract",
			Group:       group,
			Description: "Calculates the difference between two numbers (num1 - num2). Strings convertible to numbers are accepted.",
			Category:    "Math Operations",
			Args: []tool.ArgSpec{
				{Name: "num1", Type: tool.ArgTypeFloat, Required: true, Description: "The number to subtract from."},
				{Name: "num2", Type: tool.ArgTypeFloat, Required: true, Description: "The number to subtract."},
			},
			ReturnType:      tool.ArgTypeFloat,
			ReturnHelp:      "Returns the difference of num1 - num2 as a float64. Both inputs are expected to be (or be coercible to) numbers.",
			Example:         `tool.Subtract(10, 4.5) // returns 5.5`,
			ErrorConditions: "Returns an 'ErrInternalTool' if arguments cannot be processed as float64 (should be caught by validation).",
		},
		Func: toolSubtract, // From tools_math.go
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Multiply",
			Group:       group,
			Description: "Calculates the product of two numbers. Strings convertible to numbers are accepted.",
			Category:    "Math Operations",
			Args: []tool.ArgSpec{
				{Name: "num1", Type: tool.ArgTypeFloat, Required: true, Description: "The first number."},
				{Name: "num2", Type: tool.ArgTypeFloat, Required: true, Description: "The second number."},
			},
			ReturnType:      tool.ArgTypeFloat,
			ReturnHelp:      "Returns the product of num1 and num2 as a float64. Both inputs are expected to be (or be coercible to) numbers.",
			Example:         `tool.Multiply(6, 7.0) // returns 42.0`,
			ErrorConditions: "Returns an 'ErrInternalTool' if arguments cannot be processed as float64 (should be caught by validation).",
		},
		Func: toolMultiply, // From tools_math.go
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Divide",
			Group:       group,
			Description: "Calculates the division of two numbers (num1 / num2). Returns float. Handles division by zero.",
			Category:    "Math Operations",
			Args: []tool.ArgSpec{
				{Name: "num1", Type: tool.ArgTypeFloat, Required: true, Description: "The dividend."},
				{Name: "num2", Type: tool.ArgTypeFloat, Required: true, Description: "The divisor."},
			},
			ReturnType:      tool.ArgTypeFloat,
			ReturnHelp:      "Returns the result of num1 / num2 as a float64. Both inputs are expected to be (or be coercible to) numbers.",
			Example:         `tool.Divide(10, 4) // returns 2.5`,
			ErrorConditions: "Returns 'ErrDivisionByZero' if num2 is 0. Returns an 'ErrInternalTool' if arguments cannot be processed as float64 (should be caught by validation).",
		},
		Func: toolDivide, // From tools_math.go
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Modulo",
			Group:       group,
			Description: "Calculates the modulo (remainder) of two integers (num1 % num2). Handles division by zero.",
			Category:    "Math Operations",
			Args: []tool.ArgSpec{
				{Name: "num1", Type: tool.ArgTypeInt, Required: true, Description: "The dividend (must be integer)."},
				{Name: "num2", Type: tool.ArgTypeInt, Required: true, Description: "The divisor (must be integer)."},
			},
			ReturnType:      tool.ArgTypeInt,
			ReturnHelp:      "Returns the remainder of num1 % num2 as an int64. Both inputs must be integers.",
			Example:         `tool.Modulo(10, 3) // returns 1`,
			ErrorConditions: "Returns 'ErrDivisionByZero' if num2 is 0. Returns an 'ErrInternalTool' if arguments cannot be processed as int64 (should be caught by validation).",
		},
		Func: toolModulo, // From tools_math.go
	},
	// Add other math tools like Power, Sqrt, Abs, Max, Min, Round, Floor, Ceil, Random here
	// if their tool... functions and ToolImplementation structs are defined.
}
