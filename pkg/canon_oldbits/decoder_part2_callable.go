// NeuroScript Version: 0.6.2
// File version: 6
// Purpose: CallableExpr decoder with unambiguous CE+version+layout header,
//          plus strong debug (offsets, hex sniff) for header/arg failures.
// Filename: pkg/canon/decoder_part2_callable.go
// Risk rating: MEDIUM

package canon

import (
	"encoding/hex"
	"fmt"
	"io"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// readByte reads one byte from the underlying stream.
func (r *canonReader) readByte() (byte, error) {
	return r.r.ReadByte()
}

func (r *canonReader) curOff() int64 {
	return r.r.Size() - int64(r.r.Len())
}

func (r *canonReader) peekN(n int) []byte {
	if n <= 0 {
		return nil
	}
	buf := make([]byte, n)
	pos := r.curOff()
	read, _ := r.r.Read(buf)
	if read > 0 {
		_, _ = r.r.Seek(pos, io.SeekStart)
	}
	return buf[:read]
}

// readCallableExpr (no guessing):
//
//	byte[2] magic "CE" (0x43, 0x45)
//	byte    version (0x01)
//	byte    layout  (CallLayoutHeader|CallLayoutNodeTarget)
//	payload per layout
func (r *canonReader) readCallableExpr() (*ast.CallableExprNode, error) {
	off := r.curOff()

	m1, err := r.readByte()
	if err != nil {
		return nil, fmt.Errorf("callable: read magic[0] at off=%d: %w", off, err)
	}
	m2, err := r.readByte()
	if err != nil {
		return nil, fmt.Errorf("callable: read magic[1] at off=%d: %w", off+1, err)
	}
	ver, err := r.readByte()
	if err != nil {
		return nil, fmt.Errorf("callable: read version at off=%d: %w", off+2, err)
	}
	layout, err := r.readByte()
	if err != nil {
		return nil, fmt.Errorf("callable: read layout at off=%d: %w", off+3, err)
	}

	if m1 != CallMagic1 || m2 != CallMagic2 || ver != CallWireVersion {
		sniff := hex.EncodeToString(r.peekN(8))
		return nil, fmt.Errorf(
			"callable: bad header at off=%d: got [%02X %02X] ver=%02X layout=%02X; next=%s",
			off, m1, m2, ver, layout, sniff,
		)
	}

	switch layout {
	case CallLayoutHeader:
		return r.readCallableHeader()
	case CallLayoutNodeTarget:
		return r.readCallableNodeTarget()
	default:
		return nil, fmt.Errorf("callable: unknown layout code %d at off=%d", layout, off)
	}
}

func (r *canonReader) readCallableHeader() (*ast.CallableExprNode, error) {
	start := r.curOff()

	isTool, err := r.readBool()
	if err != nil {
		return nil, fmt.Errorf("callable(header): isTool at off=%d: %w", start, err)
	}
	name, err := r.readString()
	if err != nil {
		return nil, fmt.Errorf("callable(header): name at off=%d: %w", r.curOff(), err)
	}
	argc, err := r.readVarint()
	if err != nil {
		return nil, fmt.Errorf("callable(header): argc at off=%d: %w", r.curOff(), err)
	}
	args := make([]ast.Expression, argc)
	for i := 0; i < int(argc); i++ {
		argOff := r.curOff()
		n, err := r.readNode()
		if err != nil {
			return nil, fmt.Errorf("callable(header): arg[%d] at off=%d: %w", i, argOff, err)
		}
		e, ok := n.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("callable(header): arg[%d] at off=%d: expected ast.Expression, got %T (next=%s)",
				i, argOff, n, hex.EncodeToString(r.peekN(8)))
		}
		args[i] = e
	}
	return &ast.CallableExprNode{
		BaseNode: ast.BaseNode{NodeKind: types.KindCallableExpr},
		Target: ast.CallTarget{
			BaseNode: ast.BaseNode{NodeKind: types.KindVariable}, // normalized in anneal
			IsTool:   isTool,
			Name:     name,
		},
		Arguments: args,
	}, nil
}

func (r *canonReader) readCallableNodeTarget() (*ast.CallableExprNode, error) {
	start := r.curOff()

	n, err := r.readNode()
	if err != nil {
		return nil, fmt.Errorf("callable(node-target): target at off=%d: %w", start, err)
	}
	vn, ok := n.(*ast.VariableNode)
	if !ok {
		return nil, fmt.Errorf("callable(node-target): expected *ast.VariableNode at off=%d, got %T", start, n)
	}
	argc, err := r.readVarint()
	if err != nil {
		return nil, fmt.Errorf("callable(node-target): argc at off=%d: %w", r.curOff(), err)
	}
	args := make([]ast.Expression, argc)
	for i := 0; i < int(argc); i++ {
		argOff := r.curOff()
		a, err := r.readNode()
		if err != nil {
			return nil, fmt.Errorf("callable(node-target): arg[%d] at off=%d: %w", i, argOff, err)
		}
		e, ok := a.(ast.Expression)
		if !ok {
			return nil, fmt.Errorf("callable(node-target): arg[%d] at off=%d: expected ast.Expression, got %T (next=%s)",
				i, argOff, a, hex.EncodeToString(r.peekN(8)))
		}
		args[i] = e
	}
	return &ast.CallableExprNode{
		BaseNode: ast.BaseNode{NodeKind: types.KindCallableExpr},
		Target: ast.CallTarget{
			BaseNode: ast.BaseNode{NodeKind: types.KindVariable}, // normalized in anneal
			IsTool:   false,
			Name:     vn.Name,
		},
		Arguments: args,
	}, nil
}
