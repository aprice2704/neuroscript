// NeuroScript Version: 0.3.1
// File version: 0.2.10
// filename: pkg/core/ai_wm.go
// Changes:
// - Ensured 'providerAllowsEmptyInlineKey' is correctly used in ListWorkerDefinitionsForDisplay.
// - Verified standard import paths.
package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/google/uuid"
)

const (
	defaultDefinitionsFile     = "ai_worker_definitions.json"
	defaultPerformanceDataFile = "ai_worker_performance_data.json"
	statelessInstanceIDPrefix  = "stateless-"
)

type AIWorkerManager struct {
	definitions     map[string]*AIWorkerDefinition
	activeInstances map[string]*AIWorkerInstance
	rateTrackers    map[string]*WorkerRateTracker

	definitionsBaseFilename     string
	performanceDataBaseFilename string
	sandboxDir                  string

	mu        sync.RWMutex
	logger    logging.Logger
	llmClient LLMClient
}

func NewAIWorkerManager(
	logger logging.Logger,
	sandboxDir string,
	llmClient LLMClient,
	initialDefinitionsContent string,
	initialPerformanceContent string,
) (*AIWorkerManager, error) {

	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil for AIWorkerManager")
	}
	if sandboxDir == "" {
		logger.Warn("AIWorkerManager: sandboxDir is empty during initialization.")
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
		logger.Debugf("AIWorkerManager: No initial definitions content provided.")
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

func (m *AIWorkerManager) loadWorkerDefinitionsFromContent(jsonBytes []byte) error {
	if len(jsonBytes) == 0 {
		m.logger.Debugf("loadWorkerDefinitionsFromContent: Provided content is empty. No definitions loaded.")
		m.definitions = make(map[string]*AIWorkerDefinition)
		return nil
	}
	var defs []*AIWorkerDefinition
	if err := json.Unmarshal(jsonBytes, &defs); err != nil {
		m.logger.Errorf("loadWorkerDefinitionsFromContent: Failed to unmarshal definitions JSON: %v", err)
		m.definitions = make(map[string]*AIWorkerDefinition)
		return NewRuntimeError(ErrorCodeInternal, "failed to unmarshal definitions data from content", err)
	}
	newDefinitions := make(map[string]*AIWorkerDefinition)
	namesEncountered := make(map[string]string)
	for _, def := range defs {
		if def == nil {
			m.logger.Warnf("loadWorkerDefinitionsFromContent: Encountered a nil definition. Skipping.")
			continue
		}
		originalID := def.DefinitionID
		currentName := def.Name
		if def.DefinitionID == "" {
			newID := uuid.NewString()
			m.logger.Debugf("loadWorkerDefinitionsFromContent: Definition (Name: '%s') has empty ID. Assigning new ID: %s", def.Name, newID)
			def.DefinitionID = newID
		}
		if existingDefID, nameFound := namesEncountered[def.Name]; nameFound {
			if existingDefID != def.DefinitionID {
				m.logger.Warnf("loadWorkerDefinitionsFromContent: Duplicate AIWorkerDefinition name '%s'. Existing ID: '%s', New ID: '%s'.", def.Name, existingDefID, def.DefinitionID)
			} else {
				m.logger.Warnf("loadWorkerDefinitionsFromContent: Duplicate entry for AIWorkerDefinition (Name: '%s', ID: '%s').", def.Name, def.DefinitionID)
			}
		} else {
			namesEncountered[def.Name] = def.DefinitionID
		}
		if _, idExists := newDefinitions[def.DefinitionID]; idExists && originalID != "" {
			m.logger.Warnf("loadWorkerDefinitionsFromContent: AIWorkerDefinition ID '%s' (Name: '%s') appears multiple times. Last occurrence used.", def.DefinitionID, currentName)
		}
		if def.Status == "" {
			def.Status = DefinitionStatusActive
			m.logger.Debugf("Definition (Name: '%s', ID: '%s') status defaulted to '%s'.", def.Name, def.DefinitionID, def.Status)
		}
		newDefinitions[def.DefinitionID] = def
	}
	m.definitions = newDefinitions
	m.logger.Debugf("Successfully loaded/reloaded %d worker definitions from content.", len(m.definitions))
	return nil
}

func (m *AIWorkerManager) prepareDefinitionsForSaving() (string, error) {
	defsToSave := make([]*AIWorkerDefinition, 0, len(m.definitions))
	for _, def := range m.definitions {
		if def == nil {
			continue
		}
		if def.AggregatePerformanceSummary == nil {
			def.AggregatePerformanceSummary = &AIWorkerPerformanceSummary{}
			m.logger.Warnf("prepareDefinitionsForSaving: Def (Name: '%s', ID: '%s') had nil AggregatePerformanceSummary; initialized.", def.Name, def.DefinitionID)
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
	m.logger.Debugf("Successfully prepared %d worker definitions for saving.", len(defsToSave))
	return string(data), nil
}

func (m *AIWorkerManager) resolveAPIKey(auth APIKeySource) (string, error) {
	m.logger.Debugf("Resolving API key with method: %s", auth.Method)
	switch auth.Method {
	case APIKeyMethodEnvVar:
		if auth.Value == "" {
			err := NewRuntimeError(ErrorCodeArgMismatch, "API key method 'env_var' but no env var name specified", ErrInvalidArgument)
			m.logger.Warnf("resolveAPIKey: %s", err.Message)
			return "", err
		}
		key := os.Getenv(auth.Value)
		if key == "" {
			err := NewRuntimeError(ErrorCodeConfiguration, fmt.Sprintf("env var '%s' for API key not found or empty", auth.Value), ErrAPIKeyNotFound)
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
		err := NewRuntimeError(ErrorCodeNotImplemented, errMessage, ErrFeatureNotImplemented)
		m.logger.Errorf("resolveAPIKey: %s", errMessage)
		return "", err
	default:
		errMessage := fmt.Sprintf("unknown API key source method: '%s'", auth.Method)
		err := NewRuntimeError(ErrorCodeArgMismatch, errMessage, ErrInvalidArgument)
		m.logger.Errorf("resolveAPIKey: %s", errMessage)
		return "", err
	}
}

func (m *AIWorkerManager) initializeRateTrackersUnsafe() {
	newRateTrackers := make(map[string]*WorkerRateTracker)
	for defID, def := range m.definitions {
		if def == nil {
			m.logger.Warnf("initializeRateTrackersUnsafe: Nil definition for ID '%s'. Skipping tracker.", defID)
			continue
		}
		activeCount := 0
		if def.AggregatePerformanceSummary != nil {
			activeCount = def.AggregatePerformanceSummary.ActiveInstancesCount
		}
		newRateTrackers[defID] = &WorkerRateTracker{
			DefinitionID:           defID,
			RequestsMinuteMarker:   time.Now(),
			TokensMinuteMarker:     time.Now(),
			TokensDayMarker:        time.Now(),
			CurrentActiveInstances: activeCount,
		}
		m.logger.Debugf("Initialized rate tracker for Def (Name: '%s', ID: %s), ActiveInstances: %d", def.Name, defID, activeCount)
	}
	m.rateTrackers = newRateTrackers
	m.logger.Debugf("Re-initialized all rate trackers. Total: %d", len(m.rateTrackers))
}

func (m *AIWorkerManager) loadRetiredInstancePerformanceDataFromContent(jsonBytes []byte) error {
	m.logger.Debug("loadRetiredInstancePerformanceDataFromContent called.")
	if len(jsonBytes) == 0 {
		m.logger.Debugf("loadRetiredInstancePerformanceDataFromContent: Empty content. No historical performance loaded.")
		return nil
	}
	var retiredInfos []*RetiredInstanceInfo
	if err := json.Unmarshal(jsonBytes, &retiredInfos); err != nil {
		m.logger.Errorf("loadRetiredInstancePerformanceDataFromContent: Failed to unmarshal performance data: %v", err)
		return NewRuntimeError(ErrorCodeInternal, "failed to unmarshal performance data from content", err)
	}
	m.logger.Debugf("Unmarshalled %d RetiredInstanceInfo records. Processing to update summaries pending.", len(retiredInfos))
	return nil
}

func (m *AIWorkerManager) prepareRetiredInstanceForAppending(existingJsonContent string, instanceInfoToAdd *RetiredInstanceInfo) (string, error) {
	if instanceInfoToAdd == nil {
		return existingJsonContent, NewRuntimeError(ErrorCodeArgMismatch, "instanceInfoToAdd cannot be nil", ErrInvalidArgument)
	}
	var allInfos []*RetiredInstanceInfo
	if existingJsonContent != "" && existingJsonContent != "null" {
		if err := json.Unmarshal([]byte(existingJsonContent), &allInfos); err != nil {
			m.logger.Errorf("prepareRetiredInstanceForAppending: Failed to unmarshal existing perf data: '%s'. Error: %v. Will save only new record.", existingJsonContent, err)
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
		return "", NewRuntimeError(ErrorCodeInternal, "failed to marshal updated performance data", err)
	}
	m.logger.Debugf("Prepared performance data for appending. Total records: %d.", len(allInfos))
	return string(newData), nil
}

func (m *AIWorkerManager) GetSandboxDir() string {
	return m.sandboxDir
}

// ListWorkerDefinitionsForDisplay retrieves AIWorkerDefinitions embellished with TUI-relevant status.
func (m *AIWorkerManager) ListWorkerDefinitionsForDisplay() ([]*AIWorkerDefinitionDisplayInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.definitions == nil {
		m.logger.Warn("ListWorkerDefinitionsForDisplay: definitions map is nil.")
		return []*AIWorkerDefinitionDisplayInfo{}, nil
	}

	displayInfos := make([]*AIWorkerDefinitionDisplayInfo, 0, len(m.definitions))

	for _, def := range m.definitions {
		if def == nil {
			m.logger.Warn("ListWorkerDefinitionsForDisplay: Encountered nil definition. Skipping.")
			continue
		}

		isChatCapable := false
		if len(def.InteractionModels) == 0 {
			isChatCapable = true
		} else {
			for _, modelType := range def.InteractionModels {
				if modelType == InteractionModelConversational || modelType == InteractionModelBoth {
					isChatCapable = true
					break
				}
			}
		}

		var apiKeyStatus APIKeyStatus
		resolvedKey, errResolve := "", error(nil)

		if def.Auth.Method == "" {
			apiKeyStatus = APIKeyStatusNotConfigured
		} else if def.Auth.Method == APIKeyMethodNone {
			apiKeyStatus = APIKeyStatusFound
		} else {
			resolvedKey, errResolve = m.resolveAPIKey(def.Auth)
			if errResolve != nil {
				if errors.Is(errResolve, ErrAPIKeyNotFound) {
					apiKeyStatus = APIKeyStatusNotFound
				} else if runErr, ok := errResolve.(*RuntimeError); ok {
					switch runErr.Code {
					case ErrorCodeConfiguration, ErrorCodeArgMismatch:
						apiKeyStatus = APIKeyStatusNotConfigured
						m.logger.Warnf("API key for def '%s' (method: %s) NotConfigured/NotFound due to: %s", def.Name, def.Auth.Method, runErr.Message)
					case ErrorCodeNotImplemented:
						apiKeyStatus = APIKeyStatusError
						m.logger.Warnf("API key method '%s' for def '%s' not implemented.", def.Auth.Method, def.Name)
					default:
						apiKeyStatus = APIKeyStatusError
						m.logger.Errorf("Unexpected runtime error resolving API key for def '%s': %v", def.Name, errResolve)
					}
				} else {
					apiKeyStatus = APIKeyStatusError
					m.logger.Errorf("Non-runtime error resolving API key for def '%s': %v", def.Name, errResolve)
				}
			} else { // No error from resolveAPIKey
				if def.Auth.Method == APIKeyMethodInline && resolvedKey == "" {
					// This variable IS used now to make the decision.
					providerAllowsEmptyInlineKey := false
					switch def.Provider {
					case ProviderGoogle, ProviderOpenAI, ProviderAnthropic:
						providerAllowsEmptyInlineKey = false
					case ProviderOllama:
						providerAllowsEmptyInlineKey = true
					default:
						providerAllowsEmptyInlineKey = false
						m.logger.Debugf("Def %s (Provider: %s) uses empty inline key. Defaulting to 'key not sufficient' (NotConfigured).", def.Name, def.Provider)
					}

					if providerAllowsEmptyInlineKey {
						apiKeyStatus = APIKeyStatusFound
						m.logger.Infof("Def %s (%s) uses inline auth with empty key, considered 'Found' as provider allows it.", def.Name, def.Provider)
					} else {
						apiKeyStatus = APIKeyStatusNotConfigured
						m.logger.Infof("Def %s (%s) uses inline auth with empty key, provider requires a key. Marked as NotConfigured.", def.Name, def.Provider)
					}
				} else if resolvedKey == "" && def.Auth.Method != APIKeyMethodNone {
					apiKeyStatus = APIKeyStatusNotFound
					m.logger.Warnf("Def %s (%s) resolved to empty key via method %s without error (or error was not ErrAPIKeyNotFound). Marked as %s.", def.Name, def.Provider, def.Auth.Method, apiKeyStatus)
				} else {
					apiKeyStatus = APIKeyStatusFound
				}
			}
		}

		displayInfos = append(displayInfos, &AIWorkerDefinitionDisplayInfo{
			Definition:    def,
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
	if length < 3 {
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
