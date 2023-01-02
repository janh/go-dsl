// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package models

import (
	"fmt"
	"io"
	"strings"
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

func (h ErrorsHistory) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "End time: %s\n", h.EndTime)
	fmt.Fprintf(&b, "Period length: %s\n", Duration{Duration: h.PeriodLength})
	fmt.Fprintf(&b, "Period count: %d\n", h.PeriodCount)
	fmt.Fprintln(&b)

	printErrorsHistoryList(&b, "Downstream rtx-tx", h.DownstreamRTXTXCount)
	printErrorsHistoryList(&b, "Upstream rtx-tx", h.UpstreamRTXTXCount)

	printErrorsHistoryList(&b, "Downstream rtx-c", h.DownstreamRTXCCount)
	printErrorsHistoryList(&b, "Upstream rtx-c", h.UpstreamRTXCCount)

	printErrorsHistoryList(&b, "Downstream rtx-uc", h.DownstreamRTXUCCount)
	printErrorsHistoryList(&b, "Upstream rtx-uc", h.UpstreamRTXUCCount)

	printErrorsHistoryList(&b, "Downstream FEC", h.DownstreamFECCount)
	printErrorsHistoryList(&b, "Upstream FEC", h.UpstreamFECCount)

	printErrorsHistoryList(&b, "Downstream CRC", h.DownstreamCRCCount)
	printErrorsHistoryList(&b, "Upstream CRC", h.UpstreamCRCCount)

	printErrorsHistoryList(&b, "Downstream ES", h.DownstreamESCount)
	printErrorsHistoryList(&b, "Upstream ES", h.UpstreamESCount)

	printErrorsHistoryList(&b, "Downstream SES", h.DownstreamSESCount)
	printErrorsHistoryList(&b, "Upstream SES", h.UpstreamSESCount)

	return b.String()
}

func printErrorsHistoryList(w io.Writer, label string, data []IntValue) {
	fmt.Fprintf(w, "%s:", label)
	for _, val := range data {
		fmt.Fprintf(w, " %s", val)
	}
	fmt.Fprintf(w, "\n\n")
}
