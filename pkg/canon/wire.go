// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 2
// :: description: Defines shared constants for the canonical binary wire format.
// :: latestChange: Stabilized magicNumber to 0x01 instead of volatile types.KindMarker.
// :: filename: pkg/canon/wire.go
// :: serialization: go

package canon

// magicNumber is a stable fingerprint ("NSC" + version) to identify a valid canonical blob.
// It was previously dynamically tied to types.KindMarker, which broke decoding when the AST grew.
var magicNumber = []byte{'N', 'S', 'C', 0x01}

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
