// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package models

type State int

const (
	StateUnknown State = iota
	StateDown
	StateDownIdle
	StateDownSilent
	StateInit
	StateInitHandshake
	StateInitChannelDiscovery
	StateInitTraining
	StateInitChannelAnalysisExchange
	StateShowtime
	StateError
)

func (s State) String() string {
	switch s {
	case StateDown:
		return "Down"
	case StateDownIdle:
		return "Down (Idle)"
	case StateDownSilent:
		return "Down (Silent)"
	case StateInit:
		return "Initialization"
	case StateInitHandshake:
		return "Initialization (Handshake)"
	case StateInitChannelDiscovery:
		return "Initialization (Channel Discovery)"
	case StateInitTraining:
		return "Initialization (Training)"
	case StateInitChannelAnalysisExchange:
		return "Initialization (Channel Analysis / Exchange)"
	case StateShowtime:
		return "Showtime"
	case StateError:
		return "Error"
	}
	return "Unknown"
}
