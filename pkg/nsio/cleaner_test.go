package nsio

import (
	"bytes"
	"testing"
)

func TestCleanNS(t *testing.T) {
	const mb = 1 << 20

	tests := []struct {
		name      string
		input     []byte
		max       int
		want      string // empty means we expect an error
		wantError bool
	}{
		{
			name:  "plain ascii ok",
			input: []byte("func main() means\n\tset x = 1\nendfunc\n"),
			max:   mb,
			want:  "func main() means\n\tset x = 1\nendfunc\n",
		},
		{
			name:  "strip controls <0x20 except TAB/LF",
			input: []byte("foo\x00bar\x07\tbaz\n"),
			max:   mb,
			want:  "foobar\tbaz\n",
		},
		{
			name: "BOM and CRLF normalisation",
			input: append([]byte{0xEF, 0xBB, 0xBF},
				[]byte("line1\r\nline2\r\n")...),
			max:  mb,
			want: "line1\nline2\n",
		},
		{
			name:      "reject invalid utf‑8",
			input:     []byte{0xff, 0xfe, 0xfd},
			max:       mb,
			wantError: true,
		},
		{
			name:      "reject bidi override U+202E",
			input:     []byte("a\u202Efb\n"),
			max:       mb,
			wantError: true,
		},
		{
			name:      "reject zero‑width space U+200B",
			input:     []byte("a\u200B\n"),
			max:       mb,
			wantError: true,
		},
		{
			name:      "size limit exceeded",
			input:     bytes.Repeat([]byte("A"), 1024),
			max:       512,
			wantError: true,
		},
	}

	for _, tc := range tests {
		tc := tc // capture
		t.Run(tc.name, func(t *testing.T) {
			got, err := CleanNS(bytes.NewReader(tc.input), tc.max)
			if tc.wantError {
				if err == nil {
					t.Fatalf("expected error, got nil (cleaned = %q)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tc.want {
				t.Fatalf("mismatch:\nwant %q\ngot  %q", tc.want, got)
			}
		})
	}
}
