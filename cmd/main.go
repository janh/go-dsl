// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"3e8.eu/go/dsl"

	"3e8.eu/go/dsl/cmd/cli"
	"3e8.eu/go/dsl/cmd/web"

	_ "3e8.eu/go/dsl/all"
)

var (
	defaultPrivateKey string
	defaultKnownHosts string
)

func init() {
	home, err := os.UserHomeDir()
	if err == nil {
		defaultPrivateKey = filepath.Join(home, ".ssh") + string(filepath.Separator)
		defaultKnownHosts = filepath.Join(home, ".ssh", "known_hosts")
	}
}

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
	flagSet := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
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
	privateKey := flagSet.String("private-key", defaultPrivateKey, "private key file for SSH authentication")
	knownHosts := flagSet.String("known-hosts", defaultKnownHosts, "known hosts file for SSH host key validation, validation is skipped if set to \"IGNORE\"")
	startWebServer := flagSet.Bool("web", false, "start web server instead of printing result")

	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		if err == flag.ErrHelp {
			printHelp()
			os.Exit(0)
		} else {
			os.Exit(2)
		}
	}

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

	config := buildClientConfig(clientType, flagSet.Arg(0), *user, *privateKey, *knownHosts, options)

	if *startWebServer {
		web.Run(config)
	} else {
		cli.LoadData(config)
	}
}

func wordWrap(maxLength int, str string) string {
	runes := []rune(str)

	i := 0
	nextWord := func() (space, word []rune, linebreaks int, ok bool) {
		if i == len(runes) {
			return nil, nil, 0, false
		}

		start := i
		firstNonSpace := -1
		hasNonSpace := false

		for ; i < len(runes); i++ {
			r := runes[i]

			isNonBreakingSpace := r == '\u00A0' || r == '\u2007' || r == '\u202F'
			isSpace := unicode.IsSpace(r) && !isNonBreakingSpace
			isLinebreak := r == '\n'

			if hasNonSpace && isSpace {
				break
			}

			if isLinebreak {
				linebreaks++
			}

			if !isSpace && !hasNonSpace {
				hasNonSpace = true
				firstNonSpace = i
			}
		}

		if firstNonSpace == -1 {
			firstNonSpace = i
		}

		return runes[start:firstNonSpace], runes[firstNonSpace:i], linebreaks, true
	}

	var b strings.Builder
	var line []rune

	for {
		space, word, linebreaks, ok := nextWord()
		if !ok {
			break
		}

		// this assumes every rune has a width of 1
		if (len(line) != 0 && len(word) != 0 && len(line)+len(space)+len(word) > maxLength) || linebreaks != 0 {
			fmt.Fprintln(&b, string(line))
			line = []rune{}

			if linebreaks > 1 {
				for i := 1; i < linebreaks; i++ {
					fmt.Fprintln(&b)
				}
			}
		}

		if linebreaks == 0 && len(line) != 0 && len(word) != 0 {
			line = append(line, space...)
		}
		line = append(line, word...)
	}

	fmt.Fprint(&b, string(line))

	return b.String()
}

func indentAndWordWrap(str string) string {
	wordWrappedStr := wordWrap(64, str)
	return "\t" + strings.ReplaceAll(wordWrappedStr, "\n", "\n\t")
}

func printUsage(flagSet *flag.FlagSet) {
	fmt.Print("\nUsage:")
	if len(flagSet.Name()) > 20 {
		fmt.Println()
	}
	fmt.Printf("  %s -d device [options] hostname\n\n", flagSet.Name())

	fmt.Println("List of options:")
	fmt.Println()

	var flags []*flag.Flag
	flagSet.VisitAll(func(flag *flag.Flag) {
		flags = append(flags, flag)
	})

	helpFlag := flag.Flag{
		Name:  "help",
		Usage: "print information about available options",
	}
	flags = append(flags, &helpFlag)

	sort.Slice(flags, func(i, j int) bool {
		return flags[i].Name < flags[j].Name
	})

	for _, flag := range flags {
		fmt.Print("  -" + flag.Name)
		if len(flag.Name) > 1 {
			fmt.Println()
		}

		fmt.Println(indentAndWordWrap(flag.Usage))
		if flag.DefValue != "" {
			fmt.Println(indentAndWordWrap("(default: " + flag.DefValue + ")"))
		}

		fmt.Println()
	}
}

func printHelp() {
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
			fmt.Println(indentAndWordWrap(desc))
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
		return "", nil
	}

	data, err := os.ReadFile(file)
	if err != nil && file != defaultKnownHosts {
		return "", err
	}

	return string(data), nil
}

func loadPrivateKeys(file string) ([]string, error) {
	if file == "" {
		return nil, nil
	}

	if file[len(file)-1] == filepath.Separator {
		var keys []string

		keyFileNames := []string{"id_ed25519", "id_rsa", "id_ecdsa"}
		for _, name := range keyFileNames {
			data, err := os.ReadFile(file + name)
			if err != nil && file != defaultPrivateKey {
				return []string{}, err
			}

			if err == nil {
				keys = append(keys, string(data))
			}
		}

		return keys, nil
	}

	data, err := os.ReadFile(file)
	if err != nil && file != defaultPrivateKey {
		return []string{}, err
	}

	return []string{string(data)}, nil
}

func buildClientConfig(clientType dsl.ClientType, host, user, privateKey, knownHosts string, options map[string]string) dsl.Config {
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
	}

	config := dsl.Config{
		Type:            clientType,
		Host:            host,
		User:            user,
		AuthPrivateKeys: privateKeysCallback,
		KnownHosts:      knownHosts,
		Options:         options,
	}

	return config
}
