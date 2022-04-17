// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package billion

import (
	"3e8.eu/go/dsl"
)

func init() {
	options := map[string]string{
		"Command": "name of the xdslctl command on the device",
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
		Title:              "Billion (Telnet)",
		RequiresUser:       dsl.TristateMaybe,
		SupportedAuthTypes: dsl.AuthTypePassword,
		OptionDescriptions: options,
	}
	dsl.RegisterClient("billion_telnet", newTelnet, clientDescTelnet)
}
