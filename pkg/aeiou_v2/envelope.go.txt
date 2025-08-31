// NeuroScript Version: 0.3.0
// File version: 25
// Purpose: Fixes the fuzzy parsing bug by ensuring only complete, valid magic markers are processed.
// filename: neuroscript/pkg/aeiou/envelope.go
// nlines: 242
// risk_rating: MEDIUM

package aeiou

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/nsio"
)

// --- Sentinel Errors ---
var (
	ErrInvalidJSONHeader        = errors.New("invalid JSON header")
	ErrInvalidMagicMarker       = errors.New("invalid magic marker format")
	ErrDuplicateSection         = errors.New("duplicate section found")
	ErrUnknownSection           = errors.New("unknown section type")
	ErrContentValidation        = errors.New("content validation failed")
	ErrChecksumMismatch         = errors.New("checksum mismatch")
	ErrContentBeforeFirstHeader = errors.New("content found before first section header")
	ErrEnvelopeNotFound         = errors.New("no envelope found")
	ErrEnvelopeNoStart          = errors.New("envelope START marker not found")
	ErrEnvelopeNoEnd            = errors.New("envelope END marker not found after START")
)

// SectionType defines the valid sections in an AEIOU envelope.
type SectionType string

// V2 Section Types
const (
	SectionStart           SectionType = "START"
	SectionEnd             SectionType = "END"
	SectionHeader          SectionType = "HEADER"
	SectionActions         SectionType = "ACTIONS"
	SectionEvents          SectionType = "EVENTS"
	SectionImplementations SectionType = "IMPLEMENTATIONS"
	SectionOrchestration   SectionType = "ORCHESTRATION"
	SectionUserData        SectionType = "USERDATA"
	SectionDiagnostic      SectionType = "DIAGNOSTIC"
	SectionLoop            SectionType = "LOOP"
	SectionHalt            SectionType = "HALT"
)

// Header contains metadata about the envelope and the agent that sent it.
type Header struct {
	Proto     string            `json:"proto"`
	Version   int               `json:"v"`
	AgentDID  string            `json:"agent_did,omitempty"`
	Caps      []string          `json:"caps,omitempty"`
	Budgets   map[string]int    `json:"budgets,omitempty"`
	Hashes    map[string]string `json:"hashes,omitempty"`
	Checksum  string            `json:"checksum,omitempty"`
	Signature string            `json:"signature,omitempty"`
}

// Envelope holds the parsed header and content of the five AEIOU sections.
type Envelope struct {
	Header          *Header
	Actions         string
	Events          string
	Implementations string
	Orchestration   string
	UserData        string
}

// extractEnvelope finds the first complete START...END block.
func extractEnvelope(payload string) (string, error) {
	startMarker, _ := Wrap(SectionStart, nil)
	endMarker, _ := Wrap(SectionEnd, nil)

	startIndex := strings.Index(payload, startMarker)
	if startIndex == -1 {
		return "", ErrEnvelopeNoStart
	}

	// Search for the end marker *after* the start marker
	endIndex := strings.Index(payload[startIndex:], endMarker)
	if endIndex == -1 {
		return "", ErrEnvelopeNoEnd
	}

	// The content is between the end of the start marker and the beginning of the end marker.
	contentStart := startIndex + len(startMarker)
	contentEnd := startIndex + endIndex
	return payload[contentStart:contentEnd], nil
}

// RobustParse finds the first complete AEIOU envelope in a string, extracts it, and parses it.
func RobustParse(payload string) (*Envelope, error) {
	trimmedPayload := strings.TrimSpace(payload)
	if trimmedPayload == "" {
		return nil, ErrEnvelopeNotFound
	}

	envelopeContent, err := extractEnvelope(trimmedPayload)
	if err != nil {
		return nil, err
	}

	return Parse(envelopeContent)
}

// Parse takes a clean string containing exactly one envelope's content and parses it.
func Parse(payload string) (*Envelope, error) {
	cleanedBytes, err := nsio.CleanNS(bytes.NewReader([]byte(payload)), 10*1024*1024)
	if err != nil {
		return nil, fmt.Errorf("envelope cleaning failed: %w", err)
	}

	env := &Envelope{}
	scanner := bufio.NewScanner(bytes.NewReader(cleanedBytes))
	var currentSection *string
	var seenSections = make(map[SectionType]bool)

	for scanner.Scan() {
		line := scanner.Text()

		// ** THE FIX IS HERE: Only treat a line as a marker if it is fully formed. **
		if strings.HasPrefix(line, "<<<") && strings.HasSuffix(line, ">>>") {
			parts := strings.Split(strings.Trim(line, "<>"), ":")
			if len(parts) < 3 {
				// Malformed marker, treat as content
				if currentSection != nil {
					*currentSection += line + "\n"
				}
				continue
			}
			section := SectionType(parts[2])

			if seenSections[section] && section != SectionLoop && section != SectionDiagnostic && section != SectionHalt {
				return nil, fmt.Errorf("%w: %s", ErrDuplicateSection, section)
			}

			switch section {
			case SectionHeader:
				if len(parts) > 3 {
					jsonPayload := strings.Join(parts[3:], ":")
					if err := json.Unmarshal([]byte(jsonPayload), &env.Header); err != nil {
						return nil, fmt.Errorf("%w: %v", ErrInvalidJSONHeader, err)
					}
				}
			case SectionActions:
				currentSection = &env.Actions
			case SectionEvents:
				currentSection = &env.Events
			case SectionImplementations:
				currentSection = &env.Implementations
			case SectionOrchestration:
				currentSection = &env.Orchestration
			case SectionUserData:
				currentSection = &env.UserData
			default:
				// Ignore other valid marker types
			}
			seenSections[section] = true
		} else {
			// Not a valid marker, so it's content.
			if currentSection != nil {
				*currentSection += line + "\n"
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning envelope payload: %w", err)
	}

	trimSections(env)
	return env, nil
}

// trimSections removes trailing whitespace from content.
func trimSections(env *Envelope) {
	env.Actions = strings.TrimSpace(env.Actions)
	env.Events = strings.TrimSpace(env.Events)
	env.Implementations = strings.TrimSpace(env.Implementations)
	env.Orchestration = strings.TrimSpace(env.Orchestration)
	env.UserData = strings.TrimSpace(env.UserData)
}

// Compose constructs the full envelope string from an Envelope struct.
func (e *Envelope) Compose() (string, error) {
	var contentBuilder strings.Builder
	appendSection := func(sectionType SectionType, content string) {
		if content != "" {
			marker, _ := Wrap(sectionType, nil)
			contentBuilder.WriteString(marker)
			contentBuilder.WriteString("\n")
			contentBuilder.WriteString(content)
			contentBuilder.WriteString("\n")
		}
	}
	appendSection(SectionActions, e.Actions)
	appendSection(SectionEvents, e.Events)
	appendSection(SectionImplementations, e.Implementations)
	appendSection(SectionOrchestration, e.Orchestration)
	appendSection(SectionUserData, e.UserData)
	contentString := contentBuilder.String()

	hash := sha256.Sum256([]byte(contentString))
	checksum := hex.EncodeToString(hash[:])
	if e.Header == nil {
		e.Header = &Header{Proto: "NSENVELOPE", Version: 2}
	}
	e.Header.Checksum = checksum

	header, _ := Wrap(SectionHeader, e.Header)
	start, _ := Wrap(SectionStart, nil)
	end, _ := Wrap(SectionEnd, nil)

	return fmt.Sprintf("%s\n%s\n%s%s\n", start, header, contentString, end), nil
}

// Validate checks the envelope for content and checksum integrity.
func (e *Envelope) Validate() []error {
	var errs []error
	log.Printf("[DEBUG] V2 validation is placeholder.")
	return errs
}
