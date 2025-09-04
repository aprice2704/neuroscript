// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Defines the version-aware registry for capsules.
// filename: pkg/capsule/registry.go
// nlines: 150
// risk_rating: MEDIUM

package capsule

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/mod/semver"
)

// Capsule is a simple container for ship-with-interpreter docs/specs.
type Capsule struct {
	ID       string // Fully qualified id: <name>@<version>, e.g., "capsule/aeiou@2"
	Name     string // Stable logical name, e.g. "capsule/aeiou"
	Version  string // semantic or integer version, e.g. "2"
	MIME     string // e.g. "text/markdown; charset=utf-8"
	Content  string // markdown payload
	SHA256   string // hex sha256 of Content
	Size     int    // bytes
	Priority int    // optional ordering hint for List()
}

var (
	// registry stores capsules by name, then by version.
	registry = make(map[string]map[string]Capsule)
	mu       sync.RWMutex

	// nameRE validates the <name> part of a capsule id.
	nameRE = regexp.MustCompile(`^capsule/[a-z0-9._-]+$`)
)

// ValidateName returns nil if name is well-formed.
func ValidateName(name string) error {
	if !nameRE.MatchString(name) {
		return errors.New("invalid capsule name; expected capsule/<name>")
	}
	return nil
}

// Register adds (or replaces) a capsule. Safe for concurrent use and init().
func Register(c Capsule) error {
	if err := ValidateName(c.Name); err != nil {
		return fmt.Errorf("invalid name for capsule with version %s: %w", c.Version, err)
	}
	if c.Version == "" {
		return errors.New("capsule version cannot be empty")
	}

	if c.Content != "" && c.SHA256 == "" {
		sum := sha256.Sum256([]byte(c.Content))
		c.SHA256 = hex.EncodeToString(sum[:])
	}
	c.Size = len(c.Content)
	c.ID = fmt.Sprintf("%s@%s", c.Name, c.Version)

	mu.Lock()
	defer mu.Unlock()

	if _, ok := registry[c.Name]; !ok {
		registry[c.Name] = make(map[string]Capsule)
	}
	registry[c.Name][c.Version] = c
	return nil
}

// MustRegister is like Register but panics on error (useful in init).
func MustRegister(c Capsule) {
	if err := Register(c); err != nil {
		panic(err)
	}
}

// Get returns a specific version of a capsule by name.
func Get(name, version string) (Capsule, bool) {
	mu.RLock()
	defer mu.RUnlock()
	versions, ok := registry[name]
	if !ok {
		return Capsule{}, false
	}
	c, ok := versions[version]
	return c, ok
}

// GetLatest returns the highest version of a capsule.
// It uses integer comparison for versions that are simple integers,
// otherwise it uses semantic versioning comparison.
func GetLatest(name string) (Capsule, bool) {
	mu.RLock()
	defer mu.RUnlock()
	versions, ok := registry[name]
	if !ok || len(versions) == 0 {
		return Capsule{}, false
	}

	var latestVersion string
	var versionKeys []string
	for k := range versions {
		versionKeys = append(versionKeys, k)
	}

	// Try to sort by integer first
	sort.SliceStable(versionKeys, func(i, j int) bool {
		v1, err1 := strconv.Atoi(versionKeys[i])
		v2, err2 := strconv.Atoi(versionKeys[j])
		if err1 == nil && err2 == nil {
			return v1 > v2
		}
		// Fallback to semver or string compare if not simple integers
		// Add "v" prefix if it's missing for semver compatibility.
		sv1 := versionKeys[i]
		if !strings.HasPrefix(sv1, "v") {
			sv1 = "v" + sv1
		}
		sv2 := versionKeys[j]
		if !strings.HasPrefix(sv2, "v") {
			sv2 = "v" + sv2
		}
		return semver.Compare(sv1, sv2) > 0
	})

	latestVersion = versionKeys[0]
	return versions[latestVersion], true
}

// ListVersions returns all available versions for a given capsule name, sorted.
func ListVersions(name string) ([]string, bool) {
	mu.RLock()
	defer mu.RUnlock()
	versions, ok := registry[name]
	if !ok {
		return nil, false
	}

	var versionKeys []string
	for k := range versions {
		versionKeys = append(versionKeys, k)
	}
	sort.Strings(versionKeys)
	return versionKeys, true
}

// List returns all capsules, stable-ordered by Priority then ID.
func List() []Capsule {
	mu.RLock()
	defer mu.RUnlock()
	var allCapsules []Capsule
	for _, versions := range registry {
		for _, c := range versions {
			allCapsules = append(allCapsules, c)
		}
	}

	sort.Slice(allCapsules, func(i, j int) bool {
		if allCapsules[i].Priority != allCapsules[j].Priority {
			return allCapsules[i].Priority < allCapsules[j].Priority
		}
		return allCapsules[i].ID < allCapsules[j].ID
	})
	return allCapsules
}
