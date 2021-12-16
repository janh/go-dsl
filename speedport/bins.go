// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package speedport

import (
	"strconv"
	"strings"

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/models"
)

func interpretBins(status *models.Status, valuesDSL map[string]responseVar) models.Bins {
	var bins models.Bins

	bins.Mode = status.Mode

	interpretBitsAndMode(&bins, valuesDSL)

	helpers.GenerateBandsData(&bins)

	return bins
}

func interpretBitsAndMode(bins *models.Bins, values map[string]responseVar) {
	bitsDown, lastNonZeroDown := interpretBitsList(values, "BinallocaDown")
	bitsUp, lastNonZeroUp := interpretBitsList(values, "BinallocaUp")

	if len(bitsDown) != 512 || len(bitsUp) != 512 {
		return
	}

	if bins.Mode.Type == models.ModeTypeVDSL2 {
		if lastNonZeroUp < 180 && lastNonZeroDown > 170 {
			bins.Mode.Subtype = models.ModeSubtypeProfile35b
		} else {
			bins.Mode.Subtype = models.ModeSubtypeProfile17a
		}
	}

	binCount := bins.Mode.BinCount()
	factor := binCount / 512

	bins.Bits.Downstream.Data = scaleBitsData(bitsDown, factor)
	bins.Bits.Upstream.Data = scaleBitsData(bitsUp, factor)
}

func scaleBitsData(in []int8, factor int) []int8 {
	count := len(in) * factor
	out := make([]int8, count, count)

	for i, val := range in {
		for j := i * factor; j < i*factor+factor; j++ {
			out[j] = val
		}
	}

	return out
}

func interpretBitsList(values map[string]responseVar, key string) ([]int8, int) {
	data, ok := values[key]
	if !ok {
		return nil, 0
	}

	var out []int8
	var max int

	lines := strings.Split(data.Value, "||")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		items := strings.Split(line, "|")
		if len(items) != 9 {
			return nil, 0
		}

		for i, item := range items[1:] {
			val, _ := strconv.ParseUint(item, 16, 4)
			out = append(out, int8(val))
			if val > 0 {
				max = i
			}
		}
	}

	return out, max
}
