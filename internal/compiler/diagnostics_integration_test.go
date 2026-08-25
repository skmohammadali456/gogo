package compiler

import "testing"

func TestSessionDiagnosticsCarryFileIDsFromLexerAndParser(t *testing.T) {
	s := NewSession()
	first := s.AddFile("first.gogo", "create variable ok as 1")
	second := s.AddFile("second.gogo", "create variable bad as @")
	if first == 0 || second == 0 {
		t.Fatal("expected files to be added")
	}
	_, _ = s.ParseFile(second)
	diags := s.Diagnostics.All()
	if len(diags) == 0 {
		t.Fatal("expected diagnostics")
	}
	for _, d := range diags {
		if d.FileID != second {
			t.Fatalf("diagnostic %s file id = %d, want %d", d.Code, d.FileID, second)
		}
		for _, label := range d.Labels {
			if label.FileID != second {
				t.Fatalf("label file id = %d, want %d", label.FileID, second)
			}
		}
	}
}
