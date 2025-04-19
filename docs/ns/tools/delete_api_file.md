 :: type: NSproject
 :: subtype: tool_specification
 :: version: 0.1.0
 :: id: tool-spec-delete-api-file-v0.1
 :: status: draft
 :: dependsOn: [docs/ns/tools/tool_spec_structure.md](./tool_spec_structure.md), [pkg/core/tools_file_api.go](../../pkg/core/tools_file_api.go)
 :: howToUpdate: Update if argument validation or return value changes.

 # Tool Specification: TOOL.DeleteAPIFile (v0.1)

 * **Tool Name:** TOOL.DeleteAPIFile (v0.1)
 * **Purpose:** Permanently deletes a specific file from the Gemini File API storage using its unique API resource name.
 * **NeuroScript Syntax:**
   ```neuroscript
   deleteResult := TOOL.DeleteAPIFile(apiFileName)
   ```
 * **Arguments:**
    * `api_file_name` (String): Required. The unique resource name of the file to delete, typically starting with "files/" (e.g., "files/abc123xyz").
 * **Return Value:** (String)
    * On success: A confirmation string (e.g., "Successfully deleted API file: files/abc123xyz").
    * On failure: An error string describing the failure (e.g., "Failed to delete API file 'files/...' : rpc error: code = NotFound desc = File not found"). A Go-level error will also be returned to the interpreter.
 * **Behavior:**
    1. Checks if the internal GenAI client is initialized. Returns an error if not.
    2. Validates that the `api_file_name` argument is a non-empty string.
    3. Calls the Gemini File API's `DeleteFile` endpoint with the provided `api_file_name`.
    4. If the API call succeeds, returns a success message string.
    5. If the API call fails (e.g., file not found, permission denied), returns an error message string and propagates the Go error.
 * **Security Considerations:**
    * Requires a valid `GEMINI_API_KEY` with permission to delete files.
    * Deletion is permanent and cannot be undone.
    * Network access to Google Cloud services is required.
 * **Examples:**
   ```neuroscript
   // Assume apiNameToDelete holds "files/some-generated-id"
   apiNameToDelete := "files/abc123xyz"
   result := TOOL.DeleteAPIFile(apiNameToDelete)
   IO.Print("Deletion Result:", result)

   // Example: Delete all files listed by ListAPIFiles
   allFiles := TOOL.ListAPIFiles()
   FOR fileInfo IN allFiles DO
       IF fileInfo["error"] == null AND fileInfo["name"] != null THEN
           IO.Print("Attempting to delete:", fileInfo["name"])
           deleteMsg := TOOL.DeleteAPIFile(fileInfo["name"])
           IO.Print("  Result:", deleteMsg)
       END
   END
   ```
 * **Go Implementation Notes:**
    * Implemented in `pkg/core/tools_file_api.go` as `toolDeleteAPIFile`.
    * Uses `interpreter.GenAIClient().DeleteFile()`.
    * Registered via `registerFileAPITools`.