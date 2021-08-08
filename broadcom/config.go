// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package broadcom

import (
	"3e8.eu/go/dsl"
)

type TelnetConfig struct {
	Host     string
	User     string
	Password dsl.PasswordCallback
	Command  string
}

type SSHConfig struct {
	Host        string
	User        string
	Password    dsl.PasswordCallback
	PrivateKeys dsl.PrivateKeysCallback
	KnownHosts  string
	Command     string
}
