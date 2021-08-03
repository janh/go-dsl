// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"

	"3e8.eu/go/dsl"
	"3e8.eu/go/dsl/graphs"
	"3e8.eu/go/dsl/models"

	_ "3e8.eu/go/dsl/broadcom"
	_ "3e8.eu/go/dsl/draytek"
	_ "3e8.eu/go/dsl/lantiq"
)

type optionsFlag map[string]string

func (o *optionsFlag) String() string {
	// not needed for flag parsing
	return ""
}

func (o *optionsFlag) Set(val string) error {
	if *o == nil {
		*o = make(map[string]string)
	}

	valSplit := strings.SplitN(val, "=", 2)
	if len(valSplit) != 2 {
		return errors.New("invalid format for device specific option")
	}

	(*o)[valSplit[0]] = valSplit[1]
	return nil
}

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

	var options optionsFlag

	device := flagSet.String("d", "", "device type (valid options: "+deviceTypeOptions+")")
	user := flagSet.String("u", "", "user name (optional depending on device type)")
	flagSet.Var(&options, "o", "device-specific option, in format Key=Value")
	privateKey := flagSet.String("private-key", "", "private key file for SSH authentication (by default, the system keys are used)")
	knownHosts := flagSet.String("known-hosts", "", "known hosts file for SSH host key validation, validation is skipped if set to \"IGNORE\" (by default, the system file is used)")
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

	for optionKey := range options {
		valid := false
		for option := range clientDesc.OptionDescriptions {
			if optionKey == option {
				valid = true
				break
			}
		}
		if !valid {
			exitWithUsage(flagSet, "Invalid device-specific option: "+optionKey)
		}
	}

	loadData(clientType, flagSet.Arg(0), *user, *privateKey, *knownHosts, options)
}

func printUsage(flagSet *flag.FlagSet) {
	fmt.Println("\nUsage:")
	fmt.Printf("  %s -d device [options] hostname\n\n", flagSet.Name())

	fmt.Println("List of options:")
	flagSet.PrintDefaults()
	fmt.Println()

	fmt.Println("Device-specific options:")
	fmt.Println()

	clientTypes := dsl.GetClientTypes()
	for _, clientType := range clientTypes {
		clientDesc := clientType.ClientDesc()
		if len(clientDesc.OptionDescriptions) == 0 {
			continue
		}

		fmt.Println("  " + clientType + ":")

		for key, desc := range clientDesc.OptionDescriptions {
			fmt.Println("    " + key)
			fmt.Println("\t" + desc)
		}

		fmt.Println()
	}
}

func exitWithUsage(flagSet *flag.FlagSet, message string) {
	fmt.Println(message)
	flagSet.Usage()
	os.Exit(2)
}

func loadKnownHosts(file string) (string, error) {
	if file == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", nil
		}

		file = filepath.Join(home, ".ssh", "known_hosts")
		data, _ := os.ReadFile(file)
		return string(data), nil
	}

	data, err := os.ReadFile(file)
	return string(data), err
}

func loadPrivateKeys(file string) ([]string, error) {
	if file == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, nil
		}

		var keys []string

		keyFileNames := []string{"id_ed25519", "id_rsa", "id_ecdsa"}
		for _, name := range keyFileNames {
			file = filepath.Join(home, ".ssh", name)

			data, err := os.ReadFile(file)
			if err == nil {
				keys = append(keys, string(data))
			}
		}

		return keys, nil
	}

	data, err := os.ReadFile(file)
	return []string{string(data)}, err
}

func readPassword(prompt string) string {
	fmt.Print(prompt)
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(err)
	}
	fmt.Println()

	return string(passwordBytes)
}

func loadData(clientType dsl.ClientType, host, user, privateKey, knownHosts string, options map[string]string) {
	clientDesc := clientType.ClientDesc()

	if clientDesc.RequiresKnownHosts {
		if knownHosts == "IGNORE" {
			fmt.Println("WARNING: Host key validation disabled!")
		} else {
			var err error
			knownHosts, err = loadKnownHosts(knownHosts)
			if err != nil {
				fmt.Println("failed to load known hosts file:", err)
				os.Exit(1)
			}
		}
	}

	fmt.Println()
	fmt.Print("Connecting…")

	var passwordCallback dsl.PasswordCallback
	if clientDesc.SupportedAuthTypes&dsl.AuthTypePassword != 0 {
		passwordCallback = func() string {
			fmt.Println(" password required")
			password := readPassword("Password: ")
			fmt.Print("Authenticating…")
			return password
		}
	}

	var privateKeysCallback dsl.PrivateKeysCallback
	if clientDesc.SupportedAuthTypes&dsl.AuthTypePrivateKeys != 0 {
		privateKeysCallback.Keys = func() []string {
			keys, err := loadPrivateKeys(privateKey)
			if err != nil {
				fmt.Println("failed to load private key file:", err)
				os.Exit(1)
			}
			return keys
		}

		privateKeysCallback.Passphrase = func(fingerprint string) string {
			fmt.Println(" passphrase required")
			fmt.Println("Fingerprint: " + fingerprint)
			passphrase := readPassword("Passphrase: ")
			fmt.Print("Authenticating…")
			return passphrase
		}
	}

	config := dsl.Config{
		Type:            clientType,
		Host:            host,
		User:            user,
		AuthPassword:    passwordCallback,
		AuthPrivateKeys: privateKeysCallback,
		KnownHosts:      knownHosts,
		Options:         options,
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
