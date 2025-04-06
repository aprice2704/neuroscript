// pkg/core/tools_register.go
package core

// No longer need to import checklist or composite doc packages here

// registerCoreTools defines the specs for built-in tools and registers them
// by calling registration functions from specific tool files WITHIN THE CORE PACKAGE.
func registerCoreTools(registry *ToolRegistry) {
	// Register core tool groups
	registerFsTools(registry)     // From tools_fs.go
	registerVectorTools(registry) // From tools_vector.go
	registerGitTools(registry)    // From tools_git.go
	registerStringTools(registry) // From tools_string.go
	registerShellTools(registry)  // From tools_shell.go
	// *** REMOVED composite doc registration ***
	// registerCompositeDocTools(registry) // This function is now removed/obsolete

	// Checklist registration is handled in main.go
	// Block tool registration is handled in main.go
}

// RegisterCoreTools initializes all tools defined within the core package.
// This remains the function called by main.go
func RegisterCoreTools(registry *ToolRegistry) {
	registerCoreTools(registry)
}
