// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package models

type State int

const (
	StateUnknown State = iota
	StateIdle
	StateSilent
	StateHandshake
	StateTraining
	StateShowtime
	StateError
)

func (s State) String() string {
	switch s {
	case StateIdle:
		return "Idle"
	case StateSilent:
		return "Silent"
	case StateHandshake:
		return "Handshake"
	case StateTraining:
		return "Training"
	case StateShowtime:
		return "Showtime"
	case StateError:
		return "Error"
	}
	return "Unknown"
}
