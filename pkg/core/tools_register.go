// filename: pkg/core/tools_register.go
package core

// Removed imports for blocks and checklist

// registerCoreTools defines the specs for built-in tools and registers them
// by calling registration functions from specific tool files WITHIN THE CORE PACKAGE.
// It NO LONGER registers tools from external packages like blocks or checklist.
func registerCoreTools(registry *ToolRegistry) {
	// Register core tool groups
	registerFsTools(registry)       // From tools_fs.go
	registerVectorTools(registry)   // From tools_vector.go
	registerGitTools(registry)      // From tools_git.go
	registerStringTools(registry)   // From tools_string.go
	registerShellTools(registry)    // From tools_shell.go
	registerMathTools(registry)     // From tools_math.go
	registerMetadataTools(registry) // From tools_metadata.go
}

// RegisterCoreTools initializes all tools defined within the core package.
// This remains the function called by the application setup (e.g., neurogo/app.go)
func RegisterCoreTools(registry *ToolRegistry) {
	registerCoreTools(registry)
	// Calls to external package registrations (blocks, checklist) are REMOVED from here.
}
