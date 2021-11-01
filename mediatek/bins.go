// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package mediatek

import (
	"strconv"
	"strings"

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/models"
)

func parseBins(status models.Status, adslShowbpcDs, adslShowbpcUs, adslShowsnr, vdslShowbpcDs, vdslShowbpcUs, vdslShowsnr string) models.Bins {
	var bins models.Bins

	bins.Mode = status.Mode

	var bitsDS, bitsUS, snr string
	if bins.Mode.Type == models.ModeTypeVDSL2 {
		bitsDS = vdslShowbpcDs
		bitsUS = vdslShowbpcUs
		snr = vdslShowsnr
	} else {
		bitsDS = adslShowbpcDs
		bitsUS = adslShowbpcUs
		snr = adslShowsnr
	}

	parseBits(&bins.Bits.Downstream, bitsDS)
	parseBits(&bins.Bits.Upstream, bitsUS)

	helpers.GenerateBandsData(&bins)

	parseSNR(&bins, snr)

	return bins
}

func parseBinList(str string) []string {
	items := strings.Split(str, ",")

	for i, item := range items {
		if len(item) >= 2 && item[0] == '"' && item[len(item)-1] == '"' {
			item = item[1 : len(item)-1]
		}

		items[i] = strings.TrimSpace(item)
	}

	return items
}

func parseBits(out *models.BinsBits, data string) {
	values := parseBinList(data)

	out.Data = make([]int8, len(values))

	for num, val := range values {
		valInt, err := strconv.ParseInt(val, 10, 8)
		if err == nil && valInt > 0 {
			out.Data[num] = int8(valInt)
		}
	}
}

func parseSNR(bins *models.Bins, data string) {
	values := parseBinList(data)

	bins.SNR.Downstream.GroupSize = 1
	bins.SNR.Downstream.Data = make([]float64, len(values))

	bins.SNR.Upstream.GroupSize = 1
	bins.SNR.Upstream.Data = make([]float64, len(values))

	bandDecider, err := helpers.NewBandDecider(bins.Bands)
	if err != nil {
		return
	}

	for num, val := range values {
		valFloat, err := strconv.ParseFloat(val, 64)
		if err == nil && valFloat > 0 {
			if bandDecider.IsDownstream(num) {
				bins.SNR.Downstream.Data[num] = float64(valFloat)
			} else {
				bins.SNR.Downstream.Data[num] = float64(valFloat)
			}
		}
	}
}
