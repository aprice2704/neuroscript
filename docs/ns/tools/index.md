:: type: NSproject  
:: subtype: documentation  
:: version: 0.1.5  
:: id: tool-spec-index-v0.1.5  
:: status: draft  
:: dependsOn: ./*.md  
:: howToUpdate: Update the list below when tool specification documents are added, removed, or renamed in this directory.  

# NeuroScript Tool Specifications Index

This directory contains detailed specifications for the built-in functions (grouped by category like `FS`, `git`, `String`, etc.) available within the NeuroScript language. Each specification follows a standard format to ensure clarity and consistency.

## Specification Format

* **[Tool Specification Structure Template](./tool_spec_structure.md):** Defines the standard structure used for all tool specification documents in this directory.

## Available Tool Specifications (Core Tools)

### Filesystem Tools (`FS.*`)
* **[FS.ReadFile](./fs_readfile.md):** Reads the entire content of a specific file.
* **[FS.WriteFile](./fs_writefile.md):** Writes content to a specific file, overwriting if exists.
* **[FS.ListDirectory](./fs_listdirectory.md):** Lists the contents (files and subdirectories) of a directory.
* **[FS.Mkdir](./fs_mkdir.md):** Creates a new directory (including any necessary parents).
* **[FS.LineCountFile](./fs_linecountfile.md):** Counts lines in a specified file.
* **[FS.SanitizeFilename](./fs_sanitizefilename.md):** Cleans a string for use as a filename.
* **[FS.WalkDir](./fs_walkdir.md):** Recursively walks a directory tree, listing files/subdirs found.
* **[FS.FileHash](./fs_filehash.md):** Calculates the SHA256 hash of a file.
* **[FS.MoveFile](./fs_movefile.md):** Moves or renames a file or directory. (Previously `TOOL.MoveFile`)
* **[FS.DeleteFile](./fs_deletefile.md):** Deletes a file or an empty directory.

### Vector Tools (`Vector.*`)
* **[Vector.SearchSkills](./vector_searchskills.md):** Searches the (mock) vector index for skills matching a query.
* **[Vector.VectorUpdate](./vector_vectorupdate.md):** Updates the (mock) vector index for a given file.

### Git Tools (`git.*`)
* **[git.Add](./git_add.md):** Stages changes for commit (`git add`).
* **[git.Commit](./git_commit.md):** Commits staged changes (`git commit -m`).
* **[git.NewBranch](./git_newbranch.md):** Creates and checks out a new branch (`git checkout -b`).
* **[git.Checkout](./git_checkout.md):** Checks out an existing branch or commit (`git checkout`).
* **[git.Rm](./git_rm.md):** Removes a file from the Git index (`git rm`).
* **[git.Merge](./git_merge.md):** Merges a branch into the current branch (`git merge`).
* **[git.Status](./git_status.md):** Gets the current Git repository status. (Previously `TOOL.GitStatus`)
* **[git.Pull](./git_pull.md):** Fetches and integrates changes from a remote repository (`git pull`). (Previously `TOOL.GitPull`)
* **[git.Push](./git_push.md):** Pushes local commits to a remote repository (`git push`). (Previously `TOOL.GitPush`)
* **[git.Diff](./git_diff.md):** Shows unstaged changes in the working directory (`git diff`). (Previously `TOOL.GitDiff`)

### String Tools (`String.*`)
* **[String Tools Summary](./string_summary.md):** Overview of all string manipulation tools (`StringLength`, `Substring`, `ToUpper`, `ToLower`, `TrimSpace`, `SplitString`, `SplitWords`, `JoinStrings`, `ReplaceAll`, `Contains`, `HasPrefix`, `HasSuffix`, `LineCountString`).

### Shell Tools (`Shell.*`)
* **[Shell.ExecuteCommand](./shell_executecommand.md):** Executes an arbitrary external command. (**Use with extreme caution!**)
* _(GoBuild, GoCheck, GoTest, GoFmt, GoModTidy specs needed)_

### Math Tools (`Math.*`)
* _(Add, Subtract, etc. specs needed)_

### Metadata Tools (`Metadata.*`)
* _(GetMetadata, SetMetadata specs needed)_

### List Tools (`List.*`)
* _(ListLength, GetElement, AppendToList, etc. specs needed)_

### Go AST Tools (`GoAST.*`)
* **[GoAST.UpdateImportsForMovedPackage](./go_update_imports_for_moved_package.md):** Updates Go import paths after refactoring. (Previously `TOOL.GoUpdateImports...`)
* _(GoParseFile, GoFindIdentifiers, etc. specs needed)_

### IO Tools (`IO.*`)
* _(Log, Input specs needed)_

### LLM Tools (`LLM.*`)
* _(Call specs needed)_

## Available Tool Specifications (File API Tools - External)

*These tools interact with the Gemini File API and may not be considered core language tools in the same way as the above.*

* **[FileAPI.ListAPIFiles](./list_api_files.md):** Lists files stored in the Gemini File API. (Previously `TOOL.ListAPIFiles`)
* **[FileAPI.DeleteAPIFile](./delete_api_file.md):** Deletes a file from the Gemini File API. (Previously `TOOL.DeleteAPIFile`)
* **[FileAPI.UploadFile](./upload_file.md):** Uploads a local file to the Gemini File API. (Previously `TOOL.UploadFile`)
* **[FileAPI.SyncFiles](./sync_files.md):** Synchronizes files between local and API storage. (Previously `TOOL.SyncFiles`)

## Available Tool Specifications (NeuroData Tools - External)

* **[NeuroData.QueryTable](./query_table.md):** Queries NeuroData Table (`.ndtable`) files. (Previously `TOOL.QueryTable`)

*(This list should be updated as more tool specifications are created or existing ones are updated.)*