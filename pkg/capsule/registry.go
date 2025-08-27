package capsule

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"sort"
)

// Capsule is a simple container for ship-with-interpreter docs/specs.
type Capsule struct {
	ID       string // stable logical name, e.g. "capsule/aeiou/1"
	Version  string // semantic or integer version, e.g. "1"
	MIME     string // e.g. "text/markdown; charset=utf-8"
	Content  string // markdown payload
	SHA256   string // hex sha256 of Content
	Size     int    // bytes
	Priority int    // optional ordering hint for List()
}

var (
	registry = map[string]Capsule{}
	ids      = []string{}

	// ID must look like: capsule/<name>/<integer>
	idRE = regexp.MustCompile(`^capsule/[a-z0-9._-]+/[0-9]+$`)
)

// ValidateID returns nil if id is well-formed.
func ValidateID(id string) error {
	if !idRE.MatchString(id) {
		return errors.New("invalid capsule id; expected capsule/<name>/<int>")
	}
	return nil
}

// Register adds (or replaces) a capsule by ID. Safe for init() calls.
func Register(c Capsule) error {
	if err := ValidateID(c.ID); err != nil {
		return err
	}
	if c.Content != "" && c.SHA256 == "" {
		sum := sha256.Sum256([]byte(c.Content))
		c.SHA256 = hex.EncodeToString(sum[:])
	}
	c.Size = len(c.Content)
	registry[c.ID] = c
	reindex()
	return nil
}

// MustRegister is like Register but panics on error (useful in init).
func MustRegister(c Capsule) {
	if err := Register(c); err != nil {
		panic(err)
	}
}

// Get returns a capsule by ID.
func Get(id string) (Capsule, bool) {
	c, ok := registry[id]
	return c, ok
}

// List returns all capsules, stable-ordered by Priority then ID.
func List() []Capsule {
	out := make([]Capsule, 0, len(registry))
	for _, id := range ids {
		out = append(out, registry[id])
	}
	return out
}

func reindex() {
	ids = ids[:0]
	for id := range registry {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		ci, cj := registry[ids[i]], registry[ids[j]]
		if ci.Priority != cj.Priority {
			return ci.Priority < cj.Priority
		}
		return ci.ID < cj.ID
	})
}
