// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Implements a file watcher to enable live reloading of external tool metadata.
// filename: pkg/nslsp/watcher.go
// nlines: 70
// risk_rating: MEDIUM

package nslsp

import (
	"context"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	lsp "github.com/sourcegraph/go-lsp"
)

// startFileWatcher initializes and runs a file watcher for the paths specified
// in the server's configuration. It reloads tool definitions and re-publishes
// diagnostics when a file is changed.
func (s *Server) startFileWatcher(ctx context.Context, workspaceRoot lsp.DocumentURI) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		s.logger.Printf("ERROR: Failed to create file watcher: %v", err)
		return
	}
	s.fileWatcher = watcher

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// We only care about writes to the file.
				if event.Op&fsnotify.Write == fsnotify.Write {
					s.logger.Printf("File watcher detected change in: %s. Reloading external tools.", event.Name)

					// The paths in the config are relative, so we just reload all of them.
					s.externalTools.LoadFromPaths(s.logger, string(workspaceRoot), s.config.ExternalToolMetadata)

					// Re-publish diagnostics for all open documents.
					s.logger.Println("Re-publishing diagnostics for all open documents after tool reload.")
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
				return
			}
		}
	}()

	// Add all configured metadata files to the watcher.
	for _, relPath := range s.config.ExternalToolMetadata {
		absPath := filepath.Join(string(workspaceRoot), relPath)
		if filepath.HasPrefix(string(workspaceRoot), "file://") {
			absPath = filepath.Join(string(workspaceRoot[7:]), relPath)
		}
		if err := watcher.Add(absPath); err != nil {
			s.logger.Printf("ERROR: Failed to add path to file watcher: %v", err)
		} else {
			s.logger.Printf("Watching for changes in: %s", absPath)
		}
	}
}
