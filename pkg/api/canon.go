// canon.go
func Canonicalise(tree *Tree) (blob []byte, sum [32]byte) // varint-packed
func Decode(blob []byte) (*Tree, [32]byte, error)         // verify shape only
