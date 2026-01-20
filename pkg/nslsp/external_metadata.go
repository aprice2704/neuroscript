// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 4
// :: description: Manages loading external tool AND constant definitions from JSON. Renamed from external_tools.go.
// :: latestChange: Updated LoadFromPaths to support MetadataEnvelope and load constants into memory.
// :: filename: pkg/nslsp/external_metadata.go
// :: serialization: go

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

// MetadataEnvelope mirrors the structure exported by tools_meta_exporter.
type MetadataEnvelope struct {
	Tools     []tool.ToolImplementation `json:"tools"`
	Constants map[string]any            `json:"constants"`
}

// ExternalToolManager holds tool implementations and constants loaded from metadata files.
type ExternalToolManager struct {
	mu        sync.RWMutex
	tools     map[types.FullName]tool.ToolImplementation
	constants map[string]any
	// Keep track of which items came from which file for reloading.
	sourceFileMap map[string]struct {
		Tools     []types.FullName
		Constants []string
	}
}

// NewExternalToolManager creates a new manager for external tools and constants.
func NewExternalToolManager() *ExternalToolManager {
	return &ExternalToolManager{
		tools:     make(map[types.FullName]tool.ToolImplementation),
		constants: make(map[string]any),
		sourceFileMap: make(map[string]struct {
			Tools     []types.FullName
			Constants []string
		}),
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

		// 1. Clear old data associated with this file.
		if oldData, ok := etm.sourceFileMap[absPath]; ok {
			for _, fullName := range oldData.Tools {
				delete(etm.tools, fullName)
			}
			for _, constName := range oldData.Constants {
				delete(etm.constants, constName)
			}
			logger.Printf("Cleared %d tools and %d constants from %s", len(oldData.Tools), len(oldData.Constants), absPath)
		}

		// 2. Load the new file.
		file, err := os.Open(absPath)
		if err != nil {
			logger.Printf("ERROR: Could not open external metadata file %s: %v", absPath, err)
			continue
		}
		defer file.Close()

		var envelope MetadataEnvelope
		// Attempt to decode as Envelope
		if err := json.NewDecoder(file).Decode(&envelope); err != nil {
			// Fallback: If strict envelope decoding fails, try decoding as raw []ToolImplementation
			// This handles the transition period or simple manually edited tool files.
			file.Seek(0, 0) // Rewind
			var toolsOnly []tool.ToolImplementation
			if err2 := json.NewDecoder(file).Decode(&toolsOnly); err2 == nil {
				envelope.Tools = toolsOnly
				logger.Printf("WARN: File %s parsed as raw tool list (legacy format).", absPath)
			} else {
				logger.Printf("ERROR: Failed to parse JSON from %s (tried envelope and legacy list): %v", absPath, err)
				continue
			}
		}

		// 3. Register Tools
		newToolNames := make([]types.FullName, 0, len(envelope.Tools))
		for _, impl := range envelope.Tools {
			baseName := string(impl.Spec.Group) + "." + string(impl.Spec.Name)
			fullName := types.FullName(tool.CanonicalizeToolName(baseName))
			impl.FullName = fullName
			etm.tools[fullName] = impl
			newToolNames = append(newToolNames, fullName)
		}

		// 4. Register Constants
		newConstNames := make([]string, 0, len(envelope.Constants))
		for k, v := range envelope.Constants {
			etm.constants[k] = v
			newConstNames = append(newConstNames, k)
		}

		// 5. Update Source Map
		etm.sourceFileMap[absPath] = struct {
			Tools     []types.FullName
			Constants []string
		}{
			Tools:     newToolNames,
			Constants: newConstNames,
		}

		logger.Printf("Successfully loaded %d tools and %d constants from %s", len(newToolNames), len(newConstNames), absPath)
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

// HasConstant checks if a constant name exists in the loaded metadata.
func (etm *ExternalToolManager) HasConstant(name string) bool {
	etm.mu.RLock()
	defer etm.mu.RUnlock()
	_, found := etm.constants[name]
	return found
}

// GetConstants returns all loaded constants.
func (etm *ExternalToolManager) GetConstants() map[string]any {
	etm.mu.RLock()
	defer etm.mu.RUnlock()
	// Return a copy to be safe
	out := make(map[string]any, len(etm.constants))
	for k, v := range etm.constants {
		out[k] = v
	}
	return out
}
