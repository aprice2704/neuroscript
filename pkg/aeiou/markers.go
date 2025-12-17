// :: product: FDM/NS
// :: majorVersion: 0
// :: fileVersion: 3
// :: description: Defines the AEIOU V4 envelope markers.
// :: latestChange: Updated markers to V4 to match spec and fix test failures.
// :: filename: pkg/aeiou/markers.go
// :: serialization: go
package aeiou

import "fmt"

const (
	markerPrefix = "<<<NSENV:V4"
	markerSuffix = ">>>"
)

// Wrap formats a string according to the NeuroScript V4 envelope protocol.
func Wrap(sectionType SectionType) string {
	return fmt.Sprintf("%s:%s%s", markerPrefix, sectionType, markerSuffix)
}
