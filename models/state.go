// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package models

import (
	"strings"
)

type State int

const (
	StateUnknown State = iota
	StateIdle
	StateSilent
	StateHandshake
	StateTraining
	StateShowtime
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
	}
	return "Unknown"
}

func ParseState(str string) State {
	str = strings.ToLower(strings.TrimSpace(str))

	switch {

	case strings.Contains(str, "idle"), strings.Contains(str, "ready"):
		return StateIdle

	case strings.Contains(str, "silent"):
		return StateSilent

	case strings.Contains(str, "handshake"), strings.Contains(str, "g.994"):
		return StateHandshake

	case strings.Contains(str, "train"), strings.Contains(str, "g.992"), strings.Contains(str, "g.993"),
		strings.Contains(str, "full init"), strings.Contains(str, "discovery"),
		strings.Contains(str, "analysis"), strings.Contains(str, "exchange"):
		return StateTraining

	case strings.Contains(str, "showtime"):
		return StateShowtime

	}

	return StateUnknown
}
