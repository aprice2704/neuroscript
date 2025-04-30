// filename: pkg/core/sync_types.go
package core

import (
	"context"

	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/google/generative-ai-go/genai"
	gitignore "github.com/sabhiram/go-gitignore"
)

// syncContext holds shared state and configuration for the sync operation.
// *** MODIFIED: Added interp field ***
type syncContext struct {
	ctx           context.Context
	absLocalDir   string
	filterPattern string
	ignorer       *gitignore.GitIgnore // Gitignore rules
	client        *genai.Client        // GenAI client
	logger        logging.Logger
	stats         map[string]interface{} // Statistics map
	incrementStat func(string)           // Function to increment stats
	interp        *Interpreter           // <<< ADDED: Interpreter reference needed by helpers
}

// LocalFileInfo stores details about a local file found during the walk.
type LocalFileInfo struct {
	RelPath string // Path relative to absLocalDir
	AbsPath string // Absolute path on disk
	Hash    string // SHA256 hash of file content
}

// uploadJob defines the data needed for an upload/update worker.
type uploadJob struct {
	localAbsPath    string
	relPath         string
	localHash       string
	existingApiFile *genai.File // non-nil if updating (delete first)
}

// uploadResult defines the result of an upload worker's job.
type uploadResult struct {
	job     uploadJob
	apiFile *genai.File // nil on error
	err     error
}

// SyncActions holds the lists of operations determined by comparing local and remote state.
type SyncActions struct {
	FilesToUpload []LocalFileInfo // Files only present locally
	FilesToUpdate []uploadJob     // Files present in both, but hashes differ
	FilesToDelete []*genai.File   // Files only present remotely
}
