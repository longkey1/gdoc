package gdoc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadInputFromFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
	}{
		{name: "plain content", content: "hello, world\n"},
		{name: "empty file", content: ""},
		{name: "multiline markdown", content: "# Title\n\n- item\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "input.md")
			writeFile(t, path, tt.content)

			got, err := ReadInput(path)
			if err != nil {
				t.Fatalf("ReadInput() error = %v", err)
			}
			if got != tt.content {
				t.Errorf("ReadInput() = %q, want %q", got, tt.content)
			}
		})
	}
}

func TestReadInputMissingFile(t *testing.T) {
	t.Parallel()

	if _, err := ReadInput(filepath.Join(t.TempDir(), "nosuch.md")); err == nil {
		t.Error("ReadInput() error = nil, want error")
	}
}

// ReadInput reads os.Stdin for "" and "-", so these subtests swap the
// process stdin and must not run in parallel.
func TestReadInputFromStdin(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
	}{
		{name: "empty path", filePath: ""},
		{name: "dash path", filePath: "-"},
	}

	for _, tt := range tests {
		filePath := tt.filePath
		t.Run(tt.name, func(t *testing.T) {
			const content = "from stdin\n"

			path := filepath.Join(t.TempDir(), "stdin.txt")
			writeFile(t, path, content)
			f, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			orig := os.Stdin
			os.Stdin = f
			t.Cleanup(func() { os.Stdin = orig })

			got, err := ReadInput(filePath)
			if err != nil {
				t.Fatalf("ReadInput(%q) error = %v", filePath, err)
			}
			if got != content {
				t.Errorf("ReadInput(%q) = %q, want %q", filePath, got, content)
			}
		})
	}
}
