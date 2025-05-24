Compact Tool List:
AIWorker.ExecuteStatelessTask(definition_id:string, prompt:string, config_overrides:map?) -> map
AIWorker.GetPerformanceRecords(definition_id:string, filters:map?) -> slice_map
AIWorker.LoadPerformanceData() -> string
AIWorker.LogPerformance(task_id:string, instance_id:string, definition_id:string, timestamp_start:string, timestamp_end:string, duration_ms:int, success:bool, input_context:map?, llm_metrics:map?, cost_incurred:float?, output_summary:string?, error_details:string?) -> string
AIWorker.SavePerformanceData() -> string
AIWorkerDefinition.Add(definition_id:string?, name:string?, provider:string, model_name:string, auth:map, interaction_models:slice_string?, capabilities:slice_string?, base_config:map?, cost_metrics:map?, rate_limits:map?, status:string?, default_file_contexts:slice_string?, metadata:map?) -> string
AIWorkerDefinition.Get(definition_id:string) -> map
AIWorkerDefinition.List(filters:map?) -> slice_map
AIWorkerDefinition.LoadAll() -> string
AIWorkerDefinition.Remove(definition_id:string) -> nil
AIWorkerDefinition.SaveAll() -> string
AIWorkerDefinition.Update(definition_id:string, updates:map) -> nil
AIWorkerInstance.Get(instance_id:string) -> map
AIWorkerInstance.ListActive(filters:map?) -> slice_map
AIWorkerInstance.Retire(instance_id:string, conversation_manager_handle:string, reason:string, final_status:string, final_session_token_usage:map, performance_records:slice_map?) -> nil
AIWorkerInstance.Spawn(definition_id:string, config_overrides:map?, file_contexts:slice_string?) -> map
AIWorkerInstance.UpdateStatus(instance_id:string, status:string, last_error:string?) -> nil
AIWorkerInstance.UpdateTokenUsage(instance_id:string, input_tokens:int, output_tokens:int) -> nil
Add(num1:float, num2:float) -> float
Concat(strings_list:slice_string) -> string
Contains(input_string:string, substring:string) -> bool
DeleteAPIFile(api_file_id:string) -> string
Divide(num1:float, num2:float) -> float
FS.Delete(path:string) -> string
FS.Hash(filepath:string) -> string
FS.LineCount(filepath:string) -> int
FS.List(path:string, recursive:bool?) -> slice_any
FS.Mkdir(path:string) -> string
FS.Move(source_path:string, destination_path:string) -> map
FS.Read(filepath:string) -> string
FS.SanitizeFilename(name:string) -> string
FS.Stat(path:string) -> map
FS.Walk(path:string) -> slice_any
FS.Write(filepath:string, content:string) -> string
Git.Add(paths:any) -> string
Git.Branch(name:string?, checkout:bool?, list_remote:bool?, list_all:bool?) -> any
Git.Checkout(branch:string, create:bool?) -> string
Git.Commit(message:string, add_all:bool?) -> string
Git.Diff(cached:bool?, commit1:string?, commit2:string?, path:string?) -> string
Git.Merge(branch:string) -> string
Git.Pull() -> string
Git.Push(remote:string?, branch:string?, set_upstream:bool?) -> string
Git.Rm(paths:any) -> string
Git.Status() -> map
Go.Build(target:string?) -> map
Go.Check(target:string) -> map
Go.Fmt(content:string) -> string
Go.GetModuleInfo(directory:string?) -> map
Go.Imports(content:string) -> string
Go.ListPackages(target_directory:string?, patterns:slice_string?) -> slice_map
Go.ModTidy() -> map
Go.Test(target:string?) -> map
Go.Vet(target:string?) -> map
HasPrefix(input_string:string, prefix:string) -> bool
HasSuffix(input_string:string, suffix:string) -> bool
Input(prompt:string?) -> string
Join(string_list:slice_string, separator:string) -> string
Length(input_string:string) -> int
LineCount(content_string:string) -> int
List.Append(list:slice_any, element:any?) -> slice_any
List.Contains(list:slice_any, element:any?) -> bool
List.Get(list:slice_any, index:int, default:any?) -> any
List.Head(list:slice_any) -> any
List.IsEmpty(list:slice_any) -> bool
List.Length(list:slice_any) -> int
List.Prepend(list:slice_any, element:any?) -> slice_any
List.Rest(list:slice_any) -> slice_any
List.Reverse(list:slice_any) -> slice_any
List.Slice(list:slice_any, start:int, end:int) -> slice_any
List.Sort(list:slice_any) -> slice_any
List.Tail(list:slice_any, count:int) -> slice_any
ListAPIFiles() -> slice_any
Meta.ListTools() -> string
Meta.ToolsHelp(filter:string?) -> string
Modulo(num1:int, num2:int) -> int
Multiply(num1:float, num2:float) -> float
Print(values:any) -> nil
Replace(input_string:string, old_substring:string, new_substring:string, count:int) -> string
Shell.Execute(command:string, args_list:slice_string?, directory:string?) -> map
Split(input_string:string, delimiter:string) -> slice_string
SplitWords(input_string:string) -> slice_string
Staticcheck(target:string?) -> map
Substring(input_string:string, start_index:int, length:int) -> string
Subtract(num1:float, num2:float) -> float
SyncFiles(direction:string, local_dir:string, filter_pattern:string?, ignore_gitignore:bool?) -> map
ToLower(input_string:string) -> string
ToUpper(input_string:string) -> string
Tree.AddChildNode(tree_handle:string, parent_node_id:string, new_node_id_suggestion:string?, node_type:string, value:any?, key_for_object_parent:string?) -> string
Tree.FindNodes(tree_handle:string, start_node_id:string, query_map:map, max_depth:int?, max_results:int?) -> slice_string
Tree.GetChildren(tree_handle:string, node_id:string) -> slice_string
Tree.GetNode(tree_handle:string, node_id:string) -> map
Tree.GetParent(tree_handle:string, node_id:string) -> string
Tree.LoadJSON(json_string:string) -> string
Tree.RemoveNode(tree_handle:string, node_id:string) -> nil
Tree.RemoveNodeMetadata(tree_handle:string, node_id:string, metadata_key:string) -> nil
Tree.RemoveObjectAttribute(tree_handle:string, object_node_id:string, attribute_key:string) -> nil
Tree.RenderText(tree_handle:string) -> string
Tree.SetNodeMetadata(tree_handle:string, node_id:string, metadata_key:string, metadata_value:string) -> nil
Tree.SetObjectAttribute(tree_handle:string, object_node_id:string, attribute_key:string, child_node_id:string) -> nil
Tree.SetValue(tree_handle:string, node_id:string, value:any) -> nil
Tree.ToJSON(tree_handle:string) -> string
TrimSpace(input_string:string) -> string
UploadFile(local_filepath:string, api_display_name:string?) -> map


    --------------------
    
Detailed Tool Help (Markdown):
# NeuroScript Tools Help

## `tool.AIWorker.ExecuteStatelessTask`
**Description:** Executes a stateless task using an AI Worker Definition.

**Parameters:**
* `definition_id` (`string`): 
* `prompt` (`string`): 
* `config_overrides` (`map`): (optional) 

**Returns:** (`map`)
---

## `tool.AIWorker.GetPerformanceRecords`
**Description:** Retrieves persisted performance records for a specific AI Worker Definition.

**Parameters:**
* `definition_id` (`string`): 
* `filters` (`map`): (optional) 

**Returns:** (`slice_map`)
---

## `tool.AIWorker.LoadPerformanceData`
**Description:** Reloads all worker definitions, which implicitly re-processes performance summaries from persisted data.

**Parameters:**
_None_

**Returns:** (`string`)
---

## `tool.AIWorker.LogPerformance`
**Description:** Logs a performance record for an AI Worker task.

**Parameters:**
* `task_id` (`string`): 
* `instance_id` (`string`): 
* `definition_id` (`string`): 
* `timestamp_start` (`string`): 
* `timestamp_end` (`string`): 
* `duration_ms` (`int`): 
* `success` (`bool`): 
* `input_context` (`map`): (optional) 
* `llm_metrics` (`map`): (optional) 
* `cost_incurred` (`float`): (optional) 
* `output_summary` (`string`): (optional) 
* `error_details` (`string`): (optional) 

**Returns:** (`string`)
---

## `tool.AIWorker.SavePerformanceData`
**Description:** Explicitly triggers saving of all retired instance performance data. Usually handled automatically on retire.

**Parameters:**
_None_

**Returns:** (`string`)
---

## `tool.AIWorkerDefinition.Add`
**Description:** Adds a new AI Worker Definition. Maps (base_config, etc.) are optional.

**Parameters:**
* `definition_id` (`string`): (optional) 
* `name` (`string`): (optional) 
* `provider` (`string`): 
* `model_name` (`string`): 
* `auth` (`map`): 
* `interaction_models` (`slice_string`): (optional) 
* `capabilities` (`slice_string`): (optional) 
* `base_config` (`map`): (optional) 
* `cost_metrics` (`map`): (optional) 
* `rate_limits` (`map`): (optional) 
* `status` (`string`): (optional) 
* `default_file_contexts` (`slice_string`): (optional) 
* `metadata` (`map`): (optional) 

**Returns:** (`string`)
---

## `tool.AIWorkerDefinition.Get`
**Description:** Retrieves an AI Worker Definition by its ID.

**Parameters:**
* `definition_id` (`string`): 

**Returns:** (`map`)
---

## `tool.AIWorkerDefinition.List`
**Description:** Lists all AI Worker Definitions, optionally filtered.

**Parameters:**
* `filters` (`map`): (optional) 

**Returns:** (`slice_map`)
---

## `tool.AIWorkerDefinition.LoadAll`
**Description:** Reloads all worker definitions from the configured JSON file.

**Parameters:**
_None_

**Returns:** (`string`)
---

## `tool.AIWorkerDefinition.Remove`
**Description:** Removes an AI Worker Definition if it has no active instances.

**Parameters:**
* `definition_id` (`string`): 

**Returns:** (`nil`)
---

## `tool.AIWorkerDefinition.SaveAll`
**Description:** Saves all current worker definitions to the configured JSON file.

**Parameters:**
_None_

**Returns:** (`string`)
---

## `tool.AIWorkerDefinition.Update`
**Description:** Updates fields of an existing AI Worker Definition.

**Parameters:**
* `definition_id` (`string`): 
* `updates` (`map`): 

**Returns:** (`nil`)
---

## `tool.AIWorkerInstance.Get`
**Description:** Retrieves an active AI Worker Instance's details by its ID.

**Parameters:**
* `instance_id` (`string`): 

**Returns:** (`map`)
---

## `tool.AIWorkerInstance.ListActive`
**Description:** Lists currently active AI Worker Instances, optionally filtered.

**Parameters:**
* `filters` (`map`): (optional) 

**Returns:** (`slice_map`)
---

## `tool.AIWorkerInstance.Retire`
**Description:** Retires an active AI Worker Instance, persisting its final state and performance.

**Parameters:**
* `instance_id` (`string`): 
* `conversation_manager_handle` (`string`): 
* `reason` (`string`): 
* `final_status` (`string`): 
* `final_session_token_usage` (`map`): 
* `performance_records` (`slice_map`): (optional) 

**Returns:** (`nil`)
---

## `tool.AIWorkerInstance.Spawn`
**Description:** Spawns a new AI Worker Instance and returns its details including a ConversationManager handle.

**Parameters:**
* `definition_id` (`string`): 
* `config_overrides` (`map`): (optional) 
* `file_contexts` (`slice_string`): (optional) 

**Returns:** (`map`)
---

## `tool.AIWorkerInstance.UpdateStatus`
**Description:** Updates the status and optionally the last error of an active AI Worker Instance.

**Parameters:**
* `instance_id` (`string`): 
* `status` (`string`): 
* `last_error` (`string`): (optional) 

**Returns:** (`nil`)
---

## `tool.AIWorkerInstance.UpdateTokenUsage`
**Description:** Updates the session token usage for an active AI Worker Instance.

**Parameters:**
* `instance_id` (`string`): 
* `input_tokens` (`int`): 
* `output_tokens` (`int`): 

**Returns:** (`nil`)
---

## `tool.Add`
**Description:** Calculates the sum of two numbers (integers or decimals). Strings convertible to numbers are accepted.

**Parameters:**
* `num1` (`float`): The first number (or numeric string) to add.
* `num2` (`float`): The second number (or numeric string) to add.

**Returns:** (`float`)
---

## `tool.Concat`
**Description:** Concatenates a list of strings without a separator.

**Parameters:**
* `strings_list` (`slice_string`): List of strings to concatenate.

**Returns:** (`string`)
---

## `tool.Contains`
**Description:** Checks if a string contains a substring.

**Parameters:**
* `input_string` (`string`): The string to check.
* `substring` (`string`): The substring to search for.

**Returns:** (`bool`)
---

## `tool.DeleteAPIFile`
**Description:** Deletes a specific file from the platform's File API using its ID/URI.

**Parameters:**
* `api_file_id` (`string`): The unique ID or URI of the file on the API (e.g., 'files/abcde123').

**Returns:** (`string`)
---

## `tool.Divide`
**Description:** Calculates the division of two numbers (num1 / num2). Returns float. Handles division by zero.

**Parameters:**
* `num1` (`float`): The dividend.
* `num2` (`float`): The divisor.

**Returns:** (`float`)
---

## `tool.FS.Delete`
**Description:** Deletes a file or an empty directory. Returns 'OK' on success or if path doesn't exist.

**Parameters:**
* `path` (`string`): Relative path to the file or empty directory to delete.

**Returns:** (`string`)
---

## `tool.FS.Hash`
**Description:** Calculates the SHA256 hash of a specified file. Returns the hex-encoded hash string.

**Parameters:**
* `filepath` (`string`): Relative path (within the sandbox) of the file to hash.

**Returns:** (`string`)
---

## `tool.FS.LineCount`
**Description:** Counts lines in a specified file. Returns line count as an integer.

**Parameters:**
* `filepath` (`string`): Relative path to the file.

**Returns:** (`int`)
---

## `tool.FS.List`
**Description:** Lists files and subdirectories at a given path. Returns a list of maps, each describing an entry (keys: name, path, isDir, size, modTime).

**Parameters:**
* `path` (`string`): Relative path to the directory (use '.' for current).
* `recursive` (`bool`): (optional) Whether to list recursively (default: false).

**Returns:** (`slice_any`)
---

## `tool.FS.Mkdir`
**Description:** Creates a directory. Parent directories are created if they do not exist (like mkdir -p). Returns a success message.

**Parameters:**
* `path` (`string`): Relative path of the directory to create.

**Returns:** (`string`)
---

## `tool.FS.Move`
**Description:** Moves or renames a file or directory within the sandbox. Returns a map: {'message': 'success message', 'error': nil} on success.

**Parameters:**
* `source_path` (`string`): Relative path of the source file/directory.
* `destination_path` (`string`): Relative path of the destination.

**Returns:** (`map`)
---

## `tool.FS.Read`
**Description:** Reads the entire content of a specific file. Returns the content as a string.

**Parameters:**
* `filepath` (`string`): Relative path to the file.

**Returns:** (`string`)
---

## `tool.FS.SanitizeFilename`
**Description:** Cleans a string to make it suitable for use as part of a filename.

**Parameters:**
* `name` (`string`): The string to sanitize.

**Returns:** (`string`)
---

## `tool.FS.Stat`
**Description:** Gets information about a file or directory. Returns a map containing: name(string), path(string), size_bytes(int), is_dir(bool), modified_unix(int), modified_rfc3339(string - format 2006-01-02T15:04:05.999999999Z07:00), mode_string(string), mode_perm(string).

**Parameters:**
* `path` (`string`): Relative path to the file or directory.

**Returns:** (`map`)
---

## `tool.FS.Walk`
**Description:** Recursively walks a directory, returning a list of maps describing files/subdirectories found (keys: name, path_relative, is_dir, size_bytes, modified_unix, modified_rfc3339 (format 2006-01-02T15:04:05.999999999Z07:00), mode_string). Skips the root directory itself.

**Parameters:**
* `path` (`string`): Relative path to the directory to walk.

**Returns:** (`slice_any`)
---

## `tool.FS.Write`
**Description:** Writes content to a specific file. Creates parent directories if needed. Returns 'OK' on success.

**Parameters:**
* `filepath` (`string`): Relative path to the file.
* `content` (`string`): The content to write.

**Returns:** (`string`)
---

## `tool.Git.Add`
**Description:** Add file contents to the index. Accepts a single path string or a list of path strings.

**Parameters:**
* `paths` (`any`): A single file path string or a list of file path strings to stage.

**Returns:** (`string`)
---

## `tool.Git.Branch`
**Description:** Lists existing branches or creates a new branch. By default lists local branches.

**Parameters:**
* `name` (`string`): (optional) If provided, create a branch with this name. If omitted, list existing branches.
* `checkout` (`bool`): (optional) If creating a branch (name provided), also check it out immediately (`-b` flag). Default: false.
* `list_remote` (`bool`): (optional) If listing branches (name omitted), list remote branches (`-r`). Default: false.
* `list_all` (`bool`): (optional) If listing branches (name omitted), list all branches (`-a`). Default: false.

**Returns:** (`any`)
---

## `tool.Git.Checkout`
**Description:** Switches branches or restores working tree files. Can also create a new branch before switching.

**Parameters:**
* `branch` (`string`): The name of the branch or commit to check out.
* `create` (`bool`): (optional) If true, create the branch if it doesn't exist (`-b` flag). Default: false.

**Returns:** (`string`)
---

## `tool.Git.Commit`
**Description:** Records changes to the repository.

**Parameters:**
* `message` (`string`): The commit message.
* `add_all` (`bool`): (optional) If true, stage all tracked, modified files (`git add .`) before committing. Default: false.

**Returns:** (`string`)
---

## `tool.Git.Diff`
**Description:** Shows changes between commits, commit and working tree, etc. Returns the diff output or a message indicating no changes.

**Parameters:**
* `cached` (`bool`): (optional) Show diff of staged changes against HEAD (`--cached`). Default: false.
* `commit1` (`string`): (optional) First commit/branch/tree reference. Default: Index.
* `commit2` (`string`): (optional) Second commit/branch/tree reference. Default: Working tree.
* `path` (`string`): (optional) Limit the diff to the specified file or directory path.

**Returns:** (`string`)
---

## `tool.Git.Merge`
**Description:** Join two or more development histories together.

**Parameters:**
* `branch` (`string`): The name of the branch to merge into the current branch.

**Returns:** (`string`)
---

## `tool.Git.Pull`
**Description:** Fetch from and integrate with another repository or a local branch.

**Parameters:**
_None_

**Returns:** (`string`)
---

## `tool.Git.Push`
**Description:** Updates remote refs along with associated objects.

**Parameters:**
* `remote` (`string`): (optional) The remote repository name. Default: 'origin'.
* `branch` (`string`): (optional) The local branch name to push. Default: current branch.
* `set_upstream` (`bool`): (optional) If true, set the upstream tracking configuration (`-u`). Default: false.

**Returns:** (`string`)
---

## `tool.Git.Rm`
**Description:** Remove files from the working tree and from the index.

**Parameters:**
* `paths` (`any`): A single file path string or a list of file path strings to remove.

**Returns:** (`string`)
---

## `tool.Git.Status`
**Description:** Gets the current Git repository status using 'git status --porcelain -b --untracked-files=all' and returns a structured map. Keys: 'branch', 'remote_branch', 'ahead', 'behind', 'files', 'untracked_files_present', 'is_clean', 'error'.

**Parameters:**
_None_

**Returns:** (`map`)
---

## `tool.Go.Build`
**Description:** Runs 'go build' for a specified target in the sandbox. Defaults to './...'.

**Parameters:**
* `target` (`string`): (optional) Optional. The build target (e.g., a package path or './...'). Defaults to './...'.

**Returns:** (`map`)
---

## `tool.Go.Check`
**Description:** Checks Go code validity using 'go list -e -json <target>' within the sandbox. Returns a map indicating success and error details.

**Parameters:**
* `target` (`string`): Target Go package path or file path relative to sandbox (e.g., './pkg/core', 'main.go').

**Returns:** (`map`)
---

## `tool.Go.Fmt`
**Description:** Formats Go source code using 'go/format.Source'. Returns the formatted code or an error map.

**Parameters:**
* `content` (`string`): The Go source code content to format.

**Returns:** (`string`)
---

## `tool.Go.GetModuleInfo`
**Description:** Finds and parses the go.mod file relevant to a directory by searching upwards. Returns a map with module path, go version, root directory, requires, and replaces, or nil if not found.

**Parameters:**
* `directory` (`string`): (optional) Directory (relative to sandbox) to start searching upwards for go.mod. Defaults to '.' (sandbox root).

**Returns:** (`map`)
---

## `tool.Go.Imports`
**Description:** Formats Go source code and adjusts imports using 'golang.org/x/tools/imports'. Returns the processed code or an error map.

**Parameters:**
* `content` (`string`): The Go source code content to process.

**Returns:** (`string`)
---

## `tool.Go.ListPackages`
**Description:** Runs 'go list -json' for specified patterns in a target directory. Returns a list of maps, each describing a package.

**Parameters:**
* `target_directory` (`string`): (optional) Optional. The directory relative to the sandbox root to run 'go list'. Defaults to '.' (sandbox root).
* `patterns` (`slice_string`): (optional) Optional. A list of package patterns (e.g., './...', 'example.com/project/...'). Defaults to ['./...'].

**Returns:** (`slice_map`)
---

## `tool.Go.ModTidy`
**Description:** Runs 'go mod tidy' in the sandbox to add missing and remove unused modules. Operates in the sandbox root.

**Parameters:**
_None_

**Returns:** (`map`)
---

## `tool.Go.Test`
**Description:** Runs 'go test' for a specified target in the sandbox. Defaults to './...'.

**Parameters:**
* `target` (`string`): (optional) Optional. The test target (e.g., a package path or './...'). Defaults to './...'.

**Returns:** (`map`)
---

## `tool.Go.Vet`
**Description:** Runs 'go vet' on the specified target(s) in the sandbox to report likely mistakes in Go source code. Defaults to './...'.

**Parameters:**
* `target` (`string`): (optional) Optional. The target for 'go vet' (e.g., a package path or './...'). Defaults to './...'.

**Returns:** (`map`)
---

## `tool.HasPrefix`
**Description:** Checks if a string starts with a prefix.

**Parameters:**
* `input_string` (`string`): The string to check.
* `prefix` (`string`): The prefix to check for.

**Returns:** (`bool`)
---

## `tool.HasSuffix`
**Description:** Checks if a string ends with a suffix.

**Parameters:**
* `input_string` (`string`): The string to check.
* `suffix` (`string`): The suffix to check for.

**Returns:** (`bool`)
---

## `tool.Input`
**Description:** Reads a single line of text from standard input.

**Parameters:**
* `prompt` (`string`): (optional) Optional prompt message to display to the user.

**Returns:** (`string`)
---

## `tool.Join`
**Description:** Joins elements of a list of strings with a separator.

**Parameters:**
* `string_list` (`slice_string`): List of strings to join.
* `separator` (`string`): String to place between elements.

**Returns:** (`string`)
---

## `tool.Length`
**Description:** Returns the number of UTF-8 characters (runes) in a string.

**Parameters:**
* `input_string` (`string`): The string to measure.

**Returns:** (`int`)
---

## `tool.LineCount`
**Description:** Counts the number of lines in the given string content.

**Parameters:**
* `content_string` (`string`): The string content in which to count lines.

**Returns:** (`int`)
---

## `tool.List.Append`
**Description:** Returns a *new* list with the given element added to the end.

**Parameters:**
* `list` (`slice_any`): The list to append to.
* `element` (`any`): (optional) The element to append (can be nil).

**Returns:** (`slice_any`)
---

## `tool.List.Contains`
**Description:** Checks if a list contains a specific element (using deep equality comparison).

**Parameters:**
* `list` (`slice_any`): The list to search within.
* `element` (`any`): (optional) The element to search for (can be nil).

**Returns:** (`bool`)
---

## `tool.List.Get`
**Description:** Safely gets the element at a specific index (0-based). Returns nil or the optional default value if the index is out of bounds.

**Parameters:**
* `list` (`slice_any`): The list to get from.
* `index` (`int`): The 0-based index.
* `default` (`any`): (optional) Optional default value if index is out of bounds.

**Returns:** (`any`)
---

## `tool.List.Head`
**Description:** Returns the first element of the list, or nil if the list is empty.

**Parameters:**
* `list` (`slice_any`): The list to get the head from.

**Returns:** (`any`)
---

## `tool.List.IsEmpty`
**Description:** Returns true if the list has zero elements, false otherwise.

**Parameters:**
* `list` (`slice_any`): The list to check.

**Returns:** (`bool`)
---

## `tool.List.Length`
**Description:** Returns the number of elements in a list.

**Parameters:**
* `list` (`slice_any`): The list to measure.

**Returns:** (`int`)
---

## `tool.List.Prepend`
**Description:** Returns a *new* list with the given element added to the beginning.

**Parameters:**
* `list` (`slice_any`): The list to prepend to.
* `element` (`any`): (optional) The element to prepend (can be nil).

**Returns:** (`slice_any`)
---

## `tool.List.Rest`
**Description:** Returns a *new* list containing all elements except the first. Returns an empty list if the input list has 0 or 1 element.

**Parameters:**
* `list` (`slice_any`): The list to get the rest from.

**Returns:** (`slice_any`)
---

## `tool.List.Reverse`
**Description:** Returns a *new* list with the elements in reverse order.

**Parameters:**
* `list` (`slice_any`): The list to reverse.

**Returns:** (`slice_any`)
---

## `tool.List.Slice`
**Description:** Returns a *new* list containing elements from the start index (inclusive) up to the end index (exclusive). Follows Go slice semantics (indices are clamped, invalid range returns empty list).

**Parameters:**
* `list` (`slice_any`): The list to slice.
* `start` (`int`): The starting index (inclusive).
* `end` (`int`): The ending index (exclusive).

**Returns:** (`slice_any`)
---

## `tool.List.Sort`
**Description:** Returns a *new* list with elements sorted. Restricted to lists containing only numbers (int/float) or only strings. Throws error for mixed types or non-sortable types (nil, bool, list, map).

**Parameters:**
* `list` (`slice_any`): The list to sort.

**Returns:** (`slice_any`)
---

## `tool.List.Tail`
**Description:** Returns a *new* list containing the last 'count' elements. Returns an empty list if count <= 0. Returns a copy of the whole list if count >= list length.

**Parameters:**
* `list` (`slice_any`): The list to get the tail from.
* `count` (`int`): The number of elements to take from the end.

**Returns:** (`slice_any`)
---

## `tool.ListAPIFiles`
**Description:** Lists files currently available via the platform's File API.

**Parameters:**
_None_

**Returns:** (`slice_any`)
---

## `tool.Meta.ListTools`
**Description:** Provides a compact text list (sorted alphabetically) of all currently available tools, including basic parameter information. Each tool is listed on a new line, showing its name, parameters (name:type), and return type. Example: FS.Read(filepath:string) -> string

**Parameters:**
_None_

**Returns:** (`string`)
---

## `tool.Meta.ToolsHelp`
**Description:** Provides a more extensive, Markdown-formatted list of available tools, including descriptions, parameters, and return types. Can be filtered by providing a partial tool name. Details include parameter names, types, descriptions, and return type with its description.

**Parameters:**
* `filter` (`string`): (optional) An optional string to filter tool names. Only tools whose names contain this substring will be listed. If empty or omitted, all tools are listed.

**Returns:** (`string`)
---

## `tool.Modulo`
**Description:** Calculates the modulo (remainder) of two integers (num1 % num2). Handles division by zero.

**Parameters:**
* `num1` (`int`): The dividend (must be integer).
* `num2` (`int`): The divisor (must be integer).

**Returns:** (`int`)
---

## `tool.Multiply`
**Description:** Calculates the product of two numbers. Strings convertible to numbers are accepted.

**Parameters:**
* `num1` (`float`): The first number.
* `num2` (`float`): The second number.

**Returns:** (`float`)
---

## `tool.Print`
**Description:** Prints the provided arguments to standard output, separated by spaces, followed by a newline.

**Parameters:**
* `values` (`any`): One or more values to print.

**Returns:** (`nil`)
---

## `tool.Replace`
**Description:** Replaces occurrences of a substring with another, up to a specified count.

**Parameters:**
* `input_string` (`string`): The string to perform replacements on.
* `old_substring` (`string`): The substring to be replaced.
* `new_substring` (`string`): The substring to replace with.
* `count` (`int`): Maximum number of replacements. Use -1 for all.

**Returns:** (`string`)
---

## `tool.Shell.Execute`
**Description:** Executes an arbitrary shell command. WARNING: Use with extreme caution due to security risks. Command path validation is basic. Consider using specific tools (e.g., GoBuild, GitAdd) instead.

**Parameters:**
* `command` (`string`): The command or executable path.
* `args_list` (`slice_string`): (optional) A list of string arguments for the command.
* `directory` (`string`): (optional) Optional directory (relative to sandbox) to execute the command in. Defaults to sandbox root.

**Returns:** (`map`)
---

## `tool.Split`
**Description:** Splits a string by a delimiter.

**Parameters:**
* `input_string` (`string`): The string to split.
* `delimiter` (`string`): The delimiter string.

**Returns:** (`slice_string`)
---

## `tool.SplitWords`
**Description:** Splits a string into words based on whitespace.

**Parameters:**
* `input_string` (`string`): The string to split into words.

**Returns:** (`slice_string`)
---

## `tool.Staticcheck`
**Description:** Runs 'staticcheck' on the specified target(s) in the sandbox. Reports bugs, stylistic errors, and performance issues. Defaults to './...'. Assumes 'staticcheck' is in PATH.

**Parameters:**
* `target` (`string`): (optional) Optional. The target for 'staticcheck' (e.g., a package path or './...'). Defaults to './...'.

**Returns:** (`map`)
---

## `tool.Substring`
**Description:** Returns a portion of the string (rune-based indexing), from start_index for a given length.

**Parameters:**
* `input_string` (`string`): The string to take a substring from.
* `start_index` (`int`): 0-based start index (inclusive).
* `length` (`int`): Number of characters to extract.

**Returns:** (`string`)
---

## `tool.Subtract`
**Description:** Calculates the difference between two numbers (num1 - num2). Strings convertible to numbers are accepted.

**Parameters:**
* `num1` (`float`): The number to subtract from.
* `num2` (`float`): The number to subtract.

**Returns:** (`float`)
---

## `tool.SyncFiles`
**Description:** Synchronizes files between a local sandbox directory and the platform's File API. Supports 'up' (local to API) and 'down' (API to local) directions.

**Parameters:**
* `direction` (`string`): Sync direction: 'up' (local to API) or 'down' (API to local).
* `local_dir` (`string`): Relative path (within the sandbox) of the local directory to sync.
* `filter_pattern` (`string`): (optional) Optional glob pattern (e.g., '*.go', 'data/**') to filter files being synced. Applies to filenames relative to local_dir.
* `ignore_gitignore` (`bool`): (optional) If true, ignores .gitignore rules found within the local_dir (default: false).

**Returns:** (`map`)
---

## `tool.ToLower`
**Description:** Converts a string to lowercase.

**Parameters:**
* `input_string` (`string`): The string to convert.

**Returns:** (`string`)
---

## `tool.ToUpper`
**Description:** Converts a string to uppercase.

**Parameters:**
* `input_string` (`string`): The string to convert.

**Returns:** (`string`)
---

## `tool.Tree.AddChildNode`
**Description:** Adds a new child node to an existing parent node. Returns the ID of the newly created child node.

**Parameters:**
* `tree_handle` (`string`): Handle for the tree structure.
* `parent_node_id` (`string`): ID of the node that will become the parent.
* `new_node_id_suggestion` (`string`): (optional) Optional suggested unique ID for the new node. If empty or nil, an ID will be auto-generated. Must be unique if provided.
* `node_type` (`string`): Type of the new child (e.g., 'object', 'array', 'string', 'number', 'boolean', 'null', 'checklist_item').
* `value` (`any`): (optional) Initial value if the node_type is a leaf or simple type. Ignored for 'object' and 'array' types.
* `key_for_object_parent` (`string`): (optional) If the parent is an 'object' node, this key is used to link the new child in the parent's attributes. Required for object parents.

**Returns:** (`string`)
---

## `tool.Tree.FindNodes`
**Description:** Finds nodes within a tree (starting from a specified node) that match specific criteria. Returns a list of matching node IDs.

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.
* `start_node_id` (`string`): ID of the node within the tree to start searching from. The search includes this node.
* `query_map` (`map`): Map defining search criteria. Supported keys: 'type' (string), 'value' (any), 'metadata' (map of string:string).
* `max_depth` (`int`): (optional) Maximum depth to search relative to the start node (0 for start node only, -1 for unlimited). Default: -1.
* `max_results` (`int`): (optional) Maximum number of matching node IDs to return (-1 for unlimited). Default: -1.

**Returns:** (`slice_string`)
---

## `tool.Tree.GetChildren`
**Description:** Gets a list of node IDs of the children of a given node. For object nodes, children are determined by attribute values that are node IDs. For array nodes, children are from the ordered list. Other node types return an empty list.

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.
* `node_id` (`string`): ID of the parent node.

**Returns:** (`slice_string`)
---

## `tool.Tree.GetNode`
**Description:** Retrieves detailed information about a specific node within a tree, returned as a map. The map includes 'id', 'type', 'value', 'attributes' (map), 'children' (slice of IDs), and 'parentId'.

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.
* `node_id` (`string`): The unique ID of the node to retrieve.

**Returns:** (`map`)
---

## `tool.Tree.GetParent`
**Description:** Gets the node ID of the parent of a given node. Returns an empty string for the root node or if the node has no parent.

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.
* `node_id` (`string`): ID of the node whose parent is sought.

**Returns:** (`string`)
---

## `tool.Tree.LoadJSON`
**Description:** Loads a JSON string into a new tree structure and returns a tree handle.

**Parameters:**
* `json_string` (`string`): The JSON data as a string.

**Returns:** (`string`)
---

## `tool.Tree.RemoveNode`
**Description:** Removes a node (specified by ID) and all its descendants from the tree. Cannot remove the root node.

**Parameters:**
* `tree_handle` (`string`): Handle to the tree.
* `node_id` (`string`): ID of the node to remove.

**Returns:** (`nil`)
---

## `tool.Tree.RemoveNodeMetadata`
**Description:** Removes a metadata attribute (a key-value string pair) from a node.

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.
* `node_id` (`string`): ID of the node to remove metadata from.
* `metadata_key` (`string`): The key of the metadata attribute to remove.

**Returns:** (`nil`)
---

## `tool.Tree.RemoveObjectAttribute`
**Description:** Removes an attribute (a key mapping to a child node ID) from an 'object' type node. This unlinks the child but does not delete the child node itself.

**Parameters:**
* `tree_handle` (`string`): Handle for the tree structure.
* `object_node_id` (`string`): Unique ID of the 'object' type node to modify.
* `attribute_key` (`string`): The key (name) of the attribute to remove.

**Returns:** (`nil`)
---

## `tool.Tree.RenderText`
**Description:** Renders a visual text representation of the entire tree structure identified by the given tree handle.

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure to render.

**Returns:** (`string`)
---

## `tool.Tree.SetNodeMetadata`
**Description:** Sets a metadata attribute as a key-value string pair on any node. This is separate from object attributes that link to child nodes.

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.
* `node_id` (`string`): ID of the node to set metadata on.
* `metadata_key` (`string`): The key of the metadata attribute (string).
* `metadata_value` (`string`): The value of the metadata attribute (string).

**Returns:** (`nil`)
---

## `tool.Tree.SetObjectAttribute`
**Description:** Sets or updates an attribute on an 'object' type node, mapping the attribute key to an existing child node's ID. This is for establishing parent-child relationships in object nodes.

**Parameters:**
* `tree_handle` (`string`): Handle for the tree structure.
* `object_node_id` (`string`): Unique ID of the 'object' type node to modify.
* `attribute_key` (`string`): The key (name) of the attribute to set.
* `child_node_id` (`string`): The ID of an *existing* node within the same tree to associate with the key.

**Returns:** (`nil`)
---

## `tool.Tree.SetValue`
**Description:** Sets the value of an existing leaf or simple-type node (e.g., string, number, boolean, null, checklist_item). Cannot set value on 'object' or 'array' type nodes using this tool.

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.
* `node_id` (`string`): ID of the leaf or simple-type node to modify.
* `value` (`any`): The new value for the node.

**Returns:** (`nil`)
---

## `tool.Tree.ToJSON`
**Description:** Converts a tree structure (identified by tree handle) back into a JSON string. Output is pretty-printed.

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.

**Returns:** (`string`)
---

## `tool.TrimSpace`
**Description:** Removes leading and trailing whitespace from a string.

**Parameters:**
* `input_string` (`string`): The string to trim.

**Returns:** (`string`)
---

## `tool.UploadFile`
**Description:** Uploads a local file (from the sandbox) to the platform's File API. Returns a map describing the uploaded file.

**Parameters:**
* `local_filepath` (`string`): Relative path (within the sandbox) of the local file to upload.
* `api_display_name` (`string`): (optional) Optional display name for the file on the API.

**Returns:** (`map`)
---


[green]Initial script completed.[-]
