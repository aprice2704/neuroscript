// NeuroScript Version: 0.5.2
// File version: 6
// Purpose: Implements path-lite selectors with options for case-insensitivity.
// filename: pkg/json-lite/path.go
// nlines: 177
// risk_rating: MEDIUM

package json_lite

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	maxPathSegments   = 128
	maxPathSegmentLen = 256
)

var (
	pathRegex *regexp.Regexp
)

func init() {
	const keyPart = `[a-zA-Z0-9_-]+`
	const segmentPart = keyPart + `(?:\[\d+\])*`
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

// SelectOptions provides options for the Select function.
type SelectOptions struct {
	CaseInsensitive bool
}

// ParsePath compiles a path-lite string (e.g., "a.b[0].c") into a reusable Path structure.
func ParsePath(pathStr string) (Path, error) {
	if pathStr == "" {
		return nil, fmt.Errorf("%w: path string cannot be empty", ErrInvalidPath)
	}
	if !pathRegex.MatchString(pathStr) {
		return nil, fmt.Errorf("%w: path string has invalid format", ErrInvalidPath)
	}

	var segments []PathSegment
	reader := strings.NewReader(pathStr)
	var currentSegment strings.Builder

	for reader.Len() > 0 {
		if len(segments) >= maxPathSegments {
			return nil, fmt.Errorf("%w: path exceeds maximum number of segments (%d)", ErrNestingDepthExceeded, maxPathSegments)
		}

		char, _, _ := reader.ReadRune()
		if currentSegment.Len() > maxPathSegmentLen {
			return nil, fmt.Errorf("%w: path segment exceeds maximum length (%d)", ErrInvalidArgument, maxPathSegmentLen)
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
					return nil, fmt.Errorf("%w: path index segment exceeds maximum length (%d)", ErrInvalidArgument, maxPathSegmentLen)
				}
				idxChar, _, _ := reader.ReadRune()
				if idxChar == ']' {
					inIndex = false
				} else {
					indexStr.WriteRune(idxChar)
				}
			}
			if inIndex {
				return nil, fmt.Errorf("%w: unterminated index in path '%s'", ErrInvalidPath, pathStr)
			}
			idx, err := strconv.Atoi(indexStr.String())
			if err != nil {
				return nil, fmt.Errorf("%w: invalid list index '%s' in path: %v", ErrListInvalidIndexType, indexStr.String(), err)
			}
			segments = append(segments, PathSegment{Index: idx, IsKey: false})
		default:
			currentSegment.WriteRune(char)
		}
	}

	if currentSegment.Len() > 0 {
		if currentSegment.Len() > maxPathSegmentLen {
			return nil, fmt.Errorf("%w: path segment exceeds maximum length (%d)", ErrInvalidArgument, maxPathSegmentLen)
		}
		segments = append(segments, PathSegment{Key: currentSegment.String(), IsKey: true})
	}

	return segments, nil
}

// Select retrieves a value from a nested data structure using a pre-parsed Path.
func Select(value any, path Path, options *SelectOptions) (any, error) {
	current := value
	for i, seg := range path {
		if current == nil {
			return nil, fmt.Errorf("%w: cannot select from nil value at path segment %d", ErrCollectionIsNil, i)
		}
		if seg.IsKey {
			asMap, ok := current.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("%w: expected a map to access key '%s' (segment %d)", ErrCannotAccessType, seg.Key, i)
			}

			if options != nil && options.CaseInsensitive {
				var foundVal any
				foundKey := false
				lowerKey := strings.ToLower(seg.Key)
				for k, v := range asMap {
					if strings.ToLower(k) == lowerKey {
						foundVal = v
						foundKey = true
						break
					}
				}
				if !foundKey {
					return nil, fmt.Errorf("%w: key '%s' not found (segment %d)", ErrMapKeyNotFound, seg.Key, i)
				}
				current = foundVal
			} else {
				val, keyOk := asMap[seg.Key]
				if !keyOk {
					return nil, fmt.Errorf("%w: key '%s' not found (segment %d)", ErrMapKeyNotFound, seg.Key, i)
				}
				current = val
			}

		} else { // Is Index
			asList, ok := current.([]any)
			if !ok {
				return nil, fmt.Errorf("%w: expected a list to access index %d (segment %d)", ErrCannotAccessType, seg.Index, i)
			}
			if seg.Index < 0 || seg.Index >= len(asList) {
				return nil, fmt.Errorf("%w: index %d is out of bounds for list of length %d (segment %d)", ErrListIndexOutOfBounds, seg.Index, len(asList), i)
			}
			current = asList[seg.Index]
		}
	}
	return current, nil
}
