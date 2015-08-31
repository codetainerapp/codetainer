package codetainer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	version "github.com/hashicorp/go-version"

	"github.com/BurntSushi/toml"
	docker "github.com/fsouza/go-dockerclient"
)

type Config struct {
	DockerServerUseHttps bool
	DockerServer         string
	DockerPort           int
	DatabasePath         string
	database             *Database
}

func (c *Config) Url() string {
	// TODO: make this configurable
	return "http://localhost:3000"
}

func (c *Config) GetDatabase() (*Database, error) {
	// TODO cache db

	if c.database != nil {
		return c.database, nil
	}

	db, err := NewDatabase(c.GetDatabasePath())
	if err != nil {
		return nil, err
	}
	c.database = db
	return c.database, nil
}

func (c *Config) GetDatabasePath() string {

	if c.DatabasePath == "" {
		// basePath := "/var/lib/codetainer/"
		// basePath := "./"
		c.DatabasePath = "codetainer.db"

		// if _, err := os.Stat(c.DatabasePath); err != nil {
		// if os.IsNotExist(err) {
		// err := os.MkdirAll("/var/lib/codetainer", 0700)
		// if err != nil {
		// Log.Fatal("Unable to create path for database: " + basePath)
		// }
		// } else {
		// Log.Fatal(err)
		// }
		// }
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

func (c *Config) testDockerClient() error {
	endpoint, err := c.GetDockerClient()
	if err != nil {
		return err
	}
	return endpoint.Ping()
}

func (c *Config) testDockerVersion() error {
	endpoint, err := c.GetDockerClient()
	if err != nil {
		return err
	}
	ev, err := endpoint.Version()
	if err != nil {

		return err
	}
	currVersion := ev.Get("ApiVersion")
	activeVersion, err := version.NewVersion(currVersion)
	supportedVersion, err := version.NewVersion(DockerApiVersion)
	if activeVersion.LessThan(supportedVersion) {
		return errors.New(currVersion + " version is lower than supported Docker version of " + DockerApiVersion + ". You will need to upgrade docker.")
	}
	return nil
}

func (c *Config) GetDockerEndpoint() string {
	if c.DockerServerUseHttps {
		return fmt.Sprintf("https://%s:%d", c.DockerServer, c.DockerPort)
	} else {
		return fmt.Sprintf("http://%s:%d", c.DockerServer, c.DockerPort)
	}
}

//
// Ensure a configuration is valid and all dependencies are installed.
//
func (c *Config) TestConfig() bool {
	err := c.testDockerClient()
	if err != nil {
		Log.Fatal(`Unable to connect to Docker API.  Are you sure you have
configured the Docker API to accept remote HTTP connections?

E.g., your docker service needs to have the following parameters in the
command line:

  /usr/bin/docker -d -H tcp://127.0.0.1:4500

Please also check your config.toml has the correct configuration for the DockerServer
and DockerPort:

  # Docker API server and port
  DockerServer = "localhost"
  DockerPort = 4500
`)

	}
	err = c.testDockerVersion()
	if err != nil {
		Log.Fatal(err)
	}

	return true
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
