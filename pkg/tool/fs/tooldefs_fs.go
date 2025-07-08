// NeuroScript Version: 0.4.0
// File version: 7
// Purpose: Corrected all instances of parser.ArgSpec to the correct tool.ArgSpec.
// nlines: 275 // Approximate
// risk_rating: MEDIUM
// filename: pkg/tool/fs/tooldefs_fs.go
package fs

import (
	"fmt"
	"time"

	"github.com/aprice2704/neuroscript/pkg/tool"
) // Required for time format constants used in some tool descriptions

const group = "fs"

var fsToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "Hash",
			Group:       group,
			Description: "Calculates the SHA256 hash of a specified file. Returns the hex-encoded hash string.",
			Category:    "Filesystem",
			Args: []tool.ArgSpec{
				{Name: "filepath", Type: tool.ArgTypeString, Required: true, Description: "Relative path (within the sandbox) of the file to hash."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a hex-encoded SHA256 hash string of the file's content. Returns an empty string on error.",
			Example:         `TOOL.FS.Hash(filepath: "data/my_document.txt") // Returns "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" (example for an empty file)`,
			ErrorConditions: "ErrArgumentMismatch if filepath is empty; ErrConfiguration if sandbox is not set; ErrSecurityPath (from SecureFilePath) for invalid paths; ErrFileNotFound if file does not exist; ErrPermissionDenied if file cannot be opened; ErrPathNotFile if path is a directory; ErrIOFailed for other I/O errors during open or read.",
		},
		Func: toolFileHash, // from tools_fs_hash.go
	},
	{
		Spec: tool.ToolSpec{
			Name:        "LineCount",
			Group:       group,
			Description: "Counts lines in a specified file. Returns line count as an integer.",
			Category:    "Filesystem",
			Args: []tool.ArgSpec{
				{Name: "filepath", Type: tool.ArgTypeString, Required: true, Description: "Relative path to the file."},
			},
			ReturnType:      tool.ArgTypeInt,
			ReturnHelp:      "Returns the number of lines in the specified file. Returns 0 on error or if file is empty.",
			Example:         `TOOL.FS.LineCount(filepath: "logs/app.log") // Returns 150 (example)`,
			ErrorConditions: "ErrArgumentMismatch if filepath is empty; ErrConfiguration if sandbox is not set; ErrSecurityPath for invalid paths; ErrFileNotFound; ErrPermissionDenied; ErrPathNotFile if path is a directory; ErrIOFailed for read errors. (Based on typical file tool error handling, actual implementation for toolLineCountFile in tools_fs_utils.go needed for exact errors).",
		},
		Func: toolLineCountFile, // from tools_fs_utils.go
	},
	{
		Spec: tool.ToolSpec{
			Name:        "SanitizeFilename",
			Group:       group,
			Description: "Cleans a string to make it suitable for use as part of a filename.",
			Category:    "Filesystem Utilities",
			Args: []tool.ArgSpec{
				{Name: "name", Type: tool.ArgTypeString, Required: true, Description: "The string to sanitize."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a sanitized string suitable for use as a filename component (e.g., replacing unsafe characters with underscores).",
			Example:         `TOOL.FS.SanitizeFilename(name: "My Report Final?.docx") // Returns "My_Report_Final_.docx" (example)`,
			ErrorConditions: "ErrArgumentMismatch if name is not provided or not a string. (Based on typical utility tool error handling, actual implementation for toolSanitizeFilename in tools_fs_utils.go needed for exact errors).",
		},
		Func: toolSanitizeFilename, // from tools_fs_utils.go
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Read",
			Group:       group,
			Description: "Reads the entire content of a specific file. Returns the content as a string.",
			Category:    "Filesystem",
			Args: []tool.ArgSpec{
				{Name: "filepath", Type: tool.ArgTypeString, Required: true, Description: "Relative path to the file."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns the content of the file as a string. Returns an empty string on error.",
			Example:         `TOOL.FS.Read(filepath: "config.txt") // Returns "setting=value\n..."`,
			ErrorConditions: "ErrArgumentMismatch if filepath is empty; ErrConfiguration if sandbox is not set; ErrSecurityPath (from ResolveAndSecurePath) for invalid paths; ErrFileNotFound if file does not exist; ErrPermissionDenied; ErrPathNotFile if path is a directory; ErrIOFailed for other I/O errors.",
		},
		Func: toolReadFile, // from tools_fs_read.go
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Write",
			Group:       group,
			Description: "Writes content to a specific file, overwriting it if it exists. Creates parent directories if needed. Returns 'OK' on success.",
			Category:    "Filesystem",
			Args: []tool.ArgSpec{
				{Name: "filepath", Type: tool.ArgTypeString, Required: true, Description: "Relative path to the file."},
				{Name: "content", Type: tool.ArgTypeString, Required: true, Description: "The content to write."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns 'OK' on success. Returns nil on error.",
			Example:         `TOOL.FS.Write(filepath: "output/data.json", content: "{\"key\":\"value\"}")`,
			ErrorConditions: "ErrArgumentMismatch if arguments are invalid; ErrConfiguration if sandbox is not set; ErrSecurityPath for invalid paths; ErrCannotCreateDir if parent directories cannot be created; ErrPermissionDenied if writing is not allowed; ErrPathNotFile if path exists and is a directory; ErrIOFailed for other I/O errors.",
		},
		Func: toolWriteFile, // from tools_fs_write.go
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Append",
			Group:       group,
			Description: "Appends content to a specific file. Creates the file and parent directories if needed. Returns 'OK' on success.",
			Category:    "Filesystem",
			Args: []tool.ArgSpec{
				{Name: "filepath", Type: tool.ArgTypeString, Required: true, Description: "Relative path to the file."},
				{Name: "content", Type: tool.ArgTypeString, Required: true, Description: "The content to append."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns 'OK' on success. Returns nil on error.",
			Example:         `TOOL.FS.Append(filepath: "logs/activity.log", content: "User logged in.\n")`,
			ErrorConditions: "ErrArgumentMismatch if arguments are invalid; ErrConfiguration if sandbox is not set; ErrSecurityPath for invalid paths; ErrCannotCreateDir if parent directories cannot be created; ErrPermissionDenied if writing is not allowed; ErrPathNotFile if path exists and is a directory; ErrIOFailed for other I/O errors.",
		},
		Func: toolAppendFile, // from tools_fs_write.go
	},
	{
		Spec: tool.ToolSpec{
			Name:        "List",
			Group:       group,
			Description: "Lists files and subdirectories at a given path. Returns a list of maps, each describing an entry (keys: name, path, isDir, size, modTime).",
			Category:    "Filesystem",
			Args: []tool.ArgSpec{
				{Name: "path", Type: tool.ArgTypeString, Required: true, Description: "Relative path to the directory (use '.' for current)."},
				{Name: "recursive", Type: tool.ArgTypeBool, Required: false, Description: "Whether to list recursively (default: false)."},
			},
			ReturnType:      tool.ArgTypeSliceAny, // Returns []map[string]interface{}
			ReturnHelp:      "Returns a slice of maps. Each map details a file/directory: {'name':string, 'path':string (relative to input path for recursive), 'isDir':bool, 'size':int64, 'modTime':string (RFC3339Nano)}. Returns nil on error.",
			Example:         `TOOL.FS.List(path: "mydir", recursive: true)`,
			ErrorConditions: "ErrArgumentMismatch if path is not a string or recursive is not bool/nil; ErrConfiguration if sandbox is not set; ErrSecurityPath (from ResolveAndSecurePath) for invalid path; ErrFileNotFound if path does not exist; ErrPermissionDenied; ErrPathNotDirectory if path is not a directory; ErrIOFailed for other I/O errors during listing or walking.",
		},
		Func: toolListDirectory, // from tools_fs_dirs.go
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Mkdir",
			Group:       group,
			Description: "Creates a directory. Parent directories are created if they do not exist (like mkdir -p). Returns a success message.",
			Category:    "Filesystem",
			Args: []tool.ArgSpec{
				{Name: "path", Type: tool.ArgTypeString, Required: true, Description: "Relative path of the directory to create."},
			},
			ReturnType:      tool.ArgTypeMap, // Returns map[string]interface{}{"status":"success", "message": "...", "path": "..."}
			ReturnHelp:      "Returns a map: {'status':'success', 'message':'Successfully created directory: <path>', 'path':'<path>'} on success. Returns nil on error.",
			Example:         `TOOL.FS.Mkdir(path: "new/subdir") // Returns {"status":"success", "message":"Successfully created directory: new/subdir", "path":"new/subdir"}`,
			ErrorConditions: "ErrArgumentMismatch if path is empty, '.', or not a string; ErrConfiguration if sandbox is not set; ErrSecurityPath (from ResolveAndSecurePath) for invalid path; ErrPathNotDirectory if path exists and is a file; ErrPathExists if directory already exists; ErrPermissionDenied; ErrIOFailed for other I/O errors or failure to stat; ErrCannotCreateDir if MkdirAll fails.",
		},
		Func: toolMkdir, // from tools_fs_dirs.go
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Delete",
			Group:       group,
			Description: "Deletes a file or an empty directory. Returns 'OK' on success or if path doesn't exist.",
			Category:    "Filesystem",
			Args: []tool.ArgSpec{
				{Name: "path", Type: tool.ArgTypeString, Required: true, Description: "Relative path to the file or empty directory to delete."},
			},
			ReturnType:      tool.ArgTypeString, // Returns "OK"
			ReturnHelp:      "Returns the string 'OK' on successful deletion or if the path does not exist. Returns nil on error.",
			Example:         `TOOL.FS.Delete(path: "temp/old_file.txt") // Returns "OK"`,
			ErrorConditions: "ErrArgumentMismatch if path is empty or not a string; ErrConfiguration if sandbox is not set; ErrSecurityPath (from SecureFilePath) for invalid path; ErrPreconditionFailed if directory is not empty; ErrPermissionDenied; ErrIOFailed for other I/O errors. Path not found is treated as success.",
		},
		Func: toolDeleteFile, // from tools_fs_delete.go
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Stat",
			Group:       group,
			Description: fmt.Sprintf("Gets information about a file or directory. Returns a map containing: name(string), path(string), size_bytes(int), is_dir(bool), modified_unix(int), modified_rfc3339(string - format %s), mode_string(string), mode_perm(string).", time.RFC3339Nano),
			Category:    "Filesystem",
			Args: []tool.ArgSpec{
				{Name: "path", Type: tool.ArgTypeString, Required: true, Description: "Relative path to the file or directory."},
			},
			ReturnType:      tool.ArgTypeMap,
			ReturnHelp:      "Returns a map with file/directory info: {'name', 'path', 'size_bytes', 'is_dir', 'modified_unix', 'modified_rfc3339', 'mode_string', 'mode_perm'}. Returns nil on error.",
			Example:         `TOOL.FS.Stat(path: "my_file.go")`,
			ErrorConditions: "ErrArgumentMismatch if path is empty or not a string; ErrConfiguration if sandbox is not set; ErrSecurityPath (from ResolveAndSecurePath) for invalid path; ErrFileNotFound if path does not exist; ErrPermissionDenied; ErrIOFailed for other I/O errors.",
		},
		Func: toolStat, // from tools_fs_stat.go
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Move",
			Group:       group,
			Description: "Moves or renames a file or directory within the sandbox. Returns a map: {'message': 'success message', 'error': nil} on success.",
			Category:    "Filesystem",
			Args: []tool.ArgSpec{
				{Name: "source_path", Type: tool.ArgTypeString, Required: true, Description: "Relative path of the source file/directory."},
				{Name: "destination_path", Type: tool.ArgTypeString, Required: true, Description: "Relative path of the destination."},
			},
			ReturnType:      tool.ArgTypeMap, // Returns map[string]interface{}
			ReturnHelp:      "Returns a map {'message': 'success message', 'error': nil} on success. Returns nil on error.",
			Example:         `TOOL.FS.Move(source_path: "old_name.txt", destination_path: "new_name.txt")`,
			ErrorConditions: "ErrArgumentMismatch if paths are empty, not strings, or are the same; ErrConfiguration if sandbox is not set; ErrSecurityPath (from SecureFilePath) for invalid source or destination paths; ErrFileNotFound if source path does not exist; ErrPathExists if destination path already exists; ErrPermissionDenied for source or destination; ErrIOFailed for other I/O errors during stat or rename.",
		},
		Func: toolMoveFile, // from tools_fs_move.go
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Walk",
			Group:       group,
			Description: fmt.Sprintf("Recursively walks a directory, returning a list of maps describing files/subdirectories found (keys: name, path_relative, is_dir, size_bytes, modified_unix, modified_rfc3339 (format %s), mode_string). Skips the root directory itself.", time.RFC3339Nano),
			Category:    "Filesystem",
			Args: []tool.ArgSpec{
				{Name: "path", Type: tool.ArgTypeString, Required: true, Description: "Relative path to the directory to walk."},
			},
			ReturnType:      tool.ArgTypeSliceAny, // Returns []map[string]interface{}
			ReturnHelp:      "Returns a slice of maps, each describing a file/subdir: {'name', 'path_relative', 'is_dir', 'size_bytes', 'modified_unix', 'modified_rfc3339', 'mode_string'}. Skips the root dir itself. Returns nil on error.",
			Example:         `TOOL.FS.Walk(path: "src")`,
			ErrorConditions: "ErrArgumentMismatch if path is empty or not a string; ErrConfiguration if sandbox is not set; ErrSecurityPath (from ResolveAndSecurePath) for invalid path; ErrFileNotFound if start path not found; ErrPathNotDirectory if start path is not a directory; ErrPermissionDenied for start path; ErrIOFailed for stat errors or errors during walk; ErrInternal if relative path calculation fails during walk.",
		},
		Func: toolWalkDir, // from tools_fs_walk.go
	},
}
