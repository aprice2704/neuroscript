// NeuroScript Version: 0.3.0
// File version: 1
// Purpose: Entry point for ns-lint CLI (metadata/policy/effects checks).
// filename: cmd/ns-lint/main.go
// nlines: 49
// risk_rating: MEDIUM

package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	fail := flag.Bool("fail", false, "exit nonzero if any ERROR findings are produced")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: ns-lint [--fail] <files-or-dirs>")
		os.Exit(2)
	}

	paths := flag.Args()
	findings, err := Run(paths)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ns-lint error:", err)
		os.Exit(2)
	}

	PrintFindings(findings)

	if *fail && findings.HasErrors() {
		os.Exit(1)
	}
}
