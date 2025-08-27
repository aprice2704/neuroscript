package capsule

import _ "embed"

// askloopV1Markdown is embedded to avoid escaping issues.
//
//go:embed askloop_spec_v2.md
var askloopV1Markdown string

func init() {
	MustRegister(Capsule{
		ID:       "capsule/askloop/2",
		Version:  "2",
		MIME:     "text/markdown; charset=utf-8",
		Priority: 11, // after aeiou/1 so it appears nearby in List()
		Content:  askloopV1Markdown,
	})
}
