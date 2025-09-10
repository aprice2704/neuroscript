// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Provides comprehensive, dedicated unit tests for the scopeMatch function.
// filename: pkg/policy/capability/matcher_scope_test.go
// nlines: 125
// risk_rating: MEDIUM

package capability

import "testing"

func TestScopeMatch(t *testing.T) {
	testCases := []struct {
		name     string
		resource string
		need     string
		grant    string
		want     bool
	}{
		// --- Generic/Default Cases ---
		{"Default exact match", "other", "value", "value", true},
		{"Default case-insensitive match", "other", "Value", "value", true},
		{"Default wildcard match", "other", "any-value", "*", true},
		{"Default true match", "rand", "true", "true", true},
		{"Default no match", "other", "value1", "value2", false},

		// --- Env/Secrets/Model/Sandbox/Proc Cases ---
		{"Env exact", "env", "API_KEY", "api_key", true},
		{"Env prefix", "env", "STRIPE_API_KEY", "stripe_*", true},
		{"Env suffix", "env", "DEV_API_KEY", "*_api_key", true},
		{"Env substring", "env", "PROD_API_KEY_OLD", "*api_key*", true},
		{"Env no match", "env", "API_TOKEN", "api_key", false},
		{"Env grant *", "model", "gpt-4", "*", true},

		// --- FS Cases ---
		{"FS exact", "fs", "/data/file.txt", "/data/file.txt", true},
		{"FS glob star", "fs", "/data/file.txt", "/data/*", true},
		{"FS glob star no match", "fs", "/tmp/file.txt", "/data/*", false},
		{"FS glob question mark", "fs", "/data/file1.txt", "/data/file?.txt", true},
		{"FS glob question mark no match", "fs", "/data/file10.txt", "/data/file?.txt", false},
		{"FS grant *", "fs", "/any/path/whatsoever", "*", true},

		// --- Net Cases ---
		{"Net exact host", "net", "api.example.com", "api.example.com", true},
		{"Net exact host and port", "net", "api.example.com:443", "api.example.com:443", true},
		{"Net wildcard subdomain", "net", "sub.api.example.com", "*.example.com", true},
		{"Net wildcard subdomain with port", "net", "sub.api.example.com:8080", "*.example.com:8080", true},
		{"Net base domain match", "net", "example.com", "*.example.com", true},
		{"Net host match, port mismatch", "net", "api.example.com:443", "api.example.com:8080", false},
		{"Net host match, need has port, grant no port", "net", "api.example.com:443", "api.example.com", true},
		{"Net host match, grant has port, need no port", "net", "api.example.com", "api.example.com:443", true},
		{"Net wildcard host, port match", "net", "api-1.example.com:443", "*:443", true},
		{"Net no match", "net", "api.wrong.com", "*.example.com", false},
		{"Net IP prefix match", "net", "192.168.1.50", "192.168.*", true},
		{"Net IP prefix no match", "net", "10.0.0.5", "192.168.*", false},
		{"Net grant *", "net", "anything.whatsoever:1234", "*", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := scopeMatch(tc.resource, tc.need, tc.grant); got != tc.want {
				t.Errorf("scopeMatch(res: %q, need: %q, grant: %q) = %v, want %v", tc.resource, tc.need, tc.grant, got, tc.want)
			}
		})
	}
}

func TestHostMatch(t *testing.T) {
	testCases := []struct {
		name    string
		pattern string
		host    string
		want    bool
	}{
		{"Exact match", "host.com", "host.com", true},
		{"Case-insensitive match", "Host.Com", "host.com", true},
		{"Universal wildcard", "*", "anything.com", true},
		{"Prefix wildcard", "api.*", "api.example.com", true},
		{"Subdomain wildcard", "*.example.com", "api.example.com", true},
		{"Subdomain wildcard base match", "*.example.com", "example.com", true},
		{"Subdomain wildcard no match", "*.example.com", "wrong.com", false},
		{"Simple contains", "*middle*", "a.middle.b", true},
		{"No match", "host.com", "other.com", false},
		{"Partial no match", "api.host.com", "host.com", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := hostMatch(tc.pattern, tc.host); got != tc.want {
				t.Errorf("hostMatch(pattern: %q, host: %q) = %v, want %v", tc.pattern, tc.host, got, tc.want)
			}
		})
	}
}
