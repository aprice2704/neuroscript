// filename: pkg/wm/ai_wm_default_wdefs.go
package wm

const AIWorkerDefinitions_Default string = `
[
  {
    "name": "gem1.5",
    "provider": "google",
    "model_name": "gemini-1.5-pro-latest",
    "auth": {
      "method": "env_var",
      "value": "GOOGLE_API_KEY"
    },
    "interaction_models": ["conversational", "stateless_task"],
    "capabilities": ["general", "text_generation", "reasoning"],
    "base_config": {
      "temperature": 0.7,
      "top_p": 0.9
    },
    "status": "active",
    "metadata": {
      "description": "Default Google Gemini 1.5 Pro model."
    }
  },
  {
    "name": "gpt4o",
    "provider": "openai",
    "model_name": "gpt-4o",
    "auth": {
      "method": "env_var",
      "value": "OPENAI_API_KEY"
    },
    "interaction_models": ["conversational", "stateless_task"],
    "capabilities": ["general", "text_generation", "code_generation", "reasoning"],
    "base_config": {
      "temperature": 0.7
    },
    "status": "active",
    "metadata": {
      "description": "Default OpenAI GPT-4o model."
    }
  },
  {
    "name": "ll3-loc",
    "provider": "ollama",
    "model_name": "llama3:latest",
    "auth": {
      "method": "none"
    },
    "interaction_models": ["conversational", "stateless_task"],
    "capabilities": ["general", "text_generation", "local_execution"],
    "base_config": {
      "temperature": 0.6,
      "num_ctx": 4096
    },
    "status": "active",
    "metadata": {
      "description": "Local Llama3 model via Ollama. Ensure Ollama server is running and model is pulled."
    }
  },
  {
    "name": "jun-dev",
    "provider": "google",
    "model_name": "gemini-1.5-pro-latest",
    "auth": {
      "method": "env_var",
      "value": "GOOGLE_API_KEY"
    },
    "interaction_models": ["stateless_task"],
    "capabilities": ["go_code_generation", "code_analysis", "code_completion"],
    "base_config": {
      "temperature": 0.3,
      "top_p": 0.8,
      "candidate_count": 1
    },
    "status": "active",
    "tool_allowlist": [
      "tool.ReadFile",
      "tool.WriteFile",
      "tool.ListDirectory",
      "tool.Gopls.GetDiagnostics", 
      "tool.Gopls.NotifyDidChange"
    ],
    "metadata": {
      "description": "A Gemini-based worker specialized for Go coding tasks, with a restricted toolset."
    }
  }
]
`