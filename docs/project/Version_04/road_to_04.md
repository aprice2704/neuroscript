:: title: NeuroScript Road to v0.4.0 Checklist
:: version: 004
:: file_version: 5
:: id: ns-roadmap-v0.4.0
:: status: draft
:: description: Tasks for NS v0.4.0 incl. Work-Queue, event grammar, prompt budgeting, FDM bridge, and more.
:: updated: 2025-06-07

# Vision
> **BHAG** — Multiple LLM agents (ChatGPT, Gemini, etc.) maintain the *ns* + *fdm* repos without context-window breakage.  
> Success criteria: ≤ 3 context resets / h, refactor round-trip < 5 min, divergence < 0.25 for ≥ 95 % nodes.

────────────────────────────────────────────────────────────

# 1 · Language Features
- | | **Grammar & Runtime Enhancements**
  - [ ] **Finalize `on event` grammar + dispatcher**  # NEW
  - [ ] `error` value type (+ `IsError` helper)
  - [ ] `timedate` value type
  - [ ] `event` value type
  - [ ] Fuzzy-logic operators
- | | **`must` Keyword & Error-map Standardization**
  - [ ] Define canonical `error` map: `{code, message, details}`
  - [ ] `tool` guidelines: return error map on handled faults
  - [ ] `set x = must tool()`  — fail on Go error **or** error map
  - [ ] Single-key `must map["key"] as type`
  - [ ] Multi-key atomic `must map[...] as …`
  - [ ] `must IsCorrectShape(var, shape_def)`
  - [ ] Ensure `on_error` catches all `must` panics
- | | **Prompt-Budget Helpers**
  - [ ] `prompt.Assemble(node_ids, max_tokens)` built-in  # NEW
  - [ ] `fdm.summary(node_id, max_tokens)` helper  # NEW

────────────────────────────────────────────────────────────

# 2 · Worker-Management System (AIWM)
- | | **Core Work-Queue Abstraction (in-memory v0.4)**
  - [ ] *Job structure* design
  - [ ] `tool.AIWM.AddJobToQueue(queue, job_payload)`
  - [ ] **Queue start/pause** (`PauseQueue`, `ResumeQueue`)
  - | | Worker Assignment
    - [ ] Manual / scripted worker-to-queue mapping
    - [ ] Queue config: default worker def ID, pool size
    - [ ] Auto-instantiate default workers per load
  - | | Job Lifecycle
    - [ ] States: `pending, active, completed, failed`
    - [ ] Track timestamps & retries
  - | | Result Accumulation (“Lessons Learned”)
    - [ ] Per job: store prompt, output, errors, status
    - [ ] Aggregate queue stats in memory
  - | | **Queue Management Tools**
    - [ ] `AIWM.ListQueues`
    - [ ] `AIWM.GetQueueStatus(queue)`
    - [ ] `AIWM.ListJobs(queue, filters)`
    - [ ] `AIWM.GetJobResult(job_id)`
    - [ ] Blocking waits: `WaitForJobCompletion` & `WaitForNextJobCompletion`
- | | **Delegation & Security**
  - [ ] **JobToken schema + signer**  # NEW
  - [ ] Mediator validates JobToken on every FDM tool call  # NEW
  - [ ] Agent allow/deny lists
- | | **NS Worker-Manager Bridge**
  - [ ] `fdm.enqueue_task` (wrap AddJob)
  - [ ] `fdm.wait_job(job_id)`
- | | **End-to-End Test**
  - [ ] Script: split 5 Go files → queue refactors
  - [ ] Worker processes, mediator applies edits
  - [ ] Verify edits & lessons learned stored
- | | **TUI Panel**
  - [ ] Display queue length, active workers, paused/running

────────────────────────────────────────────────────────────

# 3 · Core Interpreter / Runtime
- | | **Concurrency & Conflict Handling**
  - [ ] **Add `version` field to nodes**  # NEW
  - [ ] Mediator rejects stale writes (`ErrConflict`)  # NEW
  - [ ] Optional `on_conflict` block  # NEW
- | | Error Handling & Logging
  - [ ] Review `on_error` with new `must`
  - [ ] Rationalize loggers (single interface)
  - [ ] `ReleaseHandle` tool to prevent leaks
- | | Performance & File/FS Tools
  - [ ] Profiling pass if slowdowns observed
  - [ ] Evaluate `FS.DiffFiles` vs `NSPatch`
  - [ ] GoBuild artifact helpers (`FS.Copy`, etc.)

────────────────────────────────────────────────────────────

# 4 · FDM / NS Integration  # NEW SECTION
- | | **Finder Abstraction**
  - [ ] Mediator implements `fdm.find(query:"...")`
  - [ ] Internal routing: AST → Similarity → text scan
- | | **Summary-Node Pipeline**
  - [ ] Auto-create `file_summary` when file > 512 tokens
  - [ ] NS tool `fdm.summarize_chunk`
- | | **Metrics for BHAG**
  - [ ] Prometheus counter `context_resets_total`
  - [ ] CI smoke test fails if resets > 3 / h

────────────────────────────────────────────────────────────

# 5 · Tooling & Ecosystem
- | | Documentation
  - [ ] Update language spec for error map, `must`, events
  - [ ] Tool docs for AIWM & FDM bridge helpers
- | | Formatter
  - [ ] **`nsfmt`** — wraps long lines, respects `\` continuation
- | | Workflow Tests
  - [ ] Prompt-budget CI test: 2k-line file ⇒ summary ≤ 1024 tokens

────────────────────────────────────────────────────────────

# 6 · Example App — Language Flashcards
- | | Core features: add card, review, save/load
- | | Simple TUI or script

────────────────────────────────────────────────────────────

# 7 · NS LSP
- | | Extend grammar support: `must` enhancements, events, fuzzy
