// NeuroScript Version: 0.6.0
// File version: 3
// Purpose: Manages loading and retrieval of external tool definitions from JSON, now using the full ToolImplementation struct. FIX: Corrected sync.RWMutex typo.
// filename: pkg/nslsp/external_tools.go
// nlines: 95
// risk_rating: MEDIUM

package nslsp

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// ExternalToolManager holds tool implementations loaded from metadata files.
type ExternalToolManager struct {
	mu    sync.RWMutex
	tools map[types.FullName]tool.ToolImplementation
	// Keep track of which tools came from which file for reloading.
	sourceFileMap map[string][]types.FullName
}

// NewExternalToolManager creates a new manager for external tools.
func NewExternalToolManager() *ExternalToolManager {
	return &ExternalToolManager{
		tools:         make(map[types.FullName]tool.ToolImplementation),
		sourceFileMap: make(map[string][]types.FullName),
	}
}

// LoadFromPaths clears existing definitions for the given paths and reloads them.
func (etm *ExternalToolManager) LoadFromPaths(logger *log.Logger, workspaceRoot string, paths []string) {
	etm.mu.Lock()
	defer etm.mu.Unlock()

	for _, relPath := range paths {
		absPath := filepath.Join(workspaceRoot, relPath)
		if filepath.HasPrefix(workspaceRoot, "file://") {
			absPath = filepath.Join(workspaceRoot[7:], relPath)
		}

		// 1. Clear any old tools from this file path.
		if oldFullNames, ok := etm.sourceFileMap[absPath]; ok {
			for _, fullName := range oldFullNames {
				delete(etm.tools, fullName)
			}
			logger.Printf("Cleared %d old tool implementations from %s", len(oldFullNames), absPath)
		}

		// 2. Load the new implementations.
		file, err := os.Open(absPath)
		if err != nil {
			logger.Printf("ERROR: Could not open external tool metadata file %s: %v", absPath, err)
			continue
		}

		var newImpls []tool.ToolImplementation
		if err := json.NewDecoder(file).Decode(&newImpls); err != nil {
			file.Close()
			logger.Printf("ERROR: Failed to parse JSON from %s: %v", absPath, err)
			continue
		}
		file.Close()

		// 3. Add new implementations to the main map and update the source file map.
		newFullNames := make([]types.FullName, 0, len(newImpls))
		for _, impl := range newImpls {
			baseName := string(impl.Spec.Group) + "." + string(impl.Spec.Name)
			fullName := types.FullName(tool.CanonicalizeToolName(baseName))
			// The loaded impl doesn't have the non-exported FullName, so we set it.
			impl.FullName = fullName
			etm.tools[fullName] = impl
			newFullNames = append(newFullNames, fullName)
		}
		etm.sourceFileMap[absPath] = newFullNames
		logger.Printf("Successfully loaded %d tool implementations from %s", len(newImpls), absPath)
	}
}

// GetTool retrieves a tool implementation by its fully qualified name.
func (etm *ExternalToolManager) GetTool(name types.FullName) (tool.ToolImplementation, bool) {
	etm.mu.RLock()
	defer etm.mu.RUnlock()
	impl, found := etm.tools[name]
	return impl, found
}

// ListTools returns a slice of all loaded external tool implementations.
func (etm *ExternalToolManager) ListTools() []tool.ToolImplementation {
	etm.mu.RLock()
	defer etm.mu.RUnlock()
	list := make([]tool.ToolImplementation, 0, len(etm.tools))
	for _, impl := range etm.tools {
		list = append(list, impl)
	}
	return list
}
