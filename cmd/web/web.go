// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"3e8.eu/go/dsl"
	"3e8.eu/go/dsl/graphs"
	"3e8.eu/go/dsl/models"
)

//go:embed static templates
var files embed.FS

var c *client

func Start(config dsl.Config) {
	http.HandleFunc("/", handleRoot)

	staticFS := http.FS(files)
	fs := http.FileServer(staticFS)
	http.Handle("/static/", fs)

	http.HandleFunc("/events", handleEvents)

	http.HandleFunc("/password", handlePassword)
	http.HandleFunc("/passphrase", handlePassphrase)

	listener, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		fmt.Println("failed to start web server:", err)
		os.Exit(1)
	}

	fmt.Printf("Started web server at: http://%s\n", listener.Addr().String())

	c = newClient(config)

	var server http.Server

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

		c.close()
		server.Shutdown(context.Background())
	}()

	err = server.Serve(listener)

	if err != nil && err != http.ErrServerClosed {
		fmt.Println("failed to start web server:", err)
		os.Exit(1)
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

func getGraphString(bins models.Bins, graphFunc func(out io.Writer, data models.Bins, params graphs.GraphParams) error) string {
	buf := new(bytes.Buffer)

	err := graphFunc(buf, bins, graphs.DefaultGraphParams)
	if err != nil {
		return ""
	}

	return buf.String()
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

	receiver := make(chan stateChange, 10)
	c.RegisterReceiver(receiver)
	defer func() {
		c.UnregisterReceiver(receiver)
	}()

	for {
		select {

		case change := <-receiver:
			msg := message{State: string(change.State)}

			switch change.State {

			case StateReady:
				msg.Data = data{
					Summary:   change.Status.Summary(),
					GraphBits: getGraphString(change.Bins, graphs.DrawBitsGraph),
					GraphSNR:  getGraphString(change.Bins, graphs.DrawSNRGraph),
					GraphQLN:  getGraphString(change.Bins, graphs.DrawQLNGraph),
					GraphHlog: getGraphString(change.Bins, graphs.DrawHlogGraph),
				}

			case StatePassphraseRequired:
				msg.Data = change.Fingerprint

			case StateError:
				msg.Data = "failed to load data from device: " + change.Err.Error()

			}

			dataBytes, err := json.Marshal(msg)
			if err != nil {
				dataBytes = []byte(`{"state":"error","data":"encoding error"}`)
			}

			err = writer.WriteMessage(string(dataBytes))
			if err != nil {
				return
			}

		case <-req.Context().Done():
			return

		}
	}
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
