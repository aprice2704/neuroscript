// Types holds some shared type definitions

package types

// For dealing with tools

type ToolName string  // a tool's individual name -- no prefix or group
type FullName string  // a tool's full name, used for finding it in the registry
type ToolGroup string // the group of tools it is in e.g. fs etc.

const (
	toolPrefix = "tool"
	toolSep    = "."
)

// The only correct way to make a toolname
// Just return a single ToolName (no error) so it can be used as fn arg easily
func MakeFullName(group string, name string) (fullname FullName) {

	if len(group) == 0 {
		//		lang.Check(lang.ErrInvalidToolGroup) // give runtime a chance to bail
		return ""
	}
	if len(name) == 0 {
		//		lang.Check(lang.ErrInvalidToolName)
		return ""
	}

	return FullName(toolPrefix + toolSep + string(group) + toolSep + name)

}
