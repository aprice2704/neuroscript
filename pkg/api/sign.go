package api

// SignedAST is a container for the canonical representation of a script
// and its digital signature.
type SignedAST struct {
	Blob []byte
	Sum  [32]byte
	Sig  []byte
}
