// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 1
// :: description: Tests to ensure the codec registry is complete.
// :: latestChange: Created test TestAllKindsHaveCodecs.
// :: filename: pkg/canon/codec_registry_test.go
// :: serialization: go

package canon

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/types"
)

func TestAllKindsHaveCodecs(t *testing.T) {
	// These kinds do not correspond to standalone serialized nodes
	// and are deliberately omitted from the registry.
	exceptions := map[types.Kind]bool{
		types.KindUnknown:      true,
		types.KindMetadataLine: true,
		types.KindMarker:       true,
	}

	for k := types.KindUnknown + 1; k < types.KindMarker; k++ {
		if exceptions[k] {
			continue
		}
		if _, exists := CodecRegistry[k]; !exists {
			t.Errorf("Missing codec in CodecRegistry for node kind: %s (%d)", k.String(), int(k))
		}
	}
}
