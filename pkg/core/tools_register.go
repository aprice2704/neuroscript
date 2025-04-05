// pkg/core/tools_register.go
package core

// registerCoreTools defines the specs for built-in tools and registers them
// by calling registration functions from specific tool files.
func registerCoreTools(registry *ToolRegistry) {
	// Register tool groups
	registerFsTools(registry)           // From tools_fs.go
	registerVectorTools(registry)       // From tools_vector.go
	registerGitTools(registry)          // From tools_git.go
	registerStringTools(registry)       // From tools_string.go
	registerShellTools(registry)        // From tools_shell.go
	registerCompositeDocTools(registry) // ** NEW: From tools_composite_doc.go **

	// Example: If a tool doesn't fit a group, register it directly
	// registry.RegisterTool(ToolImplementation{
	// 	Spec: ToolSpec{ Name: "MyStandaloneTool", ... },
	// 	Func: toolMyStandaloneTool,
	// })
}
