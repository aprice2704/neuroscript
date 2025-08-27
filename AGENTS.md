# NeuroScript Go Development: Agent & AI Rules
**Revision:** 2025-Aug-24  
**Audience:** All LLMs and human agents contributing to the NeuroScript project.  

This document codifies the **CORE RULES** for AI-assisted Go development in NeuroScript.  
These rules MUST be followed consistently unless explicitly overruled by AJP.  
We use **Go 1.24+**. These rules emphasize correctness, minimalism, and context awareness.

---
## 🔴 TOP CRITICAL RULES — ALWAYS OBEY

### 1. Understand Context First
- Always review `.md` docs and relevant Go code **before** making changes.  
- Fix compiler/test failures with **minimal, targeted edits**. Do not “tidy up” unrelated code.  

### 1b. DEBUG OUTPUT
- If you don't fix all related bugs in a file in **two** attempts, you **do NOT** understand the problem, however confident you feel. **IMMEDIATELY** add debug output. Given enough debug output, all bugs are shallow.

### 2. Import Hygiene
- Never output `github.comcom`. It must always be `github.com`.  
- In `.md` files, do **not** wrap Go import paths in markdown links.

### 3. Full & Functional Files
- Always deliver **complete files**.  
- **NEVER** leave bodies stubbed with `// ... implementation ...`. That has cost hours of wasted debugging in the past.  

### 4. No File Carpet-Bombing
- Send only the **files that changed**.  
- Prefer one or two files at a time.  
- If multiple files are needed, explain why and stage them clearly.  

### 5. Split Large Files
- If a Go file grows beyond ~200 LOC, split it logically into multiple files.  
- Do this automatically when generating new code.  

### 6. File Headers & Versioning
At the top of **every modified file**, include:
```go
// NeuroScript Version: 0.3.0   // or current project version
// File version: X              // bump integer (0.1.7 → 8)
// Purpose: one-line description
// filename: path/to/file.go
// nlines: YYY   // actual line count
// risk_rating: LOW|MEDIUM|HIGH
```

### 7. Error Handling
- Never string-compare error messages.  
- Use `errors.Is` for sentinel errors.  
- Use `errors.As` for typed errors.  
- Return sentinel errors (`ErrNotFound`) or wrap with `fmt.Errorf("... %w", err)`.  

### 8. Testing
- Use the **standard library `testing` package only**.  
- No `testify` or external assertion libs.  

### 9. Bail Out on Nil
- Always check for nil before use.  
- If nil means unsafe state, **return an error immediately** (or panic if unrecoverable).  
- Don’t “limp along” on nils.  

### 10. Ask for Missing Info
- If you need a `.g4` grammar, a fixture, or a file that isn’t present, **ask immediately**.  
- Don’t hallucinate file contents.  

### 11. No Stubbing
- Avoid stubbing code. The Go compiler is your ally — it prevents half-programs.  

### 12. Strong Typing
- Do not use raw `string` or `int` for semantic fields.  
- Always create domain types (e.g., `type Status string`) and define valid values.  

---
## 📑 MARKDOWN & SPECS — SPECIAL RULE

When emitting Markdown or spec files (e.g. `agents.md`), you **MUST**:
- Wrap them in a plain text inline code block.  
- Prepend `@@@` to every line.  

This armors them against web ui rendering vagaries

**Example:**
```txt
@@@# Example Spec
@@@This is how to emit an `.md` file.
```

---
## OTHER KEY GUIDELINES

- **Stale Files:** If things break mysteriously, assume stale files. Request fresh versions.  
- **Pause for Discussion:** After questions or design discussions, wait for an explicit request before generating code.  
- **Single Update Block (optional):** For very large files, you may provide one fenced “update block” instead of the full file — but never more than one per file.  
- **Helpers:** Place general helpers in shared files (e.g. `utils.go`).  
- **Sentinel Errors:** Export them from dedicated `errors.go` files.  
- **Tests:** Use table-driven style. Verify with `errors.Is`. Cover happy paths first.  
- **Debug Logging:** Keep debug logs (`fmt.Println`, `log.Printf`) unless explicitly asked to remove.  
- **Design:** Prefer explicit, single-purpose functions. Export carefully.  

---
## SUMMARY

These rules are the **operating contract** between humans and LLM agents in NeuroScript:  
- Context first, minimal diffs.  
- Full functional code, no stubs.  
- Strong typing, strict error handling.  
- Controlled file scope and size.  
- Always emit `.md` specs as ``-prefixed inline blocks.  

Breaking these rules causes wasted time and fragile code. Following them keeps the project fast,
safe, and maintainable.