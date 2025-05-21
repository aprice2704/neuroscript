 :: type: NSproject
 :: subtype: tool_specification
 :: version: 0.1.0
 :: id: tool-spec-list-api-files-v0.1
 :: status: draft
 :: dependsOn: [docs/ns/tools/tool_spec_structure.md](./tool_spec_structure.md), [pkg/core/tools_file_api.go](../../pkg/core/tools_file_api.go)
 :: howToUpdate: Update if return value structure changes or pagination/filtering is added.

 # Tool Specification: TOOL.ListAPIFiles (v0.1)

 * **Tool Name:** TOOL.ListAPIFiles (v0.1)
 * **Purpose:** Retrieves a list of files previously uploaded to the Gemini File API that are associated with the currently configured API key.
 * **NeuroScript Syntax:**
   ```neuroscript
   apiFilesList := TOOL.ListAPIFiles()
   ```
 * **Arguments:**
    * None.
 * **Return Value:** (List)
    * On success: A list where each element is a Map representing a file stored in the API. Each map contains keys like:
        * `name` (String): The unique API resource name (e.g., "files/abc123xyz").
        * `displayName` (String): The user-provided or default display name.
        * `mimeType` (String): The detected MIME type.
        * `sizeBytes` (Number): The size of the file in bytes.
        * `createTime` (String): Creation timestamp (RFC3339 format).
        * `updateTime` (String): Last update timestamp (RFC3339 format).
        * `state` (String): Current state (e.g., "ACTIVE", "PROCESSING", "FAILED").
        * `uri` (String): The `aipi://` URI for referencing the file in prompts (available when ACTIVE).
        * `sha256Hash` (String): Hex-encoded SHA256 hash of the file content.
        * `videoMetadata` (Map|null): Video-specific metadata if applicable, otherwise null.
    * On failure: An error (e.g., API connection issues, authentication errors). If listing fails mid-way, the returned list might contain partial results along with a final map containing an `error` key describing the failure. A Go-level error will also be returned to the interpreter.
 * **Behavior:**
    1. Checks if the internal GenAI client is initialized. Returns an error if not.
    2. Calls the Gemini File API's `ListFiles` endpoint.
    3. Iterates through all available files associated with the API key.
    4. For each file, creates a map containing its metadata.
    5. Returns the list of file metadata maps.
    6. If an error occurs during iteration, returns the partial list collected so far and propagates the error.
 * **Security Considerations:**
    * Requires a valid `GEMINI_API_KEY` to be configured in the environment where the NeuroScript interpreter runs.
    * Network access to Google Cloud services is required.
    * The listed files might contain sensitive information depending on what was previously uploaded.
 * **Examples:**
   ```neuroscript
   // Get all files currently in the API storage
   allFiles := TOOL.ListAPIFiles()
   IO.Print("Files found in API:")
   FOR fileInfo IN allFiles DO
       IF fileInfo["error"] != null THEN
           IO.Print("  Error encountered during listing:", fileInfo["error"])
           BREAK // Stop processing list if error occurred
       END
       IO.Print("  - Name:", fileInfo["name"], ", Display:", fileInfo["displayName"], ", Size:", fileInfo["sizeBytes"])
   END
   ```
 * **Go Implementation Notes:**
    * Implemented in `pkg/core/tools_file_api.go` as `toolListAPIFiles`.
    * Uses `interpreter.GenAIClient().ListFiles()`.
    * Requires handling pagination using `google.golang.org/api/iterator`.
    * Registered via `registerFileAPITools`.