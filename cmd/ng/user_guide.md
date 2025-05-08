 # NeuroGo User Guide (v0.3.x - Forward Looking)

 ## 1. Overview

 `neurogo` (or `ng`) is the command-line interface for the NeuroScript ecosystem. It serves as a versatile host environment for:

 * **Executing NeuroScript (`.ns`) files:** Running scripts for automation, AI interaction, and various tasks.
 * **Managing AI Workers:** Defining, configuring, and interacting with AI agents (LLMs) using the built-in AI Worker Management system. This allows orchestrating complex workflows involving multiple AI agents.
 * **Interactive Usage:** Providing interfaces like an interactive Terminal UI (TUI) and a basic Read-Eval-Print Loop (REPL) for direct interaction with the NeuroScript environment and AI workers. A Web UI is also planned.
 * **Utilizing Tools:** Exposing a rich set of tools (filesystem operations, Go language tools, Git integration, AI worker management, etc.) accessible from NeuroScript.

 The architecture emphasizes a unified environment where core components (Interpreter, LLM Client, AI Worker Manager) are always initialized. The application's behavior is then determined by configuration flags (e.g., running a startup script, activating the TUI).

 ## 2. Installation and Setup

 1.  **Build:** Build the `neurogo` executable from the source code using the Go toolchain:
     ```bash
     go build ./cmd/ng # Note: command is now 'ng'
     ```
 2.  **API Key:** `neurogo` requires an API key for interacting with external LLM services. Set **one** of the following environment variables:
     * `NEUROSCRIPT_API_KEY="YOUR_API_KEY_HERE"` (Recommended)
     * `GEMINI_API_KEY="YOUR_API_KEY_HERE"` (Legacy, may be used as fallback)
     Alternatively, use the `-api-key` flag. If no key is provided, `neurogo` will operate with a No-Op LLM client, disabling actual AI interaction.

 ## 3. Running NeuroGo

 `neurogo`'s behavior on startup depends on the flags provided:

 * **Startup Script (`-script <path>`):** If the `-script` flag is provided, `neurogo` will execute the specified NeuroScript file upon startup. After the script finishes, the application will exit unless an interactive UI flag (`-tui` or future `-webui-port`) is also present.
 * **Terminal UI (`-tui`):** If the `-tui` flag is present, `neurogo` will launch the interactive Terminal User Interface after initialization (and after running any specified startup script).
 * **Web UI (`-webui-port <port>` - *Planned*):** A future `-webui-port` flag will launch a web-based interface accessible via a browser.
 * **Basic REPL (Default):** If no startup script is provided and no UI flag is set, `neurogo` will fall back to a basic command-line REPL for minimal interaction.

 ## 4. Command-Line Flags

 ### General Configuration
 * `-sandbox <dir>`: Specifies the root directory for secure file operations and persistence (e.g., AI Worker definitions). Defaults to `.` (current directory). **Crucial for isolating work.** Resolved to an absolute path.
 * `-insecure`: If set, disables security checks like TLS verification for LLM clients. **Use with extreme caution!**
 * `-h`, `-help`: Display the help message.

 ### Script Execution
 * `-script <file.ns>`: Path to a NeuroScript file to execute on startup.
 * `-L <path>`: Add a directory to the library path for NeuroScript execution (e.g., finding imported `.ns` files). Can be used multiple times.
 * `-target <procedure_name>`: Specifies the target procedure to run within the startup script. Defaults to `main` or a procedure specified in the script's `:: target:` metadata.
 * `-arg <argument>`: Provides an argument to the target procedure in the startup script. Can be used multiple times (passed as `arg1`, `arg2`, etc.).

 ### User Interface
 * `-tui`: Launch the interactive Terminal UI.
 * `-webui-port <port>`: (*Planned*) Launch the Web UI on the specified port (0 or flag omitted disables).

 ### LLM Configuration
 * `-api-key <key>`: Explicitly provide the LLM API Key (overrides environment variables).
 * `-api-host <hostname>`: Specify a custom API endpoint/host for the LLM service.
 * `-model <name>`: Specify the default generative model name to use (e.g., `gemini-1.5-pro-latest`). This is used if an AI Worker Definition doesn't specify its own model.

 ### Logging
 * `-log-level <level>`: Set the logging level (`debug`, `info`, `warn`, `error`). Defaults to `info`.
 * `-log-file <path>`: Path to a file where logs should be written (appends). Defaults to stderr.

 ### Deprecated / Replaced Flags
 * `-agent`, `-sync`, `-clean-api`: These modes are deprecated. Their functionality should be accessed via NeuroScript tools (e.g., `AIWorker.*` tools, `Core.SyncDirectoryUp`, `Core.CleanFileAPI`).
 * `-sync-dir`, `-sync-filter`, `-sync-ignore-gitignore`: These may still be read by specific sync-related tools but do not dictate an application mode.
 * `-allowlist`: Tool permissions are intended to be managed via the AI Worker Management system definitions in the future.

 ## 5. Interactive Usage (TUI / REPL)

 ### 5.1 Terminal UI (`-tui`)
 The TUI provides a richer interactive experience:
 * **Conversation View:** Displays the history of interactions.
 * **Input Area:** Enter prompts or commands. Press Enter to submit. Use arrow keys for history.
 * **Status Bar:** Shows connection status, activity spinner, errors, and current mode/context.
 * **Help View:** Accessible via `Ctrl+H` (toggle full help).
 * **Commands:**
     * `/`: Enter command mode (input area prefix changes). Type command and press Enter.
     * `/sync [dir] [filter]`: (Functionality may change) Trigger file synchronization. Arguments might allow specifying target directory/filter temporarily.
     * `/quit`: Exit the TUI and `neurogo`.
     * (Other commands related to AI Worker Management may be added, e.g., `/list-workers`, `/interact <worker_id>`).
 * **Multi-line Input:** Press `Ctrl+M` to open an external editor (`nsinput` or `$EDITOR`) for easier multi-line prompt entry. Save and exit the editor to submit.

 ### 5.2 Basic REPL (Default Fallback)
 If `neurogo` starts without a startup script and without the `-tui` flag, it enters a basic REPL:
 * **Prompt:** Shows a `>` prompt.
 * **Input:** Type NeuroScript code or commands.
 * **Execution:** (*Partially Implemented*) Currently logs input. Planned to execute simple statements or predefined commands.
 * **Exiting:** Type `exit` or `quit` and press Enter, or use `Ctrl+C`.

 ## 6. AI Worker Management System

 A key feature of `neurogo` is its integration with the AI Worker Management system (`ai_wm_*`). This system, configured via JSON files within the `-sandbox` directory (`ai_worker_definitions.json`, `ai_worker_performance_data.json`), allows you to:

 * **Define Worker Blueprints (`AIWorkerDefinition`):** Specify different AI worker types, including their provider (Google, OpenAI, Ollama, etc.), model, capabilities, authentication method, rate limits, default configurations, and cost metrics.
 * **Manage Instances (`AIWorkerInstance`):** Spawn stateful instances of workers for conversational tasks, manage their lifecycle, and track their performance.
 * **Execute Stateless Tasks:** Run one-off AI tasks using a definition without needing a persistent instance.
 * **Monitor Performance:** Track token usage, costs, execution times, and success rates automatically.

 You interact with this system primarily through NeuroScript using the dedicated `AIWorker.*` tools (e.g., `AIWorkerDefinition.Add`, `AIWorkerInstance.Spawn`, `AIWorker.ExecuteStatelessTask`, `AIWorker.GetPerformanceRecords`). This enables building complex, multi-agent workflows controlled by NeuroScript.

 ## 7. Examples

 ### Run a Startup Script
 ```bash
 # Make sure NEUROSCRIPT_API_KEY is set
 # Executes 'scripts/setup_workers.ns' and then exits
 ./ng -script scripts/setup_workers.ns -L ./lib
 ```

 ### Start the Interactive TUI
 ```bash
 # Make sure NEUROSCRIPT_API_KEY is set
 # Starts the TUI for interactive use
 ./ng -tui -sandbox ./my_project_sandbox -log-level debug -log-file ng.log
 ```

 ### Run a Startup Script then Enter TUI
 ```bash
 # Make sure NEUROSCRIPT_API_KEY is set
 # Executes 'init.ns', then launches the TUI
 ./ng -script init.ns -tui -sandbox ./my_project_sandbox
 ```

 ### Default to Basic REPL
 ```bash
 # Make sure NEUROSCRIPT_API_KEY is set (or use -api-key)
 # Starts the basic REPL as no script or TUI flag is given
 ./ng -sandbox ./repl_sandbox
 ```

 ### Using AI Worker Tools (Conceptual Example within a `.ns` script)
 ```neuroscript
 # Assumes definitions exist in ai_worker_definitions.json in the sandbox

 # Execute a one-off task using a 'code-analyzer' definition
 set analysisResult = CALL AIWorker.ExecuteStatelessTask(definition_id: "code-analyzer-def-id", prompt: "Analyze this code: ...")
 Log("Analysis Result:", analysisResult)

 # Spawn a conversational worker instance
 set chatWorkerHandle = CALL AIWorkerInstance.Spawn(definition_id: "chat-bot-def-id")
 # ... use chatWorkerHandle with other tools to interact ...
 CALL AIWorkerInstance.Retire(instance_id: chatWorkerHandle) # Assuming handle is the ID
 ```