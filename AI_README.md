# Guidelines and Instructions for the NeuroScript project work product files

Please read carefully and follow these rules carefully. Keep them fresh in your mind.
If you have not already done so, read all the files in doc/... before starting any work.

## Markdown and NeuroData spec files

7.  Markdown Formatting: When providing only Markdown files (e.g., README.md), prepend each non-blank line with `@@@`. This prevents the UI from rendering it. I will remove the prefix later. (This is not needed for Go files or other code formats). 
8.  * Example Markdown Line: @@@# Project Title
9.  Please convert any references you see in markdown files into relative Markdown links -- if multiple files are cited, please take care to convert them into multiple links
10. When working on specifications, please be sure to use the structure given in docs/specification_structure.md. If a spec does not adhere to this structure, ask about updating it.

## Go Development Instructions

Project Context & Awareness:

1.  Understand the Goal: Before coding, review the existing codebase. Pay close attention to documentation files (like .md files in docs/) to grasp the project's intent and architecture. Remember we have prior history on this project.
2.  Neuroscript Dependency: This project requires the neuroscript folder. Please confirm if you have access to it or notify me if it's missing. 
3.  Missing Files: If you need access to a file you cannot reach (e.g., .g4, .y, fixtures, specific source files), ask me to provide the text immediately. Do not guess its contents. 
4.  Hold Code During Discussion: If we are discussing a potential change, design idea, or approach, wait for me to explicitly ask for the updated/new code files. Do not proactively generate them during the discussion phase. 

Output Format & Structure:

5.  Provide Full Files: Always output complete Go files. Do not provide code fragments or diffs unless specifically asked. 
6.  Split Large Files: If any single Go code file exceeds roughly 300 lines, split it logically into smaller files. You do not need to ask for permission; just do it. Ensure the split maintains coherence (e.g., related functions stay together or are moved to appropriate helper files). 
10. Helper Functions: place reusable helper functions into appropriately named shared files (e.g., `utils.go`, `helpers.go`, or more specific like `parsing_helpers.go`). Ensure they are properly namespaced within the package. Existing helpers should be changed with **extreme caution**. Prefer adding new ones until proven.  
11. Package Comments: Keep package-level comments (`// Package mypackage description...`) accurate and up-to-date. If your changes significantly alter a package's purpose, check with me before proceeding with drastic modifications. 
12. Go Import Path Formatting: Within Go code blocks, especially import (...) blocks, ensure package import paths are plain string literals only (e.g., "github.com/org/repo/pkg"). Do not automatically convert them into Markdown links like path. 

Error Handling Protocol:

11. Define Sentinel Errors: Use a dedicated `errors.go` (or similar) file within each relevant package to declare exported sentinel error variables for specific, known error conditions. 
    * Use `errors.New("descriptive error message")`.
    * Example: `var ErrUserNotFound = errors.New("user not found")`
12. Handle Empty Input: Explicitly check for empty or effectively empty input (e.g., whitespace-only strings, empty slices) where relevant. Define and return a specific sentinel error, such as `ErrNoContent = errors.New("input content is empty or effectively empty")`, for this condition. 
13. Return Specific Errors: Functions should return the exact predefined sentinel error variable when that specific condition occurs. 
    * Example: `if !found { return nil, ErrUserNotFound }`
14. Wrap Errors for Context: When adding context (e.g., function name, specific IDs) to an error (sentinel or otherwise), use `fmt.Errorf` with the `%w` verb to wrap the original error. Do not just return a new string error. 
    * Example: `return fmt.Errorf("failed processing order %d: %w", orderID, err)`
15. Checking Errors in Tests: 
    * Use `errors.Is(err, TargetError)` to check if a returned error `err` is or wraps a specific `TargetError` (like `ErrUserNotFound` or `ErrNoContent`). This is the preferred method for checking specific error types. 
    * Use `if err != nil` or `if err == nil` for general checks of error presence/absence when the specific type doesn't matter. 
    * DO NOT rely on matching exact error message strings (`err.Error() == "some string"`) for testing error types. 

Testing Strategy:

1.  Happy Path First: Implement and verify tests for the expected, non-error ("happy path") scenarios before writing tests for error conditions ("unhappy paths"). 
2.  Unhappy Path Tests: Do implement unhappy path tests, but do not expect specific return values (only specific errors) in these cases unless the function contract guarantees them even on error. Focus on verifying the correct error type is returned (using `errors.Is`). 
3.  Table-Driven Tests & Helpers: For testing functions with multiple scenarios (especially tools or parsing logic), use table-driven tests (slices of test case structs) and test helper functions to reduce repetition and improve test clarity and maintainability.
4.  Use Fixtures: Strongly prefer using external fixture files (e.g., `.json`, `.txt`, `.xml` files loaded during tests) for test inputs, especially for parsing or processing structured data. Avoid large, embedded multi-line strings in test code, as fixtures are easier to read, manage, and avoid escaping issues. 
5.  **Testing Philosophy: Keep Tests Simple, Clear, and Fast to Fail**
When writing tests, prioritize clarity and simplicity above complex conditional logic within a single test case. Each test should ideally verify one specific aspect or outcome of the code under test. Use clear, descriptive names for test cases. Assertions should be direct and check only what is necessary; for instance, when checking errors, verify the error type using errors.Is rather than matching exact error message strings, which are prone to change. Similarly, only validate successful results when no error is expected. Avoid complex boolean expressions or deeply nested logic within test assertions; if different conditions need testing, prefer separate, focused test cases. This approach makes tests easier to read, understand, and maintain for both human developers and AI agents, ensuring failures point quickly and accurately to the specific condition that broke.

More rules to remember when writing tests in this project:

Standard Library: Use the built-in Go testing package exclusively (e.g., *testing.T, t.Run, t.Errorf, t.Fatalf). Avoid external assertion libraries like testify.  

Table-Driven Tests: Structure tests using tables (testCases := []struct{...}) for clarity and easy addition of new scenarios.  

Direct Function Calls: Test unexported functions (like tool implementations, e.g., toolStat) directly within the core package (not core_test).  

Helper Functions: Utilize helper functions for common setup (e.g., creating temporary directories/files, setting up minimal interpreter instances ) and teardown (e.g., using t.TempDir()  or returned cleanup functions).

Standard Assertions: Perform assertions using standard Go comparisons (==, !=, nil checks) combined with t.Errorf or t.Fatalf. Use errors.Is for checking specific error types.

Deep Comparison: Use reflect.DeepEqual for comparing complex types like maps and slices. If comparison fails, provide more detailed output by iterating through expected vs. actual elements/keys.

Error Handling: Check returned Go errors directly. For tools returning errors within their result structure (like a map), test that specific field.

Clarity: Prioritize clear test case names and assertion messages. Use t.Fatalf when a failure prevents meaningful continuation of a test case.



6.  **Use Standard Testing Package:** Do NOT use external assertion libraries like `github.com/stretchr/testify/assert` or `github.com/stretchr/testify/require`. Use the functions provided by the standard Go `testing` package (e.g., `t.Errorf`, `t.Fatalf`, `t.Helper()`) and standard Go comparison operators (`==`, `!=`) along with `errors.Is` for error checking.

Debugging & Problem Solving:

20. Systematic Debugging: If you encounter issues, avoid randomly trying multiple fixes at once ("churn"). Instead, work systematically to reduce unknown factors. Consider creating small, isolated test programs to verify assumptions or isolate problematic behavior, especially when multiple factors might be involved. 
21. Add Debug Output: If a problem isn't solved within 2-3 attempts, DO NOT HESITATE to add temporary debug logging (`fmt.Println`, `log.Printf`, etc.) to trace execution flow and variable states. You have previously been too reluctant to do this. 
22. Keep Debug Output: NEVER, EVER, EVER remove debug output lines unless I explicitly ask for a cleaned-up version of the file. 
23. Incremental Changes: Make changes incrementally. Avoid modifying stable code in unrelated packages unnecessarily. When tackling complex tasks, start with simpler cases and build up. 
24. Suggest Repetitive Fixes: If a fix requires the same simple change across multiple files (e.g., renaming a variable, updating an import), point this out and suggest a search-and-replace approach rather than providing each modified file individually. You can provide the search/replace pattern if appropriate.

Go Code Design Principles:

25. Favor Simple, Explicit Tools: When designing functions or methods, prefer clear, single-purpose implementations over "smart" functions that try to guess intent based on input format. If significantly different logic is needed based on input interpretation (not just type), split the functionality into separate, explicitly named functions. (Reference: `LineCountFile` vs. `LineCountString` example). 
26. Default to Exporting: In Go, it's generally acceptable to export functions and types by default unless they are strictly internal implementation details or have complex, non-obvious usage outside their home package.