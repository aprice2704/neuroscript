:: type: NSproject
:: subtype: tool_spec
:: version: 0.1.0
:: id: tool-spec-fs-sanitizefilename-v0.1
:: status: draft
:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_fs_utils.go, docs/script_spec.md
:: relatedTo: FS.WriteFile, FS.Mkdir
:: developedBy: AI
:: reviewedBy: User

# Tool Specification Structure Template

## Tool Specification: `FS.SanitizeFilename` (v0.1)

* **Tool Name:** `FS.SanitizeFilename` (v0.1)
* **Purpose:** Takes an input string and cleans it by removing or replacing characters that are typically disallowed or problematic in filenames, making it safer to use as part of a file path.
* **NeuroScript Syntax:** `CALL FS.SanitizeFilename(name: <String>)`
* **Arguments:**
    * `name` (String): Required. The input string that needs to be sanitized for use in a filename.
* **Return Value:** (String)
    * The sanitized string, suitable for use as a filename or path component. (Accessible via `LAST` after the `CALL`). The exact transformations (e.g., which characters are removed or replaced) depend on the underlying Go implementation (`core.SanitizeFilename`).
* **Behavior:**
    1.  Validates that exactly one argument (`name`) of type String is provided.
    2.  Calls the internal Go helper function `core.SanitizeFilename`, passing the input `name` string.
    3.  Returns the sanitized string produced by the `core.SanitizeFilename` function.
* **Security Considerations:**
    * This tool helps prevent the creation of invalid or potentially malicious filenames (e.g., containing path traversal sequences like `../`).
    * It is a utility for string manipulation and does **not** perform sandbox path validation. Tools that perform file operations (like `FS.WriteFile`, `FS.Mkdir`) must still use `SecureFilePath` for proper sandboxing.
    * The effectiveness of the sanitization depends entirely on the rules implemented in the `core.SanitizeFilename` Go function.
* **Examples:**
    ```neuroscript
    # Example 1: Sanitize a string with spaces and slashes
    SET potentially_bad_name = "My Report / Version 2"
    CALL FS.SanitizeFilename(potentially_bad_name)
    SET safe_name = LAST
    EMIT "Original: '" + potentially_bad_name + "' -> Sanitized: '" + safe_name + "'"
    # Example Output might be: 'My_Report_-_Version_2' or similar depending on implementation

    # Example 2: Use sanitized name to create a file
    SET user_input_title = "Data analysis for Q3*?"
    CALL FS.SanitizeFilename(user_input_title)
    SET filename_base = LAST
    SET full_filename = "reports/" + filename_base + ".txt"
    EMIT "Attempting to write to: " + full_filename
    CALL FS.WriteFile(full_filename, "Report content here.")
    SET write_status = LAST
    EMIT "Write status: " + write_status

    # Example 3: Sanitize a simple name (likely no changes)
    SET simple_name = "document1"
    CALL FS.SanitizeFilename(simple_name)
    SET sanitized_simple = LAST
    EMIT "Sanitized simple name: " + sanitized_simple # Expect 'document1'
    ```
* **Go Implementation Notes:**
    * Location: `pkg/core/tools_fs_utils.go`
    * Function: `toolSanitizeFilename`
    * Spec Name: `SanitizeFilename`
    * Key Go Packages: Relies on the internal `core.SanitizeFilename` helper.
    * Helpers: `core.SanitizeFilename` (this function contains the actual sanitization logic).
    * Registration: Registered by `registerFsUtilTools` within `pkg/core/tools_fs.go`. Returns `string, nil`.