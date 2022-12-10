// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fritzbox

import (
	"3e8.eu/go/dsl"
)

func init() {
	newFunc := func(config dsl.Config) (dsl.Client, error) {
		clientConfig := Config{
			Host:            config.Host,
			User:            config.User,
			Password:        config.AuthPassword,
			LoadSupportData: config.Options["LoadSupportData"] == "1",
			TLSSkipVerify:   config.Options["TLSSkipVerify"] == "1",
		}
		return NewClient(clientConfig)
	}
	clientDesc := dsl.ClientDesc{
		Title:              "FRITZ!Box",
		RequiresUser:       dsl.TristateMaybe,
		SupportedAuthTypes: dsl.AuthTypePassword,
		Options: map[string]dsl.Option{
			"LoadSupportData": dsl.Option{
				Description: "load more details from support data",
				Type:        dsl.OptionTypeBool,
			},
			"TLSSkipVerify": dsl.Option{
				Description: "skip verification of TLS certificates",
				Type:        dsl.OptionTypeBool,
			},
		},
	}
	dsl.RegisterClient("fritzbox", newFunc, clientDesc)
}
