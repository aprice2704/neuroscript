// NeuroScript Version: 0.8.0
// File version: 8
// Purpose: FIX: Refactored CanCall from a method to a standalone function to break the dependency on the ExecPolicy struct. Removed local ExecContext constants.
// filename: pkg/policy/policy.go
// nlines: 161
// risk_rating: HIGH

package policy

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/capability"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// ToolSpecProvider is an interface that a tool specification must satisfy
// for the policy engine to perform integrity checks.
type ToolSpecProvider interface {
	FullNameForChecksum() types.FullName
	ReturnTypeForChecksum() string
	ArgCountForChecksum() int
}

// LiveToolSpecFetcher is a function type for fetching tool specs at runtime.
type LiveToolSpecFetcher func(name string) (ToolSpecProvider, bool)

// ExecContext is an alias for the neutral interface type.
type ExecContext = interfaces.ExecContext

var (
	// These are now defined in interfaces, but aliased here for convenience
	// within this package. Consumers should use the re-exported constants from api.
	ContextConfig = interfaces.ContextConfig
	ContextNormal = interfaces.ContextNormal
	ContextTest   = interfaces.ContextTest
	ContextUser   = interfaces.ContextUser
)

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

// AllowAll creates a new policy that permits any tool call.
func AllowAll() *interfaces.ExecPolicy {
	return &interfaces.ExecPolicy{
		Context: ContextTest,
		Allow:   []string{"*"},
		Deny:    []string{},
		Grants: capability.GrantSet{
			Counters: capability.NewCounters(),
		},
	}
}

// CanCall enforces all security checks on the given policy in the correct order.
func CanCall(p *interfaces.ExecPolicy, t ToolMeta, fetcher LiveToolSpecFetcher) error {
	if err := validateIntegrity(p, t, fetcher); err != nil {
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

func validateIntegrity(p *interfaces.ExecPolicy, t ToolMeta, fetcher LiveToolSpecFetcher) error {
	if t.Name == "" || !validToolNameRegex.MatchString(t.Name) {
		msg := fmt.Sprintf("invalid tool name format '%s'", t.Name)
		return lang.NewRuntimeError(lang.ErrorCodeSubsystemCompromised, msg, lang.ErrSubsystemCompromised)
	}

	if fetcher == nil {
		return nil
	}

	spec, found := fetcher(t.Name)
	if !found {
		msg := fmt.Sprintf("tool spec for '%s' not found in registry for validation", t.Name)
		return lang.NewRuntimeError(lang.ErrorCodeSubsystemCompromised, msg, lang.ErrSubsystemCompromised)
	}

	expectedChecksum := CalculateChecksum(spec)
	if t.SignatureChecksum != "" && t.SignatureChecksum != expectedChecksum {
		msg := fmt.Sprintf("checksum mismatch for tool '%s'", t.Name)
		return lang.NewRuntimeError(lang.ErrorCodeSubsystemCompromised, msg, lang.ErrSubsystemCompromised)
	}
	return nil
}

// CalculateChecksum generates a stable hash of a tool's essential signature.
func CalculateChecksum(spec ToolSpecProvider) string {
	data := fmt.Sprintf("%s:%s:%d", spec.FullNameForChecksum(), spec.ReturnTypeForChecksum(), spec.ArgCountForChecksum())
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("sha256:%x", hash)
}

// disallowed decides policy outcome.
func disallowed(name string, allow, deny []string) bool {
	if matchAny(name, deny) {
		return true // Explicitly denied.
	}
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
