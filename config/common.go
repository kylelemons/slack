package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func loadFile[T any](filename string) (*T, error) {
	var out T

	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("load %T: %w", out, err) // filename already included by os package
	}

	// TODO: We can support more filetypes in the future if we want, like YAML,
	//       but for the time being let's stick with as many first party libraries as possible.
	switch ext := strings.ToLower(filepath.Ext(filename)); ext {
	case ".json", ".js":
		if err := json.NewDecoder(f).Decode(&out); err != nil {
			return nil, fmt.Errorf("parse JSON from %q as %T: %w", filename, out, err)
		}
	default:
		return nil, fmt.Errorf("unrecognized extension %q; want .json", ext)
	}
	return &out, err
}
