// NeuroScript Version: 0.3.1
// File version: 0.2.4
// filename: pkg/core/ai_wm.go
// nlines: 265 // Approximate
// risk_rating: HIGH
// Changes:
// - Auto-generate DefinitionID in loadWorkerDefinitionsFromContent if missing (log as DEBUG).
// - Added warning for duplicate definition names in loadWorkerDefinitionsFromContent.
// - Changed INFO logs to DEBUG
package core

import (
	"encoding/json"
	"fmt"
	"os" // Added for os.Getenv
	"path/filepath"
	"sync"
	"time"

	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/google/uuid" // Ensure this import is present for UUID generation
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
		return nil, fmt.Errorf("logger cannot be nil for AIWorkerManager")
	}
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
		if err := m.loadWorkerDefinitionsFromContent([]byte(initialDefinitionsContent)); err != nil {
			logger.Errorf("AIWorkerManager: Failed to load definitions from initial content: %v. Proceeding with empty definitions.", err)
		}
	} else {
		logger.Debugf("AIWorkerManager: No initial definitions content provided. Starting with an empty set of definitions.")
	}

	if initialPerformanceContent != "" {
		logger.Debugf("AIWorkerManager attempting to load performance data from provided initial content (length: %d).", len(initialPerformanceContent))
		if err := m.loadRetiredInstancePerformanceDataFromContent([]byte(initialPerformanceContent)); err != nil {
			logger.Errorf("AIWorkerManager: Failed to load performance data from initial content: %v.", err)
		}
	} else {
		logger.Debugf("AIWorkerManager: No initial performance data content provided.")
	}

	m.initializeRateTrackersUnsafe()
	// This Info log is appropriate as it's a summary of the constructor's action.
	m.logger.Infof("AIWorkerManager initialized. Loaded %d definitions. Active instances: %d. Sandbox context: '%s'", len(m.definitions), len(m.activeInstances), m.sandboxDir)
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
	if len(jsonBytes) == 0 {
		m.logger.Debugf("loadWorkerDefinitionsFromContent: Provided content is empty. No definitions loaded.")
		m.definitions = make(map[string]*AIWorkerDefinition) // Reset to empty
		return nil
	}

	var defs []*AIWorkerDefinition
	if err := json.Unmarshal(jsonBytes, &defs); err != nil {
		m.logger.Errorf("loadWorkerDefinitionsFromContent: Failed to unmarshal definitions JSON: %v", err)
		m.definitions = make(map[string]*AIWorkerDefinition)
		return NewRuntimeError(ErrorCodeInternal, "failed to unmarshal definitions data from content", err)
	}

	newDefinitions := make(map[string]*AIWorkerDefinition)
	namesEncountered := make(map[string]string) // To track names for duplicate warnings: name -> first DefinitionID

	for _, def := range defs {
		if def == nil {
			m.logger.Warnf("loadWorkerDefinitionsFromContent: Encountered a nil definition in content. Skipping.")
			continue
		}

		originalID := def.DefinitionID
		currentName := def.Name // Store name before potentially skipping due to true duplication

		if def.DefinitionID == "" {
			newID := uuid.NewString()
			m.logger.Debugf("loadWorkerDefinitionsFromContent: Definition (Name: '%s') has empty ID. Assigning new ID: %s", def.Name, newID)
			def.DefinitionID = newID
		}

		// Check for duplicate names
		if existingDefID, nameFound := namesEncountered[def.Name]; nameFound {
			// Name found. If IDs are different, it's a problematic duplicate.
			// If IDs are the same (e.g. from a misconfigured file with multiple identical entries),
			// the map assignment below will handle it (last one wins), but a warning is still good.
			if existingDefID != def.DefinitionID {
				m.logger.Warnf("loadWorkerDefinitionsFromContent: Duplicate AIWorkerDefinition name '%s' encountered. Existing ID: '%s', New/Current ID: '%s'. Ensure names are unique if distinct workers are intended, or consolidate if they are the same worker.", def.Name, existingDefID, def.DefinitionID)
			} else {
				// Same name, same ID - likely a full duplicate entry in the JSON
				m.logger.Warnf("loadWorkerDefinitionsFromContent: Duplicate entry for AIWorkerDefinition (Name: '%s', ID: '%s'). Overwriting with current entry.", def.Name, def.DefinitionID)
			}
		} else {
			namesEncountered[def.Name] = def.DefinitionID
		}

		// Check if this specific definition (by ID) has already been processed (e.g. if the JSON has exact duplicate objects)
		if _, idExists := newDefinitions[def.DefinitionID]; idExists && originalID != "" {
			// This means an explicit ID was duplicated in the source JSON.
			// The namesEncountered check would have caught logical duplicates by name with different IDs.
			m.logger.Warnf("loadWorkerDefinitionsFromContent: AIWorkerDefinition ID '%s' (Name: '%s') appears multiple times in the source. The last occurrence will be used.", def.DefinitionID, currentName)
		}

		if def.Status == "" {
			def.Status = DefinitionStatusActive
			m.logger.Debugf("Definition (Name: '%s', ID: '%s') had empty status, defaulted to '%s'.", def.Name, def.DefinitionID, def.Status)
		}
		newDefinitions[def.DefinitionID] = def
	}

	m.definitions = newDefinitions
	m.logger.Debugf("Successfully loaded/reloaded %d worker definitions from content.", len(m.definitions))
	return nil
}

// prepareDefinitionsForSaving marshals the current AI worker definitions to a JSON string.
func (m *AIWorkerManager) prepareDefinitionsForSaving() (string, error) {
	defsToSave := make([]*AIWorkerDefinition, 0, len(m.definitions))
	for _, def := range m.definitions {
		if def == nil {
			continue
		}
		if def.AggregatePerformanceSummary == nil {
			def.AggregatePerformanceSummary = &AIWorkerPerformanceSummary{}
			m.logger.Warnf("prepareDefinitionsForSaving: Definition (Name: '%s', ID: '%s') had nil AggregatePerformanceSummary; initialized.", def.Name, def.DefinitionID)
		}
		if tracker, ok := m.rateTrackers[def.DefinitionID]; ok {
			def.AggregatePerformanceSummary.ActiveInstancesCount = tracker.CurrentActiveInstances
		} else {
			def.AggregatePerformanceSummary.ActiveInstancesCount = 0
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
	m.logger.Debugf("Resolving API key with method: %s", auth.Method)
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
		key := os.Getenv(auth.Value)
		if key == "" {
			err := NewRuntimeError(ErrorCodeConfiguration,
				fmt.Sprintf("environment variable '%s' for API key not found or is empty", auth.Value),
				ErrConfiguration,
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
func (m *AIWorkerManager) initializeRateTrackersUnsafe() {
	newRateTrackers := make(map[string]*WorkerRateTracker)
	for defID, def := range m.definitions {
		if def == nil {
			m.logger.Warnf("initializeRateTrackersUnsafe: Encountered nil definition for ID '%s' (Name: '%s'). Skipping tracker initialization.", defID, def.Name)
			continue
		}
		activeCount := 0
		if def.AggregatePerformanceSummary != nil && def.AggregatePerformanceSummary.ActiveInstancesCount > 0 {
			activeCount = def.AggregatePerformanceSummary.ActiveInstancesCount
		}
		newRateTrackers[defID] = &WorkerRateTracker{
			DefinitionID:           defID,
			RequestsMinuteMarker:   time.Now(),
			TokensMinuteMarker:     time.Now(),
			TokensDayMarker:        time.Now(),
			CurrentActiveInstances: activeCount,
		}
		m.logger.Debugf("Initialized rate tracker for Definition (Name: '%s', ID: %s), ActiveInstances from summary: %d", def.Name, defID, activeCount)
	}
	m.rateTrackers = newRateTrackers
	m.logger.Debugf("Re-initialized all rate trackers. Total count: %d", len(m.rateTrackers))
}

// loadRetiredInstancePerformanceDataFromContent loads and processes performance data.
func (m *AIWorkerManager) loadRetiredInstancePerformanceDataFromContent(jsonBytes []byte) error {
	m.logger.Debug("loadRetiredInstancePerformanceDataFromContent called.")
	if len(jsonBytes) == 0 {
		m.logger.Debugf("loadRetiredInstancePerformanceDataFromContent: Provided content is empty. No historical performance loaded or processed.")
		return nil
	}
	var retiredInfos []*RetiredInstanceInfo
	if err := json.Unmarshal(jsonBytes, &retiredInfos); err != nil {
		m.logger.Errorf("loadRetiredInstancePerformanceDataFromContent: Failed to unmarshal performance data JSON: %v", err)
		return NewRuntimeError(ErrorCodeInternal, "failed to unmarshal performance data from content", err)
	}
	m.logger.Debugf("Successfully unmarshalled %d RetiredInstanceInfo records. Processing them to update definition summaries is pending full implementation.", len(retiredInfos))
	return nil
}

// prepareRetiredInstanceForAppending takes existing JSON string content of performance data,
func (m *AIWorkerManager) prepareRetiredInstanceForAppending(existingJsonContent string, instanceInfoToAdd *RetiredInstanceInfo) (string, error) {
	if instanceInfoToAdd == nil {
		return existingJsonContent, NewRuntimeError(ErrorCodeArgMismatch, "instanceInfoToAdd cannot be nil", ErrInvalidArgument)
	}
	var allInfos []*RetiredInstanceInfo
	if existingJsonContent != "" && existingJsonContent != "null" {
		if err := json.Unmarshal([]byte(existingJsonContent), &allInfos); err != nil {
			m.logger.Errorf("prepareRetiredInstanceForAppending: Failed to unmarshal existing performance data JSON: '%s'. Error: %v. Will attempt to save only new record.", existingJsonContent, err)
			allInfos = []*RetiredInstanceInfo{instanceInfoToAdd}
		} else {
			allInfos = append(allInfos, instanceInfoToAdd)
		}
	} else {
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

func (m *AIWorkerManager) GetSandboxDir() string {
	return m.sandboxDir
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
