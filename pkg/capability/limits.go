// NeuroScript Version: 0.3.0
// File version: 2
// Purpose: Limit and counter enforcement helpers. Added CheckSleep to enforce time limits.
// filename: pkg/policy/capability/limits.go
// nlines: 107
// risk_rating: MEDIUM

package capability

import (
	"errors"
	"fmt"
)

var (
	// ErrBudgetExceeded indicates a per-run or per-call budget violation.
	ErrBudgetExceeded = errors.New("budget exceeded")
	// ErrNetExceeded indicates a network bytes/calls limit violation.
	ErrNetExceeded = errors.New("network limits exceeded")
	// ErrFSExceeded indicates a filesystem bytes/calls limit violation.
	ErrFSExceeded = errors.New("filesystem limits exceeded")
	// ErrToolExceeded indicates a per-tool call count limit violation.
	ErrToolExceeded = errors.New("tool call limit exceeded")
	// ErrTimeExceeded indicates a time-based limit violation (e.g., sleep duration).
	ErrTimeExceeded = errors.New("time limit exceeded")
)

// CheckSleep validates a single sleep duration against the limit.
func (g *GrantSet) CheckSleep(seconds float64) error {
	if g.Limits.TimeMaxSleepSeconds > 0 && int(seconds) > g.Limits.TimeMaxSleepSeconds {
		return fmt.Errorf("%w: sleep duration of %.2fs exceeds limit of %ds",
			ErrTimeExceeded, seconds, g.Limits.TimeMaxSleepSeconds)
	}
	return nil
}

// CheckPerCallBudget validates a single-call spend against the per-call limit.
func (g *GrantSet) CheckPerCallBudget(currency string, cents int) error {
	per := g.Limits.BudgetPerCallCents[currency]
	if per > 0 && cents > per {
		return ErrBudgetExceeded
	}
	return nil
}

// ChargeBudget increments accumulated spend and enforces per-run budget.
func (g *GrantSet) ChargeBudget(currency string, cents int) error {
	if g.Counters == nil {
		g.Counters = NewCounters()
	}
	max := g.Limits.BudgetPerRunCents[currency]
	cur := g.Counters.BudgetSpentCents[currency]
	if max > 0 && cur+cents > max {
		return ErrBudgetExceeded
	}
	g.Counters.BudgetSpentCents[currency] = cur + cents
	return nil
}

// CountNet accounts for one network operation of given size.
func (g *GrantSet) CountNet(bytes int64) error {
	if g.Counters == nil {
		g.Counters = NewCounters()
	}
	if g.Limits.NetMaxCalls > 0 && g.Counters.NetCalls+1 > g.Limits.NetMaxCalls {
		return ErrNetExceeded
	}
	if g.Limits.NetMaxBytes > 0 && g.Counters.NetBytes+bytes > g.Limits.NetMaxBytes {
		return ErrNetExceeded
	}
	g.Counters.NetCalls++
	g.Counters.NetBytes += bytes
	return nil
}

// CountFS accounts for one filesystem operation of given size.
func (g *GrantSet) CountFS(bytes int64) error {
	if g.Counters == nil {
		g.Counters = NewCounters()
	}
	if g.Limits.FSMaxCalls > 0 && g.Counters.FSCalls+1 > g.Limits.FSMaxCalls {
		return ErrFSExceeded
	}
	if g.Limits.FSMaxBytes > 0 && g.Counters.FSBytes+bytes > g.Limits.FSMaxBytes {
		return ErrFSExceeded
	}
	g.Counters.FSCalls++
	g.Counters.FSBytes += bytes
	return nil
}

// CountToolCall increments the per-tool call counter and enforces its limit.
func (g *GrantSet) CountToolCall(tool string) error {
	if g.Counters == nil {
		g.Counters = NewCounters()
	}
	if g.Limits.ToolMaxCalls == nil {
		return nil
	}
	max, ok := g.Limits.ToolMaxCalls[tool]
	if ok && max > 0 {
		cur := g.Counters.ToolCalls[tool]
		if cur+1 > max {
			return ErrToolExceeded
		}
	}
	g.Counters.ToolCalls[tool]++
	return nil
}
