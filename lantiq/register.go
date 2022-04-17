// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lantiq

import (
	"3e8.eu/go/dsl"
)

func init() {
	options := map[string]string{
		"Command": "name of the dsl_cpe_pipe command on the device",
	}

	newTelnet := func(config dsl.Config) (dsl.Client, error) {
		telnetConfig := TelnetConfig{
			Host:     config.Host,
			User:     config.User,
			Password: config.AuthPassword,
			Command:  config.Options["Command"],
		}
		return NewTelnetClient(telnetConfig)
	}
	clientDescTelnet := dsl.ClientDesc{
		Title:              "Lantiq (Telnet)",
		RequiresUser:       dsl.TristateMaybe,
		SupportedAuthTypes: dsl.AuthTypePassword,
		OptionDescriptions: options,
	}
	dsl.RegisterClient("lantiq_telnet", newTelnet, clientDescTelnet)

	newSSH := func(config dsl.Config) (dsl.Client, error) {
		sshConfig := SSHConfig{
			Host:               config.Host,
			User:               config.User,
			Password:           config.AuthPassword,
			PrivateKeys:        config.AuthPrivateKeys,
			KnownHosts:         config.KnownHosts,
			InsecureAlgorithms: config.Options["InsecureAlgorithms"] == "1",
			Command:            config.Options["Command"],
		}
		return NewSSHClient(sshConfig)
	}
	clientDescSSH := dsl.ClientDesc{
		Title:              "Lantiq (SSH)",
		RequiresUser:       dsl.TristateYes,
		SupportedAuthTypes: dsl.AuthTypePassword | dsl.AuthTypePrivateKeys,
		RequiresKnownHosts: true,
		OptionDescriptions: map[string]string{
			"InsecureAlgorithms": "allow insecure SSH algorithms if set to 1",
			"Command":            options["Command"],
		},
	}
	dsl.RegisterClient("lantiq_ssh", newSSH, clientDescSSH)
}
