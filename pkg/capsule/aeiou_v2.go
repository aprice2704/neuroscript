package capsule

import _ "embed"

// aeiouV1Markdown is embedded from a separate file to avoid backtick escaping issues.
//
//go:embed aeiou_spec_v2.md
var aeiouV1Markdown string

func init() {
	MustRegister(Capsule{
		ID:       "capsule/aeiou/2",
		Version:  "2",
		MIME:     "text/markdown; charset=utf-8",
		Priority: 10,
		Content:  aeiouV1Markdown,
	})
}
