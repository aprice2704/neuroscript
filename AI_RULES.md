 # NeuroScript Go Development RULES (Revised 1-May-2025) by AJP
 
Follow these **RULES** **very strictly** for all Go code contributions to the NeuroScript project, unless I say otherwise. DO NOT FORGET THEM. We use golang version 1.24 or later.
 
 ## Project Setup & Awareness
 
 1.  **Understand Context:** Before coding, review project docs (`docs/`, if present) and existing code to grasp the goals and architecture. Remember the dependency on the `neuroscript` folder. ALWAYS read any .md files.
 1a. **USE THE INDEX** you should have a file `neuroscript_index.json`, **if not, ask for it**, use it to find things in the codebase **instead of making assumptions**. 
 2.  **Request Missing Files:** If you need files you don't have (e.g., `.g4`, `.y`, fixtures, source code), ask for them immediately. Don't guess contents.
 3.  **Pause for Discussion:** During design discussions, wait for an explicit request before generating new/updated code files -- especially more than 2.
 3. a. Most of the time, generate one file at a time UNLESS splitting a file
 3. b. If we experience stale files, **ask** for the latest contents to update.
 3. c. If you think the problem is build cache issues or build environment issues **you are wrong**, it has **never** been these things yet. It **may well** be that you have a stale file however, so just ask for the latest.
 
 ## Code Output & Structure
 
 4.  **Full & Functional Files:** Always provide complete Go files. **NEVER** **NEVER** **NEVER** **NEVER** **NEVER** generate code with function bodies "shorted out" or replaced by comments (e.g., `// ... implementation ...`); this wastes significant time if missed. Ensure all provided code is intended to be functional.
 5.  **Single Update Block (Optional):** *Exception:* If modifying a large existing file with localized changes, you *may* provide just the single, simple, contiguous changed block (e.g., a modified function) clearly marked for replacement, *instead* of the whole file. **Never provide more than one update block per file.** If in doubt, provide the full file.
 5a. DO NOT PROVIDE MORE THAN ONE FILE PER TURN UNLESS WE AGREE ON THIS IN ADVANCE. If you give me several files that are large and head in the wrong direction it wastes a lot of time and patience.
 6.  **File Size:** YOU MUST Split Go files logically IMMEDIATELY if they exceed 200 lines. You don't have to ask first, just do it. **THIS IS VITAL** for development efficiency (moving text between you and I).
 7.  **Helpers:** Place reusable helpers in shared files (e.g., `utils.go`, `parsing_helpers.go`). Add new helpers cautiously; modify existing ones **rarely**.
 8.  **Package Comments:** Keep `// Package ...` comments accurate.
 9.  **Import Paths:** Use plain string literals for import paths (e.g., `"path/to/pkg"`), not Markdown links.
 
 9. a. **Versioning** Please place a **version stamp** at the top of each file you modify, like this:
 
// NeuroScript Version: 0.3.0
// File version: 0.1.3 <<-- bump this right hand most number
// Simplify Inspect check, keep AST dump ON
// filename: pkg/core/tools_go_semantic.go
comment lines at the top of files thus:
// nlines: 295 (number of LoC in the file)
// risk_rating: HIGH (or medium or low depending on how central the contents of the file are to other parts of the system -- interpreter.go is high, as are many helper files for instance)

and bump the minor number in the file version each time. Please remove the last modified date/time -- it proved not to be useful.
 
 
 9. b. **Comments**: add and maintain package comments please. No comments within import blocks please. Remember: **do not** put // comments in .ns files, they are no permitted.
 
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
 26a. **BAIL OUT ON NIL** always aggressively test data structures for nil components and BAIL out of the program immediately if they are found. **DO NOT ATTEMPT TO LIMP ALONG**.
 
 ## Markdown & Specs
 
 27. **Markdown Prefix:** Prepend `@@@` (three "at" signs) to each non-blank line in *standalone* Markdown files (like READMEs) to prevent UI rendering. (Not needed for Go code).
 28. **Relative Links:** Use relative Markdown links for file references.
 29. **Spec Structure:** Adhere to `docs/specification_structure.md`. Ask if a spec needs updating to match the structure.

 
 ### Rule 30: Interpreter Setup for Testing

 Setting up a `core.Interpreter` instance correctly is crucial for testing tools and interpreter logic. Use the following methods:

 **A. Testing *Within* the `pkg/core` Package:**

 1.  **Use Dedicated Helpers:** The primary way to get a test interpreter is by using the helper functions defined in `pkg/core/helpers.go`:
     * `NewDefaultTestInterpreter(t *testing.T) (*core.Interpreter, string)`: **This is the preferred helper.** It creates an interpreter with standard test defaults.
     * `NewTestInterpreter(t *testing.T, vars map[string]interface{}, lastResult interface{}) (*core.Interpreter, string)`: Use this if you need to initialize the interpreter with specific variables or a `lastResult` value.

 2.  **Helper Functionality:** These helpers automatically handle:
     * Creating a new `core.Interpreter`.
     * Setting up a test logger that directs output to `t.Log`.
     * Setting up a `NoOpLLMClient` (no actual API calls).
     * Creating a unique temporary sandbox directory using `t.TempDir()` for test isolation.
     * Setting the interpreter's sandbox root to this temporary directory using `SetSandboxDir()`.
     * **Registering all core tools** (filesystem, Go tools, etc.) via `RegisterCoreTools()`.
     * Returning the initialized `*core.Interpreter` and the `string` path to the created temporary sandbox directory.

 3.  **Example (`*_test.go` within `pkg/core`):**
     ```go
     import (
     	"testing"
     	// ... other imports
     )

     func TestMyCoreFunction(t *testing.T) {
     	interpreter, sandboxDir := NewDefaultTestInterpreter(t) // Use the helper
     	// Set specific sandbox contents if needed
     	// ... write files to sandboxDir ...

     	// Run test logic using the 'interpreter'
     	// ...
     }
     ```

 **B. Testing *Outside* the `pkg/core` Package (e.g., in `pkg/neurogo`, `pkg/toolsets`):**

 1.  **Manual Setup Required:** The internal helper functions (`NewTestInterpreter`, `NewDefaultTestInterpreter`) are *not* exported and cannot be called directly from other packages. You must manually construct and configure the interpreter.

 2.  **Steps:**
     * Import `github.com/aprice2704/neuroscript/pkg/core`.
     * Import necessary adapter packages (e.g., `github.com/aprice2704/neuroscript/pkg/adapters` for logger/LLM).
     * Create a logger instance (usually `adapters.NewNoOpLogger()` or `adapters.NewSLogAdapter(logging.LogLevelDebug)` for test output).
     * Create an LLM client instance (usually `adapters.NewNoOpLLMClient()`).
     * Create a temporary sandbox directory using `t.TempDir()`.
     * Call the exported constructor: `interpreter, err := core.NewInterpreter(logger, llmClient, sandboxDir, nil)` (passing the temp dir path as the initial sandbox root). Handle potential errors.
     * **Register Tools:** Explicitly call the necessary *exported* tool registration functions:
         * `core.RegisterCoreTools(interpreter)` (Essential for basic functionality like filesystem access).
         * Potentially `toolsets.RegisterExtendedTools(interpreter)` (or other specific registration functions if testing tools outside the core set). Handle potential errors.
     * *(Optional but Recommended)*: Call `interpreter.SetSandboxDir(sandboxDir)` again after registration, just to be certain the `FileAPI` is fully initialized with the correct path (though `NewInterpreter` should handle the initial setting).

 3.  **Example (`*_test.go` outside `pkg/core`):**
     ```go
     import (
     	"testing"
     	"github.com/aprice2704/neuroscript/pkg/core"
     	"github.com/aprice2704/neuroscript/pkg/adapters"
        "github.com/aprice2704/neuroscript/pkg/logging"
     )

     func TestMyIntegration(t *testing.T) {
         logger := adapters.NewSLogAdapter(logging.LogLevelDebug) // Or NewNoOpLogger()
         llmClient := adapters.NewNoOpLLMClient()
         sandboxDir := t.TempDir()

         interpreter, err := core.NewInterpreter(logger, llmClient, sandboxDir, nil)
         if err != nil {
             t.Fatalf("Failed to create core.Interpreter: %v", err)
         }

         // Must register tools manually!
         err = core.RegisterCoreTools(interpreter)
         if err != nil {
             t.Fatalf("Failed to register core tools: %v", err)
         }
         // err = toolsets.RegisterExtendedTools(interpreter) // If needed
         // if err != nil {
         // 	 t.Fatalf("Failed to register extended tools: %v", err)
         // }

         // Ensure sandbox is set (redundant if NewInterpreter guarantees it, but safe)
         err = interpreter.SetSandboxDir(sandboxDir)
         if err != nil {
              t.Fatalf("Failed to set sandbox dir: %v", err)
         }

     	// Set specific sandbox contents if needed
     	// ... write files to sandboxDir ...

         // Run test logic using the 'interpreter'
         // ...
     }
     ```

 **Key Considerations:**

 * **Tool Registration:** Tools are *not* automatically registered when using `core.NewInterpreter` directly. You *must* call the relevant `Register...` functions.
 * **Sandbox:** Always use `t.TempDir()` to create isolated sandboxes for tests. Ensure the interpreter's sandbox is correctly set using `SetSandboxDir()`.
 * **Dependencies:** Use `NoOpLLMClient` for testing.
 Use adapters.NewSimpleSlogAdapter(output io.Writer, level logging.LogLevel) for most regular logging.
