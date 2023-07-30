// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package bintecelmeg

import (
	"bufio"
	"strconv"
	"strings"

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/models"
)

func interpretBinsList(data string, callback func(index int, val int)) {
	scanner := bufio.NewScanner(strings.NewReader(data))

	for scanner.Scan() {
		line := scanner.Text()

		lineSplit := strings.Split(line, ":")
		if len(lineSplit) != 2 {
			continue
		}

		base, err := strconv.Atoi(lineSplit[0])
		if err != nil {
			continue
		}

		for i, val := range strings.Fields(lineSplit[1]) {
			if valInt, err := strconv.Atoi(val); err == nil {
				callback(base+i, valInt)
			}
		}
	}
}

func interpretBinsBits(mode models.Mode, data string) (out models.BinsBits) {
	out.Data = make([]int8, mode.BinCount())

	interpretBinsList(data, func(index int, val int) {
		if index < len(out.Data) {
			out.Data[index] = int8(val)
		}
	})

	return
}

func interpretBinsSNR(mode models.Mode, bands []models.Band, data string) (out models.BinsFloat) {
	out.Data = make([]float64, mode.BinCount())
	for i := 0; i < mode.BinCount(); i++ {
		out.Data[i] = -32.5
	}

	var maxValidSNRIndex int

	interpretBinsList(data, func(index int, val int) {
		if index < len(out.Data) && val > -32 && val != 0 {
			out.Data[index] = float64(val)
			maxValidSNRIndex = index
		}
	})

	if mode.Type == models.ModeTypeVDSL2 && maxValidSNRIndex < 512 {
		out.Data = out.Data[0:512]
	}

	if len(bands) != 0 {
		maxValidBitsIndex := bands[len(bands)-1].End
		out.GroupSize = helpers.GuessSNRGroupSize(maxValidSNRIndex, maxValidBitsIndex, mode.BinCount())
	} else {
		out.GroupSize = 1
	}

	return
}

func guessVDSL2Subtype(bins *models.Bins) {
	var maxValidDownBitsIndex, maxValidUpBitsIndex int

	if len(bins.Bands.Downstream) != 0 {
		maxValidDownBitsIndex = bins.Bands.Downstream[len(bins.Bands.Downstream)-1].End
	}
	if len(bins.Bands.Upstream) != 0 {
		maxValidUpBitsIndex = bins.Bands.Upstream[len(bins.Bands.Upstream)-1].End
	}

	if maxValidDownBitsIndex < 4096 && maxValidUpBitsIndex < 4096 {
		bins.Mode.Subtype = models.ModeSubtypeProfile17a
	}
}

func interpretBins(status *models.Status, receiveStatistics, transmitStatistics map[string]string) models.Bins {
	var bins models.Bins

	bins.Mode = status.Mode

	bins.Bits.Downstream = interpretBinsBits(bins.Mode, receiveStatistics["dscarrierloadinbits"])
	bins.Bits.Upstream = interpretBinsBits(bins.Mode, transmitStatistics["uscarrierloadinbits"])

	helpers.GenerateBandsData(&bins)

	if bins.Mode.Type == models.ModeTypeVDSL2 && bins.Mode.Subtype == models.ModeSubtypeUnknown {
		guessVDSL2Subtype(&bins)
	}

	bins.SNR.Downstream = interpretBinsSNR(bins.Mode, bins.Bands.Downstream, receiveStatistics["dssnrmarginperbinindb"])

	return bins
}
