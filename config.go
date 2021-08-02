// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package dsl

type PasswordCallback func() string

func Password(password string) PasswordCallback {
	return func() string { return password }
}

type Config struct {
	Type            ClientType
	Host            string
	User            string
	AuthPassword    PasswordCallback
	AuthPrivateKeys []string
	KnownHosts      string
	Options         map[string]string
}
