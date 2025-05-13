// NeuroScript Version: 0.3.0
// File version: 0.1.0
// AI Worker Management: Definition I/O Methods
// filename: pkg/core/ai_wm_definitions_io.go
// nlines: 70 // Approximate
// risk_rating: MEDIUM
package core

import (
	"fmt"
	"os"
	"path/filepath"
)

// persistDefinitionsUnsafe prepares and writes the current definitions to their file.
// Assumes caller holds the write lock. Moved from ai_wm_definitions.go.
func (m *AIWorkerManager) persistDefinitionsUnsafe() error {
	jsonString, err := m.prepareDefinitionsForSaving() // From ai_wm.go
	if err != nil {
		return err
	}

	defPath := m.FullPathForDefinitions() // From ai_wm.go
	if defPath == "" {
		m.logger.Error("Cannot save definitions: file path is not configured in AIWorkerManager.")
		return NewRuntimeError(ErrorCodeConfiguration, "definitions file path not configured for saving", ErrConfiguration)
	}

	dir := filepath.Dir(defPath)
	if mkDirErr := os.MkdirAll(dir, 0755); mkDirErr != nil {
		m.logger.Errorf("Failed to create directory '%s' for definitions file: %v", dir, mkDirErr)
		return NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("failed to create directory for definitions file '%s'", dir), mkDirErr)
	}

	if err := os.WriteFile(defPath, []byte(jsonString), 0644); err != nil {
		m.logger.Errorf("Failed to write definitions to file '%s': %v", defPath, err)
		return NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("failed to write definitions to file '%s'", defPath), err)
	}
	m.logger.Debugf("Successfully saved definitions to %s", defPath)
	return nil
}

// LoadWorkerDefinitionsFromFile is a public method for tools/external calls to reload definitions.
// It replaces all current definitions and re-initializes rate trackers. Moved from ai_wm_definitions.go.
func (m *AIWorkerManager) LoadWorkerDefinitionsFromFile() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	defPath := m.FullPathForDefinitions()
	if defPath == "" {
		m.logger.Error("Cannot load definitions: file path is not configured in AIWorkerManager.")
		m.definitions = make(map[string]*AIWorkerDefinition)
		m.activeInstances = make(map[string]*AIWorkerInstance)
		m.initializeRateTrackersUnsafe()
		return NewRuntimeError(ErrorCodeConfiguration, "definitions file path not configured, cannot load", ErrConfiguration)
	}

	m.logger.Infof("AIWorkerManager: Public request to load worker definitions from %s", defPath)

	m.definitions = make(map[string]*AIWorkerDefinition)
	m.activeInstances = make(map[string]*AIWorkerInstance)

	contentBytes, err := os.ReadFile(defPath)
	if err != nil {
		if os.IsNotExist(err) {
			m.logger.Infof("AIWorkerManager: Definitions file '%s' not found. Manager will have no definitions.", defPath)
			m.initializeRateTrackersUnsafe()
			return nil
		}
		m.logger.Errorf("AIWorkerManager: Error reading definitions file '%s': %v", defPath, err)
		m.initializeRateTrackersUnsafe()
		return NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("failed to read definitions file '%s'", defPath), err)
	}

	if loadErr := m.loadWorkerDefinitionsFromContent(contentBytes); loadErr != nil {
		m.initializeRateTrackersUnsafe()
		return loadErr
	}

	m.initializeRateTrackersUnsafe()

	m.logger.Infof("AIWorkerManager: Public load definitions complete. %d definitions loaded from %s.", len(m.definitions), defPath)
	return nil
}

// SaveWorkerDefinitionsToFile is a public method to persist the current state of definitions.
// Moved from ai_wm_definitions.go.
func (m *AIWorkerManager) SaveWorkerDefinitionsToFile() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	defPath := m.FullPathForDefinitions()
	m.logger.Infof("AIWorkerManager: Public request to save worker definitions to %s", defPath)
	return m.persistDefinitionsUnsafe()
}
