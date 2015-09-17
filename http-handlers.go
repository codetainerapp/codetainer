// Package codetainer Codetainer API
//
// This API allows you to create, attach, and interact with codetainers.
//
//     Schemes: http, https
//     Host: localhost
//     BasePath: /api/v1
//     Version: 0.0.1
//     License: MIT http://opensource.org/licenses/MIT
//     Contact: info@codetainer.org
//
//     Consumes:
//     - application/json
//
//     Produces:
//     - application/json
//
// swagger:meta
package codetainer

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Unknwon/com"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/gorilla/mux"
)

func RouteIndex(ctx *Context) error {
	return executeTemplate(ctx, "install.html", 200, map[string]interface{}{
		"Section": "install",
	})
}

func RouteApiV1CodetainerTTY(ctx *Context) error {
	if ctx.R.Method == "POST" {
		return RouteApiV1CodetainerUpdateCurrentTTY(ctx)
	} else {
		return RouteApiV1CodetainerGetCurrentTTY(ctx)
	}
}

func RouteApiV1CodetainerFile(ctx *Context) error {
	switch ctx.R.Method {
	case "PUT":
		return RouteApiV1CodetainerFileUpload(ctx)
	case "GET":
		return RouteApiV1CodetainerFileList(ctx)
	}

	return errors.New(ctx.R.URL.String() + ": Unsupported method " + ctx.R.Method)
}

func RouteApiV1CodetainerImage(ctx *Context) error {
	switch ctx.R.Method {
	case "POST":
		return RouteApiV1CodetainerImageCreate(ctx)
	case "GET":
		return RouteApiV1CodetainerImageList(ctx)
	}

	return errors.New(ctx.R.URL.String() + ": Unsupported method " + ctx.R.Method)
}

func RouteApiV1Codetainer(ctx *Context) error {

	switch ctx.R.Method {
	case "POST":
		return RouteApiV1CodetainerCreate(ctx)
	case "GET":
		return RouteApiV1CodetainerList(ctx)
	}

	return errors.New(ctx.R.URL.String() + ": Unsupported method " + ctx.R.Method)
}

func RouteApiV1CodetainerId(ctx *Context) error {

	switch ctx.R.Method {
	case "GET":
		return RouteApiV1CodetainerGet(ctx)
	case "DELETE":
		return RouteApiV1CodetainerRemove(ctx)
	}

	return errors.New(ctx.R.URL.String() + ": Unsupported method " + ctx.R.Method)
}

// ImageList swagger:route GET /image codetainer imageList
//
// List all codetainer images.
//
// Responses:
//    default: APIErrorResponse
//        200: CodetainerImageListBody
//
func RouteApiV1CodetainerImageList(ctx *Context) error {
	db, err := GlobalConfig.GetDatabase()
	if err != nil {
		return jsonError(err, ctx.W)

	}
	images, err := db.ListCodetainerImages()
	if err != nil {
		return jsonError(err, ctx.W)

	}

	return renderJson(map[string]interface{}{
		"images": images,
	}, ctx.W)
}

// ImageCreate swagger:route POST /image codetainer imageCreate
//
// Register a Docker image to be used as a codetainer.
//
// Responses:
//    default: APIErrorResponse
//        200: CodetainerImageBody
//
func RouteApiV1CodetainerImageCreate(ctx *Context) error {

	db, err := GlobalConfig.GetDatabase()
	if err != nil {
		return jsonError(err, ctx.W)
	}

	img := CodetainerImage{}
	ctx.R.ParseForm()
	if err := parseObjectFromForm(&img, ctx.R.PostForm); err != nil {
		return jsonError(err, ctx.W)
	}

	err = img.Register(db)

	if err != nil {
		return jsonError(err, ctx.W)
	}

	return renderJson(CodetainerImageBody{Image: img}, ctx.W)
}

// UpdateCurrentTTY swagger:route POST /codetainer/{id}/tty codetainer updateCurrentTTY
//
// Update the codetainer TTY height and width.
//
// Responses:
//    default: APIErrorResponse
//        200: TTYBody
//
func RouteApiV1CodetainerUpdateCurrentTTY(ctx *Context) error {
	vars := mux.Vars(ctx.R)
	id := vars["id"]
	if id == "" {
		return jsonError(errors.New("id is required"), ctx.W)
	}

	client, err := GlobalConfig.GetDockerClient()
	if err != nil {
		return jsonError(err, ctx.W)
	}

	height := com.StrTo(ctx.R.FormValue("height")).MustInt()

	if height == 0 {
		return jsonError(errors.New("height is required"), ctx.W)
	}

	width := com.StrTo(ctx.R.FormValue("width")).MustInt()

	if width == 0 {
		return jsonError(errors.New("width is required"), ctx.W)
	}

	err = client.ResizeContainerTTY(id, height, width)

	if err != nil {
		return jsonError(err, ctx.W)
	}

	tty := TTY{Height: height, Width: width}
	return renderJson(map[string]interface{}{
		"tty": tty,
	}, ctx.W)
}

// GetCurrentTTY swagger:route GET /codetainer/{id}/tty codetainer getCurrentTTY
//
// Return the codetainer TTY height and width.
//
// Responses:
//    default: APIErrorResponse
//        200: TTYBody
//
func RouteApiV1CodetainerGetCurrentTTY(ctx *Context) error {

	vars := mux.Vars(ctx.R)
	id := vars["id"]
	if id == "" {
		return jsonError(errors.New("id is required"), ctx.W)
	}

	client, err := GlobalConfig.GetDockerClient()
	if err != nil {
		return jsonError(err, ctx.W)
	}
	col, _, err := execInContainer(client, id, []string{"tput", "cols"})
	col = strings.Trim(col, "\n")
	if err != nil {
		return jsonError(err, ctx.W)
	}
	lines, _, err := execInContainer(client, id, []string{"tput", "lines"})
	lines = strings.Trim(lines, "\n")
	if err != nil {
		return jsonError(err, ctx.W)
	}

	height, _ := strconv.Atoi(lines)
	width, _ := strconv.Atoi(col)

	tty := TTY{Height: height, Width: width}

	return renderJson(map[string]interface{}{
		"tty": tty,
	}, ctx.W)

}

// CodetainerGet swagger:route GET /codetainer/{id} codetainer codetainerGet
//
// Get a codetainer
//
// Responses:
//    default: APIErrorResponse
//        200: CodetainerBody
//
func RouteApiV1CodetainerGet(ctx *Context) error {

	if ctx.R.Method != "DELETE" {
		return jsonError(errors.New("DELETE only"), ctx.W)
	}

	vars := mux.Vars(ctx.R)
	id := vars["id"]

	db, err := GlobalConfig.GetDatabase()

	if err != nil {
		return jsonError(err, ctx.W)
	}

	codetainer := Codetainer{}
	if err = codetainer.LookupByNameOrId(id, db); err != nil {
		return jsonError(err, ctx.W)
	}

	return renderJson(CodetainerBody{Codetainer: codetainer}, ctx.W)
}

// CodetainerRemove swagger:route DELETE /codetainer/{id} codetainer codetainerRemove
//
// Remove a codetainer
//
// Responses:
//    default: APIErrorResponse
//        200: CodetainerBody
//
func RouteApiV1CodetainerRemove(ctx *Context) error {

	if ctx.R.Method != "DELETE" {
		return jsonError(errors.New("DELETE only"), ctx.W)
	}

	vars := mux.Vars(ctx.R)
	id := vars["id"]

	db, err := GlobalConfig.GetDatabase()

	if err != nil {
		return jsonError(err, ctx.W)
	}

	codetainer := Codetainer{}
	if err = codetainer.LookupByNameOrId(id, db); err != nil {
		return jsonError(err, ctx.W)
	}

	err = codetainer.Remove(db)

	if err != nil {
		return jsonError(err, ctx.W)
	}

	return renderJson(CodetainerBody{Codetainer: codetainer}, ctx.W)
}

// CodetainerStop swagger:route POST /codetainer/{id}/stop codetainer codetainerStop
//
// Stop a codetainer
//
// Responses:
//    default: APIErrorResponse
//        200: CodetainerBody
//
func RouteApiV1CodetainerStop(ctx *Context) error {

	if ctx.R.Method != "POST" {
		return jsonError(errors.New("POST only"), ctx.W)
	}

	vars := mux.Vars(ctx.R)
	id := vars["id"]

	db, err := GlobalConfig.GetDatabase()

	if err != nil {
		return jsonError(err, ctx.W)
	}

	codetainer := Codetainer{}
	if err = codetainer.LookupByNameOrId(id, db); err != nil {
		return jsonError(err, ctx.W)
	}

	err = codetainer.Stop()

	if err != nil {
		return jsonError(err, ctx.W)
	}

	return renderJson(CodetainerBody{Codetainer: codetainer}, ctx.W)
}

// ListFiles swagger:route GET /codetainer/{id}/filescodetainer listFiles
//
// List Files in a codetainer
//
// Responses:
//    default: APIErrorResponse
//        200: Object
//
func RouteApiV1CodetainerFileList(ctx *Context) error {

	vars := mux.Vars(ctx.R)
	id := vars["id"]
	if id == "" {
		return jsonError(errors.New("id is required"), ctx.W)
	}

	path := ctx.R.FormValue("path")
	if path == "" {
		return jsonError(errors.New("path is required"), ctx.W)
	}

	client, err := GlobalConfig.GetDockerClient()
	if err != nil {
		return jsonError(err, ctx.W)
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
		return jsonError(err, ctx.W)
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
		return jsonError(err, ctx.W)
	}

	errString := errorWriter.String()
	if errString != "" {
		return jsonError(errors.New(errString), ctx.W)
	}

	files, err := makeShortFiles(outputWriter.Bytes())

	if err != nil {
		return jsonError(err, ctx.W)
	}

	return renderJson(map[string]interface{}{
		"files": files,
	}, ctx.W)

}

// CodetainerCreate swagger:route POST /codetainer codetainer codetainerCreate
//
// Create a new codetainer.
//
// Responses:
//    default: APIErrorResponse
//        200: CodetainerBody
//
func RouteApiV1CodetainerCreate(ctx *Context) error {

	if ctx.R.Method != "POST" {
		return jsonError(errors.New("POST only"), ctx.W)
	}

	codetainer := Codetainer{}
	codetainerPost := CodetainerCreateParams{}
	ctx.R.ParseForm()
	if err := parseObjectFromForm(&codetainerPost, ctx.R.PostForm); err != nil {
		return jsonError(err, ctx.W)
	}
	codetainer = codetainerPost.Codetainer
	codetainer.Id = ""

	if codetainer.ImageId == "" {
		return jsonError(errors.New("image-id is required"), ctx.W)
	}

	Log.Infof("Creating codetainer from image: %s", codetainer.ImageId)

	db, err := GlobalConfig.GetDatabase()
	if err != nil {
		return jsonError(err, ctx.W)
	}

	if codetainerPost.CodetainerConfigId != "" {
		cfgId := codetainerPost.CodetainerConfigId
		Log.Debug("Looking up codetainer config profile: ", cfgId)
		cfg := &CodetainerConfig{Id: cfgId}
		err := cfg.LookupByNameOrId(db)
		if err != nil {
			return jsonError(err, ctx.W)
		}
		codetainer.Profile = cfg.Profile
	}

	err = codetainer.Create(db)

	if err != nil {
		return jsonError(err, ctx.W)
	}

	return renderJson(CodetainerBody{
		Codetainer: codetainer,
	}, ctx.W)
}

//
// CodetainerStart swagger:route POST /codetainer/{id}/start codetainer codetainerStart
//
// Start a stopped codetainer
//
// Responses:
//    default: APIErrorResponse
//        200: CodetainerBody
//
func RouteApiV1CodetainerStart(ctx *Context) error {

	if ctx.R.Method != "POST" {
		return jsonError(errors.New("POST only"), ctx.W)
	}

	vars := mux.Vars(ctx.R)
	id := vars["id"]

	db, err := GlobalConfig.GetDatabase()

	if err != nil {
		return jsonError(err, ctx.W)
	}

	codetainer := Codetainer{}
	if err = codetainer.LookupByNameOrId(id, db); err != nil {
		return jsonError(err, ctx.W)
	}

	Log.Infof("Starting codetainer: %s", codetainer.Id)
	err = codetainer.Start()

	if err != nil {
		return jsonError(err, ctx.W)
	}

	return renderJson(CodetainerBody{
		Codetainer: codetainer,
	}, ctx.W)
}

//
// CodetainerList swagger:route GET /codetainer codetainer codetainerList
//
// List all codetainers
//
// Responses:
//    default: APIErrorResponse
//        200: CodetainerListBody
//
func RouteApiV1CodetainerList(ctx *Context) error {

	db, err := GlobalConfig.GetDatabase()

	if err != nil {
		return jsonError(err, ctx.W)
	}

	containers, err := db.ListCodetainers()

	if err != nil {
		return jsonError(err, ctx.W)
	}

	return renderJson(CodetainerListBody{
		Codetainers: *containers,
	}, ctx.W)
}

//
// CodetainerFileUpload swagger:route PUT /codetainer/{id}/file codetainerFileUpload
//
// Upload a file to a codetainer
//
// Responses:
//    default: APIErrorResponse
//        200: CodetainerListBody
//
func RouteApiV1CodetainerFileUpload(ctx *Context) error {

	db, err := GlobalConfig.GetDatabase()

	if err != nil {
		return jsonError(err, ctx.W)
	}

	vars := mux.Vars(ctx.R)
	id := vars["id"]

	codetainer := Codetainer{}
	if err = codetainer.LookupByNameOrId(id, db); err != nil {
		return jsonError(err, ctx.W)
	}

	ctx.R.ParseForm()
	file, hdr, err := ctx.R.FormFile("upload")

	dstPath := ctx.R.FormValue("dst_path")
	if dstPath == "" {
		return jsonError(errors.New("Destination folder dst_path is required"),
			ctx.W)
	}

	if err != nil {
		return jsonError(err, ctx.W)
	}

	filename := path.Base(hdr.Filename)
	dst := new(bytes.Buffer)
	io.Copy(dst, file)

	Log.Infof("Uploading %s to codetainer: %s", codetainer.Id)
	err = codetainer.UploadFile(dst.Bytes(), filename, dstPath)

	if err != nil {
		return jsonError(err, ctx.W)
	}

	return renderJson(map[string]interface{}{"success": true}, ctx.W)
}

// SendCommand swagger:route POST /codetainer/{id}/send codetainer sendCommand
//
// Send a command to a container
//
// Responses:
//    default: APIErrorResponse
//        200: GenericSuccess
func RouteApiV1CodetainerSend(ctx *Context) error {
	vars := mux.Vars(ctx.R)

	if ctx.R.Method != "POST" {
		return jsonError(errors.New("POST only"), ctx.W)
	}

	id := vars["id"]

	if id == "" {
		return jsonError(errors.New("ID of container must be provided"), ctx.W)
	}

	cmd := ctx.R.FormValue("command")

	Log.Infof("Sending command to container: %s -> %s ", id, cmd)

	connection := &ContainerConnection{id: id, web: ctx.WS}

	err := connection.SendSingleMessage(cmd)

	if err != nil {
		return jsonError(err, ctx.W)
	}

	return renderJson(map[string]interface{}{
		"success": true,
	}, ctx.W)
}

//
// FileDownload swagger:route GET /codetainer/{id}/file/download codetainer fileDownload
//
// Download a file from a codetainer.  Requires path={path of file to download}.
// Returns the multipart messsage with the file bytes.
//
// Responses:
//    default: APIErrorResponse
//        200: Object
//
func RouteApiV1CodetainerFileDownload(ctx *Context) error {
	vars := mux.Vars(ctx.R)
	id := vars["id"]

	if id == "" {
		return jsonError(errors.New("ID of container must be provided"), ctx.W)
	}

	filePath := ctx.R.FormValue("path")
	if filePath == "" {
		return jsonError(errors.New("path of file to download is required"), ctx.W)
	}

	codetainer := Codetainer{}
	db, err := GlobalConfig.GetDatabase()
	if err != nil {
		return jsonError(err, ctx.W)
	}
	if err = codetainer.LookupByNameOrId(id, db); err != nil {
		return jsonError(err, ctx.W)
	}
	fileBytes, err := codetainer.DownloadFile(filePath)

	if err != nil {
		return jsonError(err, ctx.W)
	}
	fileName := path.Base(filePath)

	http.ServeContent(ctx.W, ctx.R, fileName, time.Now(),
		bytes.NewReader(fileBytes))
	return nil
}

//
// Attach to a codetainer
//
func RouteApiV1CodetainerAttach(ctx *Context) error {
	vars := mux.Vars(ctx.R)
	id := vars["id"]

	if id == "" {
		return jsonError(errors.New("ID of container must be provided"), ctx.W)
	}

	if ctx.WS == nil {
		return jsonError(errors.New("No websocket connection for web client: "+ctx.R.URL.String()), ctx.W)
	}

	codetainer := Codetainer{}
	db, err := GlobalConfig.GetDatabase()
	if err != nil {
		return jsonError(err, ctx.W)
	}

	if err = codetainer.LookupByNameOrId(id, db); err != nil {
		return jsonError(err, ctx.W)
	}

	connection := &ContainerConnection{id: codetainer.Id, web: ctx.WS}

	err = connection.Start()

	if err != nil {
		return jsonError(err, ctx.W)
	}

	return renderJson(map[string]interface{}{
		"success": true,
	}, ctx.W)
}

//
// View codetainer
//
func RouteApiV1CodetainerView(ctx *Context) error {
	vars := mux.Vars(ctx.R)
	id := vars["id"]

	tOnly := ctx.R.FormValue("terminal-only")
	if id == "" {
		return errors.New("ID of container must be provided")
	}

	terminalOnly := false

	if tOnly == "1" || tOnly == "true" {
		terminalOnly = true
	}

	return executeRaw(ctx, "view.html", 200, map[string]interface{}{
		"Section":             "ContainerView",
		"PageIsContainerView": true,
		"ContainerId":         id,
		"terminalOnly":        terminalOnly,
	})
}
