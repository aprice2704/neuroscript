// api/sign.go  (unchanged)
type SignedAST struct {
	Blob []byte
	Sum  [32]byte
	Sig  []byte
}

// api/load.go
type LoadedUnit struct { // ← result of Load
	Tree     *Tree    // verified & vetted
	Hash     [32]byte // same as SignedAST.Sum
	Mode     RunMode  // Library | Command | EventSink
	RawBytes []byte   // canonical bytes (for cache)
}

// Only transforms/validates – it **never runs** the code.
func Load(ctx context.Context, s *SignedAST,
	cfg LoaderConfig) (*LoadedUnit, error)
