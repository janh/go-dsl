// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package mediatek

import (
	"bufio"
	"strconv"
	"strings"

	"3e8.eu/go/dsl/internal/helpers"
	"3e8.eu/go/dsl/models"
)

func parseBins(status models.Status,
	adslShowbpcDs, adslShowbpcUs, adslShowsnr,
	vdslShowbpcDs, vdslShowbpcUs, vdslShowsnr,
	wanVdsl2Dmt, wanVdsl2Qln, wanVdsl2Hlog string) models.Bins {

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

	parseBandsAndPilotTones(&bins, wanVdsl2Dmt)

	parseBits(&bins.Bits.Downstream, bitsDS)
	parseBits(&bins.Bits.Upstream, bitsUS)

	helpers.GenerateBandsData(&bins)

	parseSNR(&bins, snr)

	parseQLN(&bins, wanVdsl2Qln)
	parseHlog(&bins, wanVdsl2Hlog)

	return bins
}

func parseBandsAndPilotTones(bins *models.Bins, data string) {
	scanner := bufio.NewScanner(strings.NewReader(data))

	for scanner.Scan() {
		line := strings.ToLower(scanner.Text())

		isTxBand := strings.Contains(line, "th tx band")
		isRxBand := strings.Contains(line, "th rx band")
		isPilotTones := strings.Contains(line, "pilot tone")

		if isTxBand || isRxBand || isPilotTones {
			lineSplit := strings.SplitN(line, ":", 2)
			if len(lineSplit) != 2 {
				continue
			}

			values := strings.Fields(lineSplit[1])

			if isPilotTones {

				for _, val := range values {
					valInt, err := strconv.Atoi(val)
					if err != nil {
						continue
					}

					bins.PilotTones = append(bins.PilotTones, valInt)
				}

			} else if len(values) == 3 {

				start, errStart := strconv.Atoi(values[0])
				end, errEnd := strconv.Atoi(values[1])

				if errStart != nil || errEnd != nil {
					continue
				}

				band := models.Band{Start: start, End: end}

				if isTxBand {
					bins.Bands.Upstream = append(bins.Bands.Upstream, band)
				} else if isRxBand {
					bins.Bands.Downstream = append(bins.Bands.Downstream, band)
				}

			}
		}
	}
}

func parseQuotedBinList(str string) []string {
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
	values := parseQuotedBinList(data)

	out.Data = make([]int8, len(values))

	for num, val := range values {
		valInt, err := strconv.ParseInt(val, 10, 8)
		if err == nil && valInt > 0 {
			out.Data[num] = int8(valInt)
		}
	}
}

func parseSNR(bins *models.Bins, data string) {
	values := parseQuotedBinList(data)

	bins.SNR.Downstream.GroupSize = 1
	bins.SNR.Downstream.Data = make([]float64, len(values))
	for num := range bins.SNR.Downstream.Data {
		bins.SNR.Downstream.Data[num] = -32.5
	}

	bins.SNR.Upstream.GroupSize = 1
	bins.SNR.Upstream.Data = make([]float64, len(values))
	for num := range bins.SNR.Upstream.Data {
		bins.SNR.Upstream.Data[num] = -32.5
	}

	bandDecider, err := helpers.NewBandDecider(bins.Bands)
	if err != nil {
		return
	}

	for num, val := range values {
		valFloat, err := strconv.ParseFloat(val, 64)
		if err == nil && valFloat != 0 {
			if bandDecider.IsDownstream(num) {
				bins.SNR.Downstream.Data[num] = float64(valFloat)
			} else {
				bins.SNR.Upstream.Data[num] = float64(valFloat)
			}
		}
	}
}

func parseLineSeparatedBinList(str string) []string {
	var items []string

	scanner := bufio.NewScanner(strings.NewReader(str))

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		if len(fields) > 0 {
			val := fields[len(fields)-1]

			indexNonDigit := strings.IndexFunc(val, func(r rune) bool { return r < '0' || r > '9' })
			if indexNonDigit != -1 {
				continue
			}

			items = append(items, val)
		}
	}

	return items
}

func parseQLN(bins *models.Bins, data string) {
	values := parseLineSeparatedBinList(data)
	if len(values) == 0 {
		return
	}

	groupSize := bins.Mode.BinCount() / len(values)

	bins.QLN.Downstream.GroupSize = groupSize
	bins.QLN.Downstream.Data = make([]float64, len(values))

	bins.QLN.Upstream.GroupSize = groupSize
	bins.QLN.Upstream.Data = make([]float64, len(values))

	bandDecider, err := helpers.NewBandDecider(bins.Bands)
	if err != nil {
		return
	}

	for num, val := range values {
		valUint, err := strconv.ParseUint(val, 10, 8)
		if err == nil && valUint != 255 {
			valFloat := -23 - float64(valUint)/2

			if bandDecider.IsDownstream(num*groupSize + groupSize/2) {
				bins.QLN.Downstream.Data[num] = float64(valFloat)
			} else {
				bins.QLN.Upstream.Data[num] = float64(valFloat)
			}
		}
	}
}

func parseHlog(bins *models.Bins, data string) {
	values := parseLineSeparatedBinList(data)
	if len(values) == 0 {
		return
	}

	groupSize := bins.Mode.BinCount() / len(values)

	bins.Hlog.Downstream.GroupSize = groupSize
	bins.Hlog.Downstream.Data = make([]float64, len(values))
	for num := range bins.Hlog.Downstream.Data {
		bins.Hlog.Downstream.Data[num] = -96.3
	}

	bins.Hlog.Upstream.GroupSize = groupSize
	bins.Hlog.Upstream.Data = make([]float64, len(values))
	for num := range bins.Hlog.Upstream.Data {
		bins.Hlog.Upstream.Data[num] = -96.3
	}

	bandDecider, err := helpers.NewBandDecider(bins.Bands)
	if err != nil {
		return
	}

	for num, val := range values {
		valUint, err := strconv.ParseUint(val, 10, 10)
		if err == nil && valUint != 1023 {
			valFloat := 6 - float64(valUint)/10

			if bandDecider.IsDownstream(num*groupSize + groupSize/2) {
				bins.Hlog.Downstream.Data[num] = float64(valFloat)
			} else {
				bins.Hlog.Upstream.Data[num] = float64(valFloat)
			}
		}
	}
}
