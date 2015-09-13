package codetainer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	docker "github.com/fsouza/go-dockerclient"
)

var defaultSpec string = `
{
  "Config": {
    "NetworkDisabled": false 
  },
  "HostConfig": {
    "Privileged": false,
    "ReadonlyRootfs": false,
	"Ulimits": [{ "Name": "nofile", "Soft": 1024, "Hard": 2048 }]
  }
}`

type CodetainerProfileSpec struct {
	Config     *docker.Config     `json:"Config,omitempty" yaml:"Config,omitempty"`
	HostConfig *docker.HostConfig `json:"HostConfig,omitempty" yaml:"HostConfig,omitempty"`
}

func parseJsonSpec(reader io.Reader) (*CodetainerProfileSpec, error) {
	var s *CodetainerProfileSpec
	if err := json.NewDecoder(reader).Decode(&s); err != nil {
		return nil, err
	}
	return s, nil
}

func loadJsonSpec(path string) (*CodetainerProfileSpec, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("JSON specification file for %s not found", path)
		}
		return nil, err
	}
	defer f.Close()
	return parseJsonSpec(f)
}
