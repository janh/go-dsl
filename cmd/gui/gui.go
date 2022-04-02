// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build gui

package gui

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/webview/webview"

	"3e8.eu/go/dsl"

	"3e8.eu/go/dsl/cmd/web"
)

const Enabled = true

var w webview.WebView

func Run(config dsl.Config) {
	addr, err := web.Start(config)
	if err != nil {
		fmt.Println("failed to start web server:", err)
		os.Exit(1)
	}

	fmt.Printf("Started web server at: %s\n", addr)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		stopWebView()
	}()

	startWebView(addr)

	go func() {
		<-time.After(30 * time.Second)
		os.Exit(1)
	}()

	web.Stop()
}

func startWebView(addr string) {
	w = webview.New(false)
	defer w.Destroy()

	w.SetTitle("xDSL stats")
	w.SetSize(620, 600, webview.HintNone)

	w.Init(`document.addEventListener("DOMContentLoaded", function() { header = document.getElementsByTagName("header"); if (header.length) header[0].remove(); })`)

	w.Navigate(addr)

	w.Run()
}

func stopWebView() {
	if w != nil {
		w.Terminate()
	}
}
