// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Corrected grant satisfaction logic to handle capabilities with no required scopes.
// filename: pkg/capability/check.go
// nlines: 55
// risk_rating: HIGH

package capability

import "strings"

// Check determines if the GrantSet satisfies a required capability.
func (gs *GrantSet) Check(required Capability) bool {
	for _, grant := range gs.Grants {
		if grant.satisfies(required) {
			return true
		}
	}
	return false
}

// satisfies checks if a single grant can satisfy a required capability.
func (grant Capability) satisfies(required Capability) bool {
	if grant.Resource != required.Resource {
		return false
	}

	verbMatch := false
	for _, reqVerb := range required.Verbs {
		for _, grantVerb := range grant.Verbs {
			if grantVerb == "*" || grantVerb == reqVerb {
				verbMatch = true
				break
			}
		}
		if verbMatch {
			break
		}
	}
	if !verbMatch {
		return false
	}

	// FIX: If the required capability has no scopes, the check passes at this point.
	if len(required.Scopes) == 0 {
		return true
	}

	for _, reqScope := range required.Scopes {
		for _, grantScope := range grant.Scopes {
			if scopeMatches(grantScope, reqScope) {
				return true
			}
		}
	}

	return false
}

// scopeMatches checks if a grant scope (which can have wildcards) matches a required scope.
func scopeMatches(grantScope, requiredScope string) bool {
	if grantScope == "*" {
		return true
	}
	if strings.HasSuffix(grantScope, "*") {
		return strings.HasPrefix(requiredScope, strings.TrimSuffix(grantScope, "*"))
	}
	return grantScope == requiredScope
}
