 :: title: NeuroScript Road to v0.4.0 Checklist
 :: version: 004
 :: id: ns-roadmap-v0.4.0
 :: status: draft
 :: description: Prioritized tasks to implement NeuroScript v0.4.0, including fleshed-out Work Queue design and 'must' keyword enhancements.
 :: updated: 2025-06-03
 
 # NeuroScript v0.4.0 Development Tasks
 
 ## Vision (Summary)
 
 Achieve an AI-driven development loop where a Supervising AI (SAI) interacts with NeuroScript (ng) and a File API to:
 - Apply a prompt to each file in a file tree. 
 (We now have a basic human/ai chat facility in ng, which is great -- its been a long fight.
 
 I want to get the the point where we can deliver a tree of files to a set of AI workers, have them apply instructions to the files (code initially) then have the files updated on the local drive.
 
 For this, we need the ability to queue jobs to a work queue, then have the wm dispatch them to AIs, potentially allow the AIs to use local tools, and pass back the processed files write out to disk.
 
 I am thinking this should mostly be a ns that ng runs.)
 
 
 - Coordinate tasks, potentially using worker AIs.
 - Track progress via shared checklists (Checklist tools).
 
 ## Checklist
 
 - [ ] 1. Language Features
     - [X] **Standardize Tool Error Returns & Enhance `must` Keyword:** (Corresponds to "simplify returned arg checking")
         - [ ] Define a standard NeuroScript `error` map structure (e.g., `{"code": ..., "message": ..., "details": ...}`) for tools to return on operational errors.
         - [ ] Update tool implementation guidelines: tools must return this standard `error` map for handled operational errors. `ToolSpec.ReturnType` defines the success type.
         - [ ] Implement/formalize `set <variable> = must <tool_call_expression>`:
             - Fails script (panics to `on_error`) if the tool call results in a Go-level error OR if the tool returns a standard `error` map.
             - Otherwise, assigns the success value to `<variable>`.
         - [ ] Implement `set <variable> = must <map_variable>["<key_name>"] as <expected_type>` (Single Key):
             - Fails script (panics to `on_error`) if `<map_variable>` is an `error` map or nil, key is missing, value is an `error` map (unless `<expected_type>` is `error`), or value's type doesn't match `<expected_type>` (`string`, `int`, `list`, `map`, etc.).
             - Assigns validated value to `<variable>`.
         - [ ] Implement `set <var1>, <var2>, ... = must <map_variable>[<key1_str>, <key2_str>, ...] as <type1>, <type2>, ...` (Multiple Keys):
             - Fails script (panics to `on_error`) if `<map_variable>` is an `error` map or nil, if counts of vars/keys/types don't match, or if any key is missing, its value is an unexpected `error` map, or its type mismatches the corresponding `<type_i>`.
             - Assigns all validated values to respective variables atomically.
         - [ ] Implement `must IsCorrectShape(<variable>, <shape_definition>)` statement:
             - Introduce `IsCorrectShape()` built-in function that takes a variable and a shape definition (string keyword like `"list_of_string"` or a schema map).
             - `must IsCorrectShape(...)` fails script (panics to `on_error`) if `IsCorrectShape()` returns `false`.
         - [ ] Ensure `on_error` blocks correctly catch panics triggered by all forms of `must`.
         - [ ] Provide a built-in `IsError(value)` helper function/tool to check if a value is a standard `error` map.
     - [-] ~~simplify returned arg checking~~ (Superseded by "Standardize Tool Error Returns & Enhance `must` Keyword" above)
 
 - [ ] 2. WM (Worker Management) System
     - [X] **Define Core Work Queue Abstraction (v0.4.0 Scope - In-Memory):**
         - [ ] **Job Submission:** Design `Job` structure. Implement `tool.AIWM.AddJobToQueue(queue_name, job_payload_map)` for scripts to add jobs. Queues are in-memory for this release.
         - [ ] **Worker Assignment:**
             - [ ] Allow workers to be assigned to specific queues (manual or programmatic).
             - [ ] Implement `WorkQueue` configuration for default worker definition ID and desired number of workers.
             - [ ] AIWM to auto-instantiate/manage default workers for a queue based on its configuration and load (scaled within min/max pool size if defined).
         - [ ] **Job Lifecycle & State:**
             - [ ] Queues can be explicitly `Start`ed and `Pause`d via tools (e.g., `tool.AIWM.PauseQueue`, `tool.AIWM.ResumeQueue`). Pausing prevents new job dispatch.
             - [ ] Jobs within a queue should have trackable states (e.g., pending, active, completed, failed).
         - [ ] **Result Accumulation & "Lessons Learned" (In-Memory for v0.4.0):**
             - [ ] Each queue instance to accumulate basic operational statistics (jobs processed, success/failure rates, errors).
             - [ ] Each job completion/failure on a queue generates a detailed "lesson record" (map/struct with input, prompt, output, errors, status, etc.).
             - [ ] Queues to hold these lesson records in memory for the current session. (Persistence of lessons is a future enhancement).
         - [ ] **SAI/Human Management Interface (Tools):**
             - [ ] Tools for SAI/human to `Start`/`Pause` queues.
             - [ ] Tools to gather current stats, job statuses, and recent "lessons learned" from a queue.
         - [ ] **TUI Integration:**
             - [ ] Design and implement a TUI panel section to display the status of active work queues (length, active workers, key stats, paused/running state).
         - [ ] **Script Interaction with Queues (v0.4.0 Scope):**
             - [ ] **Querying:**
                 - [ ] `tool.AIWM.ListQueues()`: List available work queue names/IDs.
                 - [ ] `tool.AIWM.GetQueueStatus(queue_name)`: Returns map with stats (length, workers, state, etc.).
                 - [ ] `tool.AIWM.ListJobs(queue_name, filters_map)`: Returns list of job summaries (ID, status).
                 - [ ] `tool.AIWM.GetJobResult(job_id)`: Returns the detailed "lesson record" for a completed/failed job.
             - [ ] **Blocking Event-like Interaction (No full async events for v0.4.0):**
                 - [ ] Design tools that allow scripts to block and wait for specific outcomes, e.g., `tool.AIWM.WaitForJobCompletion(job_id, timeout_ms)` or `tool.AIWM.WaitForNextJobCompletion(queue_name, timeout_ms)`. These tools would return job results/status or throw errors on timeout/failure.
     - [-] ~~Basic work queue with multiple workers~~ (Superseded by the more detailed "Define Core Work Queue Abstraction" above).
     - [-] ~~Result accumulation & Lessons Learned~~ (Integrated into the new Work Queue Abstraction).
     - [ ] **Refine `Apply prompt to each file in a tree` script:** Update the existing NeuroScript for batch file processing to utilize the new Work Queue system. This script will serve as a key test case and example.
     - [ ] Investigate potential for parallel step execution within a single NeuroScript procedure (longer-term research, may not be v0.4.0).
     - [ ] Design for stateful worker interaction and task lifecycle (as it pertains to workers processing jobs from queues).
     - [ ] Tooling for SAI to assign/monitor tasks on workers (covered by queue management and job query tools).
     - [ ] Agent mode permitted, allow & deny lists (Relates to worker capabilities, ensure queue can dispatch to appropriately permissioned workers).
 
 - [ ] 3. Core Interpreter / Runtime
     - [ ] Error Handling: Review `on_error` behavior, especially interaction with return and new `must` behaviors.
     - [ ] Rationalize dratted loggers.
     - [ ] Handle Management: Add explicit `ReleaseHandle` tool? Review potential memory leaks from unreleased handles.
     - [ ] Performance: Basic profiling pass (if major slowdowns observed).
     - [ ] Consider `FS.DiffFiles` tool (or is `NSPatch` sufficient?).
     - [ ] FileAPI Review: Ensure consistency/no overlap between `FS.*` tools and direct `FileAPI`.
     - [ ] `ng` -> FileAPI Sync: Design and implement mechanism/tools.
     - [ ] Build Artifacts: Review GoBuild output handling; add tools if needed (e.g., `FS.Copy`, retrieve artifacts).
 
 - [ ] 6. Tooling & Ecosystem 
     - [ ] Documentation: Update language specification and tool docs for `must` enhancements, `IsCorrectShape`, `error` map convention, and new AIWM queue tools.
     - [ ] Formatting: Begin development of `nsfmt` formatting tool for `.ns` files.
     - [ ] Workflow Test: Create end-to-end test script simulating SAI interaction with Work Queues (add jobs, monitor, retrieve lessons).
 
 - [ ] 7. Example App -- language flashcards (New)
     - [ ] Define core features (add card, review, save/load).
     - [ ] Design data structure (simple list/map, maybe JSON file).
     - [ ] Implement basic TUI or script interaction logic.
 
 - [ ] 9. NS LSP 
     - [ ] (Existing LSP tasks from original doc, ensure LSP understands new `must` syntax and `IsCorrectShape`)