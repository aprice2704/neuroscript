// NeuroScript Version: 0.8.0
// File version: 12
// Purpose: Corrected the test logic by removing 'root' from the isolatedFields map; it is a shared field.
// filename: pkg/interpreter/interpreter_clone_internal_test.go

package interpreter_test

import (
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
)

// areFieldsEqual checks if two reflected fields are equal.
// It safely handles unexported fields, pointers, interfaces, and value types.
func areFieldsEqual(t *testing.T, fieldName string, v1, v2 reflect.Value) bool {
	t.Helper()

	if v1.Kind() == reflect.Interface {
		if v1.IsNil() && v2.IsNil() {
			return true
		}
		if v1.IsNil() != v2.IsNil() {
			return false
		}
		v1 = v1.Elem()
		v2 = v2.Elem()
	}

	switch v1.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan, reflect.UnsafePointer:
		if v1.IsNil() && v2.IsNil() {
			return true
		}
		if v1.IsNil() != v2.IsNil() {
			return false
		}
		return v1.Pointer() == v2.Pointer()
	}

	if !v1.CanInterface() {
		switch v1.Kind() {
		case reflect.Bool:
			return v1.Bool() == v2.Bool()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return v1.Int() == v2.Int()
		case reflect.String:
			return v1.String() == v2.String()
		default:
			return false
		}
	}

	return reflect.DeepEqual(v1.Interface(), v2.Interface())
}

// TestInterpreter_Clone_Integrity uses reflection to ensure that the clone() method
// correctly handles every field in the Interpreter struct.
func TestInterpreter_Clone_Integrity(t *testing.T) {
	t.Logf("[DEBUG] Turn 1: Starting TestInterpreter_Clone_Integrity.")
	// The harness is not used here because we are testing the unexported `clone` method directly.
	// We still need a valid parent interpreter to start with.
	parent, err := interpreter.NewTestInterpreter(t, nil, nil, true)
	if err != nil {
		t.Fatalf("Failed to create parent interpreter: %v", err)
	}
	parent.SetSandboxDir("/test/sandbox/path")
	t.Logf("[DEBUG] Turn 2: Parent interpreter created.")

	clone := parent.Clone()
	t.Logf("[DEBUG] Turn 3: Interpreter cloned.")

	if clone.SandboxDir() != parent.SandboxDir() {
		t.Errorf("Sandbox path was not propagated to clone. Parent: '%s', Clone: '%s'",
			parent.SandboxDir(), clone.SandboxDir())
	}

	isolatedFields := map[string]bool{
		"id":              true,
		"state":           true,
		"tools":           true,
		"evaluate":        true,
		"cloneRegistry":   true,
		"cloneRegistryMu": true,
	}

	parentVal := reflect.ValueOf(parent).Elem()
	cloneVal := reflect.ValueOf(clone).Elem()
	structType := parentVal.Type()
	checkedFields := make(map[string]bool)

	t.Logf("[DEBUG] Turn 4: Beginning reflection-based field comparison.")
	for i := 0; i < parentVal.NumField(); i++ {
		fieldName := structType.Field(i).Name
		parentField := parentVal.Field(i)
		cloneField := cloneVal.Field(i)

		if fieldName == "objectCacheMu" || fieldName == "turnCtx" {
			checkedFields[fieldName] = true
			continue
		}

		fieldsAreEqual := areFieldsEqual(t, fieldName, parentField, cloneField)
		isIsolated := isolatedFields[fieldName]
		checkedFields[fieldName] = true

		if isIsolated {
			if fieldsAreEqual {
				t.Errorf("Field '%s' should be ISOLATED in the clone, but it was identical to the parent's.", fieldName)
			}
		} else {
			if !fieldsAreEqual {
				t.Errorf("Field '%s' should be SHARED between parent and clone, but it was different.", fieldName)
			}
		}
	}

	for i := 0; i < structType.NumField(); i++ {
		fieldName := structType.Field(i).Name
		if !checkedFields[fieldName] {
			t.Errorf("New/unhandled field '%s' found in Interpreter struct. Please update the 'isolatedFields' map in TestInterpreter_Clone_Integrity to specify if this field should be shared or isolated.", fieldName)
		}
	}
	t.Logf("[DEBUG] Turn 5: Field comparison complete.")
}
