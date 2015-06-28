package main

import (
	"errors"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/gorilla/mux"
)

func RouteIndex(ctx *Context) error {
	return executeTemplate(ctx, "install.html", 200, map[string]interface{}{
		"Section": "install",
	})
}

//
// Stop a codetainer
//
func RouteApiV1CodetainerStop(ctx *Context) error {
	// TODO
	return nil
}

//
// Start a codetainer
//
func RouteApiV1CodetainerStart(ctx *Context) error {
	// TODO
	return nil
}

//
// List all running codetainers
//
func RouteApiV1CodetainerList(ctx *Context) error {
	endpoint := GlobalConfig.GetDockerEndpoint()
	_, err := docker.NewClient(endpoint)
	if err != nil {

	}

	return nil
	// TODO
}

//
// Attach to a codetainer
//
func RouteApiV1CodetainerAttach(ctx *Context) error {
	vars := mux.Vars(ctx.R)
	id := vars["id"]

	if id == "" {
		return errors.New("ID of container must be provided")
	}

	if ctx.WS == nil {
		return errors.New("No websocket connection for web client")
	}

	connection := &ContainerConnection{id: id, web: ctx.WS}

	connection.Start()

	return nil
}
