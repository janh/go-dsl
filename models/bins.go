// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package models

type BinType int

const (
	BinTypeNone BinType = iota
	BinTypeUpstream
	BinTypeDownstream
)

type Bin struct {
	Type BinType

	// SNR is the signal-to-noise ratio in dB, valid range: -32 to 95
	SNR float64

	// Bits is the number of bits modulated onto the subcarrier, valid range: 0 to 15
	Bits int8

	// QLN is the level of noise present on the line without any xDSL signal in dBm/Hz, valid range: -150 to -23
	QLN float64

	// Hlog is the channel characteristic and estimates the attenuation in dB, valid range: -96.2 to 6
	Hlog float64
}

type Bins struct {
	Mode Mode
	Bins []Bin
}
