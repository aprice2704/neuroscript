// filename: pkg/core/tools_register.go
package core

import "fmt" // Keep fmt

// registerCoreTools defines the specs for built-in tools and registers them
// by calling registration functions from specific tool files WITHIN THE CORE PACKAGE.
func registerCoreTools(registry *ToolRegistry) error {
	// Register core tool groups, checking for errors
	if err := registerFsTools(registry); err != nil { // Assumes exists in tools_fs.go (or similar)
		return fmt.Errorf("failed registering FS tools: %w", err)
	}
	if err := registerVectorTools(registry); err != nil { // Assumes exists in tools_vector.go
		return fmt.Errorf("failed registering Vector tools: %w", err)
	}
	if err := registerGitTools(registry); err != nil { // Assumes exists in tools_git_register.go
		return fmt.Errorf("failed registering Git tools: %w", err)
	}
	if err := registerStringTools(registry); err != nil { // Assumes exists in tools_string.go
		return fmt.Errorf("failed registering String tools: %w", err)
	}
	if err := registerShellTools(registry); err != nil { // Assumes exists in tools_shell.go
		return fmt.Errorf("failed registering Shell tools: %w", err)
	}
	if err := registerMathTools(registry); err != nil { // Assumes exists in tools_math.go
		return fmt.Errorf("failed registering Math tools: %w", err)
	}
	if err := registerMetadataTools(registry); err != nil { // Assumes exists in tools_metadata.go
		return fmt.Errorf("failed registering Metadata tools: %w", err)
	}
	if err := registerListTools(registry); err != nil { // Assumes exists in tools_list_register.go
		return fmt.Errorf("failed registering List tools: %w", err)
	}
	if err := registerGoAstTools(registry); err != nil { // Assumes exists in tools_go_ast.go
		return fmt.Errorf("failed registering Go AST tools: %w", err)
	}
	if err := registerIOTools(registry); err != nil { // Assumes exists in tools_io.go
		return fmt.Errorf("failed registering IO tools: %w", err)
	}
	if err := registerFileAPITools(registry); err != nil { // Assumes exists in tools_file_api.go
		return fmt.Errorf("failed registering File API tools: %w", err)
	}

	// +++ ADDED: Call registerLLMTools (assuming signature is registerLLMTools(registry *ToolRegistry) error) +++
	if err := registerLLMTools(registry); err != nil { // Assumes exists in llm_tools.go
		return fmt.Errorf("failed registering LLM tools: %w", err)
	}

	return nil // Success
}

// RegisterCoreTools initializes all tools defined within the core package.
// This remains the function called by the application setup (e.g., neurogo/app.go)
func RegisterCoreTools(registry *ToolRegistry) error {
	if registry == nil {
		// Add a check for nil registry for robustness
		return fmt.Errorf("RegisterCoreTools called with a nil registry")
	}
	if err := registerCoreTools(registry); err != nil {
		return err // Propagate error
	}
	// External package tool registrations (like blocks, checklist)
	// should happen *after* core tools are registered, typically
	// in the main application setup (e.g., pkg/neurogo/app_agent.go).
	return nil // Success
}

// NOTE: The implementations for registerFsTools, registerVectorTools, etc.,
// must exist in their respective files within the core package.
// Ensure registerLLMTools in pkg/core/llm_tools.go is defined as
// func registerLLMTools(registry *ToolRegistry) error
