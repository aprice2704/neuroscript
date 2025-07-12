// api/reexport.go
package api

import "github.com/aprice2704/neuroscript/pkg/lang"

type (
	Position = lang.Position // type alias, zero overhead
	Kind     = ast.Kind
	Node     = ast.Node
	Tree     = ast.Tree
)
