// NeuroScript Version: 0.3.0
// File version: 0.1.3
// AI Worker Management: Performance Tracking and Persistence (Error and Type Conversion Corrected)
// filename: pkg/core/ai_wm_performance.go

package core

import (
	"encoding/json"
	"fmt"
	"os" // Required for filepath.Join if FullPathForPerformanceData was inlined, but it's on 'm'
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
	// "github.com/aprice2704/neuroscript/pkg/logging"
	// "github.com/google/uuid" // For TaskID if not provided externally
)

// logPerformanceRecordUnsafe is an internal method to log a performance record.
// It updates the aggregate summary on the corresponding AIWorkerDefinition.
// This method assumes the caller holds the necessary locks (typically Write Lock).
func (m *AIWorkerManager) logPerformanceRecordUnsafe(record *PerformanceRecord) error {
	if record == nil {
		return lang.NewRuntimeError(ErrorCodeArgMismatch, "cannot log a nil performance record", ErrInvalidArgument)
	}
	if record.DefinitionID == "" {
		return lang.NewRuntimeError(ErrorCodeArgMismatch, "performance record is missing DefinitionID", ErrInvalidArgument)
	}

	if record.TaskID == "" {
		m.logger.Warnf("Performance record for DefID %s is missing TaskID.", record.DefinitionID)
		// Depending on policy, we might want to assign a UUID here or reject.
		// For now, proceed with warning.
	}

	def, defExists := m.definitions[record.DefinitionID]
	if !defExists {
		m.logger.Warnf("Definition ID '%s' not found when trying to log performance record (TaskID: %s). Summary will not be updated.", record.DefinitionID, record.TaskID)
		return lang.NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("definition ID '%s' not found when logging performance record (TaskID: %s)", record.DefinitionID, record.TaskID), ErrNotFound)
	}

	summary := def.AggregatePerformanceSummary // summary is *AIWorkerPerformanceSummary
	if summary == nil {
		m.logger.Infof("Initializing nil AggregatePerformanceSummary for DefID: %s during performance logging.", def.DefinitionID)
		def.AggregatePerformanceSummary = &AIWorkerPerformanceSummary{}
		summary = def.AggregatePerformanceSummary // Point summary to the newly created struct
	}

	summary.TotalTasksAttempted++
	if record.Success {
		summary.SuccessfulTasks++
	} else {
		summary.FailedTasks++
	}

	if record.LLMMetrics != nil {
		if tokens, ok := record.LLMMetrics["input_tokens"]; ok {
			val, converted := toInt64(tokens) // Using existing helper
			if converted {
				summary.TotalTokensProcessed += val
			} else {
				m.logger.Warnf("Could not convert input_tokens '%v' to int64 for DefID %s, TaskID %s", tokens, record.DefinitionID, record.TaskID)
			}
		}
		if tokens, ok := record.LLMMetrics["output_tokens"]; ok {
			val, converted := toInt64(tokens) // Using existing helper
			if converted {
				summary.TotalTokensProcessed += val
			} else {
				m.logger.Warnf("Could not convert output_tokens '%v' to int64 for DefID %s, TaskID %s", tokens, record.DefinitionID, record.TaskID)
			}
		}
	}

	summary.TotalCostIncurred += record.CostIncurred

	if record.TimestampEnd.After(summary.LastActivityTimestamp) {
		summary.LastActivityTimestamp = record.TimestampEnd
	}

	if record.SupervisorFeedback != nil && record.SupervisorFeedback.Rating != 0 {
		// TODO: Implement rolling average calculation for AverageQualityScore if needed
		// For now, this field is more of a placeholder or manually set/interpreted.
		// summary.TotalQualityScoreSum += record.SupervisorFeedback.Rating
		// summary.QualityScoreRatedTasks++
	}

	if summary.TotalTasksAttempted > 0 {
		summary.AverageSuccessRate = float64(summary.SuccessfulTasks) / float64(summary.TotalTasksAttempted)
		// TODO: Implement rolling average for duration if needed
		// summary.AverageDurationMs = ((summary.AverageDurationMs * float64(summary.TotalTasksAttempted-1)) + float64(record.DurationMs)) / float64(summary.TotalTasksAttempted)
	} else {
		summary.AverageSuccessRate = 0
		// summary.AverageDurationMs = float64(record.DurationMs)
	}

	m.logger.Debugf("Performance logged and summary updated for DefID: %s, TaskID: %s. Success: %t", record.DefinitionID, record.TaskID, record.Success)
	return nil
}

// GetPerformanceRecordsForDefinition retrieves all persisted performance records associated with a definition.
func (m *AIWorkerManager) GetPerformanceRecordsForDefinition(definitionID string, filters map[string]interface{}) ([]*PerformanceRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if _, exists := m.definitions[definitionID]; !exists {
		return nil, lang.NewRuntimeError(ErrorCodeKeyNotFound, fmt.Sprintf("worker definition with ID '%s' not found when getting performance records", definitionID), ErrNotFound)
	}

	allRetiredData, err := m.loadAllRetiredInstanceDataUnsafe() // This now correctly calls the refactored path logic
	if err != nil {
		// loadAllRetiredInstanceDataUnsafe already logs specifics and returns appropriate error types
		if os.IsNotExist(err) { // This check might be redundant if loadAllRetiredInstanceDataUnsafe handles it
			m.logger.Debugf("GetPerformanceRecordsForDefinition: Performance data file not found. Returning empty list for DefID %s.", definitionID)
			return []*PerformanceRecord{}, nil
		}
		// Error is already logged by loadAllRetiredInstanceDataUnsafe
		return nil, err
	}

	var results []*PerformanceRecord
	for _, retiredInfo := range allRetiredData {
		if retiredInfo.DefinitionID == definitionID {
			for _, rec := range retiredInfo.PerformanceRecords {
				// Ensure rec is not nil before accessing, though it shouldn't be if data is clean
				if rec != nil && rec.DefinitionID == definitionID && m.matchesPerformanceRecordFilters(rec, filters) {
					results = append(results, rec)
				}
			}
		}
	}

	m.logger.Debugf("GetPerformanceRecordsForDefinition: Found %d records for DefID %s matching filters.", len(results), definitionID)
	return results, nil
}

// loadAllRetiredInstanceDataUnsafe loads all RetiredInstanceInfo from the JSON file.
// This method now uses FullPathForPerformanceData() to get the file path.
func (m *AIWorkerManager) loadAllRetiredInstanceDataUnsafe() ([]RetiredInstanceInfo, error) {
	perfPath := m.FullPathForPerformanceData()
	if perfPath == "" {
		m.logger.Error("Cannot load performance data: file path is not configured in AIWorkerManager.")
		return nil, lang.NewRuntimeError(ErrorCodeConfiguration, "performance data file path not configured", ErrConfiguration)
	}

	data, err := os.ReadFile(perfPath)
	if err != nil {
		if os.IsNotExist(err) {
			m.logger.Debugf("Performance data file '%s' not found, returning empty set.", perfPath)
			return []RetiredInstanceInfo{}, nil // Return nil error as per previous logic for os.IsNotExist
		}
		m.logger.Errorf("Error reading performance data file '%s': %v", perfPath, err)
		return nil, lang.NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("error reading performance data file '%s'", perfPath), err)
	}
	if len(data) == 0 {
		m.logger.Debugf("Performance data file '%s' is empty.", perfPath)
		return []RetiredInstanceInfo{}, nil
	}

	var allInfo []RetiredInstanceInfo
	if err := json.Unmarshal(data, &allInfo); err != nil {
		m.logger.Errorf("Failed to unmarshal performance data from '%s': %v", perfPath, err)
		return nil, lang.NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("failed to unmarshal performance data from '%s'", perfPath), err)
	}
	m.logger.Debugf("Successfully loaded %d RetiredInstanceInfo records from %s", len(allInfo), perfPath)
	return allInfo, nil
}

// matchesPerformanceRecordFilters checks if a single performance record matches the given criteria.
func (m *AIWorkerManager) matchesPerformanceRecordFilters(record *PerformanceRecord, filters map[string]interface{}) bool {
	if len(filters) == 0 {
		return true
	}
	if record == nil {
		return false
	}

	for key, expectedValue := range filters {
		filterKey := strings.ToLower(key)
		match := false

		switch filterKey {
		case "taskid", "task_id":
			if taskID, ok := expectedValue.(string); ok && record.TaskID == taskID {
				match = true
			}
		case "instanceid", "instance_id":
			if id, ok := expectedValue.(string); ok {
				if id == "stateless" && strings.HasPrefix(record.InstanceID, statelessInstanceIDPrefix) {
					match = true
				} else if record.InstanceID == id {
					match = true
				}
			}
		case "definitionid", "definition_id":
			if id, ok := expectedValue.(string); ok && record.DefinitionID == id {
				match = true
			}
		case "success":
			if success, ok := expectedValue.(bool); ok && record.Success == success {
				match = true
			}
		case "durationms_gt", "duration_ms_gt":
			if durVal, ok := toInt64(expectedValue); ok && record.DurationMs > durVal {
				match = true
			}
		case "durationms_lt", "duration_ms_lt":
			if durVal, ok := toInt64(expectedValue); ok && record.DurationMs < durVal {
				match = true
			}
		case "costincurred_gt", "cost_incurred_gt":
			if costVal, ok := toFloat64(expectedValue); ok && record.CostIncurred > costVal {
				match = true
			}
		case "costincurred_lt", "cost_incurred_lt":
			if costVal, ok := toFloat64(expectedValue); ok && record.CostIncurred < costVal {
				match = true
			}
		default:
			m.logger.Debugf("AIWorkerManager.matchesPerformanceRecordFilters: Unknown or unhandled filter key '%s'", filterKey)
		}

		if !match {
			return false
		}
	}
	return true
}
