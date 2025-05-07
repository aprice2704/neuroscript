// NeuroScript Version: 0.3.1
// File version: 0.1.1 // Added ReadFile, WriteFile, DeleteFile, ListDirectory, Mkdir, MoveFile
// Defines ToolImplementation structs for selected File System tools.
// filename: pkg/core/tooldefs_fs.go

package core

// fstoolsToRegister contains ToolImplementation definitions for a subset of File System tools.
// This array is intended to be concatenated with other similar arrays in a central
// registrar (e.g., zz_core_tools_registrar.go) to be processed by AddToolImplementations.
//
// It's crucial that if a tool is listed here, its original init() based registration
// (if any, in its own file like tools_fs_read.go) is removed to avoid double registration.
var fstoolsToRegister = []ToolImplementation{
	{
		Spec: ToolSpec{
			Name:        "FileHash",
			Description: "Calculates the SHA256 hash of a specified file within the sandbox. Returns the hex-encoded hash string.",
			Args: []ArgSpec{
				{Name: "filepath", Type: ArgTypeString, Required: true, Description: "The relative path (within the sandbox) of the file to hash."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolFileHash, // Assumes toolFileHash is defined in pkg/core/tools_fs_hash.go
	},
	{
		Spec: ToolSpec{
			Name:        "FileStat",
			Description: "Gets information about a file or directory within the sandbox. Returns a map with 'name', 'path', 'size', 'is_dir', 'mod_time', or nil if not found.",
			Args: []ArgSpec{
				{Name: "path", Type: ArgTypeString, Required: true, Description: "The relative path to the file or directory to stat."},
			},
			ReturnType: ArgTypeMap, // Returns a map or nil (nil indicates not found, compatible with map type for script)
		},
		Func: toolStat, // Assumes toolStat is defined in pkg/core/tools_fs_stat.go
	},
	{
		Spec: ToolSpec{
			Name:        "WalkDir",
			Description: "Recursively walks a directory within the sandbox and returns a list of maps containing file/directory information (name, path, isDir, size, modTime).",
			Args: []ArgSpec{
				{Name: "path", Type: ArgTypeString, Required: true, Description: "The relative path to the directory to walk."},
			},
			ReturnType: ArgTypeSliceMap, // Returns []map[string]interface{}
		},
		Func: toolWalkDir, // Assumes toolWalkDir is defined in pkg/core/tools_fs_walk.go
	},
	{ // Added ReadFile
		Spec: ToolSpec{
			Name:        "ReadFile",
			Description: "Reads the entire content of a file within the sandbox. Returns the content as a string.",
			Args: []ArgSpec{
				{Name: "filepath", Type: ArgTypeString, Required: true, Description: "The relative path to the file within the sandbox."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolReadFile, // Assumes toolReadFile is defined in pkg/core/tools_fs_read.go
	},
	{ // Added WriteFile
		Spec: ToolSpec{
			Name:        "WriteFile",
			Description: "Writes content to a specified file within the sandbox. Creates the file if it doesn't exist, overwrites if it does.",
			Args: []ArgSpec{
				{Name: "filepath", Type: ArgTypeString, Required: true, Description: "The relative path to the file within the sandbox."},
				{Name: "content", Type: ArgTypeString, Required: true, Description: "The string content to write to the file."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolWriteFile, // Assumes toolWriteFile is defined in pkg/core/tools_fs_write.go
	},
	{ // Added DeleteFile (using existing toolDeleteFileImpl structure)
		Spec: ToolSpec{
			Name:        "DeleteFile",
			Description: "Deletes a file or an empty directory within the sandbox.",
			Args: []ArgSpec{
				{Name: "path", Type: ArgTypeString, Required: true, Description: "The relative path to the file or empty directory to delete."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolDeleteFile, // Assumes toolDeleteFile is defined in pkg/core/tools_fs_delete.go
	},
	{ // Added ListDirectory (using existing toolListDirectoryImpl structure)
		Spec: ToolSpec{
			Name: "ListDirectory",
			Description: "Lists the contents (files and subdirectories) of a specified directory. " +
				"Returns a list of maps, each containing 'name', 'path', 'isDir', 'size', 'modTime'.",
			Args: []ArgSpec{
				{Name: "path", Type: ArgTypeString, Required: true, Description: "The relative path to the directory to list."},
				{Name: "recursive", Type: ArgTypeBool, Required: false, Description: "If true, list contents recursively. Defaults to false."},
			},
			ReturnType: ArgTypeSliceMap,
		},
		Func: toolListDirectory, // Assumes toolListDirectory is defined in pkg/core/tools_fs_dirs.go
	},
	{ // Added Mkdir (using existing toolMkdirImpl structure)
		Spec: ToolSpec{
			Name:        "Mkdir",
			Description: "Creates a new directory (including any necessary parents) within the sandbox.",
			Args: []ArgSpec{
				{Name: "path", Type: ArgTypeString, Required: true, Description: "The relative path of the directory to create."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolMkdir, // Assumes toolMkdir is defined in pkg/core/tools_fs_dirs.go
	},
	{ // Added MoveFile (using existing toolMoveFileImpl structure)
		Spec: ToolSpec{
			Name:        "MoveFile",
			Description: "Moves or renames a file or directory within the sandbox.",
			Args: []ArgSpec{
				{Name: "source", Type: ArgTypeString, Required: true, Description: "The current relative path to the file/directory within the sandbox."},
				{Name: "destination", Type: ArgTypeString, Required: true, Description: "The desired new relative path for the file/directory within the sandbox."},
			},
			ReturnType: ArgTypeMap,
		},
		Func: toolMoveFile, // Assumes toolMoveFile is defined in pkg/core/tools_fs_move.go
	},
}
