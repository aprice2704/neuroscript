:: type: NSproject
:: subtype: tool_spec
:: version: 0.1.0
:: id: tool-spec-git-gitpush-v0.1
:: status: draft
:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_git_register.go, pkg/core/tools_git.go, docs/script_spec.md
:: relatedTo: Git.GitPull, Git.GitCommit, Git.GitStatus
:: developedBy: AI
:: reviewedBy: User

# Tool Specification Structure Template

## Tool Specification: `Git.GitPush` (v0.1)

* **Tool Name:** `Git.GitPush` (v0.1)
* **Purpose:** Updates the remote repository (usually 'origin') with local commits from the current branch. This is equivalent to running `git push` within the project's sandbox directory, pushing the current branch to its configured upstream destination.
* **NeuroScript Syntax:** `CALL Git.GitPush()`
* **Arguments:** None.
* **Return Value:** (String)
    * On success: A string indicating the push was successful, potentially including output from the `git push` command. Example: `"GitPush successful.\nOutput:\nEverything up-to-date\n"` or `"GitPush successful.\nOutput:\nTo github.com:user/repo.git\n   abc1234..def5678  main -> main\n"`. (Accessible via `LAST` after the `CALL`).
    * On failure: A string describing the failure, typically wrapping the output from the `git push` command. Examples: `"GitPush failed: ..."`. This can occur due to network errors, authentication issues, the remote having changes that need to be pulled first (rejected push), or no upstream branch being configured. (Accessible via `LAST` after the `CALL`).
* **Behavior:**
    1.  Validates that no arguments are provided.
    2.  Retrieves the interpreter's `sandboxDir`.
    3.  Constructs a command equivalent to `git push`.
    4.  Executes the `git push` command within the `sandboxDir` using the `toolExec` helper function. This involves contacting the configured remote repository.
    5.  If the command execution fails (e.g., network error, remote not found, authentication required/failed, push rejected), returns an error message string containing details from `toolExec` (which includes Git's standard error/output).
    6.  If the command execution succeeds (local commits successfully sent to the remote), returns a success message string, including any standard output from the `git push` command.
* **Security Considerations:**
    * Executes the `git push` command within the configured sandbox directory.
    * Requires network access to contact the remote repository. Firewall rules may need to allow outgoing connections.
    * May require authentication (e.g., SSH keys, HTTPS credentials) configured in the execution environment's Git settings to push to the remote repository. The tool itself does not handle credential input.
    * Modifies the state of the *remote* Git repository by sending local commit objects and updating remote branch references.
    * Relies on the `toolExec` helper for command execution.
    * Assumes the `git` executable is available in the environment's PATH where the NeuroScript interpreter is running.
* **Examples:**
    ```neuroscript
    # Example 1: Push committed changes (assuming remote 'origin' is set up and commits exist)
    # Assume Git.GitCommit was called successfully before
    EMIT "Attempting to push changes..."
    CALL Git.GitPush()
    SET push_status = LAST
    EMIT "Git Push status: " + push_status
    # Output will vary depending on whether push was successful, rejected, etc.

    # Example 2: Attempt to push when up-to-date
    EMIT "Attempting to push when likely up-to-date..."
    CALL Git.GitPush()
    SET push_up_to_date_status = LAST
    EMIT "Git Push (up-to-date) status: " + push_up_to_date_status
    # Expected output often includes "Everything up-to-date"

    # Example 3: (Illustrative) Push might fail if behind remote
    # Assume remote has changes not yet pulled locally
    # CALL Git.GitPush()
    # SET push_rejected_status = LAST
    # EMIT "Git Push (rejected) status: " + push_rejected_status
    # Expect error containing text like "rejected" or "failed to push some refs"
    ```
* **Go Implementation Notes:**
    * Location: `pkg/core/tools_git.go`
    * Function: `toolGitPush`
    * Spec Name: `GitPush` (in `pkg/core/tools_git_register.go`)
    * Key Go Packages: `fmt`
    * Helpers: `core.toolExec`
    * Registration: Registered by `registerGitTools` within `pkg/core/tools_register.go`. Returns `string, error`. Rejections or other issues are reported via the wrapped error from `toolExec`.