// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Tests the version-aware capsule registry.
// filename: pkg/capsule/registry_test.go
// nlines: 150
// risk_rating: LOW
package capsule_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strconv"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/api"
	"github.com/aprice2704/neuroscript/pkg/capsule"
)

func TestRegisterComputesSHAWhenEmpty(t *testing.T) {
	name := "capsule/sha-demo"
	content := "hello, capsule"

	if err := capsule.Register(capsule.Capsule{
		Name:    name,
		Version: "1",
		MIME:    "text/markdown; charset=utf-8",
		Content: content,
	}); err != nil {
		t.Fatalf("Register: %v", err)
	}

	c, ok := capsule.Get(name, "1")
	if !ok {
		t.Fatalf("Get(%q, '1') not found", name)
	}
	sum := sha256.Sum256([]byte(content))
	want := hex.EncodeToString(sum[:])
	if c.SHA256 != want {
		t.Fatalf("SHA mismatch: got %s, want %s", c.SHA256, want)
	}
	if c.Size != len(content) {
		t.Fatalf("Size mismatch: got %d, want %d", c.Size, len(content))
	}
	if c.ID != "capsule/sha-demo@1" {
		t.Errorf("Expected ID to be 'capsule/sha-demo@1', got %s", c.ID)
	}
}

func TestMustRegisterPanicsOnBadName(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("MustRegister should panic on invalid name")
		}
	}()
	capsule.MustRegister(capsule.Capsule{
		Name:    "Capsule/BadUpper", // invalid: uppercase "C"
		Version: "1",
		MIME:    "text/plain",
		Content: "x",
	})
}

func TestListOrderingByPriorityThenID(t *testing.T) {
	// Same priority, order by ID
	a := capsule.Capsule{Name: "capsule/sorta", Version: "1", MIME: "text/plain", Content: "A", Priority: 20}
	b := capsule.Capsule{Name: "capsule/sortb", Version: "1", MIME: "text/plain", Content: "B", Priority: 20}
	// Lower priority sorts first
	lo := capsule.Capsule{Name: "capsule/low", Version: "1", MIME: "text/plain", Content: "L", Priority: 10}

	for _, c := range []capsule.Capsule{a, b, lo} {
		capsule.MustRegister(c)
	}

	list := capsule.List()
	var got []string
	for _, c := range list {
		if c.Name == a.Name || c.Name == b.Name || c.Name == lo.Name {
			got = append(got, c.ID)
		}
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 test capsules in List(), got %d", len(got))
	}

	want := []string{"capsule/low@1", "capsule/sorta@1", "capsule/sortb@1"}
	sort.Strings(got)
	sort.Strings(want)
	if !equalStrings(got, want) {
		t.Fatalf("order mismatch: got %v, want %v", got, want)
	}
}

func TestGetLatest(t *testing.T) {
	name := "capsule/version-test"
	versions := []string{"1", "10", "2"}
	for _, v := range versions {
		capsule.MustRegister(capsule.Capsule{Name: name, Version: v, Content: "v" + v})
	}

	latest, ok := capsule.GetLatest(name)
	if !ok {
		t.Fatalf("GetLatest(%q) failed", name)
	}
	if latest.Version != "10" {
		t.Errorf("GetLatest version mismatch: got %s, want 10", latest.Version)
	}

	// Test with semver
	semverName := "capsule/semver-test"
	semverVersions := []string{"1.0.0", "1.1.0", "0.9.0"}
	for _, v := range semverVersions {
		capsule.MustRegister(capsule.Capsule{Name: semverName, Version: v, Content: "v" + v})
	}
	latestSemver, ok := capsule.GetLatest(semverName)
	if !ok {
		t.Fatalf("GetLatest(%q) failed", semverName)
	}
	if latestSemver.Version != "1.1.0" {
		t.Errorf("GetLatest semver mismatch: got %s, want 1.1.0", latestSemver.Version)
	}
}

func TestListVersions(t *testing.T) {
	name := "capsule/list-versions-test"
	versions := []string{"1", "3", "2"}
	for _, v := range versions {
		capsule.MustRegister(capsule.Capsule{Name: name, Version: v})
	}

	vlist, ok := capsule.ListVersions(name)
	if !ok {
		t.Fatalf("ListVersions(%q) failed", name)
	}
	sort.Strings(vlist)
	want := []string{"1", "2", "3"}
	if !equalStrings(vlist, want) {
		t.Errorf("ListVersions mismatch: got %v, want %v", vlist, want)
	}
}

func TestValidateNameCases(t *testing.T) {
	cases := []struct {
		name  string
		valid bool
	}{
		{"capsule/aeiou", true},
		{"capsule/foo-bar_9", true},
		{"Capsule/bad", false},       // uppercase not allowed
		{"capsule/Bad", false},       // uppercase in name
		{"capsule/space bad", false}, // space
		{"capsule/missingver", true}, // this is now valid
		{"capsule/", false},          // empty name
		{"foo/bar", false},           // wrong prefix
	}
	for _, tc := range cases {
		err := capsule.ValidateName(tc.name)
		if tc.valid && err != nil {
			t.Errorf("ValidateName(%q) unexpected error: %v", tc.name, err)
		}
		if !tc.valid && err == nil {
			t.Errorf("ValidateName(%q) expected error, got nil", tc.name)
		}
	}
}

func TestToolListAndRead(t *testing.T) {
	// This test relies on the loader having run.
	const name = "capsule/aeiou"

	tool := api.NewCapsuleTool()
	ctx := context.Background()

	// List all
	all := tool.List(ctx, nil)
	// We can't know the exact ID because version is now dynamic,
	// so we check if *any* version of the capsule is present.
	found := false
	for id := range all {
		if _, err := strconv.Atoi(id[len(name)+1:]); err == nil {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("tool.List(nil) did not include any version of %q", name)
	}

	// Read latest
	latest, ok := capsule.GetLatest(name)
	if !ok {
		t.Fatalf("Could not get latest version of %q for test", name)
	}

	content, ver, sha, mime, ok := tool.Read(ctx, latest.ID)
	if !ok {
		t.Fatalf("tool.Read(%q) ok=false", latest.ID)
	}
	if content == "" || ver == "" || sha == "" || mime == "" {
		t.Fatalf("tool.Read returned empty fields")
	}

	// Read not found
	if _, _, _, _, ok := tool.Read(ctx, "capsule/does-not-exist@1"); ok {
		t.Fatalf("tool.Read(nonexistent) ok=true, want false")
	}
}

// Helper
func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sort.Strings(a)
	sort.Strings(b)
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
