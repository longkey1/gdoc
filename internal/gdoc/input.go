package gdoc

import (
	"fmt"
	"io"
	"os"
)

// ReadInput reads content from the given file path, or from stdin if filePath is empty or "-".
func ReadInput(filePath string) (string, error) {
	var r io.Reader
	if filePath == "" || filePath == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(filePath)
		if err != nil {
			return "", fmt.Errorf("unable to open input file: %v", err)
		}
		defer f.Close()
		r = f
	}

	b, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("unable to read input: %v", err)
	}
	return string(b), nil
}
