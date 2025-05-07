// NeuroScript Version: 0.3.0
// File version: 0.1.5
// Correct remaining undefined errors
// filename: pkg/core/tools_file_api.go

package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// toolListAPIFiles implements TOOL.ListAPIFiles
func toolListAPIFiles(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	interpreter.Logger().Info("Tool: ListAPIFiles] Listing files via API...")
	client, err := checkGenAIClient(interpreter)
	if err != nil {
		errMsg := fmt.Sprintf("ListAPIFiles failed to get GenAI client: %v", err)
		interpreter.Logger().Error("Tool: ListAPIFiles] %s", errMsg)
		// Use ErrorCodeLLMError and wrap ErrLLMNotConfigured
		return map[string]interface{}{"error": errMsg}, NewRuntimeError(ErrorCodeLLMError, errMsg, ErrLLMNotConfigured)
	}

	syncCtx := &syncContext{client: client, logger: interpreter.Logger(), ctx: context.Background()}
	apiFiles, listErr := listExistingAPIFiles(syncCtx)
	if listErr != nil {
		errMsg := fmt.Sprintf("ListAPIFiles call failed: %v", listErr)
		interpreter.Logger().Error("Tool: ListAPIFiles] %s", errMsg)
		// Using ErrorCodeInternal for general API call failures.
		// If listExistingAPIFiles returns specific sentinel errors, they should be preserved or wrapped.
		return map[string]interface{}{"error": errMsg}, NewRuntimeError(ErrorCodeInternal, errMsg, listErr)
	}

	resultList := make([]interface{}, 0, len(apiFiles))
	for _, file := range apiFiles {
		fileMap := map[string]interface{}{
			"name":         file.Name,
			"display_name": file.DisplayName,
			"uri":          file.URI,
			"size_bytes":   file.SizeBytes,
			"create_time":  file.CreateTime.Format(time.RFC3339Nano),
			"update_time":  file.UpdateTime.Format(time.RFC3339Nano),
			"sha256_hash":  fmt.Sprintf("%x", file.Sha256Hash),
			"mime_type":    file.MIMEType,
			"state":        file.State.String(),
		}
		resultList = append(resultList, fileMap)
	}

	interpreter.Logger().Infof("Tool: ListAPIFiles] Found %d files.", len(resultList))
	return resultList, nil
}

// toolDeleteAPIFile implements TOOL.DeleteAPIFile
func toolDeleteAPIFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "DeleteAPIFile expects 1 argument: file_name", ErrInvalidArgument)
	}
	fileName, ok := args[0].(string)
	if !ok || fileName == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "DeleteAPIFile expects a non-empty string file_name", ErrInvalidArgument)
	}
	interpreter.Logger().Infof("Tool: DeleteAPIFile] Requesting deletion of API file: %s", fileName)

	if !strings.HasPrefix(fileName, "files/") {
		interpreter.Logger().Warnf("Tool: DeleteAPIFile] Filename '%s' does not start with 'files/'. This might not be a valid File API name.", fileName)
	}

	client, err := checkGenAIClient(interpreter)
	if err != nil {
		errMsg := fmt.Sprintf("DeleteAPIFile failed to get GenAI client: %v", err)
		interpreter.Logger().Error("Tool: DeleteAPIFile] %s", errMsg)
		// Use ErrorCodeLLMError and wrap ErrLLMNotConfigured
		return map[string]interface{}{"error": errMsg}, NewRuntimeError(ErrorCodeLLMError, errMsg, ErrLLMNotConfigured)
	}

	deleteErr := client.DeleteFile(context.Background(), fileName)
	if deleteErr != nil {
		errMsg := fmt.Sprintf("DeleteAPIFile API call failed for '%s': %v", fileName, deleteErr)
		interpreter.Logger().Error("Tool: DeleteAPIFile] %s", errMsg)
		// Using ErrorCodeInternal. If client.DeleteFile returns specific sentinel errors (e.g., for not found),
		// those should be checked and mapped appropriately (e.g., to ErrorCodeKeyNotFound and ErrFileNotFound).
		return map[string]interface{}{"error": errMsg}, NewRuntimeError(ErrorCodeInternal, errMsg, deleteErr)
	}

	successMsg := fmt.Sprintf("Deletion requested successfully for API file: %s", fileName)
	interpreter.Logger().Infof("Tool: DeleteAPIFile] %s", successMsg)
	return map[string]interface{}{"message": successMsg, "error": nil}, nil
}

// toolUploadFile implements TOOL.UploadFile
func toolUploadFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "UploadFile expects 1 or 2 arguments: local_path, [display_name]", ErrInvalidArgument)
	}
	localPathRel, ok := args[0].(string)
	if !ok || localPathRel == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "UploadFile expects a non-empty string local_path", ErrInvalidArgument)
	}

	var displayName string
	if len(args) > 1 && args[1] != nil {
		displayName, ok = args[1].(string)
		if !ok {
			return nil, NewRuntimeError(ErrorCodeArgMismatch, "UploadFile display_name must be a string if provided", ErrInvalidArgument)
		}
	}
	if displayName == "" {
		displayName = filepath.Base(localPathRel)
	}

	interpreter.Logger().Infof("Tool: UploadFile] Requesting upload for local path: %s as display name: %s", localPathRel, displayName)

	// SecureFilePath already returns a RuntimeError wrapping a sentinel error (e.g., ErrInvalidPath, ErrPathViolation)
	absLocalPath, secErr := SecureFilePath(localPathRel, interpreter.SandboxDir())
	if secErr != nil {
		// secErr is already a *RuntimeError, so we can return it directly.
		interpreter.Logger().Error("Tool: UploadFile] SecureFilePath failed for '%s': %v", localPathRel, secErr)
		return map[string]interface{}{"error": secErr.Error()}, secErr
	}

	info, statErr := os.Stat(absLocalPath)
	if statErr != nil {
		errMsg := ""
		var rtErr *RuntimeError
		if os.IsNotExist(statErr) {
			errMsg = fmt.Sprintf("local file not found at '%s'", localPathRel)
			// Use ErrorCodeKeyNotFound and wrap ErrFileNotFound
			rtErr = NewRuntimeError(ErrorCodeKeyNotFound, errMsg, ErrFileNotFound)
		} else {
			errMsg = fmt.Sprintf("failed to stat local file '%s'", localPathRel)
			// Use ErrorCodeToolSpecific (since ErrorCodeFileRead is not defined) and wrap the original statErr
			rtErr = NewRuntimeError(ErrorCodeToolSpecific, errMsg, statErr)
		}
		interpreter.Logger().Error("Tool: UploadFile] %s (resolved: %s): %v", errMsg, absLocalPath, statErr)
		return map[string]interface{}{"error": errMsg}, rtErr
	}
	if info.IsDir() {
		errMsg := fmt.Sprintf("local path '%s' is a directory, not a file", localPathRel)
		interpreter.Logger().Error("Tool: UploadFile] %s (resolved: %s)", errMsg, absLocalPath)
		// Use ErrorCodeArgMismatch and wrap ErrInvalidArgument (since ErrPathIsDirectory is not defined)
		return map[string]interface{}{"error": errMsg}, NewRuntimeError(ErrorCodeArgMismatch, errMsg, ErrInvalidArgument)
	}

	client, err := checkGenAIClient(interpreter)
	if err != nil {
		errMsg := fmt.Sprintf("UploadFile failed to get GenAI client: %v", err)
		interpreter.Logger().Error("Tool: UploadFile] %s", errMsg)
		// Use ErrorCodeLLMError and wrap ErrLLMNotConfigured
		return map[string]interface{}{"error": errMsg}, NewRuntimeError(ErrorCodeLLMError, errMsg, ErrLLMNotConfigured)
	}

	apiFile, uploadErr := HelperUploadAndPollFile(context.Background(), absLocalPath, displayName, client, interpreter.Logger())
	if uploadErr != nil {
		errMsg := fmt.Sprintf("UploadFile failed for '%s': %v", localPathRel, uploadErr)
		interpreter.Logger().Error("Tool: UploadFile] %s", errMsg)
		// Using ErrorCodeLLMError if it's an LLM/API interaction issue, otherwise ErrorCodeInternal.
		// HelperUploadAndPollFile should ideally return wrapped sentinel errors.
		var returnErr error
		if _, ok := uploadErr.(*RuntimeError); ok {
			returnErr = uploadErr
		} else {
			// Defaulting to ErrorCodeInternal, but could be ErrorCodeLLMError if appropriate
			returnErr = NewRuntimeError(ErrorCodeInternal, errMsg, uploadErr)
		}

		if apiFile != nil {
			return map[string]interface{}{"error": errMsg, "api_name": apiFile.Name, "display_name": apiFile.DisplayName, "state": apiFile.State.String()}, returnErr
		}
		return map[string]interface{}{"error": errMsg}, returnErr
	}
	if apiFile == nil {
		errMsg := fmt.Sprintf("UploadFile failed for '%s': HelperUploadAndPollFile returned nil file without error", localPathRel)
		interpreter.Logger().Error("Tool: UploadFile] %s", errMsg)
		return map[string]interface{}{"error": errMsg}, NewRuntimeError(ErrorCodeInternal, errMsg, ErrInternalTool) // Wrap ErrInternalTool
	}

	interpreter.Logger().Infof("Tool: UploadFile] Successfully uploaded '%s' as '%s' (API Name: %s, URI: %s)", localPathRel, displayName, apiFile.Name, apiFile.URI)
	resultMap := map[string]interface{}{
		"error":        nil,
		"api_name":     apiFile.Name,
		"display_name": apiFile.DisplayName,
		"uri":          apiFile.URI,
		"size_bytes":   apiFile.SizeBytes,
		"create_time":  apiFile.CreateTime.Format(time.RFC3339Nano),
		"update_time":  apiFile.UpdateTime.Format(time.RFC3339Nano),
		"sha256_hash":  fmt.Sprintf("%x", apiFile.Sha256Hash),
		"mime_type":    apiFile.MIMEType,
		"state":        apiFile.State.String(),
	}
	return resultMap, nil
}

// toolSyncFiles is defined in pkg/core/tools_file_api_sync.go
// Its spec is included here for registration.

func init() {
	// This debug print is kept as per Rule 22.
	fmt.Println(">>>>>>>>>>>> DEBUG: pkg/core/tools_file_api.go init() CALLED <<<<<<<<<<<<")

	fileApiTools := []ToolImplementation{
		{
			Spec: ToolSpec{
				Name:        "ListAPIFiles",
				Description: "Lists files currently available in the File API.",
				Args:        []ArgSpec{},
				ReturnType:  ArgTypeSliceMap,
			},
			Func: toolListAPIFiles,
		},
		{
			Spec: ToolSpec{
				Name:        "DeleteAPIFile",
				Description: "Deletes a specific file from the File API using its full name (e.g., 'files/...').",
				Args: []ArgSpec{
					{Name: "file_name", Type: ArgTypeString, Required: true, Description: "The full API name of the file to delete."},
				},
				ReturnType: ArgTypeMap,
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
				ReturnType: ArgTypeMap,
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
				ReturnType: ArgTypeMap,
			},
			Func: toolSyncFiles, // Defined in tools_file_api_sync.go
		},
	}
	// Ensure AddToolImplementations is called correctly.
	AddToolImplementations(fileApiTools...)
	// This debug print is kept as per Rule 22.
	fmt.Printf(">>>>>>>>>>>> DEBUG: pkg/core/tools_file_api.go AddToolImplementations called for %d tools <<<<<<<<<<<<\n", len(fileApiTools))
}
