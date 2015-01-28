package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
)

type Embryo struct {
	serverHost string
	port       int
	proxy      *httputil.ReverseProxy
}

func NewEmbryo() (embryo *Embryo) {
	addr := "localhost"
	port := 8888
	scheme := "http"

	serverUrl, err := url.ParseRequestURI(fmt.Sprintf(scheme+"://%s:%d", addr, port))
	logE(err)

	embryo = &Embryo{
		serverHost: serverUrl.String()[len(scheme+"://"):],
		port:       port,
		proxy:      httputil.NewSingleHostReverseProxy(serverUrl),
	}

	return
}

func (e *Embryo) Run() {
	go func() {
		err := http.ListenAndServe(e.serverHost, e)
		logE(err)
	}()

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill)
	<-ch

	os.Exit(1)
}

func (e *Embryo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.proxy.ServeHTTP(w, r)
}

func getFreePort() (port int) {
	conn, err := net.Listen("tcp", ":80")
	logE(err)

	port = conn.Addr().(*net.TCPAddr).Port
	logE(conn.Close())
	return
}
