// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lantiq

import (
	"3e8.eu/go/dsl/models"
)

const (
	lineStateNotInitialized             = 0x0
	lineStateException                  = 0x1
	lineStateNotUpdated                 = 0x10
	lineStateIdleRequest                = 0xff
	lineStateIdle                       = 0x100
	lineStateSilentRequest              = 0x1ff
	lineStateSilent                     = 0x200
	lineStateHandshake                  = 0x300
	lineStateBondingClr                 = 0x310
	lineStateFullInit                   = 0x380
	lineStateShortInit                  = 0x3c0
	lineStateDiscovery                  = 0x400
	lineStateTraining                   = 0x500
	lineStateAnalysis                   = 0x600
	lineStateExchange                   = 0x700
	lineStateShowtimeNoSync             = 0x800
	lineStateShowtimeTCSync             = 0x801
	lineStateFastretrain                = 0x900
	lineStateLowpowerL2                 = 0xa00
	lineStateLoopdiagnosticActive       = 0xb00
	lineStateLoopdiagnosticDataExchange = 0xb10
	lineStateLoopdiagnosticDataRequest  = 0xb20
	lineStateLoopdiagnosticComplete     = 0xc00
	lineStateResync                     = 0xd00
)

func parseLineState(state uint64) models.State {
	switch state {

	case lineStateIdleRequest, lineStateIdle:
		return models.StateIdle

	case lineStateSilentRequest, lineStateSilent:
		return models.StateSilent

	case lineStateHandshake:
		return models.StateHandshake

	case lineStateFullInit, lineStateDiscovery, lineStateTraining, lineStateAnalysis, lineStateExchange:
		return models.StateTraining

	case lineStateShowtimeNoSync, lineStateShowtimeTCSync:
		return models.StateShowtime

	case lineStateException:
		return models.StateError

	}

	return models.StateUnknown
}
