package compiler

import "testing"

func TestDiagnosticsAfterLexerAndParserRecoveryCases(t *testing.T) {
	cases := map[string]string{
		"unterminated-string":       "create variable x as \"oops\ncreate variable y as 1",
		"invalid-escape":            "create variable x as \"\\q\"\ncreate variable y as 1",
		"unterminated-comment":      "create variable x as 1 /* open",
		"invalid-character":         "create variable x as @\ncreate variable y as 1",
		"unsupported-unicode-digit": "create variable x as ৯",
		"missing-expression":        "create variable x as ;\ncreate variable y as 1",
		"missing-delimiter":         "create function f(a { return 1",
		"unexpected-token":          "} create variable y as 1",
		"nested-malformed":          "create variable x as [ { bad: , ok: 1 }",
	}
	for name, text := range cases {
		t.Run(name, func(t *testing.T) {
			s := NewSession()
			id := s.AddFile(name+".gogo", text)
			if id == 0 {
				t.Fatal("file was not added")
			}
			_, _ = s.ParseFile(id)
			diags := s.Diagnostics.All()
			if len(diags) == 0 {
				t.Fatal("expected diagnostics")
			}
			if len(diags) > 8 {
				t.Fatalf("unexpected diagnostic cascade (%d): %#v", len(diags), diags)
			}
			for _, d := range diags {
				if d.FileID != id {
					t.Fatalf("diagnostic %s file id = %d, want %d", d.Code, d.FileID, id)
				}
				if d.Code == "" {
					t.Fatal("diagnostic missing stable code")
				}
				if !d.Span.IsValid() {
					t.Fatalf("diagnostic %s has invalid span: %#v", d.Code, d.Span)
				}
			}
		})
	}
}
