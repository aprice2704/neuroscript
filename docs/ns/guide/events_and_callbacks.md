# Host Integration Guide: Handling Event Handler Errors

**Audience:** API Team & Host Application Developers
**Version:** 1.0
**Date:** 2025-09-08

---

## 1. Overview: From Events to Callbacks

Previously, a runtime error within a NeuroScript `on event` handler would cause the interpreter to emit a second, special `system:handler:error` event. This approach was found to be architecturally fragile; an issue in the event-dispatch system could prevent the error from being reported, leading to silent failures.

To create a more robust and reliable error reporting channel, this system has been **completely replaced** by a direct, out-of-band **host callback mechanism**.

---

## 2. The Callback API

When initializing a NeuroScript interpreter, the host application can now provide a function pointer that will be invoked synchronously whenever an event handler fails.

This is configured using the `WithEventHandlerErrorCallback` option:

```go
// From pkg/interpreter/interpreter_options.go

func WithEventHandlerErrorCallback(
    f func(eventName, source string, err *lang.RuntimeError),
) InterpreterOption
```

- **`eventName`**: The name of the event whose handler failed (e.g., `"user:login"`).
- **`source`**: The source string provided when the original event was emitted.
- **`err`**: The `*lang.RuntimeError` containing the error code, message, and position of the failure.

If this callback is not provided, handler errors are simply ignored, and the interpreter will not panic.

---

## 3. Host Responsibilities & Best Practices

The callback function is a critical integration point for monitoring and stability. It is executed synchronously on the same goroutine that called `EmitEvent`.

### ✅ DO:

- **Log Extensively**: This is the primary purpose of the callback. Log the `eventName`, `source`, and all fields of the `RuntimeError` (`Code`, `Message`, `Position`). This provides the essential diagnostic information for debugging faulty scripts.
- **Increment Metrics**: Use the callback to feed monitoring systems. At a minimum, increment a counter for handler errors, ideally with tags for the event name.
  *Example: `metrics.Incr("neuroscript.handler.errors", tags={"event":eventName})`*
- **Keep it Fast**: The callback is a blocking call. It should execute quickly and avoid any long-running operations like network calls or heavy I/O. Defer complex processing to a separate goroutine if necessary.

### ❌ DO NOT:

- **Do Not Panic**: Panicking within the callback will crash the entire host application. The callback should be a safe, resilient boundary.
- **Do Not Re-emit Events**: Do not call `interp.EmitEvent()` from within the callback. This would re-introduce the recursive, unstable pattern that the callback system was designed to replace.
- **Do Not Perform Blocking I/O**: Avoid any operations that could block for an indeterminate amount of time.

---

## 4. Example Implementation

Here is a canonical example of how a host application should configure its interpreter.

```go
package main

import (
    "log/slog"
    "os"

    "[github.com/aprice2704/neuroscript/pkg/interpreter](https://github.com/aprice2704/neuroscript/pkg/interpreter)"
    "[github.com/aprice2704/neuroscript/pkg/lang](https://github.com/aprice2704/neuroscript/pkg/lang)"
)

// myHandlerErrorCallback is a well-behaved implementation of the error handler.
func myHandlerErrorCallback(eventName, source string, err *lang.RuntimeError) {
    // 1. Log with structured details.
    slog.Error("NeuroScript event handler failed",
        "event.name", eventName,
        "event.source", source,
        "error.code", err.Code.String(),
        "error.message", err.Message,
        "error.position", err.Position.String(),
    )

    // 2. Increment a monitoring metric (pseudo-code).
    // my_metrics.Counter("handler_errors_total", "event", eventName).Inc()

    // 3. Optional: Implement a circuit breaker.
    // if myCircuitBreaker.RecordFailure(eventName) > 5 {
    //     slog.Warn("Circuit breaker tripped for event handler", "event.name", eventName)
    //     // Host could choose to deregister the handler here.
    // }
}

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

    // Create the interpreter with the robust error callback.
    interp := interpreter.NewInterpreter(
        interpreter.WithEventHandlerErrorCallback(myHandlerErrorCallback),
        // ... other options like WithLogger ...
    )

    // ... load and run scripts ...
}
```