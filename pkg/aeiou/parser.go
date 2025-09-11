// NeuroScript Version: 0.7.0
// File version: 7
// Purpose: Corrected a regression where whitespace around markers was not being handled, by re-introducing trimming before regex matching.
// filename: aeiou/parser.go
// nlines: 175
// risk_rating: HIGH

package aeiou

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

const (
	// MaxSectionSize is the maximum size for any single section (512 KiB).
	MaxSectionSize = 512 * 1024
	// MaxEnvelopeSize is the maximum total size for the envelope content (1 MiB).
	MaxEnvelopeSize = 1024 * 1024
)

var markerRegex = regexp.MustCompile(`^<<<NSENV:V3:(START|USERDATA|SCRATCHPAD|OUTPUT|ACTIONS|END)>>>$`)

// Parse reads from an io.Reader, validates the AEIOU v3 envelope structure,
// and returns a populated Envelope struct and any non-fatal lints.
func Parse(r io.Reader) (*Envelope, []Lint, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), MaxEnvelopeSize+1024) // Buffer large enough for whole envelope
	return parseWithScanner(scanner)
}

// parseWithScanner contains the core parsing logic, operating on a pre-configured scanner.
func parseWithScanner(scanner *bufio.Scanner) (*Envelope, []Lint, error) {
	env := &Envelope{}
	var lints []Lint
	var currentSection *string
	var seenSections = make(map[SectionType]bool)
	var sectionOrder []SectionType
	var totalBytes int
	var sectionBytes int
	var firstLineInSection bool

	// Expect START marker first, allowing for surrounding whitespace
	if !scanner.Scan() {
		return nil, nil, fmt.Errorf("%w: expected START marker, found EOF", ErrMarkerInvalid)
	}
	trimmedFirstLine := strings.TrimSpace(scanner.Text())
	match := markerRegex.FindStringSubmatch(trimmedFirstLine)
	if len(match) < 2 || SectionType(match[1]) != SectionStart {
		return nil, nil, fmt.Errorf("%w: expected START marker, found: %s", ErrMarkerInvalid, scanner.Text())
	}
	totalBytes += len(scanner.Bytes())
	seenSections[SectionStart] = true
	sectionOrder = append(sectionOrder, SectionStart)

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		lineBytes := len(scanner.Bytes())
		totalBytes += lineBytes + 1 // +1 for the newline

		if totalBytes > MaxEnvelopeSize {
			return nil, nil, ErrPayloadTooLarge
		}

		match := markerRegex.FindStringSubmatch(trimmedLine)
		if len(match) > 1 { // It's a marker
			// When a new section starts, trim the content of the previous one.
			if currentSection != nil {
				*currentSection = strings.TrimRight(*currentSection, " \t\r\n")
			}

			sectionBytes = 0 // Reset section counter on new marker
			firstLineInSection = true
			section := SectionType(match[1])

			// END marker terminates parsing
			if section == SectionEnd {
				finalEnv, err := validateAndFinalize(env, seenSections, sectionOrder)
				return finalEnv, lints, err
			}

			if seenSections[section] {
				currentSection = nil // Ignore content of duplicate sections
				lints = append(lints, Lint{
					Code:    LintCodeDuplicateSection,
					Message: fmt.Sprintf("duplicate section '%s' ignored (first instance is used)", section),
				})
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
			default: // Should be unreachable unless a START marker is duplicated
				return nil, lints, fmt.Errorf("%w: unknown or misplaced section type: %s", ErrMarkerInvalid, section)
			}
		} else { // It's content
			sectionBytes += lineBytes + 1
			if sectionBytes > MaxSectionSize {
				return nil, lints, ErrPayloadTooLarge
			}
			if currentSection != nil {
				if !firstLineInSection {
					*currentSection += "\n"
				}
				*currentSection += line
				firstLineInSection = false
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, lints, fmt.Errorf("error scanning envelope: %w", err)
	}

	// If we reach here, the END marker was missing. Before returning, trim the last section.
	if currentSection != nil {
		*currentSection = strings.TrimRight(*currentSection, " \t\r\n")
	}
	return nil, lints, fmt.Errorf("%w: missing END marker", ErrMarkerInvalid)
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
		// Skip START since it's handled pre-loop
		if section == SectionStart {
			continue
		}
		currentIdx, ok := orderMap[section]
		if !ok {
			continue
		}
		if currentIdx < lastIdx {
			return nil, fmt.Errorf("%w: section '%s' appeared out of order", ErrSectionOrder, section)
		}
		lastIdx = currentIdx
	}

	// Required Sections Check
	if !seen[SectionUserData] {
		return nil, fmt.Errorf("%w: required section USERDATA not found", ErrSectionMissing)
	}
	if !seen[SectionActions] {
		return nil, fmt.Errorf("%w: required section ACTIONS not found", ErrSectionMissing)
	}

	return env, nil
}
