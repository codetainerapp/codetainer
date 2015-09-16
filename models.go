package codetainer

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/gorilla/schema"
	"github.com/pborman/uuid"
)

func parseObjectFromForm(p interface{}, form url.Values) error {
	decoder := schema.NewDecoder()
	// r.PostForm is a map of our POST form values
	err := decoder.Decode(p, form)
	return err
}

type CodetainerConfig struct {
	// id
	Id   string `xorm:"not null unique pk varchar(128)" json:"id" schema:"id"`
	Name string `xorm:"not null unique" json:"name" schema:"-"`
	// profile string with the CodetainerSpec format
	Profile   string    `json:"profile" schema:"profile"`
	CreatedAt time.Time `schema:"-"`
	UpdatedAt time.Time `schema:"-"`
	Enabled   bool
}

func (c *CodetainerConfig) LookupByNameOrId(db *Database) error {
	id := c.Id
	if c.Lookup(db) != nil {
		c.Name = id
		c.Id = ""
		err := c.Lookup(db)
		if err != nil {
			return errors.New("No config found: " + id)
		}
		return err
	}
	return nil
}

func (c *CodetainerConfig) Lookup(db *Database) error {
	Log.Debug("Looking up: ", c)
	has, err := db.engine.Get(c)
	if err != nil {
		return err
	}
	if !has {
		return errors.New("No codetainer config found: " + c.Id)
	}
	return nil
}

func (c *CodetainerConfig) GetProfileSpec() (*CodetainerProfileSpec, error) {
	return parseJsonSpec(strings.NewReader(c.Profile))
}

func (c *CodetainerConfig) Validate() error {
	_, err := c.GetProfileSpec()
	return err
}

func (c *CodetainerConfig) Save(db *Database) error {
	if c.Id != "" {
		_, err := db.engine.Update(c)
		return err
	} else {
		c.Id = uuid.New()
		_, err := db.engine.Insert(c)
		return err
	}
}

//
// Container image.
//
// swagger:parameters imageCreate
type CodetainerImage struct {
	Id                  string    `xorm:"varchar(128) not null unique" json:"id" schema:"id"`
	DefaultStartCommand string    `json:"command" schema:"command"`
	Description         string    `json:"description" schema:"description"`
	Tags                []string  `schema:"-"`
	CreatedAt           time.Time `schema:"-"`
	UpdatedAt           time.Time `schema:"-"`
	Enabled             bool
}

func (img *CodetainerImage) Register(db *Database) error {
	// check if image is in docker
	image := lookupImageInDocker(img.Id)

	if image != nil {
		img.Id = image.ID
		img.Tags = image.RepoTags
		img.Enabled = true

		Log.Info("Registering New Image: ", img)
		_, err := db.engine.Insert(img)
		return err

	} else {
		return errors.New("No image found in docker: " + img.Id)
	}

	return nil
}

func createTarFile(fileData []byte, fileName string) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	var file = struct {
		Name string
		Body []byte
	}{Name: fileName, Body: fileData}

	hdr := &tar.Header{
		Name: file.Name,
		Mode: 0600,
		Size: int64(len(file.Body)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return nil, err
	}
	if _, err := tw.Write(file.Body); err != nil {
		return nil, err
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}
	return buf, nil
}

//
// Codetainer data structure.
//
// swagger:parameters codetainerCreate
type Codetainer struct {
	Id        string    `schema:"id" json:"id"`
	Name      string    `schema:"name" json:"name"`
	ImageId   string    `schema:"image-id" json:"image-id"`
	Defunct   bool      `schema"-"`          // false if active
	Running   bool      `schema"-" xorm:"-"` // true if running
	Profile   string    `schema:"-"`
	CreatedAt time.Time `schema:"-"`
	UpdatedAt time.Time `schema:"-"`
}

func (c *Codetainer) GetProfileSpec() (*CodetainerProfileSpec, error) {
	return parseJsonSpec(strings.NewReader(c.Profile))
}

func (codetainer *Codetainer) DownloadFile(filePath string) ([]byte, error) {
	client, err := GlobalConfig.GetDockerClient()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	opts := docker.CopyFromContainerOptions{
		OutputStream: &buf,
		Container:    codetainer.Id,
		Resource:     filePath,
	}
	err = client.CopyFromContainer(opts)
	if err != nil {
		return nil, err
	}
	br := bytes.NewReader(buf.Bytes())
	tr := tar.NewReader(br)
	var resultBuf bytes.Buffer
	hdr, err := tr.Next()
	if err != nil {
		return nil, err
	}
	if hdr.FileInfo().IsDir() {
		// it's tarred
		return nil, errors.New("File is a directory:" + filePath)
	}

	io.Copy(&resultBuf, tr)
	return resultBuf.Bytes(), nil

}

//
// Upload a file to a `dstPath` in a container.
//
func (codetainer *Codetainer) UploadFile(
	fileData []byte,
	fileName string,
	dstFolder string) error {
	client, err := GlobalConfig.GetDockerClient()
	if err != nil {
		return err
	}

	buf, err := createTarFile(fileData, fileName)
	if err != nil {
		return err
	}

	fi := bytes.NewReader(buf.Bytes())

	opts := docker.UploadToContainerOptions{Path: dstFolder}
	opts.InputStream = fi
	Log.Debug("Writing file to codetainer")
	return client.UploadToContainer(codetainer.Id, opts)
}

func (codetainer *Codetainer) Remove(db *Database) error {
	Log.Debug("Removing codetainer:", codetainer.Id)
	client, err := GlobalConfig.GetDockerClient()
	if err != nil {
		return err
	}

	opts := docker.RemoveContainerOptions{
		ID:    codetainer.Id,
		Force: true,
	}

	err = client.RemoveContainer(opts)

	if err != nil {
		return err
	}

	codetainer.Defunct = true

	// for some reason engine.Update() isn't working for shit here
	// needless to say, only ever call this function when Id is populated
	// by the ORMe
	_, err = db.engine.Exec("Update codetainer set defunct = 1 where id = ?", codetainer.Id)
	return err
}

func (codetainer *Codetainer) Stop() error {
	client, err := GlobalConfig.GetDockerClient()
	if err != nil {
		return err
	}

	return client.StopContainer(codetainer.Id, 30)
}

func (codetainer *Codetainer) Start() error {
	client, err := GlobalConfig.GetDockerClient()
	if err != nil {
		return err
	}

	// TODO fetch config for codetainer
	return client.StartContainer(codetainer.Id, &docker.HostConfig{
		Binds: []string{
			GlobalConfig.UtilsPath() + ":/codetainer/utils:ro",
		},
	})
}

func (codetainer *Codetainer) LookupByNameOrId(id string, db *Database) error {
	codetainer.Id = id
	if codetainer.Lookup(db) != nil {
		codetainer.Name = id
		codetainer.Id = ""
		err := codetainer.Lookup(db)
		if err != nil {
			return errors.New("No codetainer found: " + id)
		}
		return err
	}
	return nil
}

func (codetainer *Codetainer) Lookup(db *Database) error {
	Log.Debug("Looking up: ", codetainer)
	has, err := db.engine.Get(codetainer)
	if err != nil {
		return err
	}
	if !has {
		return errors.New("No codetainer found: " + codetainer.Id)
	}
	return nil
}

func (codetainer *Codetainer) Create(db *Database) error {
	client, err := GlobalConfig.GetDockerClient()

	if err != nil {
		return err
	}

	image, err := db.LookupCodetainerImage(codetainer.ImageId)

	if err != nil {
		return err
	}

	if image == nil {
		return errors.New("no image found")
	}

	codetainer.ImageId = image.Id
	defaultConfig := &docker.Config{
		OpenStdin: true,
		Tty:       true,
		Image:     image.Id,
	}
	defaultHostConfig := &docker.HostConfig{
		Binds: []string{
			GlobalConfig.UtilsPath() + ":/codetainer/utils:ro",
		},
	}

	if codetainer.Profile != "" {
		config, err := codetainer.GetProfileSpec()
		if err != nil {
			Log.Error("Invalid profile spec", err)
			return err
		}
		if config.Config != nil {
			defaultConfig = config.Config
			defaultConfig.Tty = true
			defaultConfig.Image = image.Id
			defaultConfig.OpenStdin = true
		}
		if config.HostConfig != nil {
			defaultHostConfig = config.HostConfig
			defaultHostConfig.Binds = []string{
				GlobalConfig.UtilsPath() + ":/codetainer/utils:ro",
			}
		}

	}

	opts := docker.CreateContainerOptions{
		Name:       codetainer.Name,
		Config:     defaultConfig,
		HostConfig: defaultHostConfig,
	}

	Log.Debugf("Creating codetainer with opts: %+v\n", opts)
	c, err := client.CreateContainer(opts)

	if err != nil {
		return err
	}

	err = client.StartContainer(c.ID, defaultHostConfig)

	if err != nil {
		return err
	}

	codetainer.Id = c.ID
	return codetainer.Save(db)
}

func (c *Codetainer) Save(db *Database) error {
	_, err := db.engine.Insert(c)
	return err
}

type ShortFileInfo struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	IsDir   bool      `json:"is_dir"`
	IsLink  bool      `json:"is_link"`
	ModTime time.Time `json:"modified_time"`
}

func NewShortFileInfo(f os.FileInfo) *ShortFileInfo {
	fi := ShortFileInfo{}
	fi.Name = f.Name()
	fi.Size = f.Size()
	fi.IsDir = f.IsDir()
	fi.ModTime = f.ModTime()
	fi.IsLink = (f.Mode()&os.ModeType)&os.ModeSymlink > 0

	return &fi
}

func makeShortFiles(data []byte) (*[]ShortFileInfo, error) {
	files := make([]ShortFileInfo, 0)
	if err := json.Unmarshal(data, &files); err != nil {
		return nil, err
	}

	return &files, nil
}

//
// Return type for errors
//
// swagger:response APIErrorResponse
type APIErrorResponse struct {
	Error   bool   `json:"error" description:"set if an error is returned"`
	Message string `json:"message" description:"error message string"`
}

//
// TTY parameters for a codetainer
//
// swagger:parameters updateCurrentTTY
type TTY struct {
	Height int `json:"height" description:"height of tty"`
	Width  int `json:"width" description:"width of tty"`
}

//
// TTY response
//
// swagger:response TTYBody
type TTYBody struct {
	Tty TTY `json:"tty"`
}

//
// CodetainerImage response
//
// swagger:response CodetainerImageBody
type CodetainerImageBody struct {
	Image CodetainerImage `json:"image"`
}

//
// CodetainerImageList response
//
// swagger:response CodetainerImageListBody
type CodetainerImageListBody struct {
	Images []CodetainerImage `json:"images"`
}

//
// Codetainer Create Params
//
// swagger:parameters CodetainerCreate
type CodetainerCreateParams struct {
	Codetainer
	CodetainerConfigId string `schema:"codetainer-config-id" json:"codetainer-config-id"`
}

//
// Codetainer response
//
// swagger:response CodetainerBody
type CodetainerBody struct {
	Codetainer `json:"codetainer"`
}

//
// CodetainerList response
//
// swagger:response CodetainerListBody
type CodetainerListBody struct {
	Codetainers []Codetainer `json:"codetainers"`
}

type GenericSuccess struct {
	Success bool `json:"success"`
}
