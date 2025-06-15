// NeuroScript Version: 0.4.1
// File version: 5
// Purpose: Implemented error interface for ErrorValue to allow proper error propagation.
// filename: pkg/core/values.go
// nlines: 185
// risk_rating: MEDIUM

package core

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Value is the interface that all NeuroScript runtime values must implement.
type Value interface {
	Type() NeuroScriptType
	String() string
	IsTruthy() bool
}

// --- Primitive Value Wrappers ---

// StringValue wraps a string.
type StringValue struct{ Value string }

func (v StringValue) Type() NeuroScriptType { return TypeString }
func (v StringValue) String() string        { return v.Value }
func (v StringValue) IsTruthy() bool        { return len(v.Value) > 0 }

// NumberValue wraps a float64. The interpreter should convert all numbers to float64.
type NumberValue struct{ Value float64 }

func (v NumberValue) Type() NeuroScriptType { return TypeNumber }
func (v NumberValue) String() string        { return strconv.FormatFloat(v.Value, 'f', -1, 64) }
func (v NumberValue) IsTruthy() bool        { return v.Value != 0 }

// BoolValue wraps a bool.
type BoolValue struct{ Value bool }

func (v BoolValue) Type() NeuroScriptType { return TypeBoolean }
func (v BoolValue) String() string        { return strconv.FormatBool(v.Value) }
func (v BoolValue) IsTruthy() bool        { return v.Value }

// BytesValue wraps arbitrary binary data.
type BytesValue struct{ Value []byte }

func (v BytesValue) Type() NeuroScriptType { return TypeBytes } // <-- add TypeBytes to your enum
func (v BytesValue) String() string        { return fmt.Sprintf("bytes(%d)", len(v.Value)) }
func (v BytesValue) IsTruthy() bool        { return len(v.Value) > 0 }

// NilValue represents the nil value.
type NilValue struct{}

func (v NilValue) Type() NeuroScriptType { return TypeNil }
func (v NilValue) String() string        { return "nil" }
func (v NilValue) IsTruthy() bool        { return false }

// --- Complex Value Types ---

// ListValue wraps a slice of Value.
type ListValue struct {
	Value []Value
}

func (v ListValue) Type() NeuroScriptType { return TypeList }
func (v ListValue) String() string {
	items := make([]string, len(v.Value))
	for i, item := range v.Value {
		items[i] = item.String()
	}
	return fmt.Sprintf("[%s]", strings.Join(items, ", "))
}
func (v ListValue) IsTruthy() bool { return len(v.Value) > 0 }

// MapValue wraps a map of string to Value.
type MapValue struct {
	Value map[string]Value
}

func (v MapValue) Type() NeuroScriptType { return TypeMap }
func (v MapValue) String() string {
	items := make([]string, 0, len(v.Value))
	for k, val := range v.Value {
		items = append(items, fmt.Sprintf("%q: %s", k, val.String()))
	}
	return fmt.Sprintf("{%s}", strings.Join(items, ", "))
}
func (v MapValue) IsTruthy() bool { return len(v.Value) > 0 }

// ErrorValue represents a structured error, conforming to the standardized error map.
type ErrorValue struct {
	Value map[string]Value
}

func (v ErrorValue) Type() NeuroScriptType { return TypeError }
func (v ErrorValue) String() string {
	if msgVal, ok := v.Value["message"]; ok {
		return fmt.Sprintf("error: %s", msgVal.String())
	}
	return "error: (unspecified)"
}
func (v ErrorValue) IsTruthy() bool { return false }
func (v ErrorValue) Error() string  { return v.String() }

// EventValue represents a structured event, holding its name, source, and payload.
type EventValue struct {
	Value map[string]Value
}

func (v EventValue) Type() NeuroScriptType { return TypeEvent }
func (v EventValue) String() string {
	if nameVal, ok := v.Value["name"]; ok {
		return fmt.Sprintf("event: %s", nameVal.String())
	}
	return "event: (unnamed)"
}
func (v EventValue) IsTruthy() bool { return true }

// TimedateValue wraps Go's time.Time for use in NeuroScript.
type TimedateValue struct {
	Value time.Time
}

func (v TimedateValue) Type() NeuroScriptType { return TypeTimedate }
func (v TimedateValue) String() string        { return v.Value.Format(time.RFC3339Nano) }
func (v TimedateValue) IsTruthy() bool        { return !v.Value.IsZero() }

// FuzzyValue represents a value with a degree of membership between 0.0 and 1.0.
type FuzzyValue struct {
	// μ represents the membership degree, clamped to the range [0.0, 1.0].
	μ float64
}

// NewFuzzyValue creates a FuzzyValue, clamping the input to the valid range.
func NewFuzzyValue(val float64) FuzzyValue {
	if val < 0.0 {
		val = 0.0
	}
	if val > 1.0 {
		val = 1.0
	}
	return FuzzyValue{μ: val}
}

func (v FuzzyValue) Type() NeuroScriptType { return TypeFuzzy }
func (v FuzzyValue) String() string        { return strconv.FormatFloat(v.μ, 'f', -1, 64) }
func (v FuzzyValue) IsTruthy() bool        { return v.μ > 0.5 }

// FunctionValue wraps a Procedure.
type FunctionValue struct{ Value Procedure }

func (v FunctionValue) Type() NeuroScriptType { return TypeFunction }
func (v FunctionValue) String() string        { return fmt.Sprintf("<function %s>", v.Value.Name) }
func (v FunctionValue) IsTruthy() bool        { return true }

// ToolValue wraps a ToolImplementation.
type ToolValue struct{ Value ToolImplementation }

func (v ToolValue) Type() NeuroScriptType { return TypeTool }
func (v ToolValue) String() string        { return fmt.Sprintf("<tool %s>", v.Value.Spec.Name) }
func (v ToolValue) IsTruthy() bool        { return true }

// --- Constructors for Complex Types ---

// NewListValue is a helper to create a ListValue, ensuring the slice is initialized.
func NewListValue(val []Value) ListValue {
	if val == nil {
		val = []Value{}
	}
	return ListValue{Value: val}
}

// NewMapValue is a helper to create a MapValue, ensuring the map is initialized.
func NewMapValue(val map[string]Value) MapValue {
	if val == nil {
		val = make(map[string]Value)
	}
	return MapValue{Value: val}
}
