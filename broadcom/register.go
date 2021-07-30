// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package broadcom

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
		RequiresUser:       dsl.TristateMaybe,
		SupportedAuthTypes: dsl.AuthTypePassword,
	}
	dsl.RegisterClient("broadcom_telnet", newTelnet, clientDescTelnet)

	newSSH := func(config dsl.Config) (dsl.Client, error) {
		sshConfig := SSHConfig{
			Host:       config.Host,
			User:       config.User,
			Password:   config.AuthPassword,
			PrivateKey: config.AuthPrivateKey,
			KnownHost:  config.KnownHost,
		}
		return NewSSHClient(sshConfig)
	}
	clientDescSSH := dsl.ClientDesc{
		RequiresUser:       dsl.TristateYes,
		SupportedAuthTypes: dsl.AuthTypePassword | dsl.AuthTypePrivateKey,
		RequiresKnownHost:  true,
	}
	dsl.RegisterClient("broadcom_ssh", newSSH, clientDescSSH)
}
