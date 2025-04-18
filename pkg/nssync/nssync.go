// filename: pkg/nssync/nssync.go
package nssync

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	gitignore "github.com/sabhiram/go-gitignore"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	// *** Use the correct genai import path based on your go.mod ***
	genai "google.golang.org/genai" // Corrected import path

	"github.com/aprice2704/neuroscript/pkg/core" // Adjust import path as needed
)

// RemoteFile struct remains the same
type RemoteFile struct {
	Name        string
	DisplayName string
	SizeBytes   int64
	SHA256Hash  string
	State       string
	UpdateTime  time.Time
	URI         string
}

// SyncerConfig struct remains the same
type SyncerConfig struct {
	APIKey          string
	DryRun          bool
	IgnoreGitignore bool
	IncludePatterns []string
	ExcludePatterns []string
}

// Syncer struct updated to include FileService client
type Syncer struct {
	config    SyncerConfig
	client    *genai.Client      // Main genai client
	fs        *genai.FileService // Specific client for File API operations
	interp    *core.Interpreter  // Interpreter for core facility access
	gitIgnore *gitignore.GitIgnore
	logger    *log.Logger
}

// NewSyncer creates a new Syncer instance.
func NewSyncer(config SyncerConfig, interp *core.Interpreter, logger *log.Logger) (*Syncer, error) {
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}
	if interp == nil {
		return nil, fmt.Errorf("interpreter cannot be nil")
	}
	// Use the exported getter method
	sandboxDir := interp.SandboxDir() // Use Exported Method
	if sandboxDir == "" {
		return nil, fmt.Errorf("interpreter provided to Syncer must have a non-empty sandbox directory configured")
	}

	// Initialize Gemini Client
	ctx := context.Background()
	// Use option.WithAPIKey directly
	client, err := genai.NewClient(ctx, option.WithAPIKey(config.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	// Get the FileService client from the main client
	fsClient := client.FileService() // Get the service client

	s := &Syncer{
		config: config,
		client: client,
		fs:     fsClient, // Store the FileService client
		interp: interp,
		logger: logger,
	}

	// Load .gitignore (using sandboxDir)
	if !config.IgnoreGitignore {
		gitignorePath := filepath.Join(sandboxDir, ".gitignore")
		if _, err := os.Stat(gitignorePath); err == nil {
			s.logger.Printf("Info: Loading .gitignore from %s", gitignorePath)
			gi, err := gitignore.CompileIgnoreFile(gitignorePath)
			if err != nil {
				s.logger.Printf("Warning: Failed to compile .gitignore file %s: %v", gitignorePath, err)
			} else {
				s.gitIgnore = gi
			}
		} else if !os.IsNotExist(err) {
			s.logger.Printf("Warning: Error checking .gitignore file %s: %v", gitignorePath, err)
		} else {
			s.logger.Printf("Info: No .gitignore file found at %s", gitignorePath)
		}
	}

	return s, nil
}

// Sync synchronizes the provided list of relative local file paths with the remote API.
func (s *Syncer) Sync(ctx context.Context, relativeFiles []string) error {
	s.logger.Printf("Starting sync for %d potential local files...", len(relativeFiles))
	if s.config.DryRun {
		s.logger.Println("--- DRY RUN MODE ---")
	}

	// 1. Get Remote Files
	remoteFiles, err := s.listRemoteFiles(ctx) // Uses s.fs internally now
	if err != nil {
		return fmt.Errorf("failed to list remote files: %w", err)
	}
	// Build remoteFileMap (logic remains the same)
	remoteFileMap := make(map[string]RemoteFile)
	for _, rf := range remoteFiles {
		if rf.State == "ACTIVE" && rf.DisplayName != "" {
			if existing, exists := remoteFileMap[rf.DisplayName]; exists {
				s.logger.Printf("Warning: Duplicate remote DisplayName '%s' found (Names: %s, %s). Using the one updated later.", rf.DisplayName, existing.Name, rf.Name)
				if rf.UpdateTime.After(existing.UpdateTime) {
					remoteFileMap[rf.DisplayName] = rf
				}
			} else {
				remoteFileMap[rf.DisplayName] = rf
			}
		} else {
			s.logger.Printf("Debug: Skipping remote file Name: %s, DisplayName: '%s', State: %s during map build", rf.Name, rf.DisplayName, rf.State)
		}
	}
	s.logger.Printf("Found %d ACTIVE remote files with DisplayNames mapped.", len(remoteFileMap))

	// 2. Prepare tracking maps/slices (remains the same)
	localFileSet := make(map[string]bool)
	filesToUpload := make(map[string]string)
	filesToDelete := []RemoteFile{}
	var errorsEncountered []string

	// 3. Determine Uploads
	s.logger.Println("Analyzing local files for upload actions...")
	for _, relPath := range relativeFiles {
		cleanRelPath := filepath.ToSlash(filepath.Clean(relPath))
		if cleanRelPath == "." || cleanRelPath == "" {
			continue
		}
		localFileSet[cleanRelPath] = true

		// Apply gitignore (remains the same)
		if s.gitIgnore != nil && s.gitIgnore.MatchesPath(cleanRelPath) {
			s.logger.Printf("Info: Skipping '%s' due to .gitignore", cleanRelPath)
			continue
		}

		// Calculate local file hash using the INTERPRETER'S tool execution method
		// Use ExecuteToolCall
		hashResult, err := s.interp.ExecuteToolCall("FileHash", cleanRelPath) // Default SHA256
		if err != nil {
			// Use errors.Is and the CORRECTED core.ErrFileNotFound
			if errors.Is(err, core.ErrPathViolation) || errors.Is(err, core.ErrFileNotFound) {
				s.logger.Printf("Warning: Cannot get hash for local file '%s' (likely missing or path issue): %v. Skipping.", cleanRelPath, err)
				// Don't add error here, just skip the file
			} else {
				s.logger.Printf("Warning: Failed to hash local file '%s': %v. Skipping upload.", cleanRelPath, err)
				errorsEncountered = append(errorsEncountered, fmt.Sprintf("hash failed for %s: %v", cleanRelPath, err))
			}
			continue // Skip this file if hashing failed
		}

		localHash, ok := hashResult.(string)
		if !ok || localHash == "" {
			s.logger.Printf("Warning: Tool FileHash returned unexpected type (%T) or empty hash for '%s'. Skipping.", hashResult, cleanRelPath)
			errorsEncountered = append(errorsEncountered, fmt.Sprintf("invalid hash result for %s", cleanRelPath))
			continue
		}

		// Compare with remote file (remains the same)
		remoteFile, existsRemotely := remoteFileMap[cleanRelPath]
		if existsRemotely {
			if remoteFile.SHA256Hash != localHash {
				s.logger.Printf("Action: Plan upload for '%s' (local hash %s... != remote hash %s...)", cleanRelPath, localHash[:8], remoteFile.SHA256Hash[:8])
				filesToUpload[cleanRelPath] = localHash
			} else {
				s.logger.Printf("Info: Skipping upload for '%s' (hashes match)", cleanRelPath)
			}
		} else {
			s.logger.Printf("Action: Plan upload for new file '%s' (local hash %s...)", cleanRelPath, localHash[:8])
			filesToUpload[cleanRelPath] = localHash
		}
	}

	// 4. Determine Deletions (remains the same logic)
	s.logger.Println("Analyzing remote files for delete actions...")
	for displayName, remoteFile := range remoteFileMap {
		isPotentiallyManaged := false
		for _, relPath := range relativeFiles {
			cleanRelPath := filepath.ToSlash(filepath.Clean(relPath))
			if cleanRelPath == displayName {
				if s.gitIgnore == nil || !s.gitIgnore.MatchesPath(cleanRelPath) {
					isPotentiallyManaged = true
				}
				break
			}
		}
		if !isPotentiallyManaged {
			s.logger.Printf("Info: Skipping deletion check for remote file '%s' (Name: %s) as its display name wasn't passed or was gitignored.", displayName, remoteFile.Name)
			continue
		}
		if _, existsLocally := localFileSet[displayName]; !existsLocally {
			s.logger.Printf("Action: Plan delete for remote file '%s' (Name: %s) as it's not present or valid locally.", displayName, remoteFile.Name)
			filesToDelete = append(filesToDelete, remoteFile)
		}
	}

	// Report analysis errors (remains the same)
	if len(errorsEncountered) > 0 {
		return fmt.Errorf("analysis phase encountered %d error(s): %s", len(errorsEncountered), strings.Join(errorsEncountered, "; "))
	}

	// 5. Execute Actions
	var wg sync.WaitGroup
	errChan := make(chan error, len(filesToUpload)+len(filesToDelete)+1)

	if !s.config.DryRun {
		s.logger.Println("--- EXECUTING ACTIONS ---")
		// Perform Uploads
		for relPath := range filesToUpload {
			wg.Add(1)
			go func(p string) {
				defer wg.Done()
				s.logger.Printf("Executing: Uploading %s", p)
				// Read file content using ExecuteToolCall
				fileContentResult, err := s.interp.ExecuteToolCall("ReadFile", p)
				if err != nil {
					errChan <- fmt.Errorf("failed to read local file %s for upload: %w", p, err)
					return
				}
				// Check type assertion
				fileContentBytes, ok := fileContentResult.([]byte)
				if !ok {
					contentStr, okStr := fileContentResult.(string)
					if !okStr {
						errChan <- fmt.Errorf("tool ReadFile returned unexpected type (%T) for %s", fileContentResult, p)
						return
					}
					fileContentBytes = []byte(contentStr)
				}

				// Upload using the relative path as the DisplayName
				if _, err := s.uploadFile(ctx, p, fileContentBytes); err != nil { // Uses corrected uploadFile
					errChan <- fmt.Errorf("failed to upload %s: %w", p, err)
				} else {
					s.logger.Printf("Success: Uploaded %s", p)
				}
			}(relPath)
		}

		// Perform Deletions (logic remains the same, uses corrected deleteFile)
		for _, rf := range filesToDelete {
			wg.Add(1)
			go func(f RemoteFile) {
				defer wg.Done()
				s.logger.Printf("Executing: Deleting remote %s (Name: %s)", f.DisplayName, f.Name)
				if err := s.deleteFile(ctx, f.Name); err != nil { // Uses corrected deleteFile
					errChan <- fmt.Errorf("failed to delete %s (Name: %s): %w", f.DisplayName, f.Name, err)
				} else {
					s.logger.Printf("Success: Deleted remote %s (Name: %s)", f.DisplayName, f.Name)
				}
			}(rf)
		}

		wg.Wait()
		close(errChan)

		// Collect and report execution errors (remains the same)
		var syncErrors []string
		for err := range errChan {
			s.logger.Printf("Error during sync execution: %v", err)
			syncErrors = append(syncErrors, err.Error())
		}
		if len(syncErrors) > 0 {
			return fmt.Errorf("sync execution encountered %d error(s): %s", len(syncErrors), strings.Join(syncErrors, "; "))
		}
		s.logger.Println("--- EXECUTION COMPLETE ---")
	} else {
		s.logger.Println("--- DRY RUN COMPLETE (no changes made) ---")
	}

	s.logger.Println("Sync process finished.")
	return nil
}

// ClearRemote deletes all files managed by the API key.
func (s *Syncer) ClearRemote(ctx context.Context) error {
	// Logic remains the same, relies on corrected listRemoteFiles and deleteFile
	s.logger.Println("Starting remote file clearing process...")
	if s.config.DryRun {
		s.logger.Println("--- DRY RUN MODE ---")
	}

	remoteFiles, err := s.listRemoteFiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to list remote files for clearing: %w", err)
	}
	if len(remoteFiles) == 0 {
		s.logger.Println("No remote files found to clear.")
		return nil
	}
	s.logger.Printf("Found %d remote files to potentially delete.", len(remoteFiles))

	var wg sync.WaitGroup
	errChan := make(chan error, len(remoteFiles)+1)

	for _, rf := range remoteFiles {
		if s.config.DryRun {
			s.logger.Printf("Dry Run: Would delete remote file Name: %s (DisplayName: '%s', State: %s)", rf.Name, rf.DisplayName, rf.State)
		} else {
			wg.Add(1)
			go func(f RemoteFile) {
				defer wg.Done()
				s.logger.Printf("Executing: Deleting remote %s (Name: %s)", f.DisplayName, f.Name)
				if err := s.deleteFile(ctx, f.Name); err != nil { // Uses corrected deleteFile
					errChan <- fmt.Errorf("failed to delete %s (Name: %s): %w", f.DisplayName, f.Name, err)
				} else {
					s.logger.Printf("Success: Deleted remote %s (Name: %s)", f.DisplayName, f.Name)
				}
			}(rf)
		}
	}

	if !s.config.DryRun {
		wg.Wait()
		close(errChan)
		var clearErrors []string
		for err := range errChan {
			s.logger.Printf("Error during clear remote execution: %v", err)
			clearErrors = append(clearErrors, err.Error())
		}
		if len(clearErrors) > 0 {
			return fmt.Errorf("clear remote encountered %d error(s): %s", len(clearErrors), strings.Join(clearErrors, "; "))
		}
		s.logger.Println("Successfully cleared all detected remote files.")
	} else {
		s.logger.Println("--- DRY RUN COMPLETE (no files deleted) ---")
	}

	return nil
}

// --- API Interaction Helpers (Using FileService) ---

// listRemoteFiles uses the FileService client.
func (s *Syncer) listRemoteFiles(ctx context.Context) ([]RemoteFile, error) {
	s.logger.Println("Listing remote files via API...")
	var remoteFiles []RemoteFile
	// Use FileService client
	iter := s.fs.ListFiles(ctx) // Use FileService
	for {
		file, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed during remote file iteration: %w", err)
		}
		if file == nil {
			s.logger.Printf("Warning: Iterator returned nil file, skipping")
			continue
		}
		// Conversion logic remains the same
		remoteFiles = append(remoteFiles, RemoteFile{
			Name:        file.Name,
			DisplayName: file.DisplayName,
			SizeBytes:   file.SizeBytes,
			SHA256Hash:  fmt.Sprintf("%x", file.SHA256Hash), // Ensure hex format
			State:       file.State.String(),
			UpdateTime:  file.UpdateTime,
			URI:         file.URI,
		})
		s.logger.Printf("Debug: Found remote file Name: %s, DisplayName: '%s', State: %s, Size: %d, Hash: %s...",
			file.Name, file.DisplayName, file.State.String(), file.SizeBytes, fmt.Sprintf("%x", file.SHA256Hash)[:8])
	}
	s.logger.Printf("Finished listing remote files. Total found: %d", len(remoteFiles))
	return remoteFiles, nil
}

// uploadFile uses the FileService client.
func (s *Syncer) uploadFile(ctx context.Context, displayName string, content []byte) (*genai.File, error) {
	s.logger.Printf("API: Uploading content for DisplayName '%s' (%d bytes)", displayName, len(content))

	// Pass DisplayName directly as the third argument to fs.UploadFile
	// Correct usage: Use strings.NewReader for the content (io.Reader).
	uploadedFile, err := s.fs.UploadFile(ctx, "", displayName, strings.NewReader(string(content))) // Pass displayName & reader
	if err != nil {
		return nil, fmt.Errorf("api upload call failed for %s: %w", displayName, err)
	}
	if uploadedFile == nil {
		return nil, fmt.Errorf("api upload for %s returned a nil file object unexpectedly", displayName)
	}

	s.logger.Printf("API: Upload successful for '%s'. Remote Name: %s, State: %s", displayName, uploadedFile.Name, uploadedFile.State.String())
	return uploadedFile, nil
}

// deleteFile uses the FileService client.
func (s *Syncer) deleteFile(ctx context.Context, remoteFileName string) error {
	s.logger.Printf("API: Requesting deletion of remote file: %s", remoteFileName)
	// Use FileService client
	err := s.fs.DeleteFile(ctx, remoteFileName) // Use FileService
	if err != nil {
		return fmt.Errorf("api delete call failed for %s: %w", remoteFileName, err)
	}
	s.logger.Printf("API: Deletion request successful for %s", remoteFileName)
	return nil
}

// --- Utility Functions ---

// FindFiles uses TOOL.WalkDir and applies filters.
func FindFiles(interp *core.Interpreter, baseDirRel string, includes, excludes []string, gitIgnore *gitignore.GitIgnore, logger *log.Logger) ([]string, error) {
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}
	if interp == nil {
		return nil, fmt.Errorf("FindFiles requires a valid interpreter")
	}

	logger.Printf("FindFiles: Walking directory '%s' using TOOL.WalkDir", baseDirRel)

	// Call the WalkDir tool using ExecuteToolCall
	walkResultIntf, err := interp.ExecuteToolCall("WalkDir", baseDirRel) // Use ExecuteToolCall
	if err != nil {
		logger.Printf("Error: TOOL.WalkDir failed for '%s': %v", baseDirRel, err)
		if errors.Is(err, core.ErrPathViolation) {
			return nil, fmt.Errorf("walking directory failed: %w", core.ErrPathViolation)
		}
		// Check if the error indicates the start path didn't exist
		// Note: This depends on how WalkDir wraps os.ErrNotExist. Checking substring might be needed.
		if errors.Is(err, os.ErrNotExist) || errors.Is(err, core.ErrFileNotFound) || strings.Contains(err.Error(), "Start path not found") {
			logger.Printf("FindFiles: TOOL.WalkDir reported start path '%s' not found.", baseDirRel)
			return []string{}, nil // Not an error for FindFiles, just no files found
		}
		return nil, fmt.Errorf("failed to walk directory '%s': %w", baseDirRel, err)
	}

	// Process result (Type assertion)
	if walkResultIntf == nil {
		logger.Printf("FindFiles: TOOL.WalkDir returned nil for '%s' (directory likely doesn't exist or is empty).", baseDirRel)
		return []string{}, nil
	}

	// Type assertion logic remains the same
	var walkResultList []map[string]interface{}
	switch v := walkResultIntf.(type) {
	case []map[string]interface{}:
		walkResultList = v
	case []interface{}:
		for _, item := range v {
			if itemMap, ok := item.(map[string]interface{}); ok {
				walkResultList = append(walkResultList, itemMap)
			} else {
				logger.Printf("Error: TOOL.WalkDir result list contained unexpected item type: %T", item)
				return nil, fmt.Errorf("unexpected item type in WalkDir result: %w", core.ErrInternalTool)
			}
		}
	default:
		logger.Printf("Error: TOOL.WalkDir returned unexpected type: %T", walkResultIntf)
		return nil, fmt.Errorf("unexpected result type from WalkDir: %w", core.ErrInternalTool)
	}

	logger.Printf("FindFiles: TOOL.WalkDir returned %d entries for '%s'. Filtering...", len(walkResultList), baseDirRel)

	// Filtering logic (remains the same)
	var files []string
	effectiveIncludes := includes
	if len(effectiveIncludes) == 0 {
		effectiveIncludes = []string{"**/*"}
	}

	for _, entryMap := range walkResultList {
		relPathIntf, okPath := entryMap["path"]
		isDirIntf, okIsDir := entryMap["isDir"]
		if !okPath || !okIsDir {
			logger.Printf("Warning: Skipping entry due to missing 'path' or 'isDir' key: %+v", entryMap)
			continue
		}
		relPath, okPathStr := relPathIntf.(string)
		isDir, okIsDirBool := isDirIntf.(bool)
		if !okPathStr || !okIsDirBool {
			logger.Printf("Warning: Skipping entry due to unexpected type for 'path' (%T) or 'isDir' (%T): %+v", relPathIntf, isDirIntf, entryMap)
			continue
		}
		relPathSlash := filepath.ToSlash(relPath)
		if isDir {
			continue // Skip directories
		}

		// Exclude patterns
		excluded := false
		for _, pattern := range excludes {
			match, _ := doublestar.Match(pattern, relPathSlash)
			if match {
				logger.Printf("Debug: Path '%s' matches exclude pattern '%s'", relPathSlash, pattern)
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		// Gitignore
		if gitIgnore != nil && gitIgnore.MatchesPath(relPathSlash) {
			logger.Printf("Debug: Skipping gitignored file: %s", relPathSlash)
			continue
		}

		// Include patterns
		included := false
		for _, pattern := range effectiveIncludes {
			match, _ := doublestar.Match(pattern, relPathSlash)
			if match {
				logger.Printf("Debug: File '%s' matches include pattern '%s'", relPathSlash, pattern)
				included = true
				break
			}
		}
		if included {
			logger.Printf("Debug: Adding included file: %s", relPathSlash)
			files = append(files, relPathSlash)
		} else {
			logger.Printf("Debug: Skipping file '%s' (did not match include patterns)", relPathSlash)
		}
	}

	logger.Printf("FindFiles finished. Found %d matching files after filtering.", len(files))
	return files, nil
}
