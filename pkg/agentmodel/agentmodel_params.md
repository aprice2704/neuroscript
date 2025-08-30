Of course. To create a unified `agentmodel` object that can represent the full spectrum of API options across all major LLM vendors, we need to define a comprehensive set of fields and values. This "superset" structure will allow you to configure any available feature and will include switches to indicate which vendors support each specific control.

Below is a detailed breakdown of the necessary fields, categorized for clarity. This structure can serve as the blueprint for your `agentmodel` objects.

### I. Core Generation Parameters

These parameters control the fundamental aspects of the text generation process, such as its length, randomness, and when it should stop.

| Parameter | Type | Description | Vendor Support & Specifics |
| :--- | :--- | :--- | :--- |
| **`model`** | `string` | The specific model ID to use for the request. | **All Vendors**: ✅ \<br\> *(e.g., `gpt-4o`, `gemini-1.5-pro`, `claude-3-opus-20240229`, `command-r-plus`, `mistral-large-latest`, `Llama-4-Scout-17B`)* |
| **`max_tokens`** | `integer` | The maximum number of tokens to generate in the response. | **All Vendors**: ✅ \<br\> - **Meta**: `max_completion_tokens` [1] \<br\> - **Google**: `maxOutputTokens` [2] \<br\> - **Others**: `max_tokens` [3, 4, 5, 6] |
| **`stop_sequences`** | `array[string]` | A list of text sequences that will cause the model to stop generating further tokens. | **All Vendors**: ✅ \<br\> - **OpenAI/Meta/Mistral**: `stop` [3, 5] \<br\> - **Google**: `stopSequences` [2] \<br\> - **Anthropic/Cohere**: `stop_sequences` [4, 6] |
| **`stream`** | `boolean` | If `true`, the response will be sent incrementally as server-sent events. | **All Vendors**: ✅ [3, 1, 5, 7, 6] \<br\> *(Google's implementation is a separate method, `streamGenerateContent`)* |

### II. Sampling and Determinism Controls

These parameters influence the probabilistic nature of the model's output, managing the balance between creativity and predictability.

| Parameter | Type | Description | Vendor Support & Specifics |
| :--- | :--- | :--- | :--- |
| **`temperature`** | `float` | Controls randomness. Lower values (e.g., 0.1) make the output more deterministic, while higher values (e.g., 0.9) make it more random. | **All Vendors**: ✅ \<br\> - **OpenAI/Meta**: Range `0.0` to `2.0` [5] \<br\> - **Anthropic**: Range `0.0` to `1.0` [6] \<br\> - **Mistral**: Recommended `0.0` to `0.7` [3] \<br\> - **Cohere**: Non-negative float [4] \<br\> - **Google**: Range `0.0` to `1.0` [2] |
| **`top_p`** | `float` | Nucleus sampling. The model considers only the tokens with the highest probability mass that add up to this value. | **All Vendors**: ✅ \<br\> - **Cohere**: `p` (Range `0.01` to `0.99`) [4] \<br\> - **Google**: `topP` [8] \<br\> - **Others**: `top_p` (Range `0.0` to `1.0`) [3, 5, 6] |
| **`top_k`** | `integer` | The model samples from the `k` most likely next tokens. | **Google/Anthropic/Cohere/Meta**: ✅ \<br\> - **Cohere**: `k` (Range `0` to `500`) [4, 9] \<br\> - **Google**: `topK` [2] \<br\> - **Others**: `top_k` [6] \<br\> **OpenAI/Mistral**: ❌ |
| **`seed`** | `integer` | If specified, the model will make a best effort to return deterministic results. | **OpenAI/Cohere/Mistral/Meta**: ✅ \<br\> - **Mistral**: `random_seed` [3] \<br\> - **Others**: `seed` [4, 5] \<br\> **Google/Anthropic**: ❌ |
| **`n`** | `integer` | The number of different completions to generate for the prompt. | **OpenAI/Meta**: ✅ [5] \<br\> **Others**: ❌ |
| **`logprobs`** | `boolean` or `integer` | Whether to return the log probabilities of the output tokens. | **OpenAI/Cohere/Meta**: ✅ \<br\> - **OpenAI**: `logprobs` (integer, returns top `n` logprobs) [5] \<br\> - **Cohere/Meta**: `logprobs` (boolean) [4] \<br\> **Google/Anthropic/Mistral**: ❌ |
| **`logit_bias`** | `object` | A map of token IDs to a bias value (-100 to 100) to increase or decrease the likelihood of specific tokens appearing. | **OpenAI (via Azure)**: ✅ [5] \<br\> **Others**: ❌ |

### III. Repetition and Content Control

These parameters help prevent the model from repeating itself and allow for fine-grained control over the generated content.

| Parameter | Type | Description | Vendor Support & Specifics |
| :--- | :--- | :--- | :--- |
| **`presence_penalty`** | `float` | Penalizes new tokens based on whether they have appeared in the text so far, encouraging the model to introduce new topics. | **OpenAI/Cohere/Mistral/Meta/Google**: ✅ \<br\> - **Range**: Typically `-2.0` to `2.0` for most [3, 5]; `0.0` to `1.0` for Cohere.[4] |
| **`frequency_penalty`** | `float` | Penalizes new tokens based on their existing frequency in the text so far, discouraging repetition of the same words. | **OpenAI/Cohere/Mistral/Meta/Google**: ✅ \<br\> - **Range**: Typically `-2.0` to `2.0` for most [3, 5]; `0.0` to `1.0` for Cohere.[4] |
| **`repetition_penalty`** | `float` | An alternative, multiplicative penalty for repetition. Values \> 1 penalize, \< 1 encourage. | **Meta**: ✅ (Range `0.01` to `5.0`) \<br\> **Others**: ❌ |

### IV. Agentic Capabilities: Tools & Structured Outputs

These are the core parameters that enable an LLM to function as an agent by interacting with external systems and producing reliable, machine-readable output.

| Parameter | Type | Description | Vendor Support & Specifics |
| :--- | :--- | :--- | :--- |
| **`tools`** | `array[object]` | A list of tool definitions (e.g., functions with name, description, and JSON schema for parameters) that the model can call. | **All Vendors**: ✅ [3, 4, 1, 5, 10, 6] |
| **`tool_choice`** | `string` or `object` | Controls how the model uses the provided tools. | **All Vendors**: ✅ \<br\> - **Common Values**: `none`, `auto` (default), `required` (Meta).[1] \<br\> - **Forcing a specific tool**: Supported by Meta, Anthropic, and Mistral via an object specifying the tool name.[3, 1, 6] |
| **`parallel_tool_calls`** | `boolean` | A switch to enable or disable the model's ability to call multiple tools in a single turn. | **OpenAI/Google/Mistral/Meta**: ✅ \<br\> - **Mistral**: `parallel_tool_calls` (defaults to `true`) [3] \<br\> - **Anthropic**: `disable_parallel_tool_use` (boolean, inverted logic) [11] \<br\> **Cohere**: ❌ |
| **`response_format`** | `object` | Specifies the format for the model's output, such as forcing a valid JSON object. | **All Vendors**: ✅ \<br\> - **Simple JSON Mode**: `{ "type": "json_object" }` (OpenAI, Cohere, Mistral).[3, 4] \<br\> - **JSON Schema Enforcement**: `{ "type": "json_schema", "json_schema": {...} }` (Meta, Cohere, Mistral).[3, 4, 1] Google and Anthropic support this via their tool definition. |

### V. Vendor-Specific & Advanced Features

These are unique or specialized parameters offered by specific vendors that provide advanced control over the model's behavior or enable unique functionalities.

| Parameter | Type | Description | Vendor Support & Specifics |
| :--- | :--- | :--- | :--- |
| **`reasoning_effort`** | `string` | Constrains the amount of internal reasoning or "thinking" the model performs before answering. | **Meta/Mistral/OpenAI (GPT-5)**: ✅ \<br\> - **Meta**: `reasoning_effort` (Enum: `none`, `low`, `medium`, `high`) \<br\> - **Mistral**: `prompt_mode` (Enum: `reasoning`) for Magistral models [3] \<br\> - **OpenAI (GPT-5)**: `reasoning_effort` (Enum: `minimal`) [12] \<br\> **Others**: ❌ |
| **`thinking`** | `object` | Anthropic's specific implementation for exposing the model's chain-of-thought. Requires a `budget_tokens` value. | **Anthropic**: ✅ [6, 13] \<br\> **Others**: ❌ |
| **`documents`** | `array[object]` | A list of documents for the model to use for Retrieval-Augmented Generation (RAG) to provide grounded, cited answers. | **Cohere**: ✅ [4] \<br\> **Others**: ❌ (Handled via other mechanisms like File Search tools) |
| **`citation_options`** | `object` | Options for controlling how citations are generated when using the `documents` parameter. | **Cohere**: ✅ (Enum: `FAST`, `ACCURATE`) [4] \<br\> **Others**: ❌ |
| **`safe_prompt`** | `boolean` | Injects a safety prompt before the user's conversation to enforce guardrails. | **Mistral**: ✅ [3] \<br\> **Others**: ❌ (Safety is handled via other mechanisms like `safetySettings` in Google [10]) |
| **`user`** | `string` | A unique identifier for the end-user to help the vendor monitor and detect abuse. | **OpenAI/Meta**: ✅ [1, 5] \<br\> **Anthropic**: ✅ (`metadata.user_id`) [6] \<br\> **Others**: ❌ |

By implementing a data structure that accommodates this superset of fields, your `agentmodel` objects will be equipped to leverage the full, nuanced power of each LLM provider's API, ensuring maximum flexibility and control for your users.