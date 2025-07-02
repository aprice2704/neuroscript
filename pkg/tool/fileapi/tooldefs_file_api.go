// NeuroScript Version: 0.3.1
// File version: 0.1.3 // Use MakeUnimplementedToolFunc factory directly for stubs.
// Defines ToolImplementation structs for File API tools.
// filename: pkg/tool/fileapi/tooldefs_file_api.go

// nlines: 85 // Approximate
// risk_rating: MEDIUM

package fileapi

// fileApiToolsToRegister holds the definitions for File API tools.
// This slice is appended in zz_core_tools_registrar.go
var fileApiToolsToRegister = []ToolImplementation{
	{
		Spec: ToolSpec{
			Name:		"ListAPIFiles",
			Description:	"Lists files currently available via the platform's File API.",
			Args:		[]ArgSpec{},		// No arguments expected
			ReturnType:	ArgTypeSliceAny,	// Expect []map[string]interface{} describing files
		},
		// Use the factory to create the stub function
		Func:	toolListAPIFiles,
	},
	{
		Spec: ToolSpec{
			Name:		"DeleteAPIFile",
			Description:	"Deletes a specific file from the platform's File API using its ID/URI.",
			Args: []ArgSpec{
				{Name: "api_file_id", Type: ArgTypeString, Required: true, Description: "The unique ID or URI of the file on the API (e.g., 'files/abcde123')."},
			},
			ReturnType:	ArgTypeString,	// e.g., "OK" or confirmation message
		},
		// Use the factory to create the stub function
		Func:	toolDeleteAPIFile,
	},
	{
		Spec: ToolSpec{
			Name:		"UploadFile",
			Description:	"Uploads a local file (from the sandbox) to the platform's File API. Returns a map describing the uploaded file.",
			Args: []ArgSpec{
				{Name: "local_filepath", Type: ArgTypeString, Required: true, Description: "Relative path (within the sandbox) of the local file to upload."},
				{Name: "api_display_name", Type: ArgTypeString, Required: false, Description: "Optional display name for the file on the API."},
			},
			ReturnType:	ArgTypeMap,	// map[string]interface{} like { "name": "...", "uri": "...", "sizeBytes": ..., ... }
		},
		// Use the factory to create the stub function
		// Note: The actual implementation for UploadFile might still call HelperUploadAndPollFile
		Func:	toolUploadFile,
	},
	{
		Spec: ToolSpec{
			Name:		"SyncFiles",
			Description:	"Synchronizes files between a local sandbox directory and the platform's File API. Supports 'up' (local to API) and 'down' (API to local) directions.",
			Args: []ArgSpec{
				{Name: "direction", Type: ArgTypeString, Required: true, Description: "Sync direction: 'up' (local to API) or 'down' (API to local)."},
				{Name: "local_dir", Type: ArgTypeString, Required: true, Description: "Relative path (within the sandbox) of the local directory to sync."},
				{Name: "filter_pattern", Type: ArgTypeString, Required: false, Description: "Optional glob pattern (e.g., '*.go', 'data/**') to filter files being synced. Applies to filenames relative to local_dir."},
				{Name: "ignore_gitignore", Type: ArgTypeBool, Required: false, Description: "If true, ignores .gitignore rules found within the local_dir (default: false)."},
			},
			ReturnType:	ArgTypeMap,	// map[string]interface{} with sync statistics
		},
		// Assumes toolSyncFiles exists in sync_logic.go with the correct ToolFunc signature.
		Func:	toolSyncFiles,
	},
}

// Ensure toolSyncFiles (from sync_logic.go) matches the type, otherwise compilation fails here.
// This check remains useful.
var _ ToolFunc = toolSyncFiles