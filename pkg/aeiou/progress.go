// NeuroScript Version: 0.7.0
// File version: 1
// Purpose: Implements the mandatory host-side progress guard digest calculation.
// filename: aeiou/progress.go
// nlines: 41
// risk_rating: LOW

package aeiou

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// ComputeHostDigest calculates a stable digest of the host-observed outputs
// for a turn, as required by the mandatory progress guard. It normalizes the
// content and strips control tokens to prevent loops based on varying token details.
func ComputeHostDigest(output, scratchpad string) string {
	// Normalize and strip tokens from both inputs
	normalizedOutput := normalizeAndStrip(output)
	normalizedScratchpad := normalizeAndStrip(scratchpad)

	// Construct the canonical string to be hashed
	digestInput := fmt.Sprintf("OUT|%s\nSCR|%s", normalizedOutput, normalizedScratchpad)

	// Compute the digest
	hash := sha256.Sum256([]byte(digestInput))
	return hex.EncodeToString(hash[:])
}

// normalizeAndStrip processes a multi-line string by removing control tokens,
// trimming trailing whitespace from each line, and joining them with '\n'.
func normalizeAndStrip(input string) string {
	var sb strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(input))
	for scanner.Scan() {
		line := scanner.Text()
		// Strip control tokens
		if strings.HasPrefix(line, TokenMarkerPrefix) && strings.HasSuffix(line, TokenMarkerSuffix) {
			continue
		}
		// Normalize by trimming trailing space and appending with a consistent newline
		sb.WriteString(strings.TrimRight(line, " \t"))
		sb.WriteString("\n")
	}
	return sb.String()
}
