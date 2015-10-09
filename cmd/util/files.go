package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	Version = "0.1.0"
	app     = kingpin.New("files", "Command for listing a file in JSON")
	appPath = app.Flag("path", "path to list").Required().Short('s').String()
)

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

//
// Go helper utility for container
// to perform file commands.
//
// Will output FileInfo to JSON
//
func main() {

	app.Version(Version)
	_, err := app.Parse(os.Args[1:])

	if err != nil {
		app.Usage([]string{})
		log.Fatal(err)
		os.Exit(1)
	}

	files, err := ioutil.ReadDir(*appPath)
	if err != nil {
		log.Fatal(err)
	}

	sfiles := make([]*ShortFileInfo, 0)
	for _, f := range files {
		sfile := NewShortFileInfo(f)
		sfiles = append(sfiles, sfile)
	}
	bytes, err := json.Marshal(sfiles)

	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.Write(bytes)

	os.Exit(0)
}
