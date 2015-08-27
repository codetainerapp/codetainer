package codetainer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	docker "github.com/fsouza/go-dockerclient"
)

type Config struct {
	DockerServerUseHttps bool
	DockerServer         string
	DockerPort           int
	DatabasePath         string
}

func (c *Config) GetDatabasePath() string {

	if c.DatabasePath != "" {
		basePath := "/var/lib/codetainer/"
		c.DatabasePath = basePath + "codetainer.db"

		if _, err := os.Stat(c.DatabasePath); err != nil {
			if os.IsNotExist(err) {
				os.MkdirAll("/var/lib/codetainer", 0600)
			} else {
				Log.Fatal(err)
			}
		}
	}
	return c.DatabasePath
}

func (c *Config) UtilsPath() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return dir + "/util"
}

func (c *Config) GetDockerClient() (*docker.Client, error) {
	endpoint := c.GetDockerEndpoint()
	return docker.NewClient(endpoint)
}

func (c *Config) GetDockerEndpoint() string {
	if c.DockerServerUseHttps {
		return fmt.Sprintf("https://%s:%d", c.DockerServer, c.DockerPort)
	} else {
		return fmt.Sprintf("http://%s:%d", c.DockerServer, c.DockerPort)
	}
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
