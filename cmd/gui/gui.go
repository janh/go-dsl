// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build gui

package gui

import (
	"bytes"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/adrg/xdg"
	"github.com/webview/webview"

	"3e8.eu/go/dsl"
	jsgraphs "3e8.eu/go/dsl/graphs/javascript"

	"3e8.eu/go/dsl/cmd/config"
	"3e8.eu/go/dsl/cmd/web/common"
)

const Enabled = true

const (
	stateInitializing  = "initializing"
	stateDisconnecting = "disconnecting"
	stateConnect       = "connect"
)

//go:embed res
var resources embed.FS

var (
	c              *common.Client
	w              webview.WebView
	startReceive   chan bool
	startDone      chan bool
	stopReceive    chan bool
	stopDone       chan bool
	inhibitReceive chan bool
	isInitialized  bool
	lastMessage    common.Message
	mutex          sync.Mutex
	mutexClient    sync.Mutex
)

func Run() {
	updateState(common.Message{State: stateConnect})

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs

		mutex.Lock()
		defer mutex.Unlock()

		stopWebView()
	}()

	initReceive()
	startWebView()

	go func() {
		<-time.After(30 * time.Second)
		os.Exit(1)
	}()

	clientStopEvents()
	clientDisconnect()

	os.Exit(0)
}

func clientConnect(clientConfig dsl.Config) {
	mutexClient.Lock()
	defer mutexClient.Unlock()

	c = common.NewClient(clientConfig)

	startReceive <- true
	<-startDone
}

func clientStopEvents() {
	mutexClient.Lock()
	defer mutexClient.Unlock()

	stopReceive <- true
	<-stopDone
}

func clientDisconnect() {
	mutexClient.Lock()
	defer mutexClient.Unlock()

	if c != nil {
		c.Close()
		c = nil
	}
}

func getMainDataURI() string {
	style, _ := resources.ReadFile("res/style.css")
	styleDSL, _ := common.Files.ReadFile("res/dsl.css")

	data := map[string]interface{}{
		"Style":    template.CSS(style),
		"StyleDSL": template.CSS(styleDSL),
		"Script":   template.JS(jsgraphs.Script()),
	}

	buf := new(bytes.Buffer)
	tpl := template.Must(template.ParseFS(resources, "res/main.html"))
	tpl.Execute(buf, data)

	return "data:text/html;charset=utf-8;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

func initReceive() {
	startReceive = make(chan bool)
	startDone = make(chan bool)

	stopReceive = make(chan bool)
	stopDone = make(chan bool)

	inhibitReceive = make(chan bool)

	go receive()
}

func receive() {
	enabled := false
	inhibited := false

	receiver := make(chan common.StateChange, 10)

	for {
		select {

		case change := <-receiver:
			mutex.Lock()
			updateState(common.GetStateMessage(change))
			mutex.Unlock()

		case <-startReceive:
			if !enabled && !inhibited {
				c.RegisterReceiver(receiver)
				enabled = true
			}
			startDone <- true

		case <-stopReceive:
			if enabled && !inhibited {
				c.UnregisterReceiver(receiver)
				enabled = false
			}
			stopDone <- true

		case inhibit := <-inhibitReceive:
			if inhibit {
				if enabled {
					c.UnregisterReceiver(receiver)
				}
				inhibited = true
			} else {
				if enabled {
					c.RegisterReceiver(receiver)
				}
				inhibited = false
			}

		}
	}
}

func updateState(msg common.Message) {
	lastMessage = msg

	if !isInitialized {
		return
	}

	w.Dispatch(func() {
		w.Eval("updateState(" + string(msg.JSON()) + ")")
	})
}

func showMessage(msg string) {
	msgJSON, _ := json.Marshal(msg)

	if !isInitialized {
		return
	}

	w.Dispatch(func() {
		w.Eval("showMessage(" + string(msgJSON) + ")")
	})
}

func setConfig() {
	configJSON, _ := json.Marshal(config.Config)

	clients := make(map[string]dsl.ClientDesc)
	for _, clientType := range dsl.GetClientTypes() {
		clients[string(clientType)] = clientType.ClientDesc()
	}
	clientsJSON, _ := json.Marshal(clients)

	w.Dispatch(func() {
		w.Eval("setConfig(" + string(configJSON) + ", " + string(clientsJSON) + ")")
	})
}

func initialized() {
	mutex.Lock()
	defer mutex.Unlock()

	isInitialized = true

	setConfig()
	updateState(lastMessage)
}

func visibilityChanged(visible bool) {
	mutex.Lock()
	defer mutex.Unlock()

	inhibitReceive <- !visible
}

func writeArchive(state common.StateChange) (path string, err error) {
	filenameBase := state.Time.Format("dsl_20060102_150405")
	filename := filenameBase + ".zip"

	var paths = []string{
		filepath.Join(xdg.UserDirs.Download, filename),
		filepath.Join(xdg.Home, filename),
	}

	var f *os.File

	for i := range paths {
		path = paths[i]
		f, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)

		if i < len(paths)-1 && errors.Is(err, fs.ErrNotExist) {
			continue
		} else if err != nil {
			return
		} else {
			break
		}
	}

	defer f.Close()

	err = common.WriteArchive(f, filenameBase, state, true)
	return
}

func save() {
	mutex.Lock()
	defer mutex.Unlock()

	if lastMessage.State != string(common.StateReady) {
		return
	}

	change := c.State()
	if change.State != common.StateReady {
		return
	}

	path, err := writeArchive(change)

	path, _ = filepath.Abs(path)

	if err == nil {
		showMessage(fmt.Sprintf("Saved to %s.", path))
	} else if errors.Is(err, fs.ErrExist) {
		showMessage(fmt.Sprintf("File %s already exists.", path))
	} else if err != nil {
		showMessage("Saving failed!")
		fmt.Println("failed to save archive:", err)
	}
}

func setPassword(data string) {
	mutex.Lock()
	defer mutex.Unlock()

	if lastMessage.State != string(common.StatePasswordRequired) {
		return
	}

	err := c.SetPassword(data)
	if err != nil {
		fmt.Println("setting password failed:", err)
	}
}

func setPassphrase(data string) {
	mutex.Lock()
	defer mutex.Unlock()

	if lastMessage.State != string(common.StatePassphraseRequired) {
		return
	}

	err := c.SetPassphrase(data)
	if err != nil {
		fmt.Println("setting passphrase failed:", err)
	}
}

func setEncryptionPassphrase(data string) {
	mutex.Lock()
	defer mutex.Unlock()

	if lastMessage.State != string(common.StateEncryptionPassphraseRequired) {
		return
	}

	err := c.SetEncryptionPassphrase(data)
	if err != nil {
		fmt.Println("setting encryption passphrase failed:", err)
	}
}

func connect(cfg json.RawMessage, remember bool) {
	mutex.Lock()
	defer mutex.Unlock()

	if lastMessage.State != stateConnect {
		return
	}

	config.Config.Options = make(map[string]string)
	err := json.Unmarshal(cfg, &config.Config)
	if err != nil {
		msg := common.Message{
			State: string(common.StateError),
			Data:  "parsing error: " + err.Error(),
		}
		updateState(msg)
		return
	}

	err = config.Validate()
	if err != nil {
		msg := common.Message{
			State: string(common.StateError),
			Data:  "configuration error: " + err.Error(),
		}
		updateState(msg)
		return
	}

	clientConfig, err := config.ClientConfig()
	if err != nil {
		msg := common.Message{
			State: string(common.StateError),
			Data:  "client error: " + err.Error(),
		}
		updateState(msg)
		return
	}

	updateState(common.Message{State: string(common.StateLoading)})

	clientConnect(clientConfig)

	if remember {
		err = config.Save()
		if err != nil {
			showMessage("Saving configuration failed!")
			fmt.Println("failed to save config:", err)
		}
	}
}

func disconnect() {
	mutex.Lock()
	defer mutex.Unlock()

	if lastMessage.State != string(common.StateReady) &&
		lastMessage.State != string(common.StatePasswordRequired) &&
		lastMessage.State != string(common.StatePassphraseRequired) &&
		lastMessage.State != string(common.StateEncryptionPassphraseRequired) &&
		lastMessage.State != string(common.StateError) &&
		lastMessage.State != string(common.StateLoading) {

		return
	}

	clientStopEvents()

	updateState(common.Message{State: stateDisconnecting})

	go func() {
		clientDisconnect()

		mutex.Lock()
		defer mutex.Unlock()

		updateState(common.Message{State: stateConnect})
	}()
}

func startWebView() {
	w = webview.New(false)
	defer w.Destroy()

	w.SetTitle("xDSL stats")

	w.SetSize(620, 300, webview.HintMin)
	w.SetSize(620, 600, webview.HintNone)

	w.Bind("goInitialized", initialized)
	w.Bind("goVisibilityChanged", visibilityChanged)
	w.Bind("goSave", save)
	w.Bind("goSetPassword", setPassword)
	w.Bind("goSetPassphrase", setPassphrase)
	w.Bind("goSetEncryptionPassphrase", setEncryptionPassphrase)
	w.Bind("goConnect", connect)
	w.Bind("goDisconnect", disconnect)

	script, _ := resources.ReadFile("res/script.js")
	w.Init(string(script))

	w.Navigate(getMainDataURI())

	w.Run()
}

func stopWebView() {
	if w != nil {
		isInitialized = false
		w.Terminate()
	}
}
