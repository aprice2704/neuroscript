// NeuroScript Version: 0.6.0
// File version: 3
// Purpose: Manages LSP server configuration. FIX: Defaults to scanning './tools/*.json' if no .nslsp.json is found.
// filename: pkg/nslsp/config.go
// nlines: 80
// risk_rating: MEDIUM

package nslsp

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	lsp "github.com/sourcegraph/go-lsp"
)

// Config holds the server's configuration settings.
type Config struct {
	ExternalToolMetadata []string `json:"nslsp.externalToolMetadata"`
}

// loadConfig searches for and loads a .nslsp.json configuration file.
// If not found, it defaults to scanning for and loading all *.json files
// in the ./tools/ directory.
func (s *Server) loadConfig(workspaceRoot lsp.DocumentURI) {
	if workspaceRoot == "" {
		s.logger.Println("No workspace root provided, cannot load configuration.")
		return
	}

	workspacePath := string(workspaceRoot)
	if strings.HasPrefix(workspacePath, "file://") {
		workspacePath = workspacePath[7:]
	}
	configPath := filepath.Join(workspacePath, ".nslsp.json")

	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// THE FIX IS HERE: Scan the ./tools directory for all .json files.
			s.logger.Printf("No configuration file at %s, defaulting to scan './tools/*.json'.", configPath)
			toolsDir := filepath.Join(workspacePath, "tools")
			files, err := os.ReadDir(toolsDir)
			if err != nil {
				s.logger.Printf("Could not read default tools directory '%s': %v", toolsDir, err)
				return // Nothing to load
			}
			var defaultPaths []string
			for _, f := range files {
				if !f.IsDir() && strings.HasSuffix(f.Name(), ".json") {
					defaultPaths = append(defaultPaths, filepath.Join("tools", f.Name()))
				}
			}
			s.config.ExternalToolMetadata = defaultPaths
			s.logger.Printf("Found %d potential tool metadata files in './tools/'.", len(defaultPaths))
		} else {
			s.logger.Printf("Error opening configuration file %s: %v", configPath, err)
			return
		}
	} else {
		// If the file was opened successfully, parse it.
		defer file.Close()
		s.logger.Printf("Loading configuration from %s", configPath)

		bytes, err := io.ReadAll(file)
		if err != nil {
			s.logger.Printf("Error reading configuration file %s: %v", configPath, err)
			return
		}

		var config Config
		if err := json.Unmarshal(bytes, &config); err != nil {
			s.logger.Printf("Error parsing configuration file %s: %v", configPath, err)
			return
		}
		s.config = config
		s.logger.Printf("Configuration loaded: %+v", s.config)
	}

	// After loading a config or setting the default, process the external tool files.
	if len(s.config.ExternalToolMetadata) > 0 {
		s.externalTools.LoadFromPaths(s.logger, string(workspaceRoot), s.config.ExternalToolMetadata)
	}
}
