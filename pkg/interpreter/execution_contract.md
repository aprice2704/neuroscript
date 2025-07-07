# Block Execution Contract

from pkg/interpreter/interpreter_steps_blocks.go


 The functions in this file (executeIf, executeWhile, executeFor) are responsible for managing
 control flow constructs. They operate under a specific contract with the main execution
 loop (`recExecuteSteps` in `interpreter_exec.go`):

 1. Block Execution: Each construct is given a block of steps (e.g., the body of a `while`
    loop) to execute. It calls `i.executeBlock()`, which in turn calls `recExecuteSteps` for that block.
 2. Standard Returns: If the block executes without issue, the construct's job is to return the
    result of the last statement, and the execution continues normally.
 3. Error Handling: If `recExecuteSteps` returns a standard runtime error, the construct does not
    handle it. It immediately propagates the error up the call stack.
 4. Control Flow Signals: `break` and `continue` are not standard errors. They are special signals
    implemented as wrapped sentinel errors (`ErrBreak`, `ErrContinue`).
    - The main execution loop (`recExecuteSteps`) has a special case: if it catches an error
      that is a `break` or `continue` signal, it *immediately stops its own execution*
      and returns the signal error up to its caller.
    - The loop constructs in this file (`executeWhile`, `executeFor`) are the designated callers
      that *must* catch these specific signals.
    - Upon catching an `ErrBreak`, the loop must terminate its own execution (e.g., via goto
      or a labeled break) and return `nil` for the error, effectively "consuming" the signal.
    - Upon catching an `ErrContinue`, the loop must skip the rest of the current iteration,
      start the next one, and also return `nil`, consuming the signal.

 This contract ensures that control signals are handled *only* by the nearest enclosing loop
 and are not accidentally caught by general-purpose `on_error` blocks.
