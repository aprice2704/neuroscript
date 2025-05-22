 # NeuroScript Tool Return and Error Conventions
 
 **Version:** 1.0.0
 **Date:** May 21, 2025
 
 ### 1. Introduction
 
 This document outlines the established conventions for how tools in NeuroScript report success, return data, and signal errors. Understanding these conventions is crucial for writing robust NeuroScript procedures and for AI systems that generate or execute NeuroScript code.
 
 The primary error handling mechanism relies on standard Go error returns at the tool's implementation level, which are then processed by the NeuroScript interpreter.
 
 ### 2. Core Go `ToolFunc` Signature
 
 All Go functions that implement NeuroScript tools adhere to the following signature, as defined in `pkg/core/tools_types.go`:
 
 ```go
 type ToolFunc func(interpreter *Interpreter, args []interface{}) (interface{}, error)
 ```
 
 * **`interface{}` (ToolResult):** This is the primary data or status returned to the NeuroScript environment if the tool's Go function executes without an exceptional error.
 * **`error` (Go Error):** This is a standard Go error.
     * A **non-`nil` `error`** signals an exceptional failure during the tool's execution at the Go level.
     * A **`nil` `error`** indicates that the tool's Go function completed its intended operation without encountering an exceptional Go-level error. The `interface{}` result is then considered valid.
 
 ### 3. NeuroScript Interpreter Handling
 
 The NeuroScript interpreter processes the `(interface{}, error)` pair returned by a tool's Go function:
 
 * **If the Go `error` is non-`nil`:**
     * The interpreter catches this error and typically converts it into a `core.RuntimeError`.
     * This `RuntimeError` triggers the script's error handling mechanisms:
         * It may cause an immediate halt of the script.
         * It can be caught by an `on_error ... endon` block within the current NeuroScript procedure.
     * When such a Go-level error occurs, the `interface{}` (ToolResult) from the Go function is generally disregarded. From the NeuroScript's perspective, a variable assigned the tool's result (e.g., via `set my_var = tool.Example()`) will likely receive `nil`, or the `last` keyword will reflect `nil` or a previous valid result. The script should *not* expect a descriptive error string *as the direct return value* of the tool in this case; the error is handled by the interpreter's state.
 * **If the Go `error` is `nil`:**
     * The interpreter considers the tool's Go-level execution successful.
     * The `interface{}` (ToolResult) is passed to the NeuroScript environment. This is the value assigned to variables or accessible via `last`.
 
 ### 4. Observed Patterns & NeuroScript-Level Handling
 
 Based on an analysis of tool implementations (e.g., in `pkg/core/tools_fs_read.go`, `pkg/core/tools_fs_write.go`, `pkg/core/tools_git.go`, `pkg/core/ai_wm_tools_execution.go`) and their definitions (e.g., `pkg/core/tooldefs_fs.go`):
 
 * **Tools Returning Data:**
     * Example: `FS.Read`.
     * Convention: On success (Go `error` is `nil`), they return the actual data (e.g., file content as a string) as the `interface{}`.
     * NeuroScript Handling:
         ```neuroscript
         set file_content = tool.FS.Read("my_file.txt")
         if file_content == nil
           emit "Error: Failed to read file or file is empty."
           # Potentially fail or handle error
         else
           # Process file_content
         endif
         # OR, more robustly if an error halts or goes to on_error:
         # on_error means
         #   emit "FS.Read failed: " + system.error_message
         # endon
         # set file_content = tool.FS.Read("my_file.txt")
         # must file_content != nil # If nil implies an error that didn't halt.
         ```
 * **Tools Primarily for Side-Effects Returning Success Messages:**
     * Examples: `FS.Write`, `Git.Add`, `Git.Commit`.
     * Convention: On success (Go `error` is `nil`), these tools return a descriptive string message (e.g., "Successfully wrote X bytes...") as the `interface{}`. Their `ToolSpec.ReturnType` is typically `ArgTypeString`.
     * NeuroScript Handling:
         ```neuroscript
         set write_result = tool.FS.Write("output.txt", "hello")
         # If an actual Go error occurred, an on_error block would likely trigger,
         # or the script would halt. 'write_result' might be nil then.
         # If no error state, 'write_result' holds the success message.
         emit "FS.Write operation status: " + write_result
         ```
         Checking for a *specific* string like "OK" is fragile unless the tool's `ToolSpec` description *guarantees* that exact string. It's safer to assume any non-`nil` string result (in the absence of a script error/halt) indicates success for these tools.
 * **Tools Returning Structured Status/Data Maps:**
     * Example: `AIWorker.ExecuteStatelessTask` (the Go tool function wrapper).
     * Convention: On success (Go `error` is `nil`), returns a map containing various pieces of information (e.g., `{"output": ..., "taskId": ..., "cost": ...}`). The `ToolSpec.ReturnType` is `ArgTypeMap`.
     * NeuroScript Handling:
         ```neuroscript
         set ai_map_result = tool.AIWorker.ExecuteStatelessTask(...)
         if ai_map_result == nil
           emit "Error: AIWorker.ExecuteStatelessTask tool call failed at interpreter level."
         else
           # Process ai_map_result, e.g., ai_map_result["output"]
           # Note: This specific tool's Go wrapper doesn't add an "error" key to this map.
           # Errors from the underlying AI service call result in the Go wrapper itself
           # returning a non-nil Go error, which the interpreter handles.
         endif
         ```
 * **Tools with Side-Effects and No Specific Return Value (Hypothetical/Future):**
     * Convention: If a tool has only side-effects and no meaningful data/message to return on success, its `ToolSpec.ReturnType` would ideally be `ArgTypeNil`. The Go `ToolFunc` would return `(nil, nil)`.
     * NeuroScript Handling:
         ```neuroscript
         call tool.SilentSideEffectTool()
         # Success is implied if execution continues and no on_error block is triggered.
         # There's no meaningful direct return value to check.
         ```
 
 ### 5. Recommendations for NeuroScript Authors (Human & AI)
 
 1.  **Primary Error Detection:** Rely on NeuroScript's `on_error ... endon` blocks to catch and handle exceptional failures originating from tools (which are propagated as `RuntimeError`s by the interpreter). If an error is critical and unrecoverable by the script, allow the script to halt.
 2.  **Checking Tool Results:**
     * **Data-Returning Tools:** After a `set var = tool.GetData()`, check if `var` is `nil`. A `nil` value often means the underlying Go tool function returned an error, which the interpreter handled, resulting in `nil` being passed to the script. Then, validate the data if non-`nil`.
     * **Message-Returning Tools:** For tools like `FS.Write` or `Git.Commit` that return a success message string, you can `emit` this string for logging. Its presence (non-`nil` string) after a call, without an `on_error` trigger, implies success. Avoid checking for an exact string like "OK" unless the `ToolSpec.Description` for that specific tool explicitly guarantees it.
     * **Map-Returning Tools:** If a tool returns a map, check for `nil` first (indicating a Go-level error during the tool call). If non-`nil`, inspect the documented keys within the map for specific results or operational status.
     * **`must` Statements:** Use `must` statements to assert expected conditions *after* a tool call if its direct return value isn't a simple status (e.g., `must file_exists_now == true` after a `tool.FS.Write` if another tool `tool.FS.FileExists` were available).
 3.  **Consult Tool Specifications:** When available, `ToolSpec.Description` and `ToolSpec.ReturnType` (often found in `tooldefs_*.go` files) provide the most reliable information on what a tool is expected to return in NeuroScript on successful execution.
 
 ### 6. Future Consistency
 
 As discussed, future work may involve making the return patterns of side-effect tools more consistent (e.g., all returning `nil` via `ArgTypeNil` for success, or all returning a standardized status map). The guidelines above reflect the *current observed conventions*.
