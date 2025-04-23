// filename: pkg/neurogo/agent_context.go
package neurogo

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/generative-ai-go/genai" // For File API types if needed
)

// HandlePrefixAgentContext defines the prefix for AgentContext handles.
const HandlePrefixAgentContext = "agentctx"

// AgentContext holds the configuration and dynamic state specific to the
// neurogo agent mode. It is managed by the App and accessed by tools
// via a handle.
type AgentContext struct {
	mu sync.RWMutex // Protects access to all fields

	// Configuration (set via startup script tools)
	sandboxDir    string
	allowlistPath string
	modelName     string // Added from meltdown.md proposal

	// File Context State
	// Map key is the relative path within the sandboxDir/project
	// Map value is the File API URI (e.g., "files/...")
	syncedFileURIs    map[string]string // Populated by TOOL.SyncDirectory or similar
	pinnedFileURIs    map[string]string // Populated by TOOL.AgentPinFile, always included
	tempRequestedURIs map[string]string // Populated by TOOL.RequestFileContext, cleared after use

	log *log.Logger // Logger for internal operations
}

// NewAgentContext creates a new, initialized AgentContext.
func NewAgentContext(logger *log.Logger) *AgentContext {
	if logger == nil {
		// Fallback if nil logger is passed, though ideally App should provide one
		logger = log.New(log.Writer(), "AgentContext: ", log.LstdFlags|log.Lshortfile)
	}
	return &AgentContext{
		syncedFileURIs:    make(map[string]string),
		pinnedFileURIs:    make(map[string]string),
		tempRequestedURIs: make(map[string]string),
		log:               logger,
		// Initialize defaults for config? Or rely solely on startup script?
		// For now, leave them zero/empty. Startup script MUST set them.
	}
}

// --- Configuration Methods ---

// SetSandboxDir sets the agent's sandbox directory.
// Expected to be called by a startup tool (e.g., TOOL.AgentSetSandbox).
func (ac *AgentContext) SetSandboxDir(path string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.sandboxDir = path
	ac.log.Printf("Agent sandbox directory set to: %s", path)
}

// GetSandboxDir returns the configured sandbox directory.
func (ac *AgentContext) GetSandboxDir() string {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.sandboxDir
}

// SetAllowlistPath sets the path to the tool allowlist file.
// Expected to be called by a startup tool (e.g., TOOL.AgentSetAllowlist).
func (ac *AgentContext) SetAllowlistPath(path string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.allowlistPath = path
	ac.log.Printf("Agent allowlist path set to: %s", path)
}

// GetAllowlistPath returns the configured allowlist path.
func (ac *AgentContext) GetAllowlistPath() string {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.allowlistPath
}

// SetModelName sets the AI model name to be used.
// Expected to be called by a startup tool (e.g., TOOL.AgentSetModel).
func (ac *AgentContext) SetModelName(name string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.modelName = name
	ac.log.Printf("Agent model name set to: %s", name)
}

// GetModelName returns the configured model name.
func (ac *AgentContext) GetModelName() string {
	ac.mu.RLock()
	defer ac.mu.RUnlock()
	return ac.modelName
}

// --- File Context Management Methods ---

// UpdateSyncedURIs replaces the current map of synced file URIs.
// This is typically called after a TOOL.SyncDirectory operation.
// Input map key should be the relative path.
func (ac *AgentContext) UpdateSyncedURIs(syncedURIs map[string]string) {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.syncedFileURIs = make(map[string]string, len(syncedURIs)) // Create a new map
	for relPath, uri := range syncedURIs {
		ac.syncedFileURIs[relPath] = uri
	}
	ac.log.Printf("Updated synced file URIs map (count: %d)", len(ac.syncedFileURIs))
}

// PinFile adds a file (by its relative path and URI) to the pinned context.
// Expected to be called by TOOL.AgentPinFile.
func (ac *AgentContext) PinFile(relativePath string, uri string) error {
	if relativePath == "" || uri == "" {
		return fmt.Errorf("cannot pin file with empty relative path or URI")
	}
	ac.mu.Lock()
	defer ac.mu.Unlock()
	if _, exists := ac.pinnedFileURIs[relativePath]; exists {
		ac.log.Printf("Note: Re-pinning file '%s' (URI: %s)", relativePath, uri)
	} else {
		ac.log.Printf("Pinning file '%s' (URI: %s)", relativePath, uri)
	}
	ac.pinnedFileURIs[relativePath] = uri
	return nil
}

// UnpinFile removes a file (by its relative path) from the pinned context.
// Expected to be called by TOOL.Forget.
func (ac *AgentContext) UnpinFile(relativePath string) bool {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	_, exists := ac.pinnedFileURIs[relativePath]
	if exists {
		delete(ac.pinnedFileURIs, relativePath)
		ac.log.Printf("Unpinned file '%s'", relativePath)
	} else {
		ac.log.Printf("Attempted to unpin non-existent file '%s'", relativePath)
	}
	return exists
}

// UnpinAllFiles removes all files from the pinned context.
// Expected to be called by TOOL.ForgetAll.
func (ac *AgentContext) UnpinAllFiles() int {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	count := len(ac.pinnedFileURIs)
	if count > 0 {
		ac.pinnedFileURIs = make(map[string]string) // Reset the map
		ac.log.Printf("Unpinned all %d files", count)
	} else {
		ac.log.Println("Attempted to unpin all files, but none were pinned.")
	}
	return count
}

// AddTemporaryURI adds a file URI requested temporarily for the next turn.
// Called by agent logic after successful TOOL.RequestFileContext.
func (ac *AgentContext) AddTemporaryURI(relativePath string, uri string) error {
	if relativePath == "" || uri == "" {
		return fmt.Errorf("cannot add temporary file with empty relative path or URI")
	}
	ac.mu.Lock()
	defer ac.mu.Unlock()
	if _, exists := ac.tempRequestedURIs[relativePath]; exists {
		// Already requested, maybe log? Or just overwrite? Overwriting seems fine.
		ac.log.Printf("Note: Re-requesting temporary file '%s' (URI: %s)", relativePath, uri)
	} else {
		ac.log.Printf("Adding temporary file '%s' (URI: %s)", relativePath, uri)
	}
	ac.tempRequestedURIs[relativePath] = uri
	return nil
}

// GetURIsForNextContext compiles the list of URIs (pinned + temporary)
// for the next LLM API call and clears the temporary list.
// This is called by the agent's turn handler (handle_turn.go).
func (ac *AgentContext) GetURIsForNextContext() []*genai.File {
	ac.mu.Lock() // Full lock needed to read pinned/temp and clear temp atomically
	defer ac.mu.Unlock()

	combinedURIs := make(map[string]struct{}) // Use map for deduplication
	uriList := make([]*genai.File, 0, len(ac.pinnedFileURIs)+len(ac.tempRequestedURIs))

	// Add pinned files
	for relPath, uri := range ac.pinnedFileURIs {
		if _, exists := combinedURIs[uri]; !exists {
			combinedURIs[uri] = struct{}{}
			// Extract just the filename part for genai.File
			fileName := filepath.Base(relPath) // Or should we keep the full relative path? Check API needs.
			uriList = append(uriList, &genai.File{
				Name:        uri,      // API expects the "files/..." URI here
				DisplayName: fileName, // A user-friendly name
			})
			// ac.log.Printf("Adding pinned URI to context: %s (Display: %s)", uri, fileName) // Too verbose?
		} else {
			ac.log.Printf("Skipping duplicate pinned URI: %s (Path: %s)", uri, relPath)
		}
	}

	// Add temporary files
	for relPath, uri := range ac.tempRequestedURIs {
		if _, exists := combinedURIs[uri]; !exists {
			combinedURIs[uri] = struct{}{}
			fileName := filepath.Base(relPath)
			uriList = append(uriList, &genai.File{
				Name:        uri,
				DisplayName: fileName,
			})
			// ac.log.Printf("Adding temporary URI to context: %s (Display: %s)", uri, fileName) // Too verbose?
		} else {
			ac.log.Printf("Skipping duplicate temporary URI: %s (Path: %s)", uri, relPath)
		}
	}

	// Clear the temporary list for the next cycle
	clearedCount := len(ac.tempRequestedURIs)
	if clearedCount > 0 {
		ac.tempRequestedURIs = make(map[string]string)
		// ac.log.Printf("Cleared %d temporary file URIs.", clearedCount) // Maybe too verbose?
	}

	ac.log.Printf("Compiled %d unique URIs for next context.", len(uriList))
	return uriList
}

// LookupURI finds the File API URI for a given relative path.
// It checks pinned files first, then synced files.
// Returns the URI and true if found, otherwise empty string and false.
func (ac *AgentContext) LookupURI(relativePath string) (string, bool) {
	ac.mu.RLock()
	defer ac.mu.RUnlock()

	// Check pinned first
	uri, found := ac.pinnedFileURIs[relativePath]
	if found {
		return uri, true
	}

	// Check synced next
	uri, found = ac.syncedFileURIs[relativePath]
	if found {
		return uri, true
	}

	// Fallback: Check if the path exists in temporary (might be useful)
	uri, found = ac.tempRequestedURIs[relativePath]
	if found {
		ac.log.Printf("Warning: Looked up URI for '%s' found in temporary list, this might be unexpected.", relativePath)
		return uri, true
	}

	return "", false
}

// Helper to safely convert file path to relative path based on sandbox
func (ac *AgentContext) getRelativePath(filePath string) (string, error) {
	ac.mu.RLock()
	sandbox := ac.sandboxDir
	ac.mu.RUnlock()

	if sandbox == "" {
		return "", fmt.Errorf("sandbox directory not configured in AgentContext")
	}

	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for '%s': %w", filePath, err)
	}

	absSandboxPath, err := filepath.Abs(sandbox)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for sandbox '%s': %w", sandbox, err)
	}

	if !strings.HasPrefix(absFilePath, absSandboxPath) {
		return "", fmt.Errorf("path '%s' is outside the configured sandbox '%s'", filePath, sandbox)
	}

	relPath, err := filepath.Rel(absSandboxPath, absFilePath)
	if err != nil {
		// This should ideally not happen if the prefix check passed
		return "", fmt.Errorf("failed to get relative path for '%s' within sandbox '%s': %w", filePath, sandbox, err)
	}
	return filepath.ToSlash(relPath), nil // Use forward slashes for consistency
}
