# Rules for Command Blocks

Of course. It's wise to review the design rules before diving into the implementation. Here is the summary of the rules for command blocks that we've established.

Command Block Rules
A command block is a top-level construct in NeuroScript designed for imperative, one-off execution of tasks. It is explicitly separated from the library-style func and on definitions to ensure safety and clarity.

Strict Separation: A script file may contain command blocks or func/on definitions, but never both. This is a syntax-level rule enforced by the parser.

No Return Value: command blocks do not produce a return value. The return keyword is a syntax error inside a command block. A command either completes successfully or it fails, raising an error.

No Event Handlers: command blocks cannot contain on event handlers. They are for synchronous execution, not for long-lived, event-driven behavior.

Error Handling: on error blocks are permitted inside a command and are scoped to that command block.

No Function Definitions: command blocks cannot define new functions (func).

Can Call Functions: They are allowed to call pre-existing functions that have been defined in other library scripts.

Sequential Execution: If a script contains multiple command blocks, they are executed in the order they appear in the file.

Non-Empty: A command block (as well as func, if, while, for, on error, and on event blocks) must contain at least one valid statement. A block containing only comments or whitespace is a syntax error.

