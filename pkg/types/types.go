// Types holds some shared type definitions
package types

import (
	"fmt"
	"strings"
)

// For dealing with tools
type ToolName string  // a tool's individual name -- no prefix or group
type FullName string  // a tool's full name, used for finding it in the registry
type ToolGroup string // the group of tools it is in e.g. fs etc.

const (
	toolPrefix = "tool"
	toolSep    = "."
)

// MakeFullName creates a canonical tool name from a group and a name.
// It has a single return value for convenience. It will panic if the group or
// name is empty, or if the name contains invalid characters, as this
// represents a programmer error that should be caught early.
func MakeFullName(group string, name string) FullName {
	if group == "" || name == "" {
		panic("MakeFullName: tool group and name cannot be empty")
	}

	// Prevent the separator from appearing in the name part to ensure the
	// full name remains parsable. Groups can contain dots.
	if strings.Contains(name, toolSep) {
		panic(fmt.Sprintf(
			"MakeFullName: tool name '%s' cannot contain a dot separator",
			name,
		))
	}

	fullName := strings.Join([]string{toolPrefix, group, name}, toolSep)
	return FullName(fullName)
}

// MakeFullNameTyped is the typed convenience wrapper for MakeFullName.
func MakeFullNameTyped(group ToolGroup, name ToolName) FullName {
	return MakeFullName(string(group), string(name))
}
