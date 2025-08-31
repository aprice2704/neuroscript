// NeuroScript Version: 0.7.0
// File version: 3
// Purpose: Refactors the parser for better testability with large inputs.
// filename: aeiou/parser.go
// nlines: 128
// risk_rating: HIGH

package aeiou

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

const (
	// MaxSectionSize is the maximum size for any single section (512 KiB).
	MaxSectionSize = 512 * 1024
	// MaxEnvelopeSize is the maximum total size for the envelope content (1 MiB).
	MaxEnvelopeSize = 1024 * 1024
)

// Parse reads from an io.Reader, validates the AEIOU v3 envelope structure,
// and returns a populated Envelope struct.
func Parse(r io.Reader) (*Envelope, error) {
	scanner := bufio.NewScanner(r)
	// Set a buffer large enough for standard lines but not excessively large.
	// Tests for huge lines will need to create their own scanner.
	scanner.Buffer(make([]byte, 0, 64*1024), MaxSectionSize+1024)
	return parseWithScanner(scanner)
}

// parseWithScanner contains the core parsing logic, operating on a pre-configured scanner.
// This allows for dependency injection of a scanner for testing purposes.
func parseWithScanner(scanner *bufio.Scanner) (*Envelope, error) {
	env := &Envelope{}
	var currentSection *string
	var seenSections = make(map[SectionType]bool)
	var sectionOrder []SectionType
	var totalBytes int
	var sectionBytes int

	// Expect START marker first
	if !scanner.Scan() || scanner.Text() != Wrap(SectionStart) {
		return nil, ErrMarkerInvalid
	}
	totalBytes += len(scanner.Bytes())

	for scanner.Scan() {
		line := scanner.Text()
		lineBytes := len(scanner.Bytes())
		totalBytes += lineBytes + 1 // +1 for the newline

		if totalBytes > MaxEnvelopeSize {
			return nil, ErrPayloadTooLarge
		}

		if strings.HasPrefix(line, markerPrefix) && strings.HasSuffix(line, markerSuffix) {
			sectionBytes = 0 // Reset section counter on new marker
			parts := strings.Split(strings.Trim(line, "<>"), ":")
			if len(parts) != 3 {
				return nil, fmt.Errorf("%w: %s", ErrMarkerInvalid, line)
			}
			section := SectionType(parts[2])

			// END marker terminates parsing
			if section == SectionEnd {
				return validateAndFinalize(env, seenSections, sectionOrder)
			}

			if seenSections[section] {
				currentSection = nil // Ignore content of duplicate sections
				continue
			}
			seenSections[section] = true
			sectionOrder = append(sectionOrder, section)

			switch section {
			case SectionUserData:
				currentSection = &env.UserData
			case SectionScratchpad:
				currentSection = &env.Scratchpad
			case SectionOutput:
				currentSection = &env.Output
			case SectionActions:
				currentSection = &env.Actions
			default:
				return nil, fmt.Errorf("unknown section type: %s", section)
			}
		} else {
			sectionBytes += lineBytes + 1
			if sectionBytes > MaxSectionSize {
				return nil, ErrPayloadTooLarge
			}
			if currentSection != nil {
				*currentSection += line + "\n"
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning envelope: %w", err)
	}

	// If we reach here, the END marker was missing
	return nil, ErrMarkerInvalid
}

func validateAndFinalize(env *Envelope, seen map[SectionType]bool, order []SectionType) (*Envelope, error) {
	// Canonical Order Check
	expectedOrder := []SectionType{SectionUserData, SectionScratchpad, SectionOutput, SectionActions}
	orderMap := make(map[SectionType]int)
	for i, s := range expectedOrder {
		orderMap[s] = i
	}
	lastIdx := -1
	for _, section := range order {
		currentIdx, ok := orderMap[section]
		if !ok {
			continue
		}
		if currentIdx < lastIdx {
			return nil, ErrSectionOrder
		}
		lastIdx = currentIdx
	}

	// Required Sections Check
	if !seen[SectionUserData] || !seen[SectionActions] {
		return nil, ErrSectionMissing
	}

	// Trim trailing newlines from content
	env.UserData = strings.TrimSuffix(env.UserData, "\n")
	env.Scratchpad = strings.TrimSuffix(env.Scratchpad, "\n")
	env.Output = strings.TrimSuffix(env.Output, "\n")
	env.Actions = strings.TrimSuffix(env.Actions, "\n")

	return env, nil
}
