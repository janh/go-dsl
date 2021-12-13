// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package web

type message struct {
	State string      `json:"state"`
	Data  interface{} `json:"data,omitempty"`
}

type data struct {
	Summary   string `json:"summary"`
	GraphBits string `json:"graph_bits"`
	GraphSNR  string `json:"graph_snr"`
	GraphQLN  string `json:"graph_qln"`
	GraphHlog string `json:"graph_hlog"`
}
