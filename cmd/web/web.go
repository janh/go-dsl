// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"context"
	"embed"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"3e8.eu/go/dsl"
	jsgraphs "3e8.eu/go/dsl/graphs/javascript"

	"3e8.eu/go/dsl/cmd/web/common"
)

//go:embed static templates
var files embed.FS

var (
	c         *common.Client
	server    http.Server
	serverErr chan error
)

func Run(config dsl.Config) {
	addr, err := start(config)
	if err != nil {
		fmt.Println("failed to start web server:", err)
		os.Exit(1)
	}

	fmt.Printf("Started web server at: %s\n", addr)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs

		fmt.Println("Stopping…")

		go func() {
			<-sigs

			fmt.Println("Stopping forced.")
			os.Exit(0)
		}()

		stop()
	}()

	err = wait()
	if err != nil {
		fmt.Println("failed to start web server:", err)
		os.Exit(1)
	}
}

func start(config dsl.Config) (addr string, err error) {
	http.HandleFunc("/", handleRoot)

	static := &staticHandler{}
	static.MustAddFS("/static/", files, "static")
	static.MustAdd("/static/dsl.css", staticItemFile{common.Files, "res/dsl.css"})
	static.MustAdd("/static/graphs.js", staticItemBytes{jsgraphs.Script()})
	http.Handle("/static/", static)

	http.HandleFunc("/events", handleEvents)

	http.HandleFunc("/download", handleDownload)

	http.HandleFunc("/password", handlePassword)
	http.HandleFunc("/passphrase", handlePassphrase)

	listener, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		return
	}

	addr = "http://" + listener.Addr().String()

	c = common.NewClient(config)

	serverErr = make(chan error, 1)

	go func() {
		err := server.Serve(listener)
		if err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	return
}

func wait() error {
	return <-serverErr
}

func stop() {
	c.Close()
	serverErr <- server.Shutdown(context.Background())
}

func handleRoot(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}

	if req.Method != http.MethodGet {
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-type", "text/html; charset=utf-8")

	data, _ := files.ReadFile("templates/index.html")
	w.Write(data)
}

func handleEvents(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writer, err := newEventStreamWriter(w, req)
	if err != nil {
		http.Error(w, "500 internal server error", http.StatusInternalServerError)
		return
	}

	receiver := make(chan common.StateChange, 10)
	c.RegisterReceiver(receiver)

	shutdown := make(chan bool, 1)

	server.RegisterOnShutdown(func() {
		shutdown <- true
	})

	defer func() {
		c.UnregisterReceiver(receiver)
		writer.Close()
	}()

	for {
		select {

		case change := <-receiver:
			msg := common.GetStateMessage(change)

			err = writer.WriteMessage(string(msg.JSON()))
			if err != nil {
				return
			}

		case <-req.Context().Done():
			return

		case <-shutdown:
			return

		}
	}
}

func handleDownload(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}

	state := c.State()
	if state.State != common.StateReady {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}

	filenameBase := state.Time.Format("dsl_20060102_150405")

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filenameBase+`.zip"`)
	w.Header().Set("Cache-Control", "no-cache")

	common.WriteArchive(w, filenameBase, state)
}

func handlePassword(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}

	password := req.PostFormValue("data")

	err := c.SetPassword(password)
	if err != nil {
		http.Error(w, "403 forbidden", http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handlePassphrase(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}

	passphrase := req.PostFormValue("data")

	err := c.SetPassphrase(passphrase)
	if err != nil {
		http.Error(w, "403 forbidden", http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
