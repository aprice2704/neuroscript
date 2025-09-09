# Shape API for Go

This package provides pre-compiled `Shape` objects and convenience functions for composing and validating common NeuroScript data structures within the Go environment. It acts as a stable, public-facing API over the internal `json_lite` package.

The primary benefit of using this package is **efficiency and consistency**. Shapes are defined and parsed only once at program initialization, and the resulting compiled object is reused for all subsequent validation calls.

---

## Using the API

### Pre-compiled Shapes

You can access pre-compiled `*Shape` objects directly if you need more control or want to inspect their structure.

```go
// Get the compiled shape object for a standard NeuroScript Event.
eventShape := shape.NSEvent

// Use it to validate data with specific options.
err := eventShape.Validate(eventData, &shape.ValidateOptions{
    AllowExtra:      true,
    CaseInsensitive: true,
})
```

### Convenience Functions

For common operations, it's easier to use the provided convenience functions.

#### Composition

The `ComposeNSEvent` function creates a valid, standard NeuroScript event object. It internally validates its own output, guaranteeing a correct structure.

```go
package mypackage

import "github.com/aprice2704/neuroscript/pkg/api/shape"

func createEvent() (map[string]interface{}, error) {
    payload := map[string]interface{}{"status": "complete"}
    opts := &shape.NSEventComposeOptions{
        ID: "my-fixed-id",
        AgentID: "my-agent",
    }
    
    return shape.ComposeNSEvent("system.process.done", payload, opts)
}
```

#### Validation

The `ValidateNSEvent` function provides a simple, fast way to check if a map conforms to the canonical event structure.

```go
func processEvent(eventData map[string]interface{}) error {
    // Use the simple convenience function for validation.
    // Pass nil for default options (strict, case-sensitive).
    err := shape.ValidateNSEvent(eventData, nil)
    if err != nil {
        return fmt.Errorf("event data is not a valid ns_event: %w", err)
    }
    // At this point, you can safely access the event's structure.
    fmt.Println("Event is valid!")
    return nil
}
```

---

## How to Add a New Pre-defined Shape

To ensure consistency across the project, new canonical data structures should have their own pre-defined shapes. Follow these steps to add one in `pkg/api/shape/predefined.go`.

#### Step 1: Define the Shape Map

Create a `var` for your new shape definition as a `map[string]interface{}`. This map must follow the **Shape-Lite** specification.

```go
var myNewShapeMap = map[string]interface{}{
    "id": "string",
    "name?": "string",
    "config": map[string]interface{}{ "retries": "int" },
}
```

#### Step 2: Add an Exported Variable

Add an exported, uninitialized variable for your compiled shape.

```go
var (
    NSEvent *Shape
    MyNewObject *Shape // Add your new shape here
)
```

#### Step 3: Parse the Shape in `init()`

In the `init()` function, parse your shape map and assign the result to your exported variable. The `init` function will panic if the shape is invalid, which catches developer errors immediately.

```go
func init() {
    // ... existing NSEvent parsing ...
    MyNewObject, err = ParseShape(myNewShapeMap)
    if err != nil {
        panic(fmt.Sprintf("failed to parse MyNewObject shape: %v", err))
    }
}
```

#### Step 4: Create a Convenience Function

Create an exported convenience function for validation and/or composition.

```go
func ValidateMyNewObject(value map[string]interface{}, options *ValidateOptions) error {
    return MyNewObject.Validate(value, options)
}
```

#### Step 5: Add a Test

Add a test for your new function(s) in `predefined_test.go` to ensure they work as expected.
