// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package speedport

import (
	"3e8.eu/go/dsl"
)

func init() {
	newFunc := func(config dsl.Config) (dsl.Client, error) {
		clientConfig := Config{
			Host:          config.Host,
			Password:      config.AuthPassword,
			TLSSkipVerify: config.Options["TLSSkipVerify"] == "1",
		}
		return NewClient(clientConfig)
	}
	clientDesc := dsl.ClientDesc{
		Title:              "Speedport",
		RequiresUser:       dsl.TristateNo,
		SupportedAuthTypes: dsl.AuthTypePassword,
		Options: map[string]dsl.Option{
			"TLSSkipVerify": dsl.Option{
				Description: "skip verification of TLS certificates",
				Type:        dsl.OptionTypeBool,
			},
		},
	}
	dsl.RegisterClient("speedport", newFunc, clientDesc)
}
