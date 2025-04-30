// filename: pkg/core/tools_file_api.go
package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	// For generating unique IDs if needed by FileAPI state
)

// --- Define FileAPIClient interface (if not already defined elsewhere) ---
// type FileAPIClient interface { ... see interpreter.go ... }

// --- Define FileAPI struct (if not already defined elsewhere) ---
// type FileAPI struct { ... see file containing its definition ... }

// --- Tool Implementations ---

// toolListAPIFiles implements TOOL.ListAPIFiles
func toolListAPIFiles(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	interpreter.logger.Info("Tool: ListAPIFiles] Listing files via API...")
	client, err := checkGenAIClient(interpreter)
	if err != nil {
		errMsg := fmt.Sprintf("ListAPIFiles failed: %v", err)
		interpreter.logger.Error("Tool: ListAPIFiles] %s", errMsg)
		return map[string]interface{}{"error": errMsg}, nil // Return error in map
	}

	// Call HelperListApiFiles (which uses syncContext internally)
	// We need to create a minimal context here for the helper
	syncCtx := &syncContext{client: client, logger: interpreter.logger, ctx: context.Background()}
	apiFiles, listErr := listExistingAPIFiles(syncCtx) // Call helper
	if listErr != nil {
		errMsg := fmt.Sprintf("ListAPIFiles failed: %v", listErr)
		interpreter.logger.Error("Tool: ListAPIFiles] %s", errMsg)
		return map[string]interface{}{"error": errMsg}, nil // Return error in map
	}

	// Convert []*genai.File to []map[string]interface{} for NeuroScript
	resultList := make([]interface{}, 0, len(apiFiles))
	for _, file := range apiFiles {
		fileMap := map[string]interface{}{
			"name":         file.Name,
			"display_name": file.DisplayName,
			"uri":          file.URI,
			"size_bytes":   file.SizeBytes,
			"create_time":  file.CreateTime.Format(time.RFC3339),
			"update_time":  file.UpdateTime.Format(time.RFC3339),
			"sha256_hash":  fmt.Sprintf("%x", file.Sha256Hash), // Hex encode hash
			"mime_type":    file.MIMEType,                      // <<< CORRECTED FIELD NAME
			"state":        file.State.String(),
		}
		resultList = append(resultList, fileMap)
	}

	interpreter.logger.Info("Tool: ListAPIFiles] Found %d files.", len(resultList))
	return resultList, nil
}

// toolDeleteAPIFile implements TOOL.DeleteAPIFile
func toolDeleteAPIFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	fileName := args[0].(string)
	interpreter.logger.Info("Tool: DeleteAPIFile] Requesting deletion of API file: %s", fileName)

	if fileName == "" {
		return "DeleteAPIFile failed: File name cannot be empty.", nil
	}
	if !strings.HasPrefix(fileName, "files/") {
		interpreter.logger.Warn("Tool: DeleteAPIFile] Filename '%s' does not look like a standard File API name.", fileName)
	}

	client, err := checkGenAIClient(interpreter)
	if err != nil {
		errMsg := fmt.Sprintf("DeleteAPIFile failed: %v", err)
		interpreter.logger.Error("Tool: DeleteAPIFile] %s", errMsg)
		return map[string]interface{}{"error": errMsg}, nil
	}

	err = client.DeleteFile(context.Background(), fileName)
	if err != nil {
		errMsg := fmt.Sprintf("DeleteAPIFile API call failed for '%s': %v", fileName, err)
		interpreter.logger.Error("Tool: DeleteAPIFile] %s", errMsg)
		return map[string]interface{}{"error": errMsg}, nil
	}

	interpreter.logger.Info("Tool: DeleteAPIFile] Successfully requested deletion for: %s", fileName)
	return map[string]interface{}{"error": nil}, nil
}

// toolUploadFile implements TOOL.UploadFile
func toolUploadFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	localPathRel := args[0].(string)
	var displayName string
	if len(args) > 1 && args[1] != nil {
		displayName = args[1].(string)
	} else {
		displayName = filepath.Base(localPathRel)
	}

	interpreter.logger.Info("Tool: UploadFile] Requesting upload for local path: %s as display name: %s", localPathRel, displayName)

	absLocalPath, secErr := SecureFilePath(localPathRel, interpreter.sandboxDir)
	if secErr != nil {
		errMsg := fmt.Sprintf("UploadFile invalid local path '%s': %v", localPathRel, secErr)
		interpreter.logger.Error("Tool: UploadFile] %s", errMsg)
		return map[string]interface{}{"error": errMsg}, nil
	}

	info, statErr := os.Stat(absLocalPath)
	if statErr != nil {
		errMsg := ""
		if os.IsNotExist(statErr) {
			errMsg = fmt.Sprintf("UploadFile failed: Local file not found at '%s'", localPathRel)
		} else {
			errMsg = fmt.Sprintf("UploadFile failed stat local file '%s': %v", localPathRel, statErr)
		}
		interpreter.logger.Error("Tool: UploadFile] %s", errMsg)
		return map[string]interface{}{"error": errMsg}, nil
	}
	if info.IsDir() {
		errMsg := fmt.Sprintf("UploadFile failed: Local path '%s' is a directory, not a file", localPathRel)
		interpreter.logger.Error("Tool: UploadFile] %s", errMsg)
		return map[string]interface{}{"error": errMsg}, nil
	}

	client, err := checkGenAIClient(interpreter)
	if err != nil {
		errMsg := fmt.Sprintf("UploadFile failed: %v", err)
		interpreter.logger.Error("Tool: UploadFile] %s", errMsg)
		return map[string]interface{}{"error": errMsg}, nil
	}

	apiFile, uploadErr := HelperUploadAndPollFile(context.Background(), absLocalPath, displayName, client, interpreter.logger)

	if uploadErr != nil {
		errMsg := fmt.Sprintf("UploadFile failed for '%s': %v", localPathRel, uploadErr)
		interpreter.logger.Error("Tool: UploadFile] %s", errMsg)
		if apiFile != nil {
			return map[string]interface{}{"error": errMsg, "api_name": apiFile.Name, "display_name": apiFile.DisplayName, "state": apiFile.State.String()}, nil
		}
		return map[string]interface{}{"error": errMsg}, nil
	}

	interpreter.logger.Info("Tool: UploadFile] Successfully uploaded '%s' as '%s' (API Name: %s)", localPathRel, displayName, apiFile.Name)
	resultMap := map[string]interface{}{
		"error":        nil,
		"api_name":     apiFile.Name,
		"display_name": apiFile.DisplayName,
		"uri":          apiFile.URI,
		"size_bytes":   apiFile.SizeBytes,
		"create_time":  apiFile.CreateTime.Format(time.RFC3339),
		"update_time":  apiFile.UpdateTime.Format(time.RFC3339),
		"sha256_hash":  fmt.Sprintf("%x", apiFile.Sha256Hash),
		"mime_type":    apiFile.MIMEType, // <<< CORRECTED FIELD NAME
		"state":        apiFile.State.String(),
	}
	return resultMap, nil
}

// --- Registration Function (ADDED) ---
func registerFileAPITools(registry *ToolRegistry) error {
	tools := []ToolImplementation{
		{
			Spec: ToolSpec{
				Name:        "ListAPIFiles",
				Description: "Lists files currently available in the File API.",
				Args:        []ArgSpec{},     // No arguments
				ReturnType:  ArgTypeSliceAny, // Returns list of maps
			},
			Func: toolListAPIFiles,
		},
		{
			Spec: ToolSpec{
				Name:        "DeleteAPIFile",
				Description: "Deletes a specific file from the File API using its full name (e.g., 'files/...')",
				Args: []ArgSpec{
					{Name: "file_name", Type: ArgTypeString, Required: true, Description: "The full API name of the file to delete."},
				},
				ReturnType: ArgTypeAny, // Returns map {"error": string|null}
			},
			Func: toolDeleteAPIFile,
		},
		{
			Spec: ToolSpec{
				Name:        "UploadFile",
				Description: "Uploads a local file to the File API. Polls until file state is ACTIVE or FAILED.",
				Args: []ArgSpec{
					{Name: "local_path", Type: ArgTypeString, Required: true, Description: "Relative path to the local file to upload."},
					{Name: "display_name", Type: ArgTypeString, Required: false, Description: "Optional display name for the file in the API (defaults to local filename)."},
				},
				ReturnType: ArgTypeAny, // Returns map with file details or error
			},
			Func: toolUploadFile,
		},
		{
			Spec: ToolSpec{
				Name: "SyncFiles",
				Description: "Synchronizes files between a local directory and the File API. " +
					"Direction 'up' pushes local changes to the API. " +
					"Returns a map with statistics.",
				Args: []ArgSpec{
					{Name: "direction", Type: ArgTypeString, Required: true, Description: "Sync direction ('up' supported)."},
					{Name: "local_dir", Type: ArgTypeString, Required: true, Description: "Relative path to the local directory."},
					{Name: "filter_pattern", Type: ArgTypeString, Required: false, Description: "Optional glob pattern to filter files (e.g., '*.ns')."},
					{Name: "ignore_gitignore", Type: ArgTypeBool, Required: false, Description: "If true, ignores .gitignore files (defaults to false)."},
				},
				ReturnType: ArgTypeAny, // Returns map[string]interface{}
			},
			Func: toolSyncFiles, // Assumes toolSyncFiles is defined elsewhere (sync_logic.go)
		},
	}
	var errs []error
	for _, tool := range tools {
		if err := registry.RegisterTool(tool); err != nil {
			errs = append(errs, fmt.Errorf("register %s: %w", tool.Spec.Name, err))
		}
	}
	if len(errs) > 0 {
		errorMessages := make([]string, len(errs))
		for i, e := range errs {
			errorMessages[i] = e.Error()
		}
		return errors.New(strings.Join(errorMessages, "; "))
	}
	return nil
}
