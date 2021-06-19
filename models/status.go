// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package models

import (
	"fmt"
	"strconv"
	"strings"
)

type Status struct {
	State State
	Mode  Mode

	ActualDownstreamRate int32
	ActualUpstreamRate   int32

	AttainableDownstreamRate int32
	AttainableUpstreamRate   int32

	DownstreamInterleavingDepth int16
	UpstreamInterleavingDepth   int16

	DownstreamAttenuation float64
	UpstreamAttenuation   float64

	DownstreamSNRMargin float64
	UpstreamSNRMargin   float64

	DownstreamPower float64
	UpstreamPower   float64

	DownstreamFECCount *int64
	UpstreamFECCount   *int64

	DownstreamRTXTXCount *int64
	UpstreamRTXTXCount   *int64

	DownstreamRTXCCount *int64
	UpstreamRTXCCount   *int64

	DownstreamRTXUCCount *int64
	UpstreamRTXUCCount   *int64

	DownstreamCRCCount int64
	UpstreamCRCCount   int64

	DownstreamESCount int64
	UpstreamESCount   int64

	LinecardVendor  string
	LinecardVersion string

	ModemVendor  string
	ModemVersion string
}

func (s Status) Summary() string {
	var b strings.Builder

	fmt.Fprintf(&b, "           State:    %s\n", s.State)
	fmt.Fprintf(&b, "            Mode:    %s\n", s.Mode)
	fmt.Fprintf(&b, "       Line card:    %s %s\n", s.LinecardVendor, s.LinecardVersion)
	fmt.Fprintf(&b, "           Modem:    %s %s\n", s.ModemVendor, s.ModemVersion)
	fmt.Fprintln(&b)

	fmt.Fprintf(&b, "     Actual rate:    %8d kbit/s  %8d kbit/s\n", s.ActualDownstreamRate, s.ActualUpstreamRate)
	fmt.Fprintf(&b, " Attainable rate:    %8d kbit/s  %8d kbit/s\n", s.AttainableDownstreamRate, s.AttainableUpstreamRate)
	fmt.Fprintf(&b, "    Interleaving:    %8d         %8d\n", s.DownstreamInterleavingDepth, s.UpstreamInterleavingDepth)
	fmt.Fprintln(&b)

	fmt.Fprintf(&b, "     Attenuation:    %8.1f dB      %8.1f dB\n", s.DownstreamAttenuation, s.UpstreamAttenuation)
	fmt.Fprintf(&b, "      SNR margin:    %8.1f dB      %8.1f dB\n", s.DownstreamSNRMargin, s.UpstreamSNRMargin)
	fmt.Fprintf(&b, "  Transmit power:    %8.1f dBm     %8.1f dBm\n", s.DownstreamPower, s.UpstreamPower)
	fmt.Fprintln(&b)

	fmt.Fprintf(&b, "       FEC Count:    %8s         %8s\n", formatNullInt(s.DownstreamFECCount), formatNullInt(s.UpstreamFECCount))
	fmt.Fprintf(&b, "    RTX TX Count:    %8s         %8s\n", formatNullInt(s.DownstreamRTXTXCount), formatNullInt(s.UpstreamRTXTXCount))
	fmt.Fprintf(&b, "     RTX C Count:    %8s         %8s\n", formatNullInt(s.DownstreamRTXCCount), formatNullInt(s.UpstreamRTXCCount))
	fmt.Fprintf(&b, "    RTX UC Count:    %8s         %8s\n", formatNullInt(s.DownstreamRTXUCCount), formatNullInt(s.UpstreamRTXUCCount))
	fmt.Fprintf(&b, "       CRC Count:    %8d         %8d\n", s.DownstreamCRCCount, s.UpstreamCRCCount)
	fmt.Fprintf(&b, "        ES Count:    %8d         %8d\n", s.DownstreamESCount, s.UpstreamESCount)

	return b.String()
}

func formatNullInt(val *int64) string {
	if val != nil {
		return strconv.FormatInt(*val, 10)
	}
	return "-"
}
