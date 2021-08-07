// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package broadcom

import (
	"bufio"
	"regexp"
	"strconv"
	"strings"

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/models"
)

var regexpPbParams = regexp.MustCompile(`\((\d+),(\d+)\)`)

func parseBins(status models.Status, pbParams, bits, snr, qln, hlog string) models.Bins {
	var bins models.Bins

	bins.Mode = status.Mode

	parsePbParams(&bins, pbParams)
	parseBits(&bins, bits)
	parseSNR(&bins, snr)
	parseQLN(&bins, qln)
	parseHlog(&bins, hlog)

	helpers.GenerateBandsData(&bins)

	return bins
}

func parsePbParams(bins *models.Bins, pbParams string) {
	scanner := bufio.NewScanner(strings.NewReader(pbParams))

	for scanner.Scan() {
		line := scanner.Text()
		lineLower := strings.ToLower(line)

		if strings.HasPrefix(lineLower, "medley") && strings.HasSuffix(lineLower, "band plan") {
			break
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		lineLower := strings.ToLower(line)

		isUpstream := strings.HasPrefix(lineLower, "us:")
		isDownstream := strings.HasPrefix(lineLower, "ds:")

		if isUpstream || isDownstream {
			matches := regexpPbParams.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				start, _ := strconv.Atoi(match[1])
				end, _ := strconv.Atoi(match[2])

				band := models.Band{Start: start, End: end}

				if isUpstream {
					bins.Bands.Upstream = append(bins.Bands.Upstream, band)
				} else if isDownstream {
					bins.Bands.Downstream = append(bins.Bands.Downstream, band)
				}
			}
		}
	}
}

func parseBits(bins *models.Bins, bits string) {
	binCount := bins.Mode.BinCount()

	bins.Bits.Downstream.Data = make([]int8, binCount)
	bins.Bits.Upstream.Data = make([]int8, binCount)

	var val int64
	parseBinList(bits, bins.Bands, func(num int, str string, isDownstream bool) {
		if num >= binCount {
			return
		}

		val, _ = strconv.ParseInt(str, 10, 64)

		if val != 0 {
			if isDownstream {
				bins.Bits.Downstream.Data[num] = int8(val)
			} else {
				bins.Bits.Upstream.Data[num] = int8(val)
			}
		}
	})
}

func parseSNR(bins *models.Bins, snr string) {
	binCount := bins.Mode.BinCount()

	bins.SNR.Downstream.GroupSize = 1
	bins.SNR.Downstream.Data = make([]float64, binCount)

	bins.SNR.Upstream.GroupSize = 1
	bins.SNR.Upstream.Data = make([]float64, binCount)

	var val float64
	parseBinList(snr, bins.Bands, func(num int, str string, isDownstream bool) {
		if num >= binCount {
			return
		}

		val, _ = strconv.ParseFloat(str, 64)

		if val != 0 {
			if isDownstream {
				bins.SNR.Downstream.Data[num] = val
			} else {
				bins.SNR.Upstream.Data[num] = val
			}
		}
	})

	isValid := func(val float64) bool {
		return val != 0
	}
	adjustGroupSize(&bins.SNR.Downstream, isValid)
	adjustGroupSize(&bins.SNR.Upstream, isValid)
}

func parseQLN(bins *models.Bins, qln string) {
	binCount := bins.Mode.BinCount()

	bins.QLN.Downstream.GroupSize = 1
	bins.QLN.Downstream.Data = make([]float64, binCount)

	bins.QLN.Upstream.GroupSize = 1
	bins.QLN.Upstream.Data = make([]float64, binCount)

	var val float64
	parseBinList(qln, bins.Bands, func(num int, str string, isDownstream bool) {
		if num >= binCount {
			return
		}

		val, _ = strconv.ParseFloat(str, 64)

		if val != 0 && val != -160 {
			if isDownstream {
				bins.QLN.Downstream.Data[num] = val
			} else {
				bins.QLN.Upstream.Data[num] = val
			}
		}
	})

	isValid := func(val float64) bool {
		return val != 0
	}
	adjustGroupSize(&bins.QLN.Downstream, isValid)
	adjustGroupSize(&bins.QLN.Upstream, isValid)
}

func parseHlog(bins *models.Bins, hlog string) {
	binCount := bins.Mode.BinCount()

	bins.Hlog.Downstream.GroupSize = 1
	bins.Hlog.Downstream.Data = make([]float64, binCount)
	for num := range bins.Hlog.Downstream.Data {
		bins.Hlog.Downstream.Data[num] = -96.3
	}

	bins.Hlog.Upstream.GroupSize = 1
	bins.Hlog.Upstream.Data = make([]float64, binCount)
	for num := range bins.Hlog.Upstream.Data {
		bins.Hlog.Upstream.Data[num] = -96.3
	}

	var val float64
	var err error
	parseBinList(hlog, bins.Bands, func(num int, str string, isDownstream bool) {
		if num >= binCount {
			return
		}

		val, err = strconv.ParseFloat(str, 64)

		if err == nil && val > -96 {
			if isDownstream {
				bins.Hlog.Downstream.Data[num] = val
			} else {
				bins.Hlog.Upstream.Data[num] = val
			}
		}
	})

	isValid := func(val float64) bool {
		return val > -96
	}
	adjustGroupSize(&bins.Hlog.Downstream, isValid)
	adjustGroupSize(&bins.Hlog.Upstream, isValid)
}

func parseBinList(text string, bands models.BandsDownUp, handler func(int, string, bool)) {
	scanner := bufio.NewScanner(strings.NewReader(text))

	for scanner.Scan() {
		line := scanner.Text()
		lineLower := strings.ToLower(line)

		if strings.HasPrefix(lineLower, "tone number") {
			break
		}
	}

	bandDecider, err := helpers.NewBandDecider(bands)
	if err != nil {
		return
	}

	for scanner.Scan() {
		line := scanner.Text()

		data := strings.Fields(line)
		if len(data) == 2 {
			num, _ := strconv.Atoi(data[0])
			str := data[1]

			handler(num, str, bandDecider.IsDownstream(num))
		}
	}
}

func determineGroupSize(bins *models.BinsFloat, isValid func(val float64) bool) (groupSize int) {
	groupSize = 1

	for ; groupSize < 16; groupSize *= 2 {
		nextGroupSize := groupSize * 2

		if len(bins.Data)%nextGroupSize != 0 {
			return
		}

		for i := 0; i < len(bins.Data)-nextGroupSize; i += nextGroupSize {
			a := bins.Data[i]
			b := bins.Data[i+groupSize]

			if isValid(a) && isValid(b) && a != b {
				return
			}
		}

	}

	return
}

func adjustGroupSize(bins *models.BinsFloat, isValid func(val float64) bool) {
	groupSize := determineGroupSize(bins, isValid)
	if groupSize == 1 {
		return
	}

	newData := make([]float64, len(bins.Data)/groupSize)
	for num := 0; num < len(newData); num++ {
		newData[num] = bins.Data[num*groupSize+groupSize/2]
	}

	bins.GroupSize = groupSize
	bins.Data = newData
}
