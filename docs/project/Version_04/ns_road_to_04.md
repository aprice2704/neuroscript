:: title: NeuroScript Road to v0.4.0 Checklist
:: version: 004
:: file_version: 7
:: id: ns-roadmap-v0.4.0
:: status: draft
:: description: Tasks for NS v0.4.0 incl. Work-Queue, event grammar, prompt budgeting, FDM bridge, and more.
:: updated: 2025-06-27
:: review_comment: Merged WM rename + Manual Query MVP (2025-06-27).

# Vision
> **BHAG** — Multiple LLM agents (ChatGPT, Gemini, etc.) maintain the *ns* + *fdm* repos without context-window breakage.  
> Success criteria: ≤ 3 context resets / h, refactor round-trip < 5 min, divergence < 0.25 for ≥ 95 % nodes.

────────────────────────────────────────────────────────────

# 1 · Language Features
- | | **Grammar & Runtime Enhancements**
  - [ ] **Finalize `on event` grammar + dispatcher**
  - [x] `error` value type (+ `IsError` helper)
  - [x] `timedate` value type
  - [x] `event` value type
  - [x] Fuzzy-logic operators
- | | **`must` Keyword & Error-map Standardization**
  - [x] Define canonical `error` map: `{code, message, details}`
  - [x] `tool` guidelines: return error map on handled faults
  - [x] `set x = must tool()` — fail on Go error **or** error map
  - [x] Single-key `must map["key"] as type`
  - [x] Multi-key atomic `must map[...] as …`
  - [ ] `must IsCorrectShape(var, shape_def)`
  - [x] Ensure `on_error` catches all `must` panics
- | | **Prompt-Budget Helpers**
  - [ ] `prompt.Assemble(node_ids, max_tokens)` built-in
  - [ ] `fdm.summary(node_id, max_tokens)` helper

────────────────────────────────────────────────────────────

# 2 · Worker-Management System (WM)
- | | **Core Work-Queue Abstraction (in-memory v0.4)**
  - [ ] *Job structure* design
  - [ ] `tool.WM.AddJobToQueue(queue, job_payload)`
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
    - [ ] `WM.ListQueues`
    - [ ] `WM.GetQueueStatus(queue)`
    - [ ] `WM.ListJobs(queue, filters)`
    - [ ] `WM.GetJobResult(job_id)`
    - [ ] Blocking waits: `WaitForJobCompletion` & `WaitForNextJobCompletion`
- | | **Delegation & Security**
  - [ ] **JobToken schema + signer**
  - [ ] Mediator validates JobToken on every FDM tool call
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
  - [ ] **Add `version` field to nodes**
  - [ ] Mediator rejects stale writes (`ErrConflict`)
  - [ ] Optional `on_conflict` block
- | | Error Handling & Logging
  - [x] Review `on_error` with new `must`
  - [ ] Rationalize loggers (single interface)
  - [ ] `ReleaseHandle` tool to prevent leaks
- | | Performance & File/FS Tools
  - [ ] Profiling pass if slowdowns observed
  - [ ] Evaluate `FS.DiffFiles` vs `NSPatch`
  - [ ] GoBuild artifact helpers (`FS.Copy`, etc.)

────────────────────────────────────────────────────────────

# 4 · FDM / NS Integration
- | | **Finder Abstraction**
  - [ ] Mediator implements `fdm.find(query:"...")`
  - [ ] Internal routing: AST → Similarity → text scan
- | | **Summary-Node Pipeline**
  - [ ] Auto-create `file_summary` when file > 512 tokens
  - [ ] NS tool `fdm.summarize_chunk`
- | | **Manual Query MVP**  🆕
  - [ ] Script `examples/manual_query.ns` ingests repo, prints top 5 matches
  - [ ] Integration test in `core/testdata`
  - [ ] Bench: ≤ 1.5 s end-to-end on local laptop
- | | **Metrics for BHAG**
  - [ ] Prometheus counter `context_resets_total`
  - [ ] CI smoke test fails if resets > 3 / h

────────────────────────────────────────────────────────────

# 5 · Tooling & Ecosystem
- | | Documentation
  - [x] Update language spec for error map, `must`, events
  - [ ] Tool docs for WM & FDM bridge helpers
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
