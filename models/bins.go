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
	SNR  float64
	Bits int8
	QLN  float64
	Hlog float64
}

type Bins struct {
	Mode Mode
	Bins []Bin
}
