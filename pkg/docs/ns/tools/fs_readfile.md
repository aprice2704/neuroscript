:: type: NSproject
:: subtype: tool_spec
:: version: 0.1.0
:: id: tool-spec-fs-readfile-v0.1
:: status: draft
:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_fs_read.go
:: relatedTo: FS.WriteFile
:: developedBy: AI
:: reviewedBy: User

# Tool Specification Structure Template

## Tool Specification: FS.ReadFile (v0.1)

* Tool Name: FS.ReadFile (v0.1)
* Purpose: Reads the entire content of a specified file and returns it as a string. Primarily used for accessing text-based files within the allowed workspace.
* NeuroScript Syntax: FS.ReadFile(filepath: <String>)
* Arguments:
    * filepath (String): Required. The relative path to the file within the designated sandbox directory. Absolute paths or paths attempting to traverse outside the sandbox (e.g., ../outside.txt) are forbidden.
* Return Value: (String)
    * On success: The complete content of the file as a string.
    * On error: A string describing the failure (e.g., "ReadFile path error...", "ReadFile failed: File not found...", "ReadFile failed for..."). The underlying Go error is logged internally if a logger is configured but not returned directly to the script.
* Behavior:
    1.  Validates that exactly one argument (filepath) of type String is provided.
    2.  Retrieves the interpreter's configured sandboxDir.
    3.  Uses the SecureFilePath helper to validate the filepath argument against the sandboxDir. This checks:
        * If the path is relative.
        * If the resolved absolute path is still within the sandboxDir.
    4.  If SecureFilePath returns an error (invalid path, outside sandbox), the tool returns an error message string.
    5.  If the path is secure, the tool attempts to read the entire file content using the validated absolute path (os.ReadFile).
    6.  If os.ReadFile encounters an error (e.g., file does not exist, permission denied), the tool returns an appropriate error message string.
    7.  If the file is read successfully, the tool returns its content as a single string.
* Security Considerations:
    * This tool is restricted by the interpreter's sandbox directory (sandboxDir). It cannot read files outside this directory.
    * Relies entirely on the SecureFilePath function for path validation. Any vulnerability in SecureFilePath could potentially allow unauthorized file access.
    * File read permissions are determined by the operating system user running the NeuroScript interpreter.
* Examples:
    neuroscript  # Read a configuration file  config_content = FS.ReadFile("config/settings.json")  if String.Contains(config_content, "ReadFile failed") {  IO.Log("Error reading config:", config_content)  } else {  IO.Log("Config loaded.")  # ... process config_content ...  }   # Attempt to read a file outside the sandbox (will fail)  bad_content = FS.ReadFile("../sensitive_data.txt")  IO.Log("Result of bad read:", bad_content) # Will likely log a path error message 
* Go Implementation Notes:
    * Location: pkg/core/tools_fs_read.go
    * Function: toolReadFile
    * Key Go Packages: os, fmt
    * Helpers: core.SecureFilePath
    * Registration: Called by registerFsFileTools within pkg/core/tools_fs.go (which itself is called by registerCoreTools in pkg/core/tools_register.go).