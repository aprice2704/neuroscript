// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Provides a test to ensure all node kinds are handled by the decoder.
// filename: pkg/canon/coverage_test.go
// nlines: 40
// risk_rating: MEDIUM

package canon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/types"
)

// TestAllKindsHandledInDecoder serves as a safeguard to ensure that when a new
// types.Kind is added, the developer remembers to add a corresponding case in the
// decoder's readNode function.
func TestAllKindsHandledInDecoder(t *testing.T) {
	// The test iterates from the first possible Kind up to the unexported
	// kindMarker, which acts as a sentinel for the end of the enum.
	for i := types.Kind(0); i < types.KindMarker; i++ {
		t.Run(fmt.Sprintf("Kind_%d", i), func(t *testing.T) {
			if i == types.KindUnknown {
				// KindUnknown is not expected to be handled, so we skip it.
				t.Skip("Skipping KindUnknown as it is not a valid node kind")
			}

			// Create a minimal canonical blob containing only the magic number
			// and the varint for the current Kind.
			var buf bytes.Buffer
			buf.Write(magicNumber)
			varintBuf := make([]byte, binary.MaxVarintLen64)
			n := binary.PutVarint(varintBuf, int64(i))
			buf.Write(varintBuf[:n])

			// Attempt to decode it.
			_, err := Decode(buf.Bytes())

			// The test SUCCEEDS if the error is ANYTHING OTHER THAN
			// "unhandled node kind". We expect EOF or other parsing errors
			// because the blob is incomplete, but an "unhandled" error
			// means the switch statement is missing a case.
			if err != nil && strings.Contains(err.Error(), "unhandled node kind") {
				t.Errorf("Kind %d is not handled by the decoder's readNode switch statement", i)
			}
		})
	}
}
