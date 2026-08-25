package source

import "testing"

func TestValidatePath(t *testing.T) {
	for _, path := range []string{"main.gogo", "src/app.gogo", "বাংলা.gogo"} {
		if !ValidatePath(path) {
			t.Fatalf("expected valid path %q", path)
		}
	}
	for _, path := range []string{"", "   ", "bad\x00path.gogo"} {
		if ValidatePath(path) {
			t.Fatalf("expected invalid path %q", path)
		}
	}
}

func TestValidateFileEncodingAndMetadata(t *testing.T) {
	valid := File{ID: 1, Path: "main.gogo", Text: "create variable x as \"বাংলা\""}
	if !ValidateFile(valid) {
		t.Fatal("expected valid UTF-8 source file")
	}

	invalid := File{ID: 1, Path: "main.gogo", Text: string([]byte{0xff, 0xfe})}
	if ValidateFile(invalid) {
		t.Fatal("expected invalid UTF-8 file to be rejected")
	}
}
