Interpreter created using the 'api' package.
Executing script: ./library/list_tools.ns.txt
Running procedure 'main'...
Compact Tool List:
tool.Meta.GetToolSpecificationsJSON() -> string
tool.Meta.ListTools() -> string
tool.Meta.ToolsHelp(filter:string?) -> string
tool.ai.Select(value:any, path:any, missing_ok:bool?) -> any
tool.ai.SelectMany(value:any, extracts:map, missing_ok:bool?) -> map
tool.ai.Validate(value:map, shape:map, allow_extra:bool?) -> bool
tool.error.New(code:any, message:string) -> error
tool.fs.Append(filepath:string, content:string) -> string
tool.fs.Delete(path:string) -> string
tool.fs.Hash(filepath:string) -> string
tool.fs.LineCount(filepath:string) -> int
tool.fs.List(path:string, recursive:bool?) -> slice_any
tool.fs.Mkdir(path:string) -> map
tool.fs.Move(source_path:string, destination_path:string) -> map
tool.fs.Read(filepath:string) -> string
tool.fs.SanitizeFilename(name:string) -> string
tool.fs.Stat(path:string) -> map
tool.fs.Walk(path:string) -> slice_any
tool.fs.Write(filepath:string, content:string) -> string
tool.git.Add(relative_path:string, paths:slice_any) -> string
tool.git.Branch(relative_path:string, name:string?, checkout:bool?, list_remote:bool?, list_all:bool?) -> string
tool.git.Checkout(relative_path:string, branch:string, create:bool?) -> string
tool.git.Clone(repository_url:string, relative_path:string) -> string
tool.git.Commit(relative_path:string, commit_message:string, allow_empty:bool?) -> string
tool.git.Diff(relative_path:string, cached:bool?, commit1:string?, commit2:string?, path:string?) -> string
tool.git.Merge(relative_path:string, branch:string) -> string
tool.git.Pull(relative_path:string, remote_name:string?, branch_name:string?) -> string
tool.git.Push(relative_path:string, remote_name:string?, branch_name:string?) -> string
tool.git.Reset(relative_path:string, mode:string?, commit:string?) -> string
tool.git.Rm(relative_path:string, paths:any) -> string
tool.git.Status(repo_path:string?) -> map
tool.gotools.Build(target:string?) -> map
tool.gotools.Check(target:string) -> map
tool.gotools.Fmt(content:string) -> string
tool.gotools.GetModuleInfo(directory:string?) -> map
tool.gotools.Imports(content:string) -> string
tool.gotools.ListPackages(target_directory:string?, patterns:slice_string?) -> slice_map
tool.gotools.ModTidy() -> map
tool.gotools.Staticcheck(target:string?) -> map
tool.gotools.Test(target:string?) -> map
tool.gotools.Vet(target:string?) -> map
tool.io.Input(message:string?) -> string
tool.io.Print(values:any) -> nil
tool.list.Append(list:slice_any, element:any?) -> slice_any
tool.list.Contains(list:slice_any, element:any?) -> bool
tool.list.Get(list:slice_any, index:int, default:any?) -> any
tool.list.Head(list:slice_any) -> any
tool.list.IsEmpty(list:slice_any) -> bool
tool.list.Length(list:slice_any) -> int
tool.list.Prepend(list:slice_any, element:any?) -> slice_any
tool.list.Rest(list:slice_any) -> slice_any
tool.list.Reverse(list:slice_any) -> slice_any
tool.list.Slice(list:slice_any, start:int, end:int) -> slice_any
tool.list.Sort(list:slice_any) -> slice_any
tool.list.Tail(list:slice_any, count:int) -> slice_any
tool.math.Add(num1:float, num2:float) -> float
tool.math.Divide(num1:float, num2:float) -> float
tool.math.Modulo(num1:int, num2:int) -> int
tool.math.Multiply(num1:float, num2:float) -> float
tool.math.Subtract(num1:float, num2:float) -> float
tool.script.ListFunctions() -> map
tool.script.LoadScript(script_content:string) -> map
tool.shell.Execute(command:string, args_list:slice_string?, directory:string?) -> map
tool.str.Concat(strings_list:slice_string) -> string
tool.str.Contains(input_string:string, substring:string) -> bool
tool.str.HasPrefix(input_string:string, prefix:string) -> bool
tool.str.HasSuffix(input_string:string, suffix:string) -> bool
tool.str.Join(string_list:slice_string, separator:string) -> string
tool.str.Length(input_string:string) -> int
tool.str.LineCount(content_string:string) -> int
tool.str.Replace(input_string:string, old_substring:string, new_substring:string, count:int) -> string
tool.str.Split(input_string:string, delimiter:string) -> slice_string
tool.str.SplitWords(input_string:string) -> slice_string
tool.str.Substring(input_string:string, start_index:int, length:int) -> string
tool.str.ToLower(input_string:string) -> string
tool.str.ToUpper(input_string:string) -> string
tool.str.TrimSpace(input_string:string) -> string
tool.syntax.analyzeNSSyntax(nsScriptContent:string) -> slice_map
tool.time.Now() -> timedate
tool.time.Sleep(duration_seconds:number?) -> boolean
tool.tree.AddChildNode(tree_handle:string, parent_node_id:string, new_node_id_suggestion:string?, node_type:string, value:any?, key_for_object_parent:string?) -> string
tool.tree.FindNodes(tree_handle:string, start_node_id:string, query_map:map, max_depth:int?, max_results:int?) -> slice_string
tool.tree.GetChildren(tree_handle:string, node_id:string) -> slice_string
tool.tree.GetNode(tree_handle:string, node_id:string) -> map
tool.tree.GetNodeByPath(tree_handle:string, path:string) -> map
tool.tree.GetNodeMetadata(tree_handle:string, node_id:string) -> map
tool.tree.GetParent(tree_handle:string, node_id:string) -> map
tool.tree.GetRoot(tree_handle:string) -> map
tool.tree.LoadJSON(json_string:string) -> string
tool.tree.RemoveNode(tree_handle:string, node_id:string) -> nil
tool.tree.RemoveNodeMetadata(tree_handle:string, node_id:string, metadata_key:string) -> nil
tool.tree.RemoveObjectAttribute(tree_handle:string, object_node_id:string, attribute_key:string) -> nil
tool.tree.RenderText(tree_handle:string) -> string
tool.tree.SetNodeMetadata(tree_handle:string, node_id:string, metadata_key:string, metadata_value:string) -> nil
tool.tree.SetObjectAttribute(tree_handle:string, object_node_id:string, attribute_key:string, child_node_id:string) -> nil
tool.tree.SetValue(tree_handle:string, node_id:string, value:any) -> nil
tool.tree.ToJSON(tree_handle:string) -> string


    --------------------
    
Detailed Tool Help (Markdown):
# NeuroScript Tools Help

## `tool.Meta.GetToolSpecificationsJSON`
**Description:** Provides a JSON string containing an array of all currently available tool specifications. Each object in the array represents a tool and includes its name, description, category, arguments (with their details), return type, return help, variadic status, example usage, and error conditions.

**Category:** Introspection

**Parameters:**
_None_

**Returns:** (`string`) A JSON string representing an array of ToolSpec objects. This is intended for programmatic use or detailed inspection of all tool capabilities.

**Example:**
```neuroscript
GetToolSpecificationsJSON()
```
---

## `tool.Meta.ListTools`
**Description:** Provides a compact text list (sorted alphabetically) of all currently available tools, including basic parameter information. Each tool is listed on a new line, showing its name, parameters (name:type), and return type. Example: FS.Read(filepath:string) -> string

**Category:** Introspection

**Parameters:**
_None_

**Returns:** (`string`) A string containing a newline-separated list of tool names, their parameters (name:type), and return types.

**Example:**
```neuroscript
ListTools()
```
---

## `tool.Meta.ToolsHelp`
**Description:** Provides a more extensive, Markdown-formatted list of available tools, including descriptions, parameters, and return types. Can be filtered by providing a partial tool name. Details include parameter names, types, descriptions, and return type with its description.

**Category:** Introspection

**Parameters:**
* `filter` (`string`): (optional) An optional string to filter tool names. Only tools whose names contain this substring will be listed. If empty or omitted, all tools are listed.

**Returns:** (`string`) A string in Markdown format detailing available tools, their descriptions, parameters, and return types. Output can be filtered by the optional 'filter' argument.

**Example:**
```neuroscript
ToolsHelp(filter: "FS")
Meta.ToolsHelp()
```
---

## `tool.ai.Select`
**Description:** Selects a single value from a map or list using a path, according to Path-Lite spec.

**Category:** Data Selection

**Parameters:**
* `value` (`any`): The map or list to select from.
* `path` (`any`): The string or list path to the desired value.
* `missing_ok` (`bool`): (optional) If true, returns nil if the path does not exist instead of failing.

**Returns:** (`any`) Returns the value found at the specified path.

**Example:**
```neuroscript
tool.ai.Select(my_data, "user.name")
```
---

## `tool.ai.SelectMany`
**Description:** Selects multiple values from a map or list using a map of target keys to paths.

**Category:** Data Selection

**Parameters:**
* `value` (`any`): The map or list to select from.
* `extracts` (`map`): A map where keys are new names and values are the paths to extract.
* `missing_ok` (`bool`): (optional) If true, keys for missing paths will be omitted from the result instead of failing.

**Returns:** (`map`) Returns a new map containing the extracted key-value pairs.

**Example:**
```neuroscript
tool.ai.SelectMany(my_data, {"name": "user.name", "city": "user.address.city"})
```
---

## `tool.ai.Validate`
**Description:** Validates a map against a shape definition, according to Shape-Lite spec.

**Category:** Data Validation

**Parameters:**
* `value` (`map`): The data map to validate.
* `shape` (`map`): The shape map to validate against.
* `allow_extra` (`bool`): (optional) If true, allows extra keys in the value not present in the shape.

**Returns:** (`bool`) Returns true on success, otherwise returns a validation error.

**Example:**
```neuroscript
tool.ai.Validate(my_data, my_shape, false)
```
---

## `tool.error.New`
**Description:** Constructs a standard NeuroScript error value map.

**Category:** Error Handling

**Parameters:**
* `code` (`any`): A string or integer error code.
* `message` (`string`): A human-readable error message.

**Returns:** (`error`) Returns an 'error' type value, which is a map containing 'code' and 'message'.

**Example:**
```neuroscript
set file_err = tool.Error.New("ERR_NOT_FOUND", "The specified file does not exist.")
```
---

## `tool.fs.Append`
**Description:** Appends content to a specific file. Creates the file and parent directories if needed. Returns 'OK' on success.

**Category:** Filesystem

**Parameters:**
* `filepath` (`string`): Relative path to the file.
* `content` (`string`): The content to append.

**Returns:** (`string`) Returns 'OK' on success. Returns nil on error.

**Example:**
```neuroscript
TOOL.FS.Append(filepath: "logs/activity.log", content: "User logged in.\n")
```
---

## `tool.fs.Delete`
**Description:** Deletes a file or an empty directory. Returns 'OK' on success or if path doesn't exist.

**Category:** Filesystem

**Parameters:**
* `path` (`string`): Relative path to the file or empty directory to delete.

**Returns:** (`string`) Returns the string 'OK' on successful deletion or if the path does not exist. Returns nil on error.

**Example:**
```neuroscript
TOOL.FS.Delete(path: "temp/old_file.txt") // Returns "OK"
```
---

## `tool.fs.Hash`
**Description:** Calculates the SHA256 hash of a specified file. Returns the hex-encoded hash string.

**Category:** Filesystem

**Parameters:**
* `filepath` (`string`): Relative path (within the sandbox) of the file to hash.

**Returns:** (`string`) Returns a hex-encoded SHA256 hash string of the file's content. Returns an empty string on error.

**Example:**
```neuroscript
TOOL.FS.Hash(filepath: "data/my_document.txt") // Returns "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" (example for an empty file)
```
---

## `tool.fs.LineCount`
**Description:** Counts lines in a specified file. Returns line count as an integer.

**Category:** Filesystem

**Parameters:**
* `filepath` (`string`): Relative path to the file.

**Returns:** (`int`) Returns the number of lines in the specified file. Returns 0 on error or if file is empty.

**Example:**
```neuroscript
TOOL.FS.LineCount(filepath: "logs/app.log") // Returns 150 (example)
```
---

## `tool.fs.List`
**Description:** Lists files and subdirectories at a given path. Returns a list of maps, each describing an entry (keys: name, path, isDir, size, modTime).

**Category:** Filesystem

**Parameters:**
* `path` (`string`): Relative path to the directory (use '.' for current).
* `recursive` (`bool`): (optional) Whether to list recursively (default: false).

**Returns:** (`slice_any`) Returns a slice of maps. Each map details a file/directory: {'name':string, 'path':string (relative to input path for recursive), 'isDir':bool, 'size':int64, 'modTime':string (RFC3339Nano)}. Returns nil on error.

**Example:**
```neuroscript
TOOL.FS.List(path: "mydir", recursive: true)
```
---

## `tool.fs.Mkdir`
**Description:** Creates a directory. Parent directories are created if they do not exist (like mkdir -p). Returns a success message.

**Category:** Filesystem

**Parameters:**
* `path` (`string`): Relative path of the directory to create.

**Returns:** (`map`) Returns a map: {'status':'success', 'message':'Successfully created directory: <path>', 'path':'<path>'} on success. Returns nil on error.

**Example:**
```neuroscript
TOOL.FS.Mkdir(path: "new/subdir") // Returns {"status":"success", "message":"Successfully created directory: new/subdir", "path":"new/subdir"}
```
---

## `tool.fs.Move`
**Description:** Moves or renames a file or directory within the sandbox. Returns a map: {'message': 'success message', 'error': nil} on success.

**Category:** Filesystem

**Parameters:**
* `source_path` (`string`): Relative path of the source file/directory.
* `destination_path` (`string`): Relative path of the destination.

**Returns:** (`map`) Returns a map {'message': 'success message', 'error': nil} on success. Returns nil on error.

**Example:**
```neuroscript
TOOL.FS.Move(source_path: "old_name.txt", destination_path: "new_name.txt")
```
---

## `tool.fs.Read`
**Description:** Reads the entire content of a specific file. Returns the content as a string.

**Category:** Filesystem

**Parameters:**
* `filepath` (`string`): Relative path to the file.

**Returns:** (`string`) Returns the content of the file as a string. Returns an empty string on error.

**Example:**
```neuroscript
TOOL.FS.Read(filepath: "config.txt") // Returns "setting=value\n..."
```
---

## `tool.fs.SanitizeFilename`
**Description:** Cleans a string to make it suitable for use as part of a filename.

**Category:** Filesystem Utilities

**Parameters:**
* `name` (`string`): The string to sanitize.

**Returns:** (`string`) Returns a sanitized string suitable for use as a filename component (e.g., replacing unsafe characters with underscores).

**Example:**
```neuroscript
TOOL.FS.SanitizeFilename(name: "My Report Final?.docx") // Returns "My_Report_Final_.docx" (example)
```
---

## `tool.fs.Stat`
**Description:** Gets information about a file or directory. Returns a map containing: name(string), path(string), size_bytes(int), is_dir(bool), modified_unix(int), modified_rfc3339(string - format 2006-01-02T15:04:05.999999999Z07:00), mode_string(string), mode_perm(string).

**Category:** Filesystem

**Parameters:**
* `path` (`string`): Relative path to the file or directory.

**Returns:** (`map`) Returns a map with file/directory info: {'name', 'path', 'size_bytes', 'is_dir', 'modified_unix', 'modified_rfc3339', 'mode_string', 'mode_perm'}. Returns nil on error.

**Example:**
```neuroscript
TOOL.FS.Stat(path: "my_file.go")
```
---

## `tool.fs.Walk`
**Description:** Recursively walks a directory, returning a list of maps describing files/subdirectories found (keys: name, path_relative, is_dir, size_bytes, modified_unix, modified_rfc3339 (format 2006-01-02T15:04:05.999999999Z07:00), mode_string). Skips the root directory itself.

**Category:** Filesystem

**Parameters:**
* `path` (`string`): Relative path to the directory to walk.

**Returns:** (`slice_any`) Returns a slice of maps, each describing a file/subdir: {'name', 'path_relative', 'is_dir', 'size_bytes', 'modified_unix', 'modified_rfc3339', 'mode_string'}. Skips the root dir itself. Returns nil on error.

**Example:**
```neuroscript
TOOL.FS.Walk(path: "src")
```
---

## `tool.fs.Write`
**Description:** Writes content to a specific file, overwriting it if it exists. Creates parent directories if needed. Returns 'OK' on success.

**Category:** Filesystem

**Parameters:**
* `filepath` (`string`): Relative path to the file.
* `content` (`string`): The content to write.

**Returns:** (`string`) Returns 'OK' on success. Returns nil on error.

**Example:**
```neuroscript
TOOL.FS.Write(filepath: "output/data.json", content: "{\"key\":\"value\"}")
```
---

## `tool.git.Add`
**Description:** Adds file contents to the index.

**Category:** Git

**Parameters:**
* `relative_path` (`string`): The relative path within the sandbox to the Git repository.
* `paths` (`slice_any`): A list of file paths to add to the index.

**Returns:** (`string`) Returns a success message upon completion.

**Example:**
```neuroscript
TOOL.Git.Add(relative_path: "my_repo", paths: ["file1.txt", "docs/"])
```
---

## `tool.git.Branch`
**Description:** Manages branches.

**Category:** Git

**Parameters:**
* `relative_path` (`string`): Path to the repository.
* `name` (`string`): (optional) The name of the branch to create. If omitted, lists branches.
* `checkout` (`bool`): (optional) If true, checks out the new branch after creation.
* `list_remote` (`bool`): (optional) If true, lists remote-tracking branches.
* `list_all` (`bool`): (optional) If true, lists all branches.

**Returns:** (`string`) Returns a success message or a list of branches.

**Example:**
```neuroscript
TOOL.Git.Branch(relative_path: "my_repo", name: "new-feature")
```
---

## `tool.git.Checkout`
**Description:** Switches branches or restores working tree files.

**Category:** Git

**Parameters:**
* `relative_path` (`string`): Path to the repository.
* `branch` (`string`): The branch to checkout.
* `create` (`bool`): (optional) If true, creates and checks out the branch.

**Returns:** (`string`) Returns a success message.

**Example:**
```neuroscript
TOOL.Git.Checkout(relative_path: "my_repo", branch: "main")
```
---

## `tool.git.Clone`
**Description:** Clones a Git repository into the specified relative path within the sandbox.

**Category:** Git

**Parameters:**
* `repository_url` (`string`): The URL of the Git repository to clone.
* `relative_path` (`string`): The relative path within the sandbox where the repository should be cloned.

**Returns:** (`string`) Returns a success message.

**Example:**
```neuroscript
TOOL.Git.Clone(repository_url: "https://github.com/example/repo.git", relative_path: "cloned_repos/my_repo")
```
---

## `tool.git.Commit`
**Description:** Commits staged changes.

**Category:** Git

**Parameters:**
* `relative_path` (`string`): Path to the repository.
* `commit_message` (`string`): The commit message.
* `allow_empty` (`bool`): (optional) Allow an empty commit. Defaults to false.

**Returns:** (`string`) Returns a success message.

**Example:**
```neuroscript
TOOL.Git.Commit(relative_path: "my_repo", commit_message: "Fix: bug #123")
```
---

## `tool.git.Diff`
**Description:** Shows changes between commits, commit and working tree, etc.

**Category:** Git

**Parameters:**
* `relative_path` (`string`): Path to the repository.
* `cached` (`bool`): (optional) Show staged changes.
* `commit1` (`string`): (optional) First commit for diff.
* `commit2` (`string`): (optional) Second commit for diff.
* `path` (`string`): (optional) Limit the diff to a specific path.

**Returns:** (`string`) Returns a string containing the diff output.

**Example:**
```neuroscript
TOOL.Git.Diff(relative_path: "my_repo", cached: true)
```
---

## `tool.git.Merge`
**Description:** Joins two or more development histories together.

**Category:** Git

**Parameters:**
* `relative_path` (`string`): Path to the repository.
* `branch` (`string`): The branch to merge into the current branch.

**Returns:** (`string`) Returns a success message.

**Example:**
```neuroscript
TOOL.Git.Merge(relative_path: "my_repo", branch: "feature-branch")
```
---

## `tool.git.Pull`
**Description:** Pulls the latest changes from the remote repository.

**Category:** Git

**Parameters:**
* `relative_path` (`string`): Path to the repository.
* `remote_name` (`string`): (optional) Optional. The remote to pull from. Defaults to 'origin'.
* `branch_name` (`string`): (optional) Optional. The branch to pull. Defaults to the current branch.

**Returns:** (`string`) Returns a success message.

**Example:**
```neuroscript
TOOL.Git.Pull(relative_path: "my_repo")
```
---

## `tool.git.Push`
**Description:** Pushes committed changes to a remote repository.

**Category:** Git

**Parameters:**
* `relative_path` (`string`): Path to the repository.
* `remote_name` (`string`): (optional) Optional. The remote to push to. Defaults to 'origin'.
* `branch_name` (`string`): (optional) Optional. The branch to push. Defaults to the current branch.

**Returns:** (`string`) Returns a success message.

**Example:**
```neuroscript
TOOL.Git.Push(relative_path: "my_repo")
```
---

## `tool.git.Reset`
**Description:** Resets the current HEAD to the specified state.

**Category:** Git

**Parameters:**
* `relative_path` (`string`): Path to the repository.
* `mode` (`string`): (optional) Reset mode: 'soft', 'mixed' (default), 'hard', 'merge', or 'keep'.
* `commit` (`string`): (optional) Commit to reset to. Defaults to HEAD.

**Returns:** (`string`) Returns a success message.

**Example:**
```neuroscript
TOOL.Git.Reset(relative_path: "my_repo", mode: "hard", commit: "HEAD~1")
```
---

## `tool.git.Rm`
**Description:** Removes files from the working tree and from the index.

**Category:** Git

**Parameters:**
* `relative_path` (`string`): Path to the repository.
* `paths` (`any`): A single path or a list of paths to remove.

**Returns:** (`string`) Returns a success message.

**Example:**
```neuroscript
TOOL.Git.Rm(relative_path: "my_repo", paths: "old_file.txt")
```
---

## `tool.git.Status`
**Description:** Gets the status of the Git repository in the configured sandbox directory.

**Category:** Git

**Parameters:**
* `repo_path` (`string`): (optional) Optional. Relative path to the repository within the sandbox. Defaults to the sandbox root.

**Returns:** (`map`) Returns a map containing Git status information.

**Example:**
```neuroscript
TOOL.Git.Status()
```
---

## `tool.gotools.Build`
**Description:** Runs 'go build' for a specified target in the sandbox. Defaults to './...'.

**Category:** Go Build Tools

**Parameters:**
* `target` (`string`): (optional) Optional. The build target (e.g., a package path or './...'). Defaults to './...'.

**Returns:** (`map`) Returns a map with 'stdout', 'stderr', 'exit_code' (int64), and 'success' (bool) from the 'go build <target>' command.

**Example:**
```neuroscript
tool.gotools.Build(target: "./cmd/mytool")
```
---

## `tool.gotools.Check`
**Description:** Checks Go code validity using 'go list -e -json <target>' within the sandbox. Returns a map indicating success and error details.

**Category:** Go types.Diagnostics

**Parameters:**
* `target` (`string`): Target Go package path or file path relative to sandbox (e.g., './pkg/core', 'main.go').

**Returns:** (`map`) Returns a map with 'check_success' (bool) and 'error_details' (string). 'check_success' is true if 'go list -e -json' finds no errors in the target's JSON output. 'error_details' contains messages if errors are found or if the command fails.

**Example:**
```neuroscript
tool.gotools.Check(target: "./pkg/core")
```
---

## `tool.gotools.Fmt`
**Description:** Formats Go source code using 'go/format.Source'. Returns the formatted code or an error map.

**Category:** Go Formatting

**Parameters:**
* `content` (`string`): The Go source code content to format.

**Returns:** (`string`) Returns the formatted Go source code as a string. If formatting fails (e.g., syntax error), returns a map {'formatted_content': <original_content>, 'error': <error_string>, 'success': false} and a Go-level error.

**Example:**
```neuroscript
tool.gotools.Fmt(content: "package main\nfunc main(){}")
```
---

## `tool.gotools.GetModuleInfo`
**Description:** Finds and parses the go.mod file relevant to a directory by searching upwards. Returns a map with module path, go version, root directory, requires, and replaces, or nil if not found.

**Category:** Go Build Tools

**Parameters:**
* `directory` (`string`): (optional) Directory (relative to sandbox) to start searching upwards for go.mod. Defaults to '.' (sandbox root).

**Returns:** (`map`) Returns a map containing 'modulePath', 'goVersion', 'rootDir' (absolute path to module root), 'requires' (list of maps), and 'replaces' (list of maps). Returns nil if no go.mod is found.

**Example:**
```neuroscript
tool.gotools.GetModuleInfo(directory: "cmd/mytool")
```
---

## `tool.gotools.Imports`
**Description:** Formats Go source code and adjusts imports using 'golang.org/x/tools/imports'. Returns the processed code or an error map.

**Category:** Go Formatting

**Parameters:**
* `content` (`string`): The Go source code content to process.

**Returns:** (`string`) Returns the processed Go source code (formatted and with adjusted imports) as a string. If processing fails, returns a map {'formatted_content': <original_content>, 'error': <error_string>, 'success': false} and a Go-level error.

**Example:**
```neuroscript
tool.gotools.Imports(content: "package main\nimport \"fmt\"\nfunc main(){fmt.Println(\"hello\")}")
```
---

## `tool.gotools.ListPackages`
**Description:** Runs 'go list -json' for specified patterns in a target directory. Returns a list of maps, each describing a package.

**Category:** Go Build Tools

**Parameters:**
* `target_directory` (`string`): (optional) Optional. The directory relative to the sandbox root to run 'go list'. Defaults to '.' (sandbox root).
* `patterns` (`slice_string`): (optional) Optional. A list of package patterns (e.g., './...', 'example.com/project/...'). Defaults to ['./...'].

**Returns:** (`slice_map`) Returns a slice of maps, where each map is a JSON object representing a Go package as output by 'go list -json'. Returns an empty slice on command failure or if JSON decoding fails.

**Example:**
```neuroscript
tool.gotools.ListPackages(target_directory: "pkg/core", patterns: ["./..."])
```
---

## `tool.gotools.ModTidy`
**Description:** Runs 'go mod tidy' in the sandbox to add missing and remove unused modules. Operates in the sandbox root.

**Category:** Go Build Tools

**Parameters:**
_None_

**Returns:** (`map`) Returns a map with 'stdout', 'stderr', 'exit_code' (int64), and 'success' (bool) from the 'go mod tidy' command execution.

**Example:**
```neuroscript
tool.gotools.ModTidy()
```
---

## `tool.gotools.Staticcheck`
**Description:** Runs 'staticcheck' on the specified target(s) in the sandbox. Reports bugs, stylistic errors, and performance issues. Defaults to './...'. Assumes 'staticcheck' is in PATH.

**Category:** Go types.Diagnostics

**Parameters:**
* `target` (`string`): (optional) Optional. The target for 'staticcheck' (e.g., a package path or './...'). Defaults to './...'.

**Returns:** (`map`) Returns a map with 'stdout', 'stderr', 'exit_code' (int64), and 'success' (bool) from the 'staticcheck <target>' command. 'stdout' usually contains the diagnostics.

**Example:**
```neuroscript
tool.gotools.Staticcheck(target: "./...")
```
---

## `tool.gotools.Test`
**Description:** Runs 'go test' for a specified target in the sandbox. Defaults to './...'.

**Category:** Go Build Tools

**Parameters:**
* `target` (`string`): (optional) Optional. The test target (e.g., a package path or './...'). Defaults to './...'.

**Returns:** (`map`) Returns a map with 'stdout', 'stderr', 'exit_code' (int64), and 'success' (bool) from the 'go test <target>' command.

**Example:**
```neuroscript
tool.gotools.Test(target: "./pkg/feature")
```
---

## `tool.gotools.Vet`
**Description:** Runs 'go vet' on the specified target(s) in the sandbox to report likely mistakes in Go source code. Defaults to './...'.

**Category:** Go types.Diagnostics

**Parameters:**
* `target` (`string`): (optional) Optional. The target for 'go vet' (e.g., a package path or './...'). Defaults to './...'.

**Returns:** (`map`) Returns a map with 'stdout', 'stderr', 'exit_code' (int64), and 'success' (bool) from the 'go vet <target>' command. 'stderr' usually contains the vet diagnostics.

**Example:**
```neuroscript
tool.gotools.Vet(target: "./pkg/core")
```
---

## `tool.io.Input`
**Description:** Displays a message and waits for user input from standard input. Returns the input as a string.

**Category:** Input/Output

**Parameters:**
* `message` (`string`): (optional) The message to display to the user before waiting for input. If null or empty, no prompt message is printed.

**Returns:** (`string`) Returns the string entered by the user, with trailing newline characters trimmed. Returns an empty string and an error if reading input fails.

**Example:**
```neuroscript
userName = TOOL.Input(message: "Enter your name: ")
```
---

## `tool.io.Print`
**Description:** Prints values to the standard output. If multiple values are passed in a list, they are printed space-separated.

**Category:** Input/Output

**Parameters:**
* `values` (`any`): A single value or a list of values to print. List elements will be space-separated.

**Returns:** (`nil`) Returns nil. This tool is used for its side effect of printing to standard output.

**Example:**
```neuroscript
TOOL.Print(value: "Hello World")\nTOOL.Print(values: ["Hello", 42, "World!"]) // Prints "Hello 42 World!"
```
---

## `tool.list.Append`
**Description:** Returns a *new* list with the given element added to the end.

**Category:** List Operations

**Parameters:**
* `list` (`slice_any`): The list to append to.
* `element` (`any`): (optional) The element to append (can be nil).

**Returns:** (`slice_any`) Returns a new list with the 'element' added to the end of the input 'list'. The original list is not modified.

**Example:**
```neuroscript
tool.List.Append([1, 2], 3) // returns [1, 2, 3]
```
---

## `tool.list.Contains`
**Description:** Checks if a list contains a specific element (using deep equality comparison).

**Category:** List Operations

**Parameters:**
* `list` (`slice_any`): The list to search within.
* `element` (`any`): (optional) The element to search for (can be nil).

**Returns:** (`bool`) Returns true if the 'list' contains the specified 'element', using deep equality for comparison. Returns false otherwise.

**Example:**
```neuroscript
tool.List.Contains([1, "a", true], "a") // returns true
```
---

## `tool.list.Get`
**Description:** Safely gets the element at a specific index (0-based). Returns nil or the optional default value if the index is out of bounds.

**Category:** List Operations

**Parameters:**
* `list` (`slice_any`): The list to get from.
* `index` (`int`): The 0-based index.
* `default` (`any`): (optional) Optional default value if index is out of bounds.

**Returns:** (`any`) Returns the element at the specified 0-based 'index' in the 'list'. If the index is out of bounds, it returns the provided 'default' value. If no 'default' is provided and the index is out of bounds, it returns nil.

**Example:**
```neuroscript
tool.List.Get(["a", "b", "c"], 1) // returns "b"\n tool.List.Get(["a"], 5, "default_val") // returns "default_val"
```
---

## `tool.list.Head`
**Description:** Returns the first element of the list, or nil if the list is empty.

**Category:** List Operations

**Parameters:**
* `list` (`slice_any`): The list to get the head from.

**Returns:** (`any`) Returns the first element of the 'list'. If the list is empty, it returns nil.

**Example:**
```neuroscript
tool.List.Head([1, 2, 3]) // returns 1\ntool.List.Head([]) // returns nil
```
---

## `tool.list.IsEmpty`
**Description:** Returns true if the list has zero elements, false otherwise.

**Category:** List Operations

**Parameters:**
* `list` (`slice_any`): The list to check.

**Returns:** (`bool`) Returns true if the 'list' contains zero elements, and false otherwise.

**Example:**
```neuroscript
tool.List.IsEmpty([]) // returns true\ntool.List.IsEmpty([1]) // returns false
```
---

## `tool.list.Length`
**Description:** Returns the number of elements in a list.

**Category:** List Operations

**Parameters:**
* `list` (`slice_any`): The list to measure.

**Returns:** (`int`) Returns an integer representing the number of elements in the provided list.

**Example:**
```neuroscript
tool.List.Length([1, 2, 3]) // returns 3
```
---

## `tool.list.Prepend`
**Description:** Returns a *new* list with the given element added to the beginning.

**Category:** List Operations

**Parameters:**
* `list` (`slice_any`): The list to prepend to.
* `element` (`any`): (optional) The element to prepend (can be nil).

**Returns:** (`slice_any`) Returns a new list with the 'element' added to the beginning of the input 'list'. The original list is not modified.

**Example:**
```neuroscript
tool.List.Prepend([2, 3], 1) // returns [1, 2, 3]
```
---

## `tool.list.Rest`
**Description:** Returns a *new* list containing all elements except the first. Returns an empty list if the input list has 0 or 1 element.

**Category:** List Operations

**Parameters:**
* `list` (`slice_any`): The list to get the rest from.

**Returns:** (`slice_any`) Returns a new list containing all elements of the input 'list' except the first. If the list has 0 or 1 element, it returns an empty list. The original list is not modified.

**Example:**
```neuroscript
tool.List.Rest([1, 2, 3]) // returns [2, 3]\ntool.List.Rest([1]) // returns []
```
---

## `tool.list.Reverse`
**Description:** Returns a *new* list with the elements in reverse order.

**Category:** List Operations

**Parameters:**
* `list` (`slice_any`): The list to reverse.

**Returns:** (`slice_any`) Returns a new list with the elements of the input 'list' in reverse order. The original list is not modified.

**Example:**
```neuroscript
tool.List.Reverse([1, 2, 3]) // returns [3, 2, 1]
```
---

## `tool.list.Slice`
**Description:** Returns a *new* list containing elements from the start index (inclusive) up to the end index (exclusive). Follows Go slice semantics (indices are clamped, invalid range returns empty list).

**Category:** List Operations

**Parameters:**
* `list` (`slice_any`): The list to slice.
* `start` (`int`): The starting index (inclusive).
* `end` (`int`): The ending index (exclusive).

**Returns:** (`slice_any`) Returns a new list containing elements from the 'start' index (inclusive) up to the 'end' index (exclusive). Adheres to Go's slice semantics: indices are clamped to valid ranges (0 to list length). If 'start' > 'end' after clamping, or if 'start' is out of bounds (e.g. beyond list length), an empty list is returned. The original list is not modified.

**Example:**
```neuroscript
tool.List.Slice([1, 2, 3, 4, 5], 1, 4) // returns [2, 3, 4]
```
---

## `tool.list.Sort`
**Description:** Returns a *new* list with elements sorted. Restricted to lists containing only numbers (int/float) or only strings. Throws error for mixed types or non-sortable types (nil, bool, list, map).

**Category:** List Operations

**Parameters:**
* `list` (`slice_any`): The list to sort.

**Returns:** (`slice_any`) Returns a new list with elements sorted. The list must contain either all numbers (integers or floats, which will be sorted numerically) or all strings (sorted lexicographically). The original list is not modified. Returns an empty list if the input list is empty.

**Example:**
```neuroscript
tool.List.Sort([3, 1, 2]) // returns [1, 2, 3]\ntool.List.Sort(["c", "a", "b"]) // returns ["a", "b", "c"]
```
---

## `tool.list.Tail`
**Description:** Returns a *new* list containing the last 'count' elements. Returns an empty list if count <= 0. Returns a copy of the whole list if count >= list length.

**Category:** List Operations

**Parameters:**
* `list` (`slice_any`): The list to get the tail from.
* `count` (`int`): The number of elements to take from the end.

**Returns:** (`slice_any`) Returns a new list containing the last 'count' elements from the input 'list'. If 'count' is less than or equal to 0, an empty list is returned. If 'count' is greater than or equal to the list length, a copy of the original list is returned. The original list is not modified.

**Example:**
```neuroscript
tool.List.Tail([1, 2, 3, 4, 5], 3) // returns [3, 4, 5]\ntool.List.Tail([1, 2], 5) // returns [1, 2]
```
---

## `tool.math.Add`
**Description:** Calculates the sum of two numbers (integers or decimals). Strings convertible to numbers are accepted.

**Category:** Math Operations

**Parameters:**
* `num1` (`float`): The first number (or numeric string) to add.
* `num2` (`float`): The second number (or numeric string) to add.

**Returns:** (`float`) Returns the sum of num1 and num2 as a float64. Both inputs are expected to be (or be coercible to) numbers.

**Example:**
```neuroscript
tool.Add(5, 3.5) // returns 8.5
```
---

## `tool.math.Divide`
**Description:** Calculates the division of two numbers (num1 / num2). Returns float. Handles division by zero.

**Category:** Math Operations

**Parameters:**
* `num1` (`float`): The dividend.
* `num2` (`float`): The divisor.

**Returns:** (`float`) Returns the result of num1 / num2 as a float64. Both inputs are expected to be (or be coercible to) numbers.

**Example:**
```neuroscript
tool.Divide(10, 4) // returns 2.5
```
---

## `tool.math.Modulo`
**Description:** Calculates the modulo (remainder) of two integers (num1 % num2). Handles division by zero.

**Category:** Math Operations

**Parameters:**
* `num1` (`int`): The dividend (must be integer).
* `num2` (`int`): The divisor (must be integer).

**Returns:** (`int`) Returns the remainder of num1 % num2 as an int64. Both inputs must be integers.

**Example:**
```neuroscript
tool.Modulo(10, 3) // returns 1
```
---

## `tool.math.Multiply`
**Description:** Calculates the product of two numbers. Strings convertible to numbers are accepted.

**Category:** Math Operations

**Parameters:**
* `num1` (`float`): The first number.
* `num2` (`float`): The second number.

**Returns:** (`float`) Returns the product of num1 and num2 as a float64. Both inputs are expected to be (or be coercible to) numbers.

**Example:**
```neuroscript
tool.Multiply(6, 7.0) // returns 42.0
```
---

## `tool.math.Subtract`
**Description:** Calculates the difference between two numbers (num1 - num2). Strings convertible to numbers are accepted.

**Category:** Math Operations

**Parameters:**
* `num1` (`float`): The number to subtract from.
* `num2` (`float`): The number to subtract.

**Returns:** (`float`) Returns the difference of num1 - num2 as a float64. Both inputs are expected to be (or be coercible to) numbers.

**Example:**
```neuroscript
tool.Subtract(10, 4.5) // returns 5.5
```
---

## `tool.script.ListFunctions`
**Description:** Returns a map of all currently loaded function (procedure) names to their signatures.

**Category:** Scripting

**Parameters:**
_None_

**Returns:** (`map`) Returns a map where each key is the name of a known function and the value is its signature.

**Example:**
```neuroscript
set loaded_functions = tool.script.ListFunctions()
```
---

## `tool.script.LoadScript`
**Description:** Parses a string of NeuroScript code and loads its functions and event handlers into the current interpreter's scope. Does not execute any code.

**Category:** Scripting

**Parameters:**
* `script_content` (`string`): A string containing the NeuroScript code to load.

**Returns:** (`map`) Returns a map with keys 'functions_loaded', 'event_handlers_loaded', and 'metadata', which contains the file-level metadata from the script header.

**Example:**
```neuroscript
set result = tool.script.LoadScript(":: purpose: example\nfunc f()means\nendfunc")\nemit result["metadata"]["purpose"]
```
---

## `tool.shell.Execute`
**Description:** Executes an arbitrary shell command. WARNING: Use with extreme caution due to security risks. Command path validation is basic. Consider using specific tools (e.g., GoBuild, GitAdd) instead.

**Category:** Shell Operations

**Parameters:**
* `command` (`string`): The command or executable path (must not contain path separators like '/' or '\').
* `args_list` (`slice_string`): (optional) A list of string arguments for the command.
* `directory` (`string`): (optional) Optional directory (relative to sandbox) to execute the command in. Defaults to sandbox root.

**Returns:** (`map`) Returns a map containing 'stdout' (string), 'stderr' (string), 'exit_code' (int), and 'success' (bool) of the executed command. 'success' is true if the command exits with code 0, false otherwise. The command is executed within the sandboxed environment.

**Example:**
```neuroscript
tool.shell.Execute("ls", ["-la"], "my_directory")
```
---

## `tool.str.Concat`
**Description:** Concatenates a list of strings without a separator.

**Category:** String Operations

**Parameters:**
* `strings_list` (`slice_string`): List of strings to concatenate.

**Returns:** (`string`) Returns a single string by concatenating all strings in the strings_list.

**Example:**
```neuroscript
tool.Concat(["hello", " ", "world"]) // Returns "hello world"
```
---

## `tool.str.Contains`
**Description:** Checks if a string contains a substring.

**Category:** String Operations

**Parameters:**
* `input_string` (`string`): The string to check.
* `substring` (`string`): The substring to search for.

**Returns:** (`bool`) Returns true if the input_string contains the substring, false otherwise.

**Example:**
```neuroscript
tool.Contains("hello world", "world") // Returns true
```
---

## `tool.str.HasPrefix`
**Description:** Checks if a string starts with a prefix.

**Category:** String Operations

**Parameters:**
* `input_string` (`string`): The string to check.
* `prefix` (`string`): The prefix to check for.

**Returns:** (`bool`) Returns true if the input_string starts with the prefix, false otherwise.

**Example:**
```neuroscript
tool.HasPrefix("filename.txt", "filename") // Returns true
```
---

## `tool.str.HasSuffix`
**Description:** Checks if a string ends with a suffix.

**Category:** String Operations

**Parameters:**
* `input_string` (`string`): The string to check.
* `suffix` (`string`): The suffix to check for.

**Returns:** (`bool`) Returns true if the input_string ends with the suffix, false otherwise.

**Example:**
```neuroscript
tool.HasSuffix("document.doc", ".doc") // Returns true
```
---

## `tool.str.Join`
**Description:** Joins elements of a list of strings with a separator.

**Category:** String Operations

**Parameters:**
* `string_list` (`slice_string`): List of strings to join.
* `separator` (`string`): String to place between elements.

**Returns:** (`string`) Returns a single string created by joining the elements of string_list with the separator.

**Example:**
```neuroscript
tool.Join(["apple", "banana"], ", ") // Returns "apple, banana"
```
---

## `tool.str.Length`
**Description:** Returns the number of UTF-8 characters (runes) in a string.

**Category:** String Operations

**Parameters:**
* `input_string` (`string`): The string to measure.

**Returns:** (`int`) Returns an integer representing the number of runes in the input string.

**Example:**
```neuroscript
tool.Length("hello") // Returns 5
```
---

## `tool.str.LineCount`
**Description:** Counts the number of lines in the given string content.

**Category:** String Operations

**Parameters:**
* `content_string` (`string`): The string content in which to count lines.

**Returns:** (`int`) Returns an integer representing the number of lines in the string. Lines are typically separated by '\n'. An empty string results in 0 lines. If the string is not empty and does not end with a newline, the last line is still counted.

**Example:**
```neuroscript
tool.LineCount("line1\nline2\nline3") // Returns 3
tool.LineCount("line1\nline2") // Returns 2
tool.LineCount("") // Returns 0
```
---

## `tool.str.Replace`
**Description:** Replaces occurrences of a substring with another, up to a specified count.

**Category:** String Operations

**Parameters:**
* `input_string` (`string`): The string to perform replacements on.
* `old_substring` (`string`): The substring to be replaced.
* `new_substring` (`string`): The substring to replace with.
* `count` (`int`): Maximum number of replacements. Use -1 for all.

**Returns:** (`string`) Returns the string with specified replacements made.

**Example:**
```neuroscript
tool.Replace("ababab", "ab", "cd", 2) // Returns "cdcdab"
```
---

## `tool.str.Split`
**Description:** Splits a string by a delimiter.

**Category:** String Operations

**Parameters:**
* `input_string` (`string`): The string to split.
* `delimiter` (`string`): The delimiter string.

**Returns:** (`slice_string`) Returns a slice of strings after splitting the input string by the delimiter.

**Example:**
```neuroscript
tool.Split("apple,banana,orange", ",") // Returns ["apple", "banana", "orange"]
```
---

## `tool.str.SplitWords`
**Description:** Splits a string into words based on whitespace.

**Category:** String Operations

**Parameters:**
* `input_string` (`string`): The string to split into words.

**Returns:** (`slice_string`) Returns a slice of strings, where each string is a word from the input string, with whitespace removed. Multiple spaces are treated as a single delimiter.

**Example:**
```neuroscript
tool.SplitWords("hello world  example") // Returns ["hello", "world", "example"]
```
---

## `tool.str.Substring`
**Description:** Returns a portion of the string (rune-based indexing), from start_index for a given length.

**Category:** String Operations

**Parameters:**
* `input_string` (`string`): The string to take a substring from.
* `start_index` (`int`): 0-based start index (inclusive).
* `length` (`int`): Number of characters to extract.

**Returns:** (`string`) Returns the specified substring (rune-based). Returns an empty string if length is zero or if start_index is out of bounds (after clamping). Gracefully handles out-of-bounds for non-negative start_index and length by returning available characters.

**Example:**
```neuroscript
tool.Substring("hello world", 6, 5) // Returns "world"
```
---

## `tool.str.ToLower`
**Description:** Converts a string to lowercase.

**Category:** String Operations

**Parameters:**
* `input_string` (`string`): The string to convert.

**Returns:** (`string`) Returns the lowercase version of the input string.

**Example:**
```neuroscript
tool.ToLower("HELLO") // Returns "hello"
```
---

## `tool.str.ToUpper`
**Description:** Converts a string to uppercase.

**Category:** String Operations

**Parameters:**
* `input_string` (`string`): The string to convert.

**Returns:** (`string`) Returns the uppercase version of the input string.

**Example:**
```neuroscript
tool.ToUpper("hello") // Returns "HELLO"
```
---

## `tool.str.TrimSpace`
**Description:** Removes leading and trailing whitespace from a string.

**Category:** String Operations

**Parameters:**
* `input_string` (`string`): The string to trim.

**Returns:** (`string`) Returns the string with leading and trailing whitespace removed.

**Example:**
```neuroscript
tool.TrimSpace("  hello  ") // Returns "hello"
```
---

## `tool.syntax.analyzeNSSyntax`
**Description:** Analyzes a NeuroScript string for syntax errors. Returns a list of maps, where each map details an error. Returns an empty list if no errors are found.

**Category:** Syntax Utilities

**Parameters:**
* `nsScriptContent` (`string`): The NeuroScript content to analyze.

**Returns:** (`slice_map`) Returns a list (slice) of maps. Each map represents a syntax error and contains the following keys:
- `Line`: number (1-based) - The line number of the error.
- `Column`: number (0-based) - The character types.Position in the line where the error occurred.
- `Msg`: string - The error message.
- `OffendingSymbol`: string - The text of the token that caused the error (may be empty).
- `SourceName`: string - Identifier for the source (e.g., 'nsSyntaxAnalysisToolInput').
An empty list is returned if no syntax errors are found.

**Example:**
```neuroscript
set script_to_check = `func myFunc means
  set x = 
endfunc`
set error_list = tool.analyzeNSSyntax(script_to_check)
if tool.List.IsEmpty(error_list) == false
  set first_error = tool.List.Get(error_list, 0)
  emit "First error on line " + first_error["Line"] + ": " + first_error["Msg"]
endif
```
---

## `tool.time.Now`
**Description:** Returns the current system time as a 'timedate' value.

**Category:** Time

**Parameters:**
_None_

**Returns:** (`timedate`) A 'timedate' value representing the moment the tool was called.

**Example:**
```neuroscript
`set right_now = tool.Time.Now()`
```
---

## `tool.time.Sleep`
**Description:** Pauses the script execution for a specified duration.

**Category:** Time

**Parameters:**
* `duration_seconds` (`number`): (optional) The number of seconds to sleep (can be a fraction).

**Returns:** (`boolean`) Returns true on successful completion of the sleep duration.

**Example:**
```neuroscript
`call tool.Time.Sleep(1.5)`
```
---

## `tool.tree.AddChildNode`
**Description:** Adds a new child node to an existing parent node.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle for the tree structure.
* `parent_node_id` (`string`): ID of the node that will become the parent.
* `new_node_id_suggestion` (`string`): (optional) Optional suggested unique ID for the new node.
* `node_type` (`string`): Type of the new child (e.g., 'object', 'array', 'string').
* `value` (`any`): (optional) Initial value for simple types.
* `key_for_object_parent` (`string`): (optional) Required if the parent is an 'object' node.

**Returns:** (`string`) Returns the string ID of the newly created child node.

**Example:**
```neuroscript
tool.Tree.AddChildNode(handle, "root_id", "newChild", "string", "hello", "message")
```
---

## `tool.tree.FindNodes`
**Description:** Finds nodes within a tree that match specific criteria.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.
* `start_node_id` (`string`): ID of the node to start searching from.
* `query_map` (`map`): Map defining search criteria.
* `max_depth` (`int`): (optional) Maximum depth to search.
* `max_results` (`int`): (optional) Maximum number of results to return.

**Returns:** (`slice_string`) Returns a slice of node IDs matching the query.

**Example:**
```neuroscript
tool.Tree.FindNodes(handle, "start_node_id", {\"type\":\"file\"})
```
---

## `tool.tree.GetChildren`
**Description:** Gets a list of node IDs of the children of a given 'array' type node.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.
* `node_id` (`string`): ID of the 'array' type parent node.

**Returns:** (`slice_string`) Returns a slice of child node IDs.

**Example:**
```neuroscript
tool.Tree.GetChildren(handle, "array_node_id")
```
---

## `tool.tree.GetNode`
**Description:** Retrieves detailed information about a specific node within a tree, returned as a map.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.
* `node_id` (`string`): The unique ID of the node to retrieve.

**Returns:** (`map`) Returns a map containing details of the specified node.

**Example:**
```neuroscript
tool.Tree.GetNode(handle, "node_id_123")
```
---

## `tool.tree.GetNodeByPath`
**Description:** Retrieves a node from a tree using a dot-separated path expression.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.
* `path` (`string`): Dot-separated path (e.g., 'key.0.name').

**Returns:** (`map`) Returns a map containing details of the node found at the specified path.

**Example:**
```neuroscript
tool.Tree.GetNodeByPath(handle, "data.users.1")
```
---

## `tool.tree.GetNodeMetadata`
**Description:** Retrieves the metadata attributes of a specific node as a map.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.
* `node_id` (`string`): ID of the node to get metadata from.

**Returns:** (`map`) Returns a map of the node's metadata attributes.

**Example:**
```neuroscript
tool.Tree.GetNodeMetadata(handle, "node_id")
```
---

## `tool.tree.GetParent`
**Description:** Gets the parent of a given node as a map.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.
* `node_id` (`string`): ID of the node whose parent is sought.

**Returns:** (`map`) Returns a map of the parent node, or nil if the node is the root.

**Example:**
```neuroscript
tool.Tree.GetParent(handle, "child_node_id")
```
---

## `tool.tree.GetRoot`
**Description:** Retrieves the root node of the tree as a map.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.

**Returns:** (`map`) Returns a map containing details of the root node.

**Example:**
```neuroscript
handle = tool.Tree.LoadJSON("{}"); tool.Tree.GetRoot(handle)
```
---

## `tool.tree.LoadJSON`
**Description:** Loads a JSON string into a new tree structure and returns a tree handle.

**Category:** Tree Manipulation

**Parameters:**
* `json_string` (`string`): The JSON data as a string.

**Returns:** (`string`) Returns a string handle representing the loaded tree.

**Example:**
```neuroscript
tool.Tree.LoadJSON("{\"name\": \"example\"}")
```
---

## `tool.tree.RemoveNode`
**Description:** Removes a node and all its descendants from the tree.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle to the tree.
* `node_id` (`string`): ID of the node to remove.

**Returns:** (`nil`) Returns nil on success.

**Example:**
```neuroscript
tool.Tree.RemoveNode(handle, "node_to_delete_id")
```
---

## `tool.tree.RemoveNodeMetadata`
**Description:** Removes a metadata attribute from a node.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.
* `node_id` (`string`): ID of the node to remove metadata from.
* `metadata_key` (`string`): The key of the metadata attribute to remove.

**Returns:** (`nil`) Returns nil on success.

**Example:**
```neuroscript
tool.Tree.RemoveNodeMetadata(handle, "my_node_id", "version")
```
---

## `tool.tree.RemoveObjectAttribute`
**Description:** Removes an attribute from an 'object' type node.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle for the tree structure.
* `object_node_id` (`string`): Unique ID of the 'object' type node to modify.
* `attribute_key` (`string`): The key of the attribute to remove.

**Returns:** (`nil`) Returns nil on success.

**Example:**
```neuroscript
tool.Tree.RemoveObjectAttribute(handle, "obj_id", "myChild")
```
---

## `tool.tree.RenderText`
**Description:** Renders a visual text representation of the entire tree structure.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure to render.

**Returns:** (`string`) Returns a human-readable, indented text representation of the tree.

**Example:**
```neuroscript
tool.Tree.RenderText(handle)
```
---

## `tool.tree.SetNodeMetadata`
**Description:** Sets a metadata attribute as a key-value string pair on any node.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.
* `node_id` (`string`): ID of the node to set metadata on.
* `metadata_key` (`string`): The key of the metadata attribute (string).
* `metadata_value` (`string`): The value of the metadata attribute (string).

**Returns:** (`nil`) Returns nil on success.

**Example:**
```neuroscript
tool.Tree.SetNodeMetadata(handle, "my_node_id", "version", "1.0")
```
---

## `tool.tree.SetObjectAttribute`
**Description:** Sets or updates an attribute on an 'object' type node.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle for the tree structure.
* `object_node_id` (`string`): Unique ID of the 'object' type node to modify.
* `attribute_key` (`string`): The key of the attribute to set.
* `child_node_id` (`string`): The ID of an existing node to associate with the key.

**Returns:** (`nil`) Returns nil on success.

**Example:**
```neuroscript
tool.Tree.SetObjectAttribute(handle, "obj_id", "myChild", "child_id")
```
---

## `tool.tree.SetValue`
**Description:** Sets the value of an existing leaf or simple-type node.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.
* `node_id` (`string`): ID of the leaf or simple-type node to modify.
* `value` (`any`): The new value for the node.

**Returns:** (`nil`) Returns nil on success.

**Example:**
```neuroscript
tool.Tree.SetValue(handle, "id_of_keyNode", "new_value")
```
---

## `tool.tree.ToJSON`
**Description:** Converts a tree structure back into a pretty-printed JSON string.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.

**Returns:** (`string`) Returns a pretty-printed JSON string representation of the tree.

**Example:**
```neuroscript
handle = tool.Tree.LoadJSON("{\"key\":\"value\"}"); tool.Tree.ToJSON(handle)
```
---