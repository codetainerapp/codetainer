package codetainer

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
)

// handlerFunc adapts a function to an http.Handler.
type handlerFunc func(ctx *Context) error

var (
	Store *sessions.CookieStore
)

func StartServer() {
	Log.Infof("Initializing %s (%s)", Name, Version)

	r := mux.NewRouter()
	r.StrictSlash(true)

	if DevMode {
		// dev
		Log.Debugf("Loading assets from disk.")
		r.PathPrefix("/public/").Handler(http.FileServer(http.Dir("./web/")))
	} else {
		Log.Debugf("Loading assets from memory.")
		r.PathPrefix("/public/").Handler(http.FileServer(
			&assetfs.AssetFS{
				Asset:    Asset,
				AssetDir: AssetDir,
				Prefix:   "web",
			},
		))
	}

	r.Handle("/", handlerFunc(RouteIndex))
	// API v1
	r.Handle("/api/v1/codetainer/{id}/view", handlerFunc(RouteApiV1CodetainerView))
	r.Handle("/api/v1/codetainer/{id}/tty", handlerFunc(RouteApiV1CodetainerTTY))
	r.Handle("/api/v1/codetainer/{id}/files", handlerFunc(RouteApiV1CodetainerListFiles))
	r.Handle("/api/v1/codetainer/{id}/attach", handlerFunc(RouteApiV1CodetainerAttach))
	r.Handle("/api/v1/codetainer/", handlerFunc(RouteApiV1Codetainer))
	r.Handle("/api/v1/codetainer/{id}/start", handlerFunc(RouteApiV1CodetainerStart))
	r.Handle("/api/v1/codetainer/{id}/stop", handlerFunc(RouteApiV1CodetainerStop))
	r.Handle("/api/v1/codetainer/images", handlerFunc(RouteApiV1CodetainerListImages))

	http.Handle("/", r)

	Store = sessions.NewCookieStore(securecookie.GenerateRandomKey(32))
	Store.Options = &sessions.Options{
		//Domain:   "localhost", // Chrome doesn't work with localhost domain
		Path:     "/",
		MaxAge:   3600 * 8, // 8 hours
		HttpOnly: true,
		Secure:   *appSSL,
	}

	var port string = fmt.Sprintf(":%d", 3000)
	var proto string = "http"

	if *appSSL {
		proto = "http"
	}

	Log.Infof("URL: %s://%s%s", proto, "127.0.0.1", port)

	var err error

	if *appSSL {
		// err = http.ListenAndServeTLS(port, KittyConfig.file.PublicKeyPath,
		// KittyConfig.file.PrivateKeyPath, context.ClearHandler(http.DefaultServeMux))
	} else {
		err = http.ListenAndServe(port, context.ClearHandler(http.DefaultServeMux))
	}

	if err != nil {
		Log.Error(err.Error())
	}

}

func (f handlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	source, err := GetRemoteAddr(r)
	var ws *websocket.Conn

	if strings.Contains(r.URL.String(), "/attach") {
		ws, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			Log.Error("Unable to upgrade websocket connection:", err)
		}
	}

	if err != nil {
		source = "N/A"
	}

	Log.Debugf("HTTP: %s %s %s", source, r.Method, r.URL)

	session, err := Store.Get(r, "codetainer")

	if err != nil {
		Log.Debug(err)
	}

	if session.IsNew {
		// Log.Debug("Create New Session")

		if err := session.Save(r, w); err != nil {
			Log.Fatal(err)
		}
	}
	//

	context := &Context{
		W:       w,
		R:       r,
		Session: session,
		WS:      ws,
		// current user stuff
	}

	err = f(context)

	if err != nil {
		Log.Error(err)
	}

}
