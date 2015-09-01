package codetainer

import (
	"bytes"
	"net"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	docker "github.com/jandre/go-dockerclient"
)

var DockerApiVersion string = "1.17"

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

func UrlEncoded(str string) (string, error) {
	u, err := url.Parse(str)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

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

func (c *ContainerConnection) SendSingleMessage(msg string) error {

	err := c.openSocketToContainer()

	if err != nil {
		return err
	}

	err = c.container.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		return err
	}

	err = c.container.Close()
	return err
}

func (c *ContainerConnection) openSocketToContainer() error {
	id := c.id
	endpoint := GlobalConfig.GetDockerEndpoint()
	u, err := url.Parse(endpoint + "/v" + DockerApiVersion + "/containers/" + id + "/attach/ws?logs=0&stderr=1&stdout=1&stream=1&stdin=1")
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
