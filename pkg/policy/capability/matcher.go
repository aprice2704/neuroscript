// NeuroScript Version: 0.3.0
// File version: 2
// Purpose: Capability matching helpers: fixed hostMatch logic for wildcards.
// filename: pkg/policy/capability/matcher.go
// nlines: 167
// risk_rating: MEDIUM

package capability

import (
	"path/filepath"
	"strings"
)

// CapsSatisfied returns true if every required capability in 'needs' is
// satisfied by at least one grant in 'grants'.
func CapsSatisfied(needs []Capability, grants []Capability) bool {
	for _, need := range needs {
		if !capSatisfied(need, grants) {
			return false
		}
	}
	return true
}

func capSatisfied(need Capability, grants []Capability) bool {
	for _, g := range grants {
		if strings.EqualFold(need.Resource, g.Resource) &&
			verbsSubset(need.Verbs, g.Verbs) &&
			scopesSubset(need.Resource, need.Scopes, g.Scopes) {
			return true
		}
	}
	return false
}

func verbsSubset(need, have []string) bool {
	if len(need) == 0 {
		return true
	}
	haveSet := make(map[string]struct{}, len(have))
	for _, v := range have {
		haveSet[strings.ToLower(v)] = struct{}{}
	}
	for _, v := range need {
		if _, ok := haveSet[strings.ToLower(v)]; !ok {
			return false
		}
	}
	return true
}

func scopesSubset(resource string, need, have []string) bool {
	if len(need) == 0 {
		return true
	}
	for _, n := range need {
		if !anyScopeMatches(resource, n, have) {
			return false
		}
	}
	return true
}

func anyScopeMatches(resource, n string, have []string) bool {
	for _, h := range have {
		if scopeMatch(resource, n, h) {
			return true
		}
	}
	return false
}

// scopeMatch implements minimal wildcard semantics by resource type:
//
//	env/secrets/model/sandbox/proc: exact or '*' or simple prefix/suffix wildcards.
//	fs: grant is a glob; need is a concrete path → filepath.Match(grant, need).
//	net: host[:port]; supports '*.' prefix and trailing '*' in grant. If port present in both, must match.
//	clock/rand/budget: boolean or exact token equality ("true","seed:123").
func scopeMatch(resource, need, grant string) bool {
	switch resource {
	case "env", "secrets", "model", "sandbox", "proc":
		return simpleWildcard(need, grant)
	case "fs":
		if grant == "*" {
			return true
		}
		ok, _ := filepath.Match(grant, need)
		return ok
	case "net":
		nh, np := splitHostPort(need)
		gh, gp := splitHostPort(grant)
		if gp != "" && np != "" && gp != np {
			return false
		}
		return hostMatch(gh, nh)
	default:
		if grant == "*" || strings.EqualFold(grant, need) || strings.EqualFold(grant, "true") {
			return true
		}
		return false
	}
}

func simpleWildcard(need, grant string) bool {
	if grant == "*" {
		return true
	}
	ln := strings.ToLower(need)
	lg := strings.ToLower(grant)
	if strings.HasPrefix(lg, "*") && strings.HasSuffix(lg, "*") {
		sub := strings.Trim(lg, "*")
		return strings.Contains(ln, sub)
	}
	if strings.HasPrefix(lg, "*") {
		suf := strings.TrimPrefix(lg, "*")
		return strings.HasSuffix(ln, suf)
	}
	if strings.HasSuffix(lg, "*") {
		pre := strings.TrimSuffix(lg, "*")
		return strings.HasPrefix(ln, pre)
	}
	return ln == lg
}

func splitHostPort(s string) (host, port string) {
	if i := strings.LastIndexByte(s, ':'); i > -1 {
		return s[:i], s[i+1:]
	}
	return s, ""
}

func hostMatch(pattern, host string) bool {
	lp := strings.ToLower(pattern)
	lh := strings.ToLower(host)
	if lp == "*" || lp == lh {
		return true
	}
	// FIX: Handle "*.domain.com" pattern to match base domain ("domain.com")
	// and any subdomain ("api.domain.com").
	if strings.HasPrefix(lp, "*.") {
		suf := strings.TrimPrefix(lp, "*.")
		return lh == suf || strings.HasSuffix(lh, "."+suf)
	}

	// FIX: Delegate other wildcard patterns (*.com, api.*, *middle*) to the
	// more robust simpleWildcard function.
	return simpleWildcard(host, pattern)
}
