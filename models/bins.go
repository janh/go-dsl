// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package models

type Bins struct {
	Mode Mode

	Bands BandsDownUp

	// PilotTones contains the bin indexes of all pilot tones
	PilotTones []int

	// Bits is the number of bits modulated onto the subcarrier, valid range: 0 to 15
	Bits BinsBitsDownUp

	// SNR is the signal-to-noise ratio in dB, valid range: -32 to 95
	SNR BinsFloatDownUp

	// QLN is the level of noise present on the line without any xDSL signal in dBm/Hz, valid range: -150 to -23
	QLN BinsFloatDownUp

	// Hlog is the channel characteristic and estimates the attenuation in dB, valid range: -96.2 to 6
	Hlog BinsFloatDownUp
}

type BandsDownUp struct {
	Downstream []Band
	Upstream   []Band
}

type Band struct {
	Start int
	End   int
}

type BinsBitsDownUp struct {
	Downstream BinsBits
	Upstream   BinsBits
}

type BinsBits struct {
	Data []int8
}

type BinsFloatDownUp struct {
	Downstream BinsFloat
	Upstream   BinsFloat
}

type BinsFloat struct {
	// GroupSize specifies how many bins are grouped into a single value
	GroupSize int

	// Data contains the actual data, with multiple bins grouped
	Data []float64
}
