# codetainer

![codetainer gif](codetainer.gif?raw=true)

[![Build Status](http://komanda.io:8080/api/badge/github.com/recruit2class/codetainer/status.svg?branch=master)](http://komanda.io:8080/github.com/recruit2class/codetainer)

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

### Building & Installing From Source 

```bash
# set your $GOPATH
go get github.com/recruit2class/codetainer
# you may get errors about not compiling due to Asset missing, it's ok. bindata.go needs to be created
# by `go generate` first.
cd $GOPATH/src/github.com/recruit2class/codetainer
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
time you run codetainer, please edit defaults as appropriate.

```toml
# Docker API server and port
DockerServer = "localhost"
DockerPort = 4500

# Enable TLS support (optional, if you access to Docker API over HTTPS)
# DockerServerUseHttps = true
# Certificate directory path (optional)
#   e.g. if you use Docker Machine: "~/.docker/machine/certs"
# DockerCertPath = "/path/to/certs"

# Database path (optional, default is ~/.codetainer/codetainer.db)
# DatabasePath = "/path/to/codetainer.db"
```

## Running an example codetainer

```bash
$ sudo docker pull ubuntu:14.04
$ codetainer image register ubuntu:14.04
$ codetainer create ubuntu:14.04 my-codetainer-name
$ codetainer server  # to start the API server on port 3000
```

### Embedding a codetainer in your web app 

 1. Copy [codetainer.js](web/public/javascript/codetainer.js) to your webapp. 
 2. Include `codetainer.js` and `jquery` in your web page. Create a div
to house the codetainer terminal iframe (it's `#terminal` in the example below).

 ```html 
<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8">
    <title>lsof tutorial</title>
    <link rel='stylesheet' href='/stylesheets/style.css' />
    <script src="http://code.jquery.com/jquery-1.10.1.min.js"></script>
    <script src="/javascripts/codetainer.js"></script>
    <script src="/javascripts/lsof.js"></script>
  </head>
  <body>
     <div id="terminal" data-container="YOUR CODETAINER ID HERE"> 
  </body>
</html> 
 ```

 3. Run the javascript to load the codetainer iframe from the 
codetainer API server (supply `data-container` as the id of codetainer on 
the div, or supply `codetainer` in the constructor options).

```js 
 $('#terminal').codetainer({
     terminalOnly: false,                 // set to true to show only a terminal window 
     url: "http://127.0.0.1:3000",        // replace with codetainer server URL
     container: "YOUR CONTAINER ID HERE",
     width: "100%",
     height: "100%",
  });
```

### API Documentation

*TODO*

### Profiles

*TODO* more documentation.

You can use profiles to apply Docker configs to limit CPU, memory, network access,
and more.

See [example profiles](example-profiles) for some examples of this.

Register a profile to use with codetainer using `codetainer profile register <path-to-json> <name of profile>`
and then supply `codetainer-config-id` when POST'ing to `/api/v1/codetainer` to create.

# Status

Codetainer is unstable and in active development.

