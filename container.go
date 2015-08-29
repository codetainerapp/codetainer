package codetainer

import (
	"net"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

func UrlEncoded(str string) (string, error) {
	u, err := url.Parse(str)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

var DockerApiVersion string = "v1.18"

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

		//	message = r.ReplaceAll(message, empty)
		message = []byte(url.QueryEscape(string(message)))

		c.web.WriteMessage(websocket.TextMessage, message)
	}
	c.container.Close()
}

func (c *ContainerConnection) Start() error {

	err := c.openSocketToContainer()

	if err == nil {
		go c.read()
		c.write()

	}
	return err
}

func (c *ContainerConnection) openSocketToContainer() error {
	id := c.id
	endpoint := GlobalConfig.GetDockerEndpoint()
	u, err := url.Parse(endpoint + "/" + DockerApiVersion + "/containers/" + id + "/attach/ws?logs=0&stderr=1&stdout=1&stream=1&stdin=1")
	if err != nil {
		return err
	}

	rawConn, err := net.Dial("tcp", u.Host)
	if err != nil {
		return err
	}

	wsHeaders := http.Header{
		"Origin": {"http://localhost:4500"},
	}

	wsConn, resp, err := websocket.NewClient(rawConn, u, wsHeaders, 1024, 1024)

	if err != nil {
		Log.Error("Unable to connect", resp, err)
	}
	c.container = wsConn
	return err
}
