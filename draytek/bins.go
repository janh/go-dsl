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

func parseBins(status models.Status, downstream, upstream, qln, hlog string) models.Bins {
	var bins models.Bins

	bins.Mode = status.Mode
	bins.Bins = make([]models.Bin, bins.Mode.BinCount())

	parseShowbinsData(&bins, downstream, models.BinTypeDownstream)
	parseShowbinsData(&bins, upstream, models.BinTypeUpstream)

	parseStatusQLN(&bins, qln)
	parseStatusHlog(&bins, hlog)

	return bins
}

func parseShowbinsData(bins *models.Bins, data string, binType models.BinType) {
	scanner := bufio.NewScanner(strings.NewReader(data))

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "*") {
			items := strings.Split(line, "*")
			for _, item := range items {
				readShowbinsBin(bins, item, binType)
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

func readShowbinsBin(bins *models.Bins, item string, binType models.BinType) {
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

func parseStatusQLN(bins *models.Bins, qln string) {
	parseStatusBins(qln, func(num int, val float64) {
		if val != -150 {
			bins.Bins[num].QLN = val
		}
	})
}

func parseStatusHlog(bins *models.Bins, hlog string) {
	parseStatusBins(hlog, func(num int, val float64) {
		if val != 0 {
			bins.Bins[num].Hlog = val
		}
	})
}

func parseStatusBins(data string, handler func(int, float64)) {
	scanner := bufio.NewScanner(strings.NewReader(data))

	var groupSize, groupSizeDS, groupSizeUS int

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "GroupSize") {
			if indexColon := strings.IndexRune(line, ':'); indexColon != -1 {
				lineSplit := strings.Fields(line[indexColon+1:])
				if len(lineSplit) >= 2 {
					groupSizeDS, _ = strconv.Atoi(lineSplit[0])
					groupSizeUS, _ = strconv.Atoi(lineSplit[1])
				}
			}
			continue
		}

		if strings.HasPrefix(line, "US:") {
			groupSize = groupSizeDS
		} else if strings.HasPrefix(line, "DS:") {
			groupSize = groupSizeUS
		}

		if strings.HasPrefix(line, "bin=") && groupSize != 0 {
			readStatusBin(line[4:], groupSize, handler)
		}
	}
}

func readStatusBin(line string, groupSize int, handler func(int, float64)) {
	lineSplit := strings.SplitN(line, ":", 2)
	if len(lineSplit) == 2 {
		numBaseStr := strings.TrimSpace(lineSplit[0])
		numBase, _ := strconv.Atoi(numBaseStr)

		dataSplit := strings.Split(lineSplit[1], ",")
		for i := 0; i < len(dataSplit)-1; i++ {
			valStr := strings.TrimSpace(dataSplit[i])
			val, _ := strconv.ParseFloat(valStr, 64)
			numGroup := (numBase + i) * groupSize

			for j := 0; j < groupSize; j++ {
				num := numGroup + j
				handler(num, val)
			}
		}
	}
}
