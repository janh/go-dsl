// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package fritzbox

import (
	"encoding/xml"
	"strings"

	"3e8.eu/go/dsl/models"
)

type interfaceConfigInfoType struct {
	XMLName xml.Name `xml:"Envelope"`
	Data    struct {
		NewUpstreamPower   int64 `xml:"NewUpstreamPower"`
		NewDownstreamPower int64 `xml:"NewDownstreamPower"`
	} `xml:"Body>GetInfoResponse"`
}

type interfaceConfigStatisticsTotalType struct {
	XMLName xml.Name `xml:"Envelope"`
	Data    struct {
		NewErroredSecs         int64 `xml:"NewErroredSecs"`
		NewSeverelyErroredSecs int64 `xml:"NewSeverelyErroredSecs"`
		NewFECErrors           int64 `xml:"NewFECErrors"`
		NewATUCFECErrors       int64 `xml:"NewATUCFECErrors"`
		NewCRCErrors           int64 `xml:"NewCRCErrors"`
		NewATUCCRCErrors       int64 `xml:"NewATUCCRCErrors"`
	} `xml:"Body>GetStatisticsTotalResponse"`
}

func parseTR064Data(status *models.Status, d *rawDataTR064) {
	var info interfaceConfigInfoType
	err := xml.Unmarshal([]byte(d.InterfaceConfigInfo), &info)

	if err == nil {
		if info.Data.NewUpstreamPower != 0 {
			status.UpstreamPower.FloatValue.Float = float64(info.Data.NewUpstreamPower) - 500
		}
		status.UpstreamPower.FloatValue.Valid = true

		if info.Data.NewDownstreamPower != 0 {
			status.DownstreamPower.FloatValue.Float = float64(info.Data.NewDownstreamPower) - 500
		}
		status.DownstreamPower.FloatValue.Valid = true

		// This applies for the UR8-based 7270v3, but maybe other devices need it as well
		if strings.HasPrefix(status.NearEndInventory.Version, "1.52.") {
			status.UpstreamPower.FloatValue.Float *= 0.1
			status.DownstreamPower.FloatValue.Float *= 0.1
		}

		// downstream and upstream power seem to be typically swapped for VDSL2 on Lantiq modems
		if status.DownstreamPower.FloatValue.Float < status.UpstreamPower.FloatValue.Float {
			tmp := status.DownstreamPower.FloatValue.Float
			status.DownstreamPower.FloatValue.Float = status.UpstreamPower.FloatValue.Float
			status.UpstreamPower.FloatValue.Float = tmp
		}
	}

	var statistics interfaceConfigStatisticsTotalType
	err = xml.Unmarshal([]byte(d.InterfaceConfigStatisticsTotal), &statistics)

	if err == nil {
		if !status.DownstreamESCount.Valid {
			status.DownstreamESCount.Int = statistics.Data.NewErroredSecs
			status.DownstreamESCount.Valid = true
		}

		if !status.DownstreamSESCount.Valid {
			status.DownstreamSESCount.Int = statistics.Data.NewSeverelyErroredSecs
			status.DownstreamSESCount.Valid = true
		}

		status.DownstreamFECCount.Int = statistics.Data.NewFECErrors
		status.DownstreamFECCount.Valid = true

		status.UpstreamFECCount.Int = statistics.Data.NewATUCFECErrors
		status.UpstreamFECCount.Valid = true

		status.DownstreamCRCCount.Int = statistics.Data.NewCRCErrors
		status.DownstreamCRCCount.Valid = true

		status.UpstreamCRCCount.Int = statistics.Data.NewATUCCRCErrors
		status.UpstreamCRCCount.Valid = true
	}
}
