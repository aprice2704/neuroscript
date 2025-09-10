// NeuroScript Version: 0.5.2
// File version: 9
// Purpose: Updated all filesystem tools to require trust (`RequiresTrust: true`) for enhanced security, except for pure functions.
// nlines: 300 // Approximate
// risk_rating: HIGH
// filename: pkg/tool/fs/tooldefs_fs.go
package fs

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/tool"
) // Required for time format constants used in some tool descriptions

const Group = "fs"

var FsToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "Hash",
			Group:       Group,
			Description: "Calculates the SHA256 hash of a specified file. Returns the hex-encoded hash string.",
			Category:    "Filesystem",
			Args: []tool.ArgSpec{
				{Name: "filepath", Type: tool.ArgTypeString, Required: true, Description: "Relative path (within the sandbox) of the file to hash."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a hex-encoded SHA256 hash string of the file's content. Returns an empty string on error.",
			Example:         `TOOL.FS.Hash(filepath: "data/my_document.txt")`,
			ErrorConditions: "ErrArgumentMismatch if filepath is empty; ErrConfiguration if sandbox is not set; ErrSecurityPath for invalid paths; ErrFileNotFound; ErrPermissionDenied; ErrPathNotFile if path is a directory; ErrIOFailed for I/O errors.",
		},
		Func:          toolFileHash,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"read"}},
		},
		Effects: []string{"readsFS"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "LineCount",
			Group:       Group,
			Description: "Counts lines in a specified file. Returns line count as an integer.",
			Category:    "Filesystem",
			Args: []tool.ArgSpec{
				{Name: "filepath", Type: tool.ArgTypeString, Required: true, Description: "Relative path to the file."},
			},
			ReturnType:      tool.ArgTypeInt,
			ReturnHelp:      "Returns the number of lines in the specified file. Returns 0 on error or if file is empty.",
			Example:         `TOOL.FS.LineCount(filepath: "logs/app.log")`,
			ErrorConditions: "ErrArgumentMismatch if filepath is empty; ErrConfiguration if sandbox is not set; ErrSecurityPath for invalid paths; ErrFileNotFound; ErrPermissionDenied; ErrPathNotFile if path is a directory; ErrIOFailed for read errors.",
		},
		Func:          toolLineCountFile,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"read"}},
		},
		Effects: []string{"readsFS"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "SanitizeFilename",
			Group:       Group,
			Description: "Cleans a string to make it suitable for use as part of a filename.",
			Category:    "Filesystem Utilities",
			Args: []tool.ArgSpec{
				{Name: "name", Type: tool.ArgTypeString, Required: true, Description: "The string to sanitize."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a sanitized string suitable for use as a filename component.",
			Example:         `TOOL.FS.SanitizeFilename(name: "My Report Final?.docx")`,
			ErrorConditions: "ErrArgumentMismatch if name is not provided or not a string.",
		},
		Func:          toolSanitizeFilename,
		RequiresTrust: false,
		RequiredCaps:  nil,
		Effects:       []string{"idempotent"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Read",
			Group:       Group,
			Description: "Reads the entire content of a specific file. Returns the content as a string.",
			Category:    "Filesystem",
			Args: []tool.ArgSpec{
				{Name: "filepath", Type: tool.ArgTypeString, Required: true, Description: "Relative path to the file."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the content of the file as a string. Returns an empty string on error.",
			Example:         `TOOL.FS.Read(filepath: "config.txt")`,
			ErrorConditions: "ErrArgumentMismatch if filepath is empty; ErrConfiguration if sandbox is not set; ErrSecurityPath for invalid paths; ErrFileNotFound; ErrPermissionDenied; ErrPathNotFile if path is a directory; ErrIOFailed for other I/O errors.",
		},
		Func:          toolReadFile,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"read"}},
		},
		Effects: []string{"readsFS"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Write",
			Group:       Group,
			Description: "Writes content to a specific file, overwriting it if it exists. Creates parent directories if needed. Returns 'OK' on success.",
			Category:    "Filesystem",
			Args: []tool.ArgSpec{
				{Name: "filepath", Type: tool.ArgTypeString, Required: true, Description: "Relative path to the file."},
				{Name: "content", Type: tool.ArgTypeString, Required: true, Description: "The content to write."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns 'OK' on success. Returns nil on error.",
			Example:         `TOOL.FS.Write(filepath: "output/data.json", content: "{\"key\":\"value\"}")`,
			ErrorConditions: "ErrArgumentMismatch; ErrConfiguration; ErrSecurityPath; ErrCannotCreateDir; ErrPermissionDenied; ErrPathNotFile; ErrIOFailed.",
		},
		Func:          toolWriteFile,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"write"}},
		},
		Effects: []string{"writesFS", "idempotent"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Append",
			Group:       Group,
			Description: "Appends content to a specific file. Creates the file and parent directories if needed. Returns 'OK' on success.",
			Category:    "Filesystem",
			Args: []tool.ArgSpec{
				{Name: "filepath", Type: tool.ArgTypeString, Required: true, Description: "Relative path to the file."},
				{Name: "content", Type: tool.ArgTypeString, Required: true, Description: "The content to append."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns 'OK' on success. Returns nil on error.",
			Example:         `TOOL.FS.Append(filepath: "logs/activity.log", content: "User logged in.\n")`,
			ErrorConditions: "ErrArgumentMismatch; ErrConfiguration; ErrSecurityPath; ErrCannotCreateDir; ErrPermissionDenied; ErrPathNotFile; ErrIOFailed.",
		},
		Func:          toolAppendFile,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"write"}},
		},
		Effects: []string{"writesFS"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "List",
			Group:       Group,
			Description: "Lists files and subdirectories at a given path. Returns a list of maps, each describing an entry.",
			Category:    "Filesystem",
			Args: []tool.ArgSpec{
				{Name: "path", Type: tool.ArgTypeString, Required: true, Description: "Relative path to the directory (use '.' for current)."},
				{Name: "recursive", Type: tool.ArgTypeBool, Required: false, Description: "Whether to list recursively (default: false)."},
			},
			ReturnType:      tool.ArgTypeSliceAny,
			ReturnHelp:      "Returns a slice of maps detailing files/directories. Returns nil on error.",
			Example:         `TOOL.FS.List(path: "mydir", recursive: true)`,
			ErrorConditions: "ErrArgumentMismatch; ErrConfiguration; ErrSecurityPath; ErrFileNotFound; ErrPermissionDenied; ErrPathNotDirectory; ErrIOFailed.",
		},
		Func:          toolListDirectory,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"read"}},
		},
		Effects: []string{"readsFS", "idempotent"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Mkdir",
			Group:       Group,
			Description: "Creates a directory (like mkdir -p). Returns a success message.",
			Category:    "Filesystem",
			Args: []tool.ArgSpec{
				{Name: "path", Type: tool.ArgTypeString, Required: true, Description: "Relative path of the directory to create."},
			},
			ReturnType:      tool.ArgTypeMap,
			ReturnHelp:      "Returns a map indicating success. Returns nil on error.",
			Example:         `TOOL.FS.Mkdir(path: "new/subdir")`,
			ErrorConditions: "ErrArgumentMismatch; ErrConfiguration; ErrSecurityPath; ErrPathNotDirectory; ErrPathExists; ErrPermissionDenied; ErrIOFailed; ErrCannotCreateDir.",
		},
		Func:          toolMkdir,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"write"}},
		},
		Effects: []string{"writesFS", "idempotent"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Delete",
			Group:       Group,
			Description: "Deletes a file or an empty directory. Returns 'OK' on success or if path doesn't exist.",
			Category:    "Filesystem",
			Args: []tool.ArgSpec{
				{Name: "path", Type: tool.ArgTypeString, Required: true, Description: "Relative path to the file or empty directory to delete."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns 'OK' on success. Returns nil on error.",
			Example:         `TOOL.FS.Delete(path: "temp/old_file.txt")`,
			ErrorConditions: "ErrArgumentMismatch; ErrConfiguration; ErrSecurityPath; ErrPreconditionFailed if directory is not empty; ErrPermissionDenied; ErrIOFailed.",
		},
		Func:          toolDeleteFile,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"delete"}},
		},
		Effects: []string{"writesFS", "idempotent"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Stat",
			Group:       Group,
			Description: fmt.Sprintf("Gets information about a file or directory. Returns a map of file info."),
			Category:    "Filesystem",
			Args: []tool.ArgSpec{
				{Name: "path", Type: tool.ArgTypeString, Required: true, Description: "Relative path to the file or directory."},
			},
			ReturnType:      tool.ArgTypeMap,
			ReturnHelp:      "Returns a map with file/directory info. Returns nil on error.",
			Example:         `TOOL.FS.Stat(path: "my_file.go")`,
			ErrorConditions: "ErrArgumentMismatch; ErrConfiguration; ErrSecurityPath; ErrFileNotFound; ErrPermissionDenied; ErrIOFailed.",
		},
		Func:          toolStat,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"read"}},
		},
		Effects: []string{"readsFS", "idempotent"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Move",
			Group:       Group,
			Description: "Moves or renames a file or directory within the sandbox.",
			Category:    "Filesystem",
			Args: []tool.ArgSpec{
				{Name: "source_path", Type: tool.ArgTypeString, Required: true, Description: "Relative path of the source file/directory."},
				{Name: "destination_path", Type: tool.ArgTypeString, Required: true, Description: "Relative path of the destination."},
			},
			ReturnType:      tool.ArgTypeMap,
			ReturnHelp:      "Returns a map indicating success. Returns nil on error.",
			Example:         `TOOL.FS.Move(source_path: "old_name.txt", destination_path: "new_name.txt")`,
			ErrorConditions: "ErrArgumentMismatch; ErrConfiguration; ErrSecurityPath; ErrFileNotFound; ErrPathExists; ErrPermissionDenied; ErrIOFailed.",
		},
		Func:          toolMoveFile,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"write", "delete"}},
		},
		Effects: []string{"writesFS", "idempotent"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Walk",
			Group:       Group,
			Description: fmt.Sprintf("Recursively walks a directory, returning a list of maps describing files/subdirectories found."),
			Category:    "Filesystem",
			Args: []tool.ArgSpec{
				{Name: "path", Type: tool.ArgTypeString, Required: true, Description: "Relative path to the directory to walk."},
			},
			ReturnType:      tool.ArgTypeSliceAny,
			ReturnHelp:      "Returns a slice of maps, each describing a file/subdir. Skips the root dir itself. Returns nil on error.",
			Example:         `TOOL.FS.Walk(path: "src")`,
			ErrorConditions: "ErrArgumentMismatch; ErrConfiguration; ErrSecurityPath; ErrFileNotFound; ErrPathNotDirectory; ErrPermissionDenied; ErrIOFailed; ErrInternal.",
		},
		Func:          toolWalkDir,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"read"}},
		},
		Effects: []string{"readsFS", "idempotent"},
	},
}
