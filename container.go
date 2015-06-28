package main

import (
	"net"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

type ContainerConnection struct {
	id        string
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

func (c *ContainerConnection) Start() error {

	if client, err := newContainerClient(c.id); err != nil {
		return err
	} else {
		c.container = client
	}

	go c.read()
	c.write()
	return nil
}

func newContainerClient(id string) (*websocket.Conn, error) {
	//id := "9a58ba6db179"
	u, err := url.Parse("http://komanda.io:4500/v1.5/containers/" + id + "/attach/ws?logs=1&stderr=1&stdout=1&stream=1&stdin=1")
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
