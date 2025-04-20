# NeuroGo User Guide

## Overview

`neurogo` is a command-line application designed to interact with Large Language Models (LLMs) via the Google AI API. It serves several primary purposes:

* **Interactive Agent Mode:** Chat with an LLM agent that can utilize a defined set of tools to perform tasks, including file manipulation within a secure sandbox.
* **NeuroScript Execution:** Run scripts written in the NeuroScript language (`.ns` files).
* **File Synchronization:** Synchronize local files with the Google AI File API, useful for providing context to the LLM.
* **API File Management:** Cleanly remove all files currently stored in the Google AI File API.

## Installation and Setup

1.  **Build:** Build the `neurogo` executable from the source code using the Go toolchain:
    ```bash
    go build ./cmd/neurogo
    ```
2.  **API Key:** Set the `GEMINI_API_KEY` environment variable to your Google AI API key. `neurogo` requires this key to communicate with the LLM and File APIs.
    ```bash
    export GEMINI_API_KEY="YOUR_API_KEY_HERE"
    ```

## Execution Modes

`neurogo` operates in one of several mutually exclusive modes, determined by command-line flags. The order of precedence is: `-clean-api` > `-sync` > `-script` > `-agent` (default).

* **Agent Mode (Default):** `-agent`
    * Runs an interactive chat session with the LLM agent.
    * This is the default mode if no other mode flag (`-clean-api`, `-sync`, `-script`) is specified.
    * See "Agent Mode Usage" below for interactive commands.

* **Script Mode:** `-script <file.ns>`
    * Executes the specified NeuroScript file (`.ns`).
    * Requires the path to the script file as an argument.

* **Sync Mode:** `-sync`
    * Performs a single synchronization of local files to the Google AI File API based on configuration flags (`-sync-dir`, `-sync-filter`, etc.) and then exits.

* **Clean API Mode:** `-clean-api`
    * Deletes ALL files currently associated with your API key from the Google AI File API.
    * This flag must be used alone (potentially with logging or model flags).
    * It requires explicit confirmation before proceeding. **Use with extreme caution!**

## Command-Line Flags

### Mode Selection

* `-agent`: Force execution in interactive agent mode.
* `-script <file.ns>`: Execute the specified NeuroScript file.
* `-sync`: Run file synchronization based on config flags and exit.
* `-clean-api`: Delete all files from the File API (requires confirmation).

### Sync Configuration (Used by `-sync` mode and `/sync` agent command)

* `-sync-dir <dir>`: Directory to synchronize. Defaults to `.` (current directory). Used by the `-sync` flag and as the *explicit* target for the bare `/sync` agent command if set.
* `-sync-filter <pattern>`: Glob pattern (matching filename only) to include specific files during sync (e.g., `*.go`, `data?.txt`).
* `-sync-ignore-gitignore`: If set, the `.gitignore` file in the sync directory will be ignored. Defaults to `false` (gitignore is respected).

### Agent & Script Configuration

* `-sandbox <dir>`: Specifies the root directory for secure file operations by the agent/script tools. Defaults to `.` (current directory). Relative paths are interpreted relative to where `neurogo` is run. **Important:** This also influences the default target directory for the bare `/sync` agent command if `-sync-dir` is not explicitly set.
* `-allowlist <file>`: Path to a file containing a list of tools (one per line) that the agent is permitted to use.
* `-attach <file>`: Attach a local file to the agent session context initially. The file will be uploaded to the File API. Can be used multiple times. Paths are validated against the sandbox.

### Script Execution Configuration (Used with `-script`)

* `-L <path>`: Add a directory to the library path for NeuroScript execution (e.g., finding imported modules). Can be used multiple times.
* `-target <arg>`: A specific target argument passed to the main procedure of the script.
* `-arg <arg>`: A general argument passed to the main procedure of the script. Can be used multiple times.

### LLM Configuration

* `-model <name>`: Specify the generative model name to use (e.g., `gemini-1.5-pro-latest`, `gemini-1.5-flash-latest`). If omitted, a default model is used.

### Logging

* `-debug-log <file>`: Path to a file where detailed debug logs will be written. If omitted, debug logs are discarded.
* `-llm-debug-log <file>`: Path to a file where raw LLM request/response JSON communication will be written. If omitted, these logs are discarded.

### Other

* `-h`, `-help`: Display the help message listing all flags and modes.

## Agent Mode Usage

When running in agent mode (either explicitly with `-agent` or by default), you can interact with the LLM via prompts. There are also special commands:

* `/sync`:
    * Triggers a file synchronization process *during* the agent session.
    * The target directory is determined as follows:
        1.  Uses the directory specified by `-sync-dir` if that flag was provided.
        2.  Otherwise, uses the directory specified by `-sandbox` if that flag was provided.
        3.  Otherwise, defaults to `.` (the directory where `neurogo` was started).
    * Uses the filter from `-sync-filter` and the ignore setting from `-sync-ignore-gitignore`.
    * Files successfully synced become available in the context for the *next* prompt you provide to the agent.

* `/sync <directory> [filter]`:
    * Triggers a file synchronization for the specified `<directory>` (relative to the sandbox).
    * Optionally, you can provide a glob `[filter]` pattern for this specific sync operation, overriding the global `-sync-filter` for this command only.
    * This command always respects the `.gitignore` file (it does not use the `-sync-ignore-gitignore` flag).
    * Files successfully synced become available in the context for the *next* prompt.

* `/m`:
    * Launches an external editor (`nsinput`) to allow you to enter multi-line prompts more easily.
    * Save and exit the editor to submit the prompt, or quit the editor without saving to cancel.

* `quit`:
    * Exits the `neurogo` application.

## Examples

### Run Agent Mode (Default)
```bash
# Make sure GEMINI_API_KEY is set
./neurogo
```

### Run Agent Mode with Sandbox and Initial Attachment
```bash
# Make sure GEMINI_API_KEY is set
./neurogo -agent -sandbox ./project_files -attach ./project_files/main.go
```

### Execute a NeuroScript File
```bash
# Make sure GEMINI_API_KEY is set
./neurogo -script ./scripts/generate_docs.ns -L ./scripts/lib -target main.go
```

### Synchronize Files and Exit
```bash
# Make sure GEMINI_API_KEY is set
# Sync only *.go and *.md files from ./src directory
./neurogo -sync -sync-dir ./src -sync-filter '*.[gm][od]'
```

### Clean All Files from API (Use Carefully!)
```bash
# Make sure GEMINI_API_KEY is set
./neurogo -clean-api
(Requires confirmation)
```

### Agent Mode: Sync Sandbox Directory, then Ask Question
```bash
# Make sure GEMINI_API_KEY is set
./neurogo -sandbox ./project

# Inside agent prompt:
# Prompt: /sync
# (Wait for sync confirmation)
# Prompt: Summarize the Go files in the project.
```