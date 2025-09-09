## App Note: Type Integrity for Shape Validation

**Date:** 2025-09-07 (Updated)

### 1. Context: The Type Conversion Problem

When passing data from the NeuroScript runtime (represented as `lang.Value`) to external Go APIs that perform strict data validation (like `pkg/api/shape`), a subtle but critical type conversion issue can occur.

The standard `lang.Unwrap` function, for simplicity, converts all numeric types to `float64`. However, canonical shapes, like the standard `NSEvent`, often require specific integer types (e.g., `int` or `int64` for timestamps). This discrepancy causes validation to fail unexpectedly, even when the data is semantically correct.

### 2. The Canonical Solution: `shape.UnwrapForShapeValidation`

To resolve this permanently, the type-preserving unwrapper has been centralized in the `lang` package and is now exposed as a stable part of the public `shape` API.

**`shape.UnwrapForShapeValidation(v lang.Value) interface{}`**

This function is now the **only correct way** to prepare a `lang.Value` object for validation against a `shape.Shape`. It intelligently inspects a `lang.NumberValue` and converts whole numbers to `int64`, while converting numbers with decimal parts to `float64`, ensuring type integrity for the validator.

### 3. Example Usage

**Incorrect (causes validation failures):**
```go
// DO NOT DO THIS
unwrappedData := lang.Unwrap(myLangValueMap)
err := shape.ValidateSomeShape(unwrappedData.(map[string]interface{}))
// Fails if a shape field expects an int but receives a float64 (e.g., 123.0)
```

Correct (ensures type integrity):

```Go

// THIS IS THE RIGHT WAY
unwrappedData := shape.UnwrapForShapeValidation(myLangValueMap)
err := shape.ValidateSomeShape(unwrappedData.(map[string]interface{}))
// Succeeds because integer values are preserved as int64.
```

4. Integration Rules & Guarantees
Always Use shape.UnwrapForShapeValidation for Validation: When your Go code receives a lang.Value that must be validated against a shape.Shape, you must use the exported shape.UnwrapForShapeValidation function to prepare the data before passing it to the validator.

Interpreter's Guarantee on Events: The interpreter's internal event system (EmitEvent) now uses this same centralized helper function. This guarantees that any event composed by the interpreter will have the correct data types and will successfully pass shape.ValidateNSEvent. This makes event consumption much more reliable for downstream systems like FDM.