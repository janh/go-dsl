// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package common

import (
	"encoding/json"
)

type Message struct {
	State string      `json:"state"`
	Data  interface{} `json:"data,omitempty"`
}

func (m Message) JSON() []byte {
	dataBytes, err := json.Marshal(m)
	if err != nil {
		dataBytes = []byte(`{"state":"error","data":"encoding error"}`)
	}

	return dataBytes
}

type MessageData struct {
	Summary string          `json:"summary"`
	Bins    json.RawMessage `json:"bins"`
	History json.RawMessage `json:"history"`
}
