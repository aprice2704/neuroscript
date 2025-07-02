 # AI Worker Definition Specification

 This document outlines the structure and fields for defining AI Workers within the NeuroScript system. These definitions are typically stored in a JSON array format.

 ## Root Structure

 The definitions consist of a JSON array, where each element in the array is an object representing a single AI Worker definition.

 json  [  {  // Worker Definition 1  },  {  // Worker Definition 2  }  // ... more worker definitions  ] 

 ## Worker Definition Object

 Each AI Worker Definition object has the following fields:

 ### name
 * Type: string
 * Description: A unique identifier for this worker definition. This name is used to reference the worker within the system.
 * Example: "google-gemini-1.5-pro"

 ### provider
 * Type: string
 * Description: The name of the AI model provider.
 * Example: "google", "openai", "ollama"

 ### model_name
 * Type: string
 * Description: The specific model name as recognized by the provider. This might include versioning information (e.g., -latest).
 * Example: "gemini-1.5-pro-latest", "gpt-4o", "llama3:latest"

 ### auth
 * Type: object
 * Description: Specifies the authentication method and necessary details for accessing the model.
 * Fields:
     * method (string): The authentication method to use.
         * "env_var": Indicates the authentication credential (e.g., API key) should be retrieved from an environment variable.
         * "none": Indicates no explicit authentication is required (e.g., for local models).
     * value (string, optional): The name of the environment variable if method is "env_var".
 * Example:
     json  "auth": {  "method": "env_var",  "value": "GOOGLE_API_KEY"  } 
     json  "auth": {  "method": "none"  } 

 ### interaction_models
 * Type: array of string
 * Description: Defines how the AI worker is expected to interact and maintain state. This field is crucial for determining the worker's behavior in different scenarios. An AI worker can support one or more interaction models.
 * Possible Values:
     * "conversational":
         * Use Case: Suitable for multi-turn dialogues where the AI needs to remember previous parts of the conversation and build upon them.
         * State Management: The worker is expected to maintain conversation history or be provided with it for subsequent requests.
         * Behavior: Responses are context-aware, considering the flow of the ongoing interaction.
         * Example Models: General-purpose chatbots, long-form content generation assistants.
     * "stateless_task":
         * Use Case: Designed for single, self-contained tasks where each request is independent of others. The AI processes the input and returns an output without needing to remember past interactions for this specific task instance.
         * State Management: No intrinsic memory of previous requests is assumed or required for the immediate task. If context from a broader session is needed, it must be explicitly provided in each request.
         * Behavior: The worker focuses solely on the current input to produce a relevant output. This is suitable for tool-like invocations or specific, bounded operations.
         * Example Models: Specialized tools (like the "coding-assistant-gemini"), function calling, data transformation, or one-shot queries.
 * Clarification:
     * A worker defined as ["conversational", "stateless_task"] can be used in either mode. The system or calling agent will decide which mode is appropriate for a given operation.
     * A worker defined only as ["stateless_task"] (like "coding-assistant-gemini") should primarily be used for specific, bounded operations where conversation history is not the main driver of the interaction. While it might be part of a larger conversational flow orchestrated by an agent, its individual calls are treated as discrete tasks.
     * A worker defined only as ["conversational"] would be expected to always leverage conversation history.
 * Example: ["conversational", "stateless_task"], ["stateless_task"]

 ### capabilities
 * Type: array of string
 * Description: A list of keywords describing the AI's strengths and intended use cases. This helps in selecting the appropriate worker for a task.
 * Example Values: "general", "text_generation", "reasoning", "code_generation", "local_execution", "go_code_generation", "code_analysis", "code_completion"
 * Example: ["general", "text_generation", "reasoning"]

 ### base_config
 * Type: object
 * Description: Default configuration parameters for the model. These can often be overridden at the time of a specific request. The exact available parameters depend on the provider and model.
 * Common Parameters:
     * temperature (number): Controls randomness. Lower values are more deterministic.
     * top_p (number): Nucleus sampling parameter.
     * num_ctx (integer): Context window size (specific to some models like Ollama).
     * candidate_count (integer): Number of candidate responses to generate.
 * Example:
     json  "base_config": {  "temperature": 0.7,  "top_p": 0.9  } 

 ### status
 * Type: string
 * Description: Indicates the operational status of the worker definition.
 * Possible Values:
     * "active": The worker is available for use.
     * "inactive" or "disabled" (implied): The worker is configured but not currently available.
 * Example: "active"

 ### tool_allowlist (Optional)
 * Type: array of string
 * Description: If present, this field specifies a list of tools (identified by their registered names) that this AI worker is permitted to use. If omitted or empty, the worker might have access to a default set of tools or no tools, depending on system configuration. This is particularly useful for creating specialized workers with restricted capabilities.
 * Example:
     json  "tool_allowlist": [  "tool.ReadFile",  "tool.WriteFile",  "tool.ListDirectory",  "tool.Gopls.GetDiagnostics",  "tool.Gopls.NotifyDidChange"  ] 

 ### metadata
 * Type: object
 * Description: An open-ended field for storing additional information about the worker definition.
 * Common Sub-fields:
     * description (string): A human-readable description of the worker.
 * Example:
     json  "metadata": {  "description": "Default Google Gemini 1.5 Pro model."  } 

 ## Example: Full AI Worker Definition (Google Gemini 1.5 Pro)

 json  {  "name": "google-gemini-1.5-pro",  "provider": "google",  "model_name": "gemini-1.5-pro-latest",  "auth": {  "method": "env_var",  "value": "GOOGLE_API_KEY"  },  "interaction_models": ["conversational", "stateless_task"],  "capabilities": ["general", "text_generation", "reasoning"],  "base_config": {  "temperature": 0.7,  "top_p": 0.9  },  "status": "active",  "metadata": {  "description": "Default Google Gemini 1.5 Pro model."  }  } 


 ## Example: Specialized Coding Assistant (Gemini)

 json  {  "name": "coding-assistant-gemini",  "provider": "google",  "model_name": "gemini-1.5-pro-latest",  "auth": {  "method": "env_var",  "value": "GOOGLE_API_KEY"  },  "interaction_models": ["stateless_task"],  "capabilities": ["go_code_generation", "code_analysis", "code_completion"],  "base_config": {  "temperature": 0.3,  "top_p": 0.8,  "candidate_count": 1  },  "status": "active",  "tool_allowlist": [  "tool.ReadFile",  "tool.WriteFile",  "tool.ListDirectory",  "tool.Gopls.GetDiagnostics",  "tool.Gopls.NotifyDidChange"  ],  "metadata": {  "description": "A Gemini-based worker specialized for Go coding tasks, with a restricted toolset."  }  } 
