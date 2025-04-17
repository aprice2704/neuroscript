package nssync

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs" // Use io/fs for WalkDir
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time" // Used by genai.File

	"github.com/bmatcuk/doublestar/v4"
	gitignore "github.com/sabhiram/go-gitignore" // Alias for clarity
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	// Use the correct package import path based on your successful 'go get'
	genai "google.golang.org/genai"
)

// RemoteFile represents a file in the Gemini Files API storage.
// We extract relevant fields from the genai.File struct.
type RemoteFile struct {
	Name        string    // API's unique identifier (e.g., "files/abc123def")
	DisplayName string    // User-provided name (should match relative local path)
	SizeBytes   int64     // Size in bytes
	SHA256Hash  string    // SHA256 hash provided by the API (as hex string)
	State       string    // Processing state (e.g., "ACTIVE", "PROCESSING")
	UpdateTime  time.Time // Last update time from API
	URI         string    // URI for use in prompts
}

// SyncerConfig holds the configuration for the synchronization process.
type SyncerConfig struct {
	LocalDir        string
	APIKey          string
	DryRun          bool
	IgnoreGitignore bool
}

// Syncer handles the synchronization logic.
type Syncer struct {
	config     SyncerConfig
	client     *genai.Client // Use the genai client
	gitIgnore  *gitignore.GitIgnore
	baseAbsDir string // Absolute path to the local directory base
	logger     *log.Logger
}

// NewSyncer creates a new Syncer instance.
func NewSyncer(config SyncerConfig, logger *log.Logger) (*Syncer, error) {
	if logger == nil {
		logger = log.New(io.Discard, "", 0) // Default to discard if nil
	}

	// 1. Validate LocalDir is below CWD
	if err := checkPathSafety(config.LocalDir); err != nil {
		return nil, fmt.Errorf("path safety check failed: %w", err)
	}

	absDir, err := filepath.Abs(config.LocalDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %w", config.LocalDir, err)
	}

	// 2. Initialize Gemini Client using google.golang.org/genai
	ctx := context.Background() // Or use context from caller

	// Correct way to initialize the client with just an API key
	client, err := genai.NewClient(ctx, option.WithAPIKey(config.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}
	// Consider adding client.Close() in a cleanup phase if the Syncer lifecycle allows

	s := &Syncer{
		config:     config,
		client:     client, // Store the genai.Client
		baseAbsDir: absDir,
		logger:     logger,
	}

	// 3. Load .gitignore if needed
	if !config.IgnoreGitignore {
		gitignorePath := filepath.Join(absDir, ".gitignore")
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

// --- Core Sync Logic ---

// Sync synchronizes the provided list of relative local file paths with the remote API.
func (s *Syncer) Sync(ctx context.Context, relativeFiles []string) error {
	s.logger.Printf("Starting sync for %d potential local files...", len(relativeFiles))
	if s.config.DryRun {
		s.logger.Println("--- DRY RUN MODE ---")
	}

	// 1. Get Remote Files and build a map keyed by DisplayName
	remoteFiles, err := s.listRemoteFiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to list remote files: %w", err)
	}
	remoteFileMap := make(map[string]RemoteFile) // Map DisplayName -> RemoteFile
	for _, rf := range remoteFiles {
		// Only consider files that are processed and ready ("ACTIVE")
		// and have a DisplayName for mapping.
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

	// 2. Prepare tracking maps/slices
	localFileSet := make(map[string]bool)    // Tracks files passed in relativeFiles arg
	filesToUpload := make(map[string]string) // Map relative path -> local SHA256 hash
	filesToDelete := []RemoteFile{}          // List of RemoteFile structs to delete

	// 3. Determine Uploads: Iterate through provided local files
	s.logger.Println("Analyzing local files for upload actions...")
	for _, relPath := range relativeFiles {
		cleanRelPath := filepath.ToSlash(filepath.Clean(relPath))
		if cleanRelPath == "." || cleanRelPath == "" {
			continue
		}
		localFileSet[cleanRelPath] = true // Mark this path as managed locally

		absPath := filepath.Join(s.baseAbsDir, cleanRelPath)

		// Apply gitignore if loaded
		if s.gitIgnore != nil && s.gitIgnore.MatchesPath(cleanRelPath) {
			s.logger.Printf("Info: Skipping '%s' due to .gitignore", cleanRelPath)
			continue
		}

		// Check local file existence and type
		stat, err := os.Stat(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				s.logger.Printf("Warning: Local file '%s' targeted for sync not found, skipping.", cleanRelPath)
			} else {
				s.logger.Printf("Warning: Cannot stat local file '%s': %v", cleanRelPath, err)
			}
			continue
		}
		if stat.IsDir() {
			s.logger.Printf("Debug: Skipping directory '%s'", cleanRelPath)
			continue
		}

		// Calculate local file hash
		localHash, err := getFileHash(absPath)
		if err != nil {
			s.logger.Printf("Warning: Failed to hash local file '%s': %v. Skipping upload.", cleanRelPath, err)
			continue
		}

		// Compare with remote file (if exists and active)
		remoteFile, existsRemotely := remoteFileMap[cleanRelPath]
		if existsRemotely {
			// Remote file with the same DisplayName exists and is ACTIVE
			if remoteFile.SHA256Hash != localHash {
				s.logger.Printf("Action: Plan upload for '%s' (local hash %s... != remote hash %s...)", cleanRelPath, localHash[:8], remoteFile.SHA256Hash[:8])
				filesToUpload[cleanRelPath] = localHash // Mark for upload
			} else {
				s.logger.Printf("Info: Skipping upload for '%s' (hashes match)", cleanRelPath)
			}
		} else {
			// Remote file doesn't exist (or isn't ACTIVE)
			s.logger.Printf("Action: Plan upload for new file '%s' (local hash %s...)", cleanRelPath, localHash[:8])
			filesToUpload[cleanRelPath] = localHash // Mark for upload
		}
	}

	// 4. Determine Deletions: Iterate through active remote files
	s.logger.Println("Analyzing remote files for delete actions...")
	for displayName, remoteFile := range remoteFileMap {
		isPotentiallyManaged := false
		for _, relPath := range relativeFiles {
			cleanRelPath := filepath.ToSlash(filepath.Clean(relPath))
			if cleanRelPath == displayName {
				isPotentiallyManaged = true
				break
			}
		}

		if !isPotentiallyManaged {
			s.logger.Printf("Info: Skipping deletion check for remote file '%s' (Name: %s) as its display name doesn't correspond to any file passed to this sync operation.", displayName, remoteFile.Name)
			continue
		}

		if _, existsLocally := localFileSet[displayName]; !existsLocally {
			// It exists remotely (and is ACTIVE), but not in the final local set. Delete it.
			s.logger.Printf("Action: Plan delete for remote file '%s' (Name: %s) as it's not present or ignored locally.", displayName, remoteFile.Name)
			filesToDelete = append(filesToDelete, remoteFile)
		}
	}

	// 5. Execute Actions (if not DryRun)
	var wg sync.WaitGroup
	errChan := make(chan error, len(filesToUpload)+len(filesToDelete))

	if !s.config.DryRun {
		s.logger.Println("--- EXECUTING ACTIONS ---")
		// Perform Uploads (concurrently)
		for relPath := range filesToUpload {
			wg.Add(1)
			go func(p string) {
				defer wg.Done()
				s.logger.Printf("Executing: Uploading %s", p)
				absPath := filepath.Join(s.baseAbsDir, p)
				// Upload using the relative path as the DisplayName
				if _, err := s.uploadFile(ctx, absPath, p); err != nil {
					errChan <- fmt.Errorf("failed to upload %s: %w", p, err)
				} else {
					s.logger.Printf("Success: Uploaded %s", p)
				}
			}(relPath)
		}

		// Perform Deletions (concurrently)
		for _, rf := range filesToDelete {
			wg.Add(1)
			go func(f RemoteFile) {
				defer wg.Done()
				s.logger.Printf("Executing: Deleting remote %s (Name: %s)", f.DisplayName, f.Name)
				if err := s.deleteFile(ctx, f.Name); err != nil {
					errChan <- fmt.Errorf("failed to delete %s (Name: %s): %w", f.DisplayName, f.Name, err)
				} else {
					s.logger.Printf("Success: Deleted remote %s (Name: %s)", f.DisplayName, f.Name)
				}
			}(rf)
		}

		wg.Wait() // Wait for all uploads and deletes to finish
		close(errChan)

		// Collect and report errors
		var syncErrors []string
		for err := range errChan {
			s.logger.Printf("Error during sync execution: %v", err)
			syncErrors = append(syncErrors, err.Error())
		}
		if len(syncErrors) > 0 {
			return fmt.Errorf("sync encountered %d error(s): %s", len(syncErrors), strings.Join(syncErrors, "; "))
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
	s.logger.Println("Starting remote file clearing process...")
	if s.config.DryRun {
		s.logger.Println("--- DRY RUN MODE ---")
	}

	// List *all* files, regardless of state.
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
	errChan := make(chan error, len(remoteFiles))

	for _, rf := range remoteFiles {
		if s.config.DryRun {
			s.logger.Printf("Dry Run: Would delete remote file Name: %s (DisplayName: '%s', State: %s)", rf.Name, rf.DisplayName, rf.State)
		} else {
			wg.Add(1)
			go func(f RemoteFile) {
				defer wg.Done()
				s.logger.Printf("Executing: Deleting remote %s (Name: %s)", f.DisplayName, f.Name)
				// Delete using the unique Name identifier
				if err := s.deleteFile(ctx, f.Name); err != nil {
					errChan <- fmt.Errorf("failed to delete %s (Name: %s): %w", f.DisplayName, f.Name, err)
				} else {
					s.logger.Printf("Success: Deleted remote %s (Name: %s)", f.DisplayName, f.Name)
				}
			}(rf)
		}
	}

	if !s.config.DryRun {
		wg.Wait() // Wait for all deletions
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

// --- API Interaction Helpers ---

// listRemoteFiles fetches the list of files from the Gemini API using the genai.Client.
// According to documentation, ListFiles is a method directly on the client.
func (s *Syncer) listRemoteFiles(ctx context.Context) ([]RemoteFile, error) {
	s.logger.Println("Listing remote files via API...")
	var remoteFiles []RemoteFile

	// Use the ListFiles method directly from the genai client
	iter := s.client.ListFiles(ctx) // This returns *FileIterator

	for {
		file, err := iter.Next()
		if err == iterator.Done {
			break // Finished iterating
		}
		if err != nil {
			return nil, fmt.Errorf("failed during remote file iteration: %w", err)
		}
		if file == nil { // Defensive check
			s.logger.Printf("Warning: Iterator returned nil file, skipping")
			continue
		}

		// Convert the genai.File struct to our internal RemoteFile struct
		remoteFiles = append(remoteFiles, RemoteFile{
			Name:        file.Name,
			DisplayName: file.DisplayName,
			SizeBytes:   file.SizeBytes,
			SHA256Hash:  fmt.Sprintf("%x", file.SHA256Hash), // Convert []byte hash to hex string
			State:       file.State.String(),                // Convert enum state to string
			UpdateTime:  file.UpdateTime,
			URI:         file.URI,
		})
		s.logger.Printf("Debug: Found remote file Name: %s, DisplayName: '%s', State: %s, Size: %d, Hash: %s...",
			file.Name, file.DisplayName, file.State.String(), file.SizeBytes, fmt.Sprintf("%x", file.SHA256Hash)[:8])
	}

	s.logger.Printf("Finished listing remote files. Total found: %d", len(remoteFiles))
	return remoteFiles, nil
}

// uploadFile uploads a single local file to the Gemini API using genai.Client.
// According to documentation, UploadFile is a method directly on the client,
// and UploadFileOptions is a type defined in the package.
func (s *Syncer) uploadFile(ctx context.Context, absPath string, displayName string) (*genai.File, error) {
	fileReader, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open local file %s: %w", absPath, err)
	}
	defer fileReader.Close() // Ensure file is closed

	s.logger.Printf("API: Uploading '%s' with DisplayName '%s'", absPath, displayName)

	// Prepare options for the upload using the genai.UploadFileOptions type.
	// This struct allows specifying DisplayName and MimeType.
	opts := &genai.UploadFileOptions{
		DisplayName: displayName,
		// Optionally set MimeType:
		// MimeType: mime.TypeByExtension(filepath.Ext(absPath)),
	}

	// Call UploadFile directly on the client.
	// Pass "" for the optional 'name' argument to let the API generate a unique resource name.
	// Pass the file reader and the options struct.
	uploadedFile, err := s.client.UploadFile(ctx, "", fileReader, opts)
	if err != nil {
		return nil, fmt.Errorf("api upload call failed for %s: %w", displayName, err)
	}
	if uploadedFile == nil {
		return nil, fmt.Errorf("api upload for %s returned a nil file object unexpectedly", displayName)
	}

	s.logger.Printf("API: Upload successful for '%s'. Remote Name: %s, State: %s", displayName, uploadedFile.Name, uploadedFile.State.String())

	return uploadedFile, nil
}

// deleteFile deletes a file from the Gemini API using its unique name via genai.Client.
// According to documentation, DeleteFile is a method directly on the client.
func (s *Syncer) deleteFile(ctx context.Context, remoteFileName string) error {
	s.logger.Printf("API: Requesting deletion of remote file: %s", remoteFileName)

	// Call DeleteFile directly on the client using the unique file Name.
	err := s.client.DeleteFile(ctx, remoteFileName)
	if err != nil {
		// Consider checking for specific errors like "not found" if needed.
		return fmt.Errorf("api delete call failed for %s: %w", remoteFileName, err)
	}

	s.logger.Printf("API: Deletion request successful for %s", remoteFileName)
	return nil
}

// --- Utility Functions ---

// checkPathSafety ensures the target directory is within the current working directory.
func checkPathSafety(targetDir string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}
	absTarget, err := filepath.Abs(targetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for target %s: %w", targetDir, err)
	}
	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for cwd %s: %w", cwd, err)
	}

	if !strings.HasPrefix(absTarget+string(filepath.Separator), absCwd+string(filepath.Separator)) {
		return nil, fmt.Errorf("target directory '%s' is not inside the current working directory '%s'", absTarget, absCwd)
	}
	if absTarget == absCwd {
		return nil, fmt.Errorf("syncing the current working directory itself ('%s') is not allowed by default", absCwd)
	}
	return nil
}

// getFileHash calculates the SHA256 hash of a file and returns it as a hex string.
func getFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s for hashing: %w", filePath, err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to read file %s for hashing: %w", filePath, err)
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// FindFiles walks the directory relative to baseDir and applies glob filters.
func FindFiles(baseDir string, includes, excludes []string, logger *log.Logger) ([]string, error) {
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}

	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("cannot get absolute path for %s: %w", baseDir, err)
	}
	var files []string

	effectiveIncludes := includes
	if len(effectiveIncludes) == 0 {
		effectiveIncludes = []string{"**/*"}
		logger.Printf("Debug: No include patterns provided, defaulting to: %v", effectiveIncludes)
	} else {
		logger.Printf("Debug: Using include patterns: %v", effectiveIncludes)
	}
	logger.Printf("Debug: Using exclude patterns: %v", excludes)

	walkRoot := os.DirFS(absBaseDir)

	err = fs.WalkDir(walkRoot, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			logger.Printf("Warning: Error accessing path %s: %v", path, err)
			return nil
		}

		relPathSlash := filepath.ToSlash(path)
		if relPathSlash == "." || relPathSlash == "" {
			return nil
		}

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
			if d.IsDir() {
				logger.Printf("Debug: Skipping excluded directory: %s", relPathSlash)
				return fs.SkipDir
			}
			logger.Printf("Debug: Skipping excluded file: %s", relPathSlash)
			return nil
		}

		if d.IsDir() {
			return nil
		}

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

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory %s: %w", baseDir, err)
	}

	logger.Printf("FindFiles finished. Found %d matching files.", len(files))
	return files, nil
}
