// NeuroScript Version: 0.3.0
// File version: 1
// Purpose: Core capability, limits, counters, and grant set types used by the policy gate.
// filename: pkg/policy/capability/capability.go
// nlines: 66
// risk_rating: MEDIUM

// Package capability defines the minimal data structures for expressing
// capabilities, limits and run-time counters, along with a GrantSet container.
package capability

// Capability expresses a unit of authority.
// Resource examples: "env","secrets","net","fs","model","sandbox","proc","clock","rand","budget".
// Verbs examples: "read","write","use","admin","exec".
// Scopes are resource-specific strings (env keys, hostnames, paths, model names, etc.).
type Capability struct {
	Resource string
	Verbs    []string
	Scopes   []string
}

// Limits encode quantitative guardrails that apply over a run.
type Limits struct {
	BudgetPerRunCents  map[string]int // currency -> max cents (CAD, USD, etc.)
	BudgetPerCallCents map[string]int
	NetMaxBytes        int64
	NetMaxCalls        int
	FSMaxBytes         int64
	FSMaxCalls         int
	ToolMaxCalls       map[string]int // tool name -> limit
}

// Counters record consumption during a run and are compared to Limits.
type Counters struct {
	BudgetSpentCents map[string]int
	NetBytes         int64
	NetCalls         int
	FSBytes          int64
	FSCalls          int
	ToolCalls        map[string]int
}

// GrantSet aggregates grants, limits and live counters for a run.
type GrantSet struct {
	Grants   []Capability
	Limits   Limits
	Counters *Counters
}

// NewCounters constructs zeroed counters with the necessary maps allocated.
func NewCounters() *Counters {
	return &Counters{
		BudgetSpentCents: map[string]int{},
		ToolCalls:        map[string]int{},
	}
}

// NewGrantSet creates a GrantSet with provided grants and limits.
func NewGrantSet(grants []Capability, limits Limits) GrantSet {
	return GrantSet{
		Grants:   grants,
		Limits:   limits,
		Counters: NewCounters(),
	}
}
