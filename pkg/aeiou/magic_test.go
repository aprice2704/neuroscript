// NeuroScript Version: 0.3.0
// File version: 1
// Purpose: Adds unit tests for the aeiou.Wrap helper function.
// filename: neuroscript/pkg/aeiou/magic_test.go
// nlines: 48
// risk_rating: LOW

package aeiou

import (
	"encoding/json"
	"testing"
)

func TestWrap(t *testing.T) {
	testCases := []struct {
		name        string
		sectionType SectionType
		payload     interface{}
		want        string
		wantErr     bool
	}{
		{
			name:        "Simple no payload",
			sectionType: SectionStart,
			payload:     nil,
			want:        "<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:START>>>",
			wantErr:     false,
		},
		{
			name:        "With JSON payload",
			sectionType: SectionHeader,
			payload:     json.RawMessage(`{"v":2}`),
			want:        `<<<NSENVELOPE_MAGIC_9E3B6F2D:V2:HEADER:{"v":2}>>>`,
			wantErr:     false,
		},
		{
			name:        "With unmarshallable payload",
			sectionType: SectionHeader,
			payload:     make(chan int), // Cannot be marshaled
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Wrap(tc.sectionType, tc.payload)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Wrap() error = %v, wantErr %v", err, tc.wantErr)
			}
			if !tc.wantErr && got != tc.want {
				t.Errorf("Wrap() = %v, want %v", got, tc.want)
			}
		})
	}
}
