package all

// filename: toolbundles/all/all.go

import (
	// By importing these packages, Go guarantees their init() functions will run.
	_ "github.com/aprice2704/neuroscript/pkg/tool/ai"
	_ "github.com/aprice2704/neuroscript/pkg/tool/errtools"
	_ "github.com/aprice2704/neuroscript/pkg/tool/fs"
	_ "github.com/aprice2704/neuroscript/pkg/tool/git"
	_ "github.com/aprice2704/neuroscript/pkg/tool/gotools"
	_ "github.com/aprice2704/neuroscript/pkg/tool/io"
	_ "github.com/aprice2704/neuroscript/pkg/tool/list"
	_ "github.com/aprice2704/neuroscript/pkg/tool/maths"
	_ "github.com/aprice2704/neuroscript/pkg/tool/meta"
	_ "github.com/aprice2704/neuroscript/pkg/tool/script"
	_ "github.com/aprice2704/neuroscript/pkg/tool/shell"
	_ "github.com/aprice2704/neuroscript/pkg/tool/strtools"
	_ "github.com/aprice2704/neuroscript/pkg/tool/syntax"
	_ "github.com/aprice2704/neuroscript/pkg/tool/time"
	_ "github.com/aprice2704/neuroscript/pkg/tool/tree"
)
