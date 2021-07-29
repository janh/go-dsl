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

func parseBins(status models.Status, bandinfo, downstream, upstream, qln, hlog string) models.Bins {
	var bins models.Bins

	bins.Mode = status.Mode

	parseStatusBandinfo(&bins, bandinfo)

	parseShowbinsData(&bins.Bits.Downstream, &bins.SNR.Downstream, bins.Mode.BinCount(), downstream)
	parseShowbinsData(&bins.Bits.Upstream, &bins.SNR.Upstream, bins.Mode.BinCount(), upstream)

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
		if snr != 0 {
			snrData[num] = snr
			*maxSNRIndex = num
		}
	}
}

func handleShowbinsSNR(binsSNR *models.BinsFloat, snrData []float64, maxSNRIndex, maxBitsIndex int) {
	if maxSNRIndex == 0 {
		return
	}

	if maxBitsIndex > 512 {

		// The SNR data of the entire frequency range is stored in the bins 0-511, with all others being zero.

		maxFactor := len(snrData) / maxSNRIndex
		var factor int
		for factor = 1; factor < maxFactor; factor *= 2 {
			// after applying factor, maxSNRIndex should be at most 10% lower than maxBitsIndex, because:
			// - maxSNRIndex > maxBitsIndex is common when SNR is too low to allocate bits
			// - maxSNRIndex < maxBitsIndex unlikely, as SNR value needed to allocate bins
			if float64(maxSNRIndex*factor)/float64(maxBitsIndex) > 0.9 {
				break
			}
		}

		binsSNR.GroupSize = factor
		binsSNR.Data = make([]float64, len(snrData)/factor)

		for num := 0; num <= maxSNRIndex; num++ {
			binsSNR.Data[num] = snrData[num]
		}

	} else {

		binsSNR.GroupSize = 1
		binsSNR.Data = snrData

	}
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

			bins.Data[num] = handler(val, ok)
		}
	}
}
