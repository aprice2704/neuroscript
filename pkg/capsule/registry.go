// NeuroScript Version: 0.7.2
// File version: 10
// Purpose: Adds the 'Description' field to the struct to store the mandatory metadata.
// Latest change: Made the *Store writable (Register/Delete) to unify the capsule API.
// filename: pkg/capsule/registry.go
// nlines: 283

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
	ID          string // Fully qualified id: <name>@<version>, e.g., "capsule/aeiou@2"
	Name        string // Stable logical name, e.g. "capsule/aeiou"
	Version     string // semantic or integer version, e.g. "2"
	Description string // ADDED: The one-line description from metadata.
	MIME        string // e.g. "text/markdown; charset=utf-8"
	Content     string // markdown payload
	SHA256      string // hex sha256 of Content
	Size        int    // bytes
	Priority    int    // optional ordering hint for List()
}

var (
	// nameRE validates the <name> part of a capsule id.
	// --- THE FIX: Removed dot from allowed characters ---
	nameRE = regexp.MustCompile(`^capsule/[a-z0-9_-]+$`)
	// --- END FIX ---
)

// ValidateName returns nil if name is well-formed.
func ValidateName(name string) error {
	if !nameRE.MatchString(name) {
		return errors.New("invalid capsule name; expected capsule/<name> with only a-z, 0-9, _, -")
	}
	if strings.Contains(name, "@") {
		return errors.New("invalid capsule name; cannot contain '@'")
	}
	return nil
}

// --- Registry ---

// Registry is a collection of capsules. It is safe for concurrent use.
type Registry struct {
	mu       sync.RWMutex
	capsules map[string]map[string]Capsule // name -> version -> Capsule
}

// NewRegistry creates a new, empty capsule registry.
func NewRegistry() *Registry {
	return &Registry{
		capsules: make(map[string]map[string]Capsule),
	}
}

// Register adds (or replaces) a capsule in the registry.
func (r *Registry) Register(c Capsule) error {
	if err := ValidateName(c.Name); err != nil {
		return fmt.Errorf("invalid name for capsule with version %s: %w", c.Version, err)
	}
	if c.Version == "" {
		return errors.New("capsule version cannot be empty")
	}
	// Enforce integer-only versions
	if _, err := strconv.Atoi(c.Version); err != nil {
		return fmt.Errorf("capsule version must be an integer, but got %q", c.Version)
	}
	if c.Description == "" {
		// This was added as a required field in the loader
		return errors.New("capsule description cannot be empty")
	}

	if c.Content != "" && c.SHA256 == "" {
		sum := sha256.Sum256([]byte(c.Content))
		c.SHA256 = hex.EncodeToString(sum[:])
	}
	c.Size = len(c.Content)
	c.ID = fmt.Sprintf("%s@%s", c.Name, c.Version)

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.capsules[c.Name]; !ok {
		r.capsules[c.Name] = make(map[string]Capsule)
	}
	r.capsules[c.Name][c.Version] = c
	return nil
}

// MustRegister is like Register but panics on error.
func (r *Registry) MustRegister(c Capsule) {
	if err := r.Register(c); err != nil {
		panic(err)
	}
}

// Delete removes all versions of a capsule by its name.
// This makes the registry behave like a map for sync purposes.
func (r *Registry) Delete(name string) error {
	if err := ValidateName(name); err != nil {
		return fmt.Errorf("invalid name for capsule delete: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.capsules, name)
	return nil
}

// Get returns a specific version of a capsule by name from this registry.
func (r *Registry) Get(name, version string) (Capsule, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	versions, ok := r.capsules[name]
	if !ok {
		return Capsule{}, false
	}
	c, ok := versions[version]
	return c, ok
}

// GetLatest returns the highest version of a capsule from this registry.
func (r *Registry) GetLatest(name string) (Capsule, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	versions, ok := r.capsules[name]
	if !ok || len(versions) == 0 {
		return Capsule{}, false
	}

	var versionKeys []string
	for k := range versions {
		versionKeys = append(versionKeys, k)
	}

	sort.SliceStable(versionKeys, func(i, j int) bool {
		v1, err1 := strconv.Atoi(versionKeys[i])
		v2, err2 := strconv.Atoi(versionKeys[j])
		if err1 == nil && err2 == nil {
			return v1 > v2
		}
		// Fallback for non-integer versions, though Register now prevents them.
		// Kept for theoretical backward compatibility or other registry sources.
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

	latestVersion := versionKeys[0]
	return versions[latestVersion], true
}

// List returns all capsules in this registry.
func (r *Registry) List() []Capsule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var allCapsules []Capsule
	for _, versions := range r.capsules {
		for _, c := range versions {
			allCapsules = append(allCapsules, c)
		}
	}
	return allCapsules
}

// --- Store ---

// Store manages a layered set of capsule registries.
type Store struct {
	mu         sync.RWMutex
	registries []*Registry
}

// NewStore creates a new store, optionally initialized with a set of registries.
func NewStore(initial ...*Registry) *Store {
	return &Store{
		registries: initial,
	}
}

// Add adds a new registry as a new layer.
func (s *Store) Add(r *Registry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.registries = append(s.registries, r)
}

// Get finds a specific capsule version, searching registries in order.
func (s *Store) Get(name, version string) (Capsule, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.registries {
		if c, ok := r.Get(name, version); ok {
			return c, true
		}
	}
	return Capsule{}, false
}

// GetLatest finds the latest version of a capsule. It searches the first
// registry that contains the capsule name and returns the latest from there.
func (s *Store) GetLatest(name string) (Capsule, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.registries {
		// Check if the name exists at all in this registry layer.
		r.mu.RLock()
		_, ok := r.capsules[name]
		r.mu.RUnlock()

		if ok {
			// If it exists, get the latest from this layer and stop searching.
			return r.GetLatest(name)
		}
	}
	return Capsule{}, false
}

// List returns all capsules from all registries, sorted by priority then ID.
// It does not handle potential duplicates across registries.
func (s *Store) List() []Capsule {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var allCapsules []Capsule
	for _, r := range s.registries {
		allCapsules = append(allCapsules, r.List()...)
	}
	sort.Slice(allCapsules, func(i, j int) bool {
		if allCapsules[i].Priority != allCapsules[j].Priority {
			return allCapsules[i].Priority < allCapsules[j].Priority
		}
		return allCapsules[i].ID < allCapsules[i].ID
	})
	return allCapsules
}

// --- ADDED: Store Write Methods ---

// writableRegistry returns the first registry (index 0), which is designated
// as the writable layer for admin tools.
func (s *Store) writableRegistry() (*Registry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.registries) == 0 {
		return nil, errors.New("cannot write: capsule store has no registries")
	}
	return s.registries[0], nil
}

// Register adds a capsule to the store's writable layer (index 0).
func (s *Store) Register(c Capsule) error {
	reg, err := s.writableRegistry()
	if err != nil {
		return err
	}
	return reg.Register(c)
}

// Delete removes a capsule from the store's writable layer (index 0).
func (s *Store) Delete(name string) error {
	reg, err := s.writableRegistry()
	if err != nil {
		return err
	}
	return reg.Delete(name)
}

// --- END: Store Write Methods ---

// --- Default Singletons (FIXED) ---

var (
	builtInRegistry      *Registry // The registry for embedded content
	defaultStore         *Store    // The default store that contains the built-in registry
	defaultSingletonOnce sync.Once
)

// initSingletons initializes the built-in registry and the default store.
func initSingletons() {
	defaultSingletonOnce.Do(func() {
		builtInRegistry = NewRegistry()
		// THE FIX: The default store now has an empty, writable registry
		// at index 0, layered over the built-in one.
		defaultStore = NewStore(NewRegistry(), builtInRegistry)
	})
}

// BuiltInRegistry returns the singleton registry for built-in capsules.
// This is primarily used by the loader (e.g., in loader.go) to populate
// the default set of capsules.
func BuiltInRegistry() *Registry {
	initSingletons()
	return builtInRegistry
}

// DefaultStore returns the singleton, layered store.
// It is initialized with the BuiltInRegistry as its first layer.
// Application code should use this to get capsules or add new registries.
func DefaultStore() *Store {
	initSingletons()
	return defaultStore
}
