// NeuroScript Version: 0.5.2
// File version: 15
// Purpose: Adds nil-safety to (*MapValue).String() method. Added concrete HandleValue type.
// Latest change: Added HandleValue struct and its methods.
// filename: pkg/lang/values.go
// nlines: 279
// risk_rating: HIGH

package lang

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// Value is the interface that all NeuroScript runtime values must implement.
type Value interface {
	Type() NeuroScriptType
	String() string
	IsTruthy() bool
}

// valueToString is a helper to get the string representation of a value,
// ensuring that StringValues are properly quoted.
func valueToString(v Value) string {
	if sv, ok := v.(StringValue); ok {
		return strconv.Quote(sv.Value)
	}
	return v.String()
}

// --- Primitive Value Wrappers ---

type StringValue struct{ Value string }

func (v StringValue) Type() NeuroScriptType { return TypeString }
func (v StringValue) String() string        { return v.Value }
func (v StringValue) IsTruthy() bool        { return len(v.Value) > 0 }

type NumberValue struct{ Value float64 }

func (v NumberValue) Type() NeuroScriptType { return TypeNumber }
func (v NumberValue) String() string        { return strconv.FormatFloat(v.Value, 'f', -1, 64) }
func (v NumberValue) IsTruthy() bool        { return v.Value != 0 }

type BoolValue struct{ Value bool }

func (v BoolValue) Type() NeuroScriptType { return TypeBoolean }
func (v BoolValue) String() string        { return strconv.FormatBool(v.Value) }
func (v BoolValue) IsTruthy() bool        { return v.Value }

type BytesValue struct{ Value []byte }

func (v BytesValue) Type() NeuroScriptType { return TypeBytes }
func (v BytesValue) String() string        { return fmt.Sprintf("bytes(%d)", len(v.Value)) }
func (v BytesValue) IsTruthy() bool        { return len(v.Value) > 0 }

type NilValue struct{}

func (v NilValue) Type() NeuroScriptType { return TypeNil }
func (v NilValue) String() string        { return "nil" }
func (v NilValue) IsTruthy() bool        { return false }

// --- Complex Value Types ---

type ListValue struct {
	Value []Value
}

func (v ListValue) Type() NeuroScriptType { return TypeList }
func (v ListValue) String() string {
	items := make([]string, len(v.Value))
	for i, item := range v.Value {
		items[i] = valueToString(item)
	}
	return fmt.Sprintf("[%s]", strings.Join(items, ", "))
}
func (v ListValue) IsTruthy() bool { return len(v.Value) > 0 }

type MapValue struct {
	Value map[string]Value
}

func (v MapValue) Type() NeuroScriptType { return TypeMap }
func (v MapValue) String() string {
	// FIX: Add nil-safety for pointer receiver
	if v.Value == nil {
		return "nil"
	}
	items := make([]string, 0, len(v.Value))
	for k, val := range v.Value {
		items = append(items, fmt.Sprintf("%q: %s", k, valueToString(val)))
	}
	return fmt.Sprintf("{%s}", strings.Join(items, ", "))
}
func (v MapValue) IsTruthy() bool { return len(v.Value) > 0 }

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

type TimedateValue struct {
	Value time.Time
}

func (v TimedateValue) Type() NeuroScriptType { return TypeTimedate }
func (v TimedateValue) String() string        { return v.Value.Format(time.RFC3339Nano) }
func (v TimedateValue) IsTruthy() bool        { return !v.Value.IsZero() }

type FuzzyValue struct {
	μ float64
}

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

type FunctionValue struct{ Value Callable }

func (v FunctionValue) Type() NeuroScriptType { return TypeFunction }
func (v FunctionValue) String() string {
	if v.Value == nil {
		return "<nil function>"
	}
	return fmt.Sprintf("function<%s>", v.Value.Name())
}
func (v FunctionValue) IsTruthy() bool { return v.Value != nil }

// ToolValue now holds an 'any' type to break the import cycle.
// The interpreter is responsible for asserting it to an interfaces.Tool.
type ToolValue struct{ Value any }

func (v ToolValue) Type() NeuroScriptType { return TypeTool }
func (v ToolValue) String() string {
	if v.Value == nil {
		return "<nil tool>"
	}
	// Attempt to get a name via a temporary interface
	type Namer interface {
		Name() string
	}
	if n, ok := v.Value.(Namer); ok {
		return fmt.Sprintf("tool<%s>", n.Name())
	}
	return "tool<unknown>"
}
func (v ToolValue) IsTruthy() bool { return v.Value != nil }

// --- Handle Value ---

// HandleValue is a concrete implementation of the interfaces.HandleValue and lang.Value.
type HandleValue struct {
	ID   string
	Kind string
}

func NewHandleValue(id, kind string) interfaces.HandleValue {
	return HandleValue{ID: id, Kind: kind}
}

// Type implements lang.Value
func (v HandleValue) Type() NeuroScriptType { return TypeHandle }

// String implements lang.Value. It provides an opaque, debug representation.
// It MUST NOT expose the internal ID for security/opaqueness.
func (v HandleValue) String() string {
	return fmt.Sprintf("<handle %s>", v.Kind)
}

// IsTruthy implements lang.Value
func (v HandleValue) IsTruthy() bool { return v.ID != "" }

// HandleID implements interfaces.HandleValue
func (v HandleValue) HandleID() string { return v.ID }

// HandleKind implements interfaces.HandleValue
func (v HandleValue) HandleKind() string { return v.Kind }

// --- Constructors for Complex Types ---

func NewListValue(val []Value) ListValue {
	if val == nil {
		val = []Value{}
	}
	return ListValue{Value: val}
}

func NewMapValue(val map[string]Value) *MapValue {
	if val == nil {
		val = make(map[string]Value)
	}
	return &MapValue{Value: val}
}

func NewErrorValue(code, message string, details Value) ErrorValue {
	if details == nil {
		details = NilValue{}
	}
	return ErrorValue{Value: map[string]Value{
		ErrorKeyCode:    StringValue{Value: code},
		ErrorKeyMessage: StringValue{Value: message},
		ErrorKeyDetails: details,
	}}
}

func NewErrorValueFromRuntimeError(re *RuntimeError) ErrorValue {
	if re == nil {
		return NewErrorValue("E_NIL", "nil runtime error provided", NilValue{})
	}

	var detailsVal Value = NilValue{}
	if re.Wrapped != nil {
		detailsVal = StringValue{Value: re.Wrapped.Error()}
	}

	return NewErrorValue(
		strconv.Itoa(int(re.Code)),
		re.Message,
		detailsVal,
	)
}
