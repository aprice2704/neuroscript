package analysis

// analysis/pass.go
type Severity uint8

const (
	SevInfo Severity = iota
	SevWarn
	SevError
)

type Diag struct {
	Pos      Position
	Severity Severity
	Pass     string
	Message  string
}

type Pass interface {
	Name() string
	Analyse(tree *Tree) []Diag
}

func RegisterPass(p Pass)
func Vet(tree *Tree) []Diag // runs all registered passes
