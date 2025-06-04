:: title: AI Worker Management (AIWM) System Design
:: version: 0.2.0
:: status: draft
:: description: Design for the AI Worker Management system, including Work Queues, Worker Instances, Definitions, and interaction patterns.
:: updated: 2025-06-03

# AI Worker Management (AIWM) System Design

## 1. Overview

The AI Worker Management (AIWM) system is responsible for defining, instantiating, managing, and monitoring AI Worker Instances. It provides a framework for executing tasks using these workers, including managing prompts, inputs, outputs, and performance logging. A key component of this system for v0.4.0 will be the introduction of **Work Queues** to manage and dispatch jobs to workers.

## 2. Core Components

### 2.1. AI Worker Definition (`AIWorkerDefinition`)
(As per existing document - content largely preserved)
* **Source:** Loaded from `.aiwd.json` files or potentially other sources.
* **Contents:**
    * `id`: Unique identifier (e.g., "google-gemini-1.5-pro-code-refactor-v1").
    * `description`: Human-readable description.
    * `model_details`: Information about the underlying LLM (vendor, model name, API version).
    * `system_prompt`: The base system prompt defining the worker's persona/task.
    * `default_config`: Default LLM parameters (temperature, max tokens, etc.).
    * `input_schema`: (Future) Defines expected input structure.
    * `output_schema`: (Future) Defines expected output structure.
    * `tool_permissions`: List of NeuroScript tools the worker is allowed to request/use (if any).
    * `metadata`: Version, author, tags, etc.
* **Management:** CRUD tools for definitions (e.g., `tool.AIWorkerDefinition.Load`, `tool.AIWorkerDefinition.Save`, `tool.AIWorkerDefinition.List`, `tool.AIWorkerDefinition.Get`).

### 2.2. AI Worker Instance (`AIWorkerInstance`)
(As per existing document - content largely preserved, with notes on relation to queues)
* Represents a running, stateful instance of an `AIWorkerDefinition`.
* Can handle multiple tasks sequentially or maintain a conversational context.
* May be explicitly created by a script/SAI or managed by a Work Queue's worker pool.
* **Attributes:**
    * `instance_id`: Unique ID for the running instance.
    * `definition_id`: The definition it's based on.
    * `status`: (e.g., idle, busy, error, retired).
    * `current_task_id`: If busy.
    * `llm_client_instance`: The actual client to interact with the LLM.
    * `performance_summary`: Aggregate performance data for tasks it has run.
    * `token_usage`: Aggregate token counts.
* **Management:** Tools for spawning, retiring, listing, and getting status. Workers might be tied to specific Work Queues.

### 2.3. Work Queues (New Detailed Section for v0.4.0)

Work Queues are central to managing collections of jobs and distributing them to available AI Worker Instances. For v0.4.0, queues will be in-memory.

* **Purpose:**
    * Decouple job submission from immediate execution.
    * Enable batch processing and parallel execution of tasks by multiple workers.
    * Provide a central point for accumulating statistics and "lessons learned" from job processing.
    * Allow for management (pause/resume) of job flows.

* **Attributes & Functionality:**
    * **Identity:** Each queue will have a unique name or ID.
    * **Job Storage (In-Memory for v0.4.0):**
        * A mechanism to hold submitted `Job` objects/payloads.
        * Jobs will have priorities (basic implementation for v0.4.0, e.g., FIFO with simple priority tiers).
    * **Worker Association:**
        * Workers (AIWorkerInstances) can be explicitly assigned to a queue.
        * **Default Worker Pool:** A queue can be configured with:
            * A default `AIWorkerDefinitionID`.
            * A desired number of worker instances (e.g., min/max or target active).
            * The AIWM will be responsible for automatically instantiating and managing these default workers, scaling them based on queue load (rudimentary scaling for v0.4.0).
    * **State Management:**
        * Queues can be `Started` (actively dispatching jobs) or `Paused` (jobs are accepted but not dispatched).
    * **Statistics & Lessons Learned (In-Memory for v0.4.0):**
        * **Stats:** The queue will accumulate operational statistics:
            * Number of jobs submitted, pending, active, completed, failed.
            * Success/failure rates.
            * Average processing time per job (approximate).
            * Aggregate costs (if LLM provides cost per call and jobs record it).
        * **"Lessons Learned" Records:** For each job processed, a detailed record (map/struct) will be generated and stored by the queue (in memory for this version). This record will include:
            * Job ID, submitted payload (prompt, inputs).
            * Worker Definition ID, Worker Instance ID used.
            * Timestamp of start/end.
            * Raw output from the AI worker.
            * Any errors reported by the worker or during processing.
            * Status of the job (e.g., `success_refactored`, `fail_ai_errors_persisted`, `fail_tool_error` as discussed for the file refactoring script).
            * For tasks like code refactoring: initial errors, post-refactor errors (if applicable).
            * (Persistence of these detailed lessons to a file or database is a future enhancement beyond v0.4.0).
    * **TUI Visibility:** Queue status (length, active workers, key stats, paused/running state) should be displayable in a dedicated AIWM panel in the NeuroScript TUI.

* **Management & Interaction (for SAI, Human via TUI/CLI, and Scripts):**
    * **NeuroScript Tools for Queue Management:**
        * `tool.AIWM.CreateQueue(queue_name, optional_default_worker_config_map)`
        * `tool.AIWM.DeleteQueue(queue_name)`
        * `tool.AIWM.PauseQueue(queue_name)`
        * `tool.AIWM.ResumeQueue(queue_name)`
        * `tool.AIWM.StartQueue(queue_name)` (if queues can be created in a non-started state)
    * **NeuroScript Tools for Job Submission & Querying:**
        * `tool.AIWM.AddJobToQueue(queue_name, job_payload_map)`: Submits a new job. `job_payload_map` would include target worker definition, prompt, input data, priority, etc.
        * `tool.AIWM.ListQueues()`: Returns a list of available queue names/IDs.
        * `tool.AIWM.GetQueueStatus(queue_name)`: Returns a map with current statistics (length, workers, state, error rates, etc.).
        * `tool.AIWM.ListJobs(queue_name, filters_map)`: Returns a list of job summaries (ID, status, priority, submission time) for a queue, with optional filters (e.g., by status: "pending", "failed").
        * `tool.AIWM.GetJobResult(job_id)`: Returns the detailed "lesson record" for a specific completed or failed job.
        * `tool.AIWM.CancelJob(job_id)` (if job is cancellable, e.g. still pending).
    * **Script Interaction for "Events" (Blocking Calls for v0.4.0):**
        * Full asynchronous event handling is postponed.
        * For v0.4.0, provide blocking tools that allow scripts to wait for specific outcomes:
            * `tool.AIWM.WaitForJobCompletion(job_id, timeout_ms)`: Blocks until the specified job completes (success or failure) or timeout. Returns job result/lesson.
            * `tool.AIWM.WaitForNextJobCompletion(queue_name, timeout_ms)`: Blocks until *any* job in the specified queue completes. Returns job result/lesson.
            * These tools will throw errors on timeout or if the queue/job cannot be monitored.

### 2.4. Job/Task Structure
(This section might need expansion or merging with Work Queue details)
* A `Job` (or `TaskPayload`) needs to be clearly defined. It will encapsulate:
    * `JobID`: Unique identifier.
    * `TargetWorkerDefinitionID`: Which type of worker should process this.
    * `SpecificWorkerInstanceID` (optional): Request a particular instance if applicable and available.
    * `Prompt`: The main instruction or question.
    * `InputData`: A map or structured data for the worker.
    * `ConfigOverrides`: Map of LLM parameters to override defaults for this job.
    * `Priority`: Integer or enum.
    * `Metadata`: Submission time, submitter ID, tags, etc.
    * `Status`: Pending, Active, CompletedSuccess, CompletedFailure, Cancelled.
    * `Result`: The output from the worker, or error details. (This would be part of the "lesson record").

### 2.5. Stateless Task Execution
(As per existing document - `tool.AIWorker.ExecuteStatelessTask` - this will likely remain as a direct way to use workers, bypassing queues for simple, immediate tasks. Queues are for managing multiple, potentially longer-running, or deferrable jobs.)

## 3. Interaction Patterns & Workflows

### 3.1. Batch Processing via Work Queue (e.g., Refactoring Script)
1.  A NeuroScript (e.g., the "UpdateNsSyntaxWM.ns" script) identifies multiple items to process (e.g., files needing refactoring).
2.  For each item, it constructs a `job_payload_map` (containing file content, specific instructions, target worker definition like "ns-refactor-v1").
3.  It submits each job to a designated `WorkQueue` using `tool.AIWM.AddJobToQueue`.
4.  The AIWM, observing the queue, dispatches these jobs to available `AIWorkerInstance`s (either dedicated or from the queue's default pool).
5.  As jobs complete, the `WorkQueue` collects their results and detailed "lesson records".
6.  The originating script (or an SAI) can:
    * Poll the queue status using `tool.AIWM.GetQueueStatus`.
    * Use `tool.AIWM.WaitForNextJobCompletion` or `tool.AIWM.WaitForJobCompletion` to process results as they come in or wait for all.
    * Retrieve detailed outcomes using `tool.AIWM.GetJobResult` for each completed job.
    * This data forms the basis for the SAI loop (analyze lessons, update prompts/definitions, etc.).

### 3.2. SAI-Managed Operations
* An SAI (which could be another NeuroScript) monitors queue statistics and lesson records.
* If a queue shows high failure rates for a particular type of job or with a specific worker definition, the SAI could:
    * Pause the queue.
    * Alert a human.
    * Attempt to modify the default worker definition's system prompt.
    * Re-queue failed jobs with modified parameters or to a different queue/worker type.
    * Adjust the number of active workers for a queue based on load or cost considerations.

## 4. Future Considerations (Beyond v0.4.0)
* **Persistence:** Persistent storage for Work Queues (job state) and "Lessons Learned" database.
* **Advanced Scheduling:** More sophisticated job prioritization, scheduling strategies (e.g., round-robin, load balancing).
* **Asynchronous Event System:** Full event-driven model for scripts to react to queue/job events without polling/blocking.
* **Distributed Workers:** Support for worker instances running on different machines.
* **Input/Output Schemas:** Enforcing schemas for job inputs and worker outputs for better reliability.

## 5. Open Questions / Design Decisions Pending
* Detailed error handling and retry mechanisms for jobs within the queue.
* Specific structure and content of the "lesson record" beyond the initial sketch.
* Mechanism for workers to report fine-grained progress on long tasks.