package main

import (
	"runtime"

	codetainer "github.com/codetainerapp/codetainer"
)

var (
	// Build SHA
	Build string
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	codetainer.Build = Build
	codetainer.Start()
}
