:: type: NSproject
:: subtype: tool_spec
:: version: 0.1.0
:: id: tool-spec-git-gitadd-v0.1
:: status: draft
:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_git_register.go, pkg/core/tools_git.go, docs/script_spec.md
:: relatedTo: Git.GitCommit, Git.GitStatus
:: developedBy: AI
:: reviewedBy: User

# Tool Specification Structure Template

## Tool Specification: `Git.GitAdd` (v0.1)

* **Tool Name:** `Git.GitAdd` (v0.1)
* **Purpose:** Stages changes in one or more specified files or directories for the next Git commit. This is equivalent to running `git add <path1> <path2> ...` within the project's sandbox directory.
* **NeuroScript Syntax:** `CALL Git.GitAdd(paths: <List[String]>)`
* **Arguments:**
    * `paths` (List of Strings): Required. A list containing one or more relative paths (strings) within the sandbox. These paths specify the files or directories whose changes should be staged.
* **Return Value:** (String)
    * On success: A string indicating success, potentially including the output from the `git add` command. Example: `"GitAdd successful for paths: [file1.txt dir/file2.go].\nOutput:\n"` (Output may be empty if successful). (Accessible via `LAST` after the `CALL`).
    * On error: A string describing the failure. This could be due to path validation errors (e.g., "GitAdd path error for '...': ...") or errors from the underlying `git add` command execution (e.g., "GitAdd failed: ..."). (Accessible via `LAST` after the `CALL`).
* **Behavior:**
    1.  Validates that exactly one argument (`paths`) is provided and that it is a List containing only Strings.
    2.  Retrieves the interpreter's `sandboxDir`.
    3.  Iterates through each path string in the input `paths` list.
    4.  For each path, uses `SecureFilePath` to validate it against the `sandboxDir`. If any path fails validation, returns an error message string immediately.
    5.  If all paths are validated, constructs a command equivalent to `git add <path1> <path2> ...` using the validated relative paths.
    6.  Executes the `git add` command within the `sandboxDir` using the `toolExec` helper function.
    7.  If the command execution fails (e.g., `git` not found, `git add` returns an error), returns an error message string containing details from `toolExec`.
    8.  If the command execution succeeds, returns a success message string, including any standard output from the `git add` command.
* **Security Considerations:**
    * Uses `SecureFilePath` to ensure all specified paths are within the configured sandbox before passing them to the `git` command.
    * Relies on the `toolExec` helper for command execution, which should execute commands within the `sandboxDir`.
    * Modifies the Git index (staging area) of the repository located within the `sandboxDir`.
    * Assumes the `git` executable is available in the environment's PATH where the NeuroScript interpreter is running.
* **Examples:**
    ```neuroscript
    # Example 1: Stage a single modified file
    CALL FS.WriteFile("README.md", "Updated content.") # Modify a file
    SET write_ok = LAST

    CALL Git.GitAdd(["README.md"])
    SET add_status = LAST
    EMIT "Git Add status: " + add_status

    # Example 2: Stage multiple files/directories
    CALL FS.WriteFile("src/main.go", "// New Go code")
    CALL FS.WriteFile("docs/guide.md", "New documentation section.")
    SET write_multi_ok = LAST

    SET files_to_add = ["src/main.go", "docs/"] # Add a file and a directory
    CALL Git.GitAdd(files_to_add)
    SET add_multi_status = LAST
    EMIT "Git Add multiple status: " + add_multi_status

    # Example 3: Attempt to add a file outside the sandbox (will fail validation)
    CALL Git.GitAdd(["../sensitive_info.txt"])
    SET add_bad_path_status = LAST
    EMIT "Git Add bad path status: " + add_bad_path_status
    # Expect error like: "GitAdd path error for '../sensitive_info.txt': Path is outside sandbox"
    ```
* **Go Implementation Notes:**
    * Location: `pkg/core/tools_git.go`
    * Function: `toolGitAdd`
    * Spec Name: `GitAdd` (in `pkg/core/tools_git_register.go`)
    * Key Go Packages: `fmt`, `strings`, `errors`
    * Helpers: `core.SecureFilePath`, `core.toolExec`
    * Registration: Registered by `registerGitTools` within `pkg/core/tools_register.go`. Returns `string, error`.