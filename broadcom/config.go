// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package broadcom

type TelnetConfig struct {
	Host     string
	User     string
	Password string
}

type SSHConfig struct {
	Host       string
	User       string
	Password   string
	PrivateKey string
	KnownHosts string
}
