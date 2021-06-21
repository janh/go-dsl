// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package draytek

import (
	"bufio"
	"strconv"
	"strings"

	"3e8.eu/go/dsl/models"
)

func parseBins(status models.Status, downstream, upstream string) models.Bins {
	var bins models.Bins

	bins.Mode = status.Mode
	bins.Bins = make([]models.Bin, bins.Mode.BinCount())

	parseBinData(&bins, downstream, models.BinTypeDownstream)
	parseBinData(&bins, upstream, models.BinTypeUpstream)

	return bins
}

func parseBinData(bins *models.Bins, data string, binType models.BinType) {
	scanner := bufio.NewScanner(strings.NewReader(data))

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "*") {
			items := strings.Split(line, "*")
			for _, item := range items {
				readBin(bins, item, binType)
			}
		}
	}

	binCount := bins.Mode.BinCount()
	if binCount > 512 {
		// There is a bug in the bin data for at least some VDSL firmwares: The SNR data of the entire
		// VDSL frequency range is stored in the bins 0-511, with all others being zero.

		var maxSnrIndex, maxBitsIndex int
		for i := 0; i < len(bins.Bins); i++ {
			if bins.Bins[i].SNR > 0 {
				maxSnrIndex = i
			}
			if bins.Bins[i].Bits > 0 {
				maxBitsIndex = i
			}
		}

		factor := binCount / 512
		if maxSnrIndex <= maxBitsIndex/factor {
			for i := binCount - 1; i > 0; i-- {
				bins.Bins[i].SNR = bins.Bins[i/factor].SNR
			}
		}
	}
}

func readBin(bins *models.Bins, item string, binType models.BinType) {
	data := strings.Fields(item)
	if len(data) == 4 {
		num, _ := strconv.Atoi(data[0])
		snr, _ := strconv.ParseFloat(data[1], 64)
		bits, _ := strconv.ParseInt(data[3], 10, 64)

		if bits != 0 || snr != 0 {
			bins.Bins[num].SNR = snr
			bins.Bins[num].Bits = int8(bits)
			bins.Bins[num].Type = binType
		}
	}
}
