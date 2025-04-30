 # NeuroScript Go Development Guidelines (Revised April 24, 2025)
 
 Please follow these guidelines carefully for all Go code contributions to the NeuroScript project.
 
 ## Project Setup & Awareness
 
 1.  **Understand Context:** Before coding, review project docs (`docs/`, if present) and existing code to grasp the goals and architecture. Remember the dependency on the `neuroscript` folder.
 1a. **USE THE INDEX** you should have a file `neuroscript_index.json`, **if not, ask for it**, use it to find things in the codebase **instead of making assumptions**. 
 2.  **Request Missing Files:** If you need files you don't have (e.g., `.g4`, `.y`, fixtures, source code), ask for them immediately. Don't guess contents.
 3.  **Pause for Discussion:** During design discussions, wait for an explicit request before generating new/updated code files -- especially more than 2.
 3. a. If you think the problem is build cache issues or build environment issues **you are wrong**, it has **never** been these things yet. It **may well** be that you have a stale file however, so just ask for the latest.
 
 ## Code Output & Structure
 
 4.  **Full & Functional Files:** Always provide complete Go files. **NEVER** generate code with function bodies "shorted out" or replaced by comments (e.g., `// ... implementation ...`); this wastes significant time if missed. Ensure all provided code is intended to be functional.
 5.  **Single Update Block (Optional):** *Exception:* If modifying a large existing file with localized changes, you *may* provide just the single, contiguous changed block (e.g., a modified function) clearly marked for replacement, *instead* of the whole file. **Never provide more than one update block per file.** If in doubt, provide the full file.
 6.  **File Size:** Split Go files logically if they exceed ~200 lines.
 7.  **Helpers:** Place reusable helpers in shared files (e.g., `utils.go`, `parsing_helpers.go`). Add new helpers cautiously; modify existing ones **rarely**.
 8.  **Package Comments:** Keep `// Package ...` comments accurate.
 9.  **Import Paths:** Use plain string literals for import paths (e.g., `"path/to/pkg"`), not Markdown links.
 
 ## Error Handling Protocol
 
 10. **Core Principle: Check Error Values, NOT Strings.**
     * **CRITICAL:** **NEVER** check error messages directly (e.g., `err.Error() == "some string"`). This is fragile and incorrect.
     * **ALWAYS Use `errors.Is`:** Check if an error matches or wraps a specific *sentinel error variable* using `errors.Is(err, TargetSentinelError)`.
     * **Use `errors.As` for Types:** If you need to check if an error is of a specific *type* (that might hold data), use `errors.As(err, &targetTypeError)`.
     * Use `if err != nil` only for generic "is there an error?" checks when the specific cause doesn't matter at that point.
 
 11. **Define Sentinel Errors:** In each package requiring specific error conditions, declare exported sentinel error variables in a dedicated `errors.go` file using `errors.New("descriptive message")`.
     * Example: `var ErrNotFound = errors.New("resource not found")`
     * Example: `var ErrInvalidInput = errors.New("input is invalid")`
     * Example: `var ErrNoContent = errors.New("input is empty or effectively empty")` (Use this for empty strings, nil/empty slices/maps where processing cannot proceed).
 
 12. **Return Exact Sentinels:** Functions must return the *exact* predefined sentinel error variable when that specific condition occurs.
     * Example: `if !found { return nil, ErrNotFound }`
 
 13. **Wrap for Context:** When adding context to any error (sentinel or otherwise), use `fmt.Errorf` with the `%w` verb to wrap the original error. This preserves the ability to check with `errors.Is`.
     * Example: `return fmt.Errorf("processing user %d: %w", userID, err)`
 
 ## Testing Strategy
 
 14. **Philosophy:** Keep tests simple, clear, focused, and fast to fail. Each test verifies one specific aspect. Use descriptive names.
 15. **Standard `testing` Package:** Use *only* the built-in Go `testing` package (`*testing.T`, `t.Run`, `t.Errorf`, `t.Fatalf`, `t.Helper()`, `t.TempDir()`). **NO external assertion libraries (like `testify`).**
 16. **Structure:** Use table-driven tests (`[]struct{...}`). **Prioritize using helper functions from `pkg/core/universal_test_helpers.go` whenever possible** to reduce repetition and standardize testing. Create other helpers as needed. Test unexported functions directly within their package.
 17. **Inputs:** Prefer external fixture files (`.txt`, `.json`, etc.) over large inline strings in tests.
 18. **Assertions:**
     * Use standard Go comparisons (`==`, `!=`, `nil` checks).
     * Use `reflect.DeepEqual` for complex types (provide detailed output on failure).
     * **Error Verification:** Check for specific errors using `errors.Is(err, ExpectedSentinelError)`. Check for general errors using `if err != nil`.
 19. **Test Flow:** Implement happy path tests first. Then, **once those pass** add unhappy path tests, focusing on verifying the *correct sentinel error* is returned (using `errors.Is`). Only check return values on success (`err == nil`).
 
 ## Debugging
 
 20. **Systematic Approach:** Isolate issues methodically. Avoid random changes. Create small test cases if needed.
 21. **Use Debug Logging:** Add temporary `fmt.Println` or `log.Printf` liberally if stuck.
 22. **KEEP Debug Output:** **NEVER** remove debug logging unless explicitly asked for a clean version.
 23. **Incremental Changes:** Modify code incrementally. Avoid unnecessary changes in unrelated stable code.
 24. **Suggest Bulk Fixes:** If a simple fix is needed across many files, suggest a search/replace pattern instead of providing all modified files.
 
 ## Go Design Principles
 
 25. **Explicit Functions:** Favor clear, single-purpose functions. Avoid "smart" functions that guess intent. If logic differs significantly based on input *interpretation*, create separate functions.
 26. **Export Sensibly:** Default to exporting types/functions unless they are purely internal implementation details or complex/unsafe for external use.
 
 ## Markdown & Specs
 
 27. **Markdown Prefix:** Prepend `@@@` (three "at" signs) to each non-blank line in *standalone* Markdown files (like READMEs) to prevent UI rendering. (Not needed for Go code).
 28. **Relative Links:** Use relative Markdown links for file references.
 29. **Spec Structure:** Adhere to `docs/specification_structure.md`. Ask if a spec needs updating to match the structure.