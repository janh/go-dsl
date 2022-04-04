// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"archive/zip"
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"3e8.eu/go/dsl"
	"3e8.eu/go/dsl/graphs"
	jsgraphs "3e8.eu/go/dsl/graphs/javascript"
	"3e8.eu/go/dsl/models"
)

//go:embed static templates
var files embed.FS

var (
	c         *client
	server    http.Server
	serverErr chan error
)

func Run(config dsl.Config) {
	addr, err := Start(config)
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

		Stop()
	}()

	err = Wait()
	if err != nil {
		fmt.Println("failed to start web server:", err)
		os.Exit(1)
	}
}

func Start(config dsl.Config) (addr string, err error) {
	http.HandleFunc("/", handleRoot)

	static := &staticHandler{}
	static.MustAddFS("/static/", files, "static")
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

	c = newClient(config)

	serverErr = make(chan error, 1)

	go func() {
		err := server.Serve(listener)
		if err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	return
}

func Wait() error {
	return <-serverErr
}

func Stop() {
	c.close()
	serverErr <- server.Shutdown(context.Background())
}

func handleGraphsScript(w http.ResponseWriter, req *http.Request) {
	reader := bytes.NewReader(jsgraphs.Script())
	http.ServeContent(w, req, "graphs.js", time.Time{}, reader)
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

func getSummaryString(status models.Status) string {
	buf := new(bytes.Buffer)

	tpl := template.Must(template.ParseFS(files, "templates/summary.html"))
	tpl.Execute(buf, status)

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
			msg := message{State: string(change.State)}

			switch change.State {

			case StateReady:
				msg.Data = data{
					Summary: getSummaryString(change.Status),
					Bins:    jsgraphs.EncodeBins(change.Bins),
					History: jsgraphs.EncodeBinsHistory(change.BinsHistory),
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
	if state.State != StateReady {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}

	filenameBase := state.Time.Format("dsl_20060102_150405")

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filenameBase+`.zip"`)
	w.Header().Set("Cache-Control", "no-cache")

	archive := zip.NewWriter(w)

	fileWriter, _ := archive.Create(filenameBase + "_summary.txt")
	io.WriteString(fileWriter, state.Status.Summary())

	fileWriter, _ = archive.Create(filenameBase + "_raw.txt")
	fileWriter.Write(state.RawData)

	fileWriter, _ = archive.Create(filenameBase + "_bits.svg")
	graphs.DrawBitsGraph(fileWriter, state.Bins, graphs.DefaultGraphParams)

	fileWriter, _ = archive.Create(filenameBase + "_snr.svg")
	graphs.DrawSNRGraph(fileWriter, state.Bins, graphs.DefaultGraphParams)

	fileWriter, _ = archive.Create(filenameBase + "_snr_minmax.svg")
	graphs.DrawSNRGraphWithHistory(fileWriter, state.Bins, state.BinsHistory, graphs.DefaultGraphParams)

	fileWriter, _ = archive.Create(filenameBase + "_qln.svg")
	graphs.DrawQLNGraph(fileWriter, state.Bins, graphs.DefaultGraphParams)

	fileWriter, _ = archive.Create(filenameBase + "_hlog.svg")
	graphs.DrawHlogGraph(fileWriter, state.Bins, graphs.DefaultGraphParams)

	archive.Close()
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
