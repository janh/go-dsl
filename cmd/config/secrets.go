// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"github.com/BurntSushi/toml"
)

var (
	Secrets SecretsData
)

type SecretsData struct {
	Password             string
	PrivateKeyPassphrase string
}

func LoadSecrets(path string) error {
	Secrets = SecretsData{}

	_, err := toml.DecodeFile(path, &Secrets)
	return err
}
