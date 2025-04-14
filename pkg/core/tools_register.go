// filename: pkg/core/tools_register.go
package core

import "fmt" // Keep fmt

// registerCoreTools defines the specs for built-in tools and registers them
// by calling registration functions from specific tool files WITHIN THE CORE PACKAGE.
func registerCoreTools(registry *ToolRegistry) error {
	// Register core tool groups, checking for errors
	if err := registerFsTools(registry); err != nil {
		return fmt.Errorf("failed registering FS tools: %w", err)
	}
	if err := registerVectorTools(registry); err != nil {
		return fmt.Errorf("failed registering Vector tools: %w", err)
	}
	if err := registerGitTools(registry); err != nil {
		return fmt.Errorf("failed registering Git tools: %w", err)
	}
	if err := registerStringTools(registry); err != nil {
		return fmt.Errorf("failed registering String tools: %w", err)
	}
	if err := registerShellTools(registry); err != nil {
		return fmt.Errorf("failed registering Shell tools: %w", err)
	}
	if err := registerMathTools(registry); err != nil {
		return fmt.Errorf("failed registering Math tools: %w", err)
	}
	if err := registerMetadataTools(registry); err != nil {
		return fmt.Errorf("failed registering Metadata tools: %w", err)
	}
	if err := registerListTools(registry); err != nil {
		return fmt.Errorf("failed registering List tools: %w", err)
	}
	// *** ADDED: Register Go AST tools ***
	if err := registerGoAstTools(registry); err != nil {
		return fmt.Errorf("failed registering Go AST tools: %w", err)
	}
	return nil // Success
}

// RegisterCoreTools initializes all tools defined within the core package.
// This remains the function called by the application setup (e.g., neurogo/app.go)
func RegisterCoreTools(registry *ToolRegistry) error {
	if err := registerCoreTools(registry); err != nil {
		return err // Propagate error
	}
	// External package tool registrations (like blocks, checklist)
	// should happen *after* core tools are registered, typically
	// in the main application setup (e.g., pkg/neurogo/app_script.go).
	return nil // Success
}
