// parse.go
package api

type ParseMode uint8

const (
	ParsePreserveComments ParseMode = 1 << iota
	ParseSkipComments
)

func Parse(src []byte, mode ParseMode) (*Tree, error)
