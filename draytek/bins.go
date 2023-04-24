// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package draytek

import (
	"bufio"
	"regexp"
	"strconv"
	"strings"

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/models"
)

var regexpBandinfo = regexp.MustCompile(`Limits=\[\s*(\d+)-\s*(\d+)\]`)

func parseBins(status models.Status, bandinfo, downstream, upstream, snr, qln, hlog string) models.Bins {
	var bins models.Bins

	bins.Mode = status.Mode

	parseStatusBandinfo(&bins, bandinfo)

	parseShowbinsData(&bins.Bits.Downstream, &bins.SNR.Downstream, bins.Mode.BinCount(), downstream)
	parseShowbinsData(&bins.Bits.Upstream, &bins.SNR.Upstream, bins.Mode.BinCount(), upstream)

	parseStatusSNR(&bins, snr)
	parseStatusQLN(&bins, qln)
	parseStatusHlog(&bins, hlog)

	helpers.GenerateBandsData(&bins)

	return bins
}

func parseStatusBandinfo(bins *models.Bins, data string) {
	scanner := bufio.NewScanner(strings.NewReader(data))

	var bands *[]models.Band

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "US:") {
			bands = &bins.Bands.Upstream
		} else if strings.HasPrefix(line, "DS:") {
			bands = &bins.Bands.Downstream
		}

		submatches := regexpBandinfo.FindStringSubmatch(line)
		if len(submatches) == 3 {
			start, _ := strconv.Atoi(submatches[1])
			end, _ := strconv.Atoi(submatches[2])

			band := models.Band{Start: start, End: end}

			if bands != nil {
				*bands = append(*bands, band)
			}
		}
	}
}

func parseShowbinsData(binsBits *models.BinsBits, binsSNR *models.BinsFloat, binCount int, data string) {
	binsBits.Data = make([]int8, binCount)
	snrData := make([]float64, binCount)
	for i := 0; i < binCount; i++ {
		snrData[i] = -32.5
	}

	scanner := bufio.NewScanner(strings.NewReader(data))

	var maxSNRIndex, maxBitsIndex int

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "*") {
			items := strings.Split(line, "*")
			for _, item := range items {
				readShowbinsBin(binsBits, snrData, &maxSNRIndex, &maxBitsIndex, item)
			}
		}
	}

	handleShowbinsSNR(binsSNR, snrData, maxSNRIndex, maxBitsIndex)
}

func readShowbinsBin(binsBits *models.BinsBits, snrData []float64, maxSNRIndex, maxBitsIndex *int, item string) {
	data := strings.Fields(item)
	if len(data) == 4 {
		num, _ := strconv.Atoi(data[0])
		snr, _ := strconv.ParseFloat(data[1], 64)
		bits, _ := strconv.ParseInt(data[3], 10, 64)

		if bits != 0 {
			binsBits.Data[num] = int8(bits)
			*maxBitsIndex = num
		}
		if snr != 0 && snr != -32 {
			snrData[num] = snr
			*maxSNRIndex = num
		} else {
			snrData[num] = -32.5
		}
	}
}

func handleShowbinsSNR(binsSNR *models.BinsFloat, snrData []float64, maxSNRIndex, maxBitsIndex int) {
	binsSNR.GroupSize = helpers.GuessSNRGroupSize(maxSNRIndex, maxBitsIndex, len(snrData))

	if binsSNR.GroupSize == 1 {
		binsSNR.Data = snrData
	} else {
		binsSNR.Data = make([]float64, len(snrData)/binsSNR.GroupSize)
		for num := 0; num < len(binsSNR.Data); num++ {
			binsSNR.Data[num] = snrData[num]
		}
	}
}

func parseStatusSNR(bins *models.Bins, snr string) {
	parseStatusBins(&bins.SNR, snr, func(val float64, ok bool) float64 {
		if ok {
			return val
		}
		return -32.5
	})
}

func parseStatusQLN(bins *models.Bins, qln string) {
	parseStatusBins(&bins.QLN, qln, func(val float64, ok bool) float64 {
		if ok && val != -150 {
			return val
		}
		return 0
	})
}

func parseStatusHlog(bins *models.Bins, hlog string) {
	parseStatusBins(&bins.Hlog, hlog, func(val float64, ok bool) float64 {
		if ok {
			return val
		}
		return -96.3
	})
}

func parseStatusBinsHeaderItem(line string) (valUS, valDS int) {
	if indexColon := strings.IndexRune(line, ':'); indexColon != -1 {
		lineSplit := strings.Fields(line[indexColon+1:])
		if len(lineSplit) >= 2 {
			valUS, _ = strconv.Atoi(lineSplit[0])
			valDS, _ = strconv.Atoi(lineSplit[1])
		}
	}
	return
}

func parseStatusBins(bins *models.BinsFloatDownUp, data string, handler func(float64, bool) float64) {
	scanner := bufio.NewScanner(strings.NewReader(data))

	var numDataDS, numDataUS int
	var currentBins *models.BinsFloat

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "GroupSize") {
			bins.Upstream.GroupSize, bins.Downstream.GroupSize = parseStatusBinsHeaderItem(line)
			continue
		}

		if strings.HasPrefix(line, "NumData") {
			numDataUS, numDataDS = parseStatusBinsHeaderItem(line)
			continue
		}

		if strings.HasPrefix(line, "US:") {
			bins.Upstream.Data = make([]float64, numDataUS)
			currentBins = &bins.Upstream
		} else if strings.HasPrefix(line, "DS:") {
			bins.Downstream.Data = make([]float64, numDataDS)
			currentBins = &bins.Downstream
		}

		if strings.HasPrefix(line, "bin=") && currentBins != nil && len(currentBins.Data) != 0 && currentBins.GroupSize != 0 {
			readStatusBin(line[4:], currentBins, handler)
		}
	}

	return
}

func readStatusBin(line string, bins *models.BinsFloat, handler func(float64, bool) float64) {
	lineSplit := strings.SplitN(line, ":", 2)
	if len(lineSplit) == 2 {
		numBaseStr := strings.TrimSpace(lineSplit[0])
		numBase, _ := strconv.Atoi(numBaseStr)

		dataSplit := strings.Split(lineSplit[1], ",")
		for i := 0; i < len(dataSplit)-1; i++ {
			num := numBase + i

			valStr := strings.TrimSpace(dataSplit[i])
			val, err := strconv.ParseFloat(valStr, 64)
			ok := err == nil

			if num < len(bins.Data) {
				bins.Data[num] = handler(val, ok)
			}
		}
	}
}
