// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"3e8.eu/go/dsl"
	jsgraphs "3e8.eu/go/dsl/graphs/javascript"

	"3e8.eu/go/dsl/cmd/web/common"
)

//go:embed static templates
var files embed.FS

var (
	c                 *common.Client
	server            http.Server
	serverErr         chan error
	shutdownReceivers map[chan bool]bool
	shutdownMutex     sync.Mutex
	config            Config
)

func Run(clientConfig dsl.Config, webConfig Config) {
	config = webConfig

	if config.ListenAddress == "" {
		config.ListenAddress = "[::1]:0"
	}

	addr, err := start(clientConfig)
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

func start(clientConfig dsl.Config) (addr string, err error) {
	http.HandleFunc("/", handleRoot)

	static := &staticHandler{}
	static.MustAddFS("/static/", files, "static")
	static.MustAdd("/static/dsl.css", staticItemFile{common.Files, "res/dsl.css"})
	static.MustAdd("/static/graphs.js", staticItemBytes{jsgraphs.Script()})
	http.Handle("/static/", static)

	http.HandleFunc("/events", handleEvents)

	http.HandleFunc("/download", handleDownload)

	if !config.DisableInteractiveAuth {
		http.HandleFunc("/password", handlePassword)
		http.HandleFunc("/passphrase", handlePassphrase)
	}

	listener, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		return
	}

	addr = "http://" + listener.Addr().String()

	c = common.NewClient(clientConfig)

	shutdownReceivers = make(map[chan bool]bool)
	server.RegisterOnShutdown(handleOnShutdown)

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
	err := <-serverErr
	c.Close()
	return err
}

func stop() {
	serverErr <- server.Shutdown(context.Background())
}

func registerOnShutdown(receiver chan bool) {
	shutdownMutex.Lock()
	defer shutdownMutex.Unlock()

	shutdownReceivers[receiver] = true
}

func unregisterOnShutdown(receiver chan bool) {
	shutdownMutex.Lock()
	defer shutdownMutex.Unlock()

	delete(shutdownReceivers, receiver)
}

func handleOnShutdown() {
	shutdownMutex.Lock()
	defer shutdownMutex.Unlock()

	for receiver := range shutdownReceivers {
		receiver <- true
	}
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

func getStateMessage(change common.StateChange) (msg common.Message) {
	switch {

	case config.HideErrorMessages && change.State == common.StateError:
		log.Println(change.Err)

		msg.State = string(change.State)
		msg.Data = "failed to load data from device: see log message for details"

	case config.DisableInteractiveAuth &&
		(change.State == common.StatePasswordRequired ||
			change.State == common.StatePassphraseRequired):

		msg.State = string(common.StateError)
		msg.Data = "failed to load data from device: interactive authentication required but not allowed"

	default:
		msg = common.GetStateMessage(change)

	}

	return
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

	registerOnShutdown(shutdown)

	defer func() {
		unregisterOnShutdown(shutdown)
		c.UnregisterReceiver(receiver)
		writer.Close()
	}()

	for {
		select {

		case change := <-receiver:
			msg := getStateMessage(change)

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

	common.WriteArchive(w, filenameBase, state, !config.HideRawData)
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
