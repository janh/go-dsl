// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package mediatek

import (
	"3e8.eu/go/dsl"
)

func init() {
	newTelnet := func(config dsl.Config) (dsl.Client, error) {
		telnetConfig := TelnetConfig{
			Host:     config.Host,
			User:     config.User,
			Password: config.AuthPassword,
		}
		return NewTelnetClient(telnetConfig)
	}
	clientDescTelnet := dsl.ClientDesc{
		Title:              "MediaTek (Telnet)",
		RequiresUser:       dsl.TristateMaybe,
		SupportedAuthTypes: dsl.AuthTypePassword,
	}
	dsl.RegisterClient("mediatek_telnet", newTelnet, clientDescTelnet)

	newSSH := func(config dsl.Config) (dsl.Client, error) {
		sshConfig := SSHConfig{
			Host:               config.Host,
			User:               config.User,
			Password:           config.AuthPassword,
			PrivateKeys:        config.AuthPrivateKeys,
			KnownHosts:         config.KnownHosts,
			InsecureAlgorithms: config.Options["InsecureAlgorithms"] == "1",
		}
		return NewSSHClient(sshConfig)
	}
	clientDescSSH := dsl.ClientDesc{
		Title:              "MediaTek (SSH)",
		RequiresUser:       dsl.TristateYes,
		SupportedAuthTypes: dsl.AuthTypePassword | dsl.AuthTypePrivateKeys,
		RequiresKnownHosts: true,
		OptionDescriptions: map[string]string{
			"InsecureAlgorithms": "allow insecure SSH algorithms if set to 1",
		},
	}
	dsl.RegisterClient("mediatek_ssh", newSSH, clientDescSSH)
}
