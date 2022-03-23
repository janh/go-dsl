// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

import (
	"encoding/json"
)

type message struct {
	State string      `json:"state"`
	Data  interface{} `json:"data,omitempty"`
}

type data struct {
	Summary string          `json:"summary"`
	Bins    json.RawMessage `json:"bins"`
	History json.RawMessage `json:"history"`
}
