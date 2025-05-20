// NeuroScript Version: 0.3.0
// File version: 0.1.1
// AI Worker Management: Definition I/O Methods. Added LoadRetiredInstancePerformanceDataFromFile.
// filename: pkg/core/ai_wm_definitions_io.go
// nlines: 100 // Approximate
// risk_rating: MEDIUM
package core

import (
	"fmt"
	"os"
	"path/filepath"
)

// persistDefinitionsUnsafe prepares and writes the current definitions to their file.
// Assumes caller holds the write lock.
func (m *AIWorkerManager) persistDefinitionsUnsafe() error {
	jsonString, err := m.prepareDefinitionsForSaving()
	if err != nil {
		return err // Should be a RuntimeError from prepareDefinitionsForSaving
	}

	defPath := m.FullPathForDefinitions()
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
// It replaces all current definitions and re-initializes rate trackers.
func (m *AIWorkerManager) LoadWorkerDefinitionsFromFile() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	defPath := m.FullPathForDefinitions()
	if defPath == "" {
		m.logger.Error("Cannot load definitions: file path is not configured in AIWorkerManager.")
		// Reset state even if path is missing, to ensure consistency
		m.definitions = make(map[string]*AIWorkerDefinition)
		m.activeInstances = make(map[string]*AIWorkerInstance) // Clear active instances as their defs are gone
		m.initializeRateTrackersUnsafe()
		return NewRuntimeError(ErrorCodeConfiguration, "definitions file path not configured, cannot load", ErrConfiguration)
	}

	m.logger.Infof("AIWorkerManager: Public request to load worker definitions from %s", defPath)

	// Clear existing definitions and active instances before loading new ones.
	// This ensures that if loading fails, the manager is in a clean state.
	m.definitions = make(map[string]*AIWorkerDefinition)
	m.activeInstances = make(map[string]*AIWorkerInstance) // Instances depend on definitions

	contentBytes, err := os.ReadFile(defPath)
	if err != nil {
		if os.IsNotExist(err) {
			m.logger.Infof("AIWorkerManager: Definitions file '%s' not found. Manager will have no definitions.", defPath)
			m.initializeRateTrackersUnsafe() // Initialize trackers for an empty set
			return nil                       // Not an error if file doesn't exist, just means no definitions
		}
		m.logger.Errorf("AIWorkerManager: Error reading definitions file '%s': %v", defPath, err)
		m.initializeRateTrackersUnsafe() // Initialize for safety
		return NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("failed to read definitions file '%s'", defPath), err)
	}

	// loadWorkerDefinitionsFromContent will replace m.definitions
	if loadErr := m.loadWorkerDefinitionsFromContent(contentBytes); loadErr != nil {
		m.initializeRateTrackersUnsafe() // Ensure trackers are consistent even after failed load
		return loadErr                   // loadErr should be a RuntimeError
	}

	m.initializeRateTrackersUnsafe() // Re-initialize based on newly loaded definitions

	m.logger.Infof("AIWorkerManager: Public load definitions complete. %d definitions loaded from %s.", len(m.definitions), defPath)
	return nil
}

// SaveWorkerDefinitionsToFile is a public method to persist the current state of definitions.
func (m *AIWorkerManager) SaveWorkerDefinitionsToFile() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	defPath := m.FullPathForDefinitions() // Get path for logging
	m.logger.Infof("AIWorkerManager: Public request to save worker definitions to %s", defPath)
	return m.persistDefinitionsUnsafe()
}

// LoadRetiredInstancePerformanceDataFromFile loads performance data from the configured file.
// This method is additive; it processes the records from the file but doesn't clear
// existing in-memory aggregations directly unless m.loadRetiredInstancePerformanceDataFromContent does so.
// The primary purpose of loading this file is usually at startup to inform definition summaries.
func (m *AIWorkerManager) LoadRetiredInstancePerformanceDataFromFile() error {
	m.mu.Lock() // Lock for thread-safe access to manager state if needed by underlying methods
	defer m.mu.Unlock()

	perfDataPath := m.FullPathForPerformanceData()
	if perfDataPath == "" {
		m.logger.Warn("Cannot load performance data: file path is not configured in AIWorkerManager.")
		// This is not necessarily a fatal error, could just mean no historical data to load.
		return nil
	}

	m.logger.Infof("AIWorkerManager: Attempting to load performance data from %s", perfDataPath)

	contentBytes, err := os.ReadFile(perfDataPath)
	if err != nil {
		if os.IsNotExist(err) {
			m.logger.Infof("AIWorkerManager: Performance data file '%s' not found. No historical performance data loaded.", perfDataPath)
			return nil // Not an error if the file simply doesn't exist
		}
		m.logger.Errorf("AIWorkerManager: Error reading performance data file '%s': %v", perfDataPath, err)
		return NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("failed to read performance data file '%s'", perfDataPath), err)
	}

	if loadErr := m.loadRetiredInstancePerformanceDataFromContent(contentBytes); loadErr != nil {
		// loadRetiredInstancePerformanceDataFromContent should return a RuntimeError
		m.logger.Errorf("AIWorkerManager: Failed to process loaded performance data from '%s': %v", perfDataPath, loadErr)
		return loadErr
	}

	m.logger.Infof("AIWorkerManager: Successfully loaded and processed performance data from %s.", perfDataPath)
	return nil
}

// appendRetiredInstanceToFileUnsafe appends a single RetiredInstanceInfo to the performance data file.
// Assumes caller holds appropriate locks if concurrent file access is a concern for other operations.
// This is an internal helper typically called by RetireWorkerInstance.
func (m *AIWorkerManager) appendRetiredInstanceToFileUnsafe(info RetiredInstanceInfo) error {
	filePath := m.FullPathForPerformanceData()
	if filePath == "" {
		m.logger.Error("Cannot append retired instance: performance data file path not configured.")
		return NewRuntimeError(ErrorCodeConfiguration, "performance data file path not configured", ErrConfiguration)
	}

	var existingContentBytes []byte
	var readErr error
	if _, statErr := os.Stat(filePath); statErr == nil { // File exists
		existingContentBytes, readErr = os.ReadFile(filePath)
		if readErr != nil {
			m.logger.Errorf("Failed to read existing performance data file '%s' for appending: %v. Will attempt to write as new file.", filePath, readErr)
			existingContentBytes = []byte{} // Treat as empty if read fails
		}
	} else if !os.IsNotExist(statErr) { // Some other error stating the file
		m.logger.Errorf("Error checking performance data file '%s' for appending: %v. Will attempt to write as new file.", filePath, statErr)
		existingContentBytes = []byte{}
	}

	// prepareRetiredInstanceForAppending expects a string, so convert bytes if they exist.
	// If existingContentBytes is empty or only whitespace, pass empty string.
	existingContentStr := ""
	if len(existingContentBytes) > 0 {
		existingContentStr = string(existingContentBytes)
	}

	updatedJSONString, err := m.prepareRetiredInstanceForAppending(existingContentStr, &info)
	if err != nil {
		// err from prepareRetiredInstanceForAppending should be a RuntimeError
		m.logger.Errorf("Failed to prepare performance data for appending instance %s: %v", info.InstanceID, err)
		return err
	}

	dir := filepath.Dir(filePath)
	if mkDirErr := os.MkdirAll(dir, 0755); mkDirErr != nil {
		m.logger.Errorf("Failed to create directory '%s' for performance data file: %v", dir, mkDirErr)
		return NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("failed to create directory for performance data file '%s'", dir), mkDirErr)
	}

	if err := os.WriteFile(filePath, []byte(updatedJSONString), 0644); err != nil {
		m.logger.Errorf("Failed to write updated performance data to file '%s': %v", filePath, err)
		return NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("failed to write performance data to file '%s'", filePath), err)
	}

	m.logger.Debugf("Successfully appended retired instance %s to %s", info.InstanceID, filePath)
	return nil
}
