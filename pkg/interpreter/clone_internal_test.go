// NeuroScript Version: 0.8.0
// File version: 14
// Purpose: Corrects the test to treat the 'tools' field as isolated. This is now the correct behavior, as a forked interpreter gets a new "view" of the tool registry bound to its own runtime.
// filename: pkg/interpreter/clone_internal_test.go

package interpreter

import (
	"fmt"
	"reflect"
	"testing"
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
	parent, err := NewTestInterpreter(t, nil, nil, true)
	if err != nil {
		t.Fatalf("Failed to create parent interpreter: %v", err)
	}
	parent.state.sandboxDir = "/test/sandbox/path"
	t.Logf("[DEBUG] Turn 2: Parent interpreter created.")

	clone := parent.Clone()
	t.Logf("[DEBUG] Turn 3: Interpreter cloned.")

	fmt.Printf("[DEBUG] Parent tools pointer: %p, Clone tools pointer: %p\n", parent.tools, clone.tools)

	if clone.state.sandboxDir != parent.state.sandboxDir {
		t.Errorf("Sandbox path was not propagated to clone. Parent: '%s', Clone: '%s'",
			parent.state.sandboxDir, clone.state.sandboxDir)
	}

	isolatedFields := map[string]bool{
		"id":    true,
		"state": true,
		// THE FIX: The 'tools' registry object is now intentionally isolated in a clone
		// to ensure it's bound to the correct runtime. It gets a new "view".
		"tools":           true,
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

		// The 'evaluate' field no longer exists.
		if fieldName == "evaluate" {
			checkedFields[fieldName] = true
			continue
		}

		// The turnCtx field's propagation is checked by functional tests.
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
