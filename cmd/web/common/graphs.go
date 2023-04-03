// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package common

import (
	"3e8.eu/go/dsl/graphs"
)

func GetGraphTemplateData() interface{} {
	return map[string]interface{}{
		"LegendBits":           graphs.GetBitsGraphLegend().Items,
		"LegendSNR":            graphs.GetSNRGraphWithHistoryLegend().Items,
		"LegendQLN":            graphs.GetQLNGraphLegend().Items,
		"LegendHlog":           graphs.GetHlogGraphLegend().Items,
		"LegendRetransmission": graphs.GetDownstreamRetransmissionGraphLegend().Items,
		"LegendErrors":         graphs.GetDownstreamErrorsGraphLegend().Items,
		"LegendErrorSeconds":   graphs.GetDownstreamErrorSecondsGraphLegend().Items,
	}
}
