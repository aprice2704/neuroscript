package ax

// DID is a named string for clarity at API boundaries.
type DID string

// ID is the minimal "who" contract (no crypto pulled in).
type ID interface {
	DID() DID
}

// Signer optionally adds signing capability tied to the same DID.
type Signer interface {
	ID
	PublicKey() []byte
	Sign(data []byte) ([]byte, error)
}

// IdentityCap means "this object can reveal the current actor identity."
type IdentityCap interface {
	Identity() ID
}

// SignerCap is optional; only implement where signing is truly needed.
type SignerCap interface {
	Signer() Signer
}
