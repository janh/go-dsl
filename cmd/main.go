// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"

	"golang.org/x/term"

	"3e8.eu/go/dsl"
	"3e8.eu/go/dsl/graphs"
	"3e8.eu/go/dsl/models"

	_ "3e8.eu/go/dsl/broadcom"
	_ "3e8.eu/go/dsl/draytek"
)

func main() {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagSet.Usage = func() { printUsage(flagSet) }

	clientTypes := dsl.GetClientTypes()
	deviceTypeOptions := ""
	for i, clientType := range clientTypes {
		if i != 0 {
			deviceTypeOptions += ", "
		}
		deviceTypeOptions += string(clientType)
	}

	device := flagSet.String("d", "", "device type (valid options: "+deviceTypeOptions+")")
	user := flagSet.String("u", "", "user name (optional depending on device type)")
	flagSet.Parse(os.Args[1:])

	clientType := dsl.ClientType(*device)
	if !clientType.IsValid() {
		exitWithUsage(flagSet, "Invalid or missing device type.")
	}
	clientDesc := clientType.ClientDesc()

	if flagSet.NArg() == 0 {
		exitWithUsage(flagSet, "No hostname specified.")
	} else if flagSet.NArg() > 1 {
		exitWithUsage(flagSet, "Too many arguments.")
	}

	if clientDesc.RequiresUser == dsl.TristateNo && *user != "" {
		exitWithUsage(flagSet, "Username specified, but not required for device.")
	} else if clientDesc.RequiresUser == dsl.TristateYes && *user == "" {
		exitWithUsage(flagSet, "No username specified.")
	}

	loadData(clientType, flagSet.Arg(0), *user)
}

func printUsage(flagSet *flag.FlagSet) {
	fmt.Println("\nUsage:")
	fmt.Printf("  %s -d device [options] hostname\n\n", flagSet.Name())

	fmt.Println("List of options:")
	flagSet.PrintDefaults()

	fmt.Println()
}

func exitWithUsage(flagSet *flag.FlagSet, message string) {
	fmt.Println(message)
	flagSet.Usage()
	os.Exit(2)
}

func loadData(clientType dsl.ClientType, host, user string) {
	clientDesc := clientType.ClientDesc()

	var knownHost string
	if clientDesc.RequiresKnownHost {
		fmt.Println("Warning: host key validation is not yet implemented!\n")
		knownHost = "IGNORE"
	}

	var password string
	if clientDesc.SupportedAuthTypes&dsl.AuthTypePassword != 0 {
		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			panic(err)
		}
		fmt.Println()
		password = string(passwordBytes)
	}

	fmt.Println()
	fmt.Print("Connecting…")

	config := dsl.Config{
		Type:         clientType,
		Host:         host,
		User:         user,
		AuthPassword: password,
		KnownHost:    knownHost,
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
		fmt.Println(" failed", err)
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
