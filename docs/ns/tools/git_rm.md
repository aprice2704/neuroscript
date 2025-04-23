:: type: NSproject
:: subtype: tool_spec
:: version: 0.1.0
:: id: tool-spec-git-gitrm-v0.1
:: status: draft
:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_git_register.go, pkg/core/tools_git.go, docs/script_spec.md
:: relatedTo: Git.GitAdd, Git.GitCommit, FS.DeleteFile
:: developedBy: AI
:: reviewedBy: User

# Tool Specification Structure Template

## Tool Specification: `Git.GitRm` (v0.1)

* **Tool Name:** `Git.GitRm` (v0.1)
* **Purpose:** Removes a specified file from the Git index (staging area) and typically also removes it from the working directory. This is equivalent to running `git rm <path>` within the project's sandbox directory. It does *not* remove directories recursively (use `git rm -r` functionality is not exposed by this tool).
* **NeuroScript Syntax:** `CALL Git.GitRm(path: <String>)`
* **Arguments:**
    * `path` (String): Required. The relative path (within the sandbox) of the file to be removed from the Git index and working directory.
* **Return Value:** (String)
    * On success: A string indicating the file was successfully removed from the index, potentially including output from the `git rm` command. Example: `"Successfully removed path 'config.old' from git index.\nOutput:\nrm 'config.old'\n"`. (Accessible via `LAST` after the `CALL`).
    * On error: A string describing the failure. This could be due to path validation errors (e.g., "invalid path for GitRm '...'") or errors from the underlying `git rm` command execution (e.g., "failed to remove path '...': fatal: pathspec '...' did not match any files"). (Accessible via `LAST` after the `CALL`).
* **Behavior:**
    1.  Validates that exactly one argument (`path`) of type String is provided.
    2.  Retrieves the interpreter's `sandboxDir`.
    3.  Uses `SecureFilePath` to validate the `path` argument against the `sandboxDir`. If validation fails, returns an error message string.
    4.  Constructs a command equivalent to `git rm "<path>"`.
    5.  Executes the `git rm` command within the `sandboxDir` using the `toolExec` helper function.
    6.  If the command execution fails (e.g., `git` not found, path does not match a file in the index), returns an error message string containing details from `toolExec`.
    7.  If the command execution succeeds, returns a success message string, including the path specified and any standard output from the `git rm` command.
* **Security Considerations:**
    * Executes the `git rm` command within the configured sandbox directory.
    * Modifies the Git index (staging area) and typically deletes the file from the working directory. Deletions from the working directory are generally permanent unless recovered through Git history.
    * Uses `SecureFilePath` to ensure the path specified is within the sandbox before passing it to `git rm`.
    * Relies on the `toolExec` helper for command execution.
    * Assumes the `git` executable is available in the environment's PATH where the NeuroScript interpreter is running.
* **Examples:**
    ```neuroscript
    # Example 1: Create, add, commit, then remove a file
    CALL FS.WriteFile("obsolete_file.txt", "This file will be removed.")
    CALL Git.GitAdd(["obsolete_file.txt"])
    CALL Git.GitCommit("Add obsolete file for removal test")

    EMIT "Removing file 'obsolete_file.txt' from Git..."
    CALL Git.GitRm("obsolete_file.txt")
    SET rm_status = LAST
    EMIT "Git Rm status: " + rm_status
    # Need to commit the removal
    CALL Git.GitCommit("Remove obsolete_file.txt")
    EMIT "Commit removal status: " + LAST

    # Example 2: Attempt to remove a file that is not tracked or doesn't exist
    EMIT "Attempting to remove non-tracked file 'untracked_data.log'..."
    CALL Git.GitRm("untracked_data.log")
    SET not_tracked_status = LAST
    EMIT "Status for removing non-tracked file: " + not_tracked_status
    # Expect error like: "failed to remove path 'untracked_data.log': fatal: pathspec 'untracked_data.log' did not match any files"
    ```
* **Go Implementation Notes:**
    * Location: `pkg/core/tools_git.go`
    * Function: `toolGitRm`
    * Spec Name: `GitRm` (in `pkg/core/tools_git_register.go`)
    * Key Go Packages: `fmt`, `errors`
    * Helpers: `core.SecureFilePath`, `core.toolExec`, `core.ErrValidationArgValue`
    * Registration: Registered by `registerGitTools` within `pkg/core/tools_register.go`. Returns `string, error`.