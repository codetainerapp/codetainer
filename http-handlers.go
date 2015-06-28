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

	if ctx.R.Method != "POST" {
		return errors.New("POST only")
	}

	vars := mux.Vars(ctx.R)
	id := vars["id"]
	endpoint := GlobalConfig.GetDockerEndpoint()
	client, err := docker.NewClient(endpoint)
	if err != nil {
		return err
	}

	err = client.StopContainer(id, 30)

	if err != nil {
		return err
	}

	return nil
}

//
// Start a stopped codetainer
//
func RouteApiV1CodetainerStart(ctx *Context) error {

	if ctx.R.Method != "POST" {
		return errors.New("POST only")
	}
	vars := mux.Vars(ctx.R)
	id := vars["id"]
	endpoint := GlobalConfig.GetDockerEndpoint()
	client, err := docker.NewClient(endpoint)
	if err != nil {
		return err
	}

	// TODO fetch config for codetainer
	err = client.StartContainer(id, &docker.HostConfig{})

	if err != nil {
		return err
	}

	return nil
}

//
// List all running codetainers
//
func RouteApiV1CodetainerList(ctx *Context) error {
	endpoint := GlobalConfig.GetDockerEndpoint()
	client, err := docker.NewClient(endpoint)
	if err != nil {
		return err
	}
	containers, err := client.ListContainers(docker.ListContainersOptions{})

	if err != nil {
		return err
	}
	return renderJson(map[string]interface{}{
		"containers": containers,
	}, ctx.W)
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

//
// View codetainer
//
func RouteApiV1CodetainerView(ctx *Context) error {
	vars := mux.Vars(ctx.R)
	id := vars["id"]

	if id == "" {
		return errors.New("ID of container must be provided")
	}

	return executeRaw(ctx, "view.html", 200, map[string]interface{}{
		"Section":             "ContainerView",
		"PageIsContainerView": true,
		"ContainerId":         id,
	})
}
