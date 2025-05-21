@@@:: type: NSproject
@@@:: subtype: tool_spec
@@@:: version: 0.1.0
@@@:: id: tool-spec-git-gitcommit-v0.1
@@@:: status: draft
@@@:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_git_register.go, pkg/core/tools_git.go, docs/script_spec.md
@@@:: relatedTo: Git.GitAdd, Git.GitStatus
@@@:: developedBy: AI
@@@:: reviewedBy: User
@@@
@@@# Tool Specification Structure Template
@@@
@@@## Tool Specification: `Git.GitCommit` (v0.1)
@@@
@@@* **Tool Name:** `Git.GitCommit` (v0.1)
@@@* **Purpose:** Creates a new commit containing the currently staged changes. This is equivalent to running `git commit -m "<message>"` within the project's sandbox directory.
@@@* **NeuroScript Syntax:** `CALL Git.GitCommit(message: <String>)`
@@@* **Arguments:**
@@@    * `message` (String): Required. The commit message describing the changes being committed. This message cannot be an empty string.
@@@* **Return Value:** (String)
@@@    * On success: A string indicating the commit was successful, including the commit message and potentially output from the `git commit` command. Example: `"GitCommit successful. Message: "Fix typo in README".\nOutput:\n[main abc1234] Fix typo in README\n 1 file changed, 1 insertion(+), 1 deletion(-)\n"`. (Accessible via `LAST` after the `CALL`).
@@@    * On error: A string describing the failure. This could be due to an empty commit message ("commit message cannot be empty"), nothing being staged for commit, or other errors from the underlying `git commit` command execution (e.g., "GitCommit failed: ..."). (Accessible via `LAST` after the `CALL`).
@@@* **Behavior:**
@@@    1.  Validates that exactly one argument (`message`) of type String is provided.
@@@    2.  Validates that the `message` is not empty. If it is, returns an error message string.
@@@    3.  Retrieves the interpreter's `sandboxDir`.
@@@    4.  Constructs a command equivalent to `git commit -m "<message>"`.
@@@    5.  Executes the `git commit` command within the `sandboxDir` using the `toolExec` helper function.
@@@    6.  If the command execution fails (e.g., `git` not found, nothing staged, commit hook failure), returns an error message string containing details from `toolExec`.
@@@    7.  If the command execution succeeds, returns a success message string, including the original commit message and any standard output from the `git commit` command.
@@@* **Security Considerations:**
@@@    * Executes the `git commit` command within the configured sandbox directory.
@@@    * Modifies the Git repository history by creating a new commit object.
@@@    * Relies on the `toolExec` helper for command execution.
@@@    * Assumes the `git` executable is available in the environment's PATH where the NeuroScript interpreter is running.
@@@* **Examples:**
@@@    ```neuroscript
@@@    # Example 1: Commit previously staged changes
@@@    # Assume Git.GitAdd(["README.md"]) was called successfully before
@@@    SET commit_message = "Update project README with installation instructions"
@@@    CALL Git.GitCommit(commit_message)
@@@    SET commit_status = LAST
@@@    EMIT "Git Commit status: " + commit_status
@@@
@@@    # Example 2: Attempt to commit with nothing staged (will likely fail)
@@@    EMIT "Attempting commit with nothing staged..."
@@@    CALL Git.GitCommit("Add new feature X")
@@@    SET nothing_staged_status = LAST
@@@    EMIT "Status for commit with nothing staged: " + nothing_staged_status
@@@    # Expect error like: "GitCommit failed: On branch main\nnothing to commit, working tree clean..."
@@@
@@@    # Example 3: Attempt to commit with an empty message (will fail validation)
@@@    CALL Git.GitCommit("")
@@@    SET empty_message_status = LAST
@@@    EMIT "Status for commit with empty message: " + empty_message_status
@@@    # Expect error like: "commit message cannot be empty..."
@@@    ```
@@@* **Go Implementation Notes:**
@@@    * Location: `pkg/core/tools_git.go`
@@@    * Function: `toolGitCommit`
@@@    * Spec Name: `GitCommit` (in `pkg/core/tools_git_register.go`)
@@@    * Key Go Packages: `fmt`
@@@    * Helpers: `core.toolExec`, `core.ErrValidationArgValue`
@@@    * Registration: Registered by `registerGitTools` within `pkg/core/tools_register.go`. Returns `string, error`.