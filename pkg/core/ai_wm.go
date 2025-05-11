// NeuroScript Version: 0.3.1
// File version: 0.2.2 // Changed INFO logs to DEBUG
// filename: pkg/core/ai_wm.go

package core

import (
	"encoding/json"
	"fmt"
	"os" // Added for os.Getenv
	"path/filepath"
	"sync"
	"time"

	"github.com/aprice2704/neuroscript/pkg/logging"
	// "github.com/google/uuid" // Keep if other parts of the file use it.
)

// Constants for default filenames remain
const (
	defaultDefinitionsFile     = "ai_worker_definitions.json"
	defaultPerformanceDataFile = "ai_worker_performance_data.json"
	statelessInstanceIDPrefix  = "stateless-"
)

// AIWorkerManager manages AIWorkerDefinitions and AIWorkerInstances.
type AIWorkerManager struct {
	definitions     map[string]*AIWorkerDefinition
	activeInstances map[string]*AIWorkerInstance
	rateTrackers    map[string]*WorkerRateTracker
	// File paths are stored for context, but direct I/O is removed from core manager logic
	definitionsBaseFilename     string // e.g., "ai_worker_definitions.json"
	performanceDataBaseFilename string // e.g., "ai_worker_performance_data.json"
	sandboxDir                  string // Base directory for these files

	mu        sync.RWMutex
	logger    logging.Logger
	llmClient LLMClient
}

// NewAIWorkerManager creates and initializes a new AIWorkerManager.
// It no longer loads from disk directly but can accept initial content.
func NewAIWorkerManager(
	logger logging.Logger,
	sandboxDir string, // Sandbox where files are expected to be by tools
	llmClient LLMClient,
	initialDefinitionsContent string, // Optional: initial JSON content for definitions
	initialPerformanceContent string, // Optional: initial JSON content for performance data
) (*AIWorkerManager, error) {

	if logger == nil {
		// This should ideally not happen if caller ensures logger. Fallback for safety.
		// Using a basic fmt.Printf logger here as coreNoOpLogger might not be initialized yet
		// if this is part of core initialization itself. A true panic might be better.
		// For now, let's assume logger is always provided.
		return nil, fmt.Errorf("logger cannot be nil for AIWorkerManager")
	}
	// No direct file access here, so sandboxDir existence check is less critical for constructor
	// but still important for context.
	if sandboxDir == "" {
		logger.Warn("AIWorkerManager: sandboxDir is empty during initialization. File operations by tools might be ambiguous or use current working directory.")
	}

	m := &AIWorkerManager{
		definitions:                 make(map[string]*AIWorkerDefinition),
		activeInstances:             make(map[string]*AIWorkerInstance),
		rateTrackers:                make(map[string]*WorkerRateTracker),
		definitionsBaseFilename:     defaultDefinitionsFile,
		performanceDataBaseFilename: defaultPerformanceDataFile,
		sandboxDir:                  sandboxDir,
		logger:                      logger,
		llmClient:                   llmClient,
	}

	if initialDefinitionsContent != "" {
		logger.Debugf("AIWorkerManager attempting to load definitions from provided initial content (length: %d).", len(initialDefinitionsContent))
		// Pass as []byte as json.Unmarshal expects that.
		if err := m.loadWorkerDefinitionsFromContent([]byte(initialDefinitionsContent)); err != nil {
			logger.Errorf("AIWorkerManager: Failed to load definitions from initial content: %v. Proceeding with empty definitions.", err)
		}
	} else {
		logger.Debugf("AIWorkerManager: No initial definitions content provided. Starting with an empty set of definitions.") // Changed from Infof
	}

	if initialPerformanceContent != "" {
		logger.Debugf("AIWorkerManager attempting to load performance data from provided initial content (length: %d).", len(initialPerformanceContent))
		if err := m.loadRetiredInstancePerformanceDataFromContent([]byte(initialPerformanceContent)); err != nil {
			logger.Errorf("AIWorkerManager: Failed to load performance data from initial content: %v.", err)
		}
	} else {
		logger.Debugf("AIWorkerManager: No initial performance data content provided.") // Changed from Infof
	}

	// This needs to be called after m.definitions might have been populated by loadWorkerDefinitionsFromContent
	m.initializeRateTrackersUnsafe()
	logger.Debugf("AIWorkerManager initialized. Loaded %d definitions from content. Active instances: %d. Sandbox context: '%s'", len(m.definitions), len(m.activeInstances), m.sandboxDir) // Changed from Infof
	return m, nil
}

// FullPathForDefinitions provides the expected full path for the definitions file.
func (m *AIWorkerManager) FullPathForDefinitions() string {
	if m.sandboxDir == "" || m.definitionsBaseFilename == "" {
		m.logger.Warnf("Cannot determine full path for definitions: sandboxDir ('%s') or baseFilename ('%s') is empty.", m.sandboxDir, m.definitionsBaseFilename)
		return ""
	}
	return filepath.Join(m.sandboxDir, m.definitionsBaseFilename)
}

// FullPathForPerformanceData provides the expected full path for the performance data file.
func (m *AIWorkerManager) FullPathForPerformanceData() string {
	if m.sandboxDir == "" || m.performanceDataBaseFilename == "" {
		m.logger.Warnf("Cannot determine full path for performance data: sandboxDir ('%s') or baseFilename ('%s') is empty.", m.sandboxDir, m.performanceDataBaseFilename)
		return ""
	}
	return filepath.Join(m.sandboxDir, m.performanceDataBaseFilename)
}

// loadWorkerDefinitionsFromContent loads AI worker definitions from JSON byte content.
func (m *AIWorkerManager) loadWorkerDefinitionsFromContent(jsonBytes []byte) error {
	// Caller is responsible for locking if this is called outside of NewAIWorkerManager
	// and concurrent access to m.definitions is possible.
	// m.mu.Lock()
	// defer m.mu.Unlock()

	if len(jsonBytes) == 0 {
		m.logger.Debugf("loadWorkerDefinitionsFromContent: Provided content is empty. No definitions loaded.") // Changed from Infof
		m.definitions = make(map[string]*AIWorkerDefinition)                                                   // Reset to empty
		return nil
	}

	var defs []*AIWorkerDefinition
	if err := json.Unmarshal(jsonBytes, &defs); err != nil {
		m.logger.Errorf("loadWorkerDefinitionsFromContent: Failed to unmarshal definitions JSON: %v", err)
		// Potentially corrupted content, clear existing definitions to avoid partial state.
		m.definitions = make(map[string]*AIWorkerDefinition)
		return NewRuntimeError(ErrorCodeInternal, "failed to unmarshal definitions data from content", err)
	}

	newDefinitions := make(map[string]*AIWorkerDefinition)
	for _, def := range defs {
		if def == nil {
			m.logger.Warnf("loadWorkerDefinitionsFromContent: Encountered a nil definition in content. Skipping.")
			continue
		}
		if def.DefinitionID == "" {
			// This could be an issue with data integrity or an older format.
			// For now, we skip. A more robust solution might assign a UUID if name is present.
			m.logger.Warnf("loadWorkerDefinitionsFromContent: Encountered definition with empty ID (Name: '%s'). Skipping.", def.Name)
			continue
		}
		// Basic validation or default setting for status if missing
		if def.Status == "" {
			def.Status = DefinitionStatusActive // Default to active if not specified
			m.logger.Debugf("Definition '%s' (%s) had empty status, defaulted to '%s'.", def.Name, def.DefinitionID, def.Status)
		}
		newDefinitions[def.DefinitionID] = def
	}

	m.definitions = newDefinitions                                                                          // Replace existing definitions
	m.logger.Debugf("Successfully loaded/reloaded %d worker definitions from content.", len(m.definitions)) // Changed from Infof

	// Important: Rate trackers need to be re-initialized based on the newly loaded definitions.
	// This is typically done after this call by the caller (e.g. in NewAIWorkerManager or a reload tool).
	// If this method can be called stand-alone later for a "hot reload", it should also trigger tracker re-init.
	// For now, initializeRateTrackersUnsafe is called in NewAIWorkerManager after this.
	// If called independently:
	// m.initializeRateTrackersUnsafe() // Call under the same lock as m.definitions modification.

	return nil
}

// prepareDefinitionsForSaving marshals the current AI worker definitions to a JSON string.
// Caller must hold appropriate lock (usually Write Lock via m.mu.Lock()).
func (m *AIWorkerManager) prepareDefinitionsForSaving() (string, error) {
	// Assumes caller holds the write lock (m.mu.Lock())
	defsToSave := make([]*AIWorkerDefinition, 0, len(m.definitions))
	for _, def := range m.definitions {
		if def == nil {
			continue
		}

		// Ensure AggregatePerformanceSummary is not nil before accessing fields
		if def.AggregatePerformanceSummary == nil { // Should be initialized by definition creation logic
			def.AggregatePerformanceSummary = &AIWorkerPerformanceSummary{}
			m.logger.Warnf("prepareDefinitionsForSaving: DefinitionID %s had nil AggregatePerformanceSummary; initialized.", def.DefinitionID)
		}

		if tracker, ok := m.rateTrackers[def.DefinitionID]; ok {
			def.AggregatePerformanceSummary.ActiveInstancesCount = tracker.CurrentActiveInstances
		} else {
			// If no tracker, it means no active instances are being tracked for it, so 0 is correct.
			def.AggregatePerformanceSummary.ActiveInstancesCount = 0
			// This warning might be noisy if definitions are often added without immediate spawning.
			// m.logger.Debugf("prepareDefinitionsForSaving: No rate tracker found for DefID %s. ActiveInstancesCount in summary is 0.", def.DefinitionID)
		}
		defsToSave = append(defsToSave, def)
	}

	data, err := json.MarshalIndent(defsToSave, "", "  ")
	if err != nil {
		m.logger.Errorf("prepareDefinitionsForSaving: Failed to marshal worker definitions: %v", err)
		return "", NewRuntimeError(ErrorCodeInternal, "failed to marshal worker definitions for saving", err)
	}
	m.logger.Debugf("Successfully prepared %d worker definitions for saving (as JSON string content).", len(defsToSave))
	return string(data), nil
}

// resolveAPIKey resolves the API key based on the provided APIKeySource.
func (m *AIWorkerManager) resolveAPIKey(auth APIKeySource) (string, error) {
	m.logger.Debugf("Resolving API key with method: %s", auth.Method) // Value not logged for security
	switch auth.Method {
	case APIKeyMethodEnvVar:
		if auth.Value == "" {
			err := NewRuntimeError(ErrorCodeArgMismatch,
				"API key method is 'env_var' but no environment variable name (Value) was specified",
				ErrInvalidArgument,
			)
			m.logger.Warnf("resolveAPIKey: %s", err.Message)
			return "", err
		}
		key := os.Getenv(auth.Value) // Use os.Getenv directly
		if key == "" {
			err := NewRuntimeError(ErrorCodeConfiguration,
				fmt.Sprintf("environment variable '%s' for API key not found or is empty", auth.Value),
				ErrConfiguration, // Or a more specific ErrAPIKeyNotFound
			)
			m.logger.Warnf("resolveAPIKey: Environment variable '%s' for API key not found or is empty.", auth.Value)
			return "", err
		}
		m.logger.Debugf("Successfully resolved API key from environment variable '%s'", auth.Value)
		return key, nil
	case APIKeyMethodInline:
		if auth.Value == "" {
			m.logger.Debugf("API key method is '%s' but the key value is empty. This might be acceptable for some models.", APIKeyMethodInline)
		} else {
			m.logger.Debugf("Using inline API key (actual key value is not logged for security).")
		}
		return auth.Value, nil
	case APIKeyMethodNone:
		m.logger.Debugf("API key method is '%s', no key required.", APIKeyMethodNone)
		return "", nil
	case APIKeyMethodConfigPath, APIKeyMethodVault:
		errMessage := fmt.Sprintf("API key method '%s' is not yet implemented", auth.Method)
		// Consider having a shared ErrFeatureNotImplemented in errors.go
		err := NewRuntimeError(ErrorCodeNotImplemented, errMessage, fmt.Errorf("feature not implemented: %s", auth.Method))
		m.logger.Errorf("resolveAPIKey: %s", errMessage)
		return "", err
	default:
		errMessage := fmt.Sprintf("unknown API key source method: '%s'", auth.Method)
		err := NewRuntimeError(ErrorCodeArgMismatch, errMessage, ErrInvalidArgument)
		m.logger.Errorf("resolveAPIKey: %s", errMessage)
		return "", err
	}
}

// initializeRateTrackersUnsafe ensures rate trackers are set up for all loaded definitions.
// Caller must hold lock if concurrent access to m.rateTrackers or m.definitions is possible.
func (m *AIWorkerManager) initializeRateTrackersUnsafe() {
	newRateTrackers := make(map[string]*WorkerRateTracker) // Create a new map to avoid modifying in place if iterating
	for defID, def := range m.definitions {
		if def == nil {
			m.logger.Warnf("initializeRateTrackersUnsafe: Encountered nil definition for ID '%s'. Skipping tracker initialization.", defID)
			continue
		}

		activeCount := 0
		// Ensure AggregatePerformanceSummary is not nil
		if def.AggregatePerformanceSummary != nil && def.AggregatePerformanceSummary.ActiveInstancesCount > 0 {
			activeCount = def.AggregatePerformanceSummary.ActiveInstancesCount
		}

		newRateTrackers[defID] = &WorkerRateTracker{
			DefinitionID:           defID,
			RequestsMinuteMarker:   time.Now(), // Initialize markers to current time
			TokensMinuteMarker:     time.Now(),
			TokensDayMarker:        time.Now(),
			CurrentActiveInstances: activeCount, // Initialize from summary, or 0 if summary is nil/empty
			// Request counts, token counts default to 0
		}
		m.logger.Debugf("Initialized rate tracker for DefinitionID: %s, ActiveInstances from summary: %d", defID, activeCount)
	}
	m.rateTrackers = newRateTrackers // Atomically replace the old map
	m.logger.Debugf("Re-initialized all rate trackers. Total count: %d", len(m.rateTrackers))
}

// loadRetiredInstancePerformanceDataFromContent loads and processes performance data.
// Placeholder: Actual processing of records to update summaries is complex and needs care.
func (m *AIWorkerManager) loadRetiredInstancePerformanceDataFromContent(jsonBytes []byte) error {
	// m.mu.Lock()
	// defer m.mu.Unlock()
	m.logger.Debug("loadRetiredInstancePerformanceDataFromContent called.")
	if len(jsonBytes) == 0 {
		m.logger.Debugf("loadRetiredInstancePerformanceDataFromContent: Provided content is empty. No historical performance loaded or processed.") // Changed from Infof
		return nil
	}

	var retiredInfos []*RetiredInstanceInfo
	if err := json.Unmarshal(jsonBytes, &retiredInfos); err != nil {
		m.logger.Errorf("loadRetiredInstancePerformanceDataFromContent: Failed to unmarshal performance data JSON: %v", err)
		return NewRuntimeError(ErrorCodeInternal, "failed to unmarshal performance data from content", err)
	}

	m.logger.Debugf("Successfully unmarshalled %d RetiredInstanceInfo records. Processing them to update definition summaries is pending full implementation.", len(retiredInfos)) // Changed from Infof
	// TODO: Iterate through retiredInfos. For each info:
	//   1. Find the corresponding AIWorkerDefinition using DefinitionID.
	//   2. If found, iterate through its PerformanceRecords.
	//   3. For each PerformanceRecord, call a method similar to logPerformanceRecordUnsafe (or its logic)
	//      to update the AggregatePerformanceSummary on that definition.
	//      This update must be done carefully, especially if the summary was already partially populated.
	//      It might involve resetting parts of the summary and recalculating from all known records (complex),
	//      or ensuring new records are additive if that's the design.
	// For now, this method only unmarshals. The re-aggregation logic needs to be robust.
	// If definitions are loaded *after* this, then their summaries might already be from definitions.json.
	// If this is for loading an independent performance log, then aggregation is key.

	// After potentially updating many definitions' summaries, rate trackers might also need a refresh
	// if their initial state depends on these summaries (which it does for ActiveInstancesCount).
	// m.initializeRateTrackersUnsafe() // Call under the same lock as definition modifications.

	return nil
}

// prepareRetiredInstanceForAppending takes existing JSON string content of performance data,
// appends new RetiredInstanceInfo, and returns the new marshalled JSON string.
// Caller should hold the main manager lock (m.mu.Lock()) if reading/modifying shared state related to definitions.
func (m *AIWorkerManager) prepareRetiredInstanceForAppending(existingJsonContent string, instanceInfoToAdd *RetiredInstanceInfo) (string, error) {
	if instanceInfoToAdd == nil {
		return existingJsonContent, NewRuntimeError(ErrorCodeArgMismatch, "instanceInfoToAdd cannot be nil", ErrInvalidArgument)
	}

	var allInfos []*RetiredInstanceInfo
	if existingJsonContent != "" && existingJsonContent != "null" { // "null" can be valid JSON for an empty array/object if unmarshalled into empty struct
		if err := json.Unmarshal([]byte(existingJsonContent), &allInfos); err != nil {
			// If existing content is invalid, decide policy: overwrite or error?
			// For append, we probably should error or log and start fresh.
			m.logger.Errorf("prepareRetiredInstanceForAppending: Failed to unmarshal existing performance data JSON: '%s'. Error: %v. Will attempt to save only new record.", existingJsonContent, err)
			// Fallback: create a new list with just the new record if unmarshal fails
			allInfos = []*RetiredInstanceInfo{instanceInfoToAdd}
		} else {
			allInfos = append(allInfos, instanceInfoToAdd)
		}
	} else {
		// No existing content, or it's explicitly empty/null, start a new list
		allInfos = []*RetiredInstanceInfo{instanceInfoToAdd}
	}

	newData, err := json.MarshalIndent(allInfos, "", "  ")
	if err != nil {
		m.logger.Errorf("prepareRetiredInstanceForAppending: Failed to marshal updated performance data: %v", err)
		return "", NewRuntimeError(ErrorCodeInternal, "failed to marshal updated performance data", err)
	}
	m.logger.Debugf("Successfully prepared performance data for appending. Total records now: %d.", len(allInfos))
	return string(newData), nil
}

// smartTrim is a general utility.
func smartTrim(s string, length int) string {
	if len(s) <= length {
		return s
	}
	if length < 3 {
		if length <= 0 {
			return ""
		}
		return s[:length]
	}
	return s[:length-3] + "..."
}

// ifErrorToString is a general utility.
func ifErrorToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
