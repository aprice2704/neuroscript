:: type: NSproject
:: subtype: tool_spec
:: version: 0.1.0
:: id: tool-spec-git-gitstatus-v0.1
:: status: draft
:: dependsOn: docs/ns/tools/tool_spec_structure.md, pkg/core/tools_git_register.go, pkg/core/tools_git_status.go, docs/script_spec.md
:: relatedTo: Git.GitAdd, Git.GitCommit, Git.GitCheckout
:: developedBy: AI
:: reviewedBy: User

# Tool Specification Structure Template

## Tool Specification: `Git.GitStatus` (v0.1)

* **Tool Name:** `Git.GitStatus` (v0.1)
* **Purpose:** Retrieves the status of the Git repository within the sandbox, providing detailed information about the current branch, tracking status, and changes to files. It uses `git status --porcelain -b --untracked-files=all` internally and parses the output.
* **NeuroScript Syntax:** `CALL Git.GitStatus()`
* **Arguments:** None.
* **Return Value:** (Map)
    * Returns a Map containing a structured representation of the Git status. (Accessible via `LAST` after the `CALL`). The Map always contains the following keys:
        * `branch` (String | nil): The name of the current local branch. Can be `"(detached HEAD)"` if not on a branch, or the branch name if on an unborn branch (e.g., after `git init` but before first commit). `nil` only if parsing fails unexpectedly.
        * `remote_branch` (String | nil): The name of the remote tracking branch (e.g., `origin/main`), or `nil` if the current branch is not tracking a remote branch.
        * `ahead` (Number): The number of commits the local branch is ahead of its remote tracking branch. `0` if not tracking or up-to-date.
        * `behind` (Number): The number of commits the local branch is behind its remote tracking branch. `0` if not tracking or up-to-date.
        * `files` (List of Maps): A list containing information about files with changes (staged, unstaged, or untracked). Each map in the list has:
            * `path` (String): The relative path to the file within the repository. Paths with special characters may be unquoted from Git's C-style quoting.
            * `index_status` (String): A single character representing the status of the file in the index (staging area). Common values: `M` (modified), `A` (added), `D` (deleted), `R` (renamed), `C` (copied), `U` (unmerged), `?` (untracked), `!` (ignored), ` ` (unmodified).
            * `worktree_status` (String): A single character representing the status of the file in the working tree relative to the index. Common values: `M` (modified), `D` (deleted), `?` (untracked), ` ` (unmodified/tracked).
            * `original_path` (String | nil): For renamed (`R`) or copied (`C`) files in the index, this holds the original path the file was renamed/copied from. Otherwise, it is `nil`. May be unquoted.
        * `untracked_files_present` (Boolean): `true` if there are any files listed with status `??` (untracked), `false` otherwise.
        * `is_clean` (Boolean): `true` only if there are *no* staged changes, *no* unstaged changes to tracked files, AND *no* untracked files. `false` otherwise.
        * `error` (String | nil): Contains an error message string if the `git status` command failed (e.g., "Not a git repository...") or if the output parsing failed. Contains `nil` if the command executed successfully and parsing succeeded (even if the repository is not clean).
* **Behavior:**
    1.  Validates that no arguments are provided.
    2.  Executes the command `git status --porcelain -b --untracked-files=all` within the `sandboxDir` using the `toolExec` helper.
    3.  If `toolExec` returns an error (e.g., the directory is not a Git repository):
        * Creates a default result Map.
        * Sets the `error` key in the Map with a descriptive message based on the error.
        * Returns the Map.
    4.  If `toolExec` succeeds, parses the output string line by line:
        * Parses the first line (starting with `##`) to extract local branch, remote branch (if tracked), and ahead/behind counts. Handles detached HEAD and unborn branch states.
        * Parses subsequent lines to identify file statuses (staged, unstaged, untracked, renamed/copied). Extracts the status codes (index, worktree) and paths (handling C-style quoting and `->` for renames/copies).
        * Populates the `files` list with maps containing the parsed file status details.
        * Sets `untracked_files_present` if any `??` files are found.
        * Determines `is_clean` based on whether any tracked file changes or untracked files were detected.
    5.  If parsing fails at any stage, sets the `error` key in the result Map.
    6.  Returns the fully populated result Map.
* **Security Considerations:**
    * Executes the `git status` command within the configured sandbox directory.
    * Reads Git repository metadata and file status information; does not modify the repository state or files.
    * Relies on the `toolExec` helper for command execution.
    * Assumes the `git` executable is available in the environment's PATH where the NeuroScript interpreter is running.
* **Examples:**
    ```neuroscript
    # Example 1: Get status of a clean repository on main branch
    # Assume repo is clean and on 'main' tracking 'origin/main'
    CALL Git.GitStatus()
    SET status_result = LAST
    EMIT "Clean Repo Status:"
    EMIT " Branch: " + status_result["branch"]
    EMIT " Remote: " + status_result["remote_branch"]
    EMIT " Ahead: " + status_result["ahead"] + ", Behind: " + status_result["behind"]
    EMIT " Is Clean: " + status_result["is_clean"]
    EMIT " Untracked Present: " + status_result["untracked_files_present"]
    EMIT " Error: " + status_result["error"]
    EMIT " Files Count: " + List.Length(status_result["files"]) # Requires List.Length
    # Expected: main, origin/main, 0, 0, true, false, nil, 0 (if List.Length exists)

    # Example 2: Get status after modifying and adding a file
    CALL FS.WriteFile("config.yml", "new_setting: true")
    CALL Git.GitAdd(["config.yml"])
    CALL Git.GitStatus()
    SET modified_status = LAST
    EMIT "Modified Repo Status:"
    EMIT " Is Clean: " + modified_status["is_clean"] # Expect false
    EMIT " Error: " + modified_status["error"] # Expect nil
    # Loop through files to see details (requires List tools)
    IF modified_status["error"] == nil AND List.Length(modified_status["files"]) > 0 THEN
      SET changed_file = modified_status["files"][0]
      EMIT " Changed File Path: " + changed_file["path"] # Expect config.yml
      EMIT " Index Status: " + changed_file["index_status"] # Expect 'A' (Added)
      EMIT " Worktree Status: " + changed_file["worktree_status"] # Expect ' '
    ENDIF

    # Example 3: Get status in a non-git directory
    # Assume "not_a_repo" directory exists but is not a git repo
    # This requires changing the interpreter's sandboxDir, which isn't standard tool behavior.
    # Illustrative result if run in non-repo:
    # CALL Git.GitStatus()
    # SET non_repo_status = LAST
    # EMIT "Non-Repo Status Error: " + non_repo_status["error"]
    # Expect error like: "Not a git repository..."
    ```
* **Go Implementation Notes:**
    * Location: `pkg/core/tools_git_status.go`
    * Function: `toolGitStatus` (calls `parseGitStatusOutput`)
    * Spec Name: `GitStatus` (in `pkg/core/tools_git_register.go`)
    * Key Go Packages: `fmt`, `regexp`, `strconv`, `strings`
    * Helpers: `core.toolExec`, `core.parseGitStatusOutput`
    * Registration: Registered by `registerGitTools` within `pkg/core/tools_register.go`. Returns `map[string]interface{}, error`, but execution/parsing errors are placed in the returned map's `error` field instead of returning a Go error.