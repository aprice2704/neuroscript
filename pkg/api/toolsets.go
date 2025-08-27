// NeuroScript Version: 0.6.0
// File version: 2
// Purpose: Adds the new aeiou toolset to the standard registration list.
// filename: pkg/api/toolsets.go
// nlines: 21
// risk_rating: LOW

package api

import (
	// This file ensures that the 'init()' functions in all standard tool
	// packages are executed, which registers them with the global tool registry.
	// The api.New() function then adds these registered tools to each new
	// interpreter instance.

	_ "github.com/aprice2704/neuroscript/pkg/tool/aeiou_proto"
	_ "github.com/aprice2704/neuroscript/pkg/tool/fs"
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
	// NOTE: Add other standard tool packages here as they are created.
)
