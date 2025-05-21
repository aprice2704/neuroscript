:: type: NSproject
:: subtype: tool_spec
:: version: 0.1.0
:: id: tool-spec-git-gitpull-v0.1
:: status: draft
:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_git_register.go, pkg/core/tools_git.go, docs/script_spec.md
:: relatedTo: Git.GitPush, Git.GitFetch, Git.GitMerge, Git.GitStatus
:: developedBy: AI
:: reviewedBy: User

# Tool Specification Structure Template

## Tool Specification: `Git.GitPull` (v0.1)

* **Tool Name:** `Git.GitPull` (v0.1)
* **Purpose:** Fetches changes from the default remote repository (usually 'origin') for the current branch and integrates them into the local branch. This is equivalent to running `git pull` within the project's sandbox directory.
* **NeuroScript Syntax:** `CALL Git.GitPull()`
* **Arguments:** None.
* **Return Value:** (String)
    * On success (fetch and integrate successful): A string indicating the pull was successful, potentially including output from the `git pull` command. Example: `"GitPull successful.\nOutput:\nUpdating abc1234..def5678\nFast-forward\n file.txt | 2 +-\n 1 file changed, 1 insertion(+), 1 deletion(-)\n"`. (Accessible via `LAST` after the `CALL`).
    * On failure or conflict: A string describing the failure, typically wrapping the output from the `git pull` command. Examples: `"GitPull failed: ..."`. This can occur due to network errors, authentication issues, merge conflicts during integration, or if the local repository has uncommitted changes preventing the pull. (Accessible via `LAST` after the `CALL`).
* **Behavior:**
    1.  Validates that no arguments are provided.
    2.  Retrieves the interpreter's `sandboxDir`.
    3.  Constructs a command equivalent to `git pull`.
    4.  Executes the `git pull` command within the `sandboxDir` using the `toolExec` helper function. This involves contacting the configured remote repository.
    5.  If the command execution fails (e.g., network error, remote not found, authentication required/failed, merge conflicts), returns an error message string containing details from `toolExec` (which includes Git's standard error/output).
    6.  If the command execution succeeds (fetch and merge/rebase completed cleanly), returns a success message string, including any standard output from the `git pull` command.
* **Security Considerations:**
    * Executes the `git pull` command within the configured sandbox directory.
    * Requires network access to contact the remote repository. Firewall rules may need to allow outgoing connections.
    * May require authentication (e.g., SSH keys, HTTPS credentials) configured in the execution environment's Git settings to access private repositories. The tool itself does not handle credential input.
    * Modifies the Git repository state by fetching new objects, updating remote-tracking branches, and potentially creating a merge commit or fast-forwarding the current branch reference. It can also modify files in the working directory and index during the integration step.
    * Relies on the `toolExec` helper for command execution.
    * Assumes the `git` executable is available in the environment's PATH where the NeuroScript interpreter is running.
* **Examples:**
    ```neuroscript
    # Example 1: Pull changes for the current branch (assuming remote 'origin' is set up)
    EMIT "Attempting to pull latest changes..."
    CALL Git.GitPull()
    SET pull_status = LAST
    EMIT "Git Pull status: " + pull_status
    # Output will vary greatly depending on whether changes exist, conflicts occur, etc.

    # Example 2: Check status after pull
    CALL Git.GitStatus()
    SET status_after_pull = LAST
    IF status_after_pull["error"] == nil THEN
       EMIT "Repository is clean after pull: " + status_after_pull["is_clean"]
       # Check if behind count is now 0 (assuming pull was successful)
       EMIT "Commits behind remote after pull: " + status_after_pull["behind"]
    ELSE
       EMIT "Error getting status after pull: " + status_after_pull["error"]
    ENDIF
    ```
* **Go Implementation Notes:**
    * Location: `pkg/core/tools_git.go`
    * Function: `toolGitPull`
    * Spec Name: `GitPull` (in `pkg/core/tools_git_register.go`)
    * Key Go Packages: `fmt`
    * Helpers: `core.toolExec`
    * Registration: Registered by `registerGitTools` within `pkg/core/tools_register.go`. Returns `string, error`. Merge conflicts or other issues are reported via the wrapped error from `toolExec`.