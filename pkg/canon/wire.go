// NeuroScript Version: 0.6.3
// File version: 1
// Purpose: Defines shared constants for the canonical binary wire format.
// filename: pkg/canon/wire.go
// nlines: 20
// risk_rating: LOW

package canon

import "github.com/aprice2704/neuroscript/pkg/types"

// magicNumber is a dynamic fingerprint ("NSC" + version) to identify a valid canonical blob.
var magicNumber = []byte{'N', 'S', 'C', byte(types.KindMarker)}

// CallableExpr payload header immediately after KindCallableExpr:
//
//	byte[2]  Magic "CE" (0x43, 0x45)
//	byte     Version (currently 0x01)
//	byte     Layout (1 = header, 2 = node-target), payload follows as per layout.
const (
	CallMagic1 = 0x43 // 'C'
	CallMagic2 = 0x45 // 'E'

	CallWireVersion = 0x01

	CallLayoutHeader     = 0x01 // bool isTool, string name, varint argc, args...
	CallLayoutNodeTarget = 0x02 // node target (Variable), varint argc, args...
)
