// Copyright 2022 Kyle Lemons
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
