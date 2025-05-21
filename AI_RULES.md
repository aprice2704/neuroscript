# NeuroScript Go Development: Core AI Rules (Revised May 2025)

Follow these **CORE RULES** **very strictly** for all Go code contributions to the NeuroScript project, unless I (AJP) say otherwise. Your consistent adherence to these rules is crucial for project efficiency and code quality. We use Go version 1.24 or later.

## TOP 10 CRITICAL RULES - ALWAYS FOLLOW:

1.  **Understand Context First:** Before coding, review project documentation (especially `.md` files) and existing code to grasp goals and architecture. Remember the `neuroscript` folder dependency.
2.  **Use the Code Index:**
    * You should have a `neuroscript_index.json` file. **If not, ask for it.**
    * Use this index to find things in the codebase. **DO NOT MAKE ASSUMPTIONS.**
3.  **Full & Functional Files ALWAYS:**
    * Always provide complete Go files based on latest versions (we have had some issues with stale files)
    * **NEVER EVER** generate code with function bodies "shorted out" or replaced by comments (e.g., `// ... implementation ...`). This wastes significant time and has lead to us wasting several hours of effort on at least two occasions.
4.  **One Main File Per Turn:**
    * Provide only **one primary code file** per response, unless we explicitly agree otherwise (e.g., when splitting a file).
    * Delivering multiple, potentially misdirected large files is very inefficient.
5.  **Split Large Files Immediately:**
    * If a Go file exceeds **200 lines of code**, you **MUST** split it logically.
    * Do this automatically; you don't need to ask first. **THIS IS VITAL.**
6.  **Apply Versioning Stamps:**
    * At the top of **every Go file you modify**, include the following, bumping the `File version` minor number:
        ```go
        // NeuroScript Version: 0.3.0 // (Or current project version)
        // File version: 0.1.X // (X is bumped)
        // Purpose: Brief description of file's role or recent change
        // filename: path/to/your/file.go
        // nlines: YYY (actual number of lines in this file)
        // risk_rating: LOW/MEDIUM/HIGH (assess based on centrality)
        ```
7.  **Correct Error Handling (CRITICAL):**
    * **NEVER** check error messages directly (e.g., `err.Error() == "some string"`). This is fragile.
    * **ALWAYS use `errors.Is(err, TargetSentinelError)`** to check for specific sentinel errors.
    * **ALWAYS use `errors.As(err, &targetTypeError)`** to check for specific error types (that might hold data).
    * Return exact sentinel errors (e.g., `return ErrNotFound`) or wrap errors with context using `fmt.Errorf("... %w", ..., err)`.
8.  **Standard Go Testing ONLY:**
    * Use **ONLY** the built-in Go `testing` package (e.g., `*testing.T`, `t.Run`, `t.Errorf`, `t.Helper()`).
    * **NO external assertion libraries** (like `testify`).
9.  **Bail Out On Nil:**
    * Always aggressively test data structures for `nil` components.
    * If a `nil` component is found that prevents safe continuation, **BAIL OUT of the program or function immediately** (e.g., return an error, or `panic` if appropriate for an unrecoverable state in a command-line tool context, though errors are preferred for library code). **DO NOT ATTEMPT TO LIMP ALONG.**
    * **Avoid** using nil detection on returned values -- return an error instead of a nil.
10. **Request Missing Information:**
    * If you need files (e.g., `.g4`, fixtures, other source code) or clarification that you don't have, **ask for them immediately.** Do not guess contents or specifications.

---

## OTHER KEY GUIDELINES

While the Top 10 are paramount, also remember these important points:

* **Stale Files:** If you suspect build/cache issues, it's almost certainly a stale file. Ask for the latest version.
* **Pause for Discussion:** During design discussions, wait for an explicit request before generating code, especially multiple files.
* **Single Update Block (Optional):** For minor changes in large files, you *may* provide just the changed block, clearly marked. If in doubt, send the whole file. NEVER more than one update block per file.
* **Helpers:** Place reusable helpers in shared files (e.g., `utils.go`).
* **Import Paths:** Use plain string literals for Go import paths.
* **Package Comments:** Keep `// Package ...` comments accurate. No comments within import blocks. `.ns` files do not permit `//` comments.
* **Sentinel Errors:** Define exported sentinel errors in `errors.go` files (e.g., `var ErrNotFound = errors.New(...)`). Return these exact sentinels or wrap them with `fmt.Errorf("... %w", ..., err)`.
* **Testing Strategy:**
    * Keep tests simple, clear, table-driven. Use `pkg/core/universal_test_helpers.go`.
    * Prefer external fixtures.
    * Verify errors with `errors.Is`. Test happy path first, then unhappy paths focusing on correct sentinel errors.
* **Debugging:** Be systematic. Use and **KEEP** debug logging (`fmt.Println`, `log.Printf`) unless asked to remove it. Suggest bulk fixes with search/replace patterns if applicable.
* **Go Design:** Favor explicit, single-purpose functions. Export sensibly.
* **Markdown & Specs:** Prepend `@@@` to standalone Markdown lines. Use relative links. Follow `docs/specification_structure.md`.
* **Interpreter Setup for Testing:** This is a complex but critical rule for writing tests involving the `core.Interpreter`. Refer to the full details in the original (full version) `AI_RULES.md` when setting up interpreters for tests, especially differentiating between tests *within* `pkg/core` (use internal helpers) and *outside* `pkg/core` (manual setup and tool registration required).

---

This revised guide emphasizes the most critical rules to ensure smoother development. For full details, especially on complex topics like Interpreter Setup or specific examples, the original, more comprehensive `AI_RULES.md` (version: Revised 1-May-2025) remains a valuable reference if needed.