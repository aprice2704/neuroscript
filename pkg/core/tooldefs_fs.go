// NeuroScript Version: 0.3.1
// File version: 0.0.4 // Consolidated all FS tool definitions as literals.
// nlines: 159
// risk_rating: MEDIUM
// filename: pkg/core/tooldefs_fs.go
package core

import (
	"fmt"
	"time"
) // Required for time format constants used in some tool descriptions

var fsToolsToRegister = []ToolImplementation{
	ToolImplementation{
		Spec: ToolSpec{
			Name:        "FileHash",
			Description: "Calculates the SHA256 hash of a specified file. Returns the hex-encoded hash string.",
			Args: []ArgSpec{
				{Name: "filepath", Type: ArgTypeString, Required: true, Description: "Relative path (within the sandbox) of the file to hash."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolFileHash, // from tools_fs_hash.go
	},
	// --- LineCountFile Tool ---
	ToolImplementation{
		Spec: ToolSpec{
			Name:        "LineCountFile",
			Description: "Counts lines in a specified file. Returns line count as an integer.",
			Args: []ArgSpec{
				{Name: "filepath", Type: ArgTypeString, Required: true, Description: "Relative path to the file."},
			},
			ReturnType: ArgTypeInt,
		},
		Func: toolLineCountFile, // from tools_fs_utils.go
	},
	// --- SanitizeFilename Tool ---
	ToolImplementation{
		Spec: ToolSpec{
			Name:        "SanitizeFilename",
			Description: "Cleans a string to make it suitable for use as part of a filename.",
			Args: []ArgSpec{
				{Name: "name", Type: ArgTypeString, Required: true, Description: "The string to sanitize."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolSanitizeFilename, // from tools_fs_utils.go
	},
	// --- ReadFile Tool ---
	ToolImplementation{
		Spec: ToolSpec{
			Name:        "ReadFile",
			Description: "Reads the entire content of a specific file. Returns the content as a string.",
			Args: []ArgSpec{
				{Name: "filepath", Type: ArgTypeString, Required: true, Description: "Relative path to the file."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolReadFile, // from tools_fs_read.go
	},
	// --- WriteFile Tool ---
	ToolImplementation{
		Spec: ToolSpec{
			Name:        "WriteFile",
			Description: "Writes content to a specific file. Creates parent directories if needed. Returns 'OK' on success.",
			Args: []ArgSpec{
				{Name: "filepath", Type: ArgTypeString, Required: true, Description: "Relative path to the file."},
				{Name: "content", Type: ArgTypeString, Required: true, Description: "The content to write."},
			},
			ReturnType: ArgTypeString, // Returns "OK"
		},
		Func: toolWriteFile, // from tools_fs_write.go
	},
	// --- ListDirectory Tool ---
	ToolImplementation{
		Spec: ToolSpec{
			Name:        "ListDirectory",
			Description: "Lists files and subdirectories at a given path. Returns a list of maps, each describing an entry (keys: name, path, isDir, size, modTime).",
			Args: []ArgSpec{
				{Name: "path", Type: ArgTypeString, Required: true, Description: "Relative path to the directory (use '.' for current)."},
				{Name: "recursive", Type: ArgTypeBool, Required: false, Description: "Whether to list recursively (default: false)."},
			},
			ReturnType: ArgTypeSliceAny, // Returns []map[string]interface{}
		},
		Func: toolListDirectory, // from tools_fs_dirs.go
	},
	// --- Mkdir Tool ---
	ToolImplementation{
		Spec: ToolSpec{
			Name:        "Mkdir",
			Description: "Creates a directory. Parent directories are created if they do not exist (like mkdir -p). Returns a success message.",
			Args: []ArgSpec{
				{Name: "path", Type: ArgTypeString, Required: true, Description: "Relative path of the directory to create."},
			},
			ReturnType: ArgTypeString, // Returns a success message.
		},
		Func: toolMkdir, // from tools_fs_dirs.go
	},
	// --- DeleteFile Tool ---
	ToolImplementation{
		Spec: ToolSpec{
			Name:        "DeleteFile",
			Description: "Deletes a file or an empty directory. Returns 'OK' on success or if path doesn't exist.",
			Args: []ArgSpec{
				{Name: "path", Type: ArgTypeString, Required: true, Description: "Relative path to the file or empty directory to delete."},
			},
			ReturnType: ArgTypeString, // Returns "OK"
		},
		Func: toolDeleteFile, // from tools_fs_delete.go
	},
	// --- StatPath Tool ---
	ToolImplementation{
		Spec: ToolSpec{
			Name:        "StatPath",
			Description: fmt.Sprintf("Gets information about a file or directory. Returns a map containing: name(string), path(string), size_bytes(int), is_dir(bool), modified_unix(int), modified_rfc3339(string - format %s), mode_string(string), mode_perm(string).", time.RFC3339Nano),
			Args: []ArgSpec{
				{Name: "path", Type: ArgTypeString, Required: true, Description: "Relative path to the file or directory."},
			},
			ReturnType: ArgTypeMap,
		},
		Func: toolStat, // from tools_fs_stat.go
	},
	// --- MoveFile Tool ---
	ToolImplementation{
		Spec: ToolSpec{
			Name:        "MoveFile",
			Description: "Moves or renames a file or directory within the sandbox. Returns a map: {'message': 'success message', 'error': nil} on success.",
			Args: []ArgSpec{
				{Name: "source_path", Type: ArgTypeString, Required: true, Description: "Relative path of the source file/directory."},
				{Name: "destination_path", Type: ArgTypeString, Required: true, Description: "Relative path of the destination."},
			},
			ReturnType: ArgTypeMap, // Returns map[string]interface{}
		},
		Func: toolMoveFile, // from tools_fs_move.go
	},
	// --- WalkDir Tool ---
	ToolImplementation{
		Spec: ToolSpec{
			Name:        "WalkDir",
			Description: fmt.Sprintf("Recursively walks a directory, returning a list of maps describing files/subdirectories found (keys: name, path_relative, is_dir, size_bytes, modified_unix, modified_rfc3339 (format %s), mode_string). Skips the root directory itself.", time.RFC3339Nano),
			Args: []ArgSpec{
				{Name: "path", Type: ArgTypeString, Required: true, Description: "Relative path to the directory to walk."},
			},
			ReturnType: ArgTypeSliceAny, // Returns []map[string]interface{}
		},
		Func: toolWalkDir, // from tools_fs_walk.go
	},
}
