// :: product: FDM/NS
// :: majorVersion: 0
// :: fileVersion: 2
// :: description: Updated ComputeHostDigest to strip V4 control markers (<<<LOOP:DONE>>>) as required by spec.
// :: latestChange: Added logic to strip <<<LOOP:DONE>>> from digest calculation.
// :: filename: pkg/aeiou/progress.go
// :: serialization: go
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
		trimmed := strings.TrimSpace(line)

		// Strip V3 control tokens
		if strings.HasPrefix(line, TokenMarkerPrefix) && strings.HasSuffix(line, TokenMarkerSuffix) {
			continue
		}

		// Strip V4 control marker
		// Note: The V4 spec says "Strip control markers". <<<LOOP:DONE>>> is the only one.
		if strings.HasPrefix(trimmed, "<<<LOOP:DONE>>>") {
			continue
		}

		// Normalize by trimming trailing space and appending with a consistent newline
		sb.WriteString(strings.TrimRight(line, " \t"))
		sb.WriteString("\n")
	}
	return sb.String()
}
