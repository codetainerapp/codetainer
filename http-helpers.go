package main

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"text/template"

	"github.com/dustin/go-humanize"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
)

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

var funcs template.FuncMap = map[string]interface{}{
	"DateFormat": DateFormat,
	"PrettyNumber": func(number int64) string {
		return humanize.Comma(number)
	},
}

func jsonError(error_message error, w http.ResponseWriter) error {
	return renderJson(map[string]interface{}{
		"error":   true,
		"message": error_message.Error(),
	}, w)
}

func renderJson(data interface{}, w http.ResponseWriter) error {
	js, err := json.Marshal(data)

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

	return nil
}

func newTemplate(filename string, includeLayout bool) *template.Template {
	var file []byte
	var err error

	var base []byte
	var helpers []byte

	if DevMode {
		file, err = ioutil.ReadFile("web/" + filename)
		base, err = ioutil.ReadFile("web/layout.html")
		helpers, err = ioutil.ReadFile("web/helpers.html")
	} else {
		file, err = Asset("web/" + filename)
		base, err = Asset("web/layout.html")
		helpers, err = Asset("web/helpers.html")
	}

	if err != nil {
		Log.Error(err)
	}

	var layout string

	if includeLayout {
		layout = string(base) + string(helpers) + string(file)
	} else {
		layout = string(helpers) + string(file)
	}

	return template.Must(template.New("*").Delims("<%", "%>").Funcs(funcs).Parse(layout))
}

type Context struct {
	Session *sessions.Session
	W       http.ResponseWriter
	R       *http.Request
	WS      *websocket.Conn
}

func executeTemplate(ctx *Context, name string, status int, data interface{}) error {
	ctx.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	ctx.W.WriteHeader(status)

	return newTemplate(name, true).Execute(ctx.W, data)
}

func executeRaw(ctx *Context, name string, status int, data interface{}) error {
	ctx.W.Header().Set("Content-Type", "text/html; charset=utf-8")
	ctx.W.WriteHeader(status)

	return newTemplate(name, false).Execute(ctx.W, data)
}

func GetRemoteAddr(req *http.Request) (string, error) {

	if forwardedFor := req.Header.Get("X-FORWARDED-FOR"); forwardedFor != "" {
		if ipParsed := net.ParseIP(forwardedFor); ipParsed != nil {
			return ipParsed.String(), nil
		}
	}

	ip, _, err := net.SplitHostPort(req.RemoteAddr)

	if err != nil {
		return "", err
	}

	return ip, nil
}
