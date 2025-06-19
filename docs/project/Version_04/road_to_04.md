:: title: NeuroScript Road to v0.4.0 Checklist
:: version: 004
:: file_version: 6
:: id: ns-roadmap-v0.4.0
:: status: draft
:: description: Tasks for NS v0.4.0 incl. Work-Queue, event grammar, prompt budgeting, FDM bridge, and more.
:: updated: 2025-06-16
:: review_comment: Gemini assessment based on spec documents provided 2025-06-16.

# Vision
> **BHAG** — Multiple LLM agents (ChatGPT, Gemini, etc.) maintain the *ns* + *fdm* repos without context-window breakage.  
> Success criteria: ≤ 3 context resets / h, refactor round-trip < 5 min, divergence < 0.25 for ≥ 95 % nodes.

────────────────────────────────────────────────────────────

# 1 · Language Features
- | | **Grammar & Runtime Enhancements**
  - [ ] **Finalize `on event` grammar + dispatcher** // Gemini Review: The `event` value type is specified, but the `on event` grammar and dispatcher logic are not yet detailed.
  - [x] `error` value type (+ `IsError` helper) // Gemini Review: Specified in `must_enhancements.md` and `new_types.md`.
  - [x] `timedate` value type // Gemini Review: Specified in `new_types.md`.
  - [x] `event` value type // Gemini Review: The value type itself is specified in `new_types.md`.
  - [x] Fuzzy-logic operators // Gemini Review: Detailed semantics are specified in `ns_script_spec.md` and `new_types.md`.
- | | **`must` Keyword & Error-map Standardization**
  - [x] Define canonical `error` map: `{code, message, details}` // Gemini Review: Explicitly defined in `must_enhancements.md`.
  - [x] `tool` guidelines: return error map on handled faults // Gemini Review: Core concept in `must_enhancements.md`, superseding `tool_conventions.md`.
  - [x] `set x = must tool()`  — fail on Go error **or** error map // Gemini Review: Core feature specified in `must_enhancements.md`.
  - [x] Single-key `must map["key"] as type` // Gemini Review: Specified in `must_enhancements.md`.
  - [x] Multi-key atomic `must map[...] as …` // Gemini Review: Specified in `must_enhancements.md`.
  - [ ] `must IsCorrectShape(var, shape_def)` // Gemini Review: This advanced validation function is not described in the provided specs.
  - [x] Ensure `on_error` catches all `must` panics // Gemini Review: This is the intended behavior per `must_enhancements.md` and `ns_script_spec.md`.
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
  - [ ] **JobToken schema + signer** # NEW
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
  - [ ] **Add `version` field to nodes** # NEW
  - [ ] Mediator rejects stale writes (`ErrConflict`)  # NEW
  - [ ] Optional `on_conflict` block  # NEW
- | | Error Handling & Logging
  - [x] Review `on_error` with new `must` // Gemini Review: Implicitly addressed by the design where `must` failures trigger runtime errors caught by `on_error`.
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
  - [x] Update language spec for error map, `must`, events // Gemini Review: Specs exist but need to be consolidated into the main language spec and formal grammar.
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