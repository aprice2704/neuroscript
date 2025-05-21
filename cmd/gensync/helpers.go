package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings" // <-- Added for strings.NewReader
	"time"

	"github.com/google/generative-ai-go/genai"
)

// Pre-calculate the SHA256 hash of a single space (" ")
const emptyFileContent = " " // Use a single space for empty file representation
var emptyFileHash string     // Will be calculated in init()

func init() {
	hasher := sha256.New()
	hasher.Write([]byte(emptyFileContent))
	emptyFileHash = hex.EncodeToString(hasher.Sum(nil))
}

// IndexEntry stores metadata about a synced file
type IndexEntry struct {
	RelativePath string    `json:"relativePath"`
	ApiFileName  string    `json:"apiFileName"`
	LocalModTime int64     `json:"localModTime"`
	LocalHash    string    `json:"localHash"` // Will store emptyFileHash for zero-byte files
	MimeType     string    `json:"mimeType"`
	LastSyncTime time.Time `json:"lastSyncTime"`
}

// SyncStats tracks counts during the sync operation
type SyncStats struct {
	FilesProcessed      int
	FilesIgnored        int
	FilesUpToDate       int
	FilesUploaded       int
	FilesUpdated        int
	FilesDeletedLocally int
	FilesDeletedAPI     int
	UploadErrors        int
	DeleteErrors        int
}

// LoadIndex reads the index file from the specified root directory
func LoadIndex(rootDir string) (map[string]IndexEntry, error) {
	indexPath := filepath.Join(rootDir, indexFileName)
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]IndexEntry), nil // Not an error if file doesn't exist
		}
		return nil, fmt.Errorf("reading index file %s: %w", indexPath, err)
	}

	if len(data) == 0 { // Handle empty file case
		return make(map[string]IndexEntry), nil
	}

	var indexMap map[string]IndexEntry
	err = json.Unmarshal(data, &indexMap)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling index file %s: %w", indexPath, err)
	}
	return indexMap, nil
}

// SaveIndex writes the index map back to the index file atomically
func SaveIndex(rootDir string, indexMap map[string]IndexEntry) error {
	indexPath := filepath.Join(rootDir, indexFileName)
	data, err := json.MarshalIndent(indexMap, "", "  ") // Pretty print
	if err != nil {
		return fmt.Errorf("marshalling index: %w", err)
	}

	// Write atomically (write to temp, then rename)
	tempFile := indexPath + ".tmp"
	err = os.WriteFile(tempFile, data, 0644)
	if err != nil {
		return fmt.Errorf("writing temp index file %s: %w", tempFile, err)
	}
	defer os.Remove(tempFile) // Ensure temp file is removed even if rename fails later

	err = os.Rename(tempFile, indexPath)
	if err != nil {
		return fmt.Errorf("renaming temp index file to %s: %w", tempFile, err)
	}
	return nil
}

// CalculateFileHash computes the SHA-256 hash of a file's content
// For zero-byte files, returns the pre-calculated hash of emptyFileContent
func CalculateFileHash(filePath string) (string, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("stat failed for hashing %s: %w", filePath, err)
	}

	// --- Handle empty file ---
	if fileInfo.Size() == 0 {
		return emptyFileHash, nil
	}
	// --- End handle empty file ---

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("opening file for hashing %s: %w", filePath, err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("copying file data for hashing %s: %w", filePath, err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// UploadFile handles uploading a single file and waiting for it to become active
// For zero-byte files, uploads a single space instead.
func UploadFile(ctx context.Context, client *genai.Client, localPath, relativePath string) (*genai.File, error) {
	// Get file info first
	fileInfo, err := os.Stat(localPath)
	if err != nil {
		return nil, fmt.Errorf("stat failed for upload %s: %w", localPath, err)
	}
	isZeroByte := fileInfo.Size() == 0

	// Determine MIME type - force text/plain for our empty file workaround
	var mimeType string
	if isZeroByte {
		mimeType = "text/plain" // Use text/plain for the single space content
	} else {
		mimeType = mime.TypeByExtension(filepath.Ext(localPath))
		if mimeType == "" {
			mimeType = "application/octet-stream" // Default fallback
		}
	}

	// Use the relative path (cleaned) as the DisplayName
	uploadDisplayName := relativePath

	// Create the options struct pointer
	options := &genai.UploadFileOptions{
		MIMEType:    mimeType,
		DisplayName: uploadDisplayName,
	}

	// Prepare reader: either file content or minimal content for empty files
	var reader io.Reader
	if isZeroByte {
		log.Printf("  Handling zero-byte file %q by uploading minimal content.\n", relativePath)
		reader = strings.NewReader(emptyFileContent)
	} else {
		fileReader, err := os.Open(localPath) // Only open if not zero-byte
		if err != nil {
			return nil, fmt.Errorf("opening local file %s: %w", localPath, err)
		}
		defer fileReader.Close() // Ensure closure if opened
		reader = fileReader
	}

	// Call UploadFile with the appropriate reader
	apiFile, err := client.UploadFile(ctx, "", reader, options)

	if err != nil {
		return nil, fmt.Errorf("API upload call failed for %q: %w", relativePath, err)
	}
	log.Printf("  Upload initiated for %q -> API Name: %s, DisplayName: %q\n", relativePath, apiFile.Name, apiFile.DisplayName)

	// --- Wait for ACTIVE state (same logic as before) ---
	startTime := time.Now()
	pollInterval := 2 * time.Second
	const maxPollInterval = 15 * time.Second
	const timeout = 3 * time.Minute

	for apiFile.State == genai.FileStateProcessing {
		if time.Since(startTime) > timeout {
			log.Printf("  ERROR: File %s (API: %s) timed out in processing state after %v. Attempting delete.", relativePath, apiFile.Name, timeout)
			_ = client.DeleteFile(context.Background(), apiFile.Name)
			return nil, fmt.Errorf("file %s (API: %s) stuck in processing state", relativePath, apiFile.Name)
		}
		time.Sleep(pollInterval)
		updatedFile, err := client.GetFile(ctx, apiFile.Name)
		if err != nil {
			log.Printf("  Warning: Failed to get file status during processing check for %s: %v. Assuming failure.", apiFile.Name, err)
			return nil, fmt.Errorf("checking processing status failed for %s: %w", apiFile.Name, err)
		}
		apiFile = updatedFile
		pollInterval *= 2
		if pollInterval > maxPollInterval {
			pollInterval = maxPollInterval
		}
	}

	if apiFile.State != genai.FileStateActive {
		errMsg := fmt.Sprintf("file %s finished processing but is not ACTIVE (State: %s)", relativePath, apiFile.State)
		log.Printf("  ERROR: %s (API: %s)", errMsg, apiFile.Name)
		_ = client.DeleteFile(context.Background(), apiFile.Name)
		// --- FIX: Use constant format string ---
		return nil, fmt.Errorf("%s", errMsg) // Was: return nil, fmt.Errorf(errMsg)
		// --- END FIX ---
	}

	return apiFile, nil
}
