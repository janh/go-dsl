// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package common

import (
	"encoding/json"
)

type Message struct {
	State string      `json:"state"`
	Info  interface{} `json:"info,omitempty"`
	Data  interface{} `json:"data,omitempty"`
}

func (m Message) JSON() []byte {
	dataBytes, err := json.Marshal(m)
	if err != nil {
		dataBytes = []byte(`{"state":"error","info":"encoding error"}`)
	}

	return dataBytes
}

type MessageData struct {
	Summary       string          `json:"summary"`
	Bins          json.RawMessage `json:"bins"`
	BinsHistory   json.RawMessage `json:"bins_history"`
	ErrorsHistory json.RawMessage `json:"errors_history"`
}
