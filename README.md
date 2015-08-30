# codetainer

[![Build Status](https://drone.io/github.com/codetainerapp/codetainer/status.png)](https://drone.io/github.com/codetainerapp/codetainer/latest)

`codetainer` allows you to create code 'sandboxes' you can embed in your 
web applications (think of it like an OSS clone of [codepicnic.com](http://codepicnic.com)).

Codetainer runs as a webservice and provides APIs to create, view, and attach to the 
sandbox along with a nifty HTML terminal you can interact with the sandbox in 
realtime. It uses Docker and its introspection APIs to provide the majority
of this functionality.

Codetainer is written in Go.

# Build & Installation

## Requirements

  * Docker >=1.7 (older versions may work, untested)
  * Go >=1.4
  * [godep](https://github.com/tools/godep)

## Building & Installing From Source 

```bash
# set your $GOPATH
go get github.com/codetainerapp/codetainer
cd $GOPATH/src/github.com/codetainerapp/codetainer
make updatedeps
make
make install  # optional, to install to /opt/codetainer
```

## Configuring the web server

TODO 

## Running an example codetainer

TODO


