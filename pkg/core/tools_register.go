// filename: pkg/core/tools_register.go
package core

// registerCoreTools defines the specs for built-in tools and registers them
// by calling registration functions from specific tool files WITHIN THE CORE PACKAGE.
func registerCoreTools(registry *ToolRegistry) {
	// Register core tool groups
	registerFsTools(registry)     // From tools_fs.go
	registerVectorTools(registry) // From tools_vector.go
	registerGitTools(registry)    // From tools_git.go
	registerStringTools(registry) // From tools_string.go
	registerShellTools(registry)  // From tools_shell.go
	registerMathTools(registry)   // *** ADDED: Register math tools ***

	// Block tool registration might happen here or in main.go depending on dependencies
	// Blocks pkg itself doesn't depend on core interpreter state typically.
	// Let's assume it's registered elsewhere for now, e.g., in main.go or neurogo/app.go
	// blocks.RegisterBlockTools(registry)
}

// RegisterCoreTools initializes all tools defined within the core package.
// This remains the function called by the application setup (e.g., neurogo/app.go)
func RegisterCoreTools(registry *ToolRegistry) {
	registerCoreTools(registry)
}
