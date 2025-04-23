:: type: NSproject
:: subtype: tool_spec
:: version: 0.1.0
:: id: tool-spec-fs-linecountfile-v0.1
:: status: draft
:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_fs_utils.go, docs/script_spec.md
:: relatedTo: FS.ReadFile, FS.ListDirectory
:: developedBy: AI
:: reviewedBy: User

# Tool Specification Structure Template

## Tool Specification: `FS.LineCountFile` (v0.1)

* **Tool Name:** `FS.LineCountFile` (v0.1)
* **Purpose:** Counts the number of lines in a specified text file within the sandbox.
* **NeuroScript Syntax:** `CALL FS.LineCountFile(filepath: <String>)`
* **Arguments:**
    * `filepath` (String): Required. The relative path to the file within the sandbox whose lines should be counted.
* **Return Value:** (Number)
    * On success: An integer representing the number of lines in the file. (Accessible via `LAST` after the `CALL`).
        * An empty file returns `0`.
        * Lines are typically delimited by the newline character (`\n`).
        * If the file has content but does not end with a newline, the last line is still counted.
        * A file containing only a single newline character returns `1`.
    * On error (e.g., file not found, path validation fails, read error): Returns the integer `-1`. (Accessible via `LAST` after the `CALL`).
* **Behavior:**
    1.  Validates that exactly one argument (`filepath`) of type String is provided.
    2.  Retrieves the interpreter's configured `sandboxDir`.
    3.  Uses the `SecureFilePath` helper to validate the `filepath` argument against the `sandboxDir`. If validation fails, returns `-1`.
    4.  Attempts to read the entire content of the specified file using the validated absolute path (`os.ReadFile`).
    5.  If reading the file fails (e.g., not found, permission denied), returns `-1`.
    6.  If the file content is empty, returns `0`.
    7.  Counts the occurrences of the newline character (`\n`) in the content.
    8.  If the content is not empty and does *not* end with a newline character, increments the count by 1.
    9.  Handles the edge case where the file content is exactly `\n`, setting the count to `1`.
    10. Returns the final calculated line count as an integer (Number).
* **Security Considerations:**
    * Restricted by the interpreter's sandbox directory (`sandboxDir`) via `SecureFilePath`. Cannot access files outside the sandbox.
    * Reads file content into memory to perform the count. Very large files could potentially consume significant memory during processing, although the counting itself is generally efficient.
* **Examples:**
    ```neuroscript
    # Example 1: Count lines in a log file
    CALL FS.WriteFile("activity.log", "User logged in.\nPerformed action A.\nUser logged out.") # Create a 3-line file
    SET write_ok = LAST

    CALL FS.LineCountFile("activity.log")
    SET line_count = LAST

    IF line_count >= 0 THEN
        EMIT "Log file 'activity.log' has " + line_count + " lines." # Expect 3
    ELSE
        EMIT "Error counting lines in 'activity.log'."
    ENDBLOCK

    # Example 2: Count lines in a non-existent file
    CALL FS.LineCountFile("non_existent_file.txt")
    SET count_error = LAST
    EMIT "Line count for non-existent file: " + count_error # Expect -1

    # Example 3: Count lines in an empty file
    CALL FS.WriteFile("empty.txt", "")
    SET empty_write_ok = LAST
    CALL FS.LineCountFile("empty.txt")
    SET empty_count = LAST
    EMIT "Line count for empty file: " + empty_count # Expect 0
    ```
* **Go Implementation Notes:**
    * Location: `pkg/core/tools_fs_utils.go`
    * Function: `toolLineCountFile`
    * Spec Name: `LineCountFile`
    * Key Go Packages: `os`, `fmt`, `strings`
    * Helpers: `core.SecureFilePath`
    * Registration: Registered by `registerFsUtilTools` within `pkg/core/tools_fs.go`. Returns `int64, error` internally, but handles errors to return the count or -1.