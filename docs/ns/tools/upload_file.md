 :: type: NSproject
 :: subtype: tool_specification
 :: version: 0.1.0
 :: id: tool-spec-upload-file-v0.1
 :: status: draft
 :: dependsOn: [docs/ns/tools/tool_spec_structure.md](./tool_spec_structure.md), [pkg/core/tools_file_api.go](../../pkg/core/tools_file_api.go)
 :: howToUpdate: Update if arguments, return value structure, or default display name logic changes.

 # Tool Specification: TOOL.UploadFile (v0.1)

 * **Tool Name:** TOOL.UploadFile (v0.1)
 * **Purpose:** Uploads a local file from the NeuroScript sandbox environment to the Gemini File API storage. It waits for the file to become ACTIVE and ready for use in prompts.
 * **NeuroScript Syntax:**
   ```neuroscript
   uploadResultMap := TOOL.UploadFile(localFilePath, [optionalDisplayName])
   ```
 * **Arguments:**
    * `local_path` (String): Required. The path to the file within the NeuroScript sandbox environment that needs to be uploaded. Subject to `SecureFilePath` validation.
    * `display_name` (String): Optional. A user-friendly name to assign to the file within the API. If omitted or null, defaults to the relative path of the file from the sandbox root (if a sandbox is set) or the file's base name otherwise.
 * **Return Value:** (Map)
    * On success: A map containing the details of the successfully uploaded and activated file. The structure matches the maps returned by `TOOL.ListAPIFiles`, including keys like `name`, `displayName`, `uri`, `mimeType`, `sizeBytes`, `state` (which will be "ACTIVE"), etc.
    * On failure: An error (e.g., local file not found, upload API error, timeout waiting for ACTIVE state, security violation). A Go-level error will also be returned to the interpreter.
 * **Behavior:**
    1. Checks if the internal GenAI client is initialized. Returns an error if not.
    2. Validates the `local_path` argument (must be a non-empty string) and performs `SecureFilePath` check against the interpreter's sandbox directory.
    3. Determines the `displayName` to use (provided argument or default based on path).
    4. Checks if the local file exists and gets its size. Handles zero-byte files by uploading minimal placeholder content (" ").
    5. Determines the MIME type based on the file extension, defaulting to `application/octet-stream`.
    6. Calls the Gemini File API's `UploadFile` endpoint with the file content and determined options.
    7. Polls the API using `GetFile` until the uploaded file's state becomes `ACTIVE`.
    8. If the file becomes `ACTIVE` within the timeout period, returns a map containing the final file metadata.
    9. If the file fails to become `ACTIVE` (e.g., enters a `FAILED` state or times out), attempts to delete the failed upload from the API and returns an error.
 * **Security Considerations:**
    * Requires a valid `GEMINI_API_KEY` with permission to upload files.
    * `local_path` is validated using `SecureFilePath` to ensure it's within the allowed sandbox directory.
    * Network access to Google Cloud services is required.
    * The content of the uploaded file is transferred to Google Cloud storage.
 * **Examples:**
   ```neuroscript
   // Upload a file using default display name
   result1 := TOOL.UploadFile("data/input.txt")
   IF result1["error"] != null THEN
       IO.Print("Error uploading input.txt:", result1["error"])
   ELSE
       IO.Print("Uploaded input.txt successfully. API Name:", result1["name"])
       // Can now potentially use result1["uri"] in prompts
   END

   // Upload a file with a specific display name
   result2 := TOOL.UploadFile("images/logo.png", "ProjectLogo")
   IF result2["error"] != null THEN
       IO.Print("Error uploading logo.png:", result2["error"])
   ELSE
       IO.Print("Uploaded logo.png successfully. API Name:", result2["name"], ", URI:", result2["uri"])
   END
   ```
 * **Go Implementation Notes:**
    * Implemented in `pkg/core/tools_file_api.go` as `toolUploadFile`.
    * Uses `interpreter.GenAIClient().UploadFile()` and `interpreter.GenAIClient().GetFile()`.
    * Logic adapted from `cmd/gensync/helpers.go`.
    * Requires careful handling of file reading, zero-byte files, MIME types, and polling logic.
    * Registered via `registerFileAPITools`.