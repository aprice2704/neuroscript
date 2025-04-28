# NeuroScript Error Handling: The `on_error` Mechanism

NeuroScript employs an `on_error` mechanism for structured error handling, designed for clarity and simplicity, avoiding the complexities of traditional `try/catch/finally` blocks. It focuses on handling errors that occur during the execution of function calls (`call`), tool calls (`tool.`), or explicit `fail` statements.

## Core Concepts

1.  **Implicit Error Status**: Every operation that can fail (like `call`, `tool.` calls, `fail`, or potentially built-in errors like division by zero) implicitly returns an error status consisting of an integer `error_code` (0 for success, non-zero for failure) and a string `error_message`.
2.  **Default Propagation**: If an operation fails (non-zero `error_code`) and no `on_error` handler is active, execution of the current procedure stops immediately, and the error status propagates up the call stack. If uncaught, the script terminates. Assignments or subsequent operations in the failing statement/expression do not complete.
3.  **Clean Happy Path**: Code that doesn't need specific error handling remains uncluttered. Errors propagate automatically without explicit checks.

## `on_error` Block

Provides a way to define a scope where errors can be intercepted and handled.

**Syntax:**

```neuroscript
on_error means
    # Handler code block
    # Access error_code and error_msg here
endon
```

**Activation and Scoping:**

* When encountered, an `on_error` block becomes the *active* handler for the current lexical scope (e.g., the current `proc`).
* It replaces any handler previously defined *in the same scope*.
* It only catches errors occurring *after* its definition within its scope.

**Execution Flow:**

* If an error occurs while the handler is active, normal execution pauses.
* The code inside the `on_error ... endon` block executes.
* **Automatic Variables**: Two read-only variables are automatically available within the handler block:
    * `error_code`: (Integer) The error code from the failed operation.
    * `error_msg`: (String) The error message.
    * Attempting to assign to `error_code` or `error_msg` results in a runtime error.

**Behavior Inside the Handler:**

* **Default**: If the handler block finishes normally (reaches `endon`), the *original* error that triggered the handler continues to propagate upwards, as if the handler only logged or observed the error.
* **Errors Within Handler**: If an error occurs *inside* the handler block itself (e.g., via `fail` or a failing `call`) and is not caught by a *nested* `on_error` block:
    * The handler block stops immediately.
    * This *new* error status replaces the original one.
    * The new error propagates upwards from the handler block.
* **`clear_error` Statement**:
    * If `clear_error` is executed within the handler, it marks the *original* triggering error as handled.
    * The rest of the handler block still executes.
    * When `endon` is reached after `clear_error` was called, the original error is suppressed, and execution resumes normally *after* the `endon` statement.

**Restrictions:**

* The `return` statement is **not permitted** inside an `on_error` block. Attempting to use it causes a runtime error. Handlers manage errors, they don't return values for the surrounding function.

## Common Patterns: Retrying Operations

The `on_error` mechanism can be combined with loops (`for`, `while`) to implement patterns like retrying failed operations:

```neuroscript
proc retry_example needs max_attempts returns string means
  result = none
  success = false
  last_code = 0
  last_msg = ""

  on_error means // Handler active for the loop
    log "Handler: Caught error", error_code, error_msg
    is_retryable = (error_code == 503) // Example: Only retry code 503
    if is_retryable {
      last_code = error_code
      last_msg = error_msg
      clear_error // Suppress error for now, let loop decide
    } else {
      // Non-retryable: Let error propagate by default
    }
  endon

  for i from 1 to max_attempts means
    log "Attempt", i
    // Handler is active here
    temp_result = call potentially_failing_op()

    // Check for actual success (assume op returns non-none on success)
    // This block runs if op succeeded, OR if it failed retryably (error cleared)
    if temp_result != none {
       log "Success!"
       result = temp_result
       success = true
       break // Exit loop
    }
    // If we get here, it means op failed retryably and error was cleared
    log "Retryable failure on attempt", i
    // tool.sleep(1) // Optional delay
  endfor

  if not success {
    fail last_code, "Operation failed after " + max_attempts + " attempts. Last: " + last_msg
  }

  return result
endproc

```

This approach uses the handler to classify errors and `clear_error` to allow the loop to continue, while the loop controls the retry logic and the final success/failure determination.