// NeuroScript Version: 0.6.2
// File version: 2
// Purpose: Shared wire-format tags, magic, and versions used by encoder/decoder.
// Filename: pkg/canon/wire_tags.go
// Risk rating: LOW

package canon

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
