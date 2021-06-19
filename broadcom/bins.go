// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package broadcom

import (
	"bufio"
	"regexp"
	"strconv"
	"strings"

	"3e8.eu/go/dsl/models"
)

var regexpPbParams = regexp.MustCompile(`\((\d+),(\d+)\)`)

func parseBins(status models.Status, pbParams, bits, snr, qln, hlog string) models.Bins {
	var bins models.Bins

	bins.Mode = status.Mode
	bins.Bins = make([]models.Bin, bins.Mode.BinCount())

	parsePbParams(&bins, pbParams)
	parseBits(&bins, bits)
	parseSNR(&bins, snr)
	parseQLN(&bins, qln)
	parseHlog(&bins, hlog)

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
			binType := models.BinTypeDownstream
			if isUpstream {
				binType = models.BinTypeUpstream
			}

			matches := regexpPbParams.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				start, _ := strconv.ParseInt(match[1], 10, 64)
				end, _ := strconv.ParseInt(match[2], 10, 64)

				for num := start; num <= end; num++ {
					bins.Bins[num].Type = binType
				}
			}
		}
	}
}

func parseBits(bins *models.Bins, bits string) {
	var val int64
	parseBinList(bits, func(num int, str string) {
		val, _ = strconv.ParseInt(str, 10, 64)
		bins.Bins[num].Bits = int8(val)
	})
}

func parseSNR(bins *models.Bins, snr string) {
	var val float64
	parseBinList(snr, func(num int, str string) {
		val, _ = strconv.ParseFloat(str, 64)
		bins.Bins[num].SNR = val
	})
}

func parseQLN(bins *models.Bins, qln string) {
	var val float64
	parseBinList(qln, func(num int, str string) {
		val, _ = strconv.ParseFloat(str, 64)
		bins.Bins[num].QLN = val
	})
}

func parseHlog(bins *models.Bins, hlog string) {
	var val float64
	parseBinList(hlog, func(num int, str string) {
		val, _ = strconv.ParseFloat(str, 64)
		bins.Bins[num].Hlog = val
	})
}

func parseBinList(text string, handler func(int, string)) {
	scanner := bufio.NewScanner(strings.NewReader(text))

	for scanner.Scan() {
		line := scanner.Text()
		lineLower := strings.ToLower(line)

		if strings.HasPrefix(lineLower, "tone number") {
			break
		}
	}

	for scanner.Scan() {
		line := scanner.Text()

		data := strings.Fields(line)
		if len(data) == 2 {
			num, _ := strconv.Atoi(data[0])
			str := data[1]

			handler(num, str)
		}
	}
}
