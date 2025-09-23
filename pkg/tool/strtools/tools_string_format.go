// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Implements the 'Inspect' tool for pretty-printing variables.
// filename: pkg/tool/strtools/tools_string_format.go
// nlines: 100
// risk_rating: MEDIUM

package strtools

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

const (
	defaultMaxLength = 128
	defaultMaxDepth  = 5
)

func toolInspect(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) == 0 || len(args) > 3 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "Inspect: expected 1 to 3 arguments", lang.ErrArgumentMismatch)
	}

	target := args[0]

	maxLength := int64(defaultMaxLength)
	if len(args) > 1 {
		var ok bool
		maxLength, ok = toInt64(args[1])
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("Inspect: max_length must be an integer, got %T", args[1]), lang.ErrArgumentMismatch)
		}
	}

	maxDepth := int64(defaultMaxDepth)
	if len(args) > 2 {
		var ok bool
		maxDepth, ok = toInt64(args[2])
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("Inspect: max_depth must be an integer, got %T", args[2]), lang.ErrArgumentMismatch)
		}
	}

	s := newInspector(int(maxLength), int(maxDepth))
	s.inspect(target, 0)
	return s.String(), nil
}

type inspector struct {
	buf       bytes.Buffer
	maxLength int
	maxDepth  int
}

func newInspector(maxLength, maxDepth int) *inspector {
	return &inspector{
		maxLength: maxLength,
		maxDepth:  maxDepth,
	}
}

func (s *inspector) String() string {
	return s.buf.String()
}

func (s *inspector) inspect(v interface{}, depth int) {
	switch val := v.(type) {
	case nil:
		s.buf.WriteString("<nil>")
	case string:
		if len(val) > s.maxLength {
			s.buf.WriteString(strconv.Quote(val[:s.maxLength-3] + "..."))
		} else {
			s.buf.WriteString(strconv.Quote(val))
		}
	case []interface{}:
		if depth >= s.maxDepth {
			s.buf.WriteString("...")
			return
		}
		s.buf.WriteByte('[')
		for i, item := range val {
			if i > 0 {
				s.buf.WriteString(", ")
			}
			s.inspect(item, depth+1)
		}
		s.buf.WriteByte(']')
	case map[string]interface{}:
		if depth >= s.maxDepth {
			s.buf.WriteString("...")
			return
		}
		s.buf.WriteByte('{')
		i := 0
		for key, item := range val {
			if i > 0 {
				s.buf.WriteString(", ")
			}
			s.buf.WriteString(strconv.Quote(key))
			s.buf.WriteByte(':')
			s.inspect(item, depth+1)
			i++
		}
		s.buf.WriteByte('}')
	default:
		s.buf.WriteString(fmt.Sprintf("%v", val))
	}
}
