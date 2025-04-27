 # NeuroScript Version 0.2.0 Design Document
 
 ## Introduction
 
 NeuroScript aims to be a structured, human-readable language facilitating communication and task execution between **humans, AI agents, and computer systems**. It provides a *procedural scaffolding* for defining reusable skills, particularly enabling Large Language Models (LLMs) to leverage explicit, step-by-step reasoning rather than relying solely on implicit chain-of-thought[cite: 2]. This version (0.2.0) introduces significant refinements focused on improving readability, explicitness, and reinforcing the core communication goals based on developer feedback and practical usage patterns observed in early examples.
 
 ## High-Level Design & Execution Model
 
 ### Core Goals (Reinforced in v0.2.0)
 
 1.  **Explicit Reasoning:** Encourage clear, step-by-step logic in a code-like format[cite: 2].
 2.  **Reusable Skills:** Define procedures (`func`) that can be stored, discovered, and reused[cite: 2].
 3.  **Self-Documenting:** Emphasize structured metadata (`::`) for clarity on purpose, interface, logic, and versioning[cite: 2].
 4.  **Explicit Actor Interaction:** Clearly distinguish interactions targeting AI, Humans, or Computer systems within the script syntax.
 5.  **Structured Data:** Continue support for list (`[]`) and map (`{}`) literals[cite: 2].
 6.  **Versioning:** Maintain `file_version` at the file level and `lang_version` within procedure metadata for tracking[cite: 2].
 7.  **Simplicity (One Way):** Where practical, strive for only one clear way to accomplish a specific language task, similar to Go's philosophy [personalization: reflecting your Go background and preference].
 
 ### Execution Model
 
 NeuroScript execution revolves around the **`neurogo` (`ng`) interpreter**, implemented in Go [personalization: referencing your implementation language]. This interpreter is the ground truth for script execution.
 
 1.  **Primary Executor (`ng`):** The Go interpreter parses and executes scripts step-by-step, managing state and reliably interacting with external systems via defined interfaces (`ask...`, `call tool`).
 2.  **AI as Executor (Conceptual):** An AI can conceptually execute *simple* NeuroScript procedural logic (loops, conditions, assignments) internally, using the script as **procedural guidance** without needing `ng` for every step[cite: 2]. This fulfills the goal of teaching AIs structured procedures.
 3.  **AI as Controller:** For complex scripts or those requiring external tool interaction (`askComputer`, `call tool`), the AI acts as a *controller*, making specific requests to the `ng` interpreter to execute a defined NeuroScript function (skill) with given parameters. This is akin to a function call or using an agent tool, not arbitrary code execution.
 4.  **`ng` Interactions:** The `ng` interpreter initiates outbound requests via `askAI`, `askHuman`, and `askComputer`.
 
 This layered model allows NeuroScript to serve as both direct procedural guidance for AI reasoning and as a robust language for defining executable skills managed by the reliable `ng` interpreter.
 
 
 
 ## Language Syntax & Semantics (Version 0.2.0)
 
 This version introduces several changes to improve readability and explicitness, based on the `script_spec.md` [cite: 2] and our recent design discussions.
 
 ### General
 
 * **Keywords:** All language keywords are now **lowercase** (e.g., `func`, `set`, `if`, `end if`). This improves readability and reduces the "shouting" feel of uppercase keywords used previously.
 * **Comments:** Ignored, human-readable comments start with `#` and continue to the end of the line.
 * **Metadata:** Structured metadata uses the `::` prefix (see Metadata section below).
 * **Strictness:** The `ng` interpreter will enforce structural rules rigorously (e.g., matching block keywords, required parameters, declared returns), similar in spirit to the Go compiler's strictness.
 * **Typing:** NeuroScript remains **dynamically typed**. While structure is enforced, explicit static type declarations for variables are not required, maintaining scripting flexibility.
 
 ### Procedure Definition
 
 * Replaces `DEFINE PROCEDURE ... END`.
 * **Syntax:**
     ```neuroscript
     func <name> [needs <param1>, <param2>...] [optional <param3>...] [returns <var1>, <var2>...] means
         :: <metadata key>: <metadata value>  # Metadata follows immediately
         :: <...>
 
         # Function body statements...
         # Use 'return' statement to exit and provide values
     endfunc
     ```
 * **Clauses:**
     * `func <name>`: Begins the definition.
     * `needs`: Optional clause listing comma-separated required parameter names.
     * `optional`: Optional clause listing comma-separated optional parameter names. The script logic must handle cases where these are not provided (e.g., check `no param` or `param == ""`).
     * `returns`: Optional clause listing comma-separated names for return variables. **Multiple return values are supported.** If present, `return` statements in the body *must* provide values for these (checked by the interpreter). These names are implicitly declared within the function scope (similar to Go named returns).
     * `means`: Keyword indicating the start of the function body (following any metadata).
     * `endfunc`: Keyword terminating the function definition.
 * **Minimal:** `func <name> means ... endfunc` is valid.
 
 ### Metadata (Replaces `comment:` block)
 
 * **Syntax:** Uses the `::` prefix followed by a lowercase keyword, a colon, and the value. Applied immediately after the `func` signature or inline within the code body.
     ```neuroscript
     func MyFunc needs input means
         :: purpose: Describes what MyFunc does.
         :: inputs: - input: Description of the input parameter.
         :: output: Description of the return value.
         :: lang_version: 0.2.0
         :: assumptions: Assumes input is non-empty.
         :: side_effects: Writes to a log file.
 
         set result = input
         :: reason: Initial assignment before processing. @confidence: high
         # Regular comment here
         if no result then
             :: reason: Handle empty input case explicitly.
             return "Error: Input was empty"
         end if
         return result
     endfunc
     ```
 * **Standard Fields:** Common metadata fields like `purpose`, `inputs`, `output`, `algorithm`, `lang_version`, `assumptions`, `side_effects`, `trust_requirements`, `examples` are recommended, defined via `::`.
 * **Inline Annotations:** Use `::` prefix for inline metadata providing context (e.g., `:: reason:`, `:: confidence:`, `:: audience:`). A defined vocabulary will be developed.
 * **Format:** Values are typically plain text. Avoid complex embedded formats like YAML. Multi-line values might indent on subsequent lines (TBD precise formatting).
 
 ### Block Termination
 
 * Replaces generic `ENDBLOCK`.
 * Each block statement requires a matching, specific, lowercase end keyword:
     * `if ... [else ...] end if`
     * `for ... end for`
     * `while ... end while`
 * The interpreter strictly enforces correct nesting and matching keywords.
 
 ### Multi-line Strings & Templating
 
 * **Syntax:** Use Go-style **triple-backtick** for multi-line raw string literals.
     ```neuroscript
     set my_string = triple-backtickThis is line one.
     This is line two with an {{embedded_variable}}.triple-backtick
     ```
 * **Default Templating:** Placeholders (`{{placeholder}}`) within triple-backtick literals are **expanded by default** when the literal is evaluated, using the current variable scope. Explicit `eval()` is *not* required for these literals.
 * **Regular Strings:** Standard quoted strings (`"..."`, `'...'`) remain raw literals; placeholder expansion requires the `eval()` function as before.
 * **Literal `{{`:** *[Open Question]* A mechanism is needed to represent literal `{{` within triple-backtick strings (e.g., escape `\{{`, or doubling `{{{{` ? TBD).*
 
 ### Actor Interaction Statements
 
 * Introduces explicit keywords for interactions, replacing generic `call llm` and augmenting `call tool`. These likely act as built-in functions within expressions.
 * **`askAI(handle, question)`:** Interacts with an AI/LLM. `handle` identifies the specific AI configuration/persona/model. Returns the AI's response.
 * **`askHuman(handle, question)`:** Pauses execution and requests input from a human user. `handle` could specify interaction type (e.g., "confirm", "input"). Returns the human's response.
 * **`askComputer(handle, question)`:** Interacts with the underlying computer system or OS. `handle` specifies the subsystem (e.g., "shell", "filesystem"). `question` provides the command/query. Returns the result. (Precise scope vs. `call tool` TBD).
 
 ### Zero-Value Checks
 
 * Provides syntactic sugar for common checks against zero-values.
 * **`no <variable>`:** Evaluates to `true` if `<variable>` contains the zero value for its runtime type (`""`, `0`, `0.0`, `false`, empty/`nil` list, empty/`nil` map, `nil`).
 * **`some <variable>`:** Evaluates to `true` if `<variable>` does *not* contain its type's zero value.
 * **Boolean Warning:** Interpreter may optionally warn in debug/test modes if `no` is used with a boolean, suggesting `not` is clearer.
 
 ### Verification Steps
 
 * Provides keywords for explicit runtime checks/assertions.
 * **`must <expression>` / `mustBe <check_func>(<args>)`:** Keywords (preference over `assert`/`verify`) to perform runtime validation.
 * **Semantics:** Exact behavior TBD (e.g., does failure halt execution like an assertion, or return a special error state?). Requires defining useful built-in check functions.
 
 ### Other Constructs
 
 * **`set`, `return`, `emit`, `if/else`, `while`, `for each`, `call tool`, operators, list/map literals (`[]`, `{}`)** remain largely as defined previously, but using lowercase keywords and adhering to the "one way" principle (e.g., only `map["key"]` access).
 * `eval()` function remains for explicit placeholder expansion in regular strings.
 * `last` keyword remains available to access the result of the most recent `ask...` or `call tool` statement.
 
 
 
 ## Go Implementation Details (`neurogo` - `ng`)
 
 These points relate to the Go codebase implementing the NeuroScript interpreter.
 
 ### Dependency Injection (DI)
 
 * **Approach:** Use **Constructor Injection via interfaces**. Dependencies (like loggers, tool implementations, AI clients) should be defined by interfaces and passed into the constructors of structs that need them.
 * **Example Pattern:**
     ```go
     type MyService struct {
         logger interfaces.Logger
         aiClient interfaces.AIClient
     }
 
     func NewMyService(log interfaces.Logger, client interfaces.AIClient) *MyService {
         // ... nil checks ...
         return &MyService{logger: log, aiClient: client}
     }
     ```
 * **Benefits:** Decoupling, testability, clarity. Avoids global variables or complex DI frameworks.
 
 ### Logging
 
 * **Library:** Use the Go standard library's **`log/slog`** package.
 * **Handler:** Use `slog.TextHandler` for structured, human-readable, key-value output (not JSON). Configure log levels as needed.
 * **Integration:** Define a simple `Logger` interface (e.g., with `info`, `warn`, `error`, `debug` methods) matching `slog` methods. Create an adapter struct that implements this interface using an `*slog.Logger`. Inject the `Logger` interface into components using the DI pattern above.
 
 ### Tooling
 
 * Plan for a canonical code formatting tool: **`nsfmt`**, similar to `gofmt`, to enforce consistent style, spacing, and keyword casing.
 
 
 
 ## Future Directions / Open Questions (v0.2.0 Scope and Beyond)
 
 * **Define `try/catch/finally`:** Design and implement structured error handling, emphasizing simplicity and unambiguous operation.
 * **Define `must`/`mustBe` Semantics:** Finalize failure behavior (halt vs error) and define core check functions.
 * **Literal `{{` Handling:** Finalize mechanism for escaping `{{` within triple-backtick strings.
 * **Define `askComputer` Scope:** Clarify precise scope and implementation vs `call tool`.
 * **Define Metadata Vocab:** Establish the initial set of standard `::` metadata keywords (inline and header). Define multi-line value formatting.
 * **Add Arithmetic:** Implement built-in support for basic arithmetic operators (`+`, `-`, `*`, `/`, etc.).
 * **Implement Storage:** Implement non-mocked database/vector store integration for skill storage/retrieval[cite: 2].
 * **Refine `handle` Concept:** Flesh out the `handle` concept for `askAI`/`askHuman`/`askComputer`.
 * **Refine LLM Integration:** Context passing, configuration via `handle`[cite: 2].
 * **LSP Implementation:** Consider Language Server Protocol (LSP) implementation for editor support[cite: 2].
