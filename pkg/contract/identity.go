// Package contracts defines small, stable interfaces (the "NS Contract Layer")
// that other systems (api/fdm/zadeh/tools) can rely on without importing
// Neuroscript internals. Keep these tiny and dependency-free.
//
// Versioning: bump the comment "NS-CONTRACT vX.Y" if you add/rename methods.
// NS-CONTRACT v0.1

package contract

// DID is a stable identifier for an actor/agent (e.g., did:key:...).
type DID string

// Identity is the minimal self descriptor we expose outside NS.
type Identity interface {
	// DID returns the actor's decentralized identifier.
	DID() DID
}

// Signer provides signing capability tied to an Identity.
// Avoid adding Verify here; verification can be done by consumers with the pubkey.
type Signer interface {
	DID() DID
	// PublicKey returns the public key bytes for this identity.
	PublicKey() []byte
	// Sign returns a signature over data.
	Sign(data []byte) ([]byte, error)
}

// RuntimeIdentity is a capability interface that runtimes/interpreters can satisfy
// to expose the current actor identity (and optionally the signer).
type RuntimeIdentity interface {
	Identity() Identity
	// Optional, implement if available. Keep separate so callers can depend on
	// Identity() without dragging in signing.
	// Signer() Signer
}
