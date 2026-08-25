package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionIsDevelopmentBuild(t *testing.T) {
	if version == "" {
		t.Fatal("version must be set")
	}
}

func TestCLIEmitsJSONDiagnostics(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "gogo-test")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, out)
	}
	src := filepath.Join(t.TempDir(), "bad.gogo")
	if err := os.WriteFile(src, []byte("create variable x as @\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(bin, "-json", "-locale", "hi", src)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected diagnostic exit status, got success: %s", out)
	}
	text := string(out)
	for _, want := range []string{`"severity": "error"`, `"code": "G1000"`, "मैं इस अक्षर"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in %s", want, text)
		}
	}
}

func TestCLIInvalidUTF8JSONIncludesStableLocationAndFile(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "gogo-test")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, out)
	}
	src := filepath.Join(t.TempDir(), "bad-utf8.gogo")
	if err := os.WriteFile(src, []byte{'x', 0xff}, 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(bin, "-json", "-locale", "bn", src)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected diagnostic exit status, got success: %s", out)
	}
	var got []struct {
		Code     string `json:"code"`
		Language string `json:"language"`
		File     string `json:"file"`
		Span     struct {
			Start struct {
				Offset int `json:"offset"`
				Line   int `json:"line"`
				Column int `json:"column"`
			} `json:"start"`
		} `json:"span"`
	}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("invalid CLI JSON: %v\n%s", err, out)
	}
	if len(got) != 1 || got[0].Code != "G0002" || got[0].Language != "bn" || got[0].File != src || got[0].Span.Start.Offset != 0 || got[0].Span.Start.Line != 1 || got[0].Span.Start.Column != 1 {
		t.Fatalf("unexpected invalid UTF-8 diagnostic JSON: %#v", got)
	}
}
