package codetainer

import (
	"encoding/json"
	"os"
	"time"
)

type CodetainerImage struct {
	Id                  string `xorm:"varchar(128) not null unique" json:"id"`
	DefaultStartCommand string
	Description         string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Enabled             bool
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
