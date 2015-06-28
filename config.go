package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	DockerServer string `toml:"-"`
	DockerPort   int
}

func (c *Config) GetDockerEndpoint() string {
	return fmt.Sprintf("%s:%d", c.DockerServer, c.DockerPort)
}

var (
	ConfigPath = "config.toml"

	DefaultConfigFileSettings = `# Docker API server and port 
DockerServer = "localhost"
DockerPort = 4500`
	GlobalConfig Config
)

func NewConfig(configPath string) (*Config, error) {
	Log.Debugf("Loading %s configurations", Name)

	if configPath == "" {
		configPath = ConfigPath
	}

	config := &Config{}

	if !IsExist(configPath) {

		configData := []byte(DefaultConfigFileSettings)

		f, err := os.Create(ConfigPath)

		if err != nil {
			Log.Error(err)
			Log.Fatalf("Unable to create configuration file: %s.", ConfigPath)
		}

		_, err = f.Write(configData)

		if err != nil {
			Log.Error(err)
			Log.Fatalf("Unable to create configuration file: %s.", ConfigPath)
		}

		f.Sync()
		f.Close()

	}

	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return config, err
	}
	return config, nil
}
