// NeuroScript Version: 0.3.0
// File version: 1 // New file
// Purpose: Contains test case data for TestCapsSatisfied in matcher_test.go.
// filename: pkg/policy/capability/matcher_test_cases.go
// nlines: 212 // Approximate line count
// risk_rating: LOW

package capability

// capsSatisfiedTestCase defines the structure for capability matching tests.
// Needs and grants are strings here to allow testing MustParse and direct struct creation.
type capsSatisfiedTestCase struct {
	name            string
	needs           []string // Capability strings (MustParse format)
	grants          []string // Capability strings (MustParse format, or "*"/"*:*" for special cases)
	expectSatisfied bool
}

// capsSatisfiedTestCases holds all test data for TestCapsSatisfied.
var capsSatisfiedTestCases = []capsSatisfiedTestCase{
	// --- Basic Exact Matches ---
	{
		name:            "Exact Match: Simple",
		needs:           []string{"env:read:API_KEY"},
		grants:          []string{"env:read:API_KEY"},
		expectSatisfied: true,
	},
	{
		name:            "Exact Match: Multiple Verbs (Subset)",
		needs:           []string{"fs:read:/data"},
		grants:          []string{"fs:read,write:/data"},
		expectSatisfied: true,
	},
	{
		name:            "Exact Match: Multiple Scopes (Subset)",
		needs:           []string{"net:read:host1.com"},
		grants:          []string{"net:read:host1.com,host2.com"},
		expectSatisfied: true,
	},
	{
		name:            "Exact Match: Case Insensitive Resource",
		needs:           []string{"FS:read:/tmp"},
		grants:          []string{"fs:read:/tmp"},
		expectSatisfied: true,
	},
	{
		name:            "Exact Match: Case Insensitive Verb",
		needs:           []string{"fs:READ:/tmp"},
		grants:          []string{"fs:read:/tmp"},
		expectSatisfied: true,
	},
	{
		name:            "Exact Match: Case Insensitive Scope (Default)", // Most scopes are case-sensitive, but default matcher isn't
		needs:           []string{"env:read:api_key"},
		grants:          []string{"env:read:API_KEY"},
		expectSatisfied: true, // Simple wildcard/exact matching is case-insensitive
	},

	// --- Basic Mismatches ---
	{
		name:            "Mismatch: Wrong Resource",
		needs:           []string{"net:read:host1.com"},
		grants:          []string{"fs:read:host1.com"},
		expectSatisfied: false,
	},
	{
		name:            "Mismatch: Wrong Verb",
		needs:           []string{"fs:write:/data"},
		grants:          []string{"fs:read:/data"},
		expectSatisfied: false,
	},
	{
		name:            "Mismatch: Missing Verb",
		needs:           []string{"fs:read,write:/data"},
		grants:          []string{"fs:read:/data"},
		expectSatisfied: false,
	},
	{
		name:            "Mismatch: Wrong Scope",
		needs:           []string{"env:read:KEY_A"},
		grants:          []string{"env:read:KEY_B"},
		expectSatisfied: false,
	},
	{
		name:            "Mismatch: Missing Scope",
		needs:           []string{"env:read:KEY_A,KEY_B"},
		grants:          []string{"env:read:KEY_A"},
		expectSatisfied: false, // Needs KEY_B as well
	},

	// --- Universal Wildcard Grant ---
	{
		name:            "Universal Grant: Simple Need",
		needs:           []string{"fs:read:/tmp/file.txt"},
		grants:          []string{"*:*:*"},
		expectSatisfied: true,
	},
	{
		name:            "Universal Grant: Need Multiple Verbs/Scopes",
		needs:           []string{"net:read,write:host1.com,host2.com"},
		grants:          []string{"*:*:*"},
		expectSatisfied: true,
	},
	{
		name:            "Universal Grant: Empty Need (Vacuously True)",
		needs:           []string{},
		grants:          []string{"*:*:*"},
		expectSatisfied: true,
	},

	// --- Resource Wildcard Grant ---
	{
		name:            "Resource Wildcard Grant: Simple Need",
		needs:           []string{"fs:read:/tmp/file.txt"},
		grants:          []string{"*:read:/tmp/file.txt"},
		expectSatisfied: true,
	},
	{
		name:            "Resource Wildcard Grant: Mismatch Verb",
		needs:           []string{"fs:write:/tmp/file.txt"},
		grants:          []string{"*:read:/tmp/file.txt"},
		expectSatisfied: false,
	},
	{
		name:            "Resource Wildcard Grant: Mismatch Scope",
		needs:           []string{"fs:read:/data/other.txt"},
		grants:          []string{"*:read:/tmp/file.txt"},
		expectSatisfied: false,
	},

	// --- Verb Wildcard Grant ---
	{
		name:            "Verb Wildcard Grant: Simple Need",
		needs:           []string{"fs:read:/tmp/file.txt"},
		grants:          []string{"fs:*:/tmp/file.txt"},
		expectSatisfied: true,
	},
	{
		name:            "Verb Wildcard Grant: Need Multiple Verbs",
		needs:           []string{"fs:read,write:/tmp/file.txt"},
		grants:          []string{"fs:*:/tmp/file.txt"},
		expectSatisfied: true,
	},
	{
		name:            "Verb Wildcard Grant: Mismatch Resource",
		needs:           []string{"net:read:/tmp/file.txt"},
		grants:          []string{"fs:*:/tmp/file.txt"},
		expectSatisfied: false,
	},
	{
		name:            "Verb Wildcard Grant: Mismatch Scope",
		needs:           []string{"fs:read:/data/other.txt"},
		grants:          []string{"fs:*:/tmp/file.txt"},
		expectSatisfied: false,
	},

	// --- Scope Wildcard Grant ---
	{
		name:            "Scope Wildcard Grant: Simple Need",
		needs:           []string{"fs:read:/tmp/file.txt"},
		grants:          []string{"fs:read:*"},
		expectSatisfied: true,
	},
	{
		name:            "Scope Wildcard Grant: Need Multiple Scopes",
		needs:           []string{"fs:read:/tmp/a,/tmp/b"},
		grants:          []string{"fs:read:*"},
		expectSatisfied: true,
	},
	{
		name:            "Scope Wildcard Grant: Mismatch Resource",
		needs:           []string{"net:read:/tmp/file.txt"},
		grants:          []string{"fs:read:*"},
		expectSatisfied: false,
	},
	{
		name:            "Scope Wildcard Grant: Mismatch Verb",
		needs:           []string{"fs:write:/tmp/file.txt"},
		grants:          []string{"fs:read:*"},
		expectSatisfied: false,
	},

	// --- Resource-Specific Scope Matching ---
	// FS (Glob)
	{
		name:            "FS Scope: Glob Grant Match",
		needs:           []string{"fs:read:/home/user/data.log"},
		grants:          []string{"fs:read:/home/user/*.log"},
		expectSatisfied: true,
	},
	{
		name:            "FS Scope: Glob Grant Mismatch",
		needs:           []string{"fs:read:/home/other/data.log"},
		grants:          []string{"fs:read:/home/user/*.log"},
		expectSatisfied: false,
	},
	{
		name:            "FS Scope: Glob Star Grant",
		needs:           []string{"fs:read:/any/path/anywhere"},
		grants:          []string{"fs:read:*"}, // Treated as universal scope for FS
		expectSatisfied: true,
	},
	// Net (Host)
	{
		name:            "Net Scope: *.domain Grant Match Subdomain",
		needs:           []string{"net:read:api.example.com"},
		grants:          []string{"net:read:*.example.com"},
		expectSatisfied: true,
	},
	{
		name:            "Net Scope: *.domain Grant Match Base Domain",
		needs:           []string{"net:read:example.com"},
		grants:          []string{"net:read:*.example.com"},
		expectSatisfied: true,
	},
	{
		name:            "Net Scope: *.domain Grant Mismatch Other Domain",
		needs:           []string{"net:read:api.other.com"},
		grants:          []string{"net:read:*.example.com"},
		expectSatisfied: false,
	},
	{
		name:            "Net Scope: Simple Wildcard Grant Match",
		needs:           []string{"net:read:api.example.com"},
		grants:          []string{"net:read:*.example.*"},
		expectSatisfied: true,
	},
	{
		name:            "Net Scope: Simple Wildcard Grant Mismatch",
		needs:           []string{"net:read:api.example.org"},
		grants:          []string{"net:read:*.example.com"},
		expectSatisfied: false,
	},
	{
		name:            "Net Scope: Port Match",
		needs:           []string{"net:read:host.com:8080"},
		grants:          []string{"net:read:host.com:8080"},
		expectSatisfied: true,
	},
	{
		name:            "Net Scope: Port Mismatch",
		needs:           []string{"net:read:host.com:8080"},
		grants:          []string{"net:read:host.com:9090"},
		expectSatisfied: false,
	},
	{
		name:            "Net Scope: Grant Any Port",
		needs:           []string{"net:read:host.com:8080"},
		grants:          []string{"net:read:host.com"},
		expectSatisfied: true,
	},
	{
		name:            "Net Scope: Need Any Port", // Needing any port isn't really a concept, need implies specificity
		needs:           []string{"net:read:host.com"},
		grants:          []string{"net:read:host.com:8080"},
		expectSatisfied: true, // Granting specific port satisfies need for host access
	},
	// Env (Simple Wildcard)
	{
		name:            "Env Scope: Prefix Wildcard Grant Match",
		needs:           []string{"env:read:SECRET_KEY_1"},
		grants:          []string{"env:read:SECRET_*"},
		expectSatisfied: true,
	},
	{
		name:            "Env Scope: Prefix Wildcard Grant Mismatch",
		needs:           []string{"env:read:API_KEY"},
		grants:          []string{"env:read:SECRET_*"},
		expectSatisfied: false,
	},

	// --- Multiple Needs / Grants ---
	{
		name: "Multiple Needs, One Grant",
		needs: []string{
			"fs:read:/data/a",
			"fs:read:/data/b",
		},
		grants:          []string{"fs:read:/data/*"},
		expectSatisfied: true,
	},
	{
		name: "Multiple Needs, Multiple Grants",
		needs: []string{
			"fs:read:/data/a",
			"env:read:KEY_A",
		},
		grants: []string{
			"fs:read:/data/a",
			"env:read:KEY_A",
		},
		expectSatisfied: true,
	},
	{
		name: "Multiple Needs, Missing One Grant",
		needs: []string{
			"fs:read:/data/a",
			"env:read:KEY_A",
		},
		grants:          []string{"fs:read:/data/a"},
		expectSatisfied: false,
	},

	// --- Negative Cases / Strictness ---
	// These use strings that MustParse would reject, handled in the test runner
	{
		name:            "Invalid Grant: Resource Only Wildcard (Direct Struct)",
		needs:           []string{"fs:read:/tmp"},
		grants:          []string{"*"}, // Will be converted to Capability{Resource: "*"}
		expectSatisfied: false,         // Should not satisfy because verbs/scopes are missing
	},
	{
		name:            "Invalid Grant: Resource/Verb Wildcard (Direct Struct)",
		needs:           []string{"fs:read:/tmp"},
		grants:          []string{"*:*"}, // Will be converted to Capability{Resource: "*", Verbs: []string{"*"}}
		expectSatisfied: false,           // Should not satisfy because scope is missing/doesn't match implicitly
	},
	// These use valid grants but test against incorrect resource/verb matching
	{
		name:            "Invalid Grant: Resource Wildcard Scope (Wrong Resource)",
		needs:           []string{"fs:read:/tmp"},
		grants:          []string{"net:*:*"},
		expectSatisfied: false,
	},
	{
		name:            "Invalid Grant: Verb Wildcard Scope (Wrong Resource)",
		needs:           []string{"fs:read:/tmp"},
		grants:          []string{"net:read:*"},
		expectSatisfied: false,
	},
	// This case assumes 'needs' are always valid and canonical
	// {
	// 	name:        "Invalid Need: Malformed Resource",
	// 	needs:       []string{"fs*:read:/tmp"}, // Wildcard in resource part of need is invalid conceptually
	// 	grants:      []string{"*:*:*"},
	// 	expectSatisfied: false, // Parse should ideally fail for the need, but if it didn't, match should fail here.
	// },
}
