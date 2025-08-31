// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Corrects a copy-paste error in the Wrap function's format string.
// filename: aeiou/markers.go
// nlines: 16
// risk_rating: LOW

package aeiou

import "fmt"

const (
	markerPrefix = "<<<NSENV:V3"
	markerSuffix = ">>>"
)

// Wrap formats a string according to the NeuroScript V3 envelope protocol.
func Wrap(sectionType SectionType) string {
	return fmt.Sprintf("%s:%s%s", markerPrefix, sectionType, markerSuffix)
}
