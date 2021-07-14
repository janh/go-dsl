// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package models

import (
	"fmt"
	"io"
	"strings"
)

type Status struct {
	State State
	Mode  Mode

	DownstreamActualRate ValueBandwidth
	UpstreamActualRate   ValueBandwidth

	DownstreamAttainableRate ValueBandwidth
	UpstreamAttainableRate   ValueBandwidth

	DownstreamInterleavingDepth IntValue
	UpstreamInterleavingDepth   IntValue

	DownstreamInterleavingDelay ValueMilliseconds
	UpstreamInterleavingDelay   ValueMilliseconds

	DownstreamAttenuation ValueDecibel
	UpstreamAttenuation   ValueDecibel

	DownstreamSNRMargin ValueDecibel
	UpstreamSNRMargin   ValueDecibel

	DownstreamPower ValuePower
	UpstreamPower   ValuePower

	DownstreamFECCount IntValue
	UpstreamFECCount   IntValue

	DownstreamRTXTXCount IntValue
	UpstreamRTXTXCount   IntValue

	DownstreamRTXCCount IntValue
	UpstreamRTXCCount   IntValue

	DownstreamRTXUCCount IntValue
	UpstreamRTXUCCount   IntValue

	DownstreamCRCCount IntValue
	UpstreamCRCCount   IntValue

	DownstreamESCount IntValue
	UpstreamESCount   IntValue

	FarEndInventory  Inventory
	NearEndInventory Inventory
}

func (s Status) Summary() string {
	var b strings.Builder

	fmt.Fprintf(&b, "           State:    %s\n", s.State)
	fmt.Fprintf(&b, "            Mode:    %s\n", s.Mode)
	fmt.Fprintf(&b, "          Remote:    %s\n", s.FarEndInventory)
	fmt.Fprintf(&b, "           Modem:    %s\n", s.NearEndInventory)
	fmt.Fprintln(&b)

	printValues(&b, "Actual rate", s.DownstreamActualRate, s.UpstreamActualRate)
	printValues(&b, "Attainable rate", s.DownstreamAttainableRate, s.UpstreamAttainableRate)
	printValues(&b, "Interleaving", s.DownstreamInterleavingDepth, s.UpstreamInterleavingDepth)
	printValues(&b, "Interleaving", s.DownstreamInterleavingDelay, s.UpstreamInterleavingDelay)
	fmt.Fprintln(&b)

	printValues(&b, "Attenuation", s.DownstreamAttenuation, s.UpstreamAttenuation)
	printValues(&b, "SNR margin", s.DownstreamSNRMargin, s.UpstreamSNRMargin)
	printValues(&b, "Transmit power", s.DownstreamPower, s.UpstreamPower)
	fmt.Fprintln(&b)

	printValues(&b, "FEC Count", s.DownstreamFECCount, s.UpstreamFECCount)
	printValues(&b, "RTX TX Count", s.DownstreamRTXTXCount, s.UpstreamRTXTXCount)
	printValues(&b, "RTX C Count", s.DownstreamRTXCCount, s.UpstreamRTXCCount)
	printValues(&b, "RTX UC Count", s.DownstreamRTXUCCount, s.UpstreamRTXUCCount)
	printValues(&b, "CRC Count", s.DownstreamCRCCount, s.UpstreamCRCCount)
	printValues(&b, "ES Count", s.DownstreamESCount, s.UpstreamESCount)

	return b.String()
}

func printValues(w io.Writer, label string, valDown, valUp Value) {
	fmt.Fprintf(w, "%16s:    %8s %-6s  %8s %-6s\n", label, valDown.Value(), valDown.Unit(), valUp.Value(), valUp.Unit())
}
