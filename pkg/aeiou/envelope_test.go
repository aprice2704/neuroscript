// :: product: FDM/NS
// :: majorVersion: 0
// :: fileVersion: 3
// :: description: Updated envelope tests to assert V4 markers.
// :: latestChange: Updated expectations to NSENV:V4.
// :: filename: pkg/aeiou/envelope_test.go
// :: serialization: go
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
			want:        "<<<NSENV:V4:START>>>",
		},
		{
			name:        "UserData Marker",
			sectionType: SectionUserData,
			want:        "<<<NSENV:V4:USERDATA>>>",
		},
		{
			name:        "End Marker",
			sectionType: SectionEnd,
			want:        "<<<NSENV:V4:END>>>",
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
