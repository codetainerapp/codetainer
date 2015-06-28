package main

import (
	"errors"
	"net"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

func RouteIndex(ctx *Context) error {
	return executeTemplate(ctx, "install.html", 200, map[string]interface{}{
		"Section": "install",
	})
}

type ContainerConnection struct {
	web       *websocket.Conn
	container *websocket.Conn
}

func (c *ContainerConnection) read() {
	for {
		_, message, err := c.web.ReadMessage()
		if err != nil {
			break
		}
		c.container.WriteMessage(websocket.TextMessage, message)
	}
	c.web.Close()
}

func (c *ContainerConnection) write() {
	for {
		_, message, err := c.container.ReadMessage()
		if err != nil {
			break
		}
		c.web.WriteMessage(websocket.TextMessage, message)
	}
	c.container.Close()
}

func (c *ContainerConnection) Start() {
	go c.read()
	c.write()
}

func newContainerClient() (*websocket.Conn, error) {
	id := "9a58ba6db179"
	u, err := url.Parse("http://localhost:4500/v1.5/containers/" + id + "/attach/ws?logs=1&stderr=1&stdout=1&stream=1&stdin=1")
	if err != nil {
		return nil, err
	}

	rawConn, err := net.Dial("tcp", u.Host)
	if err != nil {
		return nil, err
	}

	wsHeaders := http.Header{
		"Origin": {"http://localhost:4500"},
	}

	wsConn, resp, err := websocket.NewClient(rawConn, u, wsHeaders, 1024, 1024)

	if err != nil {
		Log.Error("Unable to connect", resp, err)
	}
	return wsConn, err
}

func RouteApiV1CodetainerAttach(ctx *Context) error {
	connection := &ContainerConnection{}
	if c, err := newContainerClient(); err != nil {
		return err
	} else {
		connection.container = c
	}

	if ctx.WS == nil {
		return errors.New("No websocket connection for web client")
	}
	connection.web = ctx.WS
	Log.Info("XXX We started")
	connection.Start()

	return nil
}
