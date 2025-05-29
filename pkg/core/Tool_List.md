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
FS.Mkdir(path:string) -> map
FS.Move(source_path:string, destination_path:string) -> map
FS.Read(filepath:string) -> string
FS.SanitizeFilename(name:string) -> string
FS.Stat(path:string) -> map
FS.Walk(path:string) -> slice_any
FS.Write(filepath:string, content:string) -> string
Git.Branch(relative_path:string, name:string?, checkout:bool?, list_remote:bool?, list_all:bool?) -> string
Git.Checkout(relative_path:string, branch:string, create:bool?) -> string
Git.Clone(repository_url:string, relative_path:string) -> string
Git.Commit(relative_path:string, commit_message:string, allow_empty:bool?) -> string
Git.Diff(relative_path:string, cached:bool?, commit1:string?, commit2:string?, path:string?) -> string
Git.Merge(relative_path:string, branch:string) -> string
Git.Pull(relative_path:string, remote_name:string?, branch_name:string?) -> string
Git.Push(relative_path:string, remote_name:string?, branch_name:string?) -> string
Git.Rm(relative_path:string, paths:any) -> string
Git.Status(repo_path:string?) -> map
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
Input(message:string?) -> string
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
Meta.GetToolSpecificationsJSON() -> string
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

**Category:** AI Worker Management

**Parameters:**
* `definition_id` (`string`): ID of the AIWorkerDefinition to use.
* `prompt` (`string`): The prompt/input text for the LLM.
* `config_overrides` (`map`): (optional) Optional map of configuration overrides for this specific execution.

**Returns:** (`map`) Returns a map: {'output': string (LLM response), 'taskId': string, 'cost': float64}. Returns nil on error.

**Example:**
```neuroscript
TOOL.AIWorker.ExecuteStatelessTask(definition_id: "google-gemini-1.5-flash", prompt: "Translate 'hello' to French.")
```

**Error Conditions:** ErrAIWorkerManagerMissing; ErrInvalidArgument for missing/invalid args; ErrConfiguration if interpreter's LLMClient is nil; Errors from AIWorkerManager.ExecuteStatelessTask (e.g., ErrDefinitionNotFound, LLM communication errors, rate limits); ErrInternal if performance record is nil without error.
---

## `tool.AIWorker.GetPerformanceRecords`
**Description:** Retrieves persisted performance records for a specific AI Worker Definition.

**Category:** AI Worker Management

**Parameters:**
* `definition_id` (`string`): ID of the AIWorkerDefinition for which to retrieve records.
* `filters` (`map`): (optional) Optional map of filters to apply to the records (e.g., {'success':true}).

**Returns:** (`slice_map`) Returns a slice of maps, where each map represents a PerformanceRecord. Returns an empty slice if no records match or exist.

**Example:**
```neuroscript
TOOL.AIWorker.GetPerformanceRecords(definition_id: "google-gemini-1.5-pro", filters: {"success":true})
```

**Error Conditions:** ErrAIWorkerManagerMissing; ErrInvalidArgument for missing/invalid args; Errors from AIWorkerManager.GetPerformanceRecordsForDefinition (e.g., file I/O for persistence, JSON parsing errors).
---

## `tool.AIWorker.LoadPerformanceData`
**Description:** Reloads all worker definitions, which implicitly re-processes performance summaries from persisted data.

**Category:** AI Worker Management

**Parameters:**
_None_

**Returns:** (`string`) Returns a string message: 'Worker definitions and associated performance summaries reloaded.'.

**Example:**
```neuroscript
TOOL.AIWorker.LoadPerformanceData()
```

**Error Conditions:** ErrAIWorkerManagerMissing; Errors from AIWorkerManager.LoadWorkerDefinitionsFromFile (e.g., file I/O, JSON parsing).
---

## `tool.AIWorker.LogPerformance`
**Description:** Logs a performance record for an AI Worker task.

**Category:** AI Worker Management

**Parameters:**
* `task_id` (`string`): Unique ID for the task.
* `instance_id` (`string`): ID of the AIWorkerInstance used.
* `definition_id` (`string`): ID of the AIWorkerDefinition used.
* `timestamp_start` (`string`): Start timestamp (RFC3339Nano or RFC3339 format).
* `timestamp_end` (`string`): End timestamp (RFC3339Nano or RFC3339 format).
* `duration_ms` (`int`): Task duration in milliseconds.
* `success` (`bool`): Whether the task was successful.
* `input_context` (`map`): (optional) Optional map of input context details.
* `llm_metrics` (`map`): (optional) Optional map of LLM-specific metrics (e.g., token counts, finish reason).
* `cost_incurred` (`float`): (optional) Optional cost incurred for this task.
* `output_summary` (`string`): (optional) Optional summary of the task output.
* `error_details` (`string`): (optional) Optional error details if success is false.

**Returns:** (`string`) Returns the TaskID string of the logged performance record.

**Example:**
```neuroscript
TOOL.AIWorker.LogPerformance(task_id: "task_abc", instance_id: "inst_123", definition_id: "def_xyz", timestamp_start: "2023-10-27T10:00:00.000Z", timestamp_end: "2023-10-27T10:00:05.123Z", duration_ms: 5123, success: true, llm_metrics: {"input_tokens":10, "output_tokens":50})
```

**Error Conditions:** ErrAIWorkerManagerMissing; ErrInvalidArgument if required arguments are missing/invalid type (e.g., timestamp format, duration_ms not int); Errors from AIWorkerManager.logPerformanceRecordUnsafe or persistDefinitionsUnsafe (e.g., file I/O for persistence).
---

## `tool.AIWorker.SavePerformanceData`
**Description:** Explicitly triggers saving of all worker definitions (which include performance summaries). Raw performance data for instances is appended when an instance is retired.

**Category:** AI Worker Management

**Parameters:**
_None_

**Returns:** (`string`) Returns a string message: 'Ensured definitions (with summaries) are saved. Raw performance data appends automatically.'.

**Example:**
```neuroscript
TOOL.AIWorker.SavePerformanceData()
```

**Error Conditions:** ErrAIWorkerManagerMissing; Errors from AIWorkerManager.persistDefinitionsUnsafe (e.g., file I/O or JSON serialization errors).
---

## `tool.AIWorkerDefinition.Add`
**Description:** Adds a new AI Worker Definition. Maps (base_config, etc.) are optional.

**Category:** AI Worker Management

**Parameters:**
* `definition_id` (`string`): (optional) Optional unique ID for the definition. Auto-generated if not provided.
* `name` (`string`): (optional) Optional human-readable name for the definition.
* `provider` (`string`): The AI provider (e.g., 'google', 'openai').
* `model_name` (`string`): The specific model name from the provider (e.g., 'gemini-1.5-pro-latest').
* `auth` (`map`): Authentication details. Map e.g., {'method':'env', 'value':'GOOGLE_API_KEY'}.
* `interaction_models` (`slice_string`): (optional) List of supported interaction models (e.g., 'chat', 'embedding').
* `capabilities` (`slice_string`): (optional) List of capabilities (e.g., 'tools', 'json_mode').
* `base_config` (`map`): (optional) Base configuration for the model (e.g., temperature, top_p).
* `cost_metrics` (`map`): (optional) Cost metrics (e.g., {'input_token_cost':0.0001, 'output_token_cost':0.0003}).
* `rate_limits` (`map`): (optional) Rate limit policy (e.g., {'max_requests_per_minute':60}).
* `status` (`string`): (optional) Initial status (e.g., 'active', 'disabled'). Defaults to 'active'.
* `default_file_contexts` (`slice_string`): (optional) List of default file context URIs.
* `metadata` (`map`): (optional) Arbitrary key-value metadata.

**Returns:** (`string`) Returns the unique DefinitionID string of the newly added AI Worker Definition.

**Example:**
```neuroscript
TOOL.AIWorkerDefinition.Add(provider: "google", model_name: "gemini-1.5-flash", auth: {"method":"env", "value":"MY_API_KEY"}, interaction_models: ["chat"])
```

**Error Conditions:** ErrAIWorkerManagerMissing if AI Worker Manager is not found in interpreter; ErrInvalidArgument if argument validation fails (e.g. missing required fields like provider, model_name, auth, or incorrect types); Errors from AIWorkerManager.AddWorkerDefinition (e.g., ErrDuplicateDefinitionID, ErrInvalidDefinition).
---

## `tool.AIWorkerDefinition.Get`
**Description:** Retrieves an AI Worker Definition by its ID.

**Category:** AI Worker Management

**Parameters:**
* `definition_id` (`string`): The unique ID of the definition to retrieve.

**Returns:** (`map`) Returns a map representing the AIWorkerDefinition struct. Returns nil if not found or on error.

**Example:**
```neuroscript
TOOL.AIWorkerDefinition.Get(definition_id: "google-gemini-1.5-pro")
```

**Error Conditions:** ErrAIWorkerManagerMissing; ErrInvalidArgument if definition_id is not provided or not a string; ErrDefinitionNotFound if definition with ID does not exist.
---

## `tool.AIWorkerDefinition.List`
**Description:** Lists all AI Worker Definitions, optionally filtered.

**Category:** AI Worker Management

**Parameters:**
* `filters` (`map`): (optional) Optional map of filters (e.g., {'provider':'google', 'status':'active'}).

**Returns:** (`slice_map`) Returns a slice of maps, where each map represents an AIWorkerDefinition. Returns an empty slice if no definitions match or exist.

**Example:**
```neuroscript
TOOL.AIWorkerDefinition.List(filters: {"provider":"google"})
```

**Error Conditions:** ErrAIWorkerManagerMissing; ErrInvalidArgument if filters is not a map.
---

## `tool.AIWorkerDefinition.LoadAll`
**Description:** Reloads all worker definitions from the configured JSON file.

**Category:** AI Worker Management

**Parameters:**
_None_

**Returns:** (`string`) Returns a string message indicating the number of definitions reloaded, e.g., 'Reloaded X worker definitions.'.

**Example:**
```neuroscript
TOOL.AIWorkerDefinition.LoadAll()
```

**Error Conditions:** ErrAIWorkerManagerMissing; Errors from AIWorkerManager.LoadWorkerDefinitionsFromFile (e.g., related to file I/O, JSON parsing, or validation of loaded definitions).
---

## `tool.AIWorkerDefinition.Remove`
**Description:** Removes an AI Worker Definition if it has no active instances.

**Category:** AI Worker Management

**Parameters:**
* `definition_id` (`string`): The unique ID of the definition to remove.

**Returns:** (`nil`) Returns nil on successful removal.

**Example:**
```neuroscript
TOOL.AIWorkerDefinition.Remove(definition_id: "old-unused-definition")
```

**Error Conditions:** ErrAIWorkerManagerMissing; ErrInvalidArgument if definition_id is missing or not a string; Errors from AIWorkerManager.RemoveWorkerDefinition (e.g., ErrDefinitionNotFound, ErrDefinitionInUse).
---

## `tool.AIWorkerDefinition.SaveAll`
**Description:** Saves all current worker definitions to the configured JSON file.

**Category:** AI Worker Management

**Parameters:**
_None_

**Returns:** (`string`) Returns a string message indicating the number of definitions saved, e.g., 'Saved X worker definitions.'.

**Example:**
```neuroscript
TOOL.AIWorkerDefinition.SaveAll()
```

**Error Conditions:** ErrAIWorkerManagerMissing; Errors from AIWorkerManager.SaveWorkerDefinitionsToFile (e.g., related to file I/O or JSON serialization).
---

## `tool.AIWorkerDefinition.Update`
**Description:** Updates fields of an existing AI Worker Definition.

**Category:** AI Worker Management

**Parameters:**
* `definition_id` (`string`): The unique ID of the definition to update.
* `updates` (`map`): A map of fields to update (e.g., {'status':'disabled', 'metadata':{'key':'new_value'}}).

**Returns:** (`nil`) Returns nil on successful update.

**Example:**
```neuroscript
TOOL.AIWorkerDefinition.Update(definition_id: "google-gemini-1.5-pro", updates: {"status":"disabled", "cost_metrics":{"input_token_cost":0.00015}})
```

**Error Conditions:** ErrAIWorkerManagerMissing; ErrInvalidArgument if definition_id or updates are missing/invalid type; Errors from AIWorkerManager.UpdateWorkerDefinition (e.g., ErrDefinitionNotFound, ErrInvalidDefinitionField).
---

## `tool.AIWorkerInstance.Get`
**Description:** Retrieves an active AI Worker Instance's details by its ID.

**Category:** AI Worker Management

**Parameters:**
* `instance_id` (`string`): The unique ID of the active instance to retrieve.

**Returns:** (`map`) Returns a map representing the AIWorkerInstance. Returns nil if not found or on error.

**Example:**
```neuroscript
TOOL.AIWorkerInstance.Get(instance_id: "instance_uuid_123")
```

**Error Conditions:** ErrAIWorkerManagerMissing; ErrInvalidArgument if instance_id is missing or not a string; ErrInstanceNotFound if instance with ID does not exist or is not active.
---

## `tool.AIWorkerInstance.ListActive`
**Description:** Lists currently active AI Worker Instances, optionally filtered.

**Category:** AI Worker Management

**Parameters:**
* `filters` (`map`): (optional) Optional map of filters (e.g., {'definition_id':'google-gemini-1.5-pro'}).

**Returns:** (`slice_map`) Returns a slice of maps, where each map represents an active AIWorkerInstance. Returns empty slice if no active instances match.

**Example:**
```neuroscript
TOOL.AIWorkerInstance.ListActive(filters: {"definition_id":"google-gemini-1.5-pro"})
```

**Error Conditions:** ErrAIWorkerManagerMissing; ErrInvalidArgument if filters is not a map.
---

## `tool.AIWorkerInstance.Retire`
**Description:** Retires an active AI Worker Instance, persisting its final state and performance.

**Category:** AI Worker Management

**Parameters:**
* `instance_id` (`string`): ID of the instance to retire.
* `conversation_manager_handle` (`string`): Handle of the associated ConversationManager to be removed.
* `reason` (`string`): Reason for retiring the instance.
* `final_status` (`string`): Final status (e.g., 'completed', 'error', 'cancelled').
* `final_session_token_usage` (`map`): Map of final token usage for the session (e.g., {'input_tokens':100, 'output_tokens':200}).
* `performance_records` (`slice_map`): (optional) Optional slice of performance record maps to log before retiring.

**Returns:** (`nil`) Returns nil on successful retirement.

**Example:**
```neuroscript
TOOL.AIWorkerInstance.Retire(instance_id: "instance_uuid_123", conversation_manager_handle: "conv_handle_abc", reason: "Task completed", final_status: "completed", final_session_token_usage: {"input_tokens":500, "output_tokens":1500})
```

**Error Conditions:** ErrAIWorkerManagerMissing; ErrInvalidArgument if required arguments are missing or of incorrect type (e.g., final_session_token_usage not a map); Errors from AIWorkerManager.RetireWorkerInstance (e.g., ErrInstanceNotFound). Failure to remove handle is logged as a warning.
---

## `tool.AIWorkerInstance.Spawn`
**Description:** Spawns a new AI Worker Instance and returns its details including a ConversationManager handle.

**Category:** AI Worker Management

**Parameters:**
* `definition_id` (`string`): ID of the AIWorkerDefinition to use for spawning.
* `config_overrides` (`map`): (optional) Optional map of configuration overrides for this instance.
* `file_contexts` (`slice_string`): (optional) Optional list of file context URIs for this instance.

**Returns:** (`map`) Returns a map representing the spawned AIWorkerInstance, including a 'conversation_manager_handle' string. Returns nil on error.

**Example:**
```neuroscript
TOOL.AIWorkerInstance.Spawn(definition_id: "google-gemini-1.5-pro", config_overrides: {"temperature":0.8})
```

**Error Conditions:** ErrAIWorkerManagerMissing; ErrInvalidArgument if validation fails for definition_id, config_overrides, or file_contexts; Errors from AIWorkerManager.SpawnWorkerInstance (e.g., ErrDefinitionNotFound, rate limit errors); ErrInternal if SpawnWorkerInstance returns nil instance without error; Errors related to interpreter.RegisterHandle if ConversationManager handle registration fails.
---

## `tool.AIWorkerInstance.UpdateStatus`
**Description:** Updates the status and optionally the last error of an active AI Worker Instance.

**Category:** AI Worker Management

**Parameters:**
* `instance_id` (`string`): ID of the active instance to update.
* `status` (`string`): New status for the instance (e.g., 'processing', 'idle', 'error').
* `last_error` (`string`): (optional) Optional error message if status is 'error'.

**Returns:** (`nil`) Returns nil on successful status update.

**Example:**
```neuroscript
TOOL.AIWorkerInstance.UpdateStatus(instance_id: "instance_uuid_123", status: "processing")
```

**Error Conditions:** ErrAIWorkerManagerMissing; ErrInvalidArgument if required arguments are missing/invalid type; Errors from AIWorkerManager.UpdateInstanceStatus (e.g., ErrInstanceNotFound).
---

## `tool.AIWorkerInstance.UpdateTokenUsage`
**Description:** Updates the session token usage for an active AI Worker Instance.

**Category:** AI Worker Management

**Parameters:**
* `instance_id` (`string`): ID of the active instance.
* `input_tokens` (`int`): Number of input tokens to add to the session total.
* `output_tokens` (`int`): Number of output tokens to add to the session total.

**Returns:** (`nil`) Returns nil on successful token usage update.

**Example:**
```neuroscript
TOOL.AIWorkerInstance.UpdateTokenUsage(instance_id: "instance_uuid_123", input_tokens: 120, output_tokens: 350)
```

**Error Conditions:** ErrAIWorkerManagerMissing; ErrInvalidArgument if required arguments are missing/invalid type; Errors from AIWorkerManager.UpdateInstanceSessionTokenUsage (e.g., ErrInstanceNotFound).
---

## `tool.Add`
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

**Error Conditions:** Returns an 'ErrInternalTool' if arguments cannot be processed as float64 (this scenario should ideally be caught by input validation before the tool function is called).
---

## `tool.Concat`
**Description:** Concatenates a list of strings without a separator.

**Category:** String Operations

**Parameters:**
* `strings_list` (`slice_string`): List of strings to concatenate.

**Returns:** (`string`) Returns a single string by concatenating all strings in the strings_list.

**Example:**
```neuroscript
tool.Concat(["hello", " ", "world"]) // Returns "hello world"
```

**Error Conditions:** Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `strings_list` is not a list of strings. May return `ErrTypeAssertionFailed` (with `ErrorCodeInternal`) if type validation fails unexpectedly.
---

## `tool.Contains`
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

**Error Conditions:** Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `input_string` or `substring` are not strings.
---

## `tool.DeleteAPIFile`
**Description:** Deletes a specific file from the platform's File API using its ID/URI.

**Parameters:**
* `api_file_id` (`string`): The unique ID or URI of the file on the API (e.g., 'files/abcde123').

**Returns:** (`string`) 
---

## `tool.Divide`
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

**Error Conditions:** Returns 'ErrDivisionByZero' if num2 is 0. Returns an 'ErrInternalTool' if arguments cannot be processed as float64 (should be caught by validation).
---

## `tool.FS.Delete`
**Description:** Deletes a file or an empty directory. Returns 'OK' on success or if path doesn't exist.

**Category:** Filesystem

**Parameters:**
* `path` (`string`): Relative path to the file or empty directory to delete.

**Returns:** (`string`) Returns the string 'OK' on successful deletion or if the path does not exist. Returns nil on error.

**Example:**
```neuroscript
TOOL.FS.Delete(path: "temp/old_file.txt") // Returns "OK"
```

**Error Conditions:** ErrArgumentMismatch if path is empty or not a string; ErrConfiguration if sandbox is not set; ErrSecurityPath (from SecureFilePath) for invalid path; ErrPreconditionFailed if directory is not empty; ErrPermissionDenied; ErrIOFailed for other I/O errors. Path not found is treated as success.
---

## `tool.FS.Hash`
**Description:** Calculates the SHA256 hash of a specified file. Returns the hex-encoded hash string.

**Category:** Filesystem

**Parameters:**
* `filepath` (`string`): Relative path (within the sandbox) of the file to hash.

**Returns:** (`string`) Returns a hex-encoded SHA256 hash string of the file's content. Returns an empty string on error.

**Example:**
```neuroscript
TOOL.FS.Hash(filepath: "data/my_document.txt") // Returns "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" (example for an empty file)
```

**Error Conditions:** ErrArgumentMismatch if filepath is empty; ErrConfiguration if sandbox is not set; ErrSecurityPath (from SecureFilePath) for invalid paths; ErrFileNotFound if file does not exist; ErrPermissionDenied if file cannot be opened; ErrPathNotFile if path is a directory; ErrIOFailed for other I/O errors during open or read.
---

## `tool.FS.LineCount`
**Description:** Counts lines in a specified file. Returns line count as an integer.

**Category:** Filesystem

**Parameters:**
* `filepath` (`string`): Relative path to the file.

**Returns:** (`int`) Returns the number of lines in the specified file. Returns 0 on error or if file is empty.

**Example:**
```neuroscript
TOOL.FS.LineCount(filepath: "logs/app.log") // Returns 150 (example)
```

**Error Conditions:** ErrArgumentMismatch if filepath is empty; ErrConfiguration if sandbox is not set; ErrSecurityPath for invalid paths; ErrFileNotFound; ErrPermissionDenied; ErrPathNotFile if path is a directory; ErrIOFailed for read errors. (Based on typical file tool error handling, actual implementation for toolLineCountFile in tools_fs_utils.go needed for exact errors).
---

## `tool.FS.List`
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

**Error Conditions:** ErrArgumentMismatch if path is not a string or recursive is not bool/nil; ErrConfiguration if sandbox is not set; ErrSecurityPath (from ResolveAndSecurePath) for invalid path; ErrFileNotFound if path does not exist; ErrPermissionDenied; ErrPathNotDirectory if path is not a directory; ErrIOFailed for other I/O errors during listing or walking.
---

## `tool.FS.Mkdir`
**Description:** Creates a directory. Parent directories are created if they do not exist (like mkdir -p). Returns a success message.

**Category:** Filesystem

**Parameters:**
* `path` (`string`): Relative path of the directory to create.

**Returns:** (`map`) Returns a map: {'status':'success', 'message':'Successfully created directory: <path>', 'path':'<path>'} on success. Returns nil on error.

**Example:**
```neuroscript
TOOL.FS.Mkdir(path: "new/subdir") // Returns {"status":"success", "message":"Successfully created directory: new/subdir", "path":"new/subdir"}
```

**Error Conditions:** ErrArgumentMismatch if path is empty, '.', or not a string; ErrConfiguration if sandbox is not set; ErrSecurityPath (from ResolveAndSecurePath) for invalid path; ErrPathNotDirectory if path exists and is a file; ErrPathExists if directory already exists; ErrPermissionDenied; ErrIOFailed for other I/O errors or failure to stat; ErrCannotCreateDir if MkdirAll fails.
---

## `tool.FS.Move`
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

**Error Conditions:** ErrArgumentMismatch if paths are empty, not strings, or are the same; ErrConfiguration if sandbox is not set; ErrSecurityPath (from SecureFilePath) for invalid source or destination paths; ErrFileNotFound if source path does not exist; ErrPathExists if destination path already exists; ErrPermissionDenied for source or destination; ErrIOFailed for other I/O errors during stat or rename.
---

## `tool.FS.Read`
**Description:** Reads the entire content of a specific file. Returns the content as a string.

**Category:** Filesystem

**Parameters:**
* `filepath` (`string`): Relative path to the file.

**Returns:** (`string`) Returns the content of the file as a string. Returns an empty string on error.

**Example:**
```neuroscript
TOOL.FS.Read(filepath: "config.txt") // Returns "setting=value\n..."
```

**Error Conditions:** ErrArgumentMismatch if filepath is empty; ErrConfiguration if sandbox is not set; ErrSecurityPath (from ResolveAndSecurePath) for invalid paths; ErrFileNotFound if file does not exist; ErrPermissionDenied; ErrPathNotFile if path is a directory; ErrIOFailed for other I/O errors.
---

## `tool.FS.SanitizeFilename`
**Description:** Cleans a string to make it suitable for use as part of a filename.

**Category:** Filesystem Utilities

**Parameters:**
* `name` (`string`): The string to sanitize.

**Returns:** (`string`) Returns a sanitized string suitable for use as a filename component (e.g., replacing unsafe characters with underscores).

**Example:**
```neuroscript
TOOL.FS.SanitizeFilename(name: "My Report Final?.docx") // Returns "My_Report_Final_.docx" (example)
```

**Error Conditions:** ErrArgumentMismatch if name is not provided or not a string. (Based on typical utility tool error handling, actual implementation for toolSanitizeFilename in tools_fs_utils.go needed for exact errors).
---

## `tool.FS.Stat`
**Description:** Gets information about a file or directory. Returns a map containing: name(string), path(string), size_bytes(int), is_dir(bool), modified_unix(int), modified_rfc3339(string - format 2006-01-02T15:04:05.999999999Z07:00), mode_string(string), mode_perm(string).

**Category:** Filesystem

**Parameters:**
* `path` (`string`): Relative path to the file or directory.

**Returns:** (`map`) Returns a map with file/directory info: {'name', 'path', 'size_bytes', 'is_dir', 'modified_unix', 'modified_rfc3339', 'mode_string', 'mode_perm'}. Returns nil on error.

**Example:**
```neuroscript
TOOL.FS.Stat(path: "my_file.go")
```

**Error Conditions:** ErrArgumentMismatch if path is empty or not a string; ErrConfiguration if sandbox is not set; ErrSecurityPath (from ResolveAndSecurePath) for invalid path; ErrFileNotFound if path does not exist; ErrPermissionDenied; ErrIOFailed for other I/O errors.
---

## `tool.FS.Walk`
**Description:** Recursively walks a directory, returning a list of maps describing files/subdirectories found (keys: name, path_relative, is_dir, size_bytes, modified_unix, modified_rfc3339 (format 2006-01-02T15:04:05.999999999Z07:00), mode_string). Skips the root directory itself.

**Category:** Filesystem

**Parameters:**
* `path` (`string`): Relative path to the directory to walk.

**Returns:** (`slice_any`) Returns a slice of maps, each describing a file/subdir: {'name', 'path_relative', 'is_dir', 'size_bytes', 'modified_unix', 'modified_rfc3339', 'mode_string'}. Skips the root dir itself. Returns nil on error.

**Example:**
```neuroscript
TOOL.FS.Walk(path: "src")
```

**Error Conditions:** ErrArgumentMismatch if path is empty or not a string; ErrConfiguration if sandbox is not set; ErrSecurityPath (from ResolveAndSecurePath) for invalid path; ErrFileNotFound if start path not found; ErrPathNotDirectory if start path is not a directory; ErrPermissionDenied for start path; ErrIOFailed for stat errors or errors during walk; ErrInternal if relative path calculation fails during walk.
---

## `tool.FS.Write`
**Description:** Writes content to a specific file. Creates parent directories if needed. Returns 'OK' on success.

**Category:** Filesystem

**Parameters:**
* `filepath` (`string`): Relative path to the file.
* `content` (`string`): The content to write.

**Returns:** (`string`) Returns a success message string like 'Successfully wrote X bytes to Y' on success. Returns an empty string on error.

**Example:**
```neuroscript
TOOL.FS.Write(filepath: "output/data.json", content: "{\"key\":\"value\"}") // Returns "Successfully wrote 15 bytes to output/data.json"
```

**Error Conditions:** ErrArgumentMismatch if filepath is empty or content is not string/nil; ErrConfiguration if sandbox is not set; ErrSecurityPath (from ResolveAndSecurePath) for invalid paths; ErrCannotCreateDir if parent directories cannot be created; ErrPermissionDenied if writing is not allowed; ErrPathNotFile if path exists and is a directory; ErrIOFailed for other I/O errors.
---

## `tool.Git.Branch`
**Description:** Manages branches: lists, creates, or creates and checks out branches in a Git repository.

**Category:** Git

**Parameters:**
* `relative_path` (`string`): Relative path to the repository within the sandbox.
* `name` (`string`): (optional) Optional. The name of the branch to create. If omitted, and other list flags are false, lists local branches.
* `checkout` (`bool`): (optional) Optional. If true and 'name' is provided, checks out the new branch after creation. Defaults to false.
* `list_remote` (`bool`): (optional) Optional. If true, lists remote-tracking branches. Defaults to false.
* `list_all` (`bool`): (optional) Optional. If true, lists all local and remote-tracking branches. Defaults to false.

**Returns:** (`string`) Returns a success message (e.g., on creation) or a string listing branches. Behavior depends on arguments.

**Example:**
```neuroscript
TOOL.Git.Branch(relative_path: "my_repo", name: "new-feature", checkout: true)\nTOOL.Git.Branch(relative_path: "my_repo", list_all: true)
```

**Error Conditions:** ErrConfiguration if sandbox not set; ErrInvalidArgument for bad args; ErrGitRepositoryNotFound; ErrGitOperationFailed for git command errors; ErrSecurityPath for invalid relative_path.
---

## `tool.Git.Checkout`
**Description:** Switches branches or restores working tree files in a Git repository.

**Category:** Git

**Parameters:**
* `relative_path` (`string`): Relative path to the repository within the sandbox.
* `branch` (`string`): The name of the branch to checkout or the commit/pathspec to restore.
* `create` (`bool`): (optional) Optional. If true, creates a new branch named by 'branch' and checks it out. Defaults to false.

**Returns:** (`string`) Returns a success message on successful checkout.

**Example:**
```neuroscript
TOOL.Git.Checkout(relative_path: "my_repo", branch: "main")\nTOOL.Git.Checkout(relative_path: "my_repo", branch: "new-feature", create: true)
```

**Error Conditions:** ErrConfiguration if sandbox not set; ErrInvalidArgument for bad args; ErrGitRepositoryNotFound; ErrGitOperationFailed for git command errors (e.g. branch not found, uncommitted changes); ErrSecurityPath for invalid relative_path.
---

## `tool.Git.Clone`
**Description:** Clones a Git repository into the specified relative path within the sandbox.

**Category:** Git

**Parameters:**
* `repository_url` (`string`): The URL of the Git repository to clone.
* `relative_path` (`string`): The relative path within the sandbox where the repository should be cloned.

**Returns:** (`string`) Returns a success message string like 'Successfully cloned <URL> to <path>.' on successful clone. Returns nil on error.

**Example:**
```neuroscript
TOOL.Git.Clone(repository_url: "https://github.com/example/repo.git", relative_path: "cloned_repos/my_repo")
```

**Error Conditions:** ErrConfiguration if sandbox directory is not set; ErrInvalidArgument if repository_url or relative_path are missing or not strings; ErrPathExists if the target relative_path already exists; ErrGitOperationFailed for errors during the 'git clone' command execution (e.g., authentication failure, repository not found, network issues); ErrSecurityPath for invalid relative_path.
---

## `tool.Git.Commit`
**Description:** Commits staged changes in the specified Git repository within the sandbox.

**Category:** Git

**Parameters:**
* `relative_path` (`string`): The relative path within the sandbox to the Git repository.
* `commit_message` (`string`): The commit message.
* `allow_empty` (`bool`): (optional) Optional. Allow an empty commit (no changes). Defaults to false.

**Returns:** (`string`) Returns a success message string like 'Successfully committed to repository <path>.' or the commit hash. Returns nil on error.

**Example:**
```neuroscript
TOOL.Git.Commit(relative_path: "my_repo", commit_message: "Fix: addressed critical bug #123")
```

**Error Conditions:** ErrConfiguration if sandbox directory is not set; ErrInvalidArgument if relative_path or commit_message are missing/invalid types; ErrGitRepositoryNotFound if the specified relative_path is not a Git repository; ErrGitOperationFailed for errors during the 'git commit' command (e.g., nothing to commit and allow_empty is false, pre-commit hooks failure); ErrSecurityPath for invalid relative_path.
---

## `tool.Git.Diff`
**Description:** Shows changes between commits, commit and working tree, etc., in a Git repository.

**Category:** Git

**Parameters:**
* `relative_path` (`string`): Relative path to the repository within the sandbox.
* `cached` (`bool`): (optional) Optional. Show staged changes (diff against HEAD). Defaults to false.
* `commit1` (`string`): (optional) Optional. First commit or tree object for diff. Defaults to index if 'cached' is true, or HEAD otherwise.
* `commit2` (`string`): (optional) Optional. Second commit or tree object for diff. Defaults to the working tree.
* `path` (`string`): (optional) Optional. Limit the diff to the specified file or directory path within the repository.

**Returns:** (`string`) Returns a string containing the diff output.

**Example:**
```neuroscript
TOOL.Git.Diff(relative_path: "my_repo", cached: true)\nTOOL.Git.Diff(relative_path: "my_repo", commit1: "HEAD~1", commit2: "HEAD", path: "src/")
```

**Error Conditions:** ErrConfiguration if sandbox not set; ErrInvalidArgument for bad args; ErrGitRepositoryNotFound; ErrGitOperationFailed for git command errors; ErrSecurityPath for invalid relative_path.
---

## `tool.Git.Merge`
**Description:** Joins two or more development histories together in a Git repository.

**Category:** Git

**Parameters:**
* `relative_path` (`string`): Relative path to the repository within the sandbox.
* `branch` (`string`): The name of the branch to merge into the current branch.

**Returns:** (`string`) Returns a success message or merge details on successful merge.

**Example:**
```neuroscript
TOOL.Git.Merge(relative_path: "my_repo", branch: "feature-branch")
```

**Error Conditions:** ErrConfiguration if sandbox not set; ErrInvalidArgument for bad args; ErrGitRepositoryNotFound; ErrGitOperationFailed for git command errors (e.g. merge conflicts); ErrSecurityPath for invalid relative_path.
---

## `tool.Git.Pull`
**Description:** Pulls the latest changes from the remote repository for the specified Git repository within the sandbox.

**Category:** Git

**Parameters:**
* `relative_path` (`string`): The relative path within the sandbox to the Git repository.
* `remote_name` (`string`): (optional) Optional. The name of the remote to pull from (e.g., 'origin'). Defaults to 'origin'.
* `branch_name` (`string`): (optional) Optional. The name of the branch to pull. Defaults to the current branch.

**Returns:** (`string`) Returns a success message string like 'Successfully pulled from <remote>/<branch> for repository <path>.' or details of the pull. Returns nil on error.

**Example:**
```neuroscript
TOOL.Git.Pull(relative_path: "my_repo")\nTOOL.Git.Pull(relative_path: "my_repo", remote_name: "upstream", branch_name: "main")
```

**Error Conditions:** ErrConfiguration if sandbox directory is not set; ErrInvalidArgument if relative_path is missing or not a string, or other args are invalid types; ErrGitRepositoryNotFound if the specified relative_path is not a Git repository; ErrGitOperationFailed for errors during the 'git pull' command execution (e.g., merge conflicts, authentication failure, network issues); ErrSecurityPath for invalid relative_path.
---

## `tool.Git.Push`
**Description:** Pushes committed changes to a remote repository.

**Category:** Git

**Parameters:**
* `relative_path` (`string`): The relative path within the sandbox to the Git repository.
* `remote_name` (`string`): (optional) Optional. The name of the remote to push to (e.g., 'origin'). Defaults to 'origin'.
* `branch_name` (`string`): (optional) Optional. The name of the local branch to push. Defaults to the current branch.

**Returns:** (`string`) Returns a success message string like 'Successfully pushed to <remote>/<branch> for repository <path>.' Returns nil on error.

**Example:**
```neuroscript
TOOL.Git.Push(relative_path: "my_repo")\nTOOL.Git.Push(relative_path: "my_repo", remote_name: "origin", branch_name: "feature/new-thing")
```

**Error Conditions:** ErrConfiguration if sandbox directory is not set; ErrInvalidArgument if relative_path is missing/invalid type; ErrGitRepositoryNotFound if the specified relative_path is not a Git repository; ErrGitOperationFailed for errors during the 'git push' command (e.g., authentication failure, non-fast-forward, network issues); ErrSecurityPath for invalid relative_path.
---

## `tool.Git.Rm`
**Description:** Removes files from the working tree and from the index in a Git repository.

**Category:** Git

**Parameters:**
* `relative_path` (`string`): Relative path to the repository within the sandbox.
* `paths` (`any`): A single file path (string) or a list of file paths ([]string) to remove relative to the repository root.

**Returns:** (`string`) Returns a success message on successful removal.

**Example:**
```neuroscript
TOOL.Git.Rm(relative_path: "my_repo", paths: "old_file.txt")\nTOOL.Git.Rm(relative_path: "my_repo", paths: ["file1.txt", "dir/file2.txt"])
```

**Error Conditions:** ErrConfiguration if sandbox not set; ErrInvalidArgument for bad args; ErrGitRepositoryNotFound; ErrGitOperationFailed for git command errors; ErrSecurityPath for invalid relative_path.
---

## `tool.Git.Status`
**Description:** Gets the status of the Git repository in the configured sandbox directory.

**Category:** Git

**Parameters:**
* `repo_path` (`string`): (optional) Optional. Relative path to the repository within the sandbox. Defaults to the sandbox root.

**Returns:** (`map`) Returns a map containing Git status information: 'current_branch' (string), 'is_clean' (bool), 'uncommitted_changes' ([]string of changed file paths), 'untracked_files' ([]string of untracked file paths), and 'error' (string, if any occurred internally). See tools_git_status.go for exact structure.

**Example:**
```neuroscript
TOOL.Git.Status() // For sandbox root\nTOOL.Git.Status(repo_path: "my_sub_repo")
```

**Error Conditions:** ErrConfiguration if sandbox directory is not set; ErrGitRepositoryNotFound if the specified path is not a Git repository; ErrIOFailed for underlying Git command execution errors or issues reading Git output; ErrInvalidArgument if repo_path is not a string.
---

## `tool.Go.Build`
**Description:** Runs 'go build' for a specified target in the sandbox. Defaults to './...'.

**Category:** Go Build Tools

**Parameters:**
* `target` (`string`): (optional) Optional. The build target (e.g., a package path or './...'). Defaults to './...'.

**Returns:** (`map`) Returns a map with 'stdout', 'stderr', 'exit_code' (int64), and 'success' (bool) from the 'go build <target>' command.

**Example:**
```neuroscript
TOOL.Go.Build(target: "./cmd/mytool")
```

**Error Conditions:** ErrInvalidArgument if optional target is not a string; ErrConfiguration if sandbox is not set; ErrInternalSecurity for path validation issues. Command execution failures are reported within the returned map.
---

## `tool.Go.Check`
**Description:** Checks Go code validity using 'go list -e -json <target>' within the sandbox. Returns a map indicating success and error details.

**Category:** Go Diagnostics

**Parameters:**
* `target` (`string`): Target Go package path or file path relative to sandbox (e.g., './pkg/core', 'main.go').

**Returns:** (`map`) Returns a map with 'check_success' (bool) and 'error_details' (string). 'check_success' is true if 'go list -e -json' finds no errors in the target's JSON output. 'error_details' contains messages if errors are found or if the command fails.

**Example:**
```neuroscript
TOOL.Go.Check(target: "./pkg/core")
```

**Error Conditions:** ErrConfiguration if sandbox is not set; ErrInternalSecurity for path validation issues. Command execution issues or JSON parsing errors result in 'check_success':false and details in 'error_details'.
---

## `tool.Go.Fmt`
**Description:** Formats Go source code using 'go/format.Source'. Returns the formatted code or an error map.

**Category:** Go Formatting

**Parameters:**
* `content` (`string`): The Go source code content to format.

**Returns:** (`string`) Returns the formatted Go source code as a string. If formatting fails (e.g., syntax error), returns a map {'formatted_content': <original_content>, 'error': <error_string>, 'success': false} and a Go-level error.

**Example:**
```neuroscript
TOOL.Go.Fmt(content: "package main\nfunc main(){}")
```

**Error Conditions:** ErrInternalTool if formatting fails internally, wrapping the original Go error from format.Source. The specific formatting error (e.g. syntax error) is in the 'error' field of the returned map if applicable.
---

## `tool.Go.GetModuleInfo`
**Description:** Finds and parses the go.mod file relevant to a directory by searching upwards. Returns a map with module path, go version, root directory, requires, and replaces, or nil if not found.

**Category:** Go Build Tools

**Parameters:**
* `directory` (`string`): (optional) Directory (relative to sandbox) to start searching upwards for go.mod. Defaults to '.' (sandbox root).

**Returns:** (`map`) Returns a map containing 'modulePath', 'goVersion', 'rootDir' (absolute path to module root), 'requires' (list of maps), and 'replaces' (list of maps). Returns nil if no go.mod is found.

**Example:**
```neuroscript
TOOL.Go.GetModuleInfo(directory: "cmd/mytool")
```

**Error Conditions:** ErrValidationTypeMismatch if directory arg is not a string; ErrInternalSecurity if sandbox is not set or for path validation errors; ErrInternalTool if FindAndParseGoMod fails for reasons other than os.ErrNotExist (e.g., parsing error, file read error). If go.mod is not found, returns nil result and nil error (not a Go-level tool error).
---

## `tool.Go.Imports`
**Description:** Formats Go source code and adjusts imports using 'golang.org/x/tools/imports'. Returns the processed code or an error map.

**Category:** Go Formatting

**Parameters:**
* `content` (`string`): The Go source code content to process.

**Returns:** (`string`) Returns the processed Go source code (formatted and with adjusted imports) as a string. If processing fails, returns a map {'formatted_content': <original_content>, 'error': <error_string>, 'success': false} and a Go-level error.

**Example:**
```neuroscript
TOOL.Go.Imports(content: "package main\nimport \"fmt\"\nfunc main(){fmt.Println(\"hello\")}")
```

**Error Conditions:** ErrInternalTool if goimports processing fails, wrapping the original error from imports.Process. The specific processing error is in the 'error' field of the returned map if applicable.
---

## `tool.Go.ListPackages`
**Description:** Runs 'go list -json' for specified patterns in a target directory. Returns a list of maps, each describing a package.

**Category:** Go Build Tools

**Parameters:**
* `target_directory` (`string`): (optional) Optional. The directory relative to the sandbox root to run 'go list'. Defaults to '.' (sandbox root).
* `patterns` (`slice_string`): (optional) Optional. A list of package patterns (e.g., './...', 'example.com/project/...'). Defaults to ['./...'].

**Returns:** (`slice_map`) Returns a slice of maps, where each map is a JSON object representing a Go package as output by 'go list -json'. Returns an empty slice on command failure or if JSON decoding fails.

**Example:**
```neuroscript
TOOL.Go.ListPackages(target_directory: "pkg/core", patterns: ["./..."])
```

**Error Conditions:** ErrValidationTypeMismatch if patterns arg contains non-string elements; ErrInternalTool if execution helper fails internally or JSON decoding fails; ErrConfiguration if sandbox is not set; ErrInternalSecurity for path validation issues. 'go list' command failures are reported in its output map rather than a Go error from the tool.
---

## `tool.Go.ModTidy`
**Description:** Runs 'go mod tidy' in the sandbox to add missing and remove unused modules. Operates in the sandbox root.

**Category:** Go Build Tools

**Parameters:**
_None_

**Returns:** (`map`) Returns a map with 'stdout', 'stderr', 'exit_code' (int64), and 'success' (bool) from the 'go mod tidy' command execution.

**Example:**
```neuroscript
TOOL.Go.ModTidy()
```

**Error Conditions:** ErrConfiguration if sandbox is not set; ErrInternalSecurity for path validation issues. Command execution failures are reported within the returned map's 'success', 'stderr', and 'exit_code' fields.
---

## `tool.Go.Test`
**Description:** Runs 'go test' for a specified target in the sandbox. Defaults to './...'.

**Category:** Go Build Tools

**Parameters:**
* `target` (`string`): (optional) Optional. The test target (e.g., a package path or './...'). Defaults to './...'.

**Returns:** (`map`) Returns a map with 'stdout', 'stderr', 'exit_code' (int64), and 'success' (bool) from the 'go test <target>' command.

**Example:**
```neuroscript
TOOL.Go.Test(target: "./pkg/feature")
```

**Error Conditions:** ErrInvalidArgument if optional target is not a string; ErrConfiguration if sandbox is not set; ErrInternalSecurity for path validation issues. Command execution failures are reported within the returned map.
---

## `tool.Go.Vet`
**Description:** Runs 'go vet' on the specified target(s) in the sandbox to report likely mistakes in Go source code. Defaults to './...'.

**Category:** Go Diagnostics

**Parameters:**
* `target` (`string`): (optional) Optional. The target for 'go vet' (e.g., a package path or './...'). Defaults to './...'.

**Returns:** (`map`) Returns a map with 'stdout', 'stderr', 'exit_code' (int64), and 'success' (bool) from the 'go vet <target>' command. 'stderr' usually contains the vet diagnostics.

**Example:**
```neuroscript
TOOL.Go.Vet(target: "./pkg/core")
```

**Error Conditions:** ErrInvalidArgument if optional target is not a string; ErrConfiguration if sandbox is not set; ErrInternalSecurity for path validation issues. Command execution failures are reported within the returned map.
---

## `tool.HasPrefix`
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

**Error Conditions:** Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `input_string` or `prefix` are not strings.
---

## `tool.HasSuffix`
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

**Error Conditions:** Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `input_string` or `suffix` are not strings.
---

## `tool.Input`
**Description:** Displays a message and waits for user input from standard input. Returns the input as a string.

**Category:** Input/Output

**Parameters:**
* `message` (`string`): (optional) The message to display to the user before waiting for input. If null or empty, no prompt message is printed.

**Returns:** (`string`) Returns the string entered by the user, with trailing newline characters trimmed. Returns an empty string and an error if reading input fails.

**Example:**
```neuroscript
userName = TOOL.Input(message: "Enter your name: ")
```

**Error Conditions:** ErrorCodeType if the prompt message argument is provided but not a string; ErrorCodeIOFailed if reading from standard input fails (e.g., EOF).
---

## `tool.Join`
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

**Error Conditions:** Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `string_list` is not a list of strings or `separator` is not a string.
---

## `tool.Length`
**Description:** Returns the number of UTF-8 characters (runes) in a string.

**Category:** String Operations

**Parameters:**
* `input_string` (`string`): The string to measure.

**Returns:** (`int`) Returns an integer representing the number of runes in the input string.

**Example:**
```neuroscript
tool.Length("hello") // Returns 5
```

**Error Conditions:** Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `input_string` is not a string.
---

## `tool.LineCount`
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

**Error Conditions:** Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `content_string` is not a string.
---

## `tool.List.Append`
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

**Error Conditions:** None expected, as input validation ensures 'list' is a slice. Appending 'nil' is allowed.
---

## `tool.List.Contains`
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

**Error Conditions:** None expected. Comparison with 'nil' elements is handled.
---

## `tool.List.Get`
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

**Error Conditions:** Returns nil or the default value if the index is out of bounds. No specific error type is returned for out-of-bounds access by design.
---

## `tool.List.Head`
**Description:** Returns the first element of the list, or nil if the list is empty.

**Category:** List Operations

**Parameters:**
* `list` (`slice_any`): The list to get the head from.

**Returns:** (`any`) Returns the first element of the 'list'. If the list is empty, it returns nil.

**Example:**
```neuroscript
tool.List.Head([1, 2, 3]) // returns 1\ntool.List.Head([]) // returns nil
```

**Error Conditions:** None expected. Returns nil for an empty list.
---

## `tool.List.IsEmpty`
**Description:** Returns true if the list has zero elements, false otherwise.

**Category:** List Operations

**Parameters:**
* `list` (`slice_any`): The list to check.

**Returns:** (`bool`) Returns true if the 'list' contains zero elements, and false otherwise.

**Example:**
```neuroscript
tool.List.IsEmpty([]) // returns true\ntool.List.IsEmpty([1]) // returns false
```

**Error Conditions:** None expected.
---

## `tool.List.Length`
**Description:** Returns the number of elements in a list.

**Category:** List Operations

**Parameters:**
* `list` (`slice_any`): The list to measure.

**Returns:** (`int`) Returns an integer representing the number of elements in the provided list.

**Example:**
```neuroscript
tool.List.Length([1, 2, 3]) // returns 3
```

**Error Conditions:** None expected, as input validation ensures 'list' is a slice. An empty list returns 0.
---

## `tool.List.Prepend`
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

**Error Conditions:** None expected, as input validation ensures 'list' is a slice. Prepending 'nil' is allowed.
---

## `tool.List.Rest`
**Description:** Returns a *new* list containing all elements except the first. Returns an empty list if the input list has 0 or 1 element.

**Category:** List Operations

**Parameters:**
* `list` (`slice_any`): The list to get the rest from.

**Returns:** (`slice_any`) Returns a new list containing all elements of the input 'list' except the first. If the list has 0 or 1 element, it returns an empty list. The original list is not modified.

**Example:**
```neuroscript
tool.List.Rest([1, 2, 3]) // returns [2, 3]\ntool.List.Rest([1]) // returns []
```

**Error Conditions:** None expected. Returns an empty list for lists with 0 or 1 element.
---

## `tool.List.Reverse`
**Description:** Returns a *new* list with the elements in reverse order.

**Category:** List Operations

**Parameters:**
* `list` (`slice_any`): The list to reverse.

**Returns:** (`slice_any`) Returns a new list with the elements of the input 'list' in reverse order. The original list is not modified.

**Example:**
```neuroscript
tool.List.Reverse([1, 2, 3]) // returns [3, 2, 1]
```

**Error Conditions:** None expected.
---

## `tool.List.Slice`
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

**Error Conditions:** Returns an empty list for invalid or out-of-bounds start/end indices. Does not return an error for range issues.
---

## `tool.List.Sort`
**Description:** Returns a *new* list with elements sorted. Restricted to lists containing only numbers (int/float) or only strings. Throws error for mixed types or non-sortable types (nil, bool, list, map).

**Category:** List Operations

**Parameters:**
* `list` (`slice_any`): The list to sort.

**Returns:** (`slice_any`) Returns a new list with elements sorted. The list must contain either all numbers (integers or floats, which will be sorted numerically) or all strings (sorted lexicographically). The original list is not modified. Returns an empty list if the input list is empty.

**Example:**
```neuroscript
tool.List.Sort([3, 1, 2]) // returns [1, 2, 3]\ntool.List.Sort(["c", "a", "b"]) // returns ["a", "b", "c"]
```

**Error Conditions:** Returns an error (ErrListCannotSortMixedTypes) if the list contains mixed types (e.g., numbers and strings), nil elements, or other non-sortable types like booleans, maps, or other lists.
---

## `tool.List.Tail`
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

**Error Conditions:** None expected. Handles various 'count' values appropriately, returning an empty list or a copy of the whole list as applicable.
---

## `tool.ListAPIFiles`
**Description:** Lists files currently available via the platform's File API.

**Parameters:**
_None_

**Returns:** (`slice_any`) 
---

## `tool.Meta.GetToolSpecificationsJSON`
**Description:** Provides a JSON string containing an array of all currently available tool specifications. Each object in the array represents a tool and includes its name, description, category, arguments (with their details), return type, return help, variadic status, example usage, and error conditions.

**Category:** Introspection

**Parameters:**
_None_

**Returns:** (`string`) A JSON string representing an array of ToolSpec objects. This is intended for programmatic use or detailed inspection of all tool capabilities.

**Example:**
```neuroscript
TOOL.Meta.GetToolSpecificationsJSON()
```

**Error Conditions:** Returns an error (ErrorCodeInternal) if JSON marshalling of the tool specifications fails. Generally does not return other errors unless the ToolRegistry is uninitialized (ErrorCodeConfiguration).
---

## `tool.Meta.ListTools`
**Description:** Provides a compact text list (sorted alphabetically) of all currently available tools, including basic parameter information. Each tool is listed on a new line, showing its name, parameters (name:type), and return type. Example: FS.Read(filepath:string) -> string

**Category:** Introspection

**Parameters:**
_None_

**Returns:** (`string`) A string containing a newline-separated list of tool names, their parameters (name:type), and return types.

**Example:**
```neuroscript
TOOL.Meta.ListTools()
```

**Error Conditions:** Generally does not return errors, unless the ToolRegistry is uninitialized (which would be an ErrorCodeConfiguration if an attempt is made to call it in such a state).
---

## `tool.Meta.ToolsHelp`
**Description:** Provides a more extensive, Markdown-formatted list of available tools, including descriptions, parameters, and return types. Can be filtered by providing a partial tool name. Details include parameter names, types, descriptions, and return type with its description.

**Category:** Introspection

**Parameters:**
* `filter` (`string`): (optional) An optional string to filter tool names. Only tools whose names contain this substring will be listed. If empty or omitted, all tools are listed.

**Returns:** (`string`) A string in Markdown format detailing available tools, their descriptions, parameters, and return types. Output can be filtered by the optional 'filter' argument.

**Example:**
```neuroscript
TOOL.Meta.ToolsHelp(filter: "FS")
TOOL.Meta.ToolsHelp()
```

**Error Conditions:** Returns ErrorCodeType if the 'filter' argument is provided but is not a string. Generally does not return other errors, unless the ToolRegistry is uninitialized (ErrorCodeConfiguration).
---

## `tool.Modulo`
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

**Error Conditions:** Returns 'ErrDivisionByZero' if num2 is 0. Returns an 'ErrInternalTool' if arguments cannot be processed as int64 (should be caught by validation).
---

## `tool.Multiply`
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

**Error Conditions:** Returns an 'ErrInternalTool' if arguments cannot be processed as float64 (should be caught by validation).
---

## `tool.Print`
**Description:** Prints values to the standard output. If multiple values are passed in a list, they are printed space-separated.

**Category:** Input/Output

**Parameters:**
* `values` (`any`): A single value or a list of values to print. List elements will be space-separated.

**Returns:** (`nil`) Returns nil. This tool is used for its side effect of printing to standard output.
**Variadic:** Yes

**Example:**
```neuroscript
TOOL.Print(value: "Hello World")\nTOOL.Print(values: ["Hello", 42, "World!"]) // Prints "Hello 42 World!"
```

**Error Conditions:** ErrArgumentMismatch if the internal 'values' argument is not provided as expected by the implementation.
---

## `tool.Replace`
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

**Error Conditions:** Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `input_string`, `old_substring`, or `new_substring` are not strings, or if `count` is not an integer.
---

## `tool.Shell.Execute`
**Description:** Executes an arbitrary shell command. WARNING: Use with extreme caution due to security risks. Command path validation is basic. Consider using specific tools (e.g., GoBuild, GitAdd) instead.

**Category:** Shell Operations

**Parameters:**
* `command` (`string`): The command or executable path (must not contain path separators like '/' or '\').
* `args_list` (`slice_string`): (optional) A list of string arguments for the command.
* `directory` (`string`): (optional) Optional directory (relative to sandbox) to execute the command in. Defaults to sandbox root.

**Returns:** (`map`) Returns a map containing 'stdout' (string), 'stderr' (string), 'exit_code' (int), and 'success' (bool) of the executed command. 'success' is true if the command exits with code 0, false otherwise. The command is executed within the sandboxed environment.

**Example:**
```neuroscript
tool.Shell.Execute("ls", ["-la"], "my_directory")
```

**Error Conditions:** Returns `ErrArgumentMismatch` if an incorrect number of arguments is provided. Returns `ErrInvalidArgument` or `ErrorCodeType` if 'command' is not a string, 'args_list' is not a list of strings, or 'directory' is not a string. Returns `ErrSecurityViolation` if the 'command' path is deemed suspicious (e.g., contains path separators or shell metacharacters). Returns `ErrInternal` if the internal FileAPI is not available. May return path-related errors (e.g., `ErrFileNotFound`, `ErrPathNotDirectory`, `ErrPermissionDenied`) if the specified 'directory' is invalid or inaccessible. If the command itself executes but fails (non-zero exit code), 'success' in the result map will be false, and 'stderr' may contain error details. OS-level execution errors are also captured in 'stderr'.
---

## `tool.Split`
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

**Error Conditions:** Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `input_string` or `delimiter` are not strings.
---

## `tool.SplitWords`
**Description:** Splits a string into words based on whitespace.

**Category:** String Operations

**Parameters:**
* `input_string` (`string`): The string to split into words.

**Returns:** (`slice_string`) Returns a slice of strings, where each string is a word from the input string, with whitespace removed. Multiple spaces are treated as a single delimiter.

**Example:**
```neuroscript
tool.SplitWords("hello world  example") // Returns ["hello", "world", "example"]
```

**Error Conditions:** Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `input_string` is not a string.
---

## `tool.Staticcheck`
**Description:** Runs 'staticcheck' on the specified target(s) in the sandbox. Reports bugs, stylistic errors, and performance issues. Defaults to './...'. Assumes 'staticcheck' is in PATH.

**Category:** Go Diagnostics

**Parameters:**
* `target` (`string`): (optional) Optional. The target for 'staticcheck' (e.g., a package path or './...'). Defaults to './...'.

**Returns:** (`map`) Returns a map with 'stdout', 'stderr', 'exit_code' (int64), and 'success' (bool) from the 'staticcheck <target>' command. 'stdout' usually contains the diagnostics.

**Example:**
```neuroscript
TOOL.Staticcheck(target: "./...")
```

**Error Conditions:** ErrInvalidArgument if optional target is not a string; ErrToolExecutionFailed if 'staticcheck' command fails (e.g. not found, or internal error), reported via the toolExecuteCommand structure.
---

## `tool.Substring`
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

**Error Conditions:** Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if arguments are not of the correct type. Returns `ErrListIndexOutOfBounds` (with `ErrorCodeBounds`) if `start_index` or `length` are negative.
---

## `tool.Subtract`
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

**Error Conditions:** Returns an 'ErrInternalTool' if arguments cannot be processed as float64 (should be caught by validation).
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

**Category:** String Operations

**Parameters:**
* `input_string` (`string`): The string to convert.

**Returns:** (`string`) Returns the lowercase version of the input string.

**Example:**
```neuroscript
tool.ToLower("HELLO") // Returns "hello"
```

**Error Conditions:** Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `input_string` is not a string.
---

## `tool.ToUpper`
**Description:** Converts a string to uppercase.

**Category:** String Operations

**Parameters:**
* `input_string` (`string`): The string to convert.

**Returns:** (`string`) Returns the uppercase version of the input string.

**Example:**
```neuroscript
tool.ToUpper("hello") // Returns "HELLO"
```

**Error Conditions:** Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `input_string` is not a string.
---

## `tool.Tree.AddChildNode`
**Description:** Adds a new child node to an existing parent node. Returns the ID of the newly created child node.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle for the tree structure.
* `parent_node_id` (`string`): ID of the node that will become the parent.
* `new_node_id_suggestion` (`string`): (optional) Optional suggested unique ID for the new node. If empty or nil, an ID will be auto-generated. Must be unique if provided.
* `node_type` (`string`): Type of the new child (e.g., 'object', 'array', 'string', 'number', 'boolean', 'null', 'checklist_item').
* `value` (`any`): (optional) Initial value if the node_type is a leaf or simple type. Ignored for 'object' and 'array' types.
* `key_for_object_parent` (`string`): (optional) If the parent is an 'object' node, this key is used to link the new child in the parent's attributes. Required for object parents.

**Returns:** (`string`) Returns the string ID of the newly created child node.

**Example:**
```neuroscript
handle = tool.Tree.LoadJSON("{\"root\":{}}"); tool.Tree.AddChildNode(handle, "actual_root_id", "newChildNodeID", "string", "new child value", "childAttributeKey")
```

**Error Conditions:** Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad/missing arguments (e.g., invalid `node_type`, missing `key_for_object_parent` when parent is 'object'). Returns `ErrTreeNotFound` if handle is invalid. Returns `ErrNodeNotFound` if `parent_node_id` does not exist. Returns `ErrNodeWrongType` if parent node type cannot accept children in the specified manner. Returns `ErrNodeIDExists` (with `ErrorCodeTreeConstraintViolation`) if `new_node_id_suggestion` (if provided) already exists.
---

## `tool.Tree.FindNodes`
**Description:** Finds nodes within a tree (starting from a specified node) that match specific criteria. Returns a list of matching node IDs.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.
* `start_node_id` (`string`): ID of the node within the tree to start searching from. The search includes this node.
* `query_map` (`map`): Map defining search criteria. Supported keys: 'id' (string), 'type' (string), 'value' (any), 'attributes' (map of string:string for child node ID checks), 'metadata' (map of string:string for direct string value metadata checks). Other keys are treated as direct metadata attribute checks.
* `max_depth` (`int`): (optional) Maximum depth to search relative to the start node (0 for start node only, -1 for unlimited). Default: -1.
* `max_results` (`int`): (optional) Maximum number of matching node IDs to return (-1 for unlimited). Default: -1.

**Returns:** (`slice_string`) Returns a slice of strings, where each string is a node ID matching the query criteria.

**Example:**
```neuroscript
handle = tool.Tree.LoadJSON("{\"root\":{\"type\":\"folder\", \"data\":{\"id\":\"child1\", \"type\":\"file\"}}}"); tool.Tree.FindNodes(handle, "id_of_root_node", {\"type\":\"file\"})
```

**Error Conditions:** Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad/missing arguments. Returns `ErrTreeNotFound` if handle is invalid. Returns `ErrNodeNotFound` if `start_node_id` does not exist. Returns `ErrTreeInvalidQuery` (with `ErrorCodeArgMismatch`) if `query_map` is malformed (e.g., incorrect value type for a query key). May return `ErrInternal` for other unexpected errors during the recursive search.
---

## `tool.Tree.GetChildren`
**Description:** Gets a list of node IDs of the children of a given 'array' type node. Other node types will result in an error.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.
* `node_id` (`string`): ID of the 'array' type parent node.

**Returns:** (`slice_string`) Returns a slice of strings, where each string is a child node ID from the specified 'array' node. Returns an empty slice if the array node has no children.

**Example:**
```neuroscript
handle = tool.Tree.LoadJSON("{\"myArray\":[{\"id\":\"child1\"}, {\"id\":\"child2\"}]}"); tool.Tree.GetChildren(handle, "id_of_myArray_node") // Returns ["child1", "child2"] if those are their actual IDs.
```

**Error Conditions:** Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad arguments. Returns `ErrTreeNotFound` if handle is invalid. Returns `ErrNodeNotFound` if `node_id` does not exist. Returns `ErrNodeWrongType` if the specified `node_id` is not an 'array' type node.
---

## `tool.Tree.GetNode`
**Description:** Retrieves detailed information about a specific node within a tree, returned as a map. The map includes 'id', 'type', 'value', 'attributes' (map), 'children' (slice of IDs), 'parent_id', and 'parent_attribute_key'.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.
* `node_id` (`string`): The unique ID of the node to retrieve.

**Returns:** (`map`) Returns a map containing details of the specified node. Structure: {'id': string, 'type': string, 'value': any, 'attributes': map[string]string, 'children': []string, 'parent_id': string, 'parent_attribute_key': string}. 'attributes' for non-object nodes will be their metadata. 'children' is primarily for array-like nodes.

**Example:**
```neuroscript
handle = tool.Tree.LoadJSON("{\"root\":{\"child\":\"value\"}}"); tool.Tree.GetNode(handle, "root_node_id") // Replace root_node_id with actual ID of the 'root' node
```

**Error Conditions:** Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad arguments. Returns `ErrTreeNotFound` if the handle is invalid. Returns `ErrNodeNotFound` if `node_id` does not exist in the tree.
---

## `tool.Tree.GetParent`
**Description:** Gets the node ID of the parent of a given node.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.
* `node_id` (`string`): ID of the node whose parent is sought.

**Returns:** (`string`) Returns the string ID of the parent node. Returns nil if the node is the root or has no explicitly set parent (which can occur if the node was detached or is the root).

**Example:**
```neuroscript
handle = tool.Tree.LoadJSON("{\"root\":{\"childKey\": {}}}"); tool.Tree.GetParent(handle, "child_node_id") // Assuming child_node_id is the ID of the node under 'childKey'
```

**Error Conditions:** Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad arguments. Returns `ErrTreeNotFound` if the handle is invalid. Returns `ErrNodeNotFound` if `node_id` does not exist.
---

## `tool.Tree.LoadJSON`
**Description:** Loads a JSON string into a new tree structure and returns a tree handle.

**Category:** Tree Manipulation

**Parameters:**
* `json_string` (`string`): The JSON data as a string.

**Returns:** (`string`) Returns a string handle representing the loaded tree. This handle is used in subsequent tree operations.

**Example:**
```neuroscript
tool.Tree.LoadJSON("{\"name\": \"example\"}") // Returns a tree handle like "tree_handle_XYZ"
```

**Error Conditions:** Returns `ErrArgumentMismatch` for incorrect argument count. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `json_string` is not a string. Returns `ErrTreeJSONUnmarshal` (with `ErrorCodeSyntax`) if JSON parsing fails. Returns `ErrInternal` for failures in tree building or handle registration.
---

## `tool.Tree.RemoveNode`
**Description:** Removes a node (specified by ID) and all its descendants from the tree. Cannot remove the root node.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle to the tree.
* `node_id` (`string`): ID of the node to remove.

**Returns:** (`nil`) Returns nil on success. Removes the node and its descendants.

**Example:**
```neuroscript
handle = tool.Tree.LoadJSON("{\"root\":{\"childKey\": {}}}"); tool.Tree.RemoveNode(handle, "id_of_child_node_under_childKey")
```

**Error Conditions:** Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad arguments. Returns `ErrTreeNotFound` if handle is invalid. Returns `ErrNodeNotFound` if `node_id` does not exist. Returns `ErrCannotRemoveRoot` (with `ErrorCodeTreeConstraintViolation`) if attempting to remove the root node. May return `ErrInternal` for inconsistent tree states (e.g., non-root node without a parent).
---

## `tool.Tree.RemoveNodeMetadata`
**Description:** Removes a metadata attribute (a key-value string pair) from a node.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.
* `node_id` (`string`): ID of the node to remove metadata from.
* `metadata_key` (`string`): The key of the metadata attribute to remove.

**Returns:** (`nil`) Returns nil on success. Removes a metadata key-value pair from the node's attributes.

**Example:**
```neuroscript
handle = tool.Tree.LoadJSON("{\"myNode\":{}}"); tool.Tree.SetNodeMetadata(handle, "id_of_myNode", "customData", "someValue"); tool.Tree.RemoveNodeMetadata(handle, "id_of_myNode", "customData")
```

**Error Conditions:** Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad/empty arguments. Returns `ErrTreeNotFound` if handle is invalid. Returns `ErrNodeNotFound` if `node_id` does not exist. Returns `ErrAttributeNotFound` if the `metadata_key` does not exist in the node's attributes.
---

## `tool.Tree.RemoveObjectAttribute`
**Description:** Removes an attribute (a key mapping to a child node ID) from an 'object' type node. This unlinks the child but does not delete the child node itself.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle for the tree structure.
* `object_node_id` (`string`): Unique ID of the 'object' type node to modify.
* `attribute_key` (`string`): The key (name) of the attribute to remove.

**Returns:** (`nil`) Returns nil on success. Removes the attribute link from the object node.

**Example:**
```neuroscript
handle = tool.Tree.LoadJSON("{\"objNode\":{\"myChildAttribute\":\"some_child_id\"}}"); tool.Tree.RemoveObjectAttribute(handle, "id_of_objNode", "myChildAttribute")
```

**Error Conditions:** Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad/empty arguments. Returns `ErrTreeNotFound` if handle is invalid. Returns `ErrNodeNotFound` if `object_node_id` does not exist. Returns `ErrTreeNodeNotObject` (with `ErrorCodeNodeWrongType`) if `object_node_id` is not an 'object' type. Returns `ErrAttributeNotFound` if the `attribute_key` does not exist on the object node.
---

## `tool.Tree.RenderText`
**Description:** Renders a visual text representation of the entire tree structure identified by the given tree handle.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure to render.

**Returns:** (`string`) Returns a string containing a human-readable, indented text representation of the tree.

**Example:**
```neuroscript
handle = tool.Tree.LoadJSON("{\"a\":{\"b\":\"c\"}}"); tool.Tree.RenderText(handle) // Returns a human-readable text tree
```

**Error Conditions:** Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad arguments. Returns `ErrTreeNotFound` if handle is invalid. May return `ErrInternal` for issues like a missing root node or other unexpected errors during the rendering process.
---

## `tool.Tree.SetNodeMetadata`
**Description:** Sets a metadata attribute as a key-value string pair on any node. This is separate from object attributes that link to child nodes.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.
* `node_id` (`string`): ID of the node to set metadata on.
* `metadata_key` (`string`): The key of the metadata attribute (string).
* `metadata_value` (`string`): The value of the metadata attribute (string).

**Returns:** (`nil`) Returns nil on success. Adds or updates a string key-value pair in the node's metadata attributes.

**Example:**
```neuroscript
handle = tool.Tree.LoadJSON("{\"myNode\":{}}"); tool.Tree.SetNodeMetadata(handle, "id_of_myNode", "version", "1.0")
```

**Error Conditions:** Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad/empty arguments. Returns `ErrTreeNotFound` if handle is invalid. Returns `ErrNodeNotFound` if `node_id` does not exist.
---

## `tool.Tree.SetObjectAttribute`
**Description:** Sets or updates an attribute on an 'object' type node, mapping the attribute key to an existing child node's ID. This is for establishing parent-child relationships in object nodes.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle for the tree structure.
* `object_node_id` (`string`): Unique ID of the 'object' type node to modify.
* `attribute_key` (`string`): The key (name) of the attribute to set.
* `child_node_id` (`string`): The ID of an *existing* node within the same tree to associate with the key.

**Returns:** (`nil`) Returns nil on success. Sets an attribute on the object node, linking it to the child node.

**Example:**
```neuroscript
handle = tool.Tree.LoadJSON("{\"objNode\":{}, \"childNode\":{}}"); tool.Tree.SetObjectAttribute(handle, "id_of_objNode", "myChildAttribute", "id_of_childNode")
```

**Error Conditions:** Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad/empty arguments. Returns `ErrTreeNotFound` if handle is invalid. Returns `ErrNodeNotFound` if `object_node_id` or `child_node_id` does not exist. Returns `ErrTreeNodeNotObject` (with `ErrorCodeNodeWrongType`) if `object_node_id` does not refer to an 'object' type node.
---

## `tool.Tree.SetValue`
**Description:** Sets the value of an existing leaf or simple-type node (e.g., string, number, boolean, null, checklist_item). Cannot set value on 'object' or 'array' type nodes using this tool.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.
* `node_id` (`string`): ID of the leaf or simple-type node to modify.
* `value` (`any`): The new value for the node.

**Returns:** (`nil`) Returns nil on success. Modifies the node's value in place.

**Example:**
```neuroscript
handle = tool.Tree.LoadJSON("{\"keyNode\":\"old_value\"}"); tool.Tree.SetValue(handle, "id_of_keyNode", "new_value")
```

**Error Conditions:** Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad arguments. Returns `ErrTreeNotFound` if the handle is invalid. Returns `ErrNodeNotFound` if `node_id` does not exist. Returns `ErrCannotSetValueOnType` (with `ErrorCodeTreeConstraintViolation`) if attempting to set value on an 'object' or 'array' node.
---

## `tool.Tree.ToJSON`
**Description:** Converts a tree structure (identified by tree handle) back into a JSON string. Output is pretty-printed.

**Category:** Tree Manipulation

**Parameters:**
* `tree_handle` (`string`): Handle to the tree structure.

**Returns:** (`string`) Returns a pretty-printed JSON string representation of the tree.

**Example:**
```neuroscript
handle = tool.Tree.LoadJSON("{\"key\":\"value\"}"); tool.Tree.ToJSON(handle) // Returns a pretty-printed JSON string.
```

**Error Conditions:** Returns `ErrArgumentMismatch` or `ErrInvalidArgument` (with `ErrorCodeType`) for bad arguments. Returns `ErrTreeNotFound` if the handle is invalid. Returns `ErrTreeJSONMarshal` (with `ErrorCodeInternal`) if marshalling to JSON fails. Returns `ErrInternal` for internal tree consistency issues (e.g., missing root node).
---

## `tool.TrimSpace`
**Description:** Removes leading and trailing whitespace from a string.

**Category:** String Operations

**Parameters:**
* `input_string` (`string`): The string to trim.

**Returns:** (`string`) Returns the string with leading and trailing whitespace removed.

**Example:**
```neuroscript
tool.TrimSpace("  hello  ") // Returns "hello"
```

**Error Conditions:** Returns `ErrArgumentMismatch` if the wrong number of arguments is provided. Returns `ErrInvalidArgument` (with `ErrorCodeType`) if `input_string` is not a string.
---

## `tool.UploadFile`
**Description:** Uploads a local file (from the sandbox) to the platform's File API. Returns a map describing the uploaded file.

**Parameters:**
* `local_filepath` (`string`): Relative path (within the sandbox) of the local file to upload.
* `api_display_name` (`string`): (optional) Optional display name for the file on the API.

**Returns:** (`map`) 
---


[green]Initial script completed.[-]
