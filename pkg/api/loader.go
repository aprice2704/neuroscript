// loader.go
type RunMode uint8

const (
	RunModeLibrary   RunMode = iota // funcs only
	RunModeCommand                  // unnamed command block, run-once
	RunModeEventSink                // one or more on-event handlers
)

func DetectRunMode(tree *Tree) RunMode

// LoaderConfig toggles caching, gas limits, secret decode key, etc.
type LoaderConfig struct {
	Cache         Cache  // optional content-addressed cache
	MaxGas        uint64 // 0 = no limit
	SecretPrivKey []byte // for secret("…", enc="age")
}

type ExecResult struct {
	Outputs any      // optional run value
	Logs    []string // or richer slog handler
}

func Load(ctx context.Context, ast *SignedAST,
	cfg LoaderConfig) (*ExecResult, error)

func Exec(ctx context.Context, ast *SignedAST,
	cfg LoaderConfig) (*ExecResult, error)
