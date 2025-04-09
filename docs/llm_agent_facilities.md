# Notes: gonsi as LLM Agent Mechanism

## 1. LLM Interaction API (LLM -> gonsi Request)

* Goal: Allow LLM to request gonsi execute a TOOL during a conversation.
* Option A: Structured Chat Response:
    * LLM responds with text containing a predefined structure (e.g., JSON) indicating a TOOL call request (tool_name, arguments).
    * gonsi parses this structured part of the response.
    * Pros: Works with standard chat/completion APIs (like the current llm.go setup [cite: uploaded:neuroscript/pkg/core/llm.go]).
    * Cons: Requires careful prompt engineering for the LLM to reliably generate the correct structure; parsing logic in gonsi can be brittle.
* Option B: Function Calling / Tool Use API:
    * Modern LLM APIs (including Gemini) often have explicit support for this.
    * gonsi declares available TOOLs (name, description, parameters - like ToolSpec [cite: uploaded:neuroscript/pkg/core/tools_types.go]) to the LLM API.
    * When the LLM needs a tool, the API response itself is a structured object indicating the tool name and arguments to call.
    * gonsi receives this object, executes the tool, and sends the result back in the next API call.
    * Pros: Most robust, standard, designed for this purpose. Avoids parsing responses.
    * Cons: Requires modifying llm.go [cite: uploaded:neuroscript/pkg/core/llm.go] to use the specific function calling mode/parameters of the Gemini API.
* Option C: MCP (Multi-Context Prompting):
    * Less common in standard commercial APIs compared to function calling. Might be overly complex or not directly supported by the current Gemini endpoint.
* Recommendation: Prioritize Option B (Function Calling). It's the industry standard and most reliable way. If that proves difficult with the specific API endpoint, fall back to Option A (Structured Chat Response).

## 2. Conversation State Management

* Requirement: gonsi must maintain the history of the interaction across multiple turns (user prompt -> LLM response -> [LLM tool request -> gonsi tool execution -> tool result -> LLM continuation] -> final response).
* Implementation:
    * The logic managing the CALL LLM needs expansion.
    * It must store the sequence of messages (user, assistant, tool_request, tool_result).
    * For each call to the LLM API (especially after a tool result), the relevant history must be formatted and included according to the API requirements (e.g., the contents array for Gemini multi-turn chat, or the specific format for function calling responses).
    * This state management likely belongs within the Go code handling the LLM interaction (llm.go and potentially interpreter_simple_steps.go where CALL LLM is handled).

## 3. Security & Limiting Actions

* Requirement: Prevent the LLM from making gonsi execute arbitrary or harmful commands/actions. This is critical.
* Mechanisms:
    * Tool Allowlist: Explicitly define which TOOLs are available to be called by the LLM. Do not expose all registered tools [cite: uploaded:neuroscript/pkg/core/tools_registry.go] if not necessary. This might involve a separate registration or filtering layer for LLM-callable tools.
    * Strict Argument Validation: Enhance ValidateAndConvertArgs [cite: uploaded:neuroscript/pkg/core/tools_validation.go] or add further checks for arguments received from the LLM for tool calls. Pay extra attention to file paths, URLs, shell command arguments, etc. Check types, formats, and potentially value ranges or patterns.
    * Path Sandboxing: Consistently use SecureFilePath [cite: uploaded:neuroscript/pkg/core/tools_helpers.go] (or equivalent) for all filesystem operations requested via TOOLs (ReadFile, WriteFile, ListDirectory, GitAdd, etc. [cite: uploaded:neuroscript/pkg/core/tools_fs.go, uploaded:neuroscript/pkg/core/tools_git.go]) to ensure they stay within the allowed working directory.
    * Restrict/Disable Dangerous Tools: TOOL.ExecuteCommand [cite: uploaded:neuroscript/pkg/core/tools_shell.go] is high-risk. Consider:
        * Not exposing it to the LLM at all.
        * If required, creating a highly restricted version that only allows specific commands from a hardcoded allowlist and performs extreme sanitization on arguments.
    * Resource Limits: Implement timeouts for LLM API calls and potentially for the execution of tools called by the LLM.
    * Deny Direct NS Execution: The LLM should only be allowed to request registered TOOLs, not ask gonsi to execute arbitrary NeuroScript code strings it generates dynamically within the agent loop.
    * Prompt Guidance (Secondary): Instructing the LLM about disallowed actions is helpful but not sufficient for security. Hardcoded checks and limitations in gonsi are essential.
* Recommendation: Implement multiple layers: Tool allowlisting, strict validation, path sandboxing, and heavy restrictions (or disabling) of ExecuteCommand. Security needs careful design.
