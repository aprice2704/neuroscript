// NeuroScript Version: 0.5.2
// File version: 9
// Purpose: Corrected NewErrorValue to return an ErrorValue struct instead of a MapValue, ensuring typeof() reports the correct type.
// filename: pkg/lang/values.go
// nlines: 220
// risk_rating: MEDIUM

package lang

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

// NumberValue wraps a float64.
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

func (v BytesValue) Type() NeuroScriptType { return TypeBytes }
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

// ErrorValue represents a structured error.
type ErrorValue struct {
	Value map[string]Value
}

func (v ErrorValue) Type() NeuroScriptType { return TypeError }
func (v ErrorValue) String() string {
	if msgVal, ok := v.Value[ErrorKeyMessage]; ok {
		return fmt.Sprintf("error: %s", msgVal.String())
	}
	return "error: (unspecified)"
}
func (v ErrorValue) IsTruthy() bool { return false }
func (v ErrorValue) Error() string  { return v.String() }

// EventValue represents a structured event.
type EventValue struct {
	Value map[string]Value
}

func (v EventValue) Type() NeuroScriptType { return TypeEvent }
func (v EventValue) String() string {
	if nameVal, ok := v.Value[EventKeyName]; ok {
		return fmt.Sprintf("event: %s", nameVal.String())
	}
	return "event: (unnamed)"
}
func (v EventValue) IsTruthy() bool { return true }

// TimedateValue wraps Go's time.Time.
type TimedateValue struct {
	Value time.Time
}

func (v TimedateValue) Type() NeuroScriptType { return TypeTimedate }
func (v TimedateValue) String() string        { return v.Value.Format(time.RFC3339Nano) }
func (v TimedateValue) IsTruthy() bool        { return !v.Value.IsZero() }

// FuzzyValue represents a value with a degree of membership between 0.0 and 1.0.
type FuzzyValue struct {
	μ float64
}

// GetValue returns the raw float64 value of the fuzzy number.
func (v FuzzyValue) GetValue() float64 {
	return v.μ
}

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

// FunctionValue wraps a Callable interface.
type FunctionValue struct{ Value Callable }

func (v FunctionValue) Type() NeuroScriptType { return TypeFunction }
func (v FunctionValue) String() string {
	if v.Value == nil {
		return "<nil function>"
	}
	return fmt.Sprintf("function<%s>", v.Value.Name())
}
func (v FunctionValue) IsTruthy() bool { return v.Value != nil }

// ToolValue wraps a Tool interface.
type ToolValue struct{ Value Tool }

func (v ToolValue) Type() NeuroScriptType { return TypeTool }
func (v ToolValue) String() string {
	if v.Value == nil {
		return "<nil tool>"
	}
	return fmt.Sprintf("tool<%s>", v.Value.Name())
}
func (v ToolValue) IsTruthy() bool { return v.Value != nil }

// --- Constructors for Complex Types ---

func NewListValue(val []Value) ListValue {
	if val == nil {
		val = []Value{}
	}
	return ListValue{Value: val}
}

func NewMapValue(val map[string]Value) MapValue {
	if val == nil {
		val = make(map[string]Value)
	}
	return MapValue{Value: val}
}

// NewErrorValue creates the standard map structure for a tool-level error
// and returns it as an ErrorValue.
func NewErrorValue(code, message string, details Value) ErrorValue {
	if details == nil {
		details = &NilValue{}
	}
	return ErrorValue{Value: map[string]Value{
		ErrorKeyCode:    StringValue{Value: code},
		ErrorKeyMessage: StringValue{Value: message},
		ErrorKeyDetails: details,
	}}
}
