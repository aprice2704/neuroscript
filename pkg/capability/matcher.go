// NeuroScript Version: 0.3.0
// File version: 5 // Bumped version
// Purpose: Capability matching helpers: Reinstates specific *.domain.com logic alongside filepath.Match fallback in hostMatch.
// filename: pkg/policy/capability/matcher.go
// nlines: 182 // Adjusted line count
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

// capSatisfied checks if a single needed capability is met by any grant.
func capSatisfied(need Capability, grants []Capability) bool {
	for _, g := range grants {
		// Check resource match (case-insensitive OR grant is wildcard)
		resourceMatch := strings.EqualFold(need.Resource, g.Resource) || g.Resource == "*"
		if !resourceMatch {
			continue // Try next grant if resources don't match
		}

		// Check verbs subset (case-insensitive OR grant verb is wildcard)
		verbsMatch := verbsSubset(need.Verbs, g.Verbs)
		if !verbsMatch {
			continue // Try next grant if verbs don't match
		}

		// Check scopes subset (uses specific rules per resource)
		scopesMatch := scopesSubset(need.Resource, need.Scopes, g.Scopes)
		if !scopesMatch {
			continue // Try next grant if scopes don't match
		}

		// If all checks pass for this grant, the need is satisfied
		return true
	}
	// If no grant satisfied the need
	return false
}

// verbsSubset checks if all needed verbs are present in the granted verbs.
func verbsSubset(need, have []string) bool {
	if len(need) == 0 {
		return true // No specific verbs needed, always satisfied
	}
	haveSet := make(map[string]struct{}, len(have))
	hasWildcard := false
	for _, v := range have {
		lowerV := strings.ToLower(v)
		haveSet[lowerV] = struct{}{}
		if lowerV == "*" {
			hasWildcard = true
		}
	}

	// If the grant includes the universal verb wildcard, it's always satisfied
	if hasWildcard {
		return true
	}

	// Otherwise, check if each needed verb is explicitly present
	for _, v := range need {
		if _, ok := haveSet[strings.ToLower(v)]; !ok {
			return false // A needed verb was not found
		}
	}
	return true // All needed verbs were found
}

// scopesSubset checks if all needed scopes are matched by any granted scope.
func scopesSubset(resource string, need, have []string) bool {
	if len(need) == 0 {
		return true // No specific scopes needed, always satisfied
	}
	// Check if *every* needed scope is matched by *at least one* granted scope
	for _, n := range need {
		if !anyScopeMatches(resource, n, have) {
			return false // A needed scope wasn't matched by any grant
		}
	}
	return true // All needed scopes were matched
}

// anyScopeMatches checks if a single needed scope 'n' is matched by any scope in 'have'.
func anyScopeMatches(resource, n string, have []string) bool {
	for _, h := range have {
		if scopeMatch(resource, n, h) {
			return true // Found a matching grant scope
		}
	}
	return false // No grant scope matched the needed scope
}

// scopeMatch implements minimal wildcard semantics by resource type:
//
//	env/secrets/model/sandbox/proc: exact or '*' or simple prefix/suffix wildcards.
//	fs: grant is a glob; need is a concrete path → filepath.Match(grant, need).
//	net: host[:port]; uses filepath.Match for host part, checks port separately.
//	clock/rand/budget: boolean or exact token equality ("true","seed:123").
func scopeMatch(resource, need, grant string) bool {
	switch resource {
	case "env", "secrets", "model", "sandbox", "proc":
		return simpleWildcard(need, grant)
	case "fs":
		// Universal grant scope matches any needed path
		if grant == "*" {
			return true
		}
		// Otherwise, treat grant as a glob pattern
		ok, _ := filepath.Match(grant, need)
		return ok
	case "net":
		nh, np := splitHostPort(need)
		gh, gp := splitHostPort(grant)
		// If ports are specified in both grant and need, they must match exactly.
		if gp != "" && np != "" && gp != np {
			return false
		}
		// If ports match (or aren't restrictively specified), check host match.
		return hostMatch(gh, nh)
	// Covers clock, rand, budget, and any other resource type
	default:
		// Universal grant scope matches anything, otherwise require exact match or "true"
		if grant == "*" || strings.EqualFold(grant, need) || strings.EqualFold(grant, "true") {
			return true
		}
		return false
	}
}

// simpleWildcard handles exact match, '*', prefix '*', suffix '*', and '*substring*' patterns (case-insensitive).
func simpleWildcard(need, grant string) bool {
	if grant == "*" {
		return true
	}
	ln := strings.ToLower(need)
	lg := strings.ToLower(grant)
	if strings.HasPrefix(lg, "*") && strings.HasSuffix(lg, "*") {
		sub := strings.Trim(lg, "*")
		return strings.Contains(ln, sub) // Check if need contains the substring
	}
	if strings.HasPrefix(lg, "*") {
		suf := strings.TrimPrefix(lg, "*")
		return strings.HasSuffix(ln, suf) // Check if need ends with suffix
	}
	if strings.HasSuffix(lg, "*") {
		pre := strings.TrimSuffix(lg, "*")
		return strings.HasPrefix(ln, pre) // Check if need starts with prefix
	}
	// If no wildcards, require exact match
	return ln == lg
}

// splitHostPort separates a host:port string.
func splitHostPort(s string) (host, port string) {
	if i := strings.LastIndexByte(s, ':'); i > -1 {
		// Basic check if this looks like an IPv6 address which also uses colons
		// This isn't foolproof but covers common cases.
		if strings.Count(s, ":") > 1 && strings.Contains(s, "]") { // Likely IPv6 like [::1]:8080
			if closingBracket := strings.LastIndexByte(s, ']'); closingBracket > i {
				// The last colon is part of the IPv6 address itself, no port specified
				return s, ""
			}
		}
		// Assume standard host:port
		return s[:i], s[i+1:]
	}
	// No colon found, assume it's just a host
	return s, ""
}

// hostMatch handles hostname matching with wildcards (case-insensitive).
// FIX: Reinstated specific *.domain.com logic AND uses filepath.Match for other wildcards.
func hostMatch(pattern, host string) bool {
	lp := strings.ToLower(pattern)
	lh := strings.ToLower(host)

	// 1. Exact match or universal pattern first
	if lp == "*" || lp == lh {
		return true
	}

	// 2. Handle the special "*.domain.com" pattern specifically
	if strings.HasPrefix(lp, "*.") {
		suf := strings.TrimPrefix(lp, "*.")
		// Check if host is exactly the suffix OR ends with ".suffix"
		if lh == suf || strings.HasSuffix(lh, "."+suf) {
			return true
		}
		// If the special pattern didn't match, fall through to general wildcard check
	}

	// 3. Use filepath.Match for any other patterns containing wildcards
	if strings.Contains(lp, "*") || strings.Contains(lp, "?") || strings.Contains(lp, "[") {
		// filepath.Match is case-sensitive on non-Windows OS by default,
		// but hostnames are generally case-insensitive, so we use lowercased versions.
		match, _ := filepath.Match(lp, lh)
		return match
	}

	// 4. If none of the above (e.g., just a plain hostname that didn't match lh exactly initially),
	// this final comparison ensures case-insensitivity was checked if needed.
	return lp == lh
}
