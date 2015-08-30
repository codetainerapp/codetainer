package codetainer

import (
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/schema"
)

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

func (img *CodetainerImage) Parse(form url.Values) error {
	decoder := schema.NewDecoder()
	// r.PostForm is a map of our POST form values
	err := decoder.Decode(img, form)
	return err
}

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

type Codetainer struct {
	Id        string
	ImageId   string
	Defunct   bool
	CreatedAt time.Time
	UpdatedAt time.Time
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
