package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	codetainer "github.com/codetainerapp/codetainer"
	kingpin "gopkg.in/alecthomas/kingpin.v1"
)

var (
	Version = "0.1.0"
	app     = kingpin.New("files", "Command for listing a file in JSON")
	appPath = app.Flag("path", "path to list").Required().Short('s').String()
)

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
		app.Usage(os.Stderr)
		log.Fatal(err)
		os.Exit(1)
	}

	files, err := ioutil.ReadDir(*appPath)
	if err != nil {
		log.Fatal(err)
	}

	sfiles := make([]*codetainer.ShortFileInfo, 0)
	for _, f := range files {
		sfile := codetainer.NewShortFileInfo(f)
		sfiles = append(sfiles, sfile)
	}
	bytes, err := json.Marshal(sfiles)

	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.Write(bytes)

	os.Exit(0)
}
