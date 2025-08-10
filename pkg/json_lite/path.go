// NeuroScript Version: 0.5.2
// File version: 5
// Purpose: Implements path-lite selectors with corrected regex and length limits.
// filename: pkg/json-lite/path.go
// nlines: 153
// risk_rating: MEDIUM

package json_lite

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

const (
	maxPathSegments   = 128
	maxPathSegmentLen = 256
)

var (
	pathRegex *regexp.Regexp
)

func init() {
	// A segment key allows letters, numbers, underscore, and hyphen.
	const keyPart = `[a-zA-Z0-9_-]+`
	// A full segment is a key part followed by optional indices.
	const segmentPart = keyPart + `(?:\[\d+\])*`
	// A valid path is one or more segments separated by dots.
	const pathPattern = `^` + segmentPart + `(?:\.` + segmentPart + `)*$`
	pathRegex = regexp.MustCompile(pathPattern)
}

// PathSegment represents one component of a parsed path, either a map key or list index.
type PathSegment struct {
	Key   string
	Index int
	IsKey bool
}

// Path is the parsed, executable representation of a path-lite string.
type Path []PathSegment

// ParsePath compiles a path-lite string (e.g., "a.b[0].c") into a reusable Path structure.
func ParsePath(pathStr string) (Path, error) {
	if pathStr == "" {
		return nil, fmt.Errorf("%w: path string cannot be empty", lang.ErrInvalidPath)
	}
	if !pathRegex.MatchString(pathStr) {
		return nil, fmt.Errorf("%w: path string has invalid format", lang.ErrInvalidPath)
	}

	var segments []PathSegment
	reader := strings.NewReader(pathStr)
	var currentSegment strings.Builder

	for reader.Len() > 0 {
		if len(segments) >= maxPathSegments {
			return nil, fmt.Errorf("%w: path exceeds maximum number of segments (%d)", lang.ErrNestingDepthExceeded, maxPathSegments)
		}

		char, _, _ := reader.ReadRune()
		if currentSegment.Len() > maxPathSegmentLen {
			return nil, fmt.Errorf("%w: path segment exceeds maximum length (%d)", lang.ErrInvalidArgument, maxPathSegmentLen)
		}

		switch char {
		case '.':
			if currentSegment.Len() > 0 {
				segments = append(segments, PathSegment{Key: currentSegment.String(), IsKey: true})
				currentSegment.Reset()
			}
		case '[':
			if currentSegment.Len() > 0 {
				segments = append(segments, PathSegment{Key: currentSegment.String(), IsKey: true})
				currentSegment.Reset()
			}
			var indexStr strings.Builder
			inIndex := true
			for reader.Len() > 0 && inIndex {
				if indexStr.Len() >= maxPathSegmentLen {
					return nil, fmt.Errorf("%w: path index segment exceeds maximum length (%d)", lang.ErrInvalidArgument, maxPathSegmentLen)
				}
				idxChar, _, _ := reader.ReadRune()
				if idxChar == ']' {
					inIndex = false
				} else {
					indexStr.WriteRune(idxChar)
				}
			}
			if inIndex {
				return nil, fmt.Errorf("%w: unterminated index in path '%s'", lang.ErrInvalidPath, pathStr)
			}
			idx, err := strconv.Atoi(indexStr.String())
			if err != nil {
				return nil, fmt.Errorf("%w: invalid list index '%s' in path: %v", lang.ErrListInvalidIndexType, indexStr.String(), err)
			}
			segments = append(segments, PathSegment{Index: idx, IsKey: false})
		default:
			currentSegment.WriteRune(char)
		}
	}

	// Final check for the last segment
	if currentSegment.Len() > 0 {
		if currentSegment.Len() > maxPathSegmentLen {
			return nil, fmt.Errorf("%w: path segment exceeds maximum length (%d)", lang.ErrInvalidArgument, maxPathSegmentLen)
		}
		segments = append(segments, PathSegment{Key: currentSegment.String(), IsKey: true})
	}

	return segments, nil
}

// Select function remains unchanged
func Select(value any, path Path) (any, error) {
	current := value
	for i, seg := range path {
		if current == nil {
			return nil, fmt.Errorf("%w: cannot select from nil value at path segment %d", lang.ErrCollectionIsNil, i)
		}
		if seg.IsKey {
			asMap, ok := current.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("%w: expected a map to access key '%s' (segment %d)", lang.ErrCannotAccessType, seg.Key, i)
			}
			current, ok = asMap[seg.Key]
			if !ok {
				return nil, fmt.Errorf("%w: key '%s' not found (segment %d)", lang.ErrMapKeyNotFound, seg.Key, i)
			}
		} else {
			asList, ok := current.([]any)
			if !ok {
				return nil, fmt.Errorf("%w: expected a list to access index %d (segment %d)", lang.ErrCannotAccessType, seg.Index, i)
			}
			if seg.Index < 0 || seg.Index >= len(asList) {
				return nil, fmt.Errorf("%w: index %d is out of bounds for list of length %d (segment %d)", lang.ErrListIndexOutOfBounds, seg.Index, len(asList), i)
			}
			current = asList[seg.Index]
		}
	}
	return current, nil
}
