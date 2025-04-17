# Tool Specification Structure Template

This document outlines the standard structure for documenting NeuroScript built-in tools within the `docs/ns/tools/` directory. Each tool should have its own markdown file (e.g., `io_input.md`, `tool_fs_movefile.md`).

## Tool Specification: `CATEGORY.ToolName` (Vx.y)

* **Tool Name:** The fully qualified name of the tool (e.g., `IO.Input`, `TOOL.MoveFile`). Include a version number (e.g., v0.1) if the spec evolves.
* **Purpose:** A brief (1-2 sentence) description of what the tool does and why it's useful.
* **NeuroScript Syntax:** A code example showing how the tool is called in NeuroScript. Include placeholders for arguments.
* **Arguments:**
    * A list describing each argument:
        * `argument_name` (Type): Description of the argument, including whether it's required or optional, and any constraints (e.g., specific format, allowed values).
* **Return Value:** (Type - e.g., Map, List, String, Number, Boolean, null)
    * Description of the value returned by the tool upon successful execution.
    * If the return type is a Map, list the expected keys and their value types/meanings. Crucially, specify how errors are indicated (e.g., an `error` key being non-null).
* **Behavior:**
    * A numbered or bulleted list describing the step-by-step execution logic of the tool.
    * Detail validation performed on arguments.
    * Specify handling of edge cases (e.g., file not found, empty input, division by zero).
    * Describe how errors are generated and returned.
* **Security Considerations:**
    * Outline any security implications of using the tool.
    * Specify restrictions (e.g., disallowed in agent mode, reliance on `SecureFilePath`).
    * Mention potential risks if used improperly.
* **Examples:** (Optional but Recommended)
    * One or more NeuroScript snippets demonstrating practical usage of the tool, potentially showing different argument combinations or expected outputs.
* **Go Implementation Notes:** (Optional - Primarily for developers implementing the tool)
    * Suggested location in the Go codebase (e.g., `pkg/core/tools_fs.go`).
    * Relevant Go packages to use (e.g., `os`, `fmt`, `go/ast`).
    * Key Go functions involved (e.g., `os.Rename`).
    * Reminder about registration (e.g., in `pkg/core/tools_register.go`).