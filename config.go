// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package dsl

type PasswordCallback func() (string, error)

func Password(password string) PasswordCallback {
	return func() (string, error) { return password, nil }
}

type PrivateKeysCallback struct {
	Keys       func() ([]string, error)
	Passphrase func(fingerprint string) (string, error)
}

func PrivateKey(key string, passphrase string) PrivateKeysCallback {
	return PrivateKeysCallback{
		Keys:       func() ([]string, error) { return []string{key}, nil },
		Passphrase: func(string) (string, error) { return passphrase, nil },
	}
}

type Config struct {
	Type            ClientType
	Host            string
	User            string
	AuthPassword    PasswordCallback
	AuthPrivateKeys PrivateKeysCallback
	KnownHosts      string
	Options         map[string]string
}
