// NeuroScript Version: 0.3.0
// File version: 3
// Purpose: Added a String() method to the Capability struct for human-readable output.
// filename: pkg/capability/capability.go
// nlines: 85
// risk_rating: LOW

// Package capability defines the minimal data structures for expressing
// capabilities, limits and run-time counters, along with a GrantSet container.
package capability

import "strings"

// Capability expresses a unit of authority.
// Resource examples: "env","secrets","net","fs","model","sandbox","proc","clock","rand","budget".
// Verbs examples: "read","write","use","admin","exec".
// Scopes are resource-specific strings (env keys, hostnames, paths, model names, etc.).
type Capability struct {
	Resource string
	Verbs    []string
	Scopes   []string
}

// String returns a human-readable representation of the capability,
// following the 'resource:verb,verb:scope,scope' format.
func (c Capability) String() string {
	var sb strings.Builder
	sb.WriteString(c.Resource)

	if len(c.Verbs) > 0 {
		sb.WriteString(":")
		sb.WriteString(strings.Join(c.Verbs, ","))
	}

	if len(c.Scopes) > 0 {
		sb.WriteString(":")
		sb.WriteString(strings.Join(c.Scopes, ","))
	}

	return sb.String()
}

// Limits encode quantitative guardrails that apply over a run.
type Limits struct {
	BudgetPerRunCents   map[string]int // currency -> max cents (CAD, USD, etc.)
	BudgetPerCallCents  map[string]int
	NetMaxBytes         int64
	NetMaxCalls         int
	FSMaxBytes          int64
	FSMaxCalls          int
	ToolMaxCalls        map[string]int // tool name -> limit
	TimeMaxSleepSeconds int
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
