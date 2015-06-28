package main

import (
	"bytes"
	"errors"
	"strings"

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
// List files in a codetainer
//
func RouteApiV1CodetainerListFiles(ctx *Context) error {

	vars := mux.Vars(ctx.R)
	id := vars["id"]
	if id == "" {
		return errors.New("id is required")
	}

	path := ctx.R.FormValue("path")
	if path == "" {
		return errors.New("path is required")
	}

	endpoint := GlobalConfig.GetDockerEndpoint()
	client, err := docker.NewClient(endpoint)
	if err != nil {
		return err
	}

	exec, err := client.CreateExec(docker.CreateExecOptions{
		AttachStderr: true,
		AttachStdin:  false,
		AttachStdout: true,
		Tty:          false,
		Cmd:          []string{"ls", path},
		Container:    id,
	})

	if err != nil {
		return err
	}

	var outputBytes []byte
	outputWriter := bytes.NewBuffer(outputBytes)
	var errorBytes []byte
	errorWriter := bytes.NewBuffer(errorBytes)

	// TODO fetch config for codetainer
	err = client.StartExec(exec.ID, docker.StartExecOptions{
		OutputStream: outputWriter,
		ErrorStream:  errorWriter,
	})

	if err != nil {
		return err
	}

	files := strings.Split(outputWriter.String(), "\n")

	// TODO: parse into string
	return renderJson(map[string]interface{}{
		"files": files,
		"error": errorWriter.String(),
	}, ctx.W)

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
