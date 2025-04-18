package main

import (
	"bufio"
	"context"

	// "encoding/json" // In helpers
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync" // <-- Added for WaitGroup
	"time"

	"github.com/google/generative-ai-go/genai"
	gitignore "github.com/sabhiram/go-gitignore"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const indexFileName = ".gemini_sync_index.json"
const maxConcurrentUploads = 8 // Adjust concurrency level as needed

// Represents a file operation needed
type syncJob struct {
	localPath      string
	relativePath   string
	currentModTime int64
	currentHash    string
	existingEntry  *IndexEntry // Pointer, nil if it's a new file upload
}

// Represents the result from a worker
type syncResult struct {
	job       syncJob     // The job this result corresponds to
	newEntry  *IndexEntry // The updated/new entry if successful
	uploadErr error       // Error during upload/update
	deleteErr error       // Error during pre-update delete (if applicable)
}

func main() {
	// --- Command Line Flags ---
	rootDir := flag.String("dir", ".", "The root directory to synchronize")
	ignoreGitignore := flag.Bool("ignore-gitignore", false, "Ignore .gitignore files")
	nuke := flag.Bool("nuke", false, "Delete ALL files from the Gemini API after confirmation.")
	flag.Parse()

	// --- Check for no args/flags ---
	if len(os.Args) == 1 {
		fmt.Println("Error: No directory specified and no other flags provided.")
		fmt.Println("\nUsage of", os.Args[0]+":")
		flag.PrintDefaults()
		os.Exit(1)
	}

	absRootDir, err := filepath.Abs(*rootDir)
	if err != nil {
		log.Fatalf("Error getting absolute path for %s: %v", *rootDir, err)
	}
	if !*nuke {
		log.Printf("Starting sync for directory: %s\n", absRootDir)
	}

	// --- Setup ---
	ctx := context.Background()
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		log.Fatal("API key not found. Please set the GOOGLE_API_KEY environment variable.")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// --- Handle NUKE Operation ---
	if *nuke {
		// ... (Nuke logic remains the same as previous version) ...
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		fmt.Println("!!! WARNING: NUKE option selected.           !!!")
		fmt.Println("!!! This will attempt to delete ALL files    !!!")
		fmt.Println("!!! associated with the current API key.     !!!")
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		fmt.Print("Type 'yes' to confirm deletion: ")

		reader := bufio.NewReader(os.Stdin)
		confirmation, _ := reader.ReadString('\n')
		confirmation = strings.TrimSpace(strings.ToLower(confirmation))

		if confirmation == "yes" {
			log.Println("Proceeding with NUKE operation...")
			nukeAllFiles(ctx, client) // Call the nuke function
			log.Println("NUKE operation finished.")
			os.Exit(0) // Exit after nuke attempt
		} else {
			log.Println("NUKE operation cancelled by user.")
			os.Exit(0) // Exit, do not proceed with sync
		}
	}

	// --- NORMAL SYNC LOGIC STARTS HERE ---
	stats := &SyncStats{}

	// --- Load Local Index ---
	log.Println("Loading local index file:", indexFileName)
	indexMap, err := LoadIndex(absRootDir) // Call helper
	if err != nil {
		log.Printf("Warning: Could not load index file: %v. Starting with empty index.", err)
		indexMap = make(map[string]IndexEntry)
	} else {
		log.Printf("Loaded %d entries from local index.", len(indexMap))
	}

	// --- Initialize Gitignore ---
	var ignorer *gitignore.GitIgnore
	if !*ignoreGitignore {
		ignorer, err = gitignore.CompileIgnoreFile(filepath.Join(absRootDir, ".gitignore"))
		if err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: Could not compile root .gitignore: %v", err)
		} else if ignorer != nil {
			log.Println("Initialized gitignore rules from root.")
		}
	}

	// --- Walk Local Directory & Collect Jobs ---
	localPathsSeen := make(map[string]bool)
	jobsToProcess := []syncJob{} // Collect jobs here

	log.Printf("Scanning local directory and identifying changes: %s\n", absRootDir)
	walkErr := filepath.WalkDir(absRootDir, func(localPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			log.Printf("Warning: Error accessing path %q: %v\n", localPath, walkErr)
			return walkErr
		}
		absLocalPath, _ := filepath.Abs(localPath)
		if absLocalPath == absRootDir {
			return nil
		} // Skip root

		relativePath, err := filepath.Rel(absRootDir, localPath)
		if err != nil {
			return nil
		} // Skip if rel path fails
		relativePath = filepath.ToSlash(relativePath)

		if ignorer != nil && ignorer.MatchesPath(relativePath) {
			stats.FilesIgnored++
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		} // Skip directories

		// --- Process File For Job Collection ---
		stats.FilesProcessed++
		localPathsSeen[relativePath] = true // Mark as seen for later deletion check

		fileInfo, err := d.Info()
		if err != nil {
			return nil
		} // Skip if info fails
		currentModTime := fileInfo.ModTime().UnixNano()
		currentHash, err := CalculateFileHash(localPath) // Call helper
		if err != nil {
			return nil
		} // Skip if hash fails

		entry, existsInIndex := indexMap[relativePath]

		if existsInIndex { // File is known
			if currentModTime == entry.LocalModTime && currentHash == entry.LocalHash {
				stats.FilesUpToDate++
				// Update LastSyncTime? Maybe do it when saving index if needed.
			} else { // File changed, needs update job
				log.Printf("Queueing update: %s\n", relativePath)
				jobsToProcess = append(jobsToProcess, syncJob{
					localPath:      localPath,
					relativePath:   relativePath,
					currentModTime: currentModTime,
					currentHash:    currentHash,
					existingEntry:  &entry, // Pass pointer to existing entry
				})
			}
		} else { // File is new, needs upload job
			log.Printf("Queueing upload: %s\n", relativePath)
			jobsToProcess = append(jobsToProcess, syncJob{
				localPath:      localPath,
				relativePath:   relativePath,
				currentModTime: currentModTime,
				currentHash:    currentHash,
				existingEntry:  nil, // Mark as new
			})
		}
		return nil
	}) // End WalkDir

	if walkErr != nil {
		log.Printf("Warning: File walk finished with error: %v\n", walkErr)
	}

	// --- Process Uploads/Updates Concurrently ---
	if len(jobsToProcess) > 0 {
		log.Printf("Processing %d uploads/updates concurrently (Max: %d)...\n", len(jobsToProcess), maxConcurrentUploads)
		jobsChan := make(chan syncJob, len(jobsToProcess))
		resultsChan := make(chan syncResult, len(jobsToProcess))
		var wg sync.WaitGroup

		// Start workers
		for i := 0; i < maxConcurrentUploads; i++ {
			go worker(ctx, client, &wg, jobsChan, resultsChan)
		}

		// Send jobs
		for _, job := range jobsToProcess {
			wg.Add(1)
			jobsChan <- job
		}
		close(jobsChan) // Signal no more jobs

		// Wait for all jobs to finish in a separate goroutine
		// so we don't block receiving results
		go func() {
			wg.Wait()
			close(resultsChan) // Close results chan when all workers done
			log.Println("All upload/update workers finished.")
		}()

		// Collect results and update index sequentially
		log.Println("Waiting for results and updating index...")
		for result := range resultsChan {
			if result.deleteErr != nil {
				log.Printf("  ERROR (Pre-update Delete): Failed to delete old API file for %s: %v\n", result.job.relativePath, result.deleteErr)
				stats.DeleteErrors++
				// Note: Upload might have proceeded anyway, leading to potential orphans if not handled
			}
			if result.uploadErr != nil {
				log.Printf("  ERROR (Upload/Update): Failed job for %s: %v\n", result.job.relativePath, result.uploadErr)
				stats.UploadErrors++
			} else if result.newEntry != nil {
				// Success! Update the index map
				indexMap[result.job.relativePath] = *result.newEntry
				if result.job.existingEntry == nil {
					stats.FilesUploaded++
					log.Printf("  Successfully uploaded new file: %s -> %s\n", result.job.relativePath, result.newEntry.ApiFileName)
				} else {
					stats.FilesUpdated++
					log.Printf("  Successfully uploaded update: %s -> %s\n", result.job.relativePath, result.newEntry.ApiFileName)
				}
			}
		}
		log.Println("Finished processing results.")

	} else {
		log.Println("No file uploads or updates needed.")
	}

	// --- Process Deletions (Sequential) ---
	log.Println("Checking for files deleted locally...")
	pathsToDeleteFromIndex := []string{}
	for relativePath, entry := range indexMap {
		if !localPathsSeen[relativePath] {
			// File in index was not seen locally -> deleted locally
			// Log details before attempting delete
			log.Printf("Detected local deletion: %s (Was API File: %s)\n", relativePath, entry.ApiFileName)
			stats.FilesDeletedLocally++ // Count intent to delete
			if entry.ApiFileName != "" {
				// Queue for deletion AFTER loop, avoid modifying map while iterating
				pathsToDeleteFromIndex = append(pathsToDeleteFromIndex, relativePath)
			} else {
				log.Printf("  Warning: No API filename found in index for locally deleted file %s, skipping API delete.", relativePath)
			}
		}
	}

	// Perform actual API deletions and index map updates
	if len(pathsToDeleteFromIndex) > 0 {
		log.Printf("Attempting %d API deletions...", len(pathsToDeleteFromIndex))
		for _, relativePath := range pathsToDeleteFromIndex {
			entry := indexMap[relativePath] // Get entry again
			log.Printf("Deleting API file: %s for locally removed %s\n", entry.ApiFileName, relativePath)
			err := client.DeleteFile(ctx, entry.ApiFileName)
			if err != nil {
				log.Printf("  ERROR: Failed to delete API file %s: %v\n", entry.ApiFileName, err)
				stats.DeleteErrors++
				// Keep in index map for next run if delete fails? Or remove anyway?
				// Let's remove from index map even if API delete fails to avoid retrying delete every time.
				delete(indexMap, relativePath)
			} else {
				log.Printf("  Successfully deleted API file: %s\n", entry.ApiFileName)
				stats.FilesDeletedAPI++
				delete(indexMap, relativePath) // Remove from index on success
			}
			// Optional small delay
			time.Sleep(50 * time.Millisecond)
		}
		log.Printf("Finished processing %d deletions.\n", len(pathsToDeleteFromIndex))
	} else {
		log.Println("No local deletions found.")
	}

	// --- Save Index ---
	log.Println("Saving updated local index...")
	err = SaveIndex(absRootDir, indexMap) // Call helper
	if err != nil {
		log.Fatalf("FATAL: Failed to save index file: %v", err)
	}
	log.Println("Index saved.")

	// --- Summary ---
	log.Println("--------------------")
	log.Println("Sync Summary:")
	log.Printf("  Local directory scanned: %s\n", absRootDir)
	log.Printf("  Total local files processed (pre-ignore): %d\n", stats.FilesProcessed)
	log.Printf("  Files ignored (.gitignore): %d\n", stats.FilesIgnored)
	log.Printf("  Files up-to-date (no change needed): %d\n", stats.FilesUpToDate)
	log.Printf("  New files uploaded: %d\n", stats.FilesUploaded)
	log.Printf("  Existing files updated: %d\n", stats.FilesUpdated)
	log.Printf("  Files deleted locally (API deletion attempted): %d\n", stats.FilesDeletedLocally)
	log.Printf("  Successful API deletions: %d\n", stats.FilesDeletedAPI)
	log.Printf("  Upload errors: %d\n", stats.UploadErrors)
	log.Printf("  API delete errors: %d\n", stats.DeleteErrors)
	log.Println("--------------------")

	if stats.UploadErrors > 0 || stats.DeleteErrors > 0 {
		log.Println("Sync completed with errors.")
		os.Exit(1)
	}
	log.Println("Sync completed successfully.")

} // end main

// --- Worker Goroutine ---
func worker(ctx context.Context, client *genai.Client, wg *sync.WaitGroup, jobs <-chan syncJob, results chan<- syncResult) {
	for job := range jobs {
		var deleteErr error
		var uploadErr error
		var newApiFile *genai.File = nil // Initialize to nil

		// If updating, delete the old file first
		if job.existingEntry != nil && job.existingEntry.ApiFileName != "" {
			// Log intent inside worker
			// log.Printf("  Worker: Deleting old API file %s for update of %s\n", job.existingEntry.ApiFileName, job.relativePath)
			deleteErr = client.DeleteFile(ctx, job.existingEntry.ApiFileName)
			if deleteErr != nil {
				// Log delete error here or pass it back in result? Pass back.
				// Log temporary failure, main goroutine handles stats.
				log.Printf("  Worker: Pre-update delete failed for %s (API: %s): %v\n", job.relativePath, job.existingEntry.ApiFileName, deleteErr)
			} else {
				// log.Printf("  Worker: Pre-update delete success for %s (API: %s)\n", job.relativePath, job.existingEntry.ApiFileName)
			}
		}

		// Proceed with upload even if delete failed (API might clean up later, or delete succeeds next run)
		newApiFile, uploadErr = UploadFile(ctx, client, job.localPath, job.relativePath) // Call helper

		// Construct result
		res := syncResult{job: job, deleteErr: deleteErr, uploadErr: uploadErr}
		if uploadErr == nil && newApiFile != nil {
			// If upload successful, create the new index entry data
			res.newEntry = &IndexEntry{
				RelativePath: job.relativePath,
				ApiFileName:  newApiFile.Name, // Use the result from UploadFile
				LocalModTime: job.currentModTime,
				LocalHash:    job.currentHash,
				MimeType:     newApiFile.MIMEType, // Use MIME type confirmed by API/helper
				LastSyncTime: time.Now(),
			}
		}
		// Send result back (even if there were errors)
		results <- res
		wg.Done() // Signal this job is done
	}
}

// --- NUKE Function (remains the same) ---
func nukeAllFiles(ctx context.Context, client *genai.Client) {
	log.Println("Fetching list of all files for deletion...")
	iter := client.ListFiles(ctx)
	filesToDelete := []*genai.File{}
	listErrorCount := 0
	for {
		file, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error fetching file list during nuke: %v. Skipping some files potentially.", err)
			listErrorCount++
			continue
		}
		filesToDelete = append(filesToDelete, file)
	}
	totalFilesFound := len(filesToDelete)
	log.Printf("Found %d files to attempt deletion (encountered %d errors during listing).", totalFilesFound, listErrorCount)

	deletedCount := 0
	deleteErrorCount := 0

	if totalFilesFound == 0 {
		log.Println("No files found to delete.")
		return
	}

	for _, file := range filesToDelete {
		log.Printf("Deleting API file: %s (DisplayName: %q)\n", file.Name, file.DisplayName)
		err := client.DeleteFile(ctx, file.Name)
		if err != nil {
			log.Printf("  ERROR deleting %s: %v\n", file.Name, err)
			deleteErrorCount++
		} else {
			deletedCount++
		}
		time.Sleep(50 * time.Millisecond)
	}

	log.Printf("Nuke summary: Attempted to delete %d files. Successful: %d, Errors: %d\n", totalFilesFound, deletedCount, deleteErrorCount)
}
