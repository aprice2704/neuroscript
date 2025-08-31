// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Corrects test cases to align with the fixed Wrap function.
// filename: aeiou/envelope_test.go
// nlines: 32
// risk_rating: LOW

package aeiou

import "testing"

func TestWrap(t *testing.T) {
	testCases := []struct {
		name        string
		sectionType SectionType
		want        string
	}{
		{
			name:        "Start Marker",
			sectionType: SectionStart,
			want:        "<<<NSENV:V3:START>>>",
		},
		{
			name:        "UserData Marker",
			sectionType: SectionUserData,
			want:        "<<<NSENV:V3:USERDATA>>>",
		},
		{
			name:        "End Marker",
			sectionType: SectionEnd,
			want:        "<<<NSENV:V3:END>>>",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Wrap(tc.sectionType)
			if got != tc.want {
				t.Errorf("Wrap() = %q, want %q", got, tc.want)
			}
		})
	}
}
