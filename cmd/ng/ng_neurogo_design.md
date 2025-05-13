 # NeuroGo (ng) Application and Package (pkg/neurogo) Design

 ## 1. Overview

 `ng` is the primary command-line interface for interacting with the NeuroScript ecosystem. It serves as the host environment for the NeuroScript interpreter (`pkg/core`), manages configuration, provides user interfaces (TUI, basic REPL, planned Web UI), and orchestrates interactions with AI models, primarily through the AI Worker Management system (`pkg/core/ai_wm_*`).

 The `pkg/neurogo` package contains the core application logic (`App`), configuration handling (`Config`), UI components (`tui`), and helper functions that bridge the command-line entry point (`cmd/ng/main.go`) with the underlying NeuroScript interpreter and AI Worker Manager.

 The refactored design moves away from strict, mutually exclusive execution modes (`-agent`, `-script`, `-sync`) towards a more unified architecture where `ng` initializes core components (Interpreter, LLM Client, AI Worker Manager) and then provides services or interfaces based on configuration flags (e.g., running a startup script, launching the TUI, starting a basic REPL, or potentially serving a Web UI). The AI Worker Manager (`ai_wm_*`) becomes central to managing potentially multiple, concurrent AI interactions.

 ## 2. Core Components and Concepts

 ### 2.1. `cmd/ng/main.go` (Entry Point)
 - **Role**: Parses command-line flags, initializes the logger, resolves paths (like the sandbox directory), creates the main `neurogo.App` instance, populates its configuration, initializes core components (LLMClient, Interpreter, AIWorkerManager) via the `App`, registers tools, optionally executes a startup script, and launches the primary user interface (TUI or REPL) or exits if only a script was run.
 - **Configuration**: Uses the standard `flag` package. Relies on `neurogo.Config` to hold parsed values.
 - **Initialization**: Orchestrates the setup sequence, ensuring essential components are ready before proceeding.
 - **Execution Flow**: Determines the final action based on flags (`-tui`) and whether a startup script (`-script`) was provided. Defaults to a basic REPL if no script is run and no TUI is requested.

 ### 2.2. `pkg/neurogo.App` (Application Core)
 - **Role**: Encapsulates the application's state and core components. Provides methods for initialization, accessing components, and executing primary tasks like running scripts.
 - **State**: Holds the `Config`, `logging.Logger`, `core.LLMClient`, `core.Interpreter`, and `core.AIWorkerManager`. Uses a `sync.RWMutex` for safe concurrent access if necessary (though current usage is primarily single-threaded during setup).
 - **Methods**: Includes `NewApp`, `CreateLLMClient`, `SetInterpreter`, `SetAIWorkerManager`, `GetInterpreter`, `GetAIWorkerManager`, `GetLogger`, `GetLLMClient`, `ExecuteScriptFile`, `loadLibraries`, `processNeuroScriptFile`, and methods implementing the `AppAccess` interface for the TUI.

 ### 2.3. `pkg/neurogo.Config` (Configuration)
 - **Role**: Struct holding configuration values parsed from flags and environment variables.
 - **Fields**: Contains paths (sandbox, startup script, libraries), API details (key, host, model name), sync parameters (for tools), and control flags (`-insecure`). Strict mode flags (`Run*Mode`) have been removed.
 - **Helpers**: Includes `NewConfig` and the `StringSliceFlag` custom type for multi-value flags (`-L`, `-arg`).

 ### 2.4. `pkg/neurogo` (Terminal UI)
 - **Role**: Provides an interactive terminal-based user interface using the `bubbletea` library.
 - **Functionality**: Displays conversation history, status information (spinner, errors), allows user input via a text area, handles commands (like `/sync`), and interacts with the `neurogo.App` instance (via the `AppAccess` interface) to get configuration, trigger actions (like sync), and potentially interact with the AI Worker Manager in the future.
 - **Activation**: Launched via the `-tui` command-line flag in `main.go`.

 ### 2.5. Basic REPL (Read-Eval-Print Loop)
 - **Role**: Provides a minimal command-line interaction loop as a fallback when `-tui` is not used and no startup script is executed.
 - **Functionality**: Reads user input line by line. Currently, it primarily recognizes `exit`/`quit`. Execution of NeuroScript statements is planned but not fully implemented; it currently logs the input and indicates non-implementation. Future versions could allow running simple statements, calling procedures, or executing specific REPL commands.
 - **Activation**: Implicitly started by `main.go` if other primary actions (TUI, script execution) are not taken.

 ### 2.6. Web UI (Future)
 - **Role**: Planned interface for interacting with `ng` via a web browser.
 - **Functionality**: Would likely involve a built-in HTTP server (`net/http`) serving static frontend assets (HTML, CSS, JS) and providing a WebSocket or REST API for interaction with the NeuroScript interpreter and AI Worker Manager.
 - **Activation**: Would likely be activated via a flag like `-webui-port <port>`.

 ### 2.7. AI Worker Manager Integration
 - **Initialization**: `main.go` creates the `core.AIWorkerManager` instance after the Logger and LLMClient are ready, passing them along with the sandbox path. The manager instance is then associated with the `neurogo.App`.
 - **Tooling**: `main.go` registers the `AIWorker.*` tools (via `core.RegisterAIWorkerTools`) with the interpreter, making them available for use in NeuroScript (e.g., in the startup script or via the REPL/TUI).
 - **Usage**: NeuroScript code executed by `ng` (startup script, REPL commands, potentially TUI actions) can now use the `AIWorker.*` tools to define, spawn, manage, and execute tasks using AI workers configured in the manager. The application itself (`ng`) acts primarily as the host environment for the manager and interpreter. Direct interaction logic (like `handleAgentTurn`) needs refactoring to operate within the context of specific worker instances or stateless tasks managed by the AI Worker Manager.

 ## 3. Key Workflows

 ### 3.1. Initialization (`main.go`)
 1. Parse Flags -> `neurogo.Config`
 2. Initialize Logger (`initializeLogger`)
 3. Create `neurogo.App` (`neurogo.NewApp`)
 4. Populate `app.Config`
 5. Create LLM Client (`app.CreateLLMClient`) -> `app.llmClient`
 6. Create Interpreter (`core.NewInterpreter`) -> `app.SetInterpreter`
 7. Create AI Worker Manager (`core.NewAIWorkerManager`) -> `app.SetAIWorkerManager`
 8. Register Tools (`core.RegisterCoreTools`, `core.RegisterAIWorkerTools`)

 ### 3.2. Startup Script Execution (`main.go` -> `app.ExecuteScriptFile`)
 1. If `-script <path>` is provided:
 2. Load Libraries (`app.loadLibraries`)
 3. Process main script file (`app.processNeuroScriptFile`) -> Adds procedures to interpreter.
 4. Determine target procedure ('main' or from `-target` / metadata).
 5. Execute target procedure (`interpreter.RunProcedure`). **Note:** Current implementation has limitations passing arguments from flags.

 ### 3.3. UI Activation (`main.go`)
 1. If `-tui` flag is set:
 2. Start TUI (`tui.Start(app)`), passing the `App` instance.
 3. Else if a startup script was executed:
 4. Exit gracefully.
 5. Else (no TUI, no script):
 6. Start basic REPL (`runRepl(ctx, app)`).

 ### 3.4. AI Worker Interaction (via NeuroScript)
 1. A NeuroScript (startup script, REPL command, etc.) calls an `AIWorker.*` tool (e.g., `AIWorker.ExecuteStatelessTask`, `AIWorkerInstance.Spawn`).
 2. The tool implementation (e.g., in `pkg/core/ai_wm_tools_*.go`) interacts with the `AIWorkerManager` instance held by the `App` (accessed via the `Interpreter`).
 3. The `AIWorkerManager` performs the requested action (validating, checking rate limits, calling `LLMClient`, updating state, logging performance).
 4. Results (or errors) are returned through the tool call back to the NeuroScript environment.

 ## 4. Configuration (`pkg/neurogo.Config`)

 Key configuration parameters managed via flags/env vars:
 - `SandboxDir` (`-sandbox`): Critical path for file operations, AI Worker Manager persistence (`ai_worker_definitions.json`, `ai_worker_performance_data.json`), and potentially other state. Resolved to an absolute path.
 - `APIKey` (`-api-key`, `GEMINI_API_KEY`, `NEUROSCRIPT_API_KEY`): The API key for the LLM service. Essential for most operations.
 - `APIHost` (`-api-host`): Optional override for the LLM API endpoint.
 - `ModelName` (`-model`): Specifies the default LLM model to use when not overridden by an AIWorkerDefinition.
 - `StartupScript` (`-script`): Path to the `.ns` file to run on startup.
 - `LibPaths` (`-L`): Directories containing NeuroScript library files (`.ns`) to be loaded before the startup script.
 - `TargetArg` (`-target`), `ProcArgs` (`-arg`): Specify the procedure and arguments for the startup script.
 - `LogLevel` (`-log-level`), `LogFile` (`-log-file`): Control logging verbosity and destination.
 - `Insecure` (`-insecure`): Disables security checks (e.g., TLS verification for LLM client). Use with caution.
 - `TuiMode` (`-tui`): Activates the terminal UI.

 ## 5. Future Work / Roadmap (.ndcl format)

 This section outlines planned enhancements, using the NeuroData Checklist format.

 ```neurodata-checklist
 :: title: NeuroGo Application Roadmap
 :: version: 0.1.0
 :: status: draft
 :: updated: 2025-05-07

 - | | 1. Core Architecture & AI WM Integration
   - [x] Refactor `main.go` to remove strict modes.
   - [x] Integrate `AIWorkerManager` initialization into `App`.
   - [x] Register `AIWorker.*` tools.
   - [ ] Refactor `handleAgentTurn` (or replace) to work with `AIWorkerManager` instances/tasks.
   - [ ] Design stateful worker interaction flow (managing ConversationManager alongside AIWorkerInstance).
   - [ ] Implement mechanism for supervisor scripts to assign/monitor tasks on workers using `ai_wm_*` tools.
   - [ ] Refine error handling and reporting from worker tasks back to the supervisor/user.

 - | | 2. User Interfaces
   - [x] Implement basic REPL fallback.
   - [ ] Enhance REPL to parse and execute simple NeuroScript statements/expressions.
   - [ ] Integrate TUI (`pkg/neurogo`) more deeply with `AIWorkerManager` (e.g., list workers, view status, interact with instances).
   - [ ] Design and implement Web UI (`-webui-port`).
     - [ ] Basic HTTP server setup.
     - [ ] API design (REST/WebSocket) for frontend interaction.
     - [ ] Frontend implementation (HTML/CSS/JS).

 - | | 3. Scripting & Execution
   - [x] Implement startup script execution (`-script`).
   - [ ] Properly implement argument passing (`-arg`, `-target`) to `RunProcedure`. Requires changes to `core.Interpreter.RunProcedure` signature (e.g., accept `map[string]interface{}`).
   - [ ] Define and implement REPL-specific commands (e.g., `!list-workers`, `!run-script <path>`).
   - [ ] Implement the "Apply prompt to files in tree" feature (potentially as a core function or a standard library script).
     - [ ] Design tool/function signature (input dir, filter, prompt, output handling).
     - [ ] Implement file walking and filtering.
     - [ ] Integrate with `AIWorker.ExecuteStatelessTask` or similar for processing each file.
     - [ ] Handle file updates (e.g., using diff/patch tools).

 - | | 4. Tooling & Ecosystem Interaction
   - [ ] Re-evaluate need for `toolsets.RegisterExtendedTools` vs. registering tools individually based on config/context.
   - [ ] Review and potentially refactor `Sync` and `CleanAPI` functionality as standard tools instead of separate modes/flags.
   - [ ] Develop `nsfmt` tool for formatting `.ns` files.
   - [ ] Add comprehensive end-to-end tests simulating supervisor-worker interactions.

 - | | 5. Configuration & Security
   - [ ] Implement more robust configuration loading (e.g., from a config file in addition to flags).
   - [ ] Refine tool allow/deny list mechanism, potentially integrating it with `AIWorkerDefinition` capabilities or security policies managed by `AIWorkerManager`.
   - [ ] Review sandbox security implementation details.
 ```

 ---