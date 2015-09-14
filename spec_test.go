package codetainer

import (
	"strings"
	"testing"
)

func TestDeserializeJson(t *testing.T) {

	c, err := parseJsonSpec(strings.NewReader(defaultSpec))
	if err != nil {
		t.Fatal(err)
	}
	if c.Config == nil {
		t.Fatal("no config found")
	}
	if c.HostConfig == nil {
		t.Fatal("no host config found")
	}
	if len((*c.HostConfig).Ulimits) != 1 {
		t.Fatal("no ulimits detected")
	}

	// log.Printf("Config %+v \n", c.Config)
	// log.Printf("HostConfig %+v \n", c.HostConfig)
}
