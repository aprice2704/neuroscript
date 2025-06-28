:: title: NeuroScript Project (includes FDM integration) Road to v0.4.0 Checklist 
:: fileVersion: 7
:: id: ns-project-roadmap-v0.4.0
:: status: draft
:: description: Tasks for NS v0.4.0 incl. event grammar, prompt budgeting, FDM bridge, and more.
:: updated: 2025-06-27


# Vision
> **BHAG** — Multiple LLM agents (ChatGPT, Gemini, etc.) maintain the *ns* + *fdm* repos without context-window breakage.  
> Success criteria: ≤ 3 context resets / h, refactor round-trip < 5 min, divergence < 0.25 for ≥ 95 % nodes.
> Current short-term goal: **AI First Light** (see section 4)

────────────────────────────────────────────────────────────

# 1 · Language Features
- |-| **Grammar & Runtime Enhancements**
  - [-] **Finalize `on event` grammar + dispatcher**
  - [x] `error` value type (+ `IsError` helper)
  - [x] `timedate` value type
  - [x] `event` value type
  - [x] Fuzzy-logic operators
- |-| **`must` Keyword & Error-map Standardization**
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

# 3 · Core Interpreter / Runtime
- | | **Concurrency & Conflict Handling**
  - [ ] **Add `version` field to nodes**
  - [ ] Mediator rejects stale writes (`ErrConflict`)
  - [ ] Optional `on_conflict` block
- | | Error Handling & Logging
  - [x] Review `on error` with new `must`
  - [ ] Rationalize loggers (single interface)
  - [ ] `ReleaseHandle` tool to prevent leaks
- | | Performance & File/FS Tools
  - [ ] Profiling pass if slowdowns observed
  - [ ] Evaluate `FS.DiffFiles` vs `NSPatch`
  - [ ] GoBuild artifact helpers (`FS.Copy`, etc.)

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
- | | **AI First Light**
  - [ ] Ingest two repos
  - [ ] Process both source and docn files into all usual indices and overlays
  - [ ] Make AI agent
  - [ ] Send instructions for query
  - [ ] Agent sends query
  - [ ] Agent gets relevant results back
- | | **Metrics for BHAG**
  - [ ] Prometheus counter `context_resets_total`
  - [ ] CI smoke test fails if resets > 3 / h

# 5 · Tooling & Ecosystem
- | | Documentation
  - [x] Update language spec for error map, `must`, events
  - [ ] Tool docs for WM & FDM bridge helpers
- | | Formatter
  - [ ] **`nsfmt`** — wraps long lines, respects `\` continuation
- | | Workflow Tests
  - [ ] Prompt-budget CI test: 2k-line file ⇒ summary ≤ 1024 tokens

# 7 · NS LSP
- |x| Extend grammar support: `must` enhancements, events, fuzzy

# 8 . Tech Debt

- [ ] Ensure CLI tools DO NOT auto-run main. Zadeh s/b done already.
- [ ] Clean up neurogo err return to zadeh main