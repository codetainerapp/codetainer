package codetainer

import (
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/gorilla/schema"
)

func parseObjectFromForm(p interface{}, form url.Values) error {
	decoder := schema.NewDecoder()
	// r.PostForm is a map of our POST form values
	err := decoder.Decode(p, form)
	return err
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

// func (img *CodetainerImage) Parse(form url.Values) error {
// decoder := schema.NewDecoder()
// // r.PostForm is a map of our POST form values
// err := decoder.Decode(img, form)
// return err
// }

func (img *CodetainerImage) Register(db *Database) error {
	// check if image is in docker
	image := lookupImageInDocker(img.Id)

	if image != nil {
		if img.DefaultStartCommand == "" {
			img.DefaultStartCommand = DefaultExecCommand
		}
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

//
// Codetainer data structure.
//
// swagger:parameters codetainerCreate
type Codetainer struct {
	Id        string    `schema:"id" json:"id"`
	Name      string    `schema:"name" json:"name"`
	ImageId   string    `schema:"image-id" json:"image-id"`
	Defunct   bool      `schema"-"`
	CreatedAt time.Time `schema:"-"`
	UpdatedAt time.Time `schema:"-"`
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

func (codetainer *Codetainer) Lookup(db *Database) error {
	has, err := db.engine.Get(&codetainer)
	if err != nil {
		return err

	}
	if !has {
		return errors.New("No codetainer found:" + codetainer.Id)
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

	// TODO: all the other configs
	c, err := client.CreateContainer(docker.CreateContainerOptions{
		Name: codetainer.Name,
		Config: &docker.Config{
			OpenStdin: true,
			Tty:       true,
			Image:     image.Id,
		},
		HostConfig: &docker.HostConfig{
			Binds: []string{
				GlobalConfig.UtilsPath() + ":/codetainer/utils:ro",
			},
		},
	})

	if err != nil {
		return err
	}

	// TODO fetch config for codetainer
	err = client.StartContainer(c.ID, &docker.HostConfig{
		Binds: []string{
			GlobalConfig.UtilsPath() + ":/codetainer/utils:ro",
		},
	})

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
	Name    string
	Size    int64
	IsDir   bool
	IsLink  bool
	ModTime time.Time
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
// Codetainer response
//
// swagger:response CodetainerBody
type CodetainerBody struct {
	Codetainer `json:"codetainer"`
}
