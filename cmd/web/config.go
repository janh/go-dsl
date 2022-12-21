// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	ListenAddress          string `toml:",omitempty"`
	HideErrorMessages      bool   `toml:",omitempty"`
	DisableInteractiveAuth bool   `toml:",omitempty"`
	HideRawData            bool   `toml:",omitempty"`
}

func (c Config) EncodeTOMLTable(enc *toml.Encoder) error {
	data := struct {
		Web  Config `toml:",omitempty"`
	}{
		Web: c,
	}

	return enc.Encode(data)
}
