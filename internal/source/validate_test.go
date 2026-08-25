package source

import "testing"

func TestValidateUTF8(t *testing.T) {
	if !ValidateUTF8("hello বাংলা हिन्दी") {
		t.Fatal("expected valid UTF-8")
	}
	if ValidateUTF8(string([]byte{0xff, 0xfe})) {
		t.Fatal("expected invalid UTF-8")
	}
}

func TestValidateFile(t *testing.T) {
	valid := File{ID: 1, Path: "main.gogo", Text: "create variable user as \"Alex\""}
	if !ValidateFile(valid) {
		t.Fatal("expected valid source file")
	}
	if ValidateFile(File{ID: 0, Path: "main.gogo", Text: "ok"}) {
		t.Fatal("expected zero ID to be invalid")
	}
	if ValidateFile(File{ID: 1, Path: "", Text: "ok"}) {
		t.Fatal("expected empty path to be invalid")
	}
}
