// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package draytek

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
		Title:              "DrayTek (Telnet)",
		RequiresUser:       dsl.TristateMaybe,
		SupportedAuthTypes: dsl.AuthTypePassword,
	}
	dsl.RegisterClient("draytek_telnet", newTelnet, clientDescTelnet)

	newWeb := func(config dsl.Config) (dsl.Client, error) {
		webConfig := WebConfig{
			Host:          config.Host,
			User:          config.User,
			Password:      config.AuthPassword,
			TLSSkipVerify: config.Options["TLSSkipVerify"] == "1",
		}
		return NewWebClient(webConfig)
	}
	clientDescWeb := dsl.ClientDesc{
		Title:              "DrayTek (Web)",
		RequiresUser:       dsl.TristateMaybe,
		SupportedAuthTypes: dsl.AuthTypePassword,
		Options: map[string]dsl.Option{
			"TLSSkipVerify": dsl.Option{
				Description: "skip verification of TLS certificates",
				Type:        dsl.OptionTypeBool,
			},
		},
	}
	dsl.RegisterClient("draytek_web", newWeb, clientDescWeb)
}
