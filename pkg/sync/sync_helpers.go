// filename: pkg/sync/sync_helpers.go
package sync

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/google/generative-ai-go/genai"
)

// --- checkGenAIClient Helper ---
// (Function remains the same)
func checkGenAIClient(interp tool.RunTime) (*genai.Client, error) {
	if interp == nil || interp.llmClient == nil {
		return nil, errors.New("interpreter or LLMClient not configured")
	}
	client := interp.llmClient.Client() // Use the interface method
	if client == nil {
		return nil, errors.New("LLM client is not a compatible GenAI client or is not initialized")
	}
	return client, nil
}

// --- HelperUploadAndPollFile Helper ---
// This function uploads a local file and polls the GenAI API until the file is Active or Failed.
func HelperUploadAndPollFile(
	ctx context.Context,
	localPath string, // Absolute path to local file *resolved by caller*
	displayName string, // Desired display name (relative path) in API
	client *genai.Client,
	logger interfaces.Logger,
) (*genai.File, error) {

	if client == nil {
		return nil, errors.New("genai client is nil")
	}
	if logger == nil {
		logger = &utils.coreNoOpLogger{} // Basic fallback
	}

	// 1. Open local file (Path assumed to be absolute and valid here)
	fileReader, err := os.Open(localPath)
	if err != nil {
		logger.Error("[HelperUpload] Failed to open local file", "path", localPath, "error", err)
		return nil, fmt.Errorf("opening local file %q: %w", localPath, err)
	}
	defer fileReader.Close()

	// 2. Upload file
	logger.Debug("[HelperUpload] Uploading file", "local_path", localPath, "display_name", displayName)
	uploadStartTime := time.Now()

	// *** CORRECTED based on VS Code signature ***
	// Create the options struct and set the DisplayName field.
	// Pass "" for the 'name' parameter to let the service generate a unique ID.
	uploadOpts := &genai.UploadFileOptions{
		DisplayName: displayName,
		// MimeType could potentially be inferred or set here if needed, e.g.:
		// MimeType: mime.TypeByExtension(filepath.Ext(localPath)),
	}
	uploadedFile, err := client.UploadFile(ctx, "", fileReader, uploadOpts) // Pass "" for name, and pointer to options struct
	if err != nil {
		logger.Error("[HelperUpload] client.UploadFile failed", "display_name", displayName, "error", err)
		// Consider logging uploadOpts content here on error for debugging
		return nil, fmt.Errorf("uploading file %q: %w", displayName, err)
	}
	// *** END CORRECTION ***

	if uploadedFile == nil {
		logger.Error("[HelperUpload] client.UploadFile returned nil file", "display_name", displayName)
		return nil, fmt.Errorf("upload API returned nil file for %q", displayName)
	}
	logger.Debug("[HelperUpload] Initial upload accepted",
		"requested_display_name", displayName,
		"remote_name", uploadedFile.Name,
		"actual_display_name", uploadedFile.DisplayName, // Check if this field is populated correctly now
		"initial_state", uploadedFile.State,
		"upload_duration_ms", time.Since(uploadStartTime).Milliseconds())

	// 3. Poll for status
	const (
		pollInterval    = 2 * time.Second
		maxPollDuration = 2 * time.Minute
	)
	pollCtx, cancelPoll := context.WithTimeout(ctx, maxPollDuration)
	defer cancelPoll()

	pollingStartTime := time.Now()
	logger.Debug("[HelperUpload] Polling status", "display_name", displayName, "remote_name", uploadedFile.Name)

	for {
		// Check context cancellation first
		if pollCtx.Err() != nil {
			logger.Error("[HelperUpload] Polling context cancelled/timed out",
				"display_name", displayName,
				"remote_name", uploadedFile.Name,
				"elapsed_ms", time.Since(pollingStartTime).Milliseconds(),
				"error", pollCtx.Err())
			return uploadedFile, fmt.Errorf("polling file status for %q timed out or was cancelled after %s: %w",
				displayName, time.Since(pollingStartTime).Round(time.Second), pollCtx.Err())
		}

		getFileStartTime := time.Now()
		file, err := client.GetFile(pollCtx, uploadedFile.Name) // Use the internal remote name
		getFileDuration := time.Since(getFileStartTime)

		if err != nil {
			logger.Error("[HelperUpload] Polling GetFile failed",
				"display_name", displayName, // Use requested name for consistency in error msg
				"remote_name", uploadedFile.Name,
				"getfile_duration_ms", getFileDuration.Milliseconds(),
				"error", err)
			return uploadedFile, fmt.Errorf("polling GetFile failed for %q: %w", displayName, err)
		}

		logger.Debug("[HelperUpload] Polling check",
			"requested_display_name", displayName,
			"remote_name", file.Name,
			"actual_display_name", file.DisplayName,
			"state", file.State,
			"getfile_duration_ms", getFileDuration.Milliseconds())

		switch file.State {
		case genai.FileStateActive:
			logger.Debug("[HelperUpload] File is ACTIVE",
				"display_name", file.DisplayName,
				"remote_name", file.Name,
				"total_polling_duration_ms", time.Since(pollingStartTime).Milliseconds())
			return file, nil // Success!
		case genai.FileStateFailed:
			logger.Error("[HelperUpload] File processing FAILED",
				"display_name", file.DisplayName,
				"remote_name", file.Name,
				"state", file.State)
			return file, fmt.Errorf("file processing failed remotely for %q (State: %s)", file.DisplayName, file.State)
		case genai.FileStateProcessing:
			// Continue polling
		default:
			logger.Warn("[HelperUpload] File has unexpected state",
				"display_name", file.DisplayName,
				"remote_name", file.Name,
				"state", file.State)
			return file, fmt.Errorf("file %q has unexpected state %q", file.DisplayName, file.State)
		}

		select {
		case <-time.After(pollInterval):
			// Continue loop
		case <-pollCtx.Done():
			logger.Debug("[HelperUpload] Polling context done during sleep")
		}
	}
}

// --- calculateFileHash ---
// (Function remains the same)
func calculateFileHash(interp tool.RunTime, relPath string) (string, error) {
	if interp == nil {
		return "", errors.New("calculateFileHash: interpreter is nil")
	}
	fileAPI := interp.FileAPI()
	if fileAPI == nil {
		interp.Logger().Error("calculateFileHash: interpreter FileAPI() returned nil!")
		return "", errors.New("calculateFileHash: interpreter FileAPI is nil")
	}

	absPath, err := fileAPI.ResolvePath(relPath)
	if err != nil {
		return "", fmt.Errorf("calculateFileHash: resolving path '%s': %w", relPath, err)
	}

	file, err := os.Open(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("%w: file not found at '%s' (resolved: %s)", lang.ErrFileNotFound, relPath, absPath)
		}
		return "", fmt.Errorf("calculateFileHash: failed to open file '%s' for hashing: %w", absPath, err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("calculateFileHash: failed to stat file '%s': %w", absPath, err)
	}
	if stat.IsDir() {
		return "", fmt.Errorf("%w: path '%s' (resolved: %s) is a directory, cannot hash", lang.ErrValidationArgValue, relPath, absPath)
	}

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("calculateFileHash: failed to read file content for '%s': %w", absPath, err)
	}

	hash := hex.EncodeToString(hasher.Sum(nil))
	interp.Logger().Debug("Calculated file hash", "path", relPath, "resolved_path", absPath, "hash", hash)
	return hash, nil
}
