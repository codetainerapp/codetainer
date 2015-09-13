package codetainer

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/opencontainers/specs"
)

// loadSpec loads the specification from the provided path.
// If the path is empty then the default path will be "config.json"
func loadSpec(path string) (*specs.LinuxRuntimeSpec, error) {
	if path == "" {
		path = "config.json"
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("JSON specification file for %s not found", path)
		}
		return nil, err
	}
	defer f.Close()
	var s *specs.LinuxRuntimeSpec
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		return nil, err
	}
	return s, nil
}
