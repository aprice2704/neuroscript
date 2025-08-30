// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Defines the core policy engine for gating tool calls and validating capabilities.
// filename: pkg/policy/policy.go
// nlines: 173
// risk_rating: HIGH

package policy

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/policy/capability"
	"github.com/aprice2704/neuroscript/pkg/tool"
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
	Context             ExecContext
	Allow               []string
	Deny                []string
	Grants              capability.GrantSet
	LiveToolSpecFetcher func(name string) (tool.ToolSpec, bool)
}

// ToolMeta describes a tool for policy evaluation.
type ToolMeta struct {
	Name              string
	RequiresTrust     bool
	RequiredCaps      []capability.Capability
	Effects           []string
	SignatureChecksum string
}

var (
	ErrTrust           = errors.New("tool requires trusted context")
	ErrPolicy          = errors.New("tool not allowed by policy")
	ErrCapability      = errors.New("capabilities not granted")
	ErrIntegrity       = errors.New("tool integrity check failed")
	validToolNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_.]+$`)
)

// CanCall enforces all security checks in the correct order.
func (p *ExecPolicy) CanCall(t ToolMeta) error {
	if err := p.validateIntegrity(t); err != nil {
		return err
	}
	if t.RequiresTrust && p.Context != ContextConfig {
		return ErrTrust
	}
	if disallowed(t.Name, p.Allow, p.Deny) {
		return ErrPolicy
	}
	if !capability.CapsSatisfied(t.RequiredCaps, p.Grants.Grants) {
		return ErrCapability
	}
	return p.Grants.CountToolCall(t.Name)
}

func (p *ExecPolicy) validateIntegrity(t ToolMeta) error {
	if t.Name == "" || !validToolNameRegex.MatchString(t.Name) {
		msg := fmt.Sprintf("invalid tool name format '%s'", t.Name)
		return lang.NewRuntimeError(lang.ErrorCodeSubsystemCompromised, msg, lang.ErrSubsystemCompromised)
	}

	if p.LiveToolSpecFetcher == nil {
		return nil
	}

	spec, found := p.LiveToolSpecFetcher(t.Name)
	if !found {
		msg := fmt.Sprintf("tool spec for '%s' not found in registry for validation", t.Name)
		return lang.NewRuntimeError(lang.ErrorCodeSubsystemCompromised, msg, lang.ErrSubsystemCompromised)
	}

	expectedChecksum := calculateMockChecksum(spec)
	if t.SignatureChecksum != "" && t.SignatureChecksum != expectedChecksum {
		msg := fmt.Sprintf("checksum mismatch for tool '%s'", t.Name)
		return lang.NewRuntimeError(lang.ErrorCodeSubsystemCompromised, msg, lang.ErrSubsystemCompromised)
	}
	return nil
}

func calculateMockChecksum(spec tool.ToolSpec) string {
	data := fmt.Sprintf("%s:%s:%d", spec.FullName, spec.ReturnType, len(spec.Args))
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("sha256:%x", hash)
}

// disallowed decides policy outcome. Deny rules override allow rules.
// If an allow list is provided (even if empty), a tool must match it.
func disallowed(name string, allow, deny []string) bool {
	if matchAny(name, deny) {
		return true // Explicitly denied.
	}
	// FIX: If the allow list is not nil, it is active. An empty active list
	// means nothing is allowed.
	if allow != nil && !matchAny(name, allow) {
		return true // Not in the allow list.
	}
	return false // Allowed.
}

func matchAny(s string, pats []string) bool {
	for _, p := range pats {
		if patMatch(s, p) {
			return true
		}
	}
	return false
}

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

func (p *ExecPolicy) MergeAllows(more ...string) {
	p.Allow = dedupMerge(p.Allow, more...)
}

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

func BuildGrants(grants []capability.Capability, limits capability.Limits) capability.GrantSet {
	return capability.GrantSet{
		Grants:   grants,
		Limits:   limits,
		Counters: capability.NewCounters(),
	}
}
