// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package models

import (
	"time"
)

type BinsHistory struct {
	SNR BinsFloatMinMaxDownUp
}

type BinsFloatMinMax struct {
	GroupSize int

	Min []float64
	Max []float64
}

type BinsFloatMinMaxDownUp struct {
	Downstream BinsFloatMinMax
	Upstream   BinsFloatMinMax
}

type ErrorsHistory struct {
	EndTime      time.Time
	PeriodLength time.Duration
	PeriodCount  int

	DownstreamRTXTXCount []IntValue
	UpstreamRTXTXCount   []IntValue

	DownstreamRTXCCount []IntValue
	UpstreamRTXCCount   []IntValue

	DownstreamRTXUCCount []IntValue
	UpstreamRTXUCCount   []IntValue

	DownstreamFECCount []IntValue
	UpstreamFECCount   []IntValue

	DownstreamCRCCount []IntValue
	UpstreamCRCCount   []IntValue

	DownstreamESCount []IntValue
	UpstreamESCount   []IntValue

	DownstreamSESCount []IntValue
	UpstreamSESCount   []IntValue
}
