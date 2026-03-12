// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 3
// :: description: Updated ComputeHostDigest to remove legacy V3 token dependencies.
// :: latestChange: Hardcoded V3 marker prefixes for legacy stripping to allow deletion of magic.go.
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
	normalizedOutput := normalizeAndStrip(output)
	normalizedScratchpad := normalizeAndStrip(scratchpad)

	digestInput := fmt.Sprintf("OUT|%s\nSCR|%s", normalizedOutput, normalizedScratchpad)

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

		// Strip V3 control tokens (legacy hardcoded to prevent import cycle post-cleanup)
		if strings.HasPrefix(line, "<<<NSMAG:V3") && strings.HasSuffix(line, ">>>") {
			continue
		}

		// Strip V4 control marker
		if strings.HasPrefix(trimmed, "<<<LOOP:DONE>>>") {
			continue
		}

		sb.WriteString(strings.TrimRight(line, " \t"))
		sb.WriteString("\n")
	}
	return sb.String()
}
