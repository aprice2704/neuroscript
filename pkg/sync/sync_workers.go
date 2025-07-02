// filename: pkg/sync/sync_workers.go
package sync

import (
	"context"
	// "errors"
	// "fmt"

	// "os"
	// "path/filepath"
	"sync"
	"time"

	"github.com/google/generative-ai-go/genai"
	// gitignore "github.com/sabhiram/go-gitignore"
)

// startUploadWorkers initializes and starts the pool of goroutines that handle file uploads/updates.
func startUploadWorkers(sc *syncContext, wg *sync.WaitGroup, actions SyncActions, resultsChan chan<- uploadResult) {
	totalUploadJobs := len(actions.FilesToUpload) + len(actions.FilesToUpdate)
	if totalUploadJobs == 0 {
		sc.logger.Debug("[API HELPER Sync] No upload/update jobs required.")
		return
	}
	const maxConcurrentUploads = 8
	jobsChan := make(chan uploadJob, totalUploadJobs)

	sc.logger.Debug("API HELPER Sync] Starting %d upload workers for %d jobs...", maxConcurrentUploads, totalUploadJobs)
	for i := 0; i < maxConcurrentUploads; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			sc.logger.Debug("HELPER Sync Worker %d] Started.", workerID)
			for job := range jobsChan {
				sc.logger.Debug("HELPER Sync Worker %d] STARTING job for: %s (Update: %t)", workerID, job.relPath, job.existingApiFile != nil)
				apiFile, uploadErr := processUploadJob(sc, job, workerID)	// Pass workerID
				sc.logger.Debug("HELPER Sync Worker %d] processUploadJob finished for: %s (Error: %v)", workerID, job.relPath, uploadErr)
				sc.logger.Debug("HELPER Sync Worker %d] Sending result to resultsChan for: %s", workerID, job.relPath)
				resultsChan <- uploadResult{job: job, apiFile: apiFile, err: uploadErr}
				sc.logger.Debug("HELPER Sync Worker %d] FINISHED job for: %s", workerID, job.relPath)
			}
			sc.logger.Debug("HELPER Sync Worker %d] Exiting (jobsChan closed).", workerID)
		}(i)
	}
	sc.logger.Debug("API HELPER Sync] %d upload workers started.", maxConcurrentUploads)

	sc.logger.Debug("API HELPER Sync] Queuing %d upload jobs...", len(actions.FilesToUpload))
	for _, fileInfo := range actions.FilesToUpload {
		jobsChan <- uploadJob{localAbsPath: fileInfo.AbsPath, relPath: fileInfo.RelPath, localHash: fileInfo.Hash, existingApiFile: nil}
	}
	sc.logger.Debug("API HELPER Sync] Queuing %d update jobs...", len(actions.FilesToUpdate))
	for _, updateJob := range actions.FilesToUpdate {
		jobsChan <- updateJob
	}
	sc.logger.Debug("[DEBUG API HELPER Sync] Finished queuing jobs.")
	close(jobsChan)
}

// processUploadJob handles the logic for a single upload/update job within a worker goroutine.
func processUploadJob(sc *syncContext, job uploadJob, workerID int) (*genai.File, error) {
	sc.logger.Debug("processUploadJob %d] Entered for: %s", workerID, job.relPath)
	defer sc.logger.Debug("processUploadJob %d] Exiting for: %s", workerID, job.relPath)
	var uploadErr error
	var apiFile *genai.File

	if job.existingApiFile != nil {	// Pre-delete logic
		sc.logger.Debug("HELPER Sync Worker %d] Deleting existing %s for update: %s", workerID, job.existingApiFile.Name, job.relPath)
		delCtx, cancelDel := context.WithTimeout(context.Background(), 30*time.Second)
		deleteErr := sc.client.DeleteFile(delCtx, job.existingApiFile.Name)
		cancelDel()
		if deleteErr != nil {
			sc.logger.Error("[ERROR Worker %d] Pre-delete fail %s (%s): %v", workerID, job.existingApiFile.Name, job.relPath, deleteErr)
		} else {
			sc.logger.Debug("Worker %d] Pre-delete OK %s (%s)", workerID, job.existingApiFile.Name, job.relPath)
		}
		time.Sleep(100 * time.Millisecond)
	}

	operation := "Uploading"
	if job.existingApiFile != nil {
		operation = "Updating"
	}
	sc.logger.Debug("[API Worker %d] %s: %s...", workerID, operation, job.relPath)	// Log start of operation

	sc.logger.Debug("processUploadJob %d] Calling HelperUploadAndPollFile: %s", workerID, job.relPath)
	uploadCtx, cancelUpload := context.WithTimeout(context.Background(), 5*time.Minute)
	// Assumes HelperUploadAndPollFile is accessible
	apiFile, uploadErr = HelperUploadAndPollFile(uploadCtx, job.localAbsPath, job.relPath, sc.client, sc.logger)
	cancelUpload()
	sc.logger.Debug("processUploadJob %d] HelperUploadAndPollFile returned: %s (Err: %v)", workerID, job.relPath, uploadErr)
	return apiFile, uploadErr
}

// startDeleteWorkers starts goroutines to process deletions.
func startDeleteWorkers(sc *syncContext, wg *sync.WaitGroup, filesToDelete []*genai.File) {
	if len(filesToDelete) == 0 {
		sc.logger.Debug("[API HELPER Sync] No delete jobs required.")
		return
	}
	const maxConcurrentDeletes = 16
	deleteJobsChan := make(chan *genai.File, len(filesToDelete))
	sc.logger.Debug("API HELPER Sync] Starting %d delete workers for %d jobs...", maxConcurrentDeletes, len(filesToDelete))

	for i := 0; i < maxConcurrentDeletes; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			sc.logger.Debug("HELPER Sync Delete Worker %d] Started.", workerID)
			for fileToDelete := range deleteJobsChan {
				if fileToDelete == nil || fileToDelete.Name == "" {
					continue
				}
				displayName := fileToDelete.DisplayName
				if displayName == "" {
					displayName = fileToDelete.Name
				}
				// Log start of delete operation? (Removed for now to match progress bar only on upload/update)
				// sc.logger.Debug("[API Delete Worker %d] Deleting: %s (%s)...", workerID, displayName, fileToDelete.Name)
				sc.logger.Debug("HELPER Sync Delete Worker %d] Deleting API File: Name=%s, DisplayName=%s", workerID, fileToDelete.Name, displayName)
				delCtx, cancelDel := context.WithTimeout(context.Background(), 30*time.Second)
				deleteErr := sc.client.DeleteFile(delCtx, fileToDelete.Name)
				cancelDel()
				if deleteErr != nil {
					sc.incrementStat("delete_errors")
					sc.logger.Error("[ERROR Delete Worker %d] Fail delete %s (%s): %v", workerID, fileToDelete.Name, displayName, deleteErr)
				} else {
					sc.incrementStat("files_deleted_api")
					sc.logger.Debug("Delete Worker %d] Deleted OK: %s (%s)", workerID, fileToDelete.Name, displayName)
				}
				time.Sleep(50 * time.Millisecond)
			}
			sc.logger.Debug("HELPER Sync Delete Worker %d] Exiting.", workerID)
		}(i)
	}
	sc.logger.Debug("API HELPER Sync] %d delete workers started.", maxConcurrentDeletes)
	sc.logger.Debug("API HELPER Sync] Queuing %d delete jobs...", len(filesToDelete))
	for _, file := range filesToDelete {
		deleteJobsChan <- file
	}
	close(deleteJobsChan)
	sc.logger.Debug("[DEBUG API HELPER Sync] Finished queuing delete jobs.")
}

// Ensure HelperUploadAndPollFile, calculateFileHash, min etc are accessible