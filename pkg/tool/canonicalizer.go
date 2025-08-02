// NeuroScript Version: 0.1.0
// File version: 2.0.0
// Purpose: Provides a function to canonicalize tool names.
// filename: pkg/tool/canonicalizer.go

package tool

import "strings"

const toolPrefix = "tool."

// CanonicalizeToolName ensures a tool name has a single, correct 'tool.' prefix.
// This function is designed to clean up malformed tool names that may have
// missing or duplicated prefixes. The canonical form is defined as "tool.<base_name>".
func CanonicalizeToolName(name string) string {
	// Start with a clean slate by lowercasing to handle case variations like "TOOL."
	processedName := strings.ToLower(name)

	// Repeatedly remove the prefix until it's gone to handle cases like "tool.tool.fs.read"
	for strings.HasPrefix(processedName, toolPrefix) {
		processedName = strings.TrimPrefix(processedName, toolPrefix)
	}

	// Add the single, correct prefix back to the base name.
	return toolPrefix + processedName
}
