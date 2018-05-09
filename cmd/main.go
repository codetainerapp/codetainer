package main

import (
	"runtime"

	codetainer "github.com/recruit2class/codetainer"
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
