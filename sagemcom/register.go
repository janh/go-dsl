// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package sagemcom

import (
	"3e8.eu/go/dsl"
)

func init() {
	newFunc := func(config dsl.Config) (dsl.Client, error) {
		clientConfig := Config{
			Host:          config.Host,
			User:          config.User,
			Password:      config.AuthPassword,
			TLSSkipVerify: config.Options["TLSSkipVerify"] == "1",
		}
		return NewClient(clientConfig)
	}
	clientDesc := dsl.ClientDesc{
		RequiresUser:       dsl.TristateMaybe,
		SupportedAuthTypes: dsl.AuthTypePassword,
		OptionDescriptions: map[string]string{
			"TLSSkipVerify": "verification of TLS certificates will be skipped if set to 1",
		},
	}
	dsl.RegisterClient("sagemcom", newFunc, clientDesc)
}
