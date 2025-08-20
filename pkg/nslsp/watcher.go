// NeuroScript Version: 0.6.0
// File version: 3
// Purpose: Implements a file watcher to enable live reloading of external tool metadata. FIX: Watches the workspace root to detect the creation of the 'tools' directory after startup.
// filename: pkg/nslsp/watcher.go
// nlines: 115
// risk_rating: HIGH

package nslsp

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	lsp "github.com/sourcegraph/go-lsp"
)

// startFileWatcher initializes and runs a file watcher. It watches both the
// configured tool files and the './tools' directory itself to detect changes,
// creations, and deletions, triggering a configuration reload and diagnostic refresh.
func (s *Server) startFileWatcher(ctx context.Context, workspaceRoot lsp.DocumentURI) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		s.logger.Printf("ERROR: Failed to create file watcher: %v", err)
		return
	}
	s.fileWatcher = watcher
	workspacePath := s.resolveWorkspacePath(workspaceRoot, "")
	toolsDir := filepath.Join(workspacePath, "tools")

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// THE FIX IS HERE: Watch for the creation of the 'tools' directory itself.
				if event.Op&fsnotify.Create == fsnotify.Create {
					// Check if the created item is our tools directory.
					if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() && event.Name == toolsDir {
						s.logger.Printf("Detected creation of tools directory: %s. Adding to watcher.", event.Name)
						if err := s.fileWatcher.Add(event.Name); err != nil {
							s.logger.Printf("ERROR: Failed to add newly created tools directory to watcher: %v", err)
						}
						// Continue to the next event, no need to reload config yet.
						continue
					}
				}

				isRelevantEvent := event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Create == fsnotify.Create ||
					event.Op&fsnotify.Remove == fsnotify.Remove

				isJSONFile := strings.HasSuffix(event.Name, ".json")

				if isRelevantEvent && isJSONFile {
					s.logger.Printf("File watcher detected relevant event '%s' for: %s.", event.Op, event.Name)
					s.loadConfig(workspaceRoot)
					s.logger.Println("Re-publishing diagnostics for all open documents after tool configuration reload.")
					openDocs := s.documentManager.GetAll()
					for uri, content := range openDocs {
						go PublishDiagnostics(ctx, s.conn, s.logger, s, uri, content)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				s.logger.Printf("ERROR: File watcher error: %v", err)
			case <-ctx.Done():
				s.logger.Println("File watcher stopping due to context cancellation.")
				return
			}
		}
	}()

	// Add all currently configured metadata files to the watcher.
	for _, relPath := range s.config.ExternalToolMetadata {
		absPath := s.resolveWorkspacePath(workspaceRoot, relPath)
		if err := watcher.Add(absPath); err != nil {
			s.logger.Printf("ERROR: Failed to add file path to watcher: %v", err)
		} else {
			s.logger.Printf("Watching for changes in file: %s", absPath)
		}
	}

	// Watch the tools directory if it exists at startup.
	if _, err := os.Stat(toolsDir); err == nil {
		if err := watcher.Add(toolsDir); err != nil {
			s.logger.Printf("ERROR: Failed to add existing tools directory to watcher: %v", err)
		} else {
			s.logger.Printf("Watching for changes in directory: %s", toolsDir)
		}
	}

	// Always watch the workspace root to detect creation of the tools directory.
	if err := watcher.Add(workspacePath); err != nil {
		s.logger.Printf("ERROR: Failed to add workspace root to watcher: %v", err)
	} else {
		s.logger.Printf("Watching for 'tools' directory creation in: %s", workspacePath)
	}
}

// resolveWorkspacePath is a helper to correctly resolve a path relative to the workspace root.
func (s *Server) resolveWorkspacePath(workspaceRoot lsp.DocumentURI, relPath string) string {
	workspacePath := string(workspaceRoot)
	if strings.HasPrefix(workspacePath, "file://") {
		workspacePath = workspacePath[7:]
	}
	return filepath.Join(workspacePath, relPath)
}
