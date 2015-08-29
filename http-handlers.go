package codetainer

import (
	"bytes"
	"errors"
	"strings"

	"github.com/Unknwon/com"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/gorilla/mux"
)

func RouteIndex(ctx *Context) error {
	return executeTemplate(ctx, "install.html", 200, map[string]interface{}{
		"Section": "install",
	})
}

func execInContainer(client *docker.Client,
	id string,
	command []string) (string, string, error) {

	exec, err := client.CreateExec(docker.CreateExecOptions{
		AttachStderr: true,
		AttachStdin:  false,
		AttachStdout: true,
		Tty:          false,
		Cmd:          command,
		Container:    id,
	})

	if err != nil {
		return "", "", err
	}

	var outputBytes []byte
	outputWriter := bytes.NewBuffer(outputBytes)
	var errorBytes []byte
	errorWriter := bytes.NewBuffer(errorBytes)

	err = client.StartExec(exec.ID, docker.StartExecOptions{
		OutputStream: outputWriter,
		ErrorStream:  errorWriter,
	})

	return outputWriter.String(), errorWriter.String(), err
}

func RouteApiV1CodetainerTTY(ctx *Context) error {
	if ctx.R.Method == "POST" {
		return RouteApiV1CodetainerUpdateCurrentTTY(ctx)
	} else {
		return RouteApiV1CodetainerGetCurrentTTY(ctx)
	}
}

func RouteApiV1CodetainerUpdateCurrentTTY(ctx *Context) error {
	vars := mux.Vars(ctx.R)
	id := vars["id"]
	if id == "" {
		return errors.New("id is required")
	}

	client, err := GlobalConfig.GetDockerClient()
	if err != nil {
		return err
	}

	height := com.StrTo(ctx.R.FormValue("height")).MustInt()

	if height == 0 {
		return errors.New("height is required")
	}

	width := com.StrTo(ctx.R.FormValue("width")).MustInt()

	if width == 0 {
		return errors.New("width is required")
	}

	err = client.ResizeContainerTTY(id, height, width)

	if err != nil {
		return err
	}
	return renderJson(map[string]interface{}{
		"success": true,
	}, ctx.W)
}

//
// Get TTY size
//
func RouteApiV1CodetainerGetCurrentTTY(ctx *Context) error {

	vars := mux.Vars(ctx.R)
	id := vars["id"]
	if id == "" {
		return errors.New("id is required")
	}

	client, err := GlobalConfig.GetDockerClient()
	if err != nil {
		return err
	}
	col, _, err := execInContainer(client, id, []string{"tput", "cols"})
	col = strings.Trim(col, "\n")
	if err != nil {
		return err
	}
	lines, _, err := execInContainer(client, id, []string{"tput", "lines"})
	lines = strings.Trim(lines, "\n")
	if err != nil {
		return err
	}

	return renderJson(map[string]interface{}{
		"col":  col,
		"rows": lines,
	}, ctx.W)

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

	client, err := GlobalConfig.GetDockerClient()
	if err != nil {
		return err
	}

	err = client.StopContainer(id, 30)

	if err != nil {
		return err
	}

	return nil
}

type FileDesc struct {
	name string
	size int64
}

func parseFiles(output string) []FileDesc {
	files := make([]FileDesc, 0)
	return files
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

	client, err := GlobalConfig.GetDockerClient()
	if err != nil {
		return err
	}

	exec, err := client.CreateExec(docker.CreateExecOptions{
		AttachStderr: true,
		AttachStdin:  false,
		AttachStdout: true,
		Tty:          false,
		Cmd:          []string{"/codetainer/utils/files", "--path", path},
		Container:    id,
	})

	if err != nil {
		return err
	}

	var outputBytes []byte
	outputWriter := bytes.NewBuffer(outputBytes)
	var errorBytes []byte
	errorWriter := bytes.NewBuffer(errorBytes)

	err = client.StartExec(exec.ID, docker.StartExecOptions{
		OutputStream: outputWriter,
		ErrorStream:  errorWriter,
	})

	if err != nil {
		return err
	}

	files, err := makeShortFiles(outputWriter.Bytes())
	if err != nil {
		return err
	}

	return renderJson(map[string]interface{}{
		"files": files,
		"error": errorWriter.String(),
	}, ctx.W)

}

//
// Create a codetainer
//
func RouteApiV1CodetainerCreate(ctx *Context) error {

	if ctx.R.Method != "POST" {
		return errors.New("POST only")
	}

	imageId := ctx.R.FormValue("image-id")
	name := ctx.R.FormValue("name")

	Log.Infof("Creating codetainer from image: %s", imageId)
	client, err := GlobalConfig.GetDockerClient()
	if err != nil {
		return err
	}

	db, err := GlobalConfig.GetDatabase()
	if err != nil {
		return err
	}

	image, err := db.LookupCodetainerImage(imageId)
	if err != nil {
		return err
	}
	if image == nil {
		return errors.New("no image found")
	}

	// TODO: all the other configs
	c, err := client.CreateContainer(docker.CreateContainerOptions{
		Name: name,
		Config: &docker.Config{
			OpenStdin: true,
			Tty:       true,
			Image:     image.Id,
		},
		HostConfig: &docker.HostConfig{
			Binds: []string{
				GlobalConfig.UtilsPath() + ":/codetainer/utils:ro",
			},
		},
	})

	if err != nil {
		Log.Error("unable to create container: "+name, err)
		return err
	}

	// TODO fetch config for codetainer
	err = client.StartContainer(c.ID, &docker.HostConfig{
		Binds: []string{
			GlobalConfig.UtilsPath() + ":/codetainer/utils:ro",
		},
	})

	if err != nil {
		Log.Error(err)
		return err
	}

	codetainer, err := db.SaveCodetainer(c.ID, imageId)

	if err != nil {
		Log.Error(err)
		return err
	}

	return renderJson(map[string]interface{}{
		"codetainer": codetainer,
		"error":      "",
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

	Log.Infof("Starting codetainer: %s", id)
	client, err := GlobalConfig.GetDockerClient()
	if err != nil {
		return err
	}

	// TODO fetch config for codetainer
	err = client.StartContainer(id, &docker.HostConfig{
		Binds: []string{
			GlobalConfig.UtilsPath() + ":/codetainer/utils:ro",
		},
	})

	if err != nil {
		Log.Error(err)
		return err
	}

	return nil
}

func RouteApiV1Codetainer(ctx *Context) error {
	if ctx.R.Method == "POST" {
		return RouteApiV1CodetainerCreate(ctx)
	} else {
		return RouteApiV1CodetainerList(ctx)
	}
}

//
// List all running codetainers
//
func RouteApiV1CodetainerList(ctx *Context) error {
	client, err := GlobalConfig.GetDockerClient()
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
// Send a command to a container
//
func RouteApiV1CodetainerSend(ctx *Context) error {
	vars := mux.Vars(ctx.R)

	if ctx.R.Method != "POST" {
		return errors.New("POST only")
	}

	id := vars["id"]

	if id == "" {
		return errors.New("ID of container must be provided")
	}

	cmd := ctx.R.FormValue("command")

	Log.Infof("Sending command to container: %s -> %s ", id, cmd)

	connection := &ContainerConnection{id: id, web: ctx.WS}

	err := connection.SendSingleMessage(cmd + "\n")

	return err
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
		return errors.New("No websocket connection for web client: " + ctx.R.URL.String())
	}

	connection := &ContainerConnection{id: id, web: ctx.WS}

	err := connection.Start()

	return err
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

//
// List images
//
func RouteApiV1CodetainerListImages(ctx *Context) error {
	db, err := GlobalConfig.GetDatabase()
	if err != nil {
		return err
	}
	images, err := db.ListCodetainerImages()
	if err != nil {
		return err
	}

	return renderJson(map[string]interface{}{
		"images": images,
	}, ctx.W)
}
