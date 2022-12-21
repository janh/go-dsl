// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"fmt"
	"io"
	"os"
	"syscall"
	"time"

	"golang.org/x/term"

	"3e8.eu/go/dsl"
	"3e8.eu/go/dsl/graphs"
	"3e8.eu/go/dsl/models"
)

func readPassword(prompt string) string {
	fmt.Print(prompt)
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(err)
	}
	fmt.Println()

	return string(passwordBytes)
}

func LoadData(config dsl.Config) {
	clientDesc := config.Type.ClientDesc()

	fmt.Println()
	fmt.Print("Connecting…")

	if clientDesc.SupportedAuthTypes&dsl.AuthTypePassword != 0 && config.AuthPassword == nil {
		config.AuthPassword = func() (string, error) {
			fmt.Println(" password required")
			password := readPassword("Password: ")
			fmt.Print("Authenticating…")
			return password, nil
		}
	}

	if clientDesc.SupportedAuthTypes&dsl.AuthTypePrivateKeys != 0 && config.AuthPrivateKeys.Passphrase == nil {
		config.AuthPrivateKeys.Passphrase = func(fingerprint string) (string, error) {
			fmt.Println(" passphrase required")
			fmt.Println("Fingerprint: " + fingerprint)
			passphrase := readPassword("Passphrase: ")
			fmt.Print("Authenticating…")
			return passphrase, nil
		}
	}

	client, err := dsl.NewClient(config)
	if err != nil {
		fmt.Println(" failed:", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println(" done")
	fmt.Print("Loading data…")

	err = client.UpdateData()
	if err != nil {
		fmt.Println(" failed:", err)
		os.Exit(1)
	}

	fmt.Println(" done")
	fmt.Println()

	fmt.Println(client.Status().Summary())

	filenameBase := time.Now().Format("dsl_20060102_150405_")

	writeFile(filenameBase+"summary.txt", []byte(client.Status().Summary()))
	writeFile(filenameBase+"raw.txt", client.RawData())

	writeGraph(filenameBase+"bits.svg", client.Bins(), graphs.DrawBitsGraph)
	writeGraph(filenameBase+"snr.svg", client.Bins(), graphs.DrawSNRGraph)
	writeGraph(filenameBase+"qln.svg", client.Bins(), graphs.DrawQLNGraph)
	writeGraph(filenameBase+"hlog.svg", client.Bins(), graphs.DrawHlogGraph)
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

func writeGraph(filename string, bins models.Bins, graphFunc func(out io.Writer, data models.Bins, params graphs.GraphParams) error) {
	f := createFile(filename)
	defer f.Close()

	err := graphFunc(f, bins, graphs.DefaultGraphParams)
	if err != nil {
		fmt.Println("failed to write graph:", err)
		os.Exit(1)
	}
}
