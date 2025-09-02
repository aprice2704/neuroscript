// NeuroScript Version: 0.6.0
// File version: 1.0.0
// Purpose: Imports all standard toolsets to enable them via the registration pattern.
// filename: pkg/toolbundles/all/all.go
// nlines: 10
// risk_rating: LOW

// Package all provides a convenient way to include all standard NeuroScript
// toolsets in a build by simply importing this package.
package all

import (
	// Filesystem toolset
	_ "github.com/aprice2704/neuroscript/pkg/tool/fs"
	// OS-level toolset
	_ "github.com/aprice2704/neuroscript/pkg/tool/os"
	// AgentModel management toolset
	_ "github.com/aprice2704/neuroscript/pkg/tool/account"
	_ "github.com/aprice2704/neuroscript/pkg/tool/agentmodel"
)
