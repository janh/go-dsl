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
		OptionDescriptions: map[string]string{
			"LoadSupportData": "load more details from support data if set to 1",
			"TLSSkipVerify":   "skip verification of TLS certificates if set to 1",
		},
	}
	dsl.RegisterClient("fritzbox", newFunc, clientDesc)
}
