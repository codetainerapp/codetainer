package main

import (
	"errors"

	"github.com/gorilla/mux"
)

func RouteIndex(ctx *Context) error {
	return executeTemplate(ctx, "install.html", 200, map[string]interface{}{
		"Section": "install",
	})
}

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
