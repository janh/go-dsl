// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"

	"3e8.eu/go/dsl"

	"3e8.eu/go/dsl/cmd/cli"
	"3e8.eu/go/dsl/cmd/config"
	"3e8.eu/go/dsl/cmd/gui"
	"3e8.eu/go/dsl/cmd/web"

	_ "3e8.eu/go/dsl/all"
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

type stringFlag struct {
	Value string
	Valid bool
}

func (s *stringFlag) String() string {
	return s.Value
}

func (s *stringFlag) Set(val string) error {
	s.Value = val
	s.Valid = true
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

	var configPath string
	flagSet.StringVar(&configPath, "config", config.DefaultConfigPath, "path to configuration file")

	var stateDir string
	flagSet.StringVar(&stateDir, "state-dir", config.DefaultStateDir, "path to state directory, set to empty string to disable persistent history")

	var secretsPath string
	flagSet.StringVar(&secretsPath, "secrets", "", "path to secrets file")

	var device stringFlag
	flagSet.Var(&device, "d", "device type (valid options: "+deviceTypeOptions+")")

	var user stringFlag
	flagSet.Var(&user, "u", "user name (optional depending on device type)")

	var options optionsFlag
	flagSet.Var(&options, "o", "device-specific option, in format Key=Value")

	privateKey := stringFlag{Value: config.DefaultPrivateKeyPath}
	flagSet.Var(&privateKey, "private-key", "private key file for SSH authentication")
	flagSet.Lookup("private-key").DefValue = privateKey.Value

	knownHosts := stringFlag{Value: config.DefaultKnownHostsPath}
	flagSet.Var(&knownHosts, "known-hosts", "known hosts file for SSH host key validation, validation is skipped if set to \"IGNORE\"")
	flagSet.Lookup("known-hosts").DefValue = knownHosts.Value

	var startWebServer bool
	flagSet.BoolVar(&startWebServer, "web", false, "start web server")
	flagSet.Lookup("web").DefValue = ""

	var startGUI bool
	if gui.Enabled {
		flagSet.BoolVar(&startGUI, "gui", false, "start graphical user interface")
		flagSet.Lookup("gui").DefValue = ""
	}

	var help bool
	flagSet.BoolVar(&help, "help", false, "print information about available options")
	flagSet.Lookup("help").DefValue = ""

	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		os.Exit(2)
	}

	if help {
		printHelp(flagSet)
		os.Exit(0)
	}

	if flagSet.NArg() > 1 {
		exitWithUsage(flagSet, "Too many arguments.")
	}

	if startWebServer && gui.Enabled && startGUI {
		exitWithUsage(flagSet, "Web interface and GUI cannot be selected together.")
	}

	err = config.Load(configPath)
	if err != nil {
		fmt.Println(err)
	}

	if secretsPath != "" {
		err = config.LoadSecrets(secretsPath)
		if err != nil {
			fmt.Println(err)
		}
	}

	if device.Valid {
		config.Config.DeviceType = dsl.ClientType(device.String())

		config.Config.User = ""
		for k := range config.Config.Options {
			delete(config.Config.Options, k)
		}
	}

	if flagSet.Arg(0) != "" {
		config.Config.Host = flagSet.Arg(0)
	}

	if user.Valid {
		config.Config.User = user.String()
	}

	if privateKey.Valid {
		config.Config.PrivateKeyPath = privateKey.String()
	}

	if knownHosts.Valid {
		config.Config.KnownHostsPath = knownHosts.String()
	}

	for k, v := range options {
		if v != "" {
			config.Config.Options[k] = v
		} else {
			delete(config.Config.Options, k)
		}
	}

	if gui.Enabled && (startGUI || len(os.Args) == 1) {
		gui.Run(stateDir)
	} else {
		err = config.Validate()
		if err != nil {
			exitWithUsage(flagSet, err.Error())
		}

		clientConfig, err := config.ClientConfig()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if startWebServer {
			web.Run(clientConfig, config.Config.Web, stateDir)
		} else {
			cli.LoadData(clientConfig)
		}
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
	fmt.Printf(`Run "%s -help" for more information.`, flagSet.Name())
	fmt.Println()
}

func printHelp(flagSet *flag.FlagSet) {
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

	fmt.Println()

	fmt.Println("Device-specific options:")
	fmt.Println()

	clientTypes := dsl.GetClientTypes()
	for _, clientType := range clientTypes {
		clientDesc := clientType.ClientDesc()
		if len(clientDesc.Options) == 0 {
			continue
		}

		fmt.Println("  " + clientType + ":")

		for key, option := range clientDesc.Options {
			desc := option.Description
			switch option.Type {
			case dsl.OptionTypeBool:
				desc += " (if set to 1)"
			case dsl.OptionTypeEnum:
				values := make([]string, 0, len(option.Values))
				for _, val := range option.Values {
					if val.Value != "" {
						values = append(values, val.Value)
					}
				}
				desc += " (valid values: " + strings.Join(values, ", ") + ")"
			}

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
