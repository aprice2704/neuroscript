// NeuroScript Version: 0.3.0
// File version: 1
// Purpose: Execution policy gate (trust, allow/deny, capabilities, per-tool limits) applied before any tool call.
// filename: pkg/runtime/policy.go
// nlines: 173
// risk_rating: HIGH

package runtime

import (
	"errors"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/policy/capability"
)

// ExecContext represents the interpreter's trust context for the current run.
type ExecContext string

const (
	ContextConfig ExecContext = "config"
	ContextNormal ExecContext = "normal"
	ContextTest   ExecContext = "test"
)

// ExecPolicy contains allow/deny lists, capability grants, and counters/limits.
type ExecPolicy struct {
	Context ExecContext
	Allow   []string            // tool name patterns allowed (glob-like)
	Deny    []string            // tool name patterns denied (deny wins)
	Grants  capability.GrantSet // capability grants + limits/counters
}

// ToolMeta describes a tool for policy evaluation. Effects support linting/caching.
type ToolMeta struct {
	Name          string
	RequiresTrust bool
	RequiredCaps  []capability.Capability
	Effects       []string // "idempotent","readsNet","readsFS","readsClock","readsRand"
}

var (
	// ErrTrust signals a trusted-only tool attempted in a non-config context.
	ErrTrust = errors.New("tool requires trusted context")
	// ErrPolicy signals a tool blocked by allow/deny policy.
	ErrPolicy = errors.New("tool not allowed by policy")
	// ErrCapability signals missing capability grants for the call.
	ErrCapability = errors.New("capabilities not granted")
)

// CanCall enforces trust context, allow/deny, capability grants, and per-tool limits.
func (p *ExecPolicy) CanCall(t ToolMeta) error {
	if t.RequiresTrust && p.Context != ContextConfig {
		return ErrTrust
	}
	if disallowed(t.Name, p.Allow, p.Deny) {
		return ErrPolicy
	}
	if !capability.CapsSatisfied(t.RequiredCaps, p.Grants.Grants) {
		return ErrCapability
	}
	// Per-tool limit (optional, zero means unlimited)
	return p.Grants.CountToolCall(t.Name)
}

// disallowed decides policy outcome for a tool name given allow/deny lists.
func disallowed(name string, allow, deny []string) bool {
	if matchAny(name, deny) {
		return true
	}
	// If allow is present, require a match.
	if len(allow) > 0 && !matchAny(name, allow) {
		return true
	}
	return false
}

func matchAny(s string, pats []string) bool {
	for _, p := range pats {
		if patMatch(s, p) {
			return true
		}
	}
	return false
}

// patMatch: very small glob on dotted identifiers: '*', leading/trailing '*', or exact.
func patMatch(s, p string) bool {
	ls := strings.ToLower(s)
	lp := strings.ToLower(p)
	if lp == "*" || ls == lp {
		return true
	}
	if strings.HasPrefix(lp, "*") && strings.HasSuffix(lp, "*") {
		sub := strings.Trim(lp, "*")
		return strings.Contains(ls, sub)
	}
	if strings.HasPrefix(lp, "*") {
		suf := strings.TrimPrefix(lp, "*")
		return strings.HasSuffix(ls, suf)
	}
	if strings.HasSuffix(lp, "*") {
		pre := strings.TrimSuffix(lp, "*")
		return strings.HasPrefix(ls, pre)
	}
	return ls == lp
}

// ---- Convenience helpers to build ExecPolicy from parsed metadata ----

// MergeAllows merges additional allow patterns (deduplicated, case-insensitive).
func (p *ExecPolicy) MergeAllows(more ...string) {
	p.Allow = dedupMerge(p.Allow, more...)
}

// MergeDenies merges additional deny patterns (deduplicated, case-insensitive).
func (p *ExecPolicy) MergeDenies(more ...string) {
	p.Deny = dedupMerge(p.Deny, more...)
}

func dedupMerge(base []string, more ...string) []string {
	seen := make(map[string]struct{}, len(base)+len(more))
	out := make([]string, 0, len(base)+len(more))
	for _, v := range base {
		l := strings.ToLower(strings.TrimSpace(v))
		if l == "" {
			continue
		}
		if _, ok := seen[l]; !ok {
			seen[l] = struct{}{}
			out = append(out, v)
		}
	}
	for _, v := range more {
		l := strings.ToLower(strings.TrimSpace(v))
		if l == "" {
			continue
		}
		if _, ok := seen[l]; !ok {
			seen[l] = struct{}{}
			out = append(out, v)
		}
	}
	return out
}

// BuildGrants constructs a capability.GrantSet from parts.
func BuildGrants(grants []capability.Capability, limits capability.Limits) capability.GrantSet {
	return capability.GrantSet{
		Grants:   grants,
		Limits:   limits,
		Counters: capability.NewCounters(),
	}
}
