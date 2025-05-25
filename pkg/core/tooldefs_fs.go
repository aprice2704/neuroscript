// NeuroScript Version: 0.3.1
// File version: 0.0.5 // Populated Category, Example, ReturnHelp, ErrorConditions fields.
// nlines: 250 // Approximate
// risk_rating: MEDIUM
// filename: pkg/core/tooldefs_fs.go
package core

import (
	"fmt"
	"time"
) // Required for time format constants used in some tool descriptions

var fsToolsToRegister = []ToolImplementation{
	{
		Spec: ToolSpec{
			Name:        "FS.Hash",
			Description: "Calculates the SHA256 hash of a specified file. Returns the hex-encoded hash string.",
			Category:    "Filesystem",
			Args: []ArgSpec{
				{Name: "filepath", Type: ArgTypeString, Required: true, Description: "Relative path (within the sandbox) of the file to hash."},
			},
			ReturnType:      ArgTypeString,
			ReturnHelp:      "Returns a hex-encoded SHA256 hash string of the file's content. Returns an empty string on error.",
			Example:         `TOOL.FS.Hash(filepath: "data/my_document.txt") // Returns "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" (example for an empty file)`,
			ErrorConditions: "ErrArgumentMismatch if filepath is empty; ErrConfiguration if sandbox is not set; ErrSecurityPath (from SecureFilePath) for invalid paths; ErrFileNotFound if file does not exist; ErrPermissionDenied if file cannot be opened; ErrPathNotFile if path is a directory; ErrIOFailed for other I/O errors during open or read.",
		},
		Func: toolFileHash, // from tools_fs_hash.go
	},
	{
		Spec: ToolSpec{
			Name:        "FS.LineCount",
			Description: "Counts lines in a specified file. Returns line count as an integer.",
			Category:    "Filesystem",
			Args: []ArgSpec{
				{Name: "filepath", Type: ArgTypeString, Required: true, Description: "Relative path to the file."},
			},
			ReturnType:      ArgTypeInt,
			ReturnHelp:      "Returns the number of lines in the specified file. Returns 0 on error or if file is empty.",
			Example:         `TOOL.FS.LineCount(filepath: "logs/app.log") // Returns 150 (example)`,
			ErrorConditions: "ErrArgumentMismatch if filepath is empty; ErrConfiguration if sandbox is not set; ErrSecurityPath for invalid paths; ErrFileNotFound; ErrPermissionDenied; ErrPathNotFile if path is a directory; ErrIOFailed for read errors. (Based on typical file tool error handling, actual implementation for toolLineCountFile in tools_fs_utils.go needed for exact errors).",
		},
		Func: toolLineCountFile, // from tools_fs_utils.go
	},
	{
		Spec: ToolSpec{
			Name:        "FS.SanitizeFilename",
			Description: "Cleans a string to make it suitable for use as part of a filename.",
			Category:    "Filesystem Utilities",
			Args: []ArgSpec{
				{Name: "name", Type: ArgTypeString, Required: true, Description: "The string to sanitize."},
			},
			ReturnType:      ArgTypeString,
			ReturnHelp:      "Returns a sanitized string suitable for use as a filename component (e.g., replacing unsafe characters with underscores).",
			Example:         `TOOL.FS.SanitizeFilename(name: "My Report Final?.docx") // Returns "My_Report_Final_.docx" (example)`,
			ErrorConditions: "ErrArgumentMismatch if name is not provided or not a string. (Based on typical utility tool error handling, actual implementation for toolSanitizeFilename in tools_fs_utils.go needed for exact errors).",
		},
		Func: toolSanitizeFilename, // from tools_fs_utils.go
	},
	{
		Spec: ToolSpec{
			Name:        "FS.Read",
			Description: "Reads the entire content of a specific file. Returns the content as a string.",
			Category:    "Filesystem",
			Args: []ArgSpec{
				{Name: "filepath", Type: ArgTypeString, Required: true, Description: "Relative path to the file."},
			},
			ReturnType:      ArgTypeString,
			ReturnHelp:      "Returns the content of the file as a string. Returns an empty string on error.",
			Example:         `TOOL.FS.Read(filepath: "config.txt") // Returns "setting=value\n..."`,
			ErrorConditions: "ErrArgumentMismatch if filepath is empty; ErrConfiguration if sandbox is not set; ErrSecurityPath (from ResolveAndSecurePath) for invalid paths; ErrFileNotFound if file does not exist; ErrPermissionDenied; ErrPathNotFile if path is a directory; ErrIOFailed for other I/O errors.",
		},
		Func: toolReadFile, // from tools_fs_read.go
	},
	{
		Spec: ToolSpec{
			Name:        "FS.Write",
			Description: "Writes content to a specific file. Creates parent directories if needed. Returns 'OK' on success.",
			Category:    "Filesystem",
			Args: []ArgSpec{
				{Name: "filepath", Type: ArgTypeString, Required: true, Description: "Relative path to the file."},
				{Name: "content", Type: ArgTypeString, Required: true, Description: "The content to write."},
			},
			ReturnType:      ArgTypeString, // Returns "Successfully wrote X bytes to Y"
			ReturnHelp:      "Returns a success message string like 'Successfully wrote X bytes to Y' on success. Returns an empty string on error.",
			Example:         `TOOL.FS.Write(filepath: "output/data.json", content: "{\"key\":\"value\"}") // Returns "Successfully wrote 15 bytes to output/data.json"`,
			ErrorConditions: "ErrArgumentMismatch if filepath is empty or content is not string/nil; ErrConfiguration if sandbox is not set; ErrSecurityPath (from ResolveAndSecurePath) for invalid paths; ErrCannotCreateDir if parent directories cannot be created; ErrPermissionDenied if writing is not allowed; ErrPathNotFile if path exists and is a directory; ErrIOFailed for other I/O errors.",
		},
		Func: toolWriteFile, // from tools_fs_write.go
	},
	{
		Spec: ToolSpec{
			Name:        "FS.List",
			Description: "Lists files and subdirectories at a given path. Returns a list of maps, each describing an entry (keys: name, path, isDir, size, modTime).",
			Category:    "Filesystem",
			Args: []ArgSpec{
				{Name: "path", Type: ArgTypeString, Required: true, Description: "Relative path to the directory (use '.' for current)."},
				{Name: "recursive", Type: ArgTypeBool, Required: false, Description: "Whether to list recursively (default: false)."},
			},
			ReturnType:      ArgTypeSliceAny, // Returns []map[string]interface{}
			ReturnHelp:      "Returns a slice of maps. Each map details a file/directory: {'name':string, 'path':string (relative to input path for recursive), 'isDir':bool, 'size':int64, 'modTime':string (RFC3339Nano)}. Returns nil on error.",
			Example:         `TOOL.FS.List(path: "mydir", recursive: true)`,
			ErrorConditions: "ErrArgumentMismatch if path is not a string or recursive is not bool/nil; ErrConfiguration if sandbox is not set; ErrSecurityPath (from ResolveAndSecurePath) for invalid path; ErrFileNotFound if path does not exist; ErrPermissionDenied; ErrPathNotDirectory if path is not a directory; ErrIOFailed for other I/O errors during listing or walking.",
		},
		Func: toolListDirectory, // from tools_fs_dirs.go
	},
	{
		Spec: ToolSpec{
			Name:        "FS.Mkdir",
			Description: "Creates a directory. Parent directories are created if they do not exist (like mkdir -p). Returns a success message.",
			Category:    "Filesystem",
			Args: []ArgSpec{
				{Name: "path", Type: ArgTypeString, Required: true, Description: "Relative path of the directory to create."},
			},
			ReturnType:      ArgTypeMap, // Returns map[string]interface{}{"status":"success", "message": "...", "path": "..."}
			ReturnHelp:      "Returns a map: {'status':'success', 'message':'Successfully created directory: <path>', 'path':'<path>'} on success. Returns nil on error.",
			Example:         `TOOL.FS.Mkdir(path: "new/subdir") // Returns {"status":"success", "message":"Successfully created directory: new/subdir", "path":"new/subdir"}`,
			ErrorConditions: "ErrArgumentMismatch if path is empty, '.', or not a string; ErrConfiguration if sandbox is not set; ErrSecurityPath (from ResolveAndSecurePath) for invalid path; ErrPathNotDirectory if path exists and is a file; ErrPathExists if directory already exists; ErrPermissionDenied; ErrIOFailed for other I/O errors or failure to stat; ErrCannotCreateDir if MkdirAll fails.",
		},
		Func: toolMkdir, // from tools_fs_dirs.go
	},
	{
		Spec: ToolSpec{
			Name:        "FS.Delete",
			Description: "Deletes a file or an empty directory. Returns 'OK' on success or if path doesn't exist.",
			Category:    "Filesystem",
			Args: []ArgSpec{
				{Name: "path", Type: ArgTypeString, Required: true, Description: "Relative path to the file or empty directory to delete."},
			},
			ReturnType:      ArgTypeString, // Returns "OK"
			ReturnHelp:      "Returns the string 'OK' on successful deletion or if the path does not exist. Returns nil on error.",
			Example:         `TOOL.FS.Delete(path: "temp/old_file.txt") // Returns "OK"`,
			ErrorConditions: "ErrArgumentMismatch if path is empty or not a string; ErrConfiguration if sandbox is not set; ErrSecurityPath (from SecureFilePath) for invalid path; ErrPreconditionFailed if directory is not empty; ErrPermissionDenied; ErrIOFailed for other I/O errors. Path not found is treated as success.",
		},
		Func: toolDeleteFile, // from tools_fs_delete.go
	},
	{
		Spec: ToolSpec{
			Name:        "FS.Stat",
			Description: fmt.Sprintf("Gets information about a file or directory. Returns a map containing: name(string), path(string), size_bytes(int), is_dir(bool), modified_unix(int), modified_rfc3339(string - format %s), mode_string(string), mode_perm(string).", time.RFC3339Nano),
			Category:    "Filesystem",
			Args: []ArgSpec{
				{Name: "path", Type: ArgTypeString, Required: true, Description: "Relative path to the file or directory."},
			},
			ReturnType:      ArgTypeMap,
			ReturnHelp:      "Returns a map with file/directory info: {'name', 'path', 'size_bytes', 'is_dir', 'modified_unix', 'modified_rfc3339', 'mode_string', 'mode_perm'}. Returns nil on error.",
			Example:         `TOOL.FS.Stat(path: "my_file.go")`,
			ErrorConditions: "ErrArgumentMismatch if path is empty or not a string; ErrConfiguration if sandbox is not set; ErrSecurityPath (from ResolveAndSecurePath) for invalid path; ErrFileNotFound if path does not exist; ErrPermissionDenied; ErrIOFailed for other I/O errors.",
		},
		Func: toolStat, // from tools_fs_stat.go
	},
	{
		Spec: ToolSpec{
			Name:        "FS.Move",
			Description: "Moves or renames a file or directory within the sandbox. Returns a map: {'message': 'success message', 'error': nil} on success.",
			Category:    "Filesystem",
			Args: []ArgSpec{
				{Name: "source_path", Type: ArgTypeString, Required: true, Description: "Relative path of the source file/directory."},
				{Name: "destination_path", Type: ArgTypeString, Required: true, Description: "Relative path of the destination."},
			},
			ReturnType:      ArgTypeMap, // Returns map[string]interface{}
			ReturnHelp:      "Returns a map {'message': 'success message', 'error': nil} on success. Returns nil on error.",
			Example:         `TOOL.FS.Move(source_path: "old_name.txt", destination_path: "new_name.txt")`,
			ErrorConditions: "ErrArgumentMismatch if paths are empty, not strings, or are the same; ErrConfiguration if sandbox is not set; ErrSecurityPath (from SecureFilePath) for invalid source or destination paths; ErrFileNotFound if source path does not exist; ErrPathExists if destination path already exists; ErrPermissionDenied for source or destination; ErrIOFailed for other I/O errors during stat or rename.",
		},
		Func: toolMoveFile, // from tools_fs_move.go
	},
	{
		Spec: ToolSpec{
			Name:        "FS.Walk",
			Description: fmt.Sprintf("Recursively walks a directory, returning a list of maps describing files/subdirectories found (keys: name, path_relative, is_dir, size_bytes, modified_unix, modified_rfc3339 (format %s), mode_string). Skips the root directory itself.", time.RFC3339Nano),
			Category:    "Filesystem",
			Args: []ArgSpec{
				{Name: "path", Type: ArgTypeString, Required: true, Description: "Relative path to the directory to walk."},
			},
			ReturnType:      ArgTypeSliceAny, // Returns []map[string]interface{}
			ReturnHelp:      "Returns a slice of maps, each describing a file/subdir: {'name', 'path_relative', 'is_dir', 'size_bytes', 'modified_unix', 'modified_rfc3339', 'mode_string'}. Skips the root dir itself. Returns nil on error.",
			Example:         `TOOL.FS.Walk(path: "src")`,
			ErrorConditions: "ErrArgumentMismatch if path is empty or not a string; ErrConfiguration if sandbox is not set; ErrSecurityPath (from ResolveAndSecurePath) for invalid path; ErrFileNotFound if start path not found; ErrPathNotDirectory if start path is not a directory; ErrPermissionDenied for start path; ErrIOFailed for stat errors or errors during walk; ErrInternal if relative path calculation fails during walk.",
		},
		Func: toolWalkDir, // from tools_fs_walk.go
	},
}
