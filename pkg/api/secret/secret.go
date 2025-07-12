package secret

// secret/secret.go
type Ref struct {
	Path string // "prod/db/main"
	Enc  string // "none"|"age"|"sealedbox"
	Raw  []byte // encoded payload
}

func Decode(ref Ref, privKey []byte) (string, error)
