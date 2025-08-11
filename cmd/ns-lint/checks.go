// NeuroScript Version: 0.3.0
// File version: 2
// Purpose: Linter core: walk files, parse metadata headers, perform basic policy/effects checks, report findings.
// filename: cmd/ns-lint/checks.go
// nlines: 163
// risk_rating: MEDIUM

package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type Finding struct {
	Path    string
	Line    int
	Level   string // "ERROR","WARN","INFO"
	Message string
}

type Findings []Finding

func (f Findings) HasErrors() bool {
	for _, x := range f {
		if x.Level == "ERROR" {
			return true
		}
	}
	return false
}

func (f Findings) add(path string, line int, level, msg string) Findings {
	return append(f, Finding{Path: path, Line: line, Level: level, Message: msg})
}

// Run lints all .ns files found under the provided paths.
func Run(paths []string) (Findings, error) {
	files, err := collectNS(paths) // defined in metadata.go
	if err != nil {
		return nil, err
	}
	var out Findings
	for _, f := range files {
		meta, firstNonMetaLine, err := parseFileHeaderMetadata(f)
		if err != nil {
			out = out.add(f, 1, "ERROR", fmt.Sprintf("parse metadata: %v", err))
			continue
		}
		out = append(out, lintFile(f, meta, firstNonMetaLine)...)
	}
	return out, nil
}

func PrintFindings(all Findings) {
	if len(all) == 0 {
		fmt.Println("ns-lint: OK (no findings)")
		return
	}
	for _, f := range all {
		fmt.Printf("%s:%d: %s: %s\n", f.Path, f.Line, f.Level, f.Message)
	}
}

// -------------------- header parsing (file-level only) --------------------

type MetaEntry struct {
	Key   string
	Value string
	Line  int
}

func parseFileHeaderMetadata(path string) ([]MetaEntry, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()

	var metas []MetaEntry
	sc := bufio.NewScanner(f)
	line := 0
	for sc.Scan() {
		line++
		s := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(s, "::") {
			kv := strings.TrimPrefix(s, "::")
			i := strings.IndexByte(kv, ':')
			if i <= 0 {
				return nil, 0, fmt.Errorf("malformed metadata line %d", line)
			}
			key := strings.TrimSpace(kv[:i])
			val := strings.TrimSpace(kv[i+1:])
			metas = append(metas, MetaEntry{Key: key, Value: val, Line: line})
			continue
		}
		// stop header parse at first non-metadata line
		return metas, line, nil
	}
	if err := sc.Err(); err != nil {
		return nil, 0, err
	}
	return metas, line + 1, nil
}

// -------------------- checks --------------------

func lintFile(path string, header []MetaEntry, bodyStart int) Findings {
	var out Findings

	// Normalize keys -> values for quick lookups
	m := map[string]MetaEntry{}
	for _, e := range header {
		m[strings.ToLower(e.Key)] = e
	}

	// 1) policyContext presence on files that look like config
	// Heuristic: if they allow trusted tools or grant admin, they should be config.
	if looksTrusted(m) {
		if v, ok := m["policycontext"]; !ok || !isOneOf(strings.ToLower(v.Value), "config") {
			out = out.add(path, pickLine(m, "policycontext", 1), "ERROR",
				"file grants admin/trusted powers but ::policyContext is not 'config'")
		}
	}

	// 2) effects vs pure consistency
	if eff, ok := m["effects"]; ok {
		if pure, ok2 := m["pure"]; ok2 && strings.EqualFold(strings.TrimSpace(pure.Value), "true") {
			// forbid effects that imply impurity
			if containsEffect(eff.Value, "readsnet") || containsEffect(eff.Value, "readsfs") ||
				containsEffect(eff.Value, "readsclock") || containsEffect(eff.Value, "readsrand") {
				out = out.add(path, eff.Line, "ERROR", "::effects indicate impurity but ::pure is true")
			}
		}
	}

	// 3) wildcard risk warnings (too broad grants)
	for key, e := range m {
		lkey := strings.ToLower(key)
		if strings.HasPrefix(lkey, "grant.fs.write") && strings.Contains(e.Value, "*") {
			out = out.add(path, e.Line, "WARN", "grant.fs.write uses wildcard '*' (discouraged)")
		}
		if strings.HasPrefix(lkey, "grant.net.read") && hasVeryBroadNet(e.Value) {
			out = out.add(path, e.Line, "WARN", "grant.net.read is overly broad (consider pinning hosts/ports)")
		}
		if strings.HasPrefix(lkey, "grant.model.admin") && strings.Contains(e.Value, "*") {
			out = out.add(path, e.Line, "WARN", "grant.model.admin: '*' allows admin on all models")
		}
	}

	// 4) budget sanity: perCall must not exceed max
	if max, ok := m["limit.budget.cad.max"]; ok {
		if per, ok2 := m["limit.budget.cad.percall"]; ok2 {
			if toInt(per.Value) > toInt(max.Value) && toInt(max.Value) > 0 {
				out = out.add(path, per.Line, "WARN", "::limit.budget.CAD.perCall exceeds ::limit.budget.CAD.max")
			}
		}
	}

	// Optional info: mark no metadata
	if len(header) == 0 {
		out = out.add(path, 1, "INFO", "no metadata header found")
	}

	_ = bodyStart
	return out
}

func looksTrusted(m map[string]MetaEntry) bool {
	// If allowlist includes tools that are commonly trusted/config-only
	if a, ok := m["policyallow"]; ok {
		val := strings.ToLower(a.Value)
		if strings.Contains(val, "tool.agentmodel.register") ||
			strings.Contains(val, "tool.agentmodel.delete") ||
			strings.Contains(val, "tool.sandbox.setprofile") ||
			strings.Contains(val, "tool.os.getenv") {
			return true
		}
	}
	// If any admin grants are present
	for k := range m {
		lk := strings.ToLower(k)
		if strings.HasPrefix(lk, "grant.model.admin") ||
			strings.HasPrefix(lk, "grant.sandbox.admin") ||
			strings.HasPrefix(lk, "grant.proc.exec") {
			return true
		}
	}
	return false
}

func pickLine(m map[string]MetaEntry, key string, fallback int) int {
	if e, ok := m[strings.ToLower(key)]; ok {
		return e.Line
	}
	return fallback
}

func isOneOf(v string, want ...string) bool {
	v = strings.TrimSpace(strings.ToLower(v))
	for _, w := range want {
		if v == strings.ToLower(w) {
			return true
		}
	}
	return false
}

func containsEffect(effects, needle string) bool {
	needle = strings.ToLower(strings.TrimSpace(needle))
	parts := splitCSV(effects)
	for _, p := range parts {
		if strings.ToLower(strings.TrimSpace(p)) == needle {
			return true
		}
	}
	return false
}

func hasVeryBroadNet(val string) bool {
	val = strings.ToLower(val)
	// crude detection of patterns like "*", "*:443", "*.com", "*.cloud", etc.
	if strings.Contains(val, "*") {
		return true
	}
	re := regexp.MustCompile(`(^|[, ])\*\.`)
	return re.MatchString(val)
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func toInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	// simple decimal-only parse; ignore errors safely
	n := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// -------------------- end --------------------
