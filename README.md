# codetainer

[![Build Status](http://komanda.io:8080/api/badge/github.com/codetainerapp/codetainer/status.svg?branch=master)](http://komanda.io:8080/github.com/codetainerapp/codetainer)

`codetainer` allows you to create code 'sandboxes' you can embed in your 
web applications (think of it like an OSS clone of [codepicnic.com](http://codepicnic.com)).

Codetainer runs as a webservice and provides APIs to create, view, and attach to the 
sandbox along with a nifty HTML terminal you can interact with the sandbox in 
realtime. It uses Docker and its introspection APIs to provide the majority
of this functionality.

Codetainer is written in Go.

For more information, see [the slides from a talk introduction](https://www.slideshare.net/JenAndre/codetainer-a-browser-code-sandbox).

# Build & Installation

## Requirements

  * Docker >=1.8 (required for file upload API)
  * Go >=1.4
  * [godep](https://github.com/tools/godep)

## Installing from `go get`


## Building & Installing From Source 

```bash
# set your $GOPATH
go get github.com/codetainerapp/codetainer
cd $GOPATH/src/github.com/codetainerapp/codetainer
# make install_deps  # if you need the dependencies like godep
make
```

This will create ./bin/codetainer.

## Configuring Docker

You must configure Docker to listen on a TCP port.

```
DOCKER_OPTS="-H tcp://127.0.0.1:4500 -H unix:///var/run/docker.sock"
```

## Configuring codetainer

See ~/.codetainer/config.toml.  This file will get auto-generated the first 
time you run codetainer.

```toml
# Docker API server and port
DockerServer = "localhost"
DockerPort = 4500
# Database path (optional, default is ~/codetainer/codetainer.db)
# DatabasePath = "/path/to/codetainer.db"
```

## Running an example codetainer

```bash
$ sudo docker pull ubuntu:14.04
$ codetainer image register ubuntu:14.04
$ codetainer create ubuntu:14.04 my-codetainer-name
$ codetainer server  # to start the API server
```

# Status

Codetainer is unstable and in active development.
