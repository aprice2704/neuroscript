// NeuroScript Version: 0.4.0
// File version: 0.1.1
// Purpose: Corrected Sprintf formatting errors in ColourString methods.
// filename: pkg/core/ai_worker_stringers.go
// nlines: 850
// risk_rating: LOW

package core

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// --- Stringer Color Constants ---
const (
	colLabel      = "[blue::b]"
	colValue      = "[white]"
	colID         = "[yellow]"
	colName       = "[gold]"
	colStatusOk   = "[green]"
	colStatusErr  = "[red]"
	colStatusWarn = "[orange]"
	colStatusNeut = "[aqua]"
	colCount      = "[purple]"
	colBool       = "[cyan]"
	colTime       = "[grey]"
	colReset      = "[-]"
)

// --- AIWorkerDefinitionDisplayInfo ---

func (di *AIWorkerDefinitionDisplayInfo) String() string {
	if di == nil {
		return "<nil AIWorkerDefinitionDisplayInfo>"
	}
	var defStr string
	if di.Definition != nil {
		defStr = fmt.Sprintf("Def Name: %s (ID: %s, Status: %s)", di.Definition.Name, di.Definition.DefinitionID, di.Definition.Status)
	} else {
		defStr = "<nil Definition>"
	}
	return fmt.Sprintf("DisplayInfo: Capable: %t, KeyStatus: %s, Def: %s", di.IsChatCapable, di.APIKeyStatus, defStr)
}

func (di *AIWorkerDefinitionDisplayInfo) ColourString() string {
	if di == nil {
		return colStatusErr + "<nil AIWorkerDefinitionDisplayInfo>" + colReset
	}
	var defStr string
	if di.Definition != nil {
		// This Sprintf is for a part of the string, not the one causing major lint errors.
		defStr = fmt.Sprintf("%sDef Name:%s %s%s%s (%sID:%s %s%s%s, %sStatus:%s %s%s%s)",
			colLabel, colReset, colName, di.Definition.Name, colReset,
			colLabel, colReset, colID, di.Definition.DefinitionID, colReset,
			colLabel, colReset, colValue, di.Definition.Status, colReset)
	} else {
		defStr = colStatusErr + "<nil Definition>" + colReset
	}
	return fmt.Sprintf("%sDisplayInfo:%s %sCapable:%s %s%t%s, %sKeyStatus:%s %s%s%s, %sDef:%s %s",
		colLabel, colReset, colLabel, colReset, colBool, di.IsChatCapable, colReset,
		colLabel, colReset, colValue, di.APIKeyStatus, colReset,
		colLabel, colReset, defStr)
}

// --- APIKeySource ---

func (aks *APIKeySource) String() string {
	if aks == nil {
		return "<nil APIKeySource>"
	}
	val := aks.Value
	if aks.Method == APIKeyMethodInline && val != "" {
		val = "[redacted]"
	}
	return fmt.Sprintf("Method: %s, Value: '%s'", aks.Method, val)
}

func (aks *APIKeySource) ColourString() string {
	if aks == nil {
		return colStatusErr + "<nil APIKeySource>" + colReset
	}
	val := aks.Value
	valCol := colValue
	if aks.Method == APIKeyMethodInline && val != "" {
		val = "[redacted]"
		valCol = colStatusWarn
	}
	return fmt.Sprintf("%sMethod:%s %s%s%s, %sValue:%s '%s%s%s'",
		colLabel, colReset, colValue, aks.Method, colReset,
		colLabel, colReset, valCol, val, colReset)
}

// --- RateLimitPolicy ---

func (rlp *RateLimitPolicy) String() string {
	if rlp == nil {
		return "<nil RateLimitPolicy>"
	}
	return fmt.Sprintf("Req/Min: %d, Tok/Min: %d, Tok/Day: %d, MaxInstances: %d",
		rlp.MaxRequestsPerMinute, rlp.MaxTokensPerMinute, rlp.MaxTokensPerDay, rlp.MaxConcurrentActiveInstances)
}

func (rlp *RateLimitPolicy) ColourString() string {
	if rlp == nil {
		return colStatusErr + "<nil RateLimitPolicy>" + colReset
	}
	return fmt.Sprintf("%sReq/Min:%s %s%d%s, %sTok/Min:%s %s%d%s, %sTok/Day:%s %s%d%s, %sMaxInstances:%s %s%d%s",
		colLabel, colReset, colCount, rlp.MaxRequestsPerMinute, colReset,
		colLabel, colReset, colCount, rlp.MaxTokensPerMinute, colReset,
		colLabel, colReset, colCount, rlp.MaxTokensPerDay, colReset,
		colLabel, colReset, colCount, rlp.MaxConcurrentActiveInstances, colReset)
}

// --- TokenUsageMetrics ---

func (tum *TokenUsageMetrics) String() string {
	if tum == nil {
		return "<nil TokenUsageMetrics>"
	}
	return fmt.Sprintf("In: %d, Out: %d, Total: %d", tum.InputTokens, tum.OutputTokens, tum.TotalTokens)
}

func (tum *TokenUsageMetrics) ColourString() string {
	if tum == nil {
		return colStatusErr + "<nil TokenUsageMetrics>" + colReset
	}
	return fmt.Sprintf("%sIn:%s %s%d%s, %sOut:%s %s%d%s, %sTotal:%s %s%d%s",
		colLabel, colReset, colCount, tum.InputTokens, colReset,
		colLabel, colReset, colCount, tum.OutputTokens, colReset,
		colLabel, colReset, colCount, tum.TotalTokens, colReset)
}

// --- SupervisorFeedback ---

func (sf *SupervisorFeedback) String() string {
	if sf == nil {
		return "<nil SupervisorFeedback>"
	}
	return fmt.Sprintf("Rating: %.1f, Supervisor: %s, Stamp: %s, Comments: %.30s...",
		sf.Rating, sf.SupervisorAgentID, sf.FeedbackTimestamp.Format(time.RFC3339), sf.Comments)
}

func (sf *SupervisorFeedback) ColourString() string {
	if sf == nil {
		return colStatusErr + "<nil SupervisorFeedback>" + colReset
	}
	return fmt.Sprintf("%sRating:%s %s%.1f%s, %sSupervisor:%s %s%s%s, %sStamp:%s %s%s%s, %sComments:%s %s%.30s...%s",
		colLabel, colReset, colValue, sf.Rating, colReset,
		colLabel, colReset, colID, sf.SupervisorAgentID, colReset,
		colLabel, colReset, colTime, sf.FeedbackTimestamp.Format(time.RFC3339), colReset,
		colLabel, colReset, colValue, sf.Comments, colReset)
}

// --- AIWorkerPerformanceSummary ---

func (ps *AIWorkerPerformanceSummary) String() string {
	if ps == nil {
		return "<nil AIWorkerPerformanceSummary>"
	}
	return fmt.Sprintf("Tasks: %d (S:%d, F:%d), SuccessRate: %.2f%%, AvgDur: %.0fms, Tokens: %d, Cost: $%.4f, ActiveInst(Live): %d, LastActivity: %s",
		ps.TotalTasksAttempted, ps.SuccessfulTasks, ps.FailedTasks, ps.AverageSuccessRate*100,
		ps.AverageDurationMs, ps.TotalTokensProcessed, ps.TotalCostIncurred, ps.ActiveInstancesCount, ps.LastActivityTimestamp.Format(time.RFC3339))
}

func (ps *AIWorkerPerformanceSummary) ColourString() string {
	if ps == nil {
		return colStatusErr + "<nil AIWorkerPerformanceSummary>" + colReset
	}
	// Corrected: This Sprintf was complex and a likely candidate for arg count mismatches if not careful.
	// Breaking it down or ensuring exact counts. Counted 45 specifiers and 45 args.
	return fmt.Sprintf("%sTasks:%s %s%d%s (%sS:%s%s%d%s, %sF:%s%s%d%s), %sSuccessRate:%s %s%.2f%%%s, %sAvgDur:%s %s%.0fms%s, %sTokens:%s %s%d%s, %sCost:%s %s$%.4f%s, %sActiveInst(Live):%s %s%d%s, %sLastActivity:%s %s%s%s",
		colLabel, colReset, colCount, ps.TotalTasksAttempted, colReset, // 5
		colLabel, colReset, colStatusOk, ps.SuccessfulTasks, colReset, // 5
		colLabel, colReset, colStatusErr, ps.FailedTasks, colReset, // 5
		colLabel, colReset, colValue, ps.AverageSuccessRate*100, colReset, // 5 (%.2f%% counts as one float specifier)
		colLabel, colReset, colValue, ps.AverageDurationMs, colReset, // 5
		colLabel, colReset, colCount, ps.TotalTokensProcessed, colReset, // 5
		colLabel, colReset, colValue, ps.TotalCostIncurred, colReset, // 5
		colLabel, colReset, colCount, ps.ActiveInstancesCount, colReset, // 5
		colLabel, colReset, colTime, ps.LastActivityTimestamp.Format(time.RFC3339), colReset) // 5
}

// --- GlobalDataSourceDefinition ---

func (dsd *GlobalDataSourceDefinition) String() string {
	if dsd == nil {
		return "<nil GlobalDataSourceDefinition>"
	}
	path := dsd.LocalPath
	if dsd.Type == DataSourceTypeFileAPI {
		path = dsd.FileAPIPath
	}
	return fmt.Sprintf("DS '%s': Type: %s, Path: '%s', RO: %t, ExtRead: %t, Filters: %v, Rec: %t",
		dsd.Name, dsd.Type, path, dsd.ReadOnly, dsd.AllowExternalReadAccess, dsd.Filters, dsd.Recursive)
}

func (dsd *GlobalDataSourceDefinition) ColourString() string {
	if dsd == nil {
		return colStatusErr + "<nil GlobalDataSourceDefinition>" + colReset
	}
	path := dsd.LocalPath
	if dsd.Type == DataSourceTypeFileAPI {
		path = dsd.FileAPIPath
	}
	return fmt.Sprintf("%sDS%s '%s%s%s': %sType:%s %s%s%s, %sPath:%s '%s%s%s', %sRO:%s %s%t%s, %sExtRead:%s %s%t%s, %sFilters:%s %s%v%s, %sRec:%s %s%t%s",
		colLabel, colReset, colName, dsd.Name, colReset,
		colLabel, colReset, colValue, dsd.Type, colReset,
		colLabel, colReset, colValue, path, colReset,
		colLabel, colReset, colBool, dsd.ReadOnly, colReset,
		colLabel, colReset, colBool, dsd.AllowExternalReadAccess, colReset,
		colLabel, colReset, colValue, dsd.Filters, colReset,
		colLabel, colReset, colBool, dsd.Recursive, colReset)
}

// --- AIWorkerDefinition ---

func (wd *AIWorkerDefinition) String() string {
	if wd == nil {
		return "<nil AIWorkerDefinition>"
	}
	var sb strings.Builder
	// DefinitionID removed as per request
	sb.WriteString(fmt.Sprintf("Name: %s\n", wd.Name))
	sb.WriteString(fmt.Sprintf("  Provider: %s, Model: %s, Status: %s\n", wd.Provider, wd.ModelName, wd.Status))
	sb.WriteString(fmt.Sprintf("  Auth: %s\n", wd.Auth.String())) // Assumes APIKeySource.String() exists

	imModelsStr := "None"
	if len(wd.InteractionModels) > 0 {
		tempModels := make([]string, len(wd.InteractionModels))
		for i, im := range wd.InteractionModels {
			tempModels[i] = string(im)
		}
		imModelsStr = strings.Join(tempModels, ", ")
	}
	sb.WriteString(fmt.Sprintf("  InteractionModels: %s\n", imModelsStr))

	capStr := "None"
	if len(wd.Capabilities) > 0 {
		capStr = strings.Join(wd.Capabilities, ", ")
	}
	sb.WriteString(fmt.Sprintf("  Capabilities: %s\n", capStr))

	sb.WriteString(fmt.Sprintf("  BaseConfig Keys: %d\n", len(wd.BaseConfig)))
	sb.WriteString(fmt.Sprintf("  CostMetrics Keys: %d\n", len(wd.CostMetrics)))
	if wd.RateLimits.MaxRequestsPerMinute > 0 || wd.RateLimits.MaxTokensPerMinute > 0 || wd.RateLimits.MaxTokensPerDay > 0 || wd.RateLimits.MaxConcurrentActiveInstances > 0 {
		sb.WriteString(fmt.Sprintf("  RateLimits: %s\n", wd.RateLimits.String()))
	}

	defFileCtxStr := "None"
	if len(wd.DefaultFileContexts) > 0 {
		defFileCtxStr = strings.Join(wd.DefaultFileContexts, ", ")
	}
	sb.WriteString(fmt.Sprintf("  DefaultFileContexts: %s\n", defFileCtxStr))

	dsRefsStr := "None"
	if len(wd.DataSourceRefs) > 0 {
		dsRefsStr = strings.Join(wd.DataSourceRefs, ", ")
	}
	sb.WriteString(fmt.Sprintf("  DataSourceRefs: %s\n", dsRefsStr))

	toolAllowStr := "None"
	if len(wd.ToolAllowlist) > 0 {
		toolAllowStr = strings.Join(wd.ToolAllowlist, ", ")
	}
	sb.WriteString(fmt.Sprintf("  ToolAllowlist: %s\n", toolAllowStr))

	toolDenyStr := "None"
	if len(wd.ToolDenylist) > 0 {
		toolDenyStr = strings.Join(wd.ToolDenylist, ", ")
	}
	sb.WriteString(fmt.Sprintf("  ToolDenylist: %s\n", toolDenyStr))

	if wd.DefaultSupervisoryAIRef != "" {
		sb.WriteString(fmt.Sprintf("  DefaultSupervisoryAIRef: %s\n", wd.DefaultSupervisoryAIRef))
	}
	if wd.AggregatePerformanceSummary != nil {
		sb.WriteString(fmt.Sprintf("  Performance: %s\n", wd.AggregatePerformanceSummary.String()))
	} else {
		sb.WriteString("  Performance: <nil>\n")
	}
	sb.WriteString(fmt.Sprintf("  Metadata Keys: %d\n", len(wd.Metadata)))
	return sb.String()
}

func (wd *AIWorkerDefinition) ColourString() string {
	if wd == nil {
		return colStatusErr + "<nil AIWorkerDefinition>" + colReset
	}
	var sb strings.Builder
	statusCol := colValue
	switch wd.Status {
	case DefinitionStatusActive:
		statusCol = colStatusOk
	case DefinitionStatusDisabled, DefinitionStatusArchived:
		statusCol = colStatusWarn
	}

	// DefinitionID removed
	sb.WriteString(fmt.Sprintf("%sName:%s %s%s%s\n", colLabel, colReset, colName, wd.Name, colReset))
	sb.WriteString(fmt.Sprintf("  %sProvider:%s %s%s%s, %sModel:%s %s%s%s, %sStatus:%s %s%s%s\n",
		colLabel, colReset, colValue, wd.Provider, colReset,
		colLabel, colReset, colValue, wd.ModelName, colReset,
		colLabel, colReset, statusCol, wd.Status, colReset))
	sb.WriteString(fmt.Sprintf("  %sAuth:%s %s\n", colLabel, colReset, wd.Auth.ColourString())) // Assumes APIKeySource.ColourString()

	imModelsStr := colValue + "None" + colReset
	if len(wd.InteractionModels) > 0 {
		tempModels := make([]string, len(wd.InteractionModels))
		for i, im := range wd.InteractionModels {
			tempModels[i] = string(im)
		}
		imModelsStr = fmt.Sprintf("%s%s%s", colValue, strings.Join(tempModels, ", "), colReset)
	}
	sb.WriteString(fmt.Sprintf("  %sInteractionModels:%s %s\n", colLabel, colReset, imModelsStr))

	capStr := colValue + "None" + colReset
	if len(wd.Capabilities) > 0 {
		capStr = fmt.Sprintf("%s%s%s", colValue, strings.Join(wd.Capabilities, ", "), colReset)
	}
	sb.WriteString(fmt.Sprintf("  %sCapabilities:%s %s\n", colLabel, colReset, capStr))

	sb.WriteString(fmt.Sprintf("  %sBaseConfig Keys:%s %s%d%s\n", colLabel, colReset, colCount, len(wd.BaseConfig), colReset))
	sb.WriteString(fmt.Sprintf("  %sCostMetrics Keys:%s %s%d%s\n", colLabel, colReset, colCount, len(wd.CostMetrics), colReset))

	if wd.RateLimits.MaxRequestsPerMinute > 0 || wd.RateLimits.MaxTokensPerMinute > 0 || wd.RateLimits.MaxTokensPerDay > 0 || wd.RateLimits.MaxConcurrentActiveInstances > 0 {
		sb.WriteString(fmt.Sprintf("  %sRateLimits:%s %s\n", colLabel, colReset, wd.RateLimits.ColourString())) // Assumes RateLimitPolicy.ColourString()
	}

	defFileCtxStr := colValue + "None" + colReset
	if len(wd.DefaultFileContexts) > 0 {
		defFileCtxStr = fmt.Sprintf("%s%s%s", colValue, strings.Join(wd.DefaultFileContexts, ", "), colReset)
	}
	sb.WriteString(fmt.Sprintf("  %sDefaultFileContexts:%s %s\n", colLabel, colReset, defFileCtxStr))

	dsRefsStr := colValue + "None" + colReset
	if len(wd.DataSourceRefs) > 0 {
		dsRefsStr = fmt.Sprintf("%s%s%s", colValue, strings.Join(wd.DataSourceRefs, ", "), colReset)
	}
	sb.WriteString(fmt.Sprintf("  %sDataSourceRefs:%s %s\n", colLabel, colReset, dsRefsStr))

	toolAllowStr := colValue + "None" + colReset
	if len(wd.ToolAllowlist) > 0 {
		toolAllowStr = fmt.Sprintf("%s%s%s", colValue, strings.Join(wd.ToolAllowlist, ", "), colReset)
	}
	sb.WriteString(fmt.Sprintf("  %sToolAllowlist:%s %s\n", colLabel, colReset, toolAllowStr))

	toolDenyStr := colValue + "None" + colReset
	if len(wd.ToolDenylist) > 0 {
		toolDenyStr = fmt.Sprintf("%s%s%s", colValue, strings.Join(wd.ToolDenylist, ", "), colReset)
	}
	sb.WriteString(fmt.Sprintf("  %sToolDenylist:%s %s\n", colLabel, colReset, toolDenyStr))

	if wd.DefaultSupervisoryAIRef != "" {
		sb.WriteString(fmt.Sprintf("  %sDefaultSupervisoryAIRef:%s %s%s%s\n", colLabel, colReset, colID, wd.DefaultSupervisoryAIRef, colReset))
	}

	if wd.AggregatePerformanceSummary != nil {
		sb.WriteString(fmt.Sprintf("  %sPerformance:%s %s\n", colLabel, colReset, wd.AggregatePerformanceSummary.ColourString())) // Assumes AIWorkerPerformanceSummary.ColourString()
	} else {
		sb.WriteString(fmt.Sprintf("  %sPerformance:%s %s<nil>%s\n", colLabel, colReset, colStatusWarn, colReset))
	}
	sb.WriteString(fmt.Sprintf("  %sMetadata Keys:%s %s%d%s\n", colLabel, colReset, colCount, len(wd.Metadata), colReset))
	return sb.String()
}

// Total args for AIWorkerDefinition.ColourString: Sum of args for each Sprintf call. No single call is overly complex to cause the "needs 18 has 20" error.
// The original error likely pointed to a specific call that I've now implicitly corrected by re-verifying each.

// --- AIWorkerInstance ---

func (wi *AIWorkerInstance) String() string {
	if wi == nil {
		return "<nil AIWorkerInstance>"
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Instance ID: %s (Def: %s)\n", wi.InstanceID, wi.DefinitionID))
	sb.WriteString(fmt.Sprintf("  Status: %s\n", wi.Status))
	if wi.PoolID != "" {
		sb.WriteString(fmt.Sprintf("  PoolID: %s\n", wi.PoolID))
	}
	if wi.CurrentTaskID != "" {
		sb.WriteString(fmt.Sprintf("  TaskID: %s\n", wi.CurrentTaskID))
	}
	sb.WriteString(fmt.Sprintf("  Tokens: %s\n", wi.SessionTokenUsage.String()))
	sb.WriteString(fmt.Sprintf("  History Turns: %d\n", len(wi.ConversationHistory)))
	sb.WriteString(fmt.Sprintf("  Last Activity: %s\n", wi.LastActivityTimestamp.Format(time.RFC3339)))
	if wi.LastError != "" {
		sb.WriteString(fmt.Sprintf("  Error: %s\n", wi.LastError))
	}
	return sb.String()
}

func (wi *AIWorkerInstance) ColourString() string {
	if wi == nil {
		return colStatusErr + "<nil AIWorkerInstance>" + colReset
	}
	var sb strings.Builder
	statusCol := colValue
	switch wi.Status {
	case InstanceStatusIdle, InstanceStatusRetiredCompleted:
		statusCol = colStatusOk
	case InstanceStatusError, InstanceStatusRetiredError, InstanceStatusRetiredExhausted, InstanceStatusContextFull, InstanceStatusRateLimited, InstanceStatusTokenLimitReached:
		statusCol = colStatusErr
	case InstanceStatusBusy, InstanceStatusInitializing:
		statusCol = colStatusWarn
	}

	sb.WriteString(fmt.Sprintf("%sInstance ID:%s %s%s%s (%sDef:%s %s%s%s)\n", colLabel, colReset, colID, wi.InstanceID, colReset, colLabel, colReset, colID, wi.DefinitionID, colReset))
	sb.WriteString(fmt.Sprintf("  %sStatus:%s %s%s%s\n", colLabel, colReset, statusCol, wi.Status, colReset))
	if wi.PoolID != "" {
		sb.WriteString(fmt.Sprintf("  %sPoolID:%s %s%s%s\n", colLabel, colReset, colID, wi.PoolID, colReset))
	}
	if wi.CurrentTaskID != "" {
		sb.WriteString(fmt.Sprintf("  %sTaskID:%s %s%s%s\n", colLabel, colReset, colID, wi.CurrentTaskID, colReset))
	}
	sb.WriteString(fmt.Sprintf("  %sTokens:%s %s\n", colLabel, colReset, wi.SessionTokenUsage.ColourString()))
	sb.WriteString(fmt.Sprintf("  %sHistory Turns:%s %s%d%s\n", colLabel, colReset, colCount, len(wi.ConversationHistory), colReset))
	sb.WriteString(fmt.Sprintf("  %sLast Activity:%s %s%s%s\n", colLabel, colReset, colTime, wi.LastActivityTimestamp.Format(time.RFC3339), colReset))
	if wi.LastError != "" {
		sb.WriteString(fmt.Sprintf("  %sError:%s %s%s%s\n", colLabel, colReset, colStatusErr, wi.LastError, colReset))
	}
	return sb.String()
}

// --- PerformanceRecord ---

func (pr *PerformanceRecord) String() string {
	if pr == nil {
		return "<nil PerformanceRecord>"
	}
	status := "FAIL"
	if pr.Success {
		status = "OK"
	}
	return fmt.Sprintf("Task: %s (Inst: %s, Def: %s) Status: %s, Dur: %dms, Cost: $%.4f, Error: %.30s...",
		pr.TaskID, pr.InstanceID, pr.DefinitionID, status, pr.DurationMs, pr.CostIncurred, pr.ErrorDetails)
}

func (pr *PerformanceRecord) ColourString() string {
	if pr == nil {
		return colStatusErr + "<nil PerformanceRecord>" + colReset
	}
	statusStr := "FAIL"
	statusCol := colStatusErr
	if pr.Success {
		statusStr = "OK"
		statusCol = colStatusOk
	}
	return fmt.Sprintf("%sTask:%s %s%s%s (%sInst:%s %s%s%s, %sDef:%s %s%s%s) %sStatus:%s %s%s%s, %sDur:%s %s%dms%s, %sCost:%s %s$%.4f%s, %sError:%s %s%.30s...%s",
		colLabel, colReset, colID, pr.TaskID, colReset,
		colLabel, colReset, colID, pr.InstanceID, colReset,
		colLabel, colReset, colID, pr.DefinitionID, colReset,
		colLabel, colReset, statusCol, statusStr, colReset,
		colLabel, colReset, colValue, pr.DurationMs, colReset,
		colLabel, colReset, colValue, pr.CostIncurred, colReset,
		colLabel, colReset, colStatusErr, pr.ErrorDetails, colReset)
}

// --- RetiredInstanceInfo ---

func (rii *RetiredInstanceInfo) String() string {
	if rii == nil {
		return "<nil RetiredInstanceInfo>"
	}
	return fmt.Sprintf("Retired Inst: %s (Def: %s), Status: %s, Reason: %s, Tokens: %s, PerfRecs: %d",
		rii.InstanceID, rii.DefinitionID, rii.FinalStatus, rii.RetirementReason, rii.SessionTokenUsage.String(), len(rii.PerformanceRecords))
}

func (rii *RetiredInstanceInfo) ColourString() string {
	if rii == nil {
		return colStatusErr + "<nil RetiredInstanceInfo>" + colReset
	}
	statusCol := colValue
	switch rii.FinalStatus {
	case InstanceStatusRetiredCompleted:
		statusCol = colStatusOk
	case InstanceStatusRetiredError, InstanceStatusRetiredExhausted:
		statusCol = colStatusErr
	default:
		statusCol = colStatusWarn
	}
	return fmt.Sprintf("%sRetired Inst:%s %s%s%s (%sDef:%s %s%s%s), %sStatus:%s %s%s%s, %sReason:%s %s%s%s, %sTokens:%s %s, %sPerfRecs:%s %s%d%s",
		colLabel, colReset, colID, rii.InstanceID, colReset,
		colLabel, colReset, colID, rii.DefinitionID, colReset,
		colLabel, colReset, statusCol, rii.FinalStatus, colReset,
		colLabel, colReset, colValue, rii.RetirementReason, colReset,
		colLabel, colReset, rii.SessionTokenUsage.ColourString(),
		colLabel, colReset, colCount, len(rii.PerformanceRecords), colReset)
}

// --- InstanceRetirementPolicy ---

func (irp *InstanceRetirementPolicy) String() string {
	if irp == nil {
		return "<nil InstanceRetirementPolicy>"
	}
	return fmt.Sprintf("MaxTasks: %d, MaxAgeHours: %d", irp.MaxTasksPerInstance, irp.MaxInstanceAgeHours)
}

func (irp *InstanceRetirementPolicy) ColourString() string {
	if irp == nil {
		return colStatusErr + "<nil InstanceRetirementPolicy>" + colReset
	}
	return fmt.Sprintf("%sMaxTasks:%s %s%d%s, %sMaxAgeHours:%s %s%d%s",
		colLabel, colReset, colCount, irp.MaxTasksPerInstance, colReset,
		colLabel, colReset, colCount, irp.MaxInstanceAgeHours, colReset)
}

// --- AIWorkerPoolDefinition ---

func (wpd *AIWorkerPoolDefinition) String() string {
	if wpd == nil {
		return "<nil AIWorkerPoolDefinition>"
	}
	return fmt.Sprintf("Pool '%s' (ID: %s): TargetDef: %s, Idle: %d, Max: %d, Retirement: [%s], DS: %d",
		wpd.Name, wpd.PoolID, wpd.TargetAIWorkerDefinitionName, wpd.MinIdleInstances, wpd.MaxTotalInstances,
		wpd.InstanceRetirementPolicy.String(), len(wpd.DataSourceRefs))
}

func (wpd *AIWorkerPoolDefinition) ColourString() string {
	if wpd == nil {
		return colStatusErr + "<nil AIWorkerPoolDefinition>" + colReset
	}
	return fmt.Sprintf("%sPool%s '%s%s%s' (%sID:%s %s%s%s): %sTargetDef:%s %s%s%s, %sIdle:%s %s%d%s, %sMax:%s %s%d%s, %sRetirement:%s [%s], %sDS:%s %s%d%s",
		colLabel, colReset, colName, wpd.Name, colReset,
		colLabel, colReset, colID, wpd.PoolID, colReset,
		colLabel, colReset, colName, wpd.TargetAIWorkerDefinitionName, colReset,
		colLabel, colReset, colCount, wpd.MinIdleInstances, colReset,
		colLabel, colReset, colCount, wpd.MaxTotalInstances, colReset,
		colLabel, colReset, wpd.InstanceRetirementPolicy.ColourString(),
		colLabel, colReset, colCount, len(wpd.DataSourceRefs), colReset)
}

// --- RetryPolicy ---

func (rp *RetryPolicy) String() string {
	if rp == nil {
		return "<nil RetryPolicy>"
	}
	return fmt.Sprintf("MaxRetries: %d, Delay: %ds", rp.MaxRetries, rp.RetryDelaySeconds)
}

func (rp *RetryPolicy) ColourString() string {
	if rp == nil {
		return colStatusErr + "<nil RetryPolicy>" + colReset
	}
	return fmt.Sprintf("%sMaxRetries:%s %s%d%s, %sDelay:%s %s%ds%s",
		colLabel, colReset, colCount, rp.MaxRetries, colReset,
		colLabel, colReset, colCount, rp.RetryDelaySeconds, colReset)
}

// --- WorkQueueDefinition ---

func (wqd *WorkQueueDefinition) String() string {
	if wqd == nil {
		return "<nil WorkQueueDefinition>"
	}
	return fmt.Sprintf("Queue '%s' (ID: %s): Pools: %v, Prio: %d, Retry: [%s], Persist: %t, DS: %d",
		wqd.Name, wqd.QueueID, wqd.AssociatedPoolNames, wqd.DefaultPriority, wqd.RetryPolicy.String(),
		wqd.PersistTasks, len(wqd.DataSourceRefs))
}

func (wqd *WorkQueueDefinition) ColourString() string {
	if wqd == nil {
		return colStatusErr + "<nil WorkQueueDefinition>" + colReset
	}
	return fmt.Sprintf("%sQueue%s '%s%s%s' (%sID:%s %s%s%s): %sPools:%s %s%v%s, %sPrio:%s %s%d%s, %sRetry:%s [%s], %sPersist:%s %s%t%s, %sDS:%s %s%d%s",
		colLabel, colReset, colName, wqd.Name, colReset,
		colLabel, colReset, colID, wqd.QueueID, colReset,
		colLabel, colReset, colValue, wqd.AssociatedPoolNames, colReset,
		colLabel, colReset, colValue, wqd.DefaultPriority, colReset,
		colLabel, colReset, wqd.RetryPolicy.ColourString(),
		colLabel, colReset, colBool, wqd.PersistTasks, colReset,
		colLabel, colReset, colCount, len(wqd.DataSourceRefs), colReset)
}

// --- WorkItemDefinition ---

func (wid *WorkItemDefinition) String() string {
	if wid == nil {
		return "<nil WorkItemDefinition>"
	}
	return fmt.Sprintf("WorkItemDef '%s' (ID: %s): Desc: %.30s..., CriteriaKeys: %d, SchemaKeys: %d, DS: %d",
		wid.Name, wid.WorkItemDefinitionID, wid.Description, len(wid.DefaultTargetWorkerCriteria),
		len(wid.DefaultPayloadSchema), len(wid.DefaultDataSourceRefs))
}

func (wid *WorkItemDefinition) ColourString() string {
	if wid == nil {
		return colStatusErr + "<nil WorkItemDefinition>" + colReset
	}
	return fmt.Sprintf("%sWorkItemDef%s '%s%s%s' (%sID:%s %s%s%s): %sDesc:%s %s%.30s...%s, %sCriteriaKeys:%s %s%d%s, %sSchemaKeys:%s %s%d%s, %sDS:%s %s%d%s",
		colLabel, colReset, colName, wid.Name, colReset,
		colLabel, colReset, colID, wid.WorkItemDefinitionID, colReset,
		colLabel, colReset, colValue, wid.Description, colReset,
		colLabel, colReset, colCount, len(wid.DefaultTargetWorkerCriteria), colReset,
		colLabel, colReset, colCount, len(wid.DefaultPayloadSchema), colReset,
		colLabel, colReset, colCount, len(wid.DefaultDataSourceRefs), colReset)
}

// --- WorkItem ---

func (wi *WorkItem) String() string {
	if wi == nil {
		return "<nil WorkItem>"
	}
	return fmt.Sprintf("WorkItem TaskID: %s (DefName: %s, Queue: %s)\n  Status: %s, Prio: %d, Retries: %d\n  PayloadKeys: %d, DS: %d\n  Error: %.30s...",
		wi.TaskID, wi.WorkItemDefinitionName, wi.QueueName, wi.Status, wi.Priority, wi.RetryCount,
		len(wi.Payload), len(wi.DataSourceRefs), wi.Error)
}

func (wi *WorkItem) ColourString() string {
	if wi == nil {
		return colStatusErr + "<nil WorkItem>" + colReset
	}
	statusCol := colValue
	switch wi.Status {
	case WorkItemStatusCompleted:
		statusCol = colStatusOk
	case WorkItemStatusFailed, WorkItemStatusCancelled:
		statusCol = colStatusErr
	case WorkItemStatusProcessing, WorkItemStatusRetrying:
		statusCol = colStatusWarn
	case WorkItemStatusPending:
		statusCol = colStatusNeut
	}

	return fmt.Sprintf("%sWorkItem TaskID:%s %s%s%s (%sDefName:%s %s%s%s, %sQueue:%s %s%s%s)\n  %sStatus:%s %s%s%s, %sPrio:%s %s%d%s, %sRetries:%s %s%d%s\n  %sPayloadKeys:%s %s%d%s, %sDS:%s %s%d%s\n  %sError:%s %s%.30s...%s",
		colLabel, colReset, colID, wi.TaskID, colReset,
		colLabel, colReset, colName, wi.WorkItemDefinitionName, colReset,
		colLabel, colReset, colName, wi.QueueName, colReset,
		colLabel, colReset, statusCol, wi.Status, colReset,
		colLabel, colReset, colValue, wi.Priority, colReset,
		colLabel, colReset, colCount, wi.RetryCount, colReset,
		colLabel, colReset, colCount, len(wi.Payload), colReset,
		colLabel, colReset, colCount, len(wi.DataSourceRefs), colReset,
		colLabel, colReset, colStatusErr, wi.Error, colReset)
}

// --- LLMCallMetrics ---

func (lcm *LLMCallMetrics) String() string {
	if lcm == nil {
		return "<nil LLMCallMetrics>"
	}
	return fmt.Sprintf("Model: %s, In: %d, Out: %d, Total: %d, Reason: %s",
		lcm.ModelUsed, lcm.InputTokens, lcm.OutputTokens, lcm.TotalTokens, lcm.FinishReason)
}

func (lcm *LLMCallMetrics) ColourString() string {
	if lcm == nil {
		return colStatusErr + "<nil LLMCallMetrics>" + colReset
	}
	return fmt.Sprintf("%sModel:%s %s%s%s, %sIn:%s %s%d%s, %sOut:%s %s%d%s, %sTotal:%s %s%d%s, %sReason:%s %s%s%s",
		colLabel, colReset, colValue, lcm.ModelUsed, colReset,
		colLabel, colReset, colCount, lcm.InputTokens, colReset,
		colLabel, colReset, colCount, lcm.OutputTokens, colReset,
		colLabel, colReset, colCount, lcm.TotalTokens, colReset,
		colLabel, colReset, colValue, lcm.FinishReason, colReset)
}

// --- AIWorkerManager ---

func (m *AIWorkerManager) String() string {
	if m == nil {
		return "<nil AIWorkerManager>"
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString("=== AIWorkerManager Status ===\n")
	sb.WriteString(fmt.Sprintf("Sandbox Directory: %s\n", m.sandboxDir))
	sb.WriteString(fmt.Sprintf("Definitions File: %s\n", m.FullPathForDefinitions()))
	sb.WriteString(fmt.Sprintf("Performance File: %s\n", m.FullPathForPerformanceData()))
	sb.WriteString(fmt.Sprintf("Total Worker Definitions: %d\n", len(m.definitions)))
	sb.WriteString(fmt.Sprintf("Total Active Instances: %d\n", len(m.activeInstances)))
	sb.WriteString(fmt.Sprintf("Total Rate Trackers: %d\n", len(m.rateTrackers)))

	if len(m.definitions) > 0 {
		sb.WriteString("\n--- Worker Definitions (Summary) ---\n")
		defs := make([]*AIWorkerDefinition, 0, len(m.definitions))
		for _, def := range m.definitions {
			if def != nil {
				defs = append(defs, def)
			}
		}
		sort.Slice(defs, func(i, j int) bool { return defs[i].Name < defs[j].Name })
		for i, def := range defs {
			activeInstances := 0
			if def.AggregatePerformanceSummary != nil {
				activeInstances = def.AggregatePerformanceSummary.ActiveInstancesCount
			}
			sb.WriteString(fmt.Sprintf("  [%d] '%s' (ID: %s, Provider: %s, Model: %s, Status: %s, ActiveInst: %d)\n",
				i+1, def.Name, def.DefinitionID, def.Provider, def.ModelName, def.Status, activeInstances))
		}
	}

	if len(m.activeInstances) > 0 {
		sb.WriteString("\n--- Active Instances (Summary) ---\n")
		instanceIDs := make([]string, 0, len(m.activeInstances))
		for id := range m.activeInstances {
			instanceIDs = append(instanceIDs, id)
		}
		sort.Strings(instanceIDs)
		for i, id := range instanceIDs {
			instance := m.activeInstances[id]
			if instance == nil {
				continue
			}
			sb.WriteString(fmt.Sprintf("  [%d] ID: %s (Def: %s, Status: %s, Task: %s, Tokens: %d)\n",
				i+1, id, instance.DefinitionID, instance.Status, instance.CurrentTaskID, instance.SessionTokenUsage.TotalTokens))
		}
	}
	sb.WriteString("==============================\n")
	return sb.String()
}

func (m *AIWorkerManager) ColourString() string {
	if m == nil {
		return colStatusErr + "<nil AIWorkerManager>" + colReset // Ensure colStatusErr and colReset are defined
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s[::b]=== AIWorkerManager Detail ===%s\n", colLabel, colReset))
	sb.WriteString(fmt.Sprintf("%sSandbox:%s %s%s%s\n", colLabel, colReset, colValue, m.sandboxDir, colReset))
	sb.WriteString(fmt.Sprintf("%sDefs:%s %s%s%s\n", colLabel, colReset, colValue, m.FullPathForDefinitions(), colReset))
	sb.WriteString(fmt.Sprintf("%sPerf:%s %s%s%s\n", colLabel, colReset, colValue, m.FullPathForPerformanceData(), colReset))
	sb.WriteString(fmt.Sprintf("%sTotal Worker Definitions:%s %s%d%s\n", colLabel, colReset, colCount, len(m.definitions), colReset))
	sb.WriteString(fmt.Sprintf("%sTotal Active Instances:%s %s%d%s\n", colLabel, colReset, colCount, len(m.activeInstances), colReset))
	sb.WriteString(fmt.Sprintf("%sTotal Rate Trackers:%s %s%d%s\n", colLabel, colReset, colCount, len(m.rateTrackers), colReset))

	if len(m.definitions) > 0 {
		sb.WriteString(fmt.Sprintf("\n%s[::b]--- Worker Definitions ---%s\n", colLabel, colReset)) // Changed from "Summary"
		defs := make([]*AIWorkerDefinition, 0, len(m.definitions))
		for _, def := range m.definitions {
			if def != nil {
				defs = append(defs, def)
			}
		}
		sort.Slice(defs, func(i, j int) bool { return defs[i].Name < defs[j].Name })

		for i, def := range defs {
			// Prepend an index for clarity in the list
			sb.WriteString(fmt.Sprintf("%s[%d]%s ", colCount, i+1, colReset))
			// Call the AIWorkerDefinition's ColourString() method directly.
			// This method is expected to be multi-line and already colorized, starting with "Name: ...".
			sb.WriteString(def.ColourString())
			// AIWorkerDefinition.ColourString() should end its output with a newline.
			// Add an additional newline for spacing between definitions if desired.
			if i < len(defs)-1 {
				sb.WriteString("\n") // Extra spacing between definition blocks
			}
		}
	}

	if len(m.activeInstances) > 0 {
		sb.WriteString(fmt.Sprintf("\n%s[::b]--- Active Instances (Summary) ---%s\n", colLabel, colReset))
		instanceIDs := make([]string, 0, len(m.activeInstances))
		for id := range m.activeInstances {
			instanceIDs = append(instanceIDs, id)
		}
		sort.Strings(instanceIDs)
		for i, id := range instanceIDs {
			instance := m.activeInstances[id]
			if instance == nil {
				continue
			}
			statusColInst := colValue // Default color
			switch instance.Status {
			case InstanceStatusIdle, InstanceStatusRetiredCompleted:
				statusColInst = colStatusOk
			case InstanceStatusError, InstanceStatusRetiredError, InstanceStatusRetiredExhausted:
				statusColInst = colStatusErr
			case InstanceStatusBusy, InstanceStatusInitializing:
				statusColInst = colStatusWarn
			default:
				// Potentially use colStatusNeut or keep colValue if other statuses are neutral
				statusColInst = colStatusNeut
			}
			// This Sprintf for active instances remains a summary line as per the original structure.
			sb.WriteString(fmt.Sprintf("  %s[%d]%s %sID:%s %s%s%s (%sDef:%s %s%s%s, %sStatus:%s %s%s%s, %sTask:%s %s%s%s, %sTokens:%s %s%d%s)\n",
				colCount, i+1, colReset,
				colLabel, colReset, colID, id, colReset,
				colLabel, colReset, colID, instance.DefinitionID, colReset,
				colLabel, colReset, statusColInst, instance.Status, colReset,
				colLabel, colReset, colID, instance.CurrentTaskID, colReset,
				colLabel, colReset, colCount, instance.SessionTokenUsage.TotalTokens, colReset))
		}
	}
	sb.WriteString(fmt.Sprintf("%s[::b]==============================%s\n", colLabel, colReset))
	return sb.String()
}

// --- WorkerRateTracker ---
// Assuming WorkerRateTracker struct fields based on ai_wm.go's initializeRateTrackersUnsafe
// This should ideally be in core/ai_wm_ratelimit.go if that's where WorkerRateTracker is defined.

func (rt *WorkerRateTracker) String() string {
	if rt == nil {
		return "<nil WorkerRateTracker>"
	}
	// rt.mu.Lock() // Uncomment if your WorkerRateTracker has a mutex and needs locking
	// defer rt.mu.Unlock()
	return fmt.Sprintf("Tracker (DefID: %s): ActiveInst: %d, Req/Min: %d, Tok/Min: %d, Tok/Day: %d, ReqMarker: %s",
		rt.DefinitionID, rt.CurrentActiveInstances, rt.RequestsLastMinute,
		rt.TokensLastMinute, rt.TokensToday, rt.RequestsMinuteMarker.Format(time.Kitchen))
}

func (rt *WorkerRateTracker) ColourString() string {
	if rt == nil {
		return colStatusErr + "<nil WorkerRateTracker>" + colReset
	}
	// rt.mu.Lock() // Uncomment if your WorkerRateTracker has a mutex and needs locking
	// defer rt.mu.Unlock()
	// Corrected Sprintf: ensuring each %d gets an int, and counts match
	return fmt.Sprintf("%sTracker (DefID:%s %s%s%s): %sActiveInst:%s %s%d%s, %sReq/Min:%s %s%d%s, %sTok/Min:%s %s%d%s, %sTok/Day:%s %s%d%s, %sReqMarker:%s %s%s%s",
		colLabel, colReset, colID, rt.DefinitionID, colReset, // 5 args for %s...(DefID:%s %s%s%s):
		colLabel, colReset, colCount, rt.CurrentActiveInstances, colReset, // 5 args for %sActiveInst:%s %s%d%s,
		colLabel, colReset, colCount, rt.RequestsLastMinute, colReset, // 5 args for %sReq/Min:%s %s%d%s,
		colLabel, colReset, colCount, rt.TokensLastMinute, colReset, // 5 args for %sTok/Min:%s %s%d%s,
		colLabel, colReset, colCount, rt.TokensToday, colReset, // 5 args for %sTok/Day:%s %s%d%s,
		colLabel, colReset, colTime, rt.RequestsMinuteMarker.Format(time.Kitchen), colReset) // 5 args for %sReqMarker:%s %s%s%s
}
