package main

import (
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
