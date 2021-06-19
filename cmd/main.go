// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"golang.org/x/term"

	"3e8.eu/go/dsl/broadcom"
	"3e8.eu/go/dsl/graphs"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("missing host")
		os.Exit(1)
	}

	host := os.Args[1]

	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(err)
	}
	fmt.Println()
	password := string(passwordBytes)

	fmt.Println()
	fmt.Print("Connecting…")

	telnetConfig := broadcom.TelnetConfig{
		Host:     host,
		Password: password,
	}
	client, err := broadcom.NewTelnetClient(telnetConfig)
	if err != nil {
		fmt.Println(" failed:", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println(" done")
	fmt.Print("Loading data…")

	err = client.UpdateData()
	if err != nil {
		fmt.Println(" failed", err)
		os.Exit(1)
	}

	fmt.Println(" done")
	fmt.Println()

	fmt.Println(client.Status().Summary())

	//filenameBase := time.Now().Format("dsl_20060102_150405_")
	_ = time.Now()
	filenameBase := ""

	writeFile(filenameBase+"summary.txt", []byte(client.Status().Summary()))
	writeFile(filenameBase+"raw.txt", client.RawData())

	graphParams := graphs.DefaultGraphParams

	f := createFile(filenameBase + "bits.svg")
	defer f.Close()
	graphs.DrawBitsGraph(f, client.Bins(), graphParams)

	f = createFile(filenameBase + "snr.svg")
	defer f.Close()
	graphs.DrawSNRGraph(f, client.Bins(), graphParams)

	f = createFile(filenameBase + "qln.svg")
	defer f.Close()
	graphs.DrawQLNGraph(f, client.Bins(), graphParams)

	f = createFile(filenameBase + "hlog.svg")
	defer f.Close()
	graphs.DrawHlogGraph(f, client.Bins(), graphParams)
}

func createFile(filename string) *os.File {
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println("failed to create file:", err)
		os.Exit(1)
	}
	return f
}

func writeFile(filename string, data []byte) {
	f := createFile(filename)
	defer f.Close()

	_, err := f.Write(data)
	if err != nil {
		fmt.Println("failed to write file:", err)
		os.Exit(1)
	}
}
