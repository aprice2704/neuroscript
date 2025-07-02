// NeuroScript Version: 0.3.1
// File version: 0.2.13
// Purpose: Core AI Worker Manager. AIWorkerDefinitions are loaded and treated as immutable. Persistence for definitions is removed.
// filename: pkg/core/ai_wm.go
// nlines: 320 // Approximate
// risk_rating: MEDIUM

package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time" // Kept for initializeRateTrackersUnsafe and other operational timestamps

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/google/uuid"
)

const (
	// defaultDefinitionsFile is still used for loading definitions.
	defaultDefinitionsFile     = "ai_worker_definitions.json"
	defaultPerformanceDataFile = "ai_worker_performance_data.json"
	statelessInstanceIDPrefix  = "stateless-"
)

type AIWorkerManager struct {
	definitions     map[string]*AIWorkerDefinition // Loaded once, then immutable (except AggregatePerformanceSummary)
	activeInstances map[string]*AIWorkerInstance
	rateTrackers    map[string]*WorkerRateTracker

	definitionsBaseFilename     string
	performanceDataBaseFilename string
	sandboxDir                  string

	mu        sync.RWMutex
	logger    interfaces.Logger
	llmClient interfaces.LLMClient
}

// String() method is in core/ai_worker_stringers.go

func NewAIWorkerManager(
	logger interfaces.Logger,
	sandboxDir string,
	llmClient interfaces.LLMClient,
	initialDefinitionsContent string, // Content for ai_worker_definitions.json
	initialPerformanceContent string, // Content for ai_worker_performance_data.json
) (*AIWorkerManager, error) {

	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil for AIWorkerManager")
	}
	if sandboxDir == "" {
		logger.Warn("AIWorkerManager: sandboxDir is empty during initialization.")
		// Depending on policy, this might be an error if definitions are expected from a specific path.
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

	// Load definitions from provided content string first, if available.
	// This simulates loading from a pre-bundled configuration.
	if initialDefinitionsContent != "" {
		logger.Debugf("AIWorkerManager attempting to load definitions from provided initial content (length: %d).", len(initialDefinitionsContent))
		if err := m.loadWorkerDefinitionsFromContent([]byte(initialDefinitionsContent)); err != nil {
			logger.Errorf("AIWorkerManager: Failed to load definitions from initial content: %v. Proceeding with empty definitions.", err)
			// Ensure definitions map is clean if loading from content fails partially
			m.definitions = make(map[string]*AIWorkerDefinition)
		}
	} else {
		// If no initial content string, attempt to load from the default file path.
		// This is the standard way to load definitions if not bundled.
		logger.Debugf("AIWorkerManager: No initial definitions content string provided. Attempting to load from file: %s", m.FullPathForDefinitions())
		// LoadWorkerDefinitionsFromFile handles its own locking and re-initializes rate trackers.
		// No need for m.mu.Lock() here as LoadWorkerDefinitionsFromFile will acquire it.
		if err := m.LoadWorkerDefinitionsFromFile(); err != nil {
			// If LoadWorkerDefinitionsFromFile returns an error, it will log it.
			// The manager will proceed, possibly with no definitions if the file wasn't found or was invalid.
			logger.Warnf("AIWorkerManager: Error during initial load from file: %v. Manager may have no definitions.", err)
		}
	}

	// Load performance data (this doesn't affect definition immutability)
	if initialPerformanceContent != "" {
		logger.Debugf("AIWorkerManager attempting to load performance data from provided initial content (length: %d).", len(initialPerformanceContent))
		if err := m.loadRetiredInstancePerformanceDataFromContent([]byte(initialPerformanceContent)); err != nil {
			logger.Errorf("AIWorkerManager: Failed to load performance data from initial content: %v.", err)
		}
	} else {
		logger.Debugf("AIWorkerManager: No initial performance data content string provided. Will load from file if it exists.")
		// loadRetiredInstancePerformanceDataFromFile handles its own locking.
		if err := m.LoadRetiredInstancePerformanceDataFromFile(); err != nil {
			logger.Warnf("AIWorkerManager: Error during initial performance data load from file: %v.", err)
		}
	}

	// Initialize rate trackers based on the loaded definitions (or lack thereof).
	// This call is made after definitions are potentially loaded by one of the above paths.
	// It needs to be under a lock because it modifies m.rateTrackers and reads m.definitions.
	m.mu.Lock()
	m.initializeRateTrackersUnsafe()
	m.mu.Unlock()

	m.logger.Infof("AIWorkerManager initialized. Loaded %d definitions. Active instances: %d. Sandbox context: '%s'", len(m.definitions), len(m.activeInstances), m.sandboxDir)
	return m, nil
}

func (m *AIWorkerManager) FullPathForDefinitions() string {
	if m.sandboxDir == "" || m.definitionsBaseFilename == "" {
		m.logger.Warnf("Cannot determine full path for definitions: sandboxDir ('%s') or baseFilename ('%s') is empty.", m.sandboxDir, m.definitionsBaseFilename)
		return ""
	}
	return filepath.Join(m.sandboxDir, m.definitionsBaseFilename)
}

func (m *AIWorkerManager) FullPathForPerformanceData() string {
	if m.sandboxDir == "" || m.performanceDataBaseFilename == "" {
		m.logger.Warnf("Cannot determine full path for performance data: sandboxDir ('%s') or baseFilename ('%s') is empty.", m.sandboxDir, m.performanceDataBaseFilename)
		return ""
	}
	return filepath.Join(m.sandboxDir, m.performanceDataBaseFilename)
}

// loadWorkerDefinitionsFromContent populates the definitions map from JSON bytes.
// This function assumes it's called under a write lock (m.mu.Lock()) if m.definitions is being modified.
// It replaces any existing definitions.
func (m *AIWorkerManager) loadWorkerDefinitionsFromContent(jsonBytes []byte) error {
	if len(jsonBytes) == 0 {
		m.logger.Debugf("loadWorkerDefinitionsFromContent: Provided content is empty. Definitions map cleared.")
		m.definitions = make(map[string]*AIWorkerDefinition) // Clear existing definitions
		return nil
	}

	var defs []*AIWorkerDefinition
	if err := json.Unmarshal(jsonBytes, &defs); err != nil {
		m.logger.Errorf("loadWorkerDefinitionsFromContent: Failed to unmarshal definitions JSON: %v", err)
		m.definitions = make(map[string]*AIWorkerDefinition) // Ensure map is clean on error
		return lang.NewRuntimeError(lang.ErrorCodeInternal, "failed to unmarshal definitions data from content", err)
	}

	newDefinitions := make(map[string]*AIWorkerDefinition)
	namesEncountered := make(map[string]string) // To check for duplicate names

	for _, def := range defs {
		if def == nil {
			m.logger.Warnf("loadWorkerDefinitionsFromContent: Encountered a nil definition. Skipping.")
			continue
		}
		if def.DefinitionID == "" {
			// This case should ideally be handled by well-formed definition files.
			// If IDs are missing, it implies the source JSON is not complete or a very old format.
			// For robustness, we could assign a new one, but it's better if definitions are complete.
			newID := uuid.NewString()
			m.logger.Warnf("loadWorkerDefinitionsFromContent: Definition (Name: '%s') has empty ID. Assigning new temporary ID for this load: %s. Source JSON should provide IDs.", def.Name, newID)
			def.DefinitionID = newID
		}

		// Check for duplicate names, which could cause issues in lookups by name.
		if existingDefID, nameFound := namesEncountered[def.Name]; nameFound {
			m.logger.Warnf("loadWorkerDefinitionsFromContent: Duplicate AIWorkerDefinition name '%s' encountered. Original ID: '%s', current definition ID: '%s'. This may lead to ambiguity if looking up by name.", def.Name, existingDefID, def.DefinitionID)
		} else {
			namesEncountered[def.Name] = def.DefinitionID
		}

		// Check for duplicate IDs.
		if _, idExists := newDefinitions[def.DefinitionID]; idExists {
			// This is a more serious issue if IDs are meant to be unique.
			m.logger.Errorf("loadWorkerDefinitionsFromContent: Duplicate AIWorkerDefinition ID '%s' (Name: '%s'). Overwriting with the last encountered definition with this ID. Please ensure definition IDs are unique in the source.", def.DefinitionID, def.Name)
		}

		if def.Status == "" {
			def.Status = DefinitionStatusActive // Default to active if not specified
			m.logger.Debugf("Definition (Name: '%s', ID: '%s') status defaulted to '%s'.", def.Name, def.DefinitionID, def.Status)
		}
		// Ensure AggregatePerformanceSummary is initialized
		if def.AggregatePerformanceSummary == nil {
			def.AggregatePerformanceSummary = &AIWorkerPerformanceSummary{}
		}

		newDefinitions[def.DefinitionID] = def
	}
	m.definitions = newDefinitions // Replace the old map with the newly loaded one
	m.logger.Debugf("Successfully loaded %d worker definitions from content.", len(m.definitions))
	return nil
}

// resolveAPIKey resolves the API key based on the Auth configuration.
// This method does not modify AIWorkerManager state and does not require locks on `m.mu`.
func (m *AIWorkerManager) resolveAPIKey(auth APIKeySource) (string, error) {
	m.logger.Debugf("Resolving API key with method: %s", auth.Method)
	switch auth.Method {
	case APIKeyMethodEnvVar:
		if auth.Value == "" {
			err := lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "API key method 'env_var' but no env var name specified", lang.ErrInvalidArgument)
			m.logger.Warnf("resolveAPIKey: %s", err.Message)
			return "", err
		}
		key := os.Getenv(auth.Value)
		if key == "" {
			err := lang.NewRuntimeError(lang.ErrorCodeConfiguration, fmt.Sprintf("env var '%s' for API key not found or empty", auth.Value), lang.ErrAPIKeyNotFound)
			m.logger.Warnf("resolveAPIKey: Env var '%s' not found or empty.", auth.Value)
			return "", err
		}
		m.logger.Debugf("Resolved API key from env var '%s'", auth.Value)
		return key, nil
	case APIKeyMethodInline:
		if auth.Value == "" {
			m.logger.Debugf("API key method is '%s' but key value is empty. May be acceptable for some models.", APIKeyMethodInline)
		} else {
			m.logger.Debugf("Using inline API key.")
		}
		return auth.Value, nil
	case APIKeyMethodNone:
		m.logger.Debugf("API key method is '%s', no key required.", APIKeyMethodNone)
		return "", nil
	case APIKeyMethodConfigPath, APIKeyMethodVault:
		errMessage := fmt.Sprintf("API key method '%s' not yet implemented", auth.Method)
		err := lang.NewRuntimeError(lang.ErrorCodeNotImplemented, errMessage, lang.ErrFeatureNotImplemented)
		m.logger.Errorf("resolveAPIKey: %s", errMessage)
		return "", err
	default:
		errMessage := fmt.Sprintf("unknown API key source method: '%s'", auth.Method)
		err := lang.NewRuntimeError(lang.ErrorCodeArgMismatch, errMessage, lang.ErrInvalidArgument)
		m.logger.Errorf("resolveAPIKey: %s", errMessage)
		return "", err
	}
}

// initializeRateTrackersUnsafe re-initializes rate trackers for all current definitions.
// Assumes caller holds the write lock (m.mu.Lock()).
func (m *AIWorkerManager) initializeRateTrackersUnsafe() {
	newRateTrackers := make(map[string]*WorkerRateTracker)
	for defID, def := range m.definitions {
		if def == nil { // Should not happen with proper loading
			m.logger.Warnf("initializeRateTrackersUnsafe: Nil definition for ID '%s'. Skipping tracker.", defID)
			continue
		}
		activeCount := 0
		// AggregatePerformanceSummary should be initialized during load if nil
		if def.AggregatePerformanceSummary != nil {
			// Preserve existing active count if tracker is being re-initialized,
			// otherwise it defaults to 0 if it's a fresh load.
			// For simplicity, on full re-init, reset to 0 unless we have a more complex state sync.
			// However, this field is also updated by instance spawning/retiring.
			// For now, let's get it from an existing tracker if one existed.
			if existingTracker, ok := m.rateTrackers[defID]; ok {
				activeCount = existingTracker.CurrentActiveInstances
			}
		}

		newRateTrackers[defID] = &WorkerRateTracker{
			DefinitionID:           defID,
			RequestsLastMinute:     0,
			TokensLastMinute:       0,
			TokensToday:            0,
			RequestsMinuteMarker:   time.Now(),
			TokensMinuteMarker:     time.Now(),
			TokensDayMarker:        time.Now(),
			CurrentActiveInstances: activeCount, // Preserve if possible, else 0
		}
		m.logger.Debugf("Initialized rate tracker for Def (Name: '%s', ID: %s), ActiveInstances: %d", def.Name, defID, activeCount)
	}
	m.rateTrackers = newRateTrackers
	m.logger.Debugf("Re-initialized all rate trackers. Total: %d", len(m.rateTrackers))
}

// loadRetiredInstancePerformanceDataFromContent processes performance data.
// Assumes caller holds lock if m.definitions is read (e.g. to update summaries).
func (m *AIWorkerManager) loadRetiredInstancePerformanceDataFromContent(jsonBytes []byte) error {
	m.logger.Debug("loadRetiredInstancePerformanceDataFromContent called.")
	if len(jsonBytes) == 0 {
		m.logger.Debugf("loadRetiredInstancePerformanceDataFromContent: Empty content. No historical performance loaded.")
		return nil
	}
	var retiredInfos []*RetiredInstanceInfo
	if err := json.Unmarshal(jsonBytes, &retiredInfos); err != nil {
		m.logger.Errorf("loadRetiredInstancePerformanceDataFromContent: Failed to unmarshal performance data: %v", err)
		return lang.NewRuntimeError(lang.ErrorCodeInternal, "failed to unmarshal performance data from content", err)
	}

	// Future: Could iterate through retiredInfos and update AggregatePerformanceSummary
	// on the corresponding m.definitions if they exist. This would require careful locking
	// if done outside initial load. For now, this simply loads them for potential later use
	// or assumes other mechanisms update summaries.
	m.logger.Debugf("Unmarshalled %d RetiredInstanceInfo records. Further processing (e.g., updating definition summaries) is not implemented in this function but could be added.", len(retiredInfos))
	return nil
}

// prepareRetiredInstanceForAppending prepares JSON string for appending performance data.
// Does not modify manager state directly.
func (m *AIWorkerManager) prepareRetiredInstanceForAppending(existingJsonContent string, instanceInfoToAdd *RetiredInstanceInfo) (string, error) {
	if instanceInfoToAdd == nil {
		return existingJsonContent, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "instanceInfoToAdd cannot be nil", lang.ErrInvalidArgument)
	}
	var allInfos []*RetiredInstanceInfo
	if existingJsonContent != "" && existingJsonContent != "null" { // "null" can be valid JSON for an empty array/object
		if err := json.Unmarshal([]byte(existingJsonContent), &allInfos); err != nil {
			m.logger.Errorf("prepareRetiredInstanceForAppending: Failed to unmarshal existing perf data (length: %d). Error: %v. Will save only new record.", len(existingJsonContent), err)
			// Potentially problematic if existing content is corrupted.
			// Decide policy: overwrite, error out, or append as new array.
			// Current: assumes it's an array and appends, or starts new if unmarshal fails badly.
			allInfos = []*RetiredInstanceInfo{instanceInfoToAdd}
		} else {
			allInfos = append(allInfos, instanceInfoToAdd)
		}
	} else {
		allInfos = []*RetiredInstanceInfo{instanceInfoToAdd}
	}
	newData, err := json.MarshalIndent(allInfos, "", "  ")
	if err != nil {
		m.logger.Errorf("prepareRetiredInstanceForAppending: Failed to marshal updated perf data: %v", err)
		return "", lang.NewRuntimeError(lang.ErrorCodeInternal, "failed to marshal updated performance data", err)
	}
	m.logger.Debugf("Prepared performance data for appending. Total records: %d.", len(allInfos))
	return string(newData), nil
}

func (m *AIWorkerManager) GetSandboxDir() string {
	return m.sandboxDir
}

// ListWorkerDefinitionsForDisplay provides information suitable for display.
// It reads m.definitions and m.rateTrackers, so it requires a read lock.
func (m *AIWorkerManager) ListWorkerDefinitionsForDisplay() ([]*AIWorkerDefinitionDisplayInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.definitions == nil { // Should be initialized in constructor
		m.logger.Warn("ListWorkerDefinitionsForDisplay: definitions map is nil.")
		return []*AIWorkerDefinitionDisplayInfo{}, nil
	}

	allDefs := make([]*AIWorkerDefinition, 0, len(m.definitions))
	for _, def := range m.definitions {
		if def != nil { // Should always be non-nil if loaded correctly
			allDefs = append(allDefs, def)
		}
	}

	sort.Slice(allDefs, func(i, j int) bool {
		nameI := strings.ToLower(allDefs[i].Name)
		nameJ := strings.ToLower(allDefs[j].Name)
		if nameI != nameJ {
			return nameI < nameJ
		}
		return allDefs[i].DefinitionID < allDefs[j].DefinitionID // Secondary sort by ID for stability
	})

	displayInfos := make([]*AIWorkerDefinitionDisplayInfo, 0, len(allDefs))

	for _, def := range allDefs {
		// Create a copy for the display info to avoid races if the original's
		// AggregatePerformanceSummary is updated concurrently (though less likely with RLock).
		// More importantly, DisplayInfo might hold transient status not part of the core def.
		defCopy := *def
		if def.AggregatePerformanceSummary != nil {
			summaryCopy := *def.AggregatePerformanceSummary
			defCopy.AggregatePerformanceSummary = &summaryCopy
		} else {
			defCopy.AggregatePerformanceSummary = &AIWorkerPerformanceSummary{} // Should be initialized on load
		}

		// Update ActiveInstancesCount from rate tracker for the display copy
		if tracker, ok := m.rateTrackers[def.DefinitionID]; ok {
			defCopy.AggregatePerformanceSummary.ActiveInstancesCount = tracker.CurrentActiveInstances
		}

		isChatCapable := false
		if len(defCopy.InteractionModels) == 0 {
			// Defaulting to true if not specified, as per original logic.
			// This could also be set during loadWorkerDefinitionsFromContent if a stricter default is desired.
			isChatCapable = true
			m.logger.Debugf("Definition '%s' has no InteractionModels specified, defaulting to IsChatCapable=true for display", defCopy.Name)
		} else {
			for _, modelType := range defCopy.InteractionModels {
				if modelType == InteractionModelConversational || modelType == InteractionModelBoth {
					isChatCapable = true
					break
				}
			}
		}

		var apiKeyStatus APIKeyStatus
		// API key resolution doesn't modify manager state, so it's safe to call here.
		resolvedKey, errResolve := "", error(nil)

		if defCopy.Auth.Method == "" {
			apiKeyStatus = APIKeyStatusNotConfigured
		} else if defCopy.Auth.Method == APIKeyMethodNone {
			apiKeyStatus = APIKeyStatusFound // "Found" in the sense that "None" is a valid, resolved state.
		} else {
			resolvedKey, errResolve = m.resolveAPIKey(defCopy.Auth)
			if errResolve != nil {
				if errors.Is(errResolve, lang.ErrAPIKeyNotFound) {
					apiKeyStatus = APIKeyStatusNotFound
				} else if runErr, ok := errResolve.(*lang.RuntimeError); ok {
					switch runErr.Code {
					case lang.ErrorCodeConfiguration, ErrorCodeArgMismatch:
						apiKeyStatus = APIKeyStatusNotConfigured
					case ErrorCodeNotImplemented:
						apiKeyStatus = APIKeyStatusError // Method known but not usable
					default:
						apiKeyStatus = APIKeyStatusError // Other resolution error
					}
				} else {
					apiKeyStatus = APIKeyStatusError // Unexpected error type
				}
				if apiKeyStatus != APIKeyStatusNotFound { // Log if not simply "not found"
					m.logger.Warnf("API key resolution for def '%s' (method: %s) resulted in status %s: %v", defCopy.Name, defCopy.Auth.Method, apiKeyStatus, errResolve)
				}
			} else { // No error resolving
				if defCopy.Auth.Method == APIKeyMethodInline && resolvedKey == "" {
					// Provider-specific check for empty inline key
					providerAllowsEmptyInlineKey := defCopy.Provider == ProviderOllama // Example: Ollama might allow empty
					if providerAllowsEmptyInlineKey {
						apiKeyStatus = APIKeyStatusFound
					} else {
						apiKeyStatus = APIKeyStatusNotConfigured // Empty inline key but provider needs one
						m.logger.Infof("Def %s (%s) uses inline auth with empty key, but provider likely requires a key. Marked as NotConfigured.", defCopy.Name, defCopy.Provider)
					}
				} else if resolvedKey == "" && defCopy.Auth.Method != APIKeyMethodNone {
					// Resolved to empty key via a method that should return a key (e.g., EnvVar existed but was empty)
					apiKeyStatus = APIKeyStatusNotFound
					m.logger.Warnf("Def %s (%s) resolved to empty key via method %s. Marked as NotFound.", defCopy.Name, defCopy.Provider, defCopy.Auth.Method)
				} else {
					apiKeyStatus = APIKeyStatusFound
				}
			}
		}

		displayInfos = append(displayInfos, &AIWorkerDefinitionDisplayInfo{
			Definition:    &defCopy, // Pass the copy
			IsChatCapable: isChatCapable,
			APIKeyStatus:  apiKeyStatus,
		})
	}
	return displayInfos, nil
}

func smartTrim(s string, length int) string {
	if len(s) <= length {
		return s
	}
	if length < 3 { // Adjusted for "..."
		if length <= 0 {
			return ""
		}
		return s[:length]
	}
	return s[:length-3] + "..."
}

func ifErrorToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
