// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package javascript

import (
	"encoding/json"

	"3e8.eu/go/dsl/models"
)

// EncodeBins returns bins data in JSON format for use with the Javascript library.
// The exact structure is not fixed and may change at any time in the future.
func EncodeBins(bins models.Bins) json.RawMessage {
	if bins.Bands.Downstream == nil {
		bins.Bands.Downstream = make([]models.Band, 0, 0)
	}

	if bins.Bands.Upstream == nil {
		bins.Bands.Upstream = make([]models.Band, 0, 0)
	}

	if bins.PilotTones == nil {
		bins.PilotTones = make([]int, 0, 0)
	}

	binsMap := map[string]interface{}{
		"BinCount":       bins.Mode.BinCount(),
		"CarrierSpacing": bins.Mode.CarrierSpacing(),
		"Bands":          bins.Bands,
		"PilotTones":     bins.PilotTones,
		"Bits": map[string]interface{}{
			"Downstream": map[string]interface{}{
				"Data": encodeListInt8(bins.Bits.Downstream.Data),
			},
			"Upstream": map[string]interface{}{
				"Data": encodeListInt8(bins.Bits.Upstream.Data),
			},
		},
		"SNR": map[string]interface{}{
			"Downstream": map[string]interface{}{
				"GroupSize": bins.SNR.Downstream.GroupSize,
				"Data":      encodeListFloat64(bins.SNR.Downstream.Data),
			},
			"Upstream": map[string]interface{}{
				"GroupSize": bins.SNR.Upstream.GroupSize,
				"Data":      encodeListFloat64(bins.SNR.Upstream.Data),
			},
		},
		"QLN": map[string]interface{}{
			"Downstream": map[string]interface{}{
				"GroupSize": bins.QLN.Downstream.GroupSize,
				"Data":      encodeListFloat64(bins.QLN.Downstream.Data),
			},
			"Upstream": map[string]interface{}{
				"GroupSize": bins.QLN.Upstream.GroupSize,
				"Data":      encodeListFloat64(bins.QLN.Upstream.Data),
			},
		},
		"Hlog": map[string]interface{}{
			"Downstream": map[string]interface{}{
				"GroupSize": bins.Hlog.Downstream.GroupSize,
				"Data":      encodeListFloat64(bins.Hlog.Downstream.Data),
			},
			"Upstream": map[string]interface{}{
				"GroupSize": bins.Hlog.Upstream.GroupSize,
				"Data":      encodeListFloat64(bins.Hlog.Upstream.Data),
			},
		},
	}

	data, _ := json.Marshal(binsMap)
	return json.RawMessage(data)
}

// EncodeBinsHistory returns bins history data in JSON format for use with the Javascript library.
// The exact structure is not fixed and may change at any time in the future.
func EncodeBinsHistory(binsHistory models.BinsHistory) json.RawMessage {
	historyMap := map[string]interface{}{
		"SNR": map[string]interface{}{
			"Downstream": map[string]interface{}{
				"GroupSize": binsHistory.SNR.Downstream.GroupSize,
				"Min":       encodeListFloat64(binsHistory.SNR.Downstream.Min),
				"Max":       encodeListFloat64(binsHistory.SNR.Downstream.Max),
			},
			"Upstream": map[string]interface{}{
				"GroupSize": binsHistory.SNR.Upstream.GroupSize,
				"Min":       encodeListFloat64(binsHistory.SNR.Upstream.Min),
				"Max":       encodeListFloat64(binsHistory.SNR.Upstream.Max),
			},
		},
	}

	data, _ := json.Marshal(historyMap)
	return json.RawMessage(data)
}
