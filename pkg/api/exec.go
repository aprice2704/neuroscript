// api/exec.go
type ExecConfig struct {
	Cache         Cache  // may be nil
	SecretPrivKey []byte // for secret("…")
	MaxGas        uint64
}

// Exec refuses anything except a **LoadedUnit**
func Exec(ctx context.Context, lu *LoadedUnit,
	cfg ExecConfig) (*ExecResult, error)
