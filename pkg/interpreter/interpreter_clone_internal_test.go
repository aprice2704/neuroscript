// NeuroScript Version: 0.7.2
// File version: 8
// Purpose: Adds an explicit check to ensure the sandboxDir is correctly propagated to the state of a cloned interpreter.
// filename: pkg/interpreter/interpreter_clone_internal_test.go

package interpreter

import (
	"reflect"
	"testing"
)

// areFieldsEqual checks if two reflected fields are equal.
// It safely handles unexported fields, pointers, interfaces, and value types.
func areFieldsEqual(t *testing.T, fieldName string, v1, v2 reflect.Value) bool {
	t.Helper()

	// If a field is an interface, we get the underlying element to check its pointer.
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

	// For types that are inherently pointers, compare their memory address.
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

	// For all other types (value types like bool, int, string, struct),
	// we must check if the field is exportable before comparing.
	if !v1.CanInterface() {
		// This is an unexported field. We cannot use DeepEqual.
		// We must compare it based on its kind.
		switch v1.Kind() {
		case reflect.Bool:
			return v1.Bool() == v2.Bool()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return v1.Int() == v2.Int()
		case reflect.String:
			return v1.String() == v2.String()
		// Add other primitive kinds here if needed in the future.
		default:
			// For unexported structs or other complex types, we cannot safely compare them.
			// The logic in the main test relies on this function returning 'false' for
			// isolated fields (which are often unexported structs like 'tools'), so this is safe.
			return false
		}
	}

	// The field is exported, so we can safely use DeepEqual.
	return reflect.DeepEqual(v1.Interface(), v2.Interface())
}

// TestInterpreter_Clone_Integrity uses reflection to ensure that the clone() method
// correctly handles every field in the Interpreter struct.
func TestInterpreter_Clone_Integrity(t *testing.T) {
	parent, err := NewTestInterpreter(t, nil, nil, true)
	if err != nil {
		t.Fatalf("Failed to create parent interpreter: %v", err)
	}
	// Set a specific value to test for propagation.
	parent.state.sandboxDir = "/test/sandbox/path"

	clone := parent.clone()

	// --- EXPLICIT CHECK FOR SANDBOX PROPAGATION ---
	// This is the critical check that was missing.
	if clone.state.sandboxDir != parent.state.sandboxDir {
		t.Errorf("Sandbox path was not propagated to clone. Parent: '%s', Clone: '%s'",
			parent.state.sandboxDir, clone.state.sandboxDir)
	}
	// --- END EXPLICIT CHECK ---

	// Fields that are EXPECTED to be different (new instances) in the clone.
	isolatedFields := map[string]bool{
		"id":              true, // A clone gets its own unique ID.
		"state":           true, // The clone gets a new, isolated state.
		"tools":           true, // The clone gets a new view of the tool registry.
		"evaluate":        true, // The clone gets its own evaluation engine.
		"root":            true, // A clone's root points to its parent, so it differs from the parent's root (which is nil).
		"cloneRegistry":   true, // The clone registry only exists on the root. Clones have a new, empty one.
		"cloneRegistryMu": true, // Each mutex is a distinct instance.
	}

	parentVal := reflect.ValueOf(parent).Elem()
	cloneVal := reflect.ValueOf(clone).Elem()
	structType := parentVal.Type()
	checkedFields := make(map[string]bool)

	for i := 0; i < parentVal.NumField(); i++ {
		fieldName := structType.Field(i).Name
		parentField := parentVal.Field(i)
		cloneField := cloneVal.Field(i)

		// The mutex and context are not directly comparable by pointer.
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

	// Sanity check to ensure the test itself is up-to-date.
	for i := 0; i < structType.NumField(); i++ {
		fieldName := structType.Field(i).Name
		if !checkedFields[fieldName] {
			t.Errorf("New/unhandled field '%s' found in Interpreter struct. Please update the 'isolatedFields' map in TestInterpreter_Clone_Integrity to specify if this field should be shared or isolated.", fieldName)
		}
	}
}
