// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fritzbox

import (
	"fmt"
	"strings"
)

type rawDataOverview struct {
	Data       string
	UpdateData string
	Legacy     bool
}

type rawDataStats struct {
	Data string
}

type rawDataSpectrum struct {
	Data string
}

type rawDataTR064 struct {
	InterfaceConfigInfo            string
	InterfaceConfigStatisticsTotal string
}

type rawDataSupport struct {
	Data string
}

type rawData struct {
	Overview    rawDataOverview
	Stats       rawDataStats
	Spectrum    rawDataSpectrum
	TR064       rawDataTR064
	SupportData rawDataSupport
}

func (d *rawData) String() string {
	var b strings.Builder

	fmt.Fprintln(&b, "////// DSL Overview\n")
	fmt.Fprintln(&b, d.Overview.Data+"\n")

	fmt.Fprintln(&b, "////// DSL Overview data\n")
	fmt.Fprintln(&b, d.Overview.UpdateData+"\n")

	fmt.Fprintln(&b, "////// DSL Stats\n")
	fmt.Fprintln(&b, d.Stats.Data+"\n")

	fmt.Fprintln(&b, "////// DSL Spectrum\n")
	fmt.Fprintln(&b, d.Spectrum.Data+"\n")

	fmt.Fprintln(&b, "////// Interface Config Info\n")
	fmt.Fprintln(&b, d.TR064.InterfaceConfigInfo+"\n")

	fmt.Fprintln(&b, "////// Interface Config Statistics Total\n")
	fmt.Fprintln(&b, d.TR064.InterfaceConfigStatisticsTotal+"\n")

	fmt.Fprintln(&b, "////// Support Data\n")
	fmt.Fprintln(&b, d.SupportData.Data+"\n")

	fmt.Fprintln(&b)

	return b.String()
}
