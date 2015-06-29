package main

import (
	"runtime"

	codetainer "github.com/codetainerapp/codetainer"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	codetainer.Start()
}
